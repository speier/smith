package cli

import (
	"fmt"
	"log"

	"github.com/speier/smith/internal/orchestrator"
	"github.com/spf13/cobra"
)

var (
	projectPath string
	maxParallel int
	dryRun      bool
)

var orchestrateCmd = &cobra.Command{
	Use:   "orchestrate",
	Short: "Start the orchestrator to coordinate agents",
	Long: `The orchestrator reads TODO.md, spawns appropriate agents,
monitors coordination via COMMS.md, and manages the workflow.`,
	Run: func(cmd *cobra.Command, args []string) {
		runOrchestrator()
	},
}

func init() {
	orchestrateCmd.Flags().StringVarP(&projectPath, "path", "p", ".", "Path to project with coordination files")
	orchestrateCmd.Flags().IntVarP(&maxParallel, "parallel", "n", 2, "Max parallel agents")
	orchestrateCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", false, "Simulate without executing")
}

func runOrchestrator() {
	fmt.Printf("🎯 Starting orchestrator (max %d parallel agents)\n", maxParallel)
	fmt.Printf("📂 Project: %s\n\n", projectPath)

	orc := orchestrator.New(orchestrator.Config{
		ProjectPath: projectPath,
		MaxParallel: maxParallel,
		DryRun:      dryRun,
	})

	if err := orc.Run(); err != nil {
		log.Fatalf("Orchestrator failed: %v", err)
	}

	fmt.Println("\n✅ Orchestrator completed")
}
