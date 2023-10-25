package tracker

import (
	"context"
	"fmt"
	"time"

	"consumptionexp/config"
	"consumptionexp/query"
	"consumptionexp/store"
	"consumptionexp/types"
	"consumptionexp/util"

	v1 "k8s.io/api/core/v1"

	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"github.com/apsdehal/go-logger"
	"k8s.io/utils/clock"
)

type resourceType int64

const (
	typeCPU resourceType = 0
	typeMem resourceType = 1
)

const (
	podNameKey       = "pod-name"
	containerNameKey = "container-name"
)

// K8sClientStore is the in-memory store for cluster-k8sclient map.
type K8sClientStore struct {
	clientmap map[string]*util.K8sClient
	container *util.ContainerClient
	logger    *logger.Logger
	nodeStore *store.Nodes
}

// NewK8sClientStore creates new instance of K8sClientStore.
func NewK8sClientStore(container *util.ContainerClient, logger *logger.Logger, nodeStore *store.Nodes) *K8sClientStore {
	return &K8sClientStore{
		clientmap: map[string]*util.K8sClient{},
		container: container,
		logger:    logger,
		nodeStore: nodeStore,
	}
}

func (s *K8sClientStore) GetOrCreate(ctx context.Context, cluster *types.Cluster) (*util.K8sClient, error) {
	key := cluster.String()
	c, exists := s.clientmap[key]
	if !exists {
		s.logger.Infof("K8sClient for cluster %s not found, creating...", key)
		v1Cluster, err := s.container.GetCluster(ctx, cluster)
		if err != nil {
			return nil, err
		}
		cluster.Cluster = v1Cluster
		c, err = util.NewK8sClient(ctx, cluster, s.nodeStore)
		if err != nil {
			return nil, err
		}
		s.clientmap[key] = c
	}
	return c, nil
}

// MetricsTracker creates new instance of metrics tracker.
type MetricsTracker struct {
	cfg        *config.Configurations
	clock      clock.WithTickerAndDelayedExecution
	nodeStore  *store.Nodes
	bigquery   *util.BigQueryClient
	container  *util.ContainerClient
	monitoring *util.MonitoringClient
	k8sclients *K8sClientStore
	logger     *logger.Logger
	cancel     func()
}

// Usage is the usage data extracted from metrics.
type Usage struct {
	StartTime time.Time
	EndTime   time.Time
	Value     float64
	AvgUsage  float64
}

// NewMetricsTracker creates a MetricsTracker object.
func NewMetricsTracker(ctx context.Context, cfg *config.Configurations, log *logger.Logger) *MetricsTracker {
	ns, err := store.NewNodeStore()
	if err != nil {
		log.Errorf("error initializing NodeCache store: %v", err)
		return nil
	}

	bq, err := util.NewBigQueryClient(ctx, cfg)
	if err != nil {
		log.Errorf("error initilizing bigquery client: %v", err)
		return nil
	}

	c, err := util.NewContainerClient(ctx)
	if err != nil {
		log.Errorf("error initilizing container client: %v", err)
		return nil
	}

	m, err := util.NewMonitoringClient(ctx)
	if err != nil {
		log.Errorf("error initilizing monitoring client: %v", err)
		return nil
	}

	return &MetricsTracker{
		cfg:        cfg,
		logger:     log,
		clock:      clock.RealClock{},
		nodeStore:  ns,
		k8sclients: NewK8sClientStore(c, log, ns),
		bigquery:   bq,
		container:  c,
		monitoring: m,
	}
}

// Start the tracker for metrics data.
func (t *MetricsTracker) Start(ctx context.Context) {
	t.logger.Infof("Starting the metrics tracker.")
	exportTicker := t.clock.NewTicker(config.ExportInterval)
	go func() {
		defer exportTicker.Stop()
		t.export(ctx)
		for {
			select {
			case <-exportTicker.C():
				t.export(ctx)
			}
		}
	}()
}

func (t *MetricsTracker) export(ctx context.Context) {
	t.logger.Infof("Start parsing and exporting metrics...")
	urs := []types.UsageRecord{}
	t.logger.Infof("Exporting CPU metrics...")
	urs = append(urs, t.exportCPU(ctx)...)
	t.logger.Infof("Exporting Memory metrics...")
	urs = append(urs, t.exportMemory(ctx)...)
	t.logger.Infof("Exporting %v usage records to bigquery...", len(urs))
	if err := t.bigquery.WriteUsage(ctx, urs); err != nil {
		t.logger.Errorf("Error exporting to bigquery: %v", err)
	}
	t.logger.Infof("Exporting to bigquery finished.")
}

