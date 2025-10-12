package coordinator

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/speier/smith/internal/eventbus"
	"github.com/speier/smith/internal/kanban"
)

// TestKanbanSyncIntegration tests the full bidirectional sync flow:
// kanban.md → DB → query → update → kanban.md
func TestKanbanSyncIntegration(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	
	// Create initial kanban.md with some tasks
	smithDir := filepath.Join(tmpDir, ".smith")
	if err := os.MkdirAll(smithDir, 0755); err != nil {
		t.Fatalf("failed to create .smith dir: %v", err)
	}
	
	kanbanPath := filepath.Join(smithDir, "kanban.md")
	initialKanban := `# Agent Kanban Board

## Backlog
<!-- Tasks waiting to be picked up -->

- [ ] task-001: Implement user authentication
- [ ] task-002: Add database migration
- [ ] task-003: Write API documentation

## WIP
<!-- Work in progress - tasks currently being worked on -->

## Review
<!-- Tasks pending review -->

## Done
<!-- Completed tasks -->

`
	if err := os.WriteFile(kanbanPath, []byte(initialKanban), 0644); err != nil {
		t.Fatalf("failed to write initial kanban.md: %v", err)
	}
	
	// Step 1: Initialize coordinator (should parse kanban.md → DB)
	coord, err := NewSQLite(tmpDir)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}
	defer coord.Close()
	
	// Verify tasks were synced to DB
	stats, err := coord.GetTaskStats()
	if err != nil {
		t.Fatalf("GetTaskStats failed: %v", err)
	}
	if stats.Backlog != 3 {
		t.Errorf("expected 3 backlog tasks after sync, got %d", stats.Backlog)
	}
	if stats.WIP != 0 {
		t.Errorf("expected 0 WIP tasks, got %d", stats.WIP)
	}
	
	// Step 2: Query available tasks from DB
	availableTasks, err := coord.GetAvailableTasks()
	if err != nil {
		t.Fatalf("GetAvailableTasks failed: %v", err)
	}
	if len(availableTasks) != 3 {
		t.Fatalf("expected 3 available tasks, got %d", len(availableTasks))
	}
	if availableTasks[0].ID != "task-001" {
		t.Errorf("expected first task to be task-001, got %s", availableTasks[0].ID)
	}
	
	// Step 3: Register an agent and claim a task (updates DB)
	ctx := context.Background()
	if err := coord.registry.Register(ctx, "agent-test", eventbus.RoleImplementation, 12345); err != nil {
		t.Fatalf("failed to register agent: %v", err)
	}
	
	if err := coord.ClaimTask("task-001", "agent-test"); err != nil {
		t.Fatalf("ClaimTask failed: %v", err)
	}
	
	// Verify DB was updated
	stats, err = coord.GetTaskStats()
	if err != nil {
		t.Fatalf("GetTaskStats failed: %v", err)
	}
	if stats.Backlog != 2 {
		t.Errorf("expected 2 backlog tasks after claim, got %d", stats.Backlog)
	}
	if stats.WIP != 1 {
		t.Errorf("expected 1 WIP task after claim, got %d", stats.WIP)
	}
	
	// Step 4: Verify kanban.md was synced back
	board, err := kanban.Parse(kanbanPath)
	if err != nil {
		t.Fatalf("failed to parse updated kanban.md: %v", err)
	}
	
	if len(board.Backlog) != 2 {
		t.Errorf("expected 2 tasks in Backlog section, got %d", len(board.Backlog))
	}
	if len(board.WIP) != 1 {
		t.Errorf("expected 1 task in WIP section, got %d", len(board.WIP))
	}
	if board.WIP[0].ID != "task-001" {
		t.Errorf("expected task-001 in WIP, got %s", board.WIP[0].ID)
	}
	if !board.WIP[0].Checked {
		t.Error("expected WIP task to be checked")
	}
	
	// Step 5: Manually update kanban.md (add a new task)
	updatedKanban := `# Agent Kanban Board

## Backlog
<!-- Tasks waiting to be picked up -->

- [ ] task-002: Add database migration
- [ ] task-003: Write API documentation
- [ ] task-004: New task added manually

## WIP
<!-- Work in progress - tasks currently being worked on -->

- [x] task-001: Implement user authentication

## Review
<!-- Tasks pending review -->

## Done
<!-- Completed tasks -->

`
	if err := os.WriteFile(kanbanPath, []byte(updatedKanban), 0644); err != nil {
		t.Fatalf("failed to write updated kanban.md: %v", err)
	}
	
	// Step 6: Re-sync from kanban.md to DB
	if err := coord.syncKanbanToDB(); err != nil {
		t.Fatalf("syncKanbanToDB failed: %v", err)
	}
	
	// Verify new task was added to DB
	stats, err = coord.GetTaskStats()
	if err != nil {
		t.Fatalf("GetTaskStats failed: %v", err)
	}
	if stats.Backlog != 3 {
		t.Errorf("expected 3 backlog tasks after re-sync, got %d", stats.Backlog)
	}
	
	availableTasks, err = coord.GetAvailableTasks()
	if err != nil {
		t.Fatalf("GetAvailableTasks failed: %v", err)
	}
	if len(availableTasks) != 3 {
		t.Errorf("expected 3 available tasks, got %d", len(availableTasks))
	}
	
	// Verify task-004 is in the list
	found := false
	for _, task := range availableTasks {
		if task.ID == "task-004" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find task-004 in available tasks")
	}
}

