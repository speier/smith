package cli

import (
	"fmt"
	"time"

	"github.com/speier/smith/internal/coordinator"
	"github.com/spf13/cobra"
)

var (
	liveMode bool
	follow   bool
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current workflow status",
	Long:  `Display task status, active locks, and agent coordination state.`,
	Run: func(cmd *cobra.Command, args []string) {
		if liveMode {
			showLiveStatus()
		} else {
			showStatus()
		}
	},
}

func init() {
	statusCmd.Flags().BoolVarP(&liveMode, "live", "l", false, "Live updating dashboard")
	statusCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow agent output")
}

func showStatus() {
	coord := coordinator.New(projectPath)

	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║  📊 Smith Status                                           ║")
	fmt.Println("╠════════════════════════════════════════════════════════════╣")

	// Show tasks
	tasks, err := coord.GetTaskStats()
	if err == nil {
		fmt.Printf("║  Tasks:                                                    ║\n")
		fmt.Printf("║    ✅ Done: %-3d                                            ║\n", tasks.Done)
		fmt.Printf("║    🔄 In Progress: %-3d                                     ║\n", tasks.InProgress)
		fmt.Printf("║    ⏳ Available: %-3d                                       ║\n", tasks.Available)
		if tasks.Blocked > 0 {
			fmt.Printf("║    ⚠️  Blocked: %-3d                                        ║\n", tasks.Blocked)
		}
	}

	fmt.Println("║                                                            ║")

	// Show locks
	locks, err := coord.GetActiveLocks()
	if err == nil && len(locks) > 0 {
		fmt.Println("║  🔒 Active Work:                                           ║")
		for _, lock := range locks {
			fmt.Printf("║    • %s (task #%s)                        ║\n",
				truncate(lock.Agent, 20), lock.TaskID)
		}
	} else {
		fmt.Println("║  🔓 No agents currently working                            ║")
	}

	fmt.Println("║                                                            ║")

	// Show messages
	messages, err := coord.GetMessages()
	if err == nil && len(messages) > 0 {
		fmt.Println("║  💬 Recent Messages:                                       ║")
		for i, msg := range messages {
			if i >= 3 {
				break
			}
			fmt.Printf("║    %s → %s                           ║\n",
				truncate(msg.From, 15), truncate(msg.To, 15))
		}
	}

	fmt.Println("╚════════════════════════════════════════════════════════════╝")
}

func showLiveStatus() {
	fmt.Println("📊 Smith Live Dashboard (Press Ctrl+C to exit)")
	fmt.Println()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		// Clear screen (ANSI escape code)
		fmt.Print("\033[H\033[2J")

		showStatus()

		fmt.Printf("\n⏱️  Last updated: %s\n", time.Now().Format("15:04:05"))

		<-ticker.C
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
