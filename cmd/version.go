package cmd

import (
	"fmt"

	"github.com/qj0r9j0vc2/kko/internal/output"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print kko version",
	RunE:  runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(_ *cobra.Command, _ []string) error {
	if cfg != nil && cfg.Output.Format == "json" {
		return output.PrintJSON(map[string]string{"version": appVersion})
	}
	fmt.Printf("kko version %s\n", appVersion)
	return nil
}
