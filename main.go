package main

import (
	"os"

	"github.com/qj0r9j0vc2/kko/cmd"
)

var version = "dev"

func main() {
	cmd.SetVersion(version)
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
