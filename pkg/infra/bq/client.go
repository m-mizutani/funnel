package bq

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/m-mizutani/bqs"
	"github.com/m-mizutani/drone/pkg/domain/interfaces"
	"github.com/m-mizutani/drone/pkg/utils"
	"github.com/m-mizutani/goerr"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type client struct {
	dataSet *bigquery.Dataset
	client  *bigquery.Client
}

func New(ctx context.Context, projectID, datasetID string, options ...option.ClientOption) (interfaces.BigQuery, error) {
	c, err := bigquery.NewClient(ctx, projectID, options...)
	if err != nil {
		return nil, goerr.Wrap(err, "Fail to create BigQuery client")
	}

	dataSet := c.Dataset(datasetID)

	return &client{
		client:  c,
		dataSet: dataSet,
	}, nil
}

func (x *client) CreateOrUpdateSchema(ctx context.Context, tableName string, schema bigquery.Schema) error {
	table := x.dataSet.Table(tableName)
	md, err := table.Metadata(ctx)
	if err != nil {
		if gerr, ok := err.(*googleapi.Error); !ok || gerr.Code != 404 {
			return goerr.Wrap(err, "failed to get metadata").With("table", tableName)
		}

		meta := &bigquery.TableMetadata{
			Schema: schema,
		}
		if err := table.Create(ctx, meta); err != nil {
			if gerr, ok := err.(*googleapi.Error); !ok || gerr.Code != 409 {
				return goerr.Wrap(err, "failed to create table").With("table", tableName)
			}
		}

		return nil
	}

	merged, err := bqs.Merge(md.Schema, schema)
	if err != nil {
		return goerr.Wrap(err, "failed to merge schema").With("table", tableName)
	}

	meta := bigquery.TableMetadataToUpdate{
		Schema: merged,
	}
	if _, err := table.Update(ctx, meta, md.ETag); err != nil {
		return goerr.Wrap(err, "failed to update schema").With("table", tableName)
	}

	return nil
}

func (x *client) Insert(ctx context.Context, tableName string, data any) error {
	if err := insertWithRetry(ctx, x.dataSet.Table(tableName), data); err != nil {
		return goerr.Wrap(err, "Fail to insert data").With("table", tableName)
	}

	return nil
}

func insertWithRetry(ctx context.Context, table *bigquery.Table, data any) error {
	// Define the maximum number of retries and the initial delay.
	const maxRetries = 12
	initialDelay := time.Millisecond * 100

	// Attempt to insert data with exponential backoff retry logic.
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Wait for the delay period before retrying.
			time.Sleep(initialDelay)
			// Increase the delay for the next retry.
			initialDelay *= 2
		}

		inserter := table.Inserter()
		err := inserter.Put(ctx, data)
		if err == nil {
			// Data inserted successfully, no need to retry.
			return nil
		}

		if e, ok := err.(*googleapi.Error); ok && e.Code == 404 {
			utils.Logger().Warn("Table not found, retrying", "table", table.FullyQualifiedName())
			continue
		}

		// If the error is not a 404, return it immediately without retrying.
		return err
	}

	// Data insertion failed after all retries.
	return errors.New("insert failed: exceeded retry limit")
}
