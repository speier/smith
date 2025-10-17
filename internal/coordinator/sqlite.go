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
	db          storage.Store
	eventBus    *eventbus.EventBus
	lockMgr     *locks.Manager
	registry    *registry.Registry
}

// NewSQLite creates a new SQLite-based coordinator
func NewSQLite(projectPath string) (*SQLiteCoordinator, error) {
	// Initialize storage (creates .smith/ directory, db, config, etc.)
	store, err := storage.InitProjectStorage(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	c := &SQLiteCoordinator{
		projectPath: projectPath,
		db:          store,
		eventBus:    eventbus.New(store),
		lockMgr:     locks.New(store),
		registry:    registry.New(store),
	}

	return c, nil
}

// Registry returns the agent registry (legacy method - use GetRegistry instead)
func (c *SQLiteCoordinator) Registry() *registry.Registry {
	return c.registry
}

// DB returns the database connection (for testing)
func (c *SQLiteCoordinator) DB() storage.Store {
	return c.db
}

// GetEventBus returns the event bus (implementing Coordinator interface)
func (c *SQLiteCoordinator) GetEventBus() EventBus {
	return &eventBusAdapter{c.eventBus}
}

// GetRegistry returns the agent registry (implementing Coordinator interface)
func (c *SQLiteCoordinator) GetRegistry() Registry {
	return &registryAdapter{c.registry}
}

// eventBusAdapter wraps eventbus.EventBus to match the Coordinator.EventBus interface
type eventBusAdapter struct {
	*eventbus.EventBus
}

func (a *eventBusAdapter) Publish(ctx context.Context, event Event) error {
	// Convert interface Event to eventbus.Event
	ebEvent := &eventbus.Event{
		AgentID:   event.AgentID,
		AgentRole: eventbus.AgentRole(event.AgentRole),
		Type:      eventbus.EventType(event.Type),
		TaskID:    event.TaskID,
		Data:      event.Data,
	}
	return a.EventBus.Publish(ctx, ebEvent)
}

func (a *eventBusAdapter) Query(ctx context.Context, filter EventFilter) ([]Event, error) {
	// Convert interface EventFilter to eventbus.EventFilter
	ebFilter := eventbus.EventFilter{}
	for _, et := range filter.EventTypes {
		ebFilter.EventTypes = append(ebFilter.EventTypes, eventbus.EventType(et))
	}

	events, err := a.EventBus.Query(ctx, ebFilter)
	if err != nil {
		return nil, err
	}

	// Convert eventbus.Events to interface Events
	result := make([]Event, len(events))
	for i, e := range events {
		result[i] = Event{
			AgentID:   e.AgentID,
			AgentRole: AgentRole(e.AgentRole),
			Type:      EventType(e.Type),
			TaskID:    e.TaskID,
			Data:      e.Data,
		}
	}
	return result, nil
}

// registryAdapter wraps registry.Registry to match the Coordinator.Registry interface
type registryAdapter struct {
	*registry.Registry
}

func (a *registryAdapter) Register(ctx context.Context, agentID string, role AgentRole, pid int) error {
	return a.Registry.Register(ctx, agentID, eventbus.AgentRole(role), pid)
}

func (a *registryAdapter) Heartbeat(ctx context.Context, agentID string) error {
	return a.Registry.Heartbeat(ctx, agentID)
}

func (a *registryAdapter) Unregister(ctx context.Context, agentID string) error {
	return a.Registry.Unregister(ctx, agentID)
}

func (a *registryAdapter) GetActiveAgents(ctx context.Context) ([]Agent, error) {
	// List all agents (role filter = nil means all roles)
	agents, err := a.Registry.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Filter for active agents only
	var activeAgents []Agent
	for _, ag := range agents {
		if ag.Status == registry.StatusActive {
			activeAgents = append(activeAgents, Agent{
				ID:     ag.ID,
				Role:   AgentRole(ag.Role),
				Status: string(ag.Status),
			})
		}
	}
	return activeAgents, nil
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

	storageStats, err := c.db.GetTaskStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get task stats: %w", err)
	}

	stats := &TaskStats{
		Backlog: storageStats.Backlog,
		WIP:     storageStats.WIP,
		Review:  storageStats.Review,
		Done:    storageStats.Done,
	}

	return stats, nil
}

