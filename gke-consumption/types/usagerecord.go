// Package types defines interface for types used.
// Follows GKE Usage Metering data structure.
package types

import (
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
)

// Label stores both the key and value of a Kubernetes label.
type Label struct {
	Key   string `bigquery:"key"`
	Value string `bigquery:"value"`
}

// String returns the string representation of a `Label` object.
func (l *Label) String() string {
	return fmt.Sprintf("%s=%s", l.Key, l.Value)
}

// ToLabelSlice converts Kubernetes-style labels (a string-to-string map) into
// a slice of `Label` objects.
func ToLabelSlice(m map[string]string) []*Label {
	var labels []*Label
	for key, value := range m {
		labels = append(labels, &Label{
			Key:   key,
			Value: value,
		})
	}
	return labels
}

// Project stores information about a GCP project.
type Project struct {
	ID string `bigquery:"id"`
}

// String returns the string representation of a `Project` object.
func (p *Project) String() string {
	return fmt.Sprintf("{id=%q}", p.ID)
}

// UsageUnit defined the base unit in which a resource usage is measured.
type UsageUnit string

const (
	// UsageUnitBytes measures a resource in bytes (e.g., network traffic).
	UsageUnitBytes UsageUnit = "bytes"
	// UsageUnitByteSeconds measures a resource by bytes * seconds (e.g., RAM).
	UsageUnitByteSeconds UsageUnit = "byte-seconds"
	// UsageUnitSeconds measures a resource by seconds (e.g., CPU).
	UsageUnitSeconds UsageUnit = "seconds"
)

// Usage holds information about the usage of a cloud resource.
type Usage struct {
	// The quantity of unit used.
	Amount float64 `bigquery:"amount"`
	// The base unit in which resource usage is measured.
	Unit UsageUnit `bigquery:"unit"`
}

// UsageRecord contains fields to describe the usage of a cloud resource.
type UsageRecord struct {
	// The unique identifier for the resource of which the usage is consumed.
	// NOTE: this field should NOT be populated to BigQuery.
	ResourceID uint64 `bigquery:"-"`
	// The GCP region in which the resource resides.
	// NOTE: this field should NOT be populated to BigQuery.
	Region string `bigquery:"-"`

	// The name of the GCE zone or region in which the cluster resides.
	ClusterLocation string `bigquery:"cluster_location"`
	// The name of the Kubernetes cluster.
	ClusterName string `bigquery:"cluster_name"`

	// The Kubernetes namespace from which the usage is generated.
	Namespace string `bigquery:"namespace"`

	// The name of the resource, which maps to a key in Kubernetes' ResourceList.
	ResourceName v1.ResourceName `bigquery:"resource_name"`
	// The SKU ID of the underlying GCP cloud resource.
	SkuID string `bigquery:"sku_id"`

	// The UNIX timestamp of when the usage began.
	StartTime time.Time `bigquery:"start_time"`
	// The UNIX timestamp of when the usage ended.
	EndTime time.Time `bigquery:"end_time"`

	// The fraction of a cloud resource used by the namespace. For a dedicated
	// cloud resource that is solely used by a single namespace, the fraction is
	// 1.0. For resources shared among multiple namespaces, the fraction is
	// calculated as the requested amount divided by the total capacity of the
	// underlying resource.
	Fraction float64 `bigquery:"fraction"`
	// The size of underlying GCP resource (e.g. "1" n1-standard-1 VM instance,
	// a "30Gi" persistent disk).
	CloudResourceSize int64 `bigquery:"cloud_resource_size"`

	// The Kubernetes labels associated with the usage record.
	Labels []*Label `bigquery:"labels"`

	// The GCP project in which the cluster resides.
	Project *Project `bigquery:"project"`

	// The amount of absolute usage of this record.
	Usage *Usage `bigquery:"usage"`
}

// String returns the string representation of a `UsageRecord` object.
func (ur *UsageRecord) String() string {
	fields := []string{
		fmt.Sprintf("cluster_location=%q", ur.ClusterLocation),
		fmt.Sprintf("cluster_name=%q", ur.ClusterName),
		fmt.Sprintf("namespace=%q", ur.Namespace),
		fmt.Sprintf("resource_name=%q", ur.ResourceName),
		fmt.Sprintf("sku_id=%q", ur.SkuID),
		fmt.Sprintf("start_time=%v", ur.StartTime),
		fmt.Sprintf("end_time=%v", ur.EndTime),
		fmt.Sprintf("fraction=%f", ur.Fraction),
		fmt.Sprintf("cloud_resource_size=%d", ur.CloudResourceSize),
		fmt.Sprintf("labels=%v", ur.Labels),
		fmt.Sprintf("project=%v", ur.Project),
		fmt.Sprintf("usage=%v", ur.Usage),
	}
	return "{" + strings.Join(fields, ", ") + "}"
}
