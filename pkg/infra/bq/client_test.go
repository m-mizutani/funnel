package bq_test

import (
	"context"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/m-mizutani/bqs"
	"github.com/m-mizutani/drone/pkg/infra/bq"
	"github.com/m-mizutani/gt"
)

func TestBigQuerySchemaUpdate(t *testing.T) {
	projectID, ok := os.LookupEnv("TEST_BIGQUERY_PROJECT_ID")
	if !ok {
		t.Skip("TEST_BIGQUERY_PROJECT_ID is not set")
	}

	datasetID, ok := os.LookupEnv("TEST_BIGQUERY_DATASET_ID")
	if !ok {
		t.Skip("TEST_BIGQUERY_DATASET_ID is not set")
	}

	ctx := context.Background()
	client := gt.R1(bq.New(ctx, projectID, datasetID)).NoError(t)

	bqClient := gt.R1(bigquery.NewClient(ctx, projectID)).NoError(t)

	tableID := time.Now().Format("test_table_20060102150405")
	table := bqClient.Dataset(datasetID).Table(tableID)
	gt.NoError(t, table.Create(ctx, &bigquery.TableMetadata{}))

	type TestData1 struct {
		Color string
	}
	type TestData2 struct {
		Number int
	}

	s1 := gt.R1(bqs.Infer(&TestData1{})).NoError(t)
	s2 := gt.R1(bqs.Infer(&TestData2{})).NoError(t)
	gt.NoError(t, client.CreateOrUpdateSchema(ctx, tableID, s1))
	gt.NoError(t, client.CreateOrUpdateSchema(ctx, tableID, s2))

	md := gt.R1(table.Metadata(ctx)).NoError(t)
	gt.A(t, md.Schema).Length(2).
		MatchThen(func(v *bigquery.FieldSchema) bool {
			return v.Name == "Color"
		}, func(t testing.TB, v *bigquery.FieldSchema) {
			gt.Equal(t, v.Type, bigquery.StringFieldType)
		}).
		MatchThen(func(v *bigquery.FieldSchema) bool {
			return v.Name == "Number"
		}, func(t testing.TB, v *bigquery.FieldSchema) {
			gt.Equal(t, v.Type, bigquery.NumericFieldType)
		})
}
