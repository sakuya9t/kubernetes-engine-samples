package types

// Labels defines the label values data structure extracted from the metrics.
type Labels struct {
	ProjectID   string
	Location    string
	ClusterName string
	Namespace   string
	PodName     string
	Container   string
}
