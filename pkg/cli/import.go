package cli

import (
	"github.com/m-mizutani/goerr"
	"github.com/m-mizutani/pacman/pkg/cli/config"
	"github.com/m-mizutani/pacman/pkg/feed/abusech"
	"github.com/m-mizutani/pacman/pkg/feed/otx"
	"github.com/urfave/cli/v2"
)

type importConfig struct {
	bq config.BigQuery
}

func subImport() *cli.Command {
	var cfg importConfig

	return &cli.Command{
		Name:    "import",
		Usage:   "Import feed data to BigQuery",
		Aliases: []string{"i"},
		Flags:   mergeFlags([]cli.Flag{}, &cfg.bq),
		Subcommands: []*cli.Command{
			subImportOtx(&cfg),
			subImportAbuseCh(&cfg),
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
				EnvVars:     []string{"PACMAN_OTX_API_KEY"},
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

			otxClient := otx.NewSubscribed(otxCfg.apiKey)
			if err := otxClient.Import(ctx.Context, bqClient); err != nil {
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

			feed := abusech.NewFeodo()
			if err := feed.Import(ctx.Context, bqClient); err != nil {
				return goerr.Wrap(err, "Fail to import OTX subscribed")
			}

			return nil
		},
	}
}
