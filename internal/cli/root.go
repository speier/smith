package cli

import (
	"fmt"
	"strings"

	"github.com/speier/smith/internal/repl"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "smith [prompt...]",
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
		// If args provided, start REPL with initial prompt
		var initialPrompt string
		if len(args) > 0 {
			initialPrompt = strings.Join(args, " ")
		}
		startREPL(initialPrompt)
	},
	Version:            "0.1.0",
	DisableFlagParsing: false,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	Args:               cobra.ArbitraryArgs,
	SilenceErrors:      true,
	SilenceUsage:       true,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Only exec command for non-interactive mode (like droid)
	rootCmd.AddCommand(execCmd)

	// Disable auto-generated commands
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}

func startREPL(initialPrompt string) {
	r, err := repl.New(".")
	if err != nil {
		fmt.Printf("Error creating REPL: %v\n", err)
		return
	}

	if err := r.Start(initialPrompt); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
