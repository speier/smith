package storage

import (
	"context"
	"time"
)

// EventStore defines the interface for event storage operations
type EventStore interface {
	// SaveEvent stores a new event
	SaveEvent(ctx context.Context, event *Event) error

	// QueryEvents retrieves events matching the filter
	QueryEvents(ctx context.Context, filter EventFilter) ([]*Event, error)
}

// Event represents a system event
type Event struct {
	ID        int64
	Timestamp time.Time
	AgentID   string
	AgentRole string
	EventType string
	TaskID    *string
	FilePath  *string
	Data      string
}

// EventFilter defines criteria for querying events
type EventFilter struct {
	EventTypes []string
	AgentID    *string
	TaskID     *string
}

// AgentStore defines the interface for agent registry operations
type AgentStore interface {
	// RegisterAgent registers or updates an agent
	RegisterAgent(ctx context.Context, agent *Agent) error

	// UpdateHeartbeat updates agent's last heartbeat time
	UpdateHeartbeat(ctx context.Context, agentID string) error

	// UnregisterAgent removes an agent
	UnregisterAgent(ctx context.Context, agentID string) error

	// GetAgent retrieves a single agent
	GetAgent(ctx context.Context, agentID string) (*Agent, error)

	// ListAgents retrieves agents, optionally filtered by role
	ListAgents(ctx context.Context, role *string) ([]*Agent, error)

	// MarkAgentDead marks agents as dead if heartbeat is too old
	MarkAgentDead(ctx context.Context, timeout time.Duration) (int, error)
}

// Agent represents an agent in the registry
type Agent struct {
	ID            string
	Role          string
	Status        string // active, idle, dead
	TaskID        *string
	PID           int
	StartedAt     time.Time
	LastHeartbeat time.Time
}

// TaskStore defines the interface for task storage operations
type TaskStore interface {
	// CreateTask creates a new task
	CreateTask(ctx context.Context, task *Task) error

	// GetTask retrieves a single task
	GetTask(ctx context.Context, taskID string) (*Task, error)

	// UpdateTask updates an existing task
	UpdateTask(ctx context.Context, task *Task) error

	// ListTasks retrieves tasks, optionally filtered by status
	ListTasks(ctx context.Context, status *string) ([]*Task, error)

	// ClaimTask atomically claims a task for an agent
	ClaimTask(ctx context.Context, taskID, agentID string) error

	// GetTaskStats returns statistics about tasks by status
	GetTaskStats(ctx context.Context) (*TaskStats, error)
}

// Task represents a task assignment
type Task struct {
	TaskID      string
	Title       string
	Description string
	AgentID     string
	AgentRole   string
	Status      string
	Result      string
	Error       string
	StartedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

// TaskStats represents task statistics
type TaskStats struct {
	Backlog int
	WIP     int
	Review  int
	Done    int
}

// LockStore defines the interface for file lock operations
type LockStore interface {
	// AcquireLocks atomically acquires locks on multiple files
	AcquireLocks(ctx context.Context, locks []*FileLock) error

	// ReleaseLocks releases locks held by an agent
	ReleaseLocks(ctx context.Context, agentID string, files []string) error

	// GetLocks retrieves all active locks
	GetLocks(ctx context.Context) ([]*FileLock, error)

	// GetLocksForAgent retrieves locks held by a specific agent
	GetLocksForAgent(ctx context.Context, agentID string) ([]*FileLock, error)
}

// FileLock represents a lock on a file
type FileLock struct {
	FilePath string
	AgentID  string
	TaskID   string
	LockedAt time.Time
}

// Store combines all storage interfaces
type Store interface {
	EventStore
	AgentStore
	TaskStore
	LockStore

	// Close closes the storage backend
	Close() error
}
