package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/speier/smith/internal/engine"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec [prompt]",
	Short: "Execute a single command (non-interactive mode)",
	Long: `Execute a single command in non-interactive mode.

This is useful for:
- Running from scripts or automation
- Single-shot commands without entering REPL
- Piping prompts from files or other commands

Examples:
  smith exec "analyze this file"
  smith exec - < prompt.txt
  echo "review the API" | smith exec -`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var prompt string

		// Read from stdin if '-' or no args
		if len(args) == 0 || (len(args) == 1 && args[0] == "-") {
			stat, err := os.Stdin.Stat()
			if err != nil {
				return fmt.Errorf("checking stdin: %w", err)
			}

			if (stat.Mode() & os.ModeCharDevice) == 0 {
				// Reading from pipe/file
				reader := bufio.NewReader(os.Stdin)
				content, err := io.ReadAll(reader)
				if err != nil {
					return fmt.Errorf("reading from stdin: %w", err)
				}
				prompt = strings.TrimSpace(string(content))
			} else {
				return fmt.Errorf("no prompt provided (use: smith exec \"prompt\" or pipe to stdin)")
			}
		} else {
			prompt = strings.Join(args, " ")
		}

		if prompt == "" {
			return fmt.Errorf("empty prompt")
		}

		// Create engine
		eng, err := engine.New(engine.Config{
			ProjectPath: ".",
		})
		if err != nil {
			return fmt.Errorf("creating engine: %w", err)
		}

		// Execute single command
		fmt.Printf("💬 Prompt: %s\n\n", prompt)
		response, err := eng.Chat(prompt)
		if err != nil {
			return fmt.Errorf("execution failed: %w", err)
		}

		fmt.Printf("🕶️ Response:\n%s\n", response)
		return nil
	},
}
