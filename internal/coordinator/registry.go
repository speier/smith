package coordinator

import (
	"time"

	"github.com/speier/smith/internal/eventbus"
	"github.com/speier/smith/internal/storage"
)

// AgentStatus represents the status of an agent
type AgentStatus string

const (
	StatusActive AgentStatus = "active"
	StatusIdle   AgentStatus = "idle"
	StatusDead   AgentStatus = "dead"
)

// RegisteredAgent represents an agent in the registry (detailed version with tracking info)
type RegisteredAgent struct {
	ID            string
	Role          eventbus.AgentRole
	Status        AgentStatus
	TaskID        *string
	PID           int
	StartedAt     time.Time
	LastHeartbeat time.Time
}

// AgentRegistry manages agent registration and heartbeat tracking
type AgentRegistry struct {
	store storage.AgentStore
}

// NewAgentRegistry creates a new agent registry
func NewAgentRegistry(store storage.AgentStore) *AgentRegistry {
	return &AgentRegistry{store: store}
}

// Agent represents an agent in the registry

// Unregister removes an agent from the registry

// FindDeadAgents finds agents that haven't sent a heartbeat within the timeout period