// TestKanbanSyncPreservesTaskOrder tests that task order is maintained
func TestKanbanSyncPreservesTaskOrder(t *testing.T) {
	tmpDir := t.TempDir()
	smithDir := filepath.Join(tmpDir, ".smith")
	if err := os.MkdirAll(smithDir, 0755); err != nil {
		t.Fatalf("failed to create .smith dir: %v", err)
	}
	
	kanbanPath := filepath.Join(smithDir, "kanban.md")
	kanbanContent := `# Agent Kanban Board

## Backlog

- [ ] task-alpha: First task
- [ ] task-beta: Second task
- [ ] task-gamma: Third task

## WIP

## Review

## Done

`
	if err := os.WriteFile(kanbanPath, []byte(kanbanContent), 0644); err != nil {
		t.Fatalf("failed to write kanban.md: %v", err)
	}
	
	coord, err := NewSQLite(tmpDir)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}
	defer coord.Close()
	
	// Get available tasks
	tasks, err := coord.GetAvailableTasks()
	if err != nil {
		t.Fatalf("GetAvailableTasks failed: %v", err)
	}
	
	// Verify order (should be by started_at which reflects insertion order)
	expectedOrder := []string{"task-alpha", "task-beta", "task-gamma"}
	for i, task := range tasks {
		if task.ID != expectedOrder[i] {
			t.Errorf("task %d: expected %s, got %s", i, expectedOrder[i], task.ID)
		}
	}
}

// TestKanbanSyncEmptySections tests handling of empty kanban sections
func TestKanbanSyncEmptySections(t *testing.T) {
	tmpDir := t.TempDir()
	smithDir := filepath.Join(tmpDir, ".smith")
	if err := os.MkdirAll(smithDir, 0755); err != nil {
		t.Fatalf("failed to create .smith dir: %v", err)
	}
	
	kanbanPath := filepath.Join(smithDir, "kanban.md")
	kanbanContent := `# Agent Kanban Board

## Backlog

## WIP

## Review

## Done

`
	if err := os.WriteFile(kanbanPath, []byte(kanbanContent), 0644); err != nil {
		t.Fatalf("failed to write kanban.md: %v", err)
	}
	
	coord, err := NewSQLite(tmpDir)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}
	defer coord.Close()
	
	// Should handle empty kanban gracefully
	stats, err := coord.GetTaskStats()
	if err != nil {
		t.Fatalf("GetTaskStats failed: %v", err)
	}
	
	if stats.Backlog != 0 || stats.WIP != 0 || stats.Review != 0 || stats.Done != 0 {
		t.Errorf("expected all stats to be 0, got: backlog=%d, wip=%d, review=%d, done=%d",
			stats.Backlog, stats.WIP, stats.Review, stats.Done)
	}
	
	// Sync back should still work
	if err := coord.syncDBToKanban(); err != nil {
		t.Fatalf("syncDBToKanban failed: %v", err)
	}
	
	// Verify kanban.md still exists and is valid
	board, err := kanban.Parse(kanbanPath)
	if err != nil {
		t.Fatalf("failed to parse kanban.md: %v", err)
	}
	
	if len(board.AllTasks()) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(board.AllTasks()))
	}
}
