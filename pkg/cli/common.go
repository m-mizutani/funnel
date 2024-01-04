package cli

import (
	"fmt"

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

func convertToBps(bytes float64) string {
	bytes = bytes * 8
	units := []string{"bps", "Kbps", "Mbps", "Gbps", "Tbps", "Pbps", "Ebps"}

	unitIndex := 0
	for bytes >= 1024 && unitIndex < len(units)-1 {
		bytes /= 1024
		unitIndex++
	}

	return fmt.Sprintf("%.2f %s", bytes, units[unitIndex])
}
