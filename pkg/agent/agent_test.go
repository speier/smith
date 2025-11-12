package agent

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/speier/smith/pkg/agent/coordinator"
	"github.com/speier/smith/internal/eventbus"
	"github.com/speier/smith/pkg/agent/storage"
)

// TestCompleteVisionFlow demonstrates the full vision:
// User chats in REPL → Tasks created → Background agents work → Tasks completed
func TestCompleteVisionFlow(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()

	// Initialize coordinator (simulates what REPL does on startup)
	coord, err := coordinator.NewBolt(tmpDir)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}
	defer func() { _ = coord.Close() }()

	// Get registry from coordinator
	reg := coord.GetRegistry()

	// === STEP 1: User chats in REPL ===
	// User: "implement user authentication and add tests"
	t.Log("👤 USER: 'implement user authentication and add tests'")

	// REPL parses this and creates tasks
	taskID1, err := coord.CreateTask(
		"Implement user authentication",
		"Add login/logout endpoints with JWT tokens",
		string(eventbus.RoleImplementation), // "keymaker"
	)
	if err != nil {
		t.Fatalf("failed to create task 1: %v", err)
	}
	t.Logf("📋 Created task: %s", taskID1)

	taskID2, err := coord.CreateTask(
		"Add authentication tests",
		"Write unit and integration tests for auth",
		string(eventbus.RoleTesting), // "sentinel"
	)
	if err != nil {
		t.Fatalf("failed to create task 2: %v", err)
	}
	t.Logf("📋 Created task: %s", taskID2)

	// === STEP 2: Start background agents ===
	t.Log("🤖 Starting background agents...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start implementation agent
	implAgent := NewImplementationAgent(Config{
		AgentID:      "agent-impl-001",
		Coordinator:  coord,
		Registry:     reg,
		PollInterval: 50 * time.Millisecond,
	})

	implDone := make(chan error, 1)
	go func() {
		t.Log("🤖 Starting implementation agent goroutine")
		err := implAgent.Start(ctx)
		t.Logf("🤖 Implementation agent stopped with: %v", err)
		implDone <- err
	}()

	// Start testing agent
	testAgent := NewTestingAgent(Config{
		AgentID:      "agent-test-001",
		Coordinator:  coord,
		Registry:     reg,
		PollInterval: 50 * time.Millisecond,
	})

	testDone := make(chan error, 1)
	go func() {
		t.Log("🤖 Starting testing agent goroutine")
		err := testAgent.Start(ctx)
		t.Logf("🤖 Testing agent stopped with: %v", err)
		testDone <- err
	}()

	// Give agents time to register and start polling
	time.Sleep(100 * time.Millisecond)

	// === STEP 3: Wait for agents to process tasks ===
	t.Log("⏳ Agents processing tasks in background...")

	// Poll until all tasks are done (with timeout)
	timeout := time.After(3 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	allDone := false
	for !allDone {
		select {
		case <-timeout:
			t.Fatal("timeout waiting for tasks to complete")
		case <-ticker.C:
			stats, err := coord.GetTaskStats()
			if err != nil {
				t.Fatalf("failed to get stats: %v", err)
			}

			t.Logf("📊 Stats: backlog=%d, wip=%d, review=%d, done=%d",
				stats.Backlog, stats.WIP, stats.Review, stats.Done)

			if stats.Done == 2 {
				allDone = true
				t.Log("✅ All tasks completed!")
			}
		}
	}

	// === STEP 4: Verify results ===
	task1, err := coord.GetTask(taskID1)
	if err != nil {
		t.Fatalf("failed to get task 1: %v", err)
	}
	if task1.Status != "done" {
		t.Errorf("task 1 should be done, got: %s", task1.Status)
	}
	if task1.Result == "" {
		t.Error("task 1 should have a result")
	}
	t.Logf("✅ Task 1 result: %s", task1.Result)

	task2, err := coord.GetTask(taskID2)
	if err != nil {
		t.Fatalf("failed to get task 2: %v", err)
	}
	if task2.Status != "done" {
		t.Errorf("task 2 should be done, got: %s", task2.Status)
	}
	if task2.Result == "" {
		t.Error("task 2 should have a result")
	}
	t.Logf("✅ Task 2 result: %s", task2.Result)

	// === STEP 5: Verify events were published ===
	allEvents, err := coord.DB().QueryEvents(ctx, storage.EventFilter{})
	if err != nil {
		t.Fatalf("failed to query events: %v", err)
	}

	var events []string
	for _, evt := range allEvents {
		// Only include events for our tasks
		if evt.TaskID != nil && (*evt.TaskID == taskID1 || *evt.TaskID == taskID2) {
			events = append(events, evt.EventType)
		}
	}

	// Should have: task_created, task_claimed, task_completed for each task
	if len(events) < 6 {
		t.Logf("Events: %v", events)
		t.Errorf("expected at least 6 events, got %d", len(events))
	}

	// === STEP 6: Stop agents gracefully ===
	cancel() // Cancel context to stop agents

	// Wait for agents to stop
	select {
	case <-implDone:
		t.Log("🛑 Implementation agent stopped")
	case <-time.After(1 * time.Second):
		t.Error("implementation agent did not stop in time")
	}

	select {
	case <-testDone:
		t.Log("🛑 Testing agent stopped")
	case <-time.After(1 * time.Second):
		t.Error("testing agent did not stop in time")
	}

	t.Log("🎉 COMPLETE VISION FLOW SUCCESSFUL!")
}

// TestMultipleAgentsConcurrent tests multiple agents of the same type
func TestMultipleAgentsConcurrent(t *testing.T) {
	tmpDir := t.TempDir()
	coord, err := coordinator.NewBolt(tmpDir)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}
	defer func() { _ = coord.Close() }()

	reg := coord.GetRegistry()

	// Create 5 implementation tasks
	taskIDs := make([]string, 5)
	for i := 0; i < 5; i++ {
		taskID, err := coord.CreateTask(
			"Implementation task",
			"Some implementation work",
			string(eventbus.RoleImplementation),
		)
		if err != nil {
			t.Fatalf("failed to create task %d: %v", i, err)
		}
		taskIDs[i] = taskID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start 3 implementation agents (they'll race to claim tasks)
	agents := make([]Agent, 3)
	doneChans := make([]chan error, 3)

	for i := 0; i < 3; i++ {
		agents[i] = NewImplementationAgent(Config{
			AgentID:      "agent-impl-00" + string(rune('1'+i)),
			Coordinator:  coord,
			Registry:     reg,
			PollInterval: 30 * time.Millisecond,
		})

		doneChans[i] = make(chan error, 1)
		i := i // capture
		go func() {
			doneChans[i] <- agents[i].Start(ctx)
		}()
	}

	// Wait for all tasks to complete
	timeout := time.After(3 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("timeout waiting for tasks")
		case <-ticker.C:
			stats, err := coord.GetTaskStats()
			if err != nil {
				t.Fatalf("failed to get stats: %v", err)
			}

			if stats.Done == 5 {
				t.Logf("✅ All 5 tasks completed by 3 concurrent agents")
				cancel() // Stop agents
				return
			}
		}
	}
}

// TestAgentErrorHandling tests that agents properly handle task failures
func TestAgentErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	coord, err := coordinator.NewBolt(tmpDir)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}
	defer func() { _ = coord.Close() }()

	reg := coord.GetRegistry()

	// Create a task
	taskID, err := coord.CreateTask(
		"Failing task",
		"This will fail",
		string(eventbus.RoleImplementation),
	)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Create a failing agent
	cfg := Config{
		AgentID:      "agent-fail-001",
		Role:         eventbus.RoleImplementation, // Match the task role
		Coordinator:  coord,
		Registry:     reg,
		PollInterval: 50 * time.Millisecond,
	}
	baseAgent := NewBaseAgent(cfg)

	// Custom executor that always fails
	failExecutor := func(ctx context.Context, task *coordinator.Task) (string, error) {
		t.Logf("Executor called for task %s, will fail", task.ID)
		return "", os.ErrInvalid // Always fail
	}

	done := make(chan error, 1)
	go func() {
		done <- baseAgent.StartLoop(ctx, failExecutor)
	}()

	// Wait for task to be attempted and failed
	timeout := time.After(1 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	taskFailed := false
	for !taskFailed {
		select {
		case <-timeout:
			t.Fatal("timeout waiting for task to fail")
		case <-ticker.C:
			task, err := coord.GetTask(taskID)
			if err != nil {
				continue
			}
			if task.Error != "" {
				taskFailed = true
				t.Logf("✅ Agent properly handled failure: %s", task.Error)

				// Verify task is back in backlog
				if task.Status != "backlog" {
					t.Errorf("expected status 'backlog', got %s", task.Status)
				}
			}
		}
	}

	cancel()
}

