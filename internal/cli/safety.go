package cli

import (
	"fmt"

	"github.com/speier/smith/internal/safety"
	"github.com/spf13/cobra"
)

var safetyCmd = &cobra.Command{
	Use:   "safety",
	Short: "Test safety system",
	Long:  `Test the command safety checking system with different auto-levels.`,
}

var safetyCheckCmd = &cobra.Command{
	Use:   "check <command>",
	Short: "Check if a command is allowed",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		command := args[0]
		level, _ := cmd.Flags().GetString("level")

		fmt.Printf("Command: %s\n", command)
		fmt.Printf("Level:   %s (%s)\n\n", level, safety.GetLevelDescription(level))

		result := safety.IsCommandAllowed(command, level)
		fmt.Println(safety.FormatCheckResult(result))

		return nil
	},
}

var safetyInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show safety system information",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🛡️  Smith Safety System")
		fmt.Println()
		fmt.Printf("Rules Version: %s\n", safety.GetVersion())
		fmt.Println()
		fmt.Println("Auto-Levels:")
		fmt.Printf("  low    - %s\n", safety.GetLevelDescription("low"))
		fmt.Printf("  medium - %s\n", safety.GetLevelDescription("medium"))
		fmt.Printf("  high   - %s\n", safety.GetLevelDescription("high"))
		fmt.Println()

		allowlist := safety.GetSessionAllowlist()
		if len(allowlist) > 0 {
			fmt.Println("Session Allowlist:")
			for _, cmd := range allowlist {
				fmt.Printf("  • %s\n", cmd)
			}
		} else {
			fmt.Println("Session Allowlist: empty")
		}

		return nil
	},
}

// Note: Safety functionality moved to sensible defaults
// Auto-level can be changed via /level in REPL
