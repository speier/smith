package coordinator

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/speier/smith/internal/eventbus"
	"github.com/speier/smith/internal/kanban"
	"github.com/speier/smith/internal/locks"
	"github.com/speier/smith/internal/registry"
	"github.com/speier/smith/internal/storage"
)

// SQLiteCoordinator implements coordination using SQLite database
type SQLiteCoordinator struct {
	projectPath string
	db          *storage.DB
	eventBus    *eventbus.EventBus
	lockMgr     *locks.Manager
	registry    *registry.Registry
}

// NewSQLite creates a new SQLite-based coordinator
func NewSQLite(projectPath string) (*SQLiteCoordinator, error) {
	// Initialize storage (creates .smith/ directory, db, kanban, config, etc.)
	db, err := storage.InitProjectStorage(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	c := &SQLiteCoordinator{
		projectPath: projectPath,
		db:          db,
		eventBus:    eventbus.New(db.DB),
		lockMgr:     locks.New(db.DB),
		registry:    registry.New(db.DB),
	}

	// Sync kanban.md to database on initialization
	if err := c.syncKanbanToDB(); err != nil {
		return nil, fmt.Errorf("failed to sync kanban to database: %w", err)
	}

	return c, nil
}

// EnsureDirectories creates the .smith directory structure
// For SQLite coordinator, this just ensures the database is initialized
func (c *SQLiteCoordinator) EnsureDirectories() error {
	// Storage initialization already happened in NewSQLite
	// This is here for interface compatibility with FileCoordinator
	return nil
}

// GetTaskStats returns statistics about tasks
func (c *SQLiteCoordinator) GetTaskStats() (*TaskStats, error) {
	ctx := context.Background()

	// Query task assignments to get counts by status
	rows, err := c.db.QueryContext(ctx, `
		SELECT status, COUNT(*) as count
		FROM task_assignments
		GROUP BY status
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query task stats: %w", err)
	}
	defer rows.Close()

	stats := &TaskStats{}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan task stats: %w", err)
		}

		switch status {
		case "backlog":
			stats.Backlog = count
		case "wip":
			stats.WIP = count
		case "review":
			stats.Review = count
		case "done":
			stats.Done = count
		}
	}

	return stats, nil
}

// GetAvailableTasks returns tasks in the backlog (not yet claimed)
func (c *SQLiteCoordinator) GetAvailableTasks() ([]Task, error) {
	ctx := context.Background()

	// Query for backlog tasks
	rows, err := c.db.QueryContext(ctx, `
		SELECT task_id, agent_role, status
		FROM task_assignments
		WHERE status = 'backlog'
		ORDER BY started_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query available tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var taskID, status string
		var role sql.NullString
		if err := rows.Scan(&taskID, &role, &status); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		roleStr := ""
		if role.Valid {
			roleStr = role.String
		}

		tasks = append(tasks, Task{
			ID:     taskID,
			Status: status,
			Role:   roleStr,
			Title:  taskID, // Title is the task ID for now
		})
	}

	return tasks, nil
}

// GetActiveLocks returns currently active file locks
func (c *SQLiteCoordinator) GetActiveLocks() ([]Lock, error) {
	ctx := context.Background()

	fileLocks, err := c.lockMgr.GetAllLocks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get locks: %w", err)
	}

	var locks []Lock
	for _, fl := range fileLocks {
		locks = append(locks, Lock{
			Agent:  fl.AgentID,
			TaskID: fl.TaskID,
			Files:  fl.FilePath, // TODO: Group by agent/task
		})
	}

	return locks, nil
}

