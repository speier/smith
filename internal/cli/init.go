package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize a new Smith project",
	Long:  `Create AGENTS.md, TODO.md, COMMS.md, and README.md in the target directory.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		initProject(path)
	},
}

func initProject(path string) {
	fmt.Printf("🚀 Initializing Smith project in %s\n", path)

	// Check if already initialized
	if _, err := os.Stat(filepath.Join(path, "AGENTS.md")); err == nil {
		fmt.Println("⚠️  Project already initialized (AGENTS.md exists)")
		return
	}

	// Create directory if needed
	if err := os.MkdirAll(path, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	files := map[string]string{
		"AGENTS.md": "# Agent Roles & Responsibilities\n\n(See existing AGENTS.md for template)",
		"TODO.md":   "# Task Board\n\n## Active Sprint\n\n*No tasks yet*",
		"COMMS.md":  "# Agent Communication & Coordination\n\n*No active work*",
		"README.md": "# Smith Project\n\nCoordinated multi-agent development.",
	}

	for filename, content := range files {
		fullPath := filepath.Join(path, filename)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			fmt.Printf("Error creating %s: %v\n", filename, err)
			continue
		}
		fmt.Printf("  ✅ Created %s\n", filename)
	}

	fmt.Println("\n🎉 Project initialized! Next steps:")
	fmt.Println("  1. Set ANTHROPIC_API_KEY environment variable")
	fmt.Println("  2. Add tasks to TODO.md")
	fmt.Println("  3. Run: smith orchestrate")
}
