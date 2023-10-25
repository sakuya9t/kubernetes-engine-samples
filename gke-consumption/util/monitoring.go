// Package util contains utility of the application.
package util

import (
	"context"
	"fmt"
	"sort"
	"time"

	"consumptionexp/config"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"

	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MonitoringClient connects with cloud monitoring service
type MonitoringClient struct {
	client *monitoring.QueryClient
}

// NewMonitoringClient creates a new instance of the monitoring client
func NewMonitoringClient(ctx context.Context) (*MonitoringClient, error) {
	client, err := monitoring.NewQueryClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("NewQueryClient: %w", err)
	}

	return &MonitoringClient{
		client: client,
	}, nil
}

func (c *MonitoringClient) QueryTimeSeries(ctx context.Context, cfg *config.Configurations, q string) ([]*monitoringpb.TimeSeriesData, error) {
	data := []*monitoringpb.TimeSeriesData{}

	req := &monitoringpb.QueryTimeSeriesRequest{
		Name:  fmt.Sprintf("projects/%s", cfg.ProjectID),
		Query: q,
	}

	it := c.client.QueryTimeSeries(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return data, fmt.Errorf("could not query time series value: %w", err)
		}
		sort.Slice(resp.PointData, func(i, j int) bool {
			t1, t2 := resp.PointData[i].GetTimeInterval().GetStartTime().AsTime(), resp.PointData[j].GetTimeInterval().GetStartTime().AsTime()
			return t1.Before(t2)
		})
		for _, pd := range resp.PointData {
			start, end := pd.GetTimeInterval().GetStartTime().AsTime(), pd.GetTimeInterval().GetEndTime().AsTime()
			if start.Equal(end) {
				duration, _ := time.ParseDuration(config.DefaultResolution)
				pd.TimeInterval.EndTime = timestamppb.New(end.Add(duration))
			}
		}
		data = append(data, resp)
	}
	return data, nil
}

// Close turns off the monitoring client
func (c *MonitoringClient) Close() {
	c.client.Close()
}
