package main

import (
	"os"

	"github.com/qj0r9j0vc2/kko/cmd"
	"github.com/qj0r9j0vc2/kko/internal/output"
)

var version = "dev"

func main() {
	cmd.SetVersion(version)
	if err := cmd.Execute(); err != nil {
		output.Errorf("%s", err)
		os.Exit(1)
	}
}
