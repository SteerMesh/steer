package main

import (
	"os"

	"github.com/SteerMesh/steer/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(cli.ExitCode(err))
	}
}
