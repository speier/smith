package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// SQLiteStore implements the Store interface using SQLite
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a SQLite-backed store
func NewSQLiteStore(db *sql.DB) *SQLiteStore {
	return &SQLiteStore{db: db}
}

// === EventStore Implementation ===

func (s *SQLiteStore) SaveEvent(ctx context.Context, event *Event) error {
	query := `
		INSERT INTO events (agent_id, agent_role, event_type, task_id, file_path, data)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := s.db.ExecContext(ctx, query,
		event.AgentID,
		event.AgentRole,
		event.EventType,
		event.TaskID,
		event.FilePath,
		event.Data,
	)
	if err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get event ID: %w", err)
	}
	event.ID = id

	return nil
}

func (s *SQLiteStore) QueryEvents(ctx context.Context, filter EventFilter) ([]*Event, error) {
	query := "SELECT id, timestamp, agent_id, agent_role, event_type, task_id, file_path, data FROM events WHERE 1=1"
	args := []interface{}{}

	if len(filter.EventTypes) > 0 {
		placeholders := make([]string, len(filter.EventTypes))
		for i, et := range filter.EventTypes {
			placeholders[i] = "?"
			args = append(args, et)
		}
		query += " AND event_type IN (" + strings.Join(placeholders, ",") + ")"
	}

	if filter.AgentID != nil {
		query += " AND agent_id = ?"
		args = append(args, *filter.AgentID)
	}

	if filter.TaskID != nil {
		query += " AND task_id = ?"
		args = append(args, *filter.TaskID)
	}

	query += " ORDER BY timestamp DESC LIMIT 1000"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		event := &Event{}
		var taskID, filePath sql.NullString
		err := rows.Scan(
			&event.ID,
			&event.Timestamp,
			&event.AgentID,
			&event.AgentRole,
			&event.EventType,
			&taskID,
			&filePath,
			&event.Data,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		if taskID.Valid {
			event.TaskID = &taskID.String
		}
		if filePath.Valid {
			event.FilePath = &filePath.String
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

// === AgentStore Implementation ===

func (s *SQLiteStore) RegisterAgent(ctx context.Context, agent *Agent) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO agents (agent_id, agent_role, status, pid, started_at, last_heartbeat)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(agent_id) DO UPDATE SET
			agent_role = excluded.agent_role,
			status = excluded.status,
			pid = excluded.pid,
			started_at = excluded.started_at,
			last_heartbeat = excluded.last_heartbeat`,
		agent.ID, agent.Role, agent.Status, agent.PID, agent.StartedAt, agent.LastHeartbeat,
	)
	if err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}
	return nil
}

func (s *SQLiteStore) UpdateHeartbeat(ctx context.Context, agentID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE agents SET last_heartbeat = CURRENT_TIMESTAMP WHERE agent_id = ?`,
		agentID,
	)
	if err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}
	return nil
}

func (s *SQLiteStore) UnregisterAgent(ctx context.Context, agentID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM agents WHERE agent_id = ?`,
		agentID,
	)
	if err != nil {
		return fmt.Errorf("failed to unregister agent: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetAgent(ctx context.Context, agentID string) (*Agent, error) {
	agent := &Agent{}
	var taskID sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT agent_id, agent_role, status, task_id, pid, started_at, last_heartbeat
		 FROM agents WHERE agent_id = ?`,
		agentID,
	).Scan(&agent.ID, &agent.Role, &agent.Status, &taskID, &agent.PID, &agent.StartedAt, &agent.LastHeartbeat)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	if taskID.Valid {
		agent.TaskID = &taskID.String
	}
	return agent, nil
}

func (s *SQLiteStore) ListAgents(ctx context.Context, role *string) ([]*Agent, error) {
	query := "SELECT agent_id, agent_role, status, task_id, pid, started_at, last_heartbeat FROM agents"
	args := []interface{}{}

	if role != nil {
		query += " WHERE agent_role = ?"
		args = append(args, *role)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}
	defer rows.Close()

	var agents []*Agent
	for rows.Next() {
		agent := &Agent{}
		var taskID sql.NullString
		err := rows.Scan(&agent.ID, &agent.Role, &agent.Status, &taskID, &agent.PID, &agent.StartedAt, &agent.LastHeartbeat)
		if err != nil {
			return nil, fmt.Errorf("failed to scan agent: %w", err)
		}
		if taskID.Valid {
			agent.TaskID = &taskID.String
		}
		agents = append(agents, agent)
	}

	return agents, rows.Err()
}

func (s *SQLiteStore) MarkAgentDead(ctx context.Context, timeout time.Duration) (int, error) {
	deadline := time.Now().Add(-timeout)
	result, err := s.db.ExecContext(ctx,
		`UPDATE agents SET status = 'dead' 
		 WHERE status = 'active' AND last_heartbeat < ?`,
		deadline,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to mark agents dead: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return int(count), nil
}

// === TaskStore Implementation ===

func (s *SQLiteStore) CreateTask(ctx context.Context, task *Task) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO task_assignments (task_id, title, description, agent_role, status, started_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		task.TaskID, task.Title, task.Description, task.AgentRole, task.Status, task.StartedAt, task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

func (s *SQLiteStore) GetTask(ctx context.Context, taskID string) (*Task, error) {
	task := &Task{}
	var agentID, agentRole, result, errMsg sql.NullString
	var completedAt sql.NullTime

	err := s.db.QueryRowContext(ctx,
		`SELECT task_id, title, description, agent_id, agent_role, status, result, error, 
		        started_at, updated_at, completed_at
		 FROM task_assignments WHERE task_id = ?`,
		taskID,
	).Scan(&task.TaskID, &task.Title, &task.Description, &agentID, &agentRole, &task.Status,
		&result, &errMsg, &task.StartedAt, &task.UpdatedAt, &completedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	if agentID.Valid {
		task.AgentID = agentID.String
	}
	if agentRole.Valid {
		task.AgentRole = agentRole.String
	}
	if result.Valid {
		task.Result = result.String
	}
	if errMsg.Valid {
		task.Error = errMsg.String
	}
	if completedAt.Valid {
		task.CompletedAt = &completedAt.Time
	}

	return task, nil
}

func (s *SQLiteStore) UpdateTask(ctx context.Context, task *Task) error {
	// Convert empty strings to NULL for foreign key fields
	var agentID interface{} = task.AgentID
	if task.AgentID == "" {
		agentID = nil
	}

	_, err := s.db.ExecContext(ctx,
		`UPDATE task_assignments 
		 SET agent_id = ?, status = ?, result = ?, error = ?, updated_at = ?, completed_at = ?
		 WHERE task_id = ?`,
		agentID, task.Status, task.Result, task.Error, task.UpdatedAt, task.CompletedAt, task.TaskID,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	return nil
}

func (s *SQLiteStore) ListTasks(ctx context.Context, status *string) ([]*Task, error) {
	query := `SELECT task_id, title, description, agent_id, agent_role, status, result, error,
	                 started_at, updated_at, completed_at FROM task_assignments`
	args := []interface{}{}

	if status != nil {
		query += " WHERE status = ?"
		args = append(args, *status)
	}

	query += " ORDER BY started_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		var agentID, agentRole, result, errMsg sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(&task.TaskID, &task.Title, &task.Description, &agentID, &agentRole,
			&task.Status, &result, &errMsg, &task.StartedAt, &task.UpdatedAt, &completedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		if agentID.Valid {
			task.AgentID = agentID.String
		}
		if agentRole.Valid {
			task.AgentRole = agentRole.String
		}
		if result.Valid {
			task.Result = result.String
		}
		if errMsg.Valid {
			task.Error = errMsg.String
		}
		if completedAt.Valid {
			task.CompletedAt = &completedAt.Time
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}

func (s *SQLiteStore) ClaimTask(ctx context.Context, taskID, agentID string) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE task_assignments 
		 SET agent_id = ?, status = 'wip', updated_at = CURRENT_TIMESTAMP
		 WHERE task_id = ? AND status = 'backlog'`,
		agentID, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to claim task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task %s is not claimable", taskID)
	}

	return nil
}

