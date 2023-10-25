package types

import (
	"fmt"
	"strconv"
	"time"

	v1 "k8s.io/api/core/v1"
)

const (
	gkePreemptibleLabel = "cloud.google.com/gke-preemptible"
	gkeSpotLabel        = "cloud.google.com/gke-spot"
)

// NodeCache defines the cache for Node resources
type NodeCache struct {
	ProjectID       string
	ClusterName     string
	ClusterLocation string
	NodeName        string
	MachineType     string
	Preemptible     bool
	Region          string
	CPUSKU          string
	CPUSize         int64
	MemSKU          string
	MemSize         int64
	LastUpdate      time.Time
}

// FromV1Node converts a v1.Node object to a NodeCache object.
func FromV1Node(node *v1.Node, cluster *Cluster) (*NodeCache, error) {
	cache := &NodeCache{
		ProjectID:       cluster.ProjectID,
		ClusterName:     cluster.Name,
		ClusterLocation: cluster.Location,
		NodeName:        node.ObjectMeta.Name,
		MachineType:     extractNodeInstanceType(node),
		Region:          extractNodeRegion(node),
		CPUSize:         node.Status.Capacity.Cpu().Value(),
		MemSize:         node.Status.Capacity.Memory().Value(),
		LastUpdate:      time.Now(),
	}

	var err error
	if cache.Preemptible, err = isPreemptible(node); err != nil {
		return nil, fmt.Errorf("unable to check if node is preemptible: %v", err)
	}

	// TODO: Also populate CPU and memory SKUs.
	return cache, nil
}

func extractNodeRegion(n *v1.Node) string {
	return readLabels(n.GetLabels(), v1.LabelZoneRegionStable, v1.LabelZoneRegion)
}

func extractNodeZone(n *v1.Node) string {
	return readLabels(n.GetLabels(), v1.LabelZoneFailureDomainStable, v1.LabelZoneFailureDomain)
}

func extractNodeInstanceType(n *v1.Node) string {
	return readLabels(n.GetLabels(), v1.LabelInstanceTypeStable, v1.LabelInstanceType)
}

func isPreemptible(node *v1.Node) (bool, error) {
	labels := node.GetLabels()
	for _, l := range []string{gkePreemptibleLabel, gkeSpotLabel} {
		val, ok := labels[l]
		if !ok {
			// Label is missing.
			continue
		}
		parsed, err := strconv.ParseBool(val)
		if err != nil {
			return false, fmt.Errorf("invalid value for label %s: %s", gkePreemptibleLabel, val)
		}
		if parsed {
			return true, nil
		}
	}
	return false, nil
}

// readLabels examines the given labels by checking the provided keys in
// sequence, and returns the first value with a key match.
func readLabels(labels map[string]string, keys ...string) string {
	for _, key := range keys {
		if val, ok := labels[key]; ok {
			return val
		}
	}
	return ""
}
