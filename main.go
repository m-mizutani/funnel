package main

import (
	"os"

	"github.com/m-mizutani/drone/pkg/cli"
)

func main() {
	if cli.Run(os.Args) != nil {
		os.Exit(1)
	}
}
