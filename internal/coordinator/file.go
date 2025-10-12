package coordinator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// Directory structure
	SmithDir   = ".smith"
	BacklogDir = "backlog"
	InboxDir   = "inbox"
	AgentsDir  = "agents"

	// Backlog subdirectories
	TodoDir   = "todo"
	WipDir    = "wip" // Work in progress
	ReviewDir = "review"
	DoneDir   = "done"
)

type FileCoordinator struct {
	projectPath string
}

type Task struct {
	ID       string
	Title    string
	Status   string // "todo", "wip", "review", "done"
	Role     string // Inferred from task type
	Priority string
	Assigned string
}

type Lock struct {
	Agent  string
	TaskID string
	Files  string
}

type Message struct {
	From    string
	To      string
	Subject string
	Body    string
}

type TaskStats struct {
	Available  int
	InProgress int
	Done       int
	Blocked    int
}

func NewFile(projectPath string) *FileCoordinator {
	return &FileCoordinator{projectPath: projectPath}
}

// EnsureDirectories creates the .smith directory structure if it doesn't exist
func (c *FileCoordinator) EnsureDirectories() error {
	dirs := []string{
		filepath.Join(c.projectPath, SmithDir, BacklogDir, TodoDir),
		filepath.Join(c.projectPath, SmithDir, BacklogDir, WipDir),
		filepath.Join(c.projectPath, SmithDir, BacklogDir, ReviewDir),
		filepath.Join(c.projectPath, SmithDir, BacklogDir, DoneDir),
		filepath.Join(c.projectPath, SmithDir, InboxDir),
		filepath.Join(c.projectPath, SmithDir, AgentsDir),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	return nil
}

func (c *FileCoordinator) GetTaskStats() (*TaskStats, error) {
	stats := &TaskStats{}

	// Count tasks in each backlog directory
	todoPath := filepath.Join(c.projectPath, SmithDir, BacklogDir, TodoDir)
	wipPath := filepath.Join(c.projectPath, SmithDir, BacklogDir, WipDir)
	reviewPath := filepath.Join(c.projectPath, SmithDir, BacklogDir, ReviewDir)
	donePath := filepath.Join(c.projectPath, SmithDir, BacklogDir, DoneDir)

	stats.Available = c.countTasks(todoPath)
	stats.InProgress = c.countTasks(wipPath)
	stats.Blocked = c.countTasks(reviewPath) // Review = blocked waiting
	stats.Done = c.countTasks(donePath)

	return stats, nil
}

func (c *FileCoordinator) countTasks(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			count++
		}
	}
	return count
}

func (c *FileCoordinator) GetAvailableTasks() ([]Task, error) {
	return c.getTasksFromDir(TodoDir, "todo")
}

func (c *FileCoordinator) getTasksFromDir(subdir, status string) ([]Task, error) {
	dir := filepath.Join(c.projectPath, SmithDir, BacklogDir, subdir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading %s directory: %w", subdir, err)
	}

	var tasks []Task
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		// Parse task file
		// TODO: Implement YAML frontmatter parsing
		tasks = append(tasks, Task{
			ID:     strings.TrimSuffix(entry.Name(), ".md"),
			Title:  strings.TrimSuffix(entry.Name(), ".md"),
			Status: status,
		})
	}

	return tasks, nil
}

func (c *FileCoordinator) GetActiveLocks() ([]Lock, error) {
	content, err := c.readFile("COMMS.md")
	if err != nil {
		return nil, err
	}

	// TODO: Parse COMMS.md for active locks
	// For now, return empty
	_ = content
	return []Lock{}, nil
}

func (c *FileCoordinator) GetMessages() ([]Message, error) {
	// TODO: Parse COMMS.md for messages
	return []Message{}, nil
}

func (c *FileCoordinator) ClaimTask(taskID, agent string) error {
	// TODO: Update TODO.md to mark task as claimed
	return fmt.Errorf("not implemented")
}

func (c *FileCoordinator) LockFiles(taskID, agent string, files []string) error {
	// TODO: Update COMMS.md to add file locks
	return fmt.Errorf("not implemented")
}

func (c *FileCoordinator) readFile(name string) (string, error) {
	path := filepath.Join(c.projectPath, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", name, err)
	}
	return string(data), nil
}

func (c *FileCoordinator) writeFile(name, content string) error {
	path := filepath.Join(c.projectPath, name)
	return os.WriteFile(path, []byte(content), 0644)
}
