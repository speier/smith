package kanban

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Task represents a task parsed from kanban.md
type Task struct {
	ID          string
	Title       string
	Description string
	Status      string // "backlog", "wip", "review", "done"
	Checked     bool   // Whether the checkbox is marked
}

// Board represents the entire kanban board
type Board struct {
	Backlog []Task
	WIP     []Task
	Review  []Task
	Done    []Task
}

var (
	// Matches: ## Backlog, ## WIP, ## Review, ## Done
	sectionRegex = regexp.MustCompile(`^##\s+(.+)$`)

	// Matches: - [ ] task-001: Title or - [x] task-001: Title
	taskRegex = regexp.MustCompile(`^-\s+\[([ x])\]\s+(.+)$`)

	// Matches: task-001: Some title
	taskIDRegex = regexp.MustCompile(`^([a-zA-Z0-9-]+):\s*(.+)$`)
)

// Parse reads kanban.md and returns a Board with all tasks
func Parse(filePath string) (*Board, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open kanban file: %w", err)
	}
	defer file.Close()

	board := &Board{}
	var currentSection string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "<!--") {
			continue
		}

		// Check for section headers
		if matches := sectionRegex.FindStringSubmatch(line); matches != nil {
			currentSection = strings.ToLower(matches[1])
			continue
		}

		// Check for task items
		if matches := taskRegex.FindStringSubmatch(line); matches != nil {
			checked := matches[1] == "x"
			taskContent := strings.TrimSpace(matches[2])

			task := Task{
				Checked: checked,
				Status:  currentSection,
			}

			// Try to extract task ID and title
			if idMatches := taskIDRegex.FindStringSubmatch(taskContent); idMatches != nil {
				task.ID = idMatches[1]
				task.Title = strings.TrimSpace(idMatches[2])
			} else {
				// No ID format, use whole content as title
				task.Title = taskContent
			}

			// Add to appropriate section
			switch currentSection {
			case "backlog":
				board.Backlog = append(board.Backlog, task)
			case "wip":
				board.WIP = append(board.WIP, task)
			case "review":
				board.Review = append(board.Review, task)
			case "done":
				board.Done = append(board.Done, task)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading kanban file: %w", err)
	}

	return board, nil
}

// AllTasks returns all tasks from all sections
func (b *Board) AllTasks() []Task {
	var all []Task
	all = append(all, b.Backlog...)
	all = append(all, b.WIP...)
	all = append(all, b.Review...)
	all = append(all, b.Done...)
	return all
}

// WriteToFile writes the board back to a kanban.md file
func (b *Board) WriteToFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create kanban file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write header
	fmt.Fprintln(writer, "# Agent Kanban Board")
	fmt.Fprintln(writer)

	// Write each section
	writeSection(writer, "Backlog", "Tasks waiting to be picked up", b.Backlog)
	writeSection(writer, "WIP", "Work in progress - tasks currently being worked on", b.WIP)
	writeSection(writer, "Review", "Tasks pending review", b.Review)
	writeSection(writer, "Done", "Completed tasks", b.Done)

	return nil
}

func writeSection(writer *bufio.Writer, title, comment string, tasks []Task) {
	fmt.Fprintf(writer, "## %s\n", title)
	fmt.Fprintf(writer, "<!-- %s -->\n\n", comment)

	if len(tasks) == 0 {
		// Empty section, just write newline
		fmt.Fprintln(writer)
		return
	}

	for _, task := range tasks {
		checkbox := " "
		if task.Checked {
			checkbox = "x"
		}

		if task.ID != "" {
			fmt.Fprintf(writer, "- [%s] %s: %s\n", checkbox, task.ID, task.Title)
		} else {
			fmt.Fprintf(writer, "- [%s] %s\n", checkbox, task.Title)
		}
	}
	fmt.Fprintln(writer)
}
