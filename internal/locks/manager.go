package locks

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	// ErrLockHeld indicates the file is already locked by another agent
	ErrLockHeld = errors.New("file is already locked")

	// ErrLockNotFound indicates the lock doesn't exist
	ErrLockNotFound = errors.New("lock not found")
)

// FileLock represents a lock on a file
type FileLock struct {
	FilePath string
	AgentID  string
	TaskID   string
	LockedAt time.Time
}

// Manager handles file locking coordination between agents
type Manager struct {
	db *sql.DB
}

// New creates a new lock Manager
func New(db *sql.DB) *Manager {
	return &Manager{db: db}
}

// Acquire attempts to acquire a lock on a file
func (m *Manager) Acquire(ctx context.Context, filePath, agentID, taskID string) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if lock already exists
	var existingAgentID string
	err = tx.QueryRowContext(
		ctx,
		"SELECT agent_id FROM file_locks WHERE file_path = ?",
		filePath,
	).Scan(&existingAgentID)

	if err == nil {
		// Lock exists
		if existingAgentID == agentID {
			// Same agent already holds the lock, that's okay
			return nil
		}
		return ErrLockHeld
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing lock: %w", err)
	}

	// Insert new lock
	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO file_locks (file_path, agent_id, task_id) VALUES (?, ?, ?)",
		filePath, agentID, taskID,
	)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Release releases a lock on a file
func (m *Manager) Release(ctx context.Context, filePath, agentID string) error {
	result, err := m.db.ExecContext(
		ctx,
		"DELETE FROM file_locks WHERE file_path = ? AND agent_id = ?",
		filePath, agentID,
	)
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rows == 0 {
		return ErrLockNotFound
	}

	return nil
}

// ReleaseAll releases all locks held by an agent
func (m *Manager) ReleaseAll(ctx context.Context, agentID string) error {
	_, err := m.db.ExecContext(
		ctx,
		"DELETE FROM file_locks WHERE agent_id = ?",
		agentID,
	)
	if err != nil {
		return fmt.Errorf("failed to release all locks: %w", err)
	}

	return nil
}

// IsLocked checks if a file is currently locked
func (m *Manager) IsLocked(ctx context.Context, filePath string) (bool, error) {
	var count int
	err := m.db.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM file_locks WHERE file_path = ?",
		filePath,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check lock status: %w", err)
	}

	return count > 0, nil
}

// GetLock retrieves information about a file lock
func (m *Manager) GetLock(ctx context.Context, filePath string) (*FileLock, error) {
	var lock FileLock
	err := m.db.QueryRowContext(
		ctx,
		"SELECT file_path, agent_id, task_id, locked_at FROM file_locks WHERE file_path = ?",
		filePath,
	).Scan(&lock.FilePath, &lock.AgentID, &lock.TaskID, &lock.LockedAt)

	if err == sql.ErrNoRows {
		return nil, ErrLockNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get lock: %w", err)
	}

	return &lock, nil
}

// GetLockedFiles returns all files currently locked by an agent
func (m *Manager) GetLockedFiles(ctx context.Context, agentID string) ([]FileLock, error) {
	rows, err := m.db.QueryContext(
		ctx,
		"SELECT file_path, agent_id, task_id, locked_at FROM file_locks WHERE agent_id = ?",
		agentID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query locked files: %w", err)
	}
	defer rows.Close()

	var locks []FileLock
	for rows.Next() {
		var lock FileLock
		if err := rows.Scan(&lock.FilePath, &lock.AgentID, &lock.TaskID, &lock.LockedAt); err != nil {
			return nil, fmt.Errorf("failed to scan lock: %w", err)
		}
		locks = append(locks, lock)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating locks: %w", err)
	}

	return locks, nil
}

// GetAllLocks returns all active file locks
func (m *Manager) GetAllLocks(ctx context.Context) ([]FileLock, error) {
	rows, err := m.db.QueryContext(
		ctx,
		"SELECT file_path, agent_id, task_id, locked_at FROM file_locks ORDER BY locked_at ASC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query all locks: %w", err)
	}
	defer rows.Close()

	var locks []FileLock
	for rows.Next() {
		var lock FileLock
		if err := rows.Scan(&lock.FilePath, &lock.AgentID, &lock.TaskID, &lock.LockedAt); err != nil {
			return nil, fmt.Errorf("failed to scan lock: %w", err)
		}
		locks = append(locks, lock)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating locks: %w", err)
	}

	return locks, nil
}
