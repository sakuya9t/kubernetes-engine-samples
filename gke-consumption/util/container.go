package util

import (
	"consumptionexp/types"
	"context"
	"fmt"

	container "cloud.google.com/go/container/apiv1"
	proto "cloud.google.com/go/container/apiv1/containerpb"
)

// ContainerClient is a wrapper for container api client.
type ContainerClient struct {
	client *container.ClusterManagerClient
}

// NewContainerClient creates a new container client.
func NewContainerClient(ctx context.Context) (*ContainerClient, error) {
	c, err := container.NewClusterManagerClient(ctx)
	if err != nil {
		return nil, err
	}

	return &ContainerClient{
		client: c,
	}, nil
}

// GetCluster retrieves GKE cluster information.
func (c *ContainerClient) GetCluster(ctx context.Context, cluster *types.Cluster) (*proto.Cluster, error) {
	req := &proto.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", cluster.ProjectID, cluster.Location, cluster.Name),
	}

	return c.client.GetCluster(ctx, req)
}

// Close closes client connection.
func (c *ContainerClient) Close() {
	c.client.Close()
}
