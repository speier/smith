package coordinator

import (
	"context"
	"fmt"
	"time"

	"github.com/speier/smith/internal/eventbus"
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

	return &SQLiteCoordinator{
		projectPath: projectPath,
		db:          db,
		eventBus:    eventbus.New(db.DB),
		lockMgr:     locks.New(db.DB),
		registry:    registry.New(db.DB),
	}, nil
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
		case "claimed":
			stats.Available = count
		case "in_progress":
			stats.InProgress = count
		case "completed":
			stats.Done = count
		case "failed":
			stats.Blocked = count
		}
	}

	return stats, nil
}

// GetAvailableTasks returns tasks that haven't been claimed yet
func (c *SQLiteCoordinator) GetAvailableTasks() ([]Task, error) {
	ctx := context.Background()

	// Query for unclaimed tasks
	// For now, we'll parse from kanban.md (future: store in DB)
	// This is a simplified implementation - you'd read from kanban.md
	rows, err := c.db.QueryContext(ctx, `
		SELECT task_id, agent_role, status, started_at
		FROM task_assignments
		WHERE status = 'claimed'
		ORDER BY started_at ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query available tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var taskID, role, status string
		var startedAt time.Time
		if err := rows.Scan(&taskID, &role, &status, &startedAt); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		tasks = append(tasks, Task{
			ID:     taskID,
			Status: status,
			Role:   role,
			Title:  taskID, // TODO: Parse from kanban.md
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

	// Check if task already claimed
	var existingAgent string
	err := c.db.QueryRowContext(ctx, `
		SELECT agent_id FROM task_assignments WHERE task_id = ?
	`, taskID).Scan(&existingAgent)

	if err == nil {
		return fmt.Errorf("task %s already claimed by %s", taskID, existingAgent)
	}

	// Insert task assignment
	_, err = c.db.ExecContext(ctx, `
		INSERT INTO task_assignments (task_id, agent_id, agent_role, status)
		VALUES (?, ?, ?, 'claimed')
	`, taskID, agent, "implementation") // TODO: Determine role

	if err != nil {
		return fmt.Errorf("failed to claim task: %w", err)
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
