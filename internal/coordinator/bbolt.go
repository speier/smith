package coordinator

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/speier/smith/internal/eventbus"
	"github.com/speier/smith/internal/locks"
	"github.com/speier/smith/internal/registry"
	"github.com/speier/smith/internal/storage"
)

// BoltCoordinator implements coordination using BBolt database
type BoltCoordinator struct {
	projectPath      string
	db               storage.Store
	eventBus         *eventbus.EventBus
	lockMgr          *locks.Manager
	registry         *registry.Registry
	currentSessionID string // Active session ID
}

// NewBolt creates a new BBolt-based coordinator
func NewBolt(projectPath string) (*BoltCoordinator, error) {
	// Initialize storage (creates .smith/ directory, db, config, etc.)
	// Using BBolt for lock-free concurrent access by multiple agents
	store, err := storage.InitProjectStorage(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	c := &BoltCoordinator{
		projectPath: projectPath,
		db:          store,
		eventBus:    eventbus.New(store),
		lockMgr:     locks.New(store),
		registry:    registry.New(store),
	}

	return c, nil
}

// Registry returns the agent registry (legacy method - use GetRegistry instead)
func (c *BoltCoordinator) Registry() *registry.Registry {
	return c.registry
}

// DB returns the database connection (for testing)
func (c *BoltCoordinator) DB() storage.Store {
	return c.db
}

// GetEventBus returns the event bus (implementing Coordinator interface)
func (c *BoltCoordinator) GetEventBus() EventBus {
	return &eventBusAdapter{c.eventBus}
}

// GetRegistry returns the agent registry (implementing Coordinator interface)
func (c *BoltCoordinator) GetRegistry() Registry {
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

	// Filter for agents that are alive (not dead) and return them with working status
	var activeAgents []Agent
	for _, ag := range agents {
		if ag.Status != registry.StatusDead {
			// Agent is "active" (green) only if it has a task assigned
			status := "idle"
			if ag.TaskID != nil && *ag.TaskID != "" {
				status = "active"
			}
			activeAgents = append(activeAgents, Agent{
				ID:     ag.ID,
				Role:   AgentRole(ag.Role),
				Status: status,
			})
		}
	}
	return activeAgents, nil
}

// EnsureDirectories creates the .smith directory structure
// For BoltCoordinator, this just ensures the database is initialized
func (c *BoltCoordinator) EnsureDirectories() error {
	// Storage initialization already happened in NewBolt
	// This is here for interface compatibility
	return nil
}

// === Session Management ===

// GetOrCreateSession returns the current session, creating one if needed
func (c *BoltCoordinator) GetOrCreateSession(ctx context.Context) (string, error) {
	if c.currentSessionID != "" {
		return c.currentSessionID, nil
	}

	// Try to find most recent active session
	sessions, err := c.db.ListSessions(ctx, 1)
	if err == nil && len(sessions) > 0 && sessions[0].Status == "active" {
		c.currentSessionID = sessions[0].SessionID
		return c.currentSessionID, nil
	}

	// Create new session
	sessionID := c.generateSessionID()
	session := &storage.Session{
		SessionID:  sessionID,
		Title:      "New Session",
		StartedAt:  time.Now(),
		LastActive: time.Now(),
		TaskCount:  0,
		Status:     "active",
	}

	if err := c.db.CreateSession(ctx, session); err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	c.currentSessionID = sessionID
	return sessionID, nil
}

// CreateNewSession archives the current session and creates a new one
func (c *BoltCoordinator) CreateNewSession(ctx context.Context) (string, error) {
	// Archive current session if exists
	if c.currentSessionID != "" {
		if err := c.db.ArchiveSession(ctx, c.currentSessionID); err != nil {
			return "", fmt.Errorf("failed to archive session: %w", err)
		}
	}

	// Create new session
	sessionID := c.generateSessionID()
	session := &storage.Session{
		SessionID:  sessionID,
		Title:      "New Session",
		StartedAt:  time.Now(),
		LastActive: time.Now(),
		TaskCount:  0,
		Status:     "active",
	}

	if err := c.db.CreateSession(ctx, session); err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	c.currentSessionID = sessionID
	return sessionID, nil
}

// SwitchSession switches to a different session
func (c *BoltCoordinator) SwitchSession(ctx context.Context, sessionID string) error {
	// Verify session exists
	session, err := c.db.GetSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Update last active time
	session.LastActive = time.Now()
	if err := c.db.UpdateSession(ctx, session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	c.currentSessionID = sessionID
	return nil
}

// GetCurrentSession returns the current session info
func (c *BoltCoordinator) GetCurrentSession(ctx context.Context) (*Session, error) {
	sessionID, err := c.GetOrCreateSession(ctx)
	if err != nil {
		return nil, err
	}

	storageSession, err := c.db.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	return &Session{
		SessionID:  storageSession.SessionID,
		Title:      storageSession.Title,
		StartedAt:  storageSession.StartedAt,
		LastActive: storageSession.LastActive,
		TaskCount:  storageSession.TaskCount,
		Status:     storageSession.Status,
	}, nil
}

// ListSessions returns recent sessions
func (c *BoltCoordinator) ListSessions(ctx context.Context, limit int) ([]*Session, error) {
	storageSessions, err := c.db.ListSessions(ctx, limit)
	if err != nil {
		return nil, err
	}

	var sessions []*Session
	for _, ss := range storageSessions {
		sessions = append(sessions, &Session{
			SessionID:  ss.SessionID,
			Title:      ss.Title,
			StartedAt:  ss.StartedAt,
			LastActive: ss.LastActive,
			TaskCount:  ss.TaskCount,
			Status:     ss.Status,
		})
	}

	return sessions, nil
}

// UpdateSessionTitle updates the title of a session (auto-generated from first task)
func (c *BoltCoordinator) UpdateSessionTitle(ctx context.Context, sessionID, title string) error {
	session, err := c.db.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.Title = title
	return c.db.UpdateSession(ctx, session)
}

func (c *BoltCoordinator) generateSessionID() string {
	now := time.Now()
	return fmt.Sprintf("session-%s-%03d",
		now.Format("2006-01-02"),
		now.UnixNano()%1000)
}

// GetTaskStats returns statistics about tasks
func (c *BoltCoordinator) GetTaskStats() (*TaskStats, error) {
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
func (c *BoltCoordinator) GetAvailableTasks() ([]Task, error) {
	ctx := context.Background()

	status := "backlog"
	storageTasks, err := c.db.ListTasks(ctx, &status)
	if err != nil {
		return nil, fmt.Errorf("failed to list available tasks: %w", err)
	}

	var tasks []Task
	for _, st := range storageTasks {
		// Check if dependencies are satisfied
		if len(st.DependsOn) > 0 {
			allDependenciesMet := true
			for _, depID := range st.DependsOn {
				depTask, err := c.db.GetTask(ctx, depID)
				if err != nil || depTask.Status != "done" {
					allDependenciesMet = false
					break
				}
			}
			// Skip this task if dependencies not met
			if !allDependenciesMet {
				continue
			}
		}

		task := Task{
			ID:              st.TaskID,
			Title:           st.Title,
			Description:     st.Description,
			Role:            st.AgentRole,
			Status:          st.Status,
			Priority:        st.Priority,
			DependsOn:       st.DependsOn,
			AgentID:         st.AgentID,
			Result:          st.Result,
			Error:           st.Error,
			StartedAt:       st.StartedAt,
			UpdatedAt:       st.UpdatedAt,
			CompletedAt:     st.CompletedAt,
			Learnings:       st.Learnings,
			TriedApproaches: st.TriedApproaches,
			Blockers:        st.Blockers,
			Notes:           st.Notes,
		}
		tasks = append(tasks, task)
	}

	// Sort by priority (high -> medium -> low)
	// Priority: 2=high, 1=medium, 0=low
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Priority > tasks[j].Priority
	})

	return tasks, nil
}

// GetTasksByStatus returns tasks filtered by status, or all tasks if status is empty
func (c *BoltCoordinator) GetTasksByStatus(status string) ([]Task, error) {
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
			ID:              st.TaskID,
			Title:           st.Title,
			Description:     st.Description,
			Role:            st.AgentRole,
			Status:          st.Status,
			AgentID:         st.AgentID,
			Result:          st.Result,
			Error:           st.Error,
			Priority:        st.Priority,
			DependsOn:       st.DependsOn,
			StartedAt:       st.StartedAt,
			UpdatedAt:       st.UpdatedAt,
			CompletedAt:     st.CompletedAt,
			Learnings:       st.Learnings,
			TriedApproaches: st.TriedApproaches,
			Blockers:        st.Blockers,
			Notes:           st.Notes,
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetActiveLocks returns currently active file locks
func (c *BoltCoordinator) GetActiveLocks() ([]Lock, error) {
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
func (c *BoltCoordinator) GetMessages() ([]Message, error) {
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
func (c *BoltCoordinator) ClaimTask(taskID, agent string) error {
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
func (c *BoltCoordinator) CreateTask(title, description, role string, opts ...TaskOption) (string, error) {
	ctx := context.Background()

	// Apply options
	options := TaskOptions{
		Priority:  1, // Default to medium priority
		DependsOn: []string{},
	}
	for _, opt := range opts {
		opt(&options)
	}

	// Ensure we have an active session
	sessionID, err := c.GetOrCreateSession(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get/create session: %w", err)
	}

	// Generate unique task ID (simple counter-based for now)
	allTasks, err := c.db.ListTasks(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list tasks: %w", err)
	}

	taskID := fmt.Sprintf("task-%03d", len(allTasks)+1)

	// Create task via TaskStore
	task := &storage.Task{
		TaskID:          taskID,
		Title:           title,
		Description:     description,
		AgentRole:       role,
		Status:          "backlog",
		Priority:        options.Priority,
		DependsOn:       options.DependsOn,
		SessionID:       sessionID,
		StartedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Learnings:       options.Learnings,
		TriedApproaches: options.TriedApproaches,
		Blockers:        options.Blockers,
		Notes:           options.Notes,
	}

	if err := c.db.CreateTask(ctx, task); err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	// Update session: increment task count and update LastActive
	session, err := c.db.GetSession(ctx, sessionID)
	if err == nil {
		session.TaskCount++
		session.LastActive = time.Now()

		// Auto-set session title from first task
		if session.TaskCount == 1 {
			session.Title = title
		}

		// Best effort - session title update is non-critical
		_ = c.db.UpdateSession(ctx, session)
	}

	// Publish task created event
	if err := c.eventBus.PublishWithData(
		ctx,
		"coordinator",
		eventbus.RoleCoordinator,
		eventbus.EventTaskCreated,
		&taskID,
		nil,
		map[string]string{"task_id": taskID, "title": title, "role": role, "session_id": sessionID},
	); err != nil {
		// Log error but don't fail task creation
		fmt.Printf("warning: failed to publish task_created event: %v\n", err)
	}

	return taskID, nil
}

// UpdateTaskStatus updates the status of a task
func (c *BoltCoordinator) UpdateTaskStatus(taskID, status string) error {
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
func (c *BoltCoordinator) CompleteTask(taskID, result string, opts ...TaskOption) error {
	ctx := context.Background()

	// Apply options
	options := &TaskOptions{}
	for _, opt := range opts {
		opt(options)
	}

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

	// Apply memory fields if provided
	if options.Learnings != "" {
		task.Learnings = options.Learnings
	}
	if len(options.TriedApproaches) > 0 {
		task.TriedApproaches = options.TriedApproaches
	}
	if len(options.Blockers) > 0 {
		task.Blockers = options.Blockers
	}
	if options.Notes != nil {
		task.Notes = options.Notes
	}

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
func (c *BoltCoordinator) FailTask(taskID, errorMsg string, opts ...TaskOption) error {
	ctx := context.Background()

	// Apply options
	options := &TaskOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Get task and move back to backlog
	task, err := c.db.GetTask(ctx, taskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	task.Status = "backlog"
	task.Error = errorMsg
	task.AgentID = ""
	task.UpdatedAt = time.Now()

	// Apply memory fields if provided (useful to record why it failed)
	if options.Learnings != "" {
		task.Learnings = options.Learnings
	}
	if len(options.TriedApproaches) > 0 {
		task.TriedApproaches = options.TriedApproaches
	}
	if len(options.Blockers) > 0 {
		task.Blockers = options.Blockers
	}
	if options.Notes != nil {
		task.Notes = options.Notes
	}

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
func (c *BoltCoordinator) GetTask(taskID string) (*Task, error) {
	ctx := context.Background()

	// Get task from TaskStore
	storageTask, err := c.db.GetTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Convert storage.Task to coordinator.Task
	task := &Task{
		ID:              storageTask.TaskID,
		Title:           storageTask.Title,
		Description:     storageTask.Description,
		Status:          storageTask.Status,
		Role:            storageTask.AgentRole,
		AgentID:         storageTask.AgentID,
		Result:          storageTask.Result,
		Error:           storageTask.Error,
		Priority:        storageTask.Priority,
		DependsOn:       storageTask.DependsOn,
		StartedAt:       storageTask.StartedAt,
		UpdatedAt:       storageTask.UpdatedAt,
		CompletedAt:     storageTask.CompletedAt,
		Learnings:       storageTask.Learnings,
		TriedApproaches: storageTask.TriedApproaches,
		Blockers:        storageTask.Blockers,
		Notes:           storageTask.Notes,
	}

	return task, nil
}

// GetRecentTasks retrieves recent tasks, optionally filtered by agent role
// This is the "memory query" API - agents use this to learn from past work
func (c *BoltCoordinator) GetRecentTasks(ctx context.Context, role string, limit int) ([]*Task, error) {
	// Get all tasks (pass nil to get all statuses)
	storageTasks, err := c.db.ListTasks(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	// Filter by role if specified
	var filtered []*storage.Task
	for _, st := range storageTasks {
		if role == "" || st.AgentRole == role {
			filtered = append(filtered, st)
		}
	}

	// Sort by UpdatedAt descending (most recent first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].UpdatedAt.After(filtered[j].UpdatedAt)
	})

	// Apply limit
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	// Convert to coordinator.Task
	tasks := make([]*Task, len(filtered))
	for i, st := range filtered {
		tasks[i] = &Task{
			ID:              st.TaskID,
			Title:           st.Title,
			Description:     st.Description,
			Status:          st.Status,
			Role:            st.AgentRole,
			AgentID:         st.AgentID,
			Result:          st.Result,
			Error:           st.Error,
			Priority:        st.Priority,
			DependsOn:       st.DependsOn,
			StartedAt:       st.StartedAt,
			UpdatedAt:       st.UpdatedAt,
			CompletedAt:     st.CompletedAt,
			Learnings:       st.Learnings,
			TriedApproaches: st.TriedApproaches,
			Blockers:        st.Blockers,
			Notes:           st.Notes,
		}
	}

	return tasks, nil
}

// LockFiles locks files for a task/agent
func (c *BoltCoordinator) LockFiles(taskID, agent string, files []string) error {
	ctx := context.Background()

	// Acquire lock on each file
	for _, file := range files {
		if err := c.lockMgr.Acquire(ctx, file, agent, taskID); err != nil {
			// If lock fails, release any we acquired (best effort)
			_ = c.lockMgr.ReleaseAll(ctx, agent)
			return fmt.Errorf("failed to lock file %s: %w", file, err)
		}

		// Publish lock event (best effort - not critical for operation)
		_ = c.eventBus.PublishWithData(
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

// GetSessionUsage retrieves token usage for a session
func (c *BoltCoordinator) GetSessionUsage(ctx context.Context, sessionID string) (*LLMUsage, error) {
	usage, err := c.db.GetSessionUsage(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Convert storage.LLMUsage to coordinator.LLMUsage
	return &LLMUsage{
		SessionID:        usage.SessionID,
		TotalTokens:      usage.TotalTokens,
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
	}, nil
}

// Close closes the database connection
func (c *BoltCoordinator) Close() error {
	return c.db.Close()
}
