package otx_test

import (
	"context"
	"testing"

	"github.com/m-mizutani/drone/pkg/feed/otx"
	"github.com/m-mizutani/drone/pkg/infra"
	"github.com/m-mizutani/drone/pkg/infra/bq"
	"github.com/m-mizutani/drone/pkg/utils"
	"github.com/m-mizutani/gt"
)

func TestSubscribedIntegration(t *testing.T) {
	var (
		bqProjectID string
		bqDatasetID string
		apiKey      string
	)

	if err := utils.LoadEnv(
		utils.Env("TEST_OTX_API_KEY", &apiKey),
		utils.Env("TEST_BIGQUERY_PROJECT_ID", &bqProjectID),
		utils.Env("TEST_BIGQUERY_DATASET_ID", &bqDatasetID),
	); err != nil {
		t.Skipf("Skip test due to lack of env variables: %v", err)
	}

	ctx := context.Background()
	bqClient := gt.R1(bq.New(ctx, bqProjectID, bqDatasetID)).NoError(t)
	clients := infra.New(infra.WithBigQuery(bqClient))

	gt.NoError(t, otx.NewSubscribed(apiKey).Import(ctx, clients))
}