func (s *SQLiteStore) GetTaskStats(ctx context.Context) (*TaskStats, error) {
	stats := &TaskStats{}
	err := s.db.QueryRowContext(ctx,
		`SELECT 
			COALESCE(SUM(CASE WHEN status = 'backlog' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'wip' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'review' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN status = 'done' THEN 1 ELSE 0 END), 0)
		 FROM task_assignments`,
	).Scan(&stats.Backlog, &stats.WIP, &stats.Review, &stats.Done)

	if err != nil {
		return nil, fmt.Errorf("failed to get task stats: %w", err)
	}

	return stats, nil
}

// === LockStore Implementation ===

func (s *SQLiteStore) AcquireLocks(ctx context.Context, locks []*FileLock) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check for existing locks
	for _, lock := range locks {
		var count int
		err := tx.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM file_locks WHERE file_path = ?`,
			lock.FilePath,
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check lock: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("file %s is already locked", lock.FilePath)
		}
	}

	// Acquire all locks
	for _, lock := range locks {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO file_locks (file_path, agent_id, task_id, locked_at)
			 VALUES (?, ?, ?, ?)`,
			lock.FilePath, lock.AgentID, lock.TaskID, lock.LockedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit locks: %w", err)
	}

	return nil
}

func (s *SQLiteStore) ReleaseLocks(ctx context.Context, agentID string, files []string) error {
	if len(files) == 0 {
		// Release all locks for agent
		_, err := s.db.ExecContext(ctx,
			`DELETE FROM file_locks WHERE agent_id = ?`,
			agentID,
		)
		if err != nil {
			return fmt.Errorf("failed to release locks: %w", err)
		}
		return nil
	}

	// Release specific files
	placeholders := make([]string, len(files))
	args := []interface{}{agentID}
	for i, file := range files {
		placeholders[i] = "?"
		args = append(args, file)
	}

	query := fmt.Sprintf(
		`DELETE FROM file_locks WHERE agent_id = ? AND file_path IN (%s)`,
		strings.Join(placeholders, ","),
	)

	_, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to release locks: %w", err)
	}

	return nil
}

func (s *SQLiteStore) GetLocks(ctx context.Context) ([]*FileLock, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT file_path, agent_id, task_id, locked_at FROM file_locks ORDER BY locked_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get locks: %w", err)
	}
	defer rows.Close()

	var locks []*FileLock
	for rows.Next() {
		lock := &FileLock{}
		err := rows.Scan(&lock.FilePath, &lock.AgentID, &lock.TaskID, &lock.LockedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lock: %w", err)
		}
		locks = append(locks, lock)
	}

	return locks, rows.Err()
}

func (s *SQLiteStore) GetLocksForAgent(ctx context.Context, agentID string) ([]*FileLock, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT file_path, agent_id, task_id, locked_at FROM file_locks WHERE agent_id = ? ORDER BY locked_at DESC`,
		agentID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get locks for agent: %w", err)
	}
	defer rows.Close()

	var locks []*FileLock
	for rows.Next() {
		lock := &FileLock{}
		err := rows.Scan(&lock.FilePath, &lock.AgentID, &lock.TaskID, &lock.LockedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan lock: %w", err)
		}
		locks = append(locks, lock)
	}

	return locks, rows.Err()
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
