package registry

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/speier/smith/internal/eventbus"
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
	db *sql.DB
}

// New creates a new agent Registry
func New(db *sql.DB) *Registry {
	return &Registry{db: db}
}

// Register registers a new agent
func (r *Registry) Register(ctx context.Context, agentID string, role eventbus.AgentRole, pid int) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO agents (agent_id, agent_role, status, pid)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(agent_id) DO UPDATE SET
			agent_role = excluded.agent_role,
			status = excluded.status,
			pid = excluded.pid,
			started_at = CURRENT_TIMESTAMP,
			last_heartbeat = CURRENT_TIMESTAMP`,
		agentID, role, StatusActive, pid,
	)
	if err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}

	return nil
}

// Unregister removes an agent from the registry
func (r *Registry) Unregister(ctx context.Context, agentID string) error {
	_, err := r.db.ExecContext(
		ctx,
		"DELETE FROM agents WHERE agent_id = ?",
		agentID,
	)
	if err != nil {
		return fmt.Errorf("failed to unregister agent: %w", err)
	}

	return nil
}

// Heartbeat updates the last heartbeat timestamp for an agent
func (r *Registry) Heartbeat(ctx context.Context, agentID string) error {
	result, err := r.db.ExecContext(
		ctx,
		"UPDATE agents SET last_heartbeat = CURRENT_TIMESTAMP WHERE agent_id = ?",
		agentID,
	)
	if err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	return nil
}

// UpdateStatus updates the status of an agent
func (r *Registry) UpdateStatus(ctx context.Context, agentID string, status AgentStatus) error {
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE agents SET status = ? WHERE agent_id = ?",
		status, agentID,
	)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// AssignTask assigns a task to an agent
func (r *Registry) AssignTask(ctx context.Context, agentID string, taskID string) error {
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE agents SET task_id = ?, status = ? WHERE agent_id = ?",
		taskID, StatusActive, agentID,
	)
	if err != nil {
		return fmt.Errorf("failed to assign task: %w", err)
	}

	return nil
}

// ClearTask clears the task assignment for an agent
func (r *Registry) ClearTask(ctx context.Context, agentID string) error {
	_, err := r.db.ExecContext(
		ctx,
		"UPDATE agents SET task_id = NULL, status = ? WHERE agent_id = ?",
		StatusIdle, agentID,
	)
	if err != nil {
		return fmt.Errorf("failed to clear task: %w", err)
	}

	return nil
}

// Get retrieves an agent by ID
func (r *Registry) Get(ctx context.Context, agentID string) (*Agent, error) {
	var agent Agent
	var taskID sql.NullString

	err := r.db.QueryRowContext(
		ctx,
		`SELECT agent_id, agent_role, status, task_id, pid, started_at, last_heartbeat
		 FROM agents WHERE agent_id = ?`,
		agentID,
	).Scan(
		&agent.ID,
		&agent.Role,
		&agent.Status,
		&taskID,
		&agent.PID,
		&agent.StartedAt,
		&agent.LastHeartbeat,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	if taskID.Valid {
		agent.TaskID = &taskID.String
	}

	return &agent, nil
}

// List retrieves all agents, optionally filtered by role
func (r *Registry) List(ctx context.Context, role *eventbus.AgentRole) ([]Agent, error) {
	query := `SELECT agent_id, agent_role, status, task_id, pid, started_at, last_heartbeat
	          FROM agents`
	var args []interface{}

	if role != nil {
		query += " WHERE agent_role = ?"
		args = append(args, *role)
	}

	query += " ORDER BY started_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var agent Agent
		var taskID sql.NullString

		err := rows.Scan(
			&agent.ID,
			&agent.Role,
			&agent.Status,
			&taskID,
			&agent.PID,
			&agent.StartedAt,
			&agent.LastHeartbeat,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}

		if taskID.Valid {
			agent.TaskID = &taskID.String
		}

		agents = append(agents, agent)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating agents: %w", err)
	}

	return agents, nil
}

// FindDeadAgents finds agents that haven't sent a heartbeat within the timeout period
func (r *Registry) FindDeadAgents(ctx context.Context, timeout time.Duration) ([]Agent, error) {
	cutoff := time.Now().Add(-timeout)

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT agent_id, agent_role, status, task_id, pid, started_at, last_heartbeat
		 FROM agents
		 WHERE last_heartbeat < ? AND status != ?
		 ORDER BY last_heartbeat ASC`,
		cutoff, StatusDead,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find dead agents: %w", err)
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var agent Agent
		var taskID sql.NullString

		err := rows.Scan(
			&agent.ID,
			&agent.Role,
			&agent.Status,
			&taskID,
			&agent.PID,
			&agent.StartedAt,
			&agent.LastHeartbeat,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}

		if taskID.Valid {
			agent.TaskID = &taskID.String
		}

		agents = append(agents, agent)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating dead agents: %w", err)
	}

	return agents, nil
}

// MarkDead marks an agent as dead
func (r *Registry) MarkDead(ctx context.Context, agentID string) error {
	return r.UpdateStatus(ctx, agentID, StatusDead)
}

// CleanupDeadAgents removes agents marked as dead from the registry
func (r *Registry) CleanupDeadAgents(ctx context.Context) (int, error) {
	result, err := r.db.ExecContext(
		ctx,
		"DELETE FROM agents WHERE status = ?",
		StatusDead,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup dead agents: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to check affected rows: %w", err)
	}

	return int(rows), nil
}
