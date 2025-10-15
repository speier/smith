package coordinator

import (
	"context"
	"os"
	"testing"

	"github.com/speier/smith/internal/eventbus"
)

func TestSQLiteCoordinator(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "smith-sqlite-coord-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create SQLite coordinator
	coord, err := NewSQLite(tmpDir)
	if err != nil {
		t.Fatalf("NewSQLite failed: %v", err)
	}
	defer coord.Close()

	// Test EnsureDirectories
	if err := coord.EnsureDirectories(); err != nil {
		t.Errorf("EnsureDirectories failed: %v", err)
	}

	// Test GetTaskStats (should be empty initially)
	stats, err := coord.GetTaskStats()
	if err != nil {
		t.Errorf("GetTaskStats failed: %v", err)
	}
	if stats.Backlog != 0 {
		t.Errorf("expected 0 backlog tasks, got %d", stats.Backlog)
	}

	// Add a task to the database
	ctx := context.Background()
	_, err = coord.db.ExecContext(ctx, `
		INSERT INTO task_assignments (task_id, agent_id, agent_role, status)
		VALUES ('task-1', NULL, NULL, 'backlog')
	`)
	if err != nil {
		t.Fatalf("failed to insert test task: %v", err)
	}

	// Test ClaimTask
	// First, register an agent
	if err := coord.registry.Register(ctx, "agent-1", eventbus.RoleImplementation, 12345); err != nil {
		t.Fatalf("failed to register agent: %v", err)
	}

	if err := coord.ClaimTask("task-1", "agent-1"); err != nil {
		t.Errorf("ClaimTask failed: %v", err)
	}

	// Verify task was claimed (should be in WIP)
	stats, err = coord.GetTaskStats()
	if err != nil {
		t.Errorf("GetTaskStats failed: %v", err)
	}
	if stats.WIP != 1 {
		t.Errorf("expected 1 WIP task after claiming, got %d", stats.WIP)
	}

	// Test LockFiles
	files := []string{"test.go", "main.go"}
	if err := coord.LockFiles("task-1", "agent-1", files); err != nil {
		t.Errorf("LockFiles failed: %v", err)
	}

	// Verify locks were acquired
	locks, err := coord.GetActiveLocks()
	if err != nil {
		t.Errorf("GetActiveLocks failed: %v", err)
	}
	if len(locks) != 2 {
		t.Errorf("expected 2 locks, got %d", len(locks))
	}

	// Test GetMessages (should only contain agent communication events)
	messages, err := coord.GetMessages()
	if err != nil {
		t.Errorf("GetMessages failed: %v", err)
	}
	// Should be 0 since we haven't sent any agent messages
	if len(messages) != 0 {
		t.Errorf("expected 0 messages (no agent communication sent), got %d", len(messages))
	}

	// Verify events were published by checking event bus directly
	events, err := coord.eventBus.Query(ctx, eventbus.EventFilter{SinceID: 0})
	if err != nil {
		t.Errorf("failed to query events: %v", err)
	}
	// Should have task_claimed + 2 file_locked events = 3 events
	if len(events) < 3 {
		t.Errorf("expected at least 3 events (1 task_claimed + 2 file_locked), got %d", len(events))
	}
}

func TestSQLiteCoordinatorLockConflict(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "smith-sqlite-coord-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	coord, err := NewSQLite(tmpDir)
	if err != nil {
		t.Fatalf("NewSQLite failed: %v", err)
	}
	defer coord.Close()

	ctx := context.Background()

	// Register two agents
	coord.registry.Register(ctx, "agent-1", eventbus.RoleImplementation, 12345)
	coord.registry.Register(ctx, "agent-2", eventbus.RoleImplementation, 12346)

	// Agent 1 locks a file
	if err := coord.LockFiles("task-1", "agent-1", []string{"shared.go"}); err != nil {
		t.Fatalf("agent-1 failed to lock file: %v", err)
	}

	// Agent 2 tries to lock the same file - should fail
	err = coord.LockFiles("task-2", "agent-2", []string{"shared.go"})
	if err == nil {
		t.Error("agent-2 should not be able to lock file already locked by agent-1")
	}
}
