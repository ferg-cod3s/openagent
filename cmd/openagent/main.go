// Package main implements the OpenAgent CLI.
package main

import (
	"os"

	"github.com/ferg-cod3s/openagent/cmd/openagent/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
