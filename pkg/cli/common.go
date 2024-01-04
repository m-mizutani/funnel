package cli

import (
	"github.com/urfave/cli/v2"
)

type flagConfig interface {
	Flags() []cli.Flag
}

func mergeFlags(base []cli.Flag, configs ...flagConfig) []cli.Flag {
	ret := base[:]
	for _, config := range configs {
		ret = append(ret, config.Flags()...)
	}

	return ret
}
