package coordinator

import (
	"context"
	"testing"
	"time"

	"github.com/speier/smith/internal/eventbus"
	"github.com/speier/smith/internal/storage"
)

// TestFullTaskFlow tests the complete lifecycle:
// 1. Create tasks (simulating REPL input: "do X and Y")
// 2. Agents claim tasks from queue
// 3. Agents execute and coordinate (avoid conflicts)
// 4. Agents complete tasks
func TestFullTaskFlow(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	coord, err := NewBolt(tmpDir)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}
	defer coord.Close()

	ctx := context.Background()

	// Step 1: User says "implement user auth and add tests" in REPL
	// This creates 2 tasks in the backlog
	taskID1, err := coord.CreateTask(
		"Implement user authentication",
		"Add login/logout endpoints with JWT tokens",
		"implementation",
	)
	if err != nil {
		t.Fatalf("failed to create task 1: %v", err)
	}

	taskID2, err := coord.CreateTask(
		"Add authentication tests",
		"Write unit and integration tests for auth endpoints",
		"testing",
	)
	if err != nil {
		t.Fatalf("failed to create task 2: %v", err)
	}

	t.Logf("Created tasks: %s, %s", taskID1, taskID2)

	// Verify tasks are in backlog
	stats, err := coord.GetTaskStats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.Backlog != 2 {
		t.Errorf("expected 2 backlog tasks, got %d", stats.Backlog)
	}

	// Step 2: Background agent picks up task from queue
	availableTasks, err := coord.GetAvailableTasks()
	if err != nil {
		t.Fatalf("failed to get available tasks: %v", err)
	}
	if len(availableTasks) != 2 {
		t.Fatalf("expected 2 available tasks, got %d", len(availableTasks))
	}

	// Verify task details
	task1, err := coord.GetTask(taskID1)
	if err != nil {
		t.Fatalf("failed to get task 1: %v", err)
	}
	if task1.Title != "Implement user authentication" {
		t.Errorf("unexpected title: %s", task1.Title)
	}
	if task1.Description != "Add login/logout endpoints with JWT tokens" {
		t.Errorf("unexpected description: %s", task1.Description)
	}
	if task1.Role != "implementation" {
		t.Errorf("unexpected role: %s", task1.Role)
	}
	if task1.Status != "backlog" {
		t.Errorf("unexpected status: %s", task1.Status)
	}

	// Step 3: Register implementation agent
	agent1 := "agent-impl-001"
	if err := coord.registry.Register(ctx, agent1, eventbus.RoleImplementation, 12345); err != nil {
		t.Fatalf("failed to register agent 1: %v", err)
	}

	// Step 4: Agent claims first task
	if err := coord.ClaimTask(taskID1, agent1); err != nil {
		t.Fatalf("failed to claim task 1: %v", err)
	}

	// Verify task is now WIP
	task1, err = coord.GetTask(taskID1)
	if err != nil {
		t.Fatalf("failed to get task 1 after claim: %v", err)
	}
	if task1.Status != "wip" {
		t.Errorf("expected status 'wip', got %s", task1.Status)
	}
	if task1.AgentID != agent1 {
		t.Errorf("expected agent %s, got %s", agent1, task1.AgentID)
	}

	// Verify stats updated
	stats, err = coord.GetTaskStats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.Backlog != 1 {
		t.Errorf("expected 1 backlog task, got %d", stats.Backlog)
	}
	if stats.WIP != 1 {
		t.Errorf("expected 1 WIP task, got %d", stats.WIP)
	}

	// Step 5: Second agent registers and claims second task
	agent2 := "agent-test-001"
	if err := coord.registry.Register(ctx, agent2, eventbus.RoleTesting, 12346); err != nil {
		t.Fatalf("failed to register agent 2: %v", err)
	}

	if err := coord.ClaimTask(taskID2, agent2); err != nil {
		t.Fatalf("failed to claim task 2: %v", err)
	}

	// Verify both tasks are WIP
	stats, err = coord.GetTaskStats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.Backlog != 0 {
		t.Errorf("expected 0 backlog tasks, got %d", stats.Backlog)
	}
	if stats.WIP != 2 {
		t.Errorf("expected 2 WIP tasks, got %d", stats.WIP)
	}

	// Step 6: Agents coordinate via EventBus (simulate file lock coordination)
	// Agent 1 locks files it's working on
	files := []string{"auth.go", "handlers.go"}
	if err := coord.LockFiles(taskID1, agent1, files); err != nil {
		t.Fatalf("failed to lock files: %v", err)
	}

	// Agent 2 works on different files (no conflict)
	testFiles := []string{"auth_test.go", "handlers_test.go"}
	if err := coord.LockFiles(taskID2, agent2, testFiles); err != nil {
		t.Fatalf("failed to lock test files: %v", err)
	}

	// Verify locks are active
	locks, err := coord.GetActiveLocks()
	if err != nil {
		t.Fatalf("failed to get locks: %v", err)
	}
	// We locked 2 files for agent1 + 2 files for agent2 = 4 total file locks
	if len(locks) != 4 {
		t.Errorf("expected 4 active file locks, got %d", len(locks))
	}

	// Step 7: Agent 1 completes its task
	result1 := "Implemented login/logout endpoints with JWT authentication"
	if err := coord.CompleteTask(taskID1, result1); err != nil {
		t.Fatalf("failed to complete task 1: %v", err)
	}

	// Verify task is done
	task1, err = coord.GetTask(taskID1)
	if err != nil {
		t.Fatalf("failed to get task 1 after completion: %v", err)
	}
	if task1.Status != "done" {
		t.Errorf("expected status 'done', got %s", task1.Status)
	}
	if task1.Result != result1 {
		t.Errorf("unexpected result: %s", task1.Result)
	}

	// Step 8: Agent 2 moves task to review
	if err := coord.UpdateTaskStatus(taskID2, "review"); err != nil {
		t.Fatalf("failed to update task 2 status: %v", err)
	}

	// Verify stats
	stats, err = coord.GetTaskStats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.WIP != 0 {
		t.Errorf("expected 0 WIP tasks, got %d", stats.WIP)
	}
	if stats.Review != 1 {
		t.Errorf("expected 1 review task, got %d", stats.Review)
	}
	if stats.Done != 1 {
		t.Errorf("expected 1 done task, got %d", stats.Done)
	}

	// Step 9: Agent 2 completes after review
	result2 := "Added comprehensive test suite for authentication"
	if err := coord.CompleteTask(taskID2, result2); err != nil {
		t.Fatalf("failed to complete task 2: %v", err)
	}

	// Final verification
	stats, err = coord.GetTaskStats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.Done != 2 {
		t.Errorf("expected 2 done tasks, got %d", stats.Done)
	}
	if stats.Backlog != 0 || stats.WIP != 0 || stats.Review != 0 {
		t.Errorf("expected all tasks done, got: backlog=%d, wip=%d, review=%d",
			stats.Backlog, stats.WIP, stats.Review)
	}

	t.Log("✅ Full task flow completed successfully!")
}

