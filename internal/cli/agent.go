package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/speier/smith/internal/agent"
	"github.com/spf13/cobra"
)

var (
	agentRole   string
	agentTaskID string
	apiKey      string
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Run a single agent (usually spawned by orchestrator)",
	Long: `Execute a single agent instance with a specific role.
This is typically called by the orchestrator, not manually.`,
	Run: func(cmd *cobra.Command, args []string) {
		runAgent()
	},
}

func init() {
	agentCmd.Flags().StringVarP(&agentRole, "role", "r", "", "Agent role (planning|implementation|testing|review)")
	agentCmd.Flags().StringVarP(&agentTaskID, "task", "t", "", "Task ID from TODO.md")
	agentCmd.Flags().StringVarP(&apiKey, "api-key", "k", "", "LLM API key (or use ANTHROPIC_API_KEY env)")

	agentCmd.MarkFlagRequired("role")
	agentCmd.MarkFlagRequired("task")
}

func runAgent() {
	// Get API key from flag or env
	key := apiKey
	if key == "" {
		key = os.Getenv("ANTHROPIC_API_KEY")
	}
	if key == "" {
		log.Fatal("API key required: --api-key flag or ANTHROPIC_API_KEY env var")
	}

	fmt.Printf("🤖 Starting %s agent for task %s\n", agentRole, agentTaskID)

	a := agent.New(agent.Config{
		Role:        agentRole,
		TaskID:      agentTaskID,
		ProjectPath: projectPath,
		APIKey:      key,
	})

	if err := a.Execute(); err != nil {
		log.Fatalf("Agent failed: %v", err)
	}

	fmt.Println("✅ Agent completed task")
}