// TestPlanningAgent tests the planning agent implementation
func TestPlanningAgent(t *testing.T) {
	tmpDir := t.TempDir()

	coord, err := coordinator.NewBolt(tmpDir)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}
	defer func() { _ = coord.Close() }()

	reg := coord.GetRegistry()

	// Create a planning task
	taskID, err := coord.CreateTask(
		"Plan authentication feature",
		"Break down auth feature into implementation tasks",
		string(eventbus.RolePlanning),
	)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Start planning agent
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	planAgent := NewPlanningAgent(Config{
		AgentID:      "agent-plan-001",
		Coordinator:  coord,
		Registry:     reg,
		PollInterval: 50 * time.Millisecond,
	})

	// Run agent in background
	go func() {
		if err := planAgent.Start(ctx); err != nil && err != context.Canceled {
			t.Logf("Planning agent error: %v", err)
		}
	}()

	// Wait for task completion
	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	taskCompleted := false
	for !taskCompleted {
		select {
		case <-timeout:
			t.Fatal("timeout waiting for planning task to complete")
		case <-ticker.C:
			stats, _ := coord.GetTaskStats()
			if stats.Done > 0 {
				taskCompleted = true
				task, err := coord.GetTask(taskID)
				if err != nil {
					t.Fatalf("failed to get task: %v", err)
				}
				if task.Result == "" {
					t.Error("expected non-empty result")
				}
				t.Logf("✅ Planning agent completed: %s", task.Result)
			}
		}
	}
}