// TestTaskFailureAndRetry tests error handling:
// 1. Create task
// 2. Agent claims and fails
// 3. Task goes back to backlog
// 4. Another agent retries successfully
func TestTaskFailureAndRetry(t *testing.T) {
	tmpDir := t.TempDir()
	coord, err := NewBolt(tmpDir)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}
	defer coord.Close()

	ctx := context.Background()

	// Create a task
	taskID, err := coord.CreateTask(
		"Refactor database layer",
		"Extract database code into separate package",
		"implementation",
	)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Agent 1 claims task
	agent1 := "agent-fail-001"
	if err := coord.registry.Register(ctx, agent1, eventbus.RoleImplementation, 99999); err != nil {
		t.Fatalf("failed to register agent: %v", err)
	}

	if err := coord.ClaimTask(taskID, agent1); err != nil {
		t.Fatalf("failed to claim task: %v", err)
	}

	// Verify task is WIP
	task, err := coord.GetTask(taskID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if task.Status != "wip" {
		t.Errorf("expected status 'wip', got %s", task.Status)
	}

	// Agent encounters error and fails the task
	errorMsg := "Database connection timeout - couldn't complete refactoring"
	if err := coord.FailTask(taskID, errorMsg); err != nil {
		t.Fatalf("failed to fail task: %v", err)
	}

	// Verify task is back in backlog with error recorded
	task, err = coord.GetTask(taskID)
	if err != nil {
		t.Fatalf("failed to get task after failure: %v", err)
	}
	if task.Status != "backlog" {
		t.Errorf("expected status 'backlog', got %s", task.Status)
	}
	if task.Error != errorMsg {
		t.Errorf("expected error message '%s', got '%s'", errorMsg, task.Error)
	}
	if task.AgentID != "" {
		t.Errorf("expected agent cleared, got %s", task.AgentID)
	}

	// Stats should show task back in backlog
	stats, err := coord.GetTaskStats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.Backlog != 1 {
		t.Errorf("expected 1 backlog task, got %d", stats.Backlog)
	}
	if stats.WIP != 0 {
		t.Errorf("expected 0 WIP tasks, got %d", stats.WIP)
	}

	// Another agent picks up the failed task and succeeds
	agent2 := "agent-retry-001"
	if err := coord.registry.Register(ctx, agent2, eventbus.RoleImplementation, 99998); err != nil {
		t.Fatalf("failed to register retry agent: %v", err)
	}

	if err := coord.ClaimTask(taskID, agent2); err != nil {
		t.Fatalf("failed to claim task on retry: %v", err)
	}

	// Complete successfully this time
	result := "Successfully refactored database layer into internal/db package"
	if err := coord.CompleteTask(taskID, result); err != nil {
		t.Fatalf("failed to complete task on retry: %v", err)
	}

	// Verify task is done
	task, err = coord.GetTask(taskID)
	if err != nil {
		t.Fatalf("failed to get task after retry: %v", err)
	}
	if task.Status != "done" {
		t.Errorf("expected status 'done', got %s", task.Status)
	}
	if task.Result != result {
		t.Errorf("unexpected result: %s", task.Result)
	}
	// Error should still be recorded for history
	if task.Error != errorMsg {
		t.Errorf("error history should be preserved: %s", task.Error)
	}

	t.Log("✅ Task failure and retry flow completed successfully!")
}

