package coordinator

import (
	"context"
	"fmt"
	"log"
)

// TaskOptions holds optional parameters for task creation
type TaskOptions struct {
	Priority        int               // 0=low, 1=medium (default), 2=high
	DependsOn       []string          // Task IDs that must be completed first
	Learnings       string            // What was learned
	TriedApproaches []string          // Approaches attempted
	Blockers        []string          // What didn't work
	Notes           map[string]string // Freeform notes
}

// TaskOption is a functional option for CreateTask
type TaskOption func(*TaskOptions)

// WithPriority sets the task priority
func WithPriority(priority int) TaskOption {
	return func(opts *TaskOptions) {
		opts.Priority = priority
	}
}

// WithDependencies sets the task dependencies
func WithDependencies(dependsOn ...string) TaskOption {
	return func(opts *TaskOptions) {
		opts.DependsOn = dependsOn
	}
}

// WithLearnings sets what was learned during task execution
func WithLearnings(learnings string) TaskOption {
	return func(opts *TaskOptions) {
		opts.Learnings = learnings
	}
}

// WithTriedApproaches records approaches that were attempted
func WithTriedApproaches(approaches ...string) TaskOption {
	return func(opts *TaskOptions) {
		opts.TriedApproaches = approaches
	}
}

// WithBlockers records what didn't work or blocked progress
func WithBlockers(blockers ...string) TaskOption {
	return func(opts *TaskOptions) {
		opts.Blockers = blockers
	}
}

// WithNotes sets freeform key-value notes
func WithNotes(notes map[string]string) TaskOption {
	return func(opts *TaskOptions) {
		opts.Notes = notes
	}
}

// Coordinator defines the interface for task and file coordination
// This interface is storage-agnostic and can be implemented with different backends
type Coordinator interface {
	EnsureDirectories() error
	GetTaskStats() (*TaskStats, error)
	GetAvailableTasks() ([]Task, error)
	GetTasksByStatus(status string) ([]Task, error)
	GetActiveLocks() ([]Lock, error)
	GetMessages() ([]Message, error)

	// Task lifecycle management
	CreateTask(title, description, role string, opts ...TaskOption) (taskID string, err error)
	ClaimTask(taskID, agent string) error
	UpdateTaskStatus(taskID, status string) error
	CompleteTask(taskID, result string, opts ...TaskOption) error
	FailTask(taskID, errorMsg string, opts ...TaskOption) error
	GetTask(taskID string) (*Task, error)
	GetRecentTasks(ctx context.Context, role string, limit int) ([]*Task, error)

	// File coordination
	LockFiles(taskID, agent string, files []string) error

	// Access to sub-systems (needed for agents)
	GetEventBus() EventBus
	GetRegistry() Registry

	// Session management
	CreateNewSession(ctx context.Context) (sessionID string, err error)
	ListSessions(ctx context.Context, limit int) ([]*Session, error)
	SwitchSession(ctx context.Context, sessionID string) error
	GetCurrentSession(ctx context.Context) (*Session, error)
}

// EventBus defines the interface for event publishing and querying
type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Query(ctx context.Context, filter EventFilter) ([]Event, error)
}

// Registry defines the interface for agent registration and heartbeat
type Registry interface {
	Register(ctx context.Context, agentID string, role AgentRole, pid int) error
	Heartbeat(ctx context.Context, agentID string) error
	Unregister(ctx context.Context, agentID string) error
	GetActiveAgents(ctx context.Context) ([]Agent, error)
}

// Event represents a system event (simplified for interface)
type Event struct {
	AgentID   string
	AgentRole AgentRole
	Type      EventType
	TaskID    *string
	Data      string
}

// EventFilter defines criteria for querying events
type EventFilter struct {
	EventTypes []EventType
}

// Agent represents an agent in the registry
type Agent struct {
	ID     string
	Role   AgentRole
	Status string
}

// EventType and AgentRole are defined as strings for interface compatibility
type EventType string
type AgentRole string

// New creates a new BBolt-based Coordinator instance
func New(projectPath string) Coordinator {
	coord, err := NewBolt(projectPath)
	if err != nil {
		log.Fatalf("Failed to create coordinator: %v", err)
	}
	return coord
}

// MustNew creates a coordinator or panics
func MustNew(projectPath string) Coordinator {
	coord, err := NewBolt(projectPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to create coordinator: %v", err))
	}
	return coord
}
