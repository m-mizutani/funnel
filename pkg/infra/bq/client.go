package bq

import (
	"context"
	"errors"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/m-mizutani/drone/pkg/domain/interfaces"
	"github.com/m-mizutani/drone/pkg/domain/model"
	"github.com/m-mizutani/drone/pkg/utils"
	"github.com/m-mizutani/goerr"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type client struct {
	dataSet *bigquery.Dataset
	client  *bigquery.Client
}

const (
	importLogTable = "import_logs"
)

func New(ctx context.Context, projectID, datasetID string, options ...option.ClientOption) (interfaces.BigQuery, error) {
	c, err := bigquery.NewClient(ctx, projectID, options...)
	if err != nil {
		return nil, goerr.Wrap(err, "Fail to create BigQuery client")
	}

	dataSet := c.Dataset(datasetID)
	table := dataSet.Table(importLogTable)
	schema, err := bigquery.InferSchema(&model.ImportLog{})
	if err != nil {
		return nil, goerr.Wrap(err, "failed to infer schema of import_logs")
	}

	meta := &bigquery.TableMetadata{
		Schema: schema,
	}
	if err := table.Create(ctx, meta); err != nil {
		if gerr, ok := err.(*googleapi.Error); !ok || gerr.Code != 409 {
			return nil, goerr.Wrap(err, "failed to create table of import_logs")
		}
	}

	return &client{
		client:  c,
		dataSet: dataSet,
	}, nil
}

func (x *client) Migrate(ctx context.Context, tableName string, schema any) error {
	table := x.dataSet.Table(tableName)
	inferredSchema, err := bigquery.InferSchema(schema)
	if err != nil {
		return goerr.Wrap(err, "failed to infer schema").With("table", tableName)
	}

	meta := &bigquery.TableMetadata{
		Schema: inferredSchema,
	}
	if err := table.Create(ctx, meta); err != nil {
		if gerr, ok := err.(*googleapi.Error); !ok || gerr.Code != 409 {
			return goerr.Wrap(err, "failed to create table").With("table", tableName)
		}
	}

	return nil
}

func (x *client) Insert(ctx context.Context, tableName string, data any) error {
	if err := insertWithRetry(ctx, x.dataSet.Table(tableName), data); err != nil {
		return goerr.Wrap(err, "Fail to insert data").With("table", tableName)
	}

	return nil
}

func (x *client) PutImportLog(ctx context.Context, log *model.ImportLog) error {
	if err := insertWithRetry(ctx, x.dataSet.Table(importLogTable), log); err != nil {
		return goerr.Wrap(err, "Fail to insert import log").With("log", log)
	}
	return nil
}

func (x *client) GetLatestImportLog(ctx context.Context, tableName string) (*model.ImportLog, error) {
	query := x.client.Query(`
		SELECT timestamp FROM ` + x.dataSet.DatasetID + `.` + importLogTable + `
		WHERE table_name = @table_name
		ORDER BY timestamp DESC
		LIMIT 1
	`)
	query.Parameters = []bigquery.QueryParameter{
		{
			Name:  "table_name",
			Value: tableName,
		},
	}

	it, err := query.Read(ctx)
	if err != nil {
		return nil, goerr.Wrap(err, "Fail to query import_logs").With("table", tableName)
	}

	var log model.ImportLog
	if err := it.Next(&log); err != nil {
		if err == iterator.Done {
			return nil, nil
		}
		return nil, goerr.Wrap(err, "Fail to get latest time").With("table", tableName)
	}

	return &log, nil
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
