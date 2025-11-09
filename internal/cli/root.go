package cli

import (
	"fmt"
	"os"

	"github.com/speier/smith/internal/frontend"
	"github.com/speier/smith/internal/session"
	"github.com/speier/smith/internal/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "smith [command]",
	Short: "The Agent Replication System",
	Long: `Smith - Inevitable. Multiplying. Building.

Mr. Anderson... Welcome back.

An agent system that duplicates itself to plan, implement, test, and review.

The Architect (planning) designs the structure.
The Keymaker (implementation) makes things work.
Sentinels (testing) hunt down bugs relentlessly.
The Oracle (review) sees quality and predicts issues.

Just chat naturally and watch the agents multiply to build your software.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create session
		sess := session.NewMockSession()

		// Get terminal size for initial UI
		width, height := 100, 40 // Default size, will auto-detect

		// Create and run chat UI
		ui := frontend.NewChatUI(sess, width, height)
		if err := ui.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running chat UI: %v\n", err)
			os.Exit(1)
		}
	},
	DisableFlagParsing: false,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	Args:               cobra.ArbitraryArgs,
	SilenceErrors:      true,
	SilenceUsage:       true,
}

// SetVersion sets the version for the CLI
func SetVersion(v string) {
	rootCmd.Version = v
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Version from internal/version
	rootCmd.Version = version.Get()

	// Customize usage template to show single usage line
	rootCmd.SetUsageTemplate(`Usage:
  {{.CommandPath}} [command]

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
Use "{{.CommandPath}} [command] --help" for more information about a command.
`)

	// Add subcommands
	rootCmd.AddCommand(execCmd)

	// Disable auto-generated commands
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}
