package util

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"consumptionexp/config"
	"consumptionexp/types"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/googleapi"
)

var (
	requestTimeout = 10 * time.Second
)

// BigQueryClient is a wrapper for bigquery api client.
type BigQueryClient struct {
	cfg    *config.Configurations
	client *bigquery.Client
}

// NewBigQueryClient creates a new instance of the bigquery client
func NewBigQueryClient(ctx context.Context, cfg *config.Configurations) (*BigQueryClient, error) {
	client, err := bigquery.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("bigquery.NewClient: %v", err)
	}
	return &BigQueryClient{
		cfg:    cfg,
		client: client,
	}, nil
}

// Close turns off the bigquery client
func (c *BigQueryClient) Close() {
	c.client.Close()
}

// WriteUsage writes usage records to the bigquery table.
func (c *BigQueryClient) WriteUsage(ctx context.Context, records []types.UsageRecord) error {
	table, err := getOrCreateTable(ctx, c.client.Dataset(c.cfg.BigQuery.DataSetName), c.cfg.BigQuery.ConsumptionTableName)
	if err != nil {
		return fmt.Errorf("error getting bigquery table: %v", err)
	}

	inserter := table.Inserter()
	if err := inserter.Put(ctx, records); err != nil {
		return fmt.Errorf("error in writting bigquery table: %v", err)
	}

	return nil
}

func getOrCreateTable(ctx context.Context, dataset *bigquery.Dataset, tableID string) (*bigquery.Table, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	table := dataset.Table(tableID)
	if _, err := table.Metadata(ctx); err != nil {
		if isNotFound(err) {
			schema, err := generateSchema()
			if err != nil {
				return nil, err
			}

			metadata := &bigquery.TableMetadata{
				Schema:           schema,
				TimePartitioning: new(bigquery.TimePartitioning),
			}
			if err := table.Create(ctx, metadata); err != nil {
				err := fmt.Errorf("failed to create table %q in dataset %q: %v", tableID, dataset.DatasetID, err)
				return nil, err
			}
		} else {
			err := fmt.Errorf("failed to get the metadata for table %q in dataset %q: %v", tableID, dataset.DatasetID, err)
			return nil, err
		}
	}
	return table, nil
}

func isNotFound(err error) bool {
	if apiErr, ok := err.(*googleapi.Error); ok {
		return apiErr.Code == http.StatusNotFound
	}
	return false
}

func generateSchema() (bigquery.Schema, error) {
	schema, err := bigquery.InferSchema(types.UsageRecord{})
	if err != nil {
		err = fmt.Errorf("cannot infer the BigQuery Schema for UsageRecord: %v", err)
		return nil, err
	}

	setFieldOptional(schema)
	return schema, nil
}

func setFieldOptional(schema bigquery.Schema) {
	for _, fieldSchema := range schema {
		fieldSchema.Required = false
		setFieldOptional(fieldSchema.Schema)
	}
}
