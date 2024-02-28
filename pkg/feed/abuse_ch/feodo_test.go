package abuse_ch_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/drone/pkg/feed/abuse_ch"
	"github.com/m-mizutani/drone/pkg/infra"
	"github.com/m-mizutani/drone/pkg/infra/bq"
	"github.com/m-mizutani/drone/pkg/utils"
	"github.com/m-mizutani/gt"
)

func TestFeodo(t *testing.T) {
	mock := bq.NewMock()
	clients := infra.New(infra.WithBigQuery(mock))
	ctx := context.Background()

	// first time
	gt.NoError(t, abuse_ch.NewFeodo().Import(ctx, clients))

	gt.A(t, mock.InsertedData).Length(1)
	firstRecords := gt.Cast[[]abuse_ch.FeodoRecord](t, mock.InsertedData[0])
	gt.A(t, firstRecords).Longer(1)

	// second time
	gt.NoError(t, abuse_ch.NewFeodo().Import(ctx, clients))
	// The second import result should not have new data
	gt.A(t, mock.InsertedData).Length(1)
}

func TestFeodoIntegration(t *testing.T) {
	bqProjectID := utils.LookupEnv(t, "TEST_BIGQUERY_PROJECT_ID")
	bqDatasetID := utils.LookupEnv(t, "TEST_BIGQUERY_DATASET_ID")

	ctx := context.Background()
	bqClient := gt.R1(bq.New(ctx, bqProjectID, bqDatasetID)).NoError(t)
	clients := infra.New(
		infra.WithBigQuery(bqClient),
	)

	gt.NoError(t, abuse_ch.NewFeodo().Import(ctx, clients))
}