// Stop the metrics tracker.
func (t *MetricsTracker) Stop() {
	t.bigquery.Close()
	t.container.Close()
	t.monitoring.Close()

	if t.cancel != nil {
		t.cancel()
	}
}

func (t *MetricsTracker) exportCPU(ctx context.Context) []types.UsageRecord {
	return t.queryUsages(ctx, query.CPUUsageTime, typeCPU)
}

func (t *MetricsTracker) exportMemory(ctx context.Context) []types.UsageRecord {
	return t.queryUsages(ctx, query.MemoryUsedBytes, typeMem)
}

func (t *MetricsTracker) queryUsages(ctx context.Context, q string, r resourceType) []types.UsageRecord {
	urs := []types.UsageRecord{}
	dp, err := t.monitoring.QueryTimeSeries(ctx, t.cfg, q)
	if err != nil {
		t.logger.Errorf("Error found querying time series: %v", err)
		return urs
	}

	t.logger.Infof("Found %d data point records.", len(dp))
	f := 0.0

	for _, d := range dp {
		labels, err := parseMetricsLabels(d.GetLabelValues())
		if err != nil {
			t.logger.Errorf("Error parsing labels: %v", err)
			continue
		}
		usage := calcUsage(d.GetPointData())

		cluster := &types.Cluster{
			ProjectID: labels.ProjectID,
			Location:  labels.Location,
			Name:      labels.ClusterName,
		}
		k8sClient, err := t.k8sclients.GetOrCreate(ctx, cluster)
		if err != nil {
			t.logger.Errorf("error getting k8sclient for cluster %v: %v", cluster, err)
			continue
		}

		pod, err := k8sClient.GetPod(ctx, labels.Namespace, labels.PodName)
		if err != nil {
			t.logger.Errorf("error occurred getting pod: %v", err)
			continue
		}

		podLabels := pod.GetLabels()
		if podLabels == nil {
			podLabels = map[string]string{}
		}
		podLabels[podNameKey] = labels.PodName
		podLabels[containerNameKey] = labels.Container

		node, err := k8sClient.GetNode(ctx, pod.Spec.NodeName)
		if err != nil {
			t.logger.Errorf("error occurred getting node: %v", err)
			continue
		}

		ur := types.UsageRecord{
			ClusterLocation: labels.Location,
			ClusterName:     labels.ClusterName,
			Namespace:       labels.Namespace,
			StartTime:       usage.StartTime,
			EndTime:         usage.EndTime,
			Labels:          types.ToLabelSlice(podLabels),
			Project: &types.Project{
				ID: labels.ProjectID,
			},
			Usage: &types.Usage{
				Amount: usage.Value,
			},
		}

		switch r {
		case typeCPU:
			ur.ResourceName = v1.ResourceCPU
			ur.Usage.Unit = types.UsageUnitSeconds
			ur.Fraction = usage.AvgUsage / float64(node.CPUSize)
			ur.CloudResourceSize = node.CPUSize
		case typeMem:
			ur.ResourceName = v1.ResourceMemory
			ur.Usage.Unit = types.UsageUnitByteSeconds
			ur.Fraction = usage.AvgUsage / float64(node.MemSize)
			ur.CloudResourceSize = node.MemSize
		}

		f += ur.Fraction

		// TODO: fill in SKU
		urs = append(urs, ur)
	}

	t.logger.Debugf("Total fraction=%f", f)
	return urs
}

// parseLabels translates an array of LabelValue to a Labels struct.
func parseMetricsLabels(labels []*monitoringpb.LabelValue) (*types.Labels, error) {
	if len(labels) != 6 {
		return nil, fmt.Errorf("error parsing label values: %v", labels)
	}
	return &types.Labels{
		ProjectID:   labels[0].GetStringValue(),
		Location:    labels[1].GetStringValue(),
		ClusterName: labels[2].GetStringValue(),
		Namespace:   labels[3].GetStringValue(),
		PodName:     labels[4].GetStringValue(),
		Container:   labels[5].GetStringValue(),
	}, nil
}

func calcUsage(pds []*monitoringpb.TimeSeriesData_PointData) Usage {
	sum, u := 0.0, 0.0
	for _, pd := range pds {
		duration := pd.GetTimeInterval().GetEndTime().AsTime().Sub(pd.GetTimeInterval().GetStartTime().AsTime()).Seconds()
		sum += pd.GetValues()[0].GetDoubleValue() * duration
		u += pd.GetValues()[0].GetDoubleValue()
	}
	startTime := pds[0].GetTimeInterval().GetStartTime().AsTime()
	endTime := pds[len(pds)-1].GetTimeInterval().GetEndTime().AsTime()
	return Usage{
		StartTime: startTime,
		EndTime:   endTime,
		Value:     sum,
		AvgUsage:  u / float64(len(pds)),
	}
}
