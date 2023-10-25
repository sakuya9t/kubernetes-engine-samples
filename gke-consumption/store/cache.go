// Package store defines DAO layer logic for a local cache db.
package store

import (
	"consumptionexp/types"
	"database/sql"
	"fmt"
	"sync"
	"time"

	// import go-sqlite3 for sqlite3 driver.
	_ "github.com/mattn/go-sqlite3"
)

const (
	cacheDBFile      = "../cache.db"
	defaultCacheLife = time.Minute * 5
)

// Nodes is the Nodes table in the cache.
type Nodes struct {
	mu sync.Mutex
	db *sql.DB
}

// NewNodeStore creates a new instance to the Nodes table in the cache.
func NewNodeStore() (*Nodes, error) {
	db, err := sql.Open("sqlite3", cacheDBFile)
	if err != nil {
		return nil, fmt.Errorf("error opening local sqlite database: %v", err)
	}
	return &Nodes{
		db: db,
	}, nil
}

// GetNodeCache retrieves from Nodes table.
func (s *Nodes) GetNodeCache(projectID, clusterName, clusterLocation, nodeName string) (*types.NodeCache, error) {
	return s.getNodeCacheByID(getNodeCacheID(projectID, clusterName, clusterLocation, nodeName))
}

func (s *Nodes) getNodeCacheByID(id string) (*types.NodeCache, error) {
	row := s.db.QueryRow("SELECT projectID, clusterName, clusterLocation, nodeName, machineType, preemptible, region, cpuSize, memSize, lastUpdate FROM nodes WHERE id=?", id)
	node := types.NodeCache{}

	var err error
	if err = row.Scan(&node.ProjectID, &node.ClusterName, &node.ClusterLocation, &node.NodeName, &node.MachineType, &node.Preemptible, &node.Region, &node.CPUSize, &node.MemSize, &node.LastUpdate); err == sql.ErrNoRows {
		return &types.NodeCache{}, fmt.Errorf("row not found, id=%s, error=%v", id, err)
	}

	return &node, nil
}

// UpsertNodeCache upserts a NodeCache record into the table.
func (s *Nodes) UpsertNodeCache(node *types.NodeCache) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := getNodeCacheID(node.ProjectID, node.ClusterName, node.ClusterLocation, node.NodeName)
	query := fmt.Sprintf("INSERT INTO nodes(id, projectID, clusterName, clusterLocation, nodeName, machineType, preemptible, region, cpuSize, memSize, lastUpdate) VALUES(%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)"+
		"ON CONFLICT(id) DO UPDATE SET machineType = %s, preemptible = %s, region = %s, cpuSize = %s, memSize = %s, lastUpdate = %s;",
		id, node.ProjectID, node.ClusterName, node.ClusterLocation, node.NodeName, node.MachineType, node.Preemptible, node.Region, node.CPUSize, node.MemSize, node.LastUpdate,
		node.MachineType, node.Preemptible, node.Region, node.CPUSize, node.MemSize, node.LastUpdate)

	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}

	return nil
}

func getNodeCacheID(projectID, clusterName, clusterLocation, nodeName string) string {
	return fmt.Sprintf("%s-%s-%s-%s", projectID, clusterName, clusterLocation, nodeName)
}
