package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for kko.

To load completions:

Bash:
  $ source <(kko completion bash)
  # To load on startup, add to ~/.bashrc:
  # echo 'source <(kko completion bash)' >> ~/.bashrc

Zsh:
  $ source <(kko completion zsh)
  # To load on startup:
  # kko completion zsh > "${fpath[1]}/_kko"

Fish:
  $ kko completion fish | source
  # To load on startup:
  # kko completion fish > ~/.config/fish/completions/kko.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
