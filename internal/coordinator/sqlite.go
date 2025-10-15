package coordinator

import (
	"context"
	"database/sql"
	"fmt"

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
	// Initialize storage (creates .smith/ directory, db, config, etc.)
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

	return c, nil
}

// Registry returns the agent registry
func (c *SQLiteCoordinator) Registry() *registry.Registry {
	return c.registry
}

// DB returns the database connection (for testing)
func (c *SQLiteCoordinator) DB() *storage.DB {
	return c.db
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
		SELECT task_id, title, description, agent_role, status
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
		var task Task
		var role sql.NullString
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &role, &task.Status); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if role.Valid {
			task.Role = role.String
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTasksByStatus returns tasks filtered by status, or all tasks if status is empty
func (c *SQLiteCoordinator) GetTasksByStatus(status string) ([]Task, error) {
	ctx := context.Background()

	var query string
	var args []interface{}

	if status == "" {
		// Get all tasks
		query = `
			SELECT task_id, title, description, agent_role, status
			FROM task_assignments
			ORDER BY started_at DESC
		`
	} else {
		// Filter by status
		query = `
			SELECT task_id, title, description, agent_role, status
			FROM task_assignments
			WHERE status = ?
			ORDER BY started_at DESC
		`
		args = append(args, status)
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		var role sql.NullString
		if err := rows.Scan(&task.ID, &task.Title, &task.Description, &role, &task.Status); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if role.Valid {
			task.Role = role.String
		}

		tasks = append(tasks, task)
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

// CreateTask creates a new task and adds it to the backlog
func (c *SQLiteCoordinator) CreateTask(title, description, role string) (string, error) {
	ctx := context.Background()

	// Generate task ID (simple incremental for now)
	var maxID int
	err := c.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(CAST(SUBSTR(task_id, 6) AS INTEGER)), 0)
		FROM task_assignments
		WHERE task_id LIKE 'task-%'
	`).Scan(&maxID)
	if err != nil && err != sql.ErrNoRows {
		return "", fmt.Errorf("failed to generate task ID: %w", err)
	}

	taskID := fmt.Sprintf("task-%03d", maxID+1)

	// Insert task into database
	_, err = c.db.ExecContext(ctx, `
		INSERT INTO task_assignments (task_id, title, description, agent_role, status)
		VALUES (?, ?, ?, ?, 'backlog')
	`, taskID, title, description, role)
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	// Publish event
	err = c.eventBus.PublishWithData(
		ctx,
		"coordinator",
		eventbus.RoleCoordinator,
		eventbus.EventTaskCreated,
		&taskID,
		nil,
		map[string]string{"task_id": taskID, "title": title},
	)
	if err != nil {
		return "", fmt.Errorf("failed to publish task created event: %w", err)
	}

	return taskID, nil
}

// UpdateTaskStatus updates the status of a task
func (c *SQLiteCoordinator) UpdateTaskStatus(taskID, status string) error {
	ctx := context.Background()

	// Validate status
	validStatuses := map[string]bool{
		"backlog": true,
		"wip":     true,
		"review":  true,
		"done":    true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	// Update status
	result, err := c.db.ExecContext(ctx, `
		UPDATE task_assignments 
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE task_id = ?
	`, status, taskID)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Publish event
	return c.eventBus.PublishWithData(
		ctx,
		"coordinator",
		eventbus.RoleCoordinator,
		eventbus.EventTaskUpdated,
		&taskID,
		nil,
		map[string]string{"task_id": taskID, "status": status},
	)
}

// CompleteTask marks a task as completed with a result
func (c *SQLiteCoordinator) CompleteTask(taskID, result string) error {
	ctx := context.Background()

	// Update task to done with result
	res, err := c.db.ExecContext(ctx, `
		UPDATE task_assignments 
		SET status = 'done', result = ?, completed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE task_id = ?
	`, result, taskID)
	if err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Publish event
	return c.eventBus.PublishWithData(
		ctx,
		"coordinator",
		eventbus.RoleCoordinator,
		eventbus.EventTaskCompleted,
		&taskID,
		nil,
		map[string]string{"task_id": taskID, "result": result},
	)
}

// FailTask marks a task as failed with an error message
func (c *SQLiteCoordinator) FailTask(taskID, errorMsg string) error {
	ctx := context.Background()

	// Update task status back to backlog and store error
	res, err := c.db.ExecContext(ctx, `
		UPDATE task_assignments 
		SET status = 'backlog', error = ?, agent_id = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE task_id = ?
	`, errorMsg, taskID)
	if err != nil {
		return fmt.Errorf("failed to fail task: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Publish event
	return c.eventBus.PublishWithData(
		ctx,
		"coordinator",
		eventbus.RoleCoordinator,
		eventbus.EventTaskFailed,
		&taskID,
		nil,
		map[string]string{"task_id": taskID, "error": errorMsg},
	)
}

// GetTask retrieves a task by ID
func (c *SQLiteCoordinator) GetTask(taskID string) (*Task, error) {
	ctx := context.Background()

	var task Task
	var agentID, agentRole, result, errMsg sql.NullString

	err := c.db.QueryRowContext(ctx, `
		SELECT task_id, title, description, agent_id, agent_role, status, result, error
		FROM task_assignments
		WHERE task_id = ?
	`, taskID).Scan(&task.ID, &task.Title, &task.Description, &agentID, &agentRole, &task.Status, &result, &errMsg)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if agentID.Valid {
		task.AgentID = agentID.String
	}
	if agentRole.Valid {
		task.Role = agentRole.String
	}
	if result.Valid {
		task.Result = result.String
	}
	if errMsg.Valid {
		task.Error = errMsg.String
	}

	return &task, nil
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
