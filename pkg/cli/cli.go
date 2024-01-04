package cli

import (
	"github.com/m-mizutani/pacman/pkg/cli/config"
	"github.com/m-mizutani/pacman/pkg/domain/types"
	"github.com/m-mizutani/pacman/pkg/utils"
	"github.com/urfave/cli/v2"
)

func Run(args []string) error {
	var (
		logger config.Logger

		logCloser func()
	)

	app := cli.App{
		Name:    "pacman",
		Flags:   mergeFlags([]cli.Flag{}, &logger),
		Version: types.AppVersion,
		Commands: []*cli.Command{
			subImport(),
		},
		Before: func(ctx *cli.Context) error {
			f, err := logger.Configure()
			if err != nil {
				return err
			}
			logCloser = f
			return nil
		},
		After: func(ctx *cli.Context) error {
			if logCloser != nil {
				logCloser()
			}
			return nil
		},
	}

	if err := app.Run(args); err != nil {
		utils.Logger().Error("Failed to run pacman", utils.ErrLog(err))
		return err
	}

	return nil
}
