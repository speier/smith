package locks

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/speier/smith/internal/eventbus"
	"github.com/speier/smith/internal/registry"
	"github.com/speier/smith/internal/storage"
)

func setupTestDB(t *testing.T) (storage.Store, *registry.Registry, func()) {
	tmpDir, err := os.MkdirTemp("", "smith-locks-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	store, err := storage.InitProjectStorage(tmpDir)
	if err != nil {
		t.Fatalf("failed to init storage: %v", err)
	}

	reg := registry.New(store)

	cleanup := func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}

	return store, reg, cleanup
}

func TestAcquireAndRelease(t *testing.T) {
	store, reg, cleanup := setupTestDB(t)
	defer cleanup()

	manager := New(store)
	ctx := context.Background()

	// Register agent first
	err := reg.Register(ctx, "agent-1", eventbus.RoleImplementation, 12345)
	if err != nil {
		t.Fatalf("failed to register agent: %v", err)
	}

	// Acquire lock
	err = manager.Acquire(ctx, "/path/to/file.go", "agent-1", "task-1")
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	// Verify lock is held
	locked, err := manager.IsLocked(ctx, "/path/to/file.go")
	if err != nil {
		t.Fatalf("failed to check lock: %v", err)
	}
	if !locked {
		t.Error("file should be locked")
	}

	// Release lock
	err = manager.Release(ctx, "/path/to/file.go", "agent-1")
	if err != nil {
		t.Fatalf("failed to release lock: %v", err)
	}

	// Verify lock is released
	locked, err = manager.IsLocked(ctx, "/path/to/file.go")
	if err != nil {
		t.Fatalf("failed to check lock: %v", err)
	}
	if locked {
		t.Error("file should not be locked")
	}
}

func TestLockConflict(t *testing.T) {
	store, reg, cleanup := setupTestDB(t)
	defer cleanup()

	manager := New(store)
	ctx := context.Background()

	// Register agents
	reg.Register(ctx, "agent-1", eventbus.RoleImplementation, 12345)
	reg.Register(ctx, "agent-2", eventbus.RoleImplementation, 12346)

	// Agent 1 acquires lock
	err := manager.Acquire(ctx, "/path/to/file.go", "agent-1", "task-1")
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	// Agent 2 tries to acquire the same lock
	err = manager.Acquire(ctx, "/path/to/file.go", "agent-2", "task-2")
	if !errors.Is(err, ErrLockHeld) {
		t.Errorf("expected ErrLockHeld, got %v", err)
	}

	// Get lock info
	lock, err := manager.GetLock(ctx, "/path/to/file.go")
	if err != nil {
		t.Fatalf("failed to get lock: %v", err)
	}

	if lock.AgentID != "agent-1" {
		t.Errorf("expected agent-1, got %s", lock.AgentID)
	}
}

func TestSameAgentReacquire(t *testing.T) {
	store, reg, cleanup := setupTestDB(t)
	defer cleanup()

	manager := New(store)
	ctx := context.Background()

	// Register agent
	reg.Register(ctx, "agent-1", eventbus.RoleImplementation, 12345)

	// Acquire lock
	err := manager.Acquire(ctx, "/path/to/file.go", "agent-1", "task-1")
	if err != nil {
		t.Fatalf("failed to acquire lock: %v", err)
	}

	// Same agent tries to acquire again (should succeed)
	err = manager.Acquire(ctx, "/path/to/file.go", "agent-1", "task-1")
	if err != nil {
		t.Errorf("same agent should be able to reacquire lock: %v", err)
	}
}

func TestReleaseAll(t *testing.T) {
	store, reg, cleanup := setupTestDB(t)
	defer cleanup()

	manager := New(store)
	ctx := context.Background()

	// Register agent
	reg.Register(ctx, "agent-1", eventbus.RoleImplementation, 12345)

	// Acquire multiple locks for same agent
	files := []string{"/file1.go", "/file2.go", "/file3.go"}
	for _, file := range files {
		err := manager.Acquire(ctx, file, "agent-1", "task-1")
		if err != nil {
			t.Fatalf("failed to acquire lock on %s: %v", file, err)
		}
	}

	// Release all locks
	err := manager.ReleaseAll(ctx, "agent-1")
	if err != nil {
		t.Fatalf("failed to release all locks: %v", err)
	}

	// Verify all locks are released
	for _, file := range files {
		locked, err := manager.IsLocked(ctx, file)
		if err != nil {
			t.Fatalf("failed to check lock on %s: %v", file, err)
		}
		if locked {
			t.Errorf("file %s should not be locked", file)
		}
	}
}

func TestGetLockedFiles(t *testing.T) {
	store, reg, cleanup := setupTestDB(t)
	defer cleanup()

	manager := New(store)
	ctx := context.Background()

	// Register agent
	reg.Register(ctx, "agent-1", eventbus.RoleImplementation, 12345)

	// Acquire locks
	files := []string{"/file1.go", "/file2.go"}
	for _, file := range files {
		err := manager.Acquire(ctx, file, "agent-1", "task-1")
		if err != nil {
			t.Fatalf("failed to acquire lock: %v", err)
		}
	}

	// Get locked files
	locks, err := manager.GetLockedFiles(ctx, "agent-1")
	if err != nil {
		t.Fatalf("failed to get locked files: %v", err)
	}

	if len(locks) != 2 {
		t.Errorf("expected 2 locks, got %d", len(locks))
	}
}

func TestGetAllLocks(t *testing.T) {
	store, reg, cleanup := setupTestDB(t)
	defer cleanup()

	manager := New(store)
	ctx := context.Background()

	// Register agents
	reg.Register(ctx, "agent-1", eventbus.RoleImplementation, 12345)
	reg.Register(ctx, "agent-2", eventbus.RoleImplementation, 12346)
	reg.Register(ctx, "agent-3", eventbus.RoleImplementation, 12347)

	// Multiple agents acquire locks
	manager.Acquire(ctx, "/file1.go", "agent-1", "task-1")
	manager.Acquire(ctx, "/file2.go", "agent-2", "task-2")
	manager.Acquire(ctx, "/file3.go", "agent-3", "task-3")

	// Get all locks
	locks, err := manager.GetAllLocks(ctx)
	if err != nil {
		t.Fatalf("failed to get all locks: %v", err)
	}

	if len(locks) != 3 {
		t.Errorf("expected 3 locks, got %d", len(locks))
	}
}
