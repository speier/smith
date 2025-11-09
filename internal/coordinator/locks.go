package coordinator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/speier/smith/internal/storage"
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
	store storage.LockStore
}

// NewLockManager creates a new lock Manager
func NewLockManager(store storage.LockStore) *Manager {
	return &Manager{store: store}
}

// Acquire attempts to acquire a lock on a file
func (m *Manager) Acquire(ctx context.Context, filePath, agentID, taskID string) error {
	lock := &storage.FileLock{
		FilePath: filePath,
		AgentID:  agentID,
		TaskID:   taskID,
		LockedAt: time.Now(),
	}

	// AcquireLocks handles the check and insert atomically
	err := m.store.AcquireLocks(ctx, []*storage.FileLock{lock})
	if err != nil {
		// Check if it's already locked by this agent
		locks, getErr := m.store.GetLocks(ctx)
		if getErr == nil {
			for _, l := range locks {
				if l.FilePath == filePath && l.AgentID == agentID {
					// Same agent already holds the lock, that's okay
					return nil
				}
			}
		}
		return ErrLockHeld
	}

	return nil
}

// Release releases a lock on a file
func (m *Manager) Release(ctx context.Context, filePath, agentID string) error {
	if err := m.store.ReleaseLocks(ctx, agentID, []string{filePath}); err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	return nil
}

// ReleaseAll releases all locks held by an agent
func (m *Manager) ReleaseAll(ctx context.Context, agentID string) error {
	if err := m.store.ReleaseLocks(ctx, agentID, nil); err != nil {
		return fmt.Errorf("failed to release all locks: %w", err)
	}

	return nil
}

// IsLocked checks if a file is currently locked
func (m *Manager) IsLocked(ctx context.Context, filePath string) (bool, error) {
	locks, err := m.store.GetLocks(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check lock: %w", err)
	}

	for _, lock := range locks {
		if lock.FilePath == filePath {
			return true, nil
		}
	}

	return false, nil
}

// GetLock retrieves lock information for a file
func (m *Manager) GetLock(ctx context.Context, filePath string) (*FileLock, error) {
	locks, err := m.store.GetLocks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get lock: %w", err)
	}

	for _, lock := range locks {
		if lock.FilePath == filePath {
			return &FileLock{
				FilePath: lock.FilePath,
				AgentID:  lock.AgentID,
				TaskID:   lock.TaskID,
				LockedAt: lock.LockedAt,
			}, nil
		}
	}

	return nil, ErrLockNotFound
}

// GetLockedFiles returns all files currently locked by an agent
func (m *Manager) GetLockedFiles(ctx context.Context, agentID string) ([]FileLock, error) {
	storageLocks, err := m.store.GetLocksForAgent(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query locked files: %w", err)
	}

	var locks []FileLock
	for _, l := range storageLocks {
		lock := FileLock{
			FilePath: l.FilePath,
			AgentID:  l.AgentID,
			TaskID:   l.TaskID,
			LockedAt: l.LockedAt,
		}
		locks = append(locks, lock)
	}

	return locks, nil
}

// GetAllLocks returns all active file locks
func (m *Manager) GetAllLocks(ctx context.Context) ([]FileLock, error) {
	storageLocks, err := m.store.GetLocks(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query all locks: %w", err)
	}

	var locks []FileLock
	for _, l := range storageLocks {
		lock := FileLock{
			FilePath: l.FilePath,
			AgentID:  l.AgentID,
			TaskID:   l.TaskID,
			LockedAt: l.LockedAt,
		}
		locks = append(locks, lock)
	}

	return locks, nil
}
