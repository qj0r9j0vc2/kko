package cmd

import (
	"fmt"

	"github.com/qj0r9j0vc2/kko/internal/config"
	"github.com/qj0r9j0vc2/kko/internal/kakao"
	"github.com/qj0r9j0vc2/kko/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile    string
	jsonOutput bool
	noColor    bool
	verbose    bool
	appVersion string

	cfg    *config.Config
	client *kakao.Client
)

func SetVersion(v string) {
	appVersion = v
}

var rootCmd = &cobra.Command{
	Use:   "kko",
	Short: "Kakao Unified CLI",
	Long:  "A fast, composable CLI for Kakao APIs — search places, get directions, send messages, manage calendar.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if jsonOutput {
			cfg.Output.Format = "json"
		}
		if noColor || !output.IsTerminal() {
			cfg.Output.Color = false
		}

		client = kakao.NewClient(cfg)
		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.config/kko/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	viper.SetDefault("output.format", "table")
	viper.SetDefault("output.color", true)
	viper.SetDefault("output.lang", "ko")
	viper.SetDefault("search.default_engine", "web")
	viper.SetDefault("search.max_results", 5)
}