// GetAvailableTasks returns tasks in the backlog (not yet claimed)
func (c *SQLiteCoordinator) GetAvailableTasks() ([]Task, error) {
	ctx := context.Background()

	status := "backlog"
	storageTasks, err := c.db.ListTasks(ctx, &status)
	if err != nil {
		return nil, fmt.Errorf("failed to list available tasks: %w", err)
	}

	var tasks []Task
	for _, st := range storageTasks {
		task := Task{
			ID:          st.TaskID,
			Title:       st.Title,
			Description: st.Description,
			Role:        st.AgentRole,
			Status:      st.Status,
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTasksByStatus returns tasks filtered by status, or all tasks if status is empty
func (c *SQLiteCoordinator) GetTasksByStatus(status string) ([]Task, error) {
	ctx := context.Background()

	var statusFilter *string
	if status != "" {
		statusFilter = &status
	}

	storageTasks, err := c.db.ListTasks(ctx, statusFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}

	var tasks []Task
	for _, st := range storageTasks {
		task := Task{
			ID:          st.TaskID,
			Title:       st.Title,
			Description: st.Description,
			Role:        st.AgentRole,
			Status:      st.Status,
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

	// Use TaskStore to claim the task
	if err := c.db.ClaimTask(ctx, taskID, agent); err != nil {
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

// CreateTask creates a new task in the system
func (c *SQLiteCoordinator) CreateTask(title, description, role string) (string, error) {
	ctx := context.Background()

	// Generate unique task ID (simple counter-based for now)
	allTasks, err := c.db.ListTasks(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list tasks: %w", err)
	}

	taskID := fmt.Sprintf("task-%03d", len(allTasks)+1)

	// Create task via TaskStore
	task := &storage.Task{
		TaskID:      taskID,
		Title:       title,
		Description: description,
		AgentRole:   role,
		Status:      "backlog",
		StartedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := c.db.CreateTask(ctx, task); err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	// Publish task created event
	if err := c.eventBus.PublishWithData(
		ctx,
		"coordinator",
		eventbus.RoleCoordinator,
		eventbus.EventTaskCreated,
		&taskID,
		nil,
		map[string]string{"task_id": taskID, "title": title, "role": role},
	); err != nil {
		// Log error but don't fail task creation
		fmt.Printf("warning: failed to publish task_created event: %v\n", err)
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

	// Get task and update status
	task, err := c.db.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.Status = status
	task.UpdatedAt = time.Now()

	if err := c.db.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
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

	// Get task and mark as done
	task, err := c.db.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.Status = "done"
	task.Result = result
	now := time.Now()
	task.CompletedAt = &now
	task.UpdatedAt = now

	if err := c.db.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
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

	// Get task and move back to backlog
	task, err := c.db.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.Status = "backlog"
	task.Error = errorMsg
	task.AgentID = ""
	task.UpdatedAt = time.Now()

	if err := c.db.UpdateTask(ctx, task); err != nil {
		return fmt.Errorf("failed to fail task: %w", err)
	}

	// Skip rows check
	var rows int64 = 1
	_ = rows
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

	// Get task from TaskStore
	storageTask, err := c.db.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Convert storage.Task to coordinator.Task
	task := &Task{
		ID:          storageTask.TaskID,
		Title:       storageTask.Title,
		Description: storageTask.Description,
		Status:      storageTask.Status,
		Role:        storageTask.AgentRole,
		AgentID:     storageTask.AgentID,
		Result:      storageTask.Result,
		Error:       storageTask.Error,
	}

	return task, nil
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
