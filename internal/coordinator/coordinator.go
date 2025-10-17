package coordinator

import (
	"context"
	"fmt"
	"log"
)

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
	CreateTask(title, description, role string) (taskID string, err error)
	ClaimTask(taskID, agent string) error
	UpdateTaskStatus(taskID, status string) error
	CompleteTask(taskID, result string) error
	FailTask(taskID, errorMsg string) error
	GetTask(taskID string) (*Task, error)

	// File coordination
	LockFiles(taskID, agent string, files []string) error

	// Access to sub-systems (needed for agents)
	GetEventBus() EventBus
	GetRegistry() Registry
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

// New creates a new SQLite-based Coordinator instance
func New(projectPath string) Coordinator {
	coord, err := NewSQLite(projectPath)
	if err != nil {
		log.Fatalf("Failed to create coordinator: %v", err)
	}
	return coord
}

// MustNew creates a coordinator or panics
func MustNew(projectPath string) Coordinator {
	coord, err := NewSQLite(projectPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to create coordinator: %v", err))
	}
	return coord
}
