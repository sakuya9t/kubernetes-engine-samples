package util

import (
	"context"
	"fmt"

	"consumptionexp/auth"
	"consumptionexp/store"
	"consumptionexp/types"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// K8sClient is a wrapper around the underlying `kubernetes.Clientset` object.
type K8sClient struct {
	cluster   *types.Cluster
	nodeStore *store.Nodes
	client    *kubernetes.Clientset
}

// NewK8sClient creates a new Client object of to a cluster.
func NewK8sClient(ctx context.Context, cluster *types.Cluster, nodeStore *store.Nodes) (*K8sClient, error) {
	otr, err := auth.OAuthTransport(ctx, cluster.Cluster)
	if err != nil {
		return nil, err
	}
	kubeclient, err := kubernetes.NewForConfig(&rest.Config{
		Host:      "https://" + cluster.Cluster.GetEndpoint(),
		Transport: otr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %v", err)
	}
	return &K8sClient{
		cluster:   cluster,
		nodeStore: nodeStore,
		client:    kubeclient,
	}, nil
}

// GetPod gets a pod with the given namespace and name.
func (c *K8sClient) GetPod(ctx context.Context, namespace, name string) (*v1.Pod, error) {
	pod, err := c.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting pod %s/%s: %v", namespace, name, err)
	}
	return pod, nil
}

// GetNode gets a node with the given name.
func (c *K8sClient) GetNode(ctx context.Context, name string) (*types.NodeCache, error) {
	node, err := c.client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error getting node %s: %v", name, err)
	}
	nc, err := types.FromV1Node(node, c.cluster)
	if err != nil {
		return nil, fmt.Errorf("error converting node %s: %v", name, err)
	}
	return nc, nil
}
