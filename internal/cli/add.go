package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	addPriority string
	addTags     []string
	interactive bool
)

var addCmd = &cobra.Command{
	Use:   "add [description]",
	Short: "Add a feature request or task to the backlog",
	Long: `Quickly add a feature request, bug fix, or task to TODO.md.
The planning agent will later enrich it with technical details.

Examples:
  smith add "Add user authentication"
  smith add "Fix slow API response" --priority high
  smith add "Refactor auth service" --tags @refactor,@backend
  smith prompt  # Interactive mode`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		description := strings.Join(args, " ")
		addToBacklog(description)
	},
}

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Interactive mode to add features with clarifying questions",
	Long:  `Start an interactive conversation to add features to the backlog.`,
	Run: func(cmd *cobra.Command, args []string) {
		interactivePrompt()
	},
}

func init() {
	addCmd.Flags().StringVarP(&addPriority, "priority", "P", "medium", "Priority: low, medium, high, critical")
	addCmd.Flags().StringSliceVarP(&addTags, "tags", "t", []string{}, "Tags (e.g., @backend,@security)")
}

func addToBacklog(description string) {
	fmt.Printf("📝 Adding to backlog: %s\n", description)

	// Read current TODO.md
	todoPath := "TODO.md"
	content, err := os.ReadFile(todoPath)
	if err != nil {
		fmt.Printf("❌ Failed to read TODO.md: %v\n", err)
		return
	}

	// Find or create "Feature Requests" section
	todoStr := string(content)

	featureRequest := fmt.Sprintf(`
### Feature Request - %s
**Priority:** %s
**Tags:** %s
**Status:** Needs Planning
**Created:** %s

**Description:**
%s

**Notes:**
- This will be enhanced by the planning agent
- Check back for structured task breakdown

---
`,
		description,
		addPriority,
		strings.Join(addTags, ", "),
		"2025-10-11", // TODO: Use actual timestamp
		description,
	)

	// Insert after "## Feature Requests" or at the top
	if strings.Contains(todoStr, "## Feature Requests") {
		parts := strings.SplitN(todoStr, "## Feature Requests", 2)
		if len(parts) == 2 {
			todoStr = parts[0] + "## Feature Requests\n" + featureRequest + parts[1]
		}
	} else {
		todoStr = "## Feature Requests\n" + featureRequest + "\n" + todoStr
	}

	// Write back
	if err := os.WriteFile(todoPath, []byte(todoStr), 0644); err != nil {
		fmt.Printf("❌ Failed to write TODO.md: %v\n", err)
		return
	}

	fmt.Println("✅ Added to backlog!")
	fmt.Println("\n💡 Next steps:")
	fmt.Println("   - Planning agent will analyze this during next watch cycle")
	fmt.Println("   - Or run: smith orchestrate --planning-only")
	fmt.Println("   - Check status: smith status")
}

func interactivePrompt() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🤖 Smith Interactive Prompt")
	fmt.Println("===============================")
	fmt.Println()

	// Get description
	fmt.Print("What would you like to build? ")
	description, _ := reader.ReadString('\n')
	description = strings.TrimSpace(description)

	if description == "" {
		fmt.Println("❌ Description cannot be empty")
		return
	}

	// Ask follow-up questions
	fmt.Print("\nAny specific requirements or constraints? (optional) ")
	requirements, _ := reader.ReadString('\n')
	requirements = strings.TrimSpace(requirements)

	// Priority
	fmt.Print("\nPriority? [low/medium/high/critical] (default: medium) ")
	priority, _ := reader.ReadString('\n')
	priority = strings.TrimSpace(priority)
	if priority == "" {
		priority = "medium"
	}

	// Tags
	fmt.Print("\nTags? (comma-separated, e.g., @backend,@security) (optional) ")
	tagsInput, _ := reader.ReadString('\n')
	tagsInput = strings.TrimSpace(tagsInput)

	// Build enhanced description
	fullDescription := description
	if requirements != "" {
		fullDescription += "\n\nRequirements:\n" + requirements
	}

	// Set flags
	addPriority = priority
	if tagsInput != "" {
		addTags = strings.Split(tagsInput, ",")
		for i := range addTags {
			addTags[i] = strings.TrimSpace(addTags[i])
		}
	}

	fmt.Println("\n📋 Summary:")
	fmt.Printf("   Description: %s\n", description)
	fmt.Printf("   Priority: %s\n", priority)
	if len(addTags) > 0 {
		fmt.Printf("   Tags: %s\n", strings.Join(addTags, ", "))
	}

	fmt.Print("\nAdd to backlog? [Y/n] ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.ToLower(strings.TrimSpace(confirm))

	if confirm == "n" || confirm == "no" {
		fmt.Println("❌ Cancelled")
		return
	}

	addToBacklog(fullDescription)
}
