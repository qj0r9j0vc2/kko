package cmd

import (
	"fmt"

	"github.com/qj0r9j0vc2/kko/internal/config"
	"github.com/qj0r9j0vc2/kko/internal/output"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage kko configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Example: `  kko config set api_key "abc123..."
  kko config set default_location.name "hakdong station"
  kko config set aliases.office "euljiro 3-ga station"`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show config file path",
	RunE:  runConfigPath,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configPathCmd)
}

func runConfigSet(_ *cobra.Command, args []string) error {
	key, value := args[0], args[1]
	if err := config.Set(key, value); err != nil {
		return err
	}

	if cfg.Output.Format == "json" {
		return output.PrintJSON(map[string]string{"key": key, "value": value, "status": "set"})
	}

	fmt.Printf("  %s %s = %s\n", output.Success("✓"), output.Label(key), value)
	return nil
}

func runConfigGet(_ *cobra.Command, args []string) error {
	key := args[0]
	value := config.Get(key)

	if cfg.Output.Format == "json" {
		return output.PrintJSON(map[string]interface{}{"key": key, "value": value})
	}

	if value == nil {
		fmt.Printf("  %s %s\n", output.Label(key+":"), output.Muted("(not set)"))
	} else {
		fmt.Printf("  %s %v\n", output.Label(key+":"), value)
	}
	return nil
}

func runConfigPath(_ *cobra.Command, _ []string) error {
	path := config.DefaultConfigPath()
	if cfg.Output.Format == "json" {
		return output.PrintJSON(map[string]string{"path": path})
	}
	fmt.Println(path)
	return nil
}