// TestConcurrentAgents tests multiple agents working simultaneously
func TestConcurrentAgents(t *testing.T) {
	tmpDir := t.TempDir()
	coord, err := NewBolt(tmpDir)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}
	defer coord.Close()

	ctx := context.Background()

	// Create 5 tasks
	taskIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		taskID, err := coord.CreateTask(
			"Task "+string(rune('A'+i)),
			"Description for task "+string(rune('A'+i)),
			"implementation",
		)
		if err != nil {
			t.Fatalf("failed to create task %d: %v", i, err)
		}
		taskIDs[i] = taskID
	}

	// Register 3 agents
	agents := []string{"agent-001", "agent-002", "agent-003"}
	for i, agent := range agents {
		if err := coord.registry.Register(ctx, agent, eventbus.RoleImplementation, 10000+i); err != nil {
			t.Fatalf("failed to register %s: %v", agent, err)
		}
	}

	// Each agent claims a task
	for i := 0; i < 3; i++ {
		if err := coord.ClaimTask(taskIDs[i], agents[i]); err != nil {
			t.Fatalf("failed to claim task %d: %v", i, err)
		}
	}

	// Verify 3 tasks are WIP, 2 still in backlog
	stats, err := coord.GetTaskStats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.WIP != 3 {
		t.Errorf("expected 3 WIP tasks, got %d", stats.WIP)
	}
	if stats.Backlog != 2 {
		t.Errorf("expected 2 backlog tasks, got %d", stats.Backlog)
	}

	// Agent 1 completes quickly
	if err := coord.CompleteTask(taskIDs[0], "Completed task A"); err != nil {
		t.Fatalf("failed to complete task 0: %v", err)
	}

	// Agent 1 picks up another task from backlog
	if err := coord.ClaimTask(taskIDs[3], agents[0]); err != nil {
		t.Fatalf("failed to claim task 3: %v", err)
	}

	// All remaining tasks are now claimed
	stats, err = coord.GetTaskStats()
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.WIP != 3 {
		t.Errorf("expected 3 WIP tasks, got %d", stats.WIP)
	}
	if stats.Backlog != 1 {
		t.Errorf("expected 1 backlog task, got %d", stats.Backlog)
	}
	if stats.Done != 1 {
		t.Errorf("expected 1 done task, got %d", stats.Done)
	}

	// Test that agent can't claim already claimed task
	err = coord.ClaimTask(taskIDs[1], agents[0]) // Task 1 is claimed by agent 2
	if err == nil {
		t.Error("expected error when claiming already-claimed task")
	}

	t.Log("✅ Concurrent agents test completed successfully!")
}

// TestEventBusIntegration tests that events are properly published
func TestEventBusIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	coord, err := NewBolt(tmpDir)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}
	defer coord.Close()

	ctx := context.Background()

	// Create a task (should publish EventTaskCreated)
	taskID, err := coord.CreateTask("Test task", "Test description", "testing")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Query events from EventStore
	time.Sleep(10 * time.Millisecond) // Give event time to be written

	events, err := coord.db.QueryEvents(ctx, storage.EventFilter{
		EventTypes: []string{"task_created"},
	})
	if err != nil {
		t.Fatalf("failed to query events: %v", err)
	}

	// Find our task event
	found := false
	for _, evt := range events {
		if evt.TaskID != nil && *evt.TaskID == taskID {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("task_created event not found")
	}

	// Register agent and claim task (should publish EventTaskClaimed)
	agent := "test-agent"
	if err := coord.registry.Register(ctx, agent, eventbus.RoleTesting, 55555); err != nil {
		t.Fatalf("failed to register agent: %v", err)
	}

	if err := coord.ClaimTask(taskID, agent); err != nil {
		t.Fatalf("failed to claim task: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	// Verify EventTaskClaimed was published
	claimEvents, err := coord.db.QueryEvents(ctx, storage.EventFilter{
		EventTypes: []string{"task_claimed"},
	})
	if err != nil {
		t.Fatalf("failed to query claim events: %v", err)
	}

	claimEventCount := 0
	for _, evt := range claimEvents {
		if evt.TaskID != nil && *evt.TaskID == taskID {
			claimEventCount++
		}
	}
	if claimEventCount != 1 {
		t.Errorf("expected 1 task_claimed event, got %d", claimEventCount)
	}

	t.Log("✅ EventBus integration test completed successfully!")
}
