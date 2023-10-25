package types

import (
	"fmt"

	proto "cloud.google.com/go/container/apiv1/containerpb"
)

// Cluster is the definition of a cluster object.
type Cluster struct {
	ProjectID string
	Name      string
	Location  string
	Cluster   *proto.Cluster
}

func (c *Cluster) String() string {
	return fmt.Sprintf("%s-%s-%s", c.ProjectID, c.Location, c.Name)
}
