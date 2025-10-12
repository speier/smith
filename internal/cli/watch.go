package cli

import (
	"fmt"
	"log"

	"github.com/speier/smith/internal/watcher"
	"github.com/spf13/cobra"
)

var (
	watchInterval int
	autoApprove   bool
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch TODO.md for changes and automatically orchestrate agents",
	Long: `Continuously monitors TODO.md for new tasks or feature requests.
When changes detected:
  1. Planning agent breaks down features into tasks
  2. Implementation agents claim and execute tasks
  3. Testing agents validate the code
  4. Review agents ensure quality
  
This is the fully autonomous mode - just add features to TODO.md and let it work.`,
	Run: func(cmd *cobra.Command, args []string) {
		runWatcher()
	},
}

func init() {
	watchCmd.Flags().StringVarP(&projectPath, "path", "p", ".", "Path to project")
	watchCmd.Flags().IntVarP(&watchInterval, "interval", "i", 5, "Check interval in seconds")
	watchCmd.Flags().BoolVarP(&autoApprove, "auto-approve", "a", false, "Auto-approve review agent decisions")
	watchCmd.Flags().IntVarP(&maxParallel, "parallel", "n", 2, "Max parallel agents")
}

func runWatcher() {
	fmt.Println("👁️  Smith Watch Mode Started")
	fmt.Printf("📂 Project: %s\n", projectPath)
	fmt.Printf("⏱️  Check interval: %ds\n", watchInterval)
	fmt.Printf("🔄 Max parallel: %d agents\n\n", maxParallel)

	if autoApprove {
		fmt.Println("⚠️  Auto-approve enabled - review decisions will be automatic")
	}

	w := watcher.New(watcher.Config{
		ProjectPath:   projectPath,
		CheckInterval: watchInterval,
		MaxParallel:   maxParallel,
		AutoApprove:   autoApprove,
	})

	if err := w.Start(); err != nil {
		log.Fatalf("Watcher failed: %v", err)
	}
}
