package cli

import (
	"github.com/m-mizutani/drone/pkg/cli/config"
	"github.com/m-mizutani/drone/pkg/feed/abuse_ch"
	"github.com/m-mizutani/drone/pkg/feed/otx"
	"github.com/m-mizutani/drone/pkg/infra"
	"github.com/m-mizutani/goerr"
	"github.com/urfave/cli/v2"
)

type importConfig struct {
	bq        config.BigQuery
	firestore config.Firestore
	sentry    config.Sentry
}

func subImport() *cli.Command {
	var cfg importConfig

	return &cli.Command{
		Name:    "import",
		Usage:   "Import feed data to BigQuery",
		Aliases: []string{"i"},
		Flags:   mergeFlags([]cli.Flag{}, &cfg.bq, &cfg.firestore, &cfg.sentry),
		Subcommands: []*cli.Command{
			subImportOtx(&cfg),
			subImportAbuseCh(&cfg),
		},
		Before: func(ctx *cli.Context) error {
			if err := cfg.sentry.Configure(); err != nil {
				return goerr.Wrap(err, "fail to configure sentry")
			}
			return nil
		},
	}
}

// -----------------------------------------
// OTX feed data import

type otxConfig struct {
	apiKey string `masq:"secret"`
}

// subImportOtx is a subcommand of "import" command
func subImportOtx(cfg *importConfig) *cli.Command {
	var otxCfg otxConfig
	return &cli.Command{
		Name:  "otx",
		Usage: "Import OTX feed data to BigQuery",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "otx-api-key",
				Usage:       "OTX API key",
				EnvVars:     []string{"DRONE_OTX_API_KEY"},
				Destination: &otxCfg.apiKey,
				Required:    true,
			},
		},
		Subcommands: []*cli.Command{
			subImportOtxSubscribed(cfg, &otxCfg),
		},
	}
}

func subImportOtxSubscribed(cfg *importConfig, otxCfg *otxConfig) *cli.Command {
	return &cli.Command{
		Name:    "subscribed",
		Aliases: []string{"s"},
		Usage:   "Import OTX subscribed feed data to BigQuery",
		Action: func(ctx *cli.Context) error {
			bqClient, err := cfg.bq.Configure(ctx.Context)
			if err != nil {
				return goerr.Wrap(err, "Fail to configure BigQuery")
			}
			dbClient, err := cfg.firestore.Configure(ctx.Context)
			if err != nil {
				return goerr.Wrap(err, "Fail to configure Firestore")
			}

			otxClient := otx.NewSubscribed(otxCfg.apiKey)
			clients := infra.New(
				infra.WithBigQuery(bqClient),
				infra.WithDatabase(dbClient),
			)
			if err := otxClient.Import(ctx.Context, clients); err != nil {
				return goerr.Wrap(err, "Fail to import OTX subscribed")
			}

			return nil
		},
	}
}

// -----------------------------------------
// Abuse.ch feed data import

func subImportAbuseCh(cfg *importConfig) *cli.Command {
	return &cli.Command{
		Name:  "abusech",
		Usage: "Import abuse.ch feed data to BigQuery",
		Subcommands: []*cli.Command{
			subImportAbuseChFeodo(cfg),
		},
	}
}

func subImportAbuseChFeodo(cfg *importConfig) *cli.Command {
	return &cli.Command{
		Name:  "feodo",
		Usage: "Import abuse.ch feodo feed data to BigQuery",
		Action: func(ctx *cli.Context) error {
			bqClient, err := cfg.bq.Configure(ctx.Context)
			if err != nil {
				return goerr.Wrap(err, "Fail to configure BigQuery")
			}
			dbClient, err := cfg.firestore.Configure(ctx.Context)
			if err != nil {
				return goerr.Wrap(err, "Fail to configure Firestore")
			}

			feed := abuse_ch.NewFeodo()
			clients := infra.New(
				infra.WithBigQuery(bqClient),
				infra.WithDatabase(dbClient),
			)
			if err := feed.Import(ctx.Context, clients); err != nil {
				return goerr.Wrap(err, "Fail to import OTX subscribed")
			}

			return nil
		},
	}
}
