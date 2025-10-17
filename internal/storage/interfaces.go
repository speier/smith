// Package storage provides the storage abstraction layer for Smith's multi-agent coordination.
//
// Architecture:
//   - Storage interfaces (EventStore, AgentStore, TaskStore, LockStore) define all operations
//   - BBolt implementation provides lock-free concurrent access for multiple agents
//   - Storage is fully swappable - implement the Store interface to use any backend
//
// Current Backend: BBolt (go.etcd.io/bbolt)
//   - Embedded key-value database (same as etcd/Kubernetes)
//   - Single-writer MVCC eliminates database lock contention
//   - Pure Go, no CGo dependencies
//   - Battle-tested in production systems
//
// Why not SQLite?
//   - SQLITE_BUSY errors with concurrent agents (4+ agents hitting lock timeouts)
//   - CGo dependency complicates cross-compilation
//   - BBolt's simple key-value model fits our access patterns better
//
// To swap backends: Implement the Store interface and change InitProjectStorage()
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
	SessionID string // Session this event belongs to
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
	Priority    int      // 0=low, 1=medium (default), 2=high
	DependsOn   []string // Task IDs that must be completed first
	SessionID   string   // Session this task belongs to
	StartedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

// Session represents a work session
type Session struct {
	SessionID  string // "session-2025-10-17-001"
	Title      string // Auto-generated from first task or user-provided
	StartedAt  time.Time
	LastActive time.Time
	TaskCount  int    // Total tasks in session
	Status     string // "active" | "archived"
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

// SessionStore defines the interface for session storage operations
type SessionStore interface {
	// CreateSession creates a new session
	CreateSession(ctx context.Context, session *Session) error

	// GetSession retrieves a single session
	GetSession(ctx context.Context, sessionID string) (*Session, error)

	// UpdateSession updates an existing session
	UpdateSession(ctx context.Context, session *Session) error

	// ListSessions retrieves sessions, sorted by LastActive (most recent first)
	ListSessions(ctx context.Context, limit int) ([]*Session, error)

	// ArchiveSession marks a session as archived
	ArchiveSession(ctx context.Context, sessionID string) error

	// GetSessionTasks retrieves all tasks for a session
	GetSessionTasks(ctx context.Context, sessionID string) ([]*Task, error)
}

// Store combines all storage interfaces
type Store interface {
	EventStore
	AgentStore
	TaskStore
	LockStore
	SessionStore

	// Close closes the storage backend
	Close() error
}