// GetMessages returns messages for agents
func (c *SQLiteCoordinator) GetMessages() ([]Message, error) {
	ctx := context.Background()

	// Query events of type "agent_message"
	events, err := c.eventBus.Query(ctx, eventbus.EventFilter{
		SinceID: 0,
		EventTypes: []eventbus.EventType{
			eventbus.EventAgentMessage,
			eventbus.EventAgentQuestion,
			eventbus.EventAgentResponse,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}

	var messages []Message
	for _, event := range events {
		// Parse message data from event
		messages = append(messages, Message{
			From:    event.AgentID,
			To:      "", // TODO: Extract from event.Data
			Subject: string(event.Type),
			Body:    event.Data,
		})
	}

	return messages, nil
}

// ClaimTask claims a task for an agent
func (c *SQLiteCoordinator) ClaimTask(taskID, agent string) error {
	ctx := context.Background()

	// Check if task exists and is available
	var currentStatus string
	var currentAgent sql.NullString
	err := c.db.QueryRowContext(ctx, `
		SELECT status, agent_id FROM task_assignments WHERE task_id = ?
	`, taskID).Scan(&currentStatus, &currentAgent)

	if err != nil {
		return fmt.Errorf("task %s not found: %w", taskID, err)
	}

	if currentStatus != "backlog" {
		return fmt.Errorf("task %s is not available (status: %s)", taskID, currentStatus)
	}

	// Update task to WIP and assign to agent
	_, err = c.db.ExecContext(ctx, `
		UPDATE task_assignments 
		SET agent_id = ?, agent_role = ?, status = 'wip'
		WHERE task_id = ?
	`, agent, "implementation", taskID) // TODO: Determine role

	if err != nil {
		return fmt.Errorf("failed to claim task: %w", err)
	}

	// Sync back to kanban.md
	if err := c.syncDBToKanban(); err != nil {
		// Log error but don't fail the claim
		fmt.Printf("Warning: failed to sync to kanban.md: %v\n", err)
	}

	// Publish event
	return c.eventBus.PublishWithData(
		ctx,
		agent,
		eventbus.RoleImplementation, // TODO: Dynamic role
		eventbus.EventTaskClaimed,
		&taskID,
		nil,
		map[string]string{"task_id": taskID},
	)
}

// LockFiles locks files for a task/agent
func (c *SQLiteCoordinator) LockFiles(taskID, agent string, files []string) error {
	ctx := context.Background()

	// Acquire lock on each file
	for _, file := range files {
		if err := c.lockMgr.Acquire(ctx, file, agent, taskID); err != nil {
			// If lock fails, release any we acquired
			c.lockMgr.ReleaseAll(ctx, agent)
			return fmt.Errorf("failed to lock file %s: %w", file, err)
		}

		// Publish lock event
		c.eventBus.PublishWithData(
			ctx,
			agent,
			eventbus.RoleImplementation,
			eventbus.EventFileLocked,
			&taskID,
			&file,
			map[string]string{"file": file},
		)
	}

	return nil
}

// Close closes the database connection
func (c *SQLiteCoordinator) Close() error {
	return c.db.Close()
}

// syncKanbanToDB reads kanban.md and syncs tasks to the database
func (c *SQLiteCoordinator) syncKanbanToDB() error {
	kanbanPath := filepath.Join(c.projectPath, ".smith", "kanban.md")

	// Parse kanban.md
	board, err := kanban.Parse(kanbanPath)
	if err != nil {
		return fmt.Errorf("failed to parse kanban.md: %w", err)
	}

	ctx := context.Background()

	// Begin transaction
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing tasks
	if _, err := tx.ExecContext(ctx, "DELETE FROM task_assignments"); err != nil {
		return fmt.Errorf("failed to clear existing tasks: %w", err)
	}

	// Insert all tasks from the board
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO task_assignments (task_id, title, agent_id, agent_role, status)
		VALUES (?, ?, NULL, NULL, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	// Insert tasks from each section
	for _, task := range board.AllTasks() {
		// Skip empty task IDs
		if task.ID == "" {
			continue
		}

		_, err = stmt.ExecContext(ctx, task.ID, task.Title, task.Status)
		if err != nil {
			return fmt.Errorf("failed to insert task %s: %w", task.ID, err)
		}
	} // Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// syncDBToKanban reads tasks from the database and writes them to kanban.md
func (c *SQLiteCoordinator) syncDBToKanban() error {
	ctx := context.Background()

	// Query all tasks grouped by status
	rows, err := c.db.QueryContext(ctx, `
		SELECT task_id, title, status
		FROM task_assignments
		ORDER BY status, started_at ASC
	`)
	if err != nil {
		return fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	// Build board from database
	board := &kanban.Board{}

	for rows.Next() {
		var taskID, title, status string
		if err := rows.Scan(&taskID, &title, &status); err != nil {
			return fmt.Errorf("failed to scan task: %w", err)
		}

		task := kanban.Task{
			ID:     taskID,
			Title:  title,
			Status: status,
		}

		// Add to appropriate section
		switch status {
		case "backlog":
			board.Backlog = append(board.Backlog, task)
		case "wip":
			task.Checked = true // Tasks in progress are checked
			board.WIP = append(board.WIP, task)
		case "review":
			task.Checked = true // Tasks in review are checked
			board.Review = append(board.Review, task)
		case "done":
			task.Checked = true // Completed tasks are checked
			board.Done = append(board.Done, task)
		}
	}

	// Write to kanban.md
	kanbanPath := filepath.Join(c.projectPath, ".smith", "kanban.md")
	if err := board.WriteToFile(kanbanPath); err != nil {
		return fmt.Errorf("failed to write kanban.md: %w", err)
	}

	return nil
}