// TestReviewAgent tests the review agent implementation
func TestReviewAgent(t *testing.T) {
	tmpDir := t.TempDir()

	coord, err := coordinator.NewBolt(tmpDir)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}
	defer func() { _ = coord.Close() }()

	reg := coord.GetRegistry()

	// Create a review task
	taskID, err := coord.CreateTask(
		"Review authentication code",
		"Review JWT implementation for security issues",
		string(eventbus.RoleReview),
	)
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Start review agent
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	reviewAgent := NewReviewAgent(Config{
		AgentID:      "agent-review-001",
		Coordinator:  coord,
		Registry:     reg,
		PollInterval: 50 * time.Millisecond,
	})

	// Run agent in background
	go func() {
		if err := reviewAgent.Start(ctx); err != nil && err != context.Canceled {
			t.Logf("Review agent error: %v", err)
		}
	}()

	// Wait for task completion
	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	taskCompleted := false
	for !taskCompleted {
		select {
		case <-timeout:
			t.Fatal("timeout waiting for review task to complete")
		case <-ticker.C:
			stats, _ := coord.GetTaskStats()
			if stats.Done > 0 {
				taskCompleted = true
				task, err := coord.GetTask(taskID)
				if err != nil {
					t.Fatalf("failed to get task: %v", err)
				}
				if task.Result == "" {
					t.Error("expected non-empty result")
				}
				t.Logf("✅ Review agent completed: %s", task.Result)
			}
		}
	}
}

// TestAllAgentTypes tests all five agent types working together
func TestAllAgentTypes(t *testing.T) {
	tmpDir := t.TempDir()

	coord, err := coordinator.NewBolt(tmpDir)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}
	defer func() { _ = coord.Close() }()

	reg := coord.GetRegistry()

	// Create tasks for each agent type
	tasks := []struct {
		title       string
		description string
		role        string
	}{
		{"Plan feature", "Break down feature", string(eventbus.RolePlanning)},
		{"Implement feature", "Write code", string(eventbus.RoleImplementation)},
		{"Test feature", "Write tests", string(eventbus.RoleTesting)},
		{"Review code", "Review implementation", string(eventbus.RoleReview)},
	}

	for _, task := range tasks {
		_, err := coord.CreateTask(task.title, task.description, task.role)
		if err != nil {
			t.Fatalf("failed to create %s task: %v", task.role, err)
		}
	}

	// Start all agent types
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	agents := []Agent{
		NewPlanningAgent(Config{
			AgentID:      "agent-plan-001",
			Coordinator:  coord,
			Registry:     reg,
			PollInterval: 30 * time.Millisecond,
		}),
		NewImplementationAgent(Config{
			AgentID:      "agent-impl-001",
			Coordinator:  coord,
			Registry:     reg,
			PollInterval: 30 * time.Millisecond,
		}),
		NewTestingAgent(Config{
			AgentID:      "agent-test-001",
			Coordinator:  coord,
			Registry:     reg,
			PollInterval: 30 * time.Millisecond,
		}),
		NewReviewAgent(Config{
			AgentID:      "agent-review-001",
			Coordinator:  coord,
			Registry:     reg,
			PollInterval: 30 * time.Millisecond,
		}),
	}

	// Start all agents
	for _, agent := range agents {
		go func(a Agent) {
			if err := a.Start(ctx); err != nil && err != context.Canceled {
				t.Logf("Agent error: %v", err)
			}
		}(agent)
	}

	// Wait for all tasks to complete
	timeout := time.After(4 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	allCompleted := false
	for !allCompleted {
		select {
		case <-timeout:
			stats, _ := coord.GetTaskStats()
			t.Fatalf("timeout: backlog=%d, wip=%d, review=%d, done=%d",
				stats.Backlog, stats.WIP, stats.Review, stats.Done)
		case <-ticker.C:
			stats, _ := coord.GetTaskStats()
			if stats.Done == 4 {
				allCompleted = true
				t.Logf("✅ All 4 agent types completed their tasks!")
			}
		}
	}
}
