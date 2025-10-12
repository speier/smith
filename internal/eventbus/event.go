package eventbus

import (
	"time"
)

// EventType represents the type of event
type EventType string

const (
	// Agent lifecycle events
	EventAgentStarted   EventType = "agent_started"
	EventAgentStopped   EventType = "agent_stopped"
	EventAgentHeartbeat EventType = "agent_heartbeat"

	// Task events
	EventTaskClaimed   EventType = "task_claimed"
	EventTaskStarted   EventType = "task_started"
	EventTaskCompleted EventType = "task_completed"
	EventTaskFailed    EventType = "task_failed"
	EventTaskAbandoned EventType = "task_abandoned"

	// File lock events
	EventFileLocked     EventType = "file_locked"
	EventFileUnlocked   EventType = "file_unlocked"
	EventFileLockWait   EventType = "file_lock_wait"
	EventFileLockFailed EventType = "file_lock_failed"

	// Communication events
	EventAgentMessage  EventType = "agent_message"
	EventAgentQuestion EventType = "agent_question"
	EventAgentResponse EventType = "agent_response"

	// Error events
	EventError EventType = "error"
)

// AgentRole represents the role of an agent
type AgentRole string

const (
	RoleCoordinator    AgentRole = "coordinator"
	RolePlanning       AgentRole = "planning"
	RoleImplementation AgentRole = "implementation"
	RoleTesting        AgentRole = "testing"
	RoleReview         AgentRole = "review"
)

// Event represents an event in the system
type Event struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	AgentID   string    `json:"agent_id"`
	AgentRole AgentRole `json:"agent_role"`
	Type      EventType `json:"event_type"`
	TaskID    *string   `json:"task_id,omitempty"`
	FilePath  *string   `json:"file_path,omitempty"`
	Data      string    `json:"data,omitempty"` // JSON-encoded additional data
}

// EventFilter defines criteria for filtering events
type EventFilter struct {
	SinceID    int64       // Only events with ID > SinceID
	AgentID    *string     // Filter by agent ID
	AgentRole  *AgentRole  // Filter by agent role
	EventTypes []EventType // Filter by event types
	TaskID     *string     // Filter by task ID
	FilePath   *string     // Filter by file path
}
