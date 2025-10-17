package registry

import (
	"context"
	"fmt"
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

// Agent represents an agent in the registry
type Agent struct {
	ID            string
	Role          eventbus.AgentRole
	Status        AgentStatus
	TaskID        *string
	PID           int
	StartedAt     time.Time
	LastHeartbeat time.Time
}

// Registry manages agent registration and heartbeat tracking
type Registry struct {
	store storage.AgentStore
}

// New creates a new agent Registry
func New(store storage.AgentStore) *Registry {
	return &Registry{store: store}
}

// Register registers a new agent
func (r *Registry) Register(ctx context.Context, agentID string, role eventbus.AgentRole, pid int) error {
	agent := &storage.Agent{
		ID:            agentID,
		Role:          string(role),
		Status:        string(StatusActive),
		PID:           pid,
		StartedAt:     time.Now(),
		LastHeartbeat: time.Now(),
	}

	if err := r.store.RegisterAgent(ctx, agent); err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}

	return nil
}

// Unregister removes an agent from the registry
func (r *Registry) Unregister(ctx context.Context, agentID string) error {
	if err := r.store.UnregisterAgent(ctx, agentID); err != nil {
		return fmt.Errorf("failed to unregister agent: %w", err)
	}

	return nil
}

// Heartbeat updates the last heartbeat timestamp for an agent
func (r *Registry) Heartbeat(ctx context.Context, agentID string) error {
	if err := r.store.UpdateHeartbeat(ctx, agentID); err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	return nil
}

// UpdateStatus updates the status of an agent
func (r *Registry) UpdateStatus(ctx context.Context, agentID string, status AgentStatus) error {
	agent, err := r.store.GetAgent(ctx, agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	agent.Status = string(status)
	if err := r.store.RegisterAgent(ctx, agent); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// AssignTask assigns a task to an agent
func (r *Registry) AssignTask(ctx context.Context, agentID string, taskID string) error {
	agent, err := r.store.GetAgent(ctx, agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	agent.TaskID = &taskID
	agent.Status = string(StatusActive)
	if err := r.store.RegisterAgent(ctx, agent); err != nil {
		return fmt.Errorf("failed to assign task: %w", err)
	}

	return nil
}

// ClearTask clears the task assignment for an agent
func (r *Registry) ClearTask(ctx context.Context, agentID string) error {
	agent, err := r.store.GetAgent(ctx, agentID)
	if err != nil {
		return fmt.Errorf("failed to get agent: %w", err)
	}

	agent.TaskID = nil
	agent.Status = string(StatusIdle)
	if err := r.store.RegisterAgent(ctx, agent); err != nil {
		return fmt.Errorf("failed to clear task: %w", err)
	}

	return nil
}

// Get retrieves an agent by ID
func (r *Registry) Get(ctx context.Context, agentID string) (*Agent, error) {
	storageAgent, err := r.store.GetAgent(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	agent := &Agent{
		ID:            storageAgent.ID,
		Role:          eventbus.AgentRole(storageAgent.Role),
		Status:        AgentStatus(storageAgent.Status),
		TaskID:        storageAgent.TaskID,
		PID:           storageAgent.PID,
		StartedAt:     storageAgent.StartedAt,
		LastHeartbeat: storageAgent.LastHeartbeat,
	}

	return agent, nil
}

// List retrieves all agents, optionally filtered by role
func (r *Registry) List(ctx context.Context, role *eventbus.AgentRole) ([]Agent, error) {
	var roleStr *string
	if role != nil {
		r := string(*role)
		roleStr = &r
	}

	storageAgents, err := r.store.ListAgents(ctx, roleStr)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	var agents []Agent
	for _, sa := range storageAgents {
		agent := Agent{
			ID:            sa.ID,
			Role:          eventbus.AgentRole(sa.Role),
			Status:        AgentStatus(sa.Status),
			TaskID:        sa.TaskID,
			PID:           sa.PID,
			StartedAt:     sa.StartedAt,
			LastHeartbeat: sa.LastHeartbeat,
		}
		agents = append(agents, agent)
	}

	return agents, nil
}

// FindDeadAgents finds agents that haven't sent a heartbeat within the timeout period
func (r *Registry) FindDeadAgents(ctx context.Context, timeout time.Duration) ([]Agent, error) {
	// Mark agents as dead in storage first
	_, err := r.store.MarkAgentDead(ctx, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to mark dead agents: %w", err)
	}

	// Get all agents and filter for dead status
	storageAgents, err := r.store.ListAgents(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	var agents []Agent
	for _, sa := range storageAgents {
		if sa.Status == string(StatusDead) {
			agent := Agent{
				ID:            sa.ID,
				Role:          eventbus.AgentRole(sa.Role),
				Status:        AgentStatus(sa.Status),
				TaskID:        sa.TaskID,
				PID:           sa.PID,
				StartedAt:     sa.StartedAt,
				LastHeartbeat: sa.LastHeartbeat,
			}
			agents = append(agents, agent)
		}
	}

	return agents, nil
}

// MarkDead marks an agent as dead
func (r *Registry) MarkDead(ctx context.Context, agentID string) error {
	return r.UpdateStatus(ctx, agentID, StatusDead)
}

// CleanupDeadAgents removes agents marked as dead from the registry
func (r *Registry) CleanupDeadAgents(ctx context.Context) (int, error) {
	// Get all dead agents
	storageAgents, err := r.store.ListAgents(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to list agents: %w", err)
	}

	count := 0
	for _, agent := range storageAgents {
		if agent.Status == string(StatusDead) {
			if err := r.store.UnregisterAgent(ctx, agent.ID); err != nil {
				return count, fmt.Errorf("failed to unregister dead agent: %w", err)
			}
			count++
		}
	}

	return count, nil
}
