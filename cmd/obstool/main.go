package main

import (
	"os"

	"github.com/observability-ui/development-tools/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
