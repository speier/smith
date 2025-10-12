package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var inboxCmd = &cobra.Command{
	Use:   "inbox",
	Short: "Check messages and clarification requests from agents",
	Long:  `View and respond to questions from agents that need your input.`,
	Run: func(cmd *cobra.Command, args []string) {
		showInbox()
	},
}

func showInbox() {
	conversationPath := filepath.Join(projectPath, "CONVERSATION.md")

	content, err := os.ReadFile(conversationPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("📭 Inbox is empty - no messages from agents")
			return
		}
		fmt.Printf("❌ Error reading inbox: %v\n", err)
		return
	}

	// Simple display for now - just show the file
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  📬 Smith Inbox                                            ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")
	fmt.Println(string(content))
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	// TODO: Parse CONVERSATION.md for unanswered questions
	// TODO: Allow interactive responses
}
