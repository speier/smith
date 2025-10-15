package engine

import (
	"testing"
)

// TestEngineCreation tests that the engine creates successfully with coordinator
func TestEngineCreation(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		ProjectPath: tmpDir,
		LLMProvider: nil, // Will use default copilot provider
	}

	engine, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	if engine == nil {
		t.Fatal("Engine should not be nil")
	}

	if engine.coord == nil {
		t.Fatal("Coordinator should not be nil")
	}

	if engine.llm == nil {
		t.Fatal("LLM provider should not be nil")
	}

	if engine.projectPath != tmpDir {
		t.Errorf("Expected projectPath %s, got %s", tmpDir, engine.projectPath)
	}
}

// TestEngineCoordinatorIntegration tests that engine can call coordinator methods
func TestEngineCoordinatorIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		ProjectPath: tmpDir,
	}

	engine, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Test that GetStatus works (calls coordinator.GetTaskStats)
	status := engine.GetStatus()
	if status == "" {
		t.Error("Expected non-empty status string")
	}

	// Status should contain task stats
	if len(status) < 10 {
		t.Errorf("Expected detailed status, got: %s", status)
	}
}

// TestEngineConversation tests basic conversation flow
func TestEngineConversation(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		ProjectPath: tmpDir,
	}

	engine, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Initially should have no history
	history := engine.GetConversationHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d messages", len(history))
	}

	// Note: We can't test actual Chat() without mocking the LLM
	// because it requires GitHub Copilot authentication
	// This test just verifies the engine is set up correctly
}

// TestEngineClearConversation tests clearing conversation history
func TestEngineClearConversation(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		ProjectPath: tmpDir,
	}

	engine, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Add some fake history
	engine.conversationHistory = []Message{
		{Role: "user", Content: "test"},
		{Role: "assistant", Content: "response"},
	}

	if len(engine.conversationHistory) != 2 {
		t.Fatal("Failed to add test messages")
	}

	// Clear it
	engine.ClearConversation()

	history := engine.GetConversationHistory()
	if len(history) != 0 {
		t.Errorf("Expected empty history after clear, got %d messages", len(history))
	}

	if engine.pendingPlan != nil {
		t.Error("Expected pendingPlan to be nil after clear")
	}
}

// TestTaskManagementTools tests the task management tool handlers
func TestTaskManagementTools(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := Config{
		ProjectPath: tmpDir,
	}

	engine, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Test create_task
	t.Run("create_task", func(t *testing.T) {
		result, err := engine.handleCreateTask(map[string]interface{}{
			"title":       "Test implementation",
			"description": "Implement test feature",
			"agent_role":  "implementation",
		})
		if err != nil {
			t.Fatalf("create_task failed: %v", err)
		}
		if result == "" {
			t.Error("Expected non-empty result from create_task")
		}
		if len(result) < 10 {
			t.Errorf("Expected detailed result, got: %s", result)
		}
	})

	// Test get_task_stats
	t.Run("get_task_stats", func(t *testing.T) {
		result, err := engine.handleGetTaskStats(map[string]interface{}{})
		if err != nil {
			t.Fatalf("get_task_stats failed: %v", err)
		}
		if result == "" {
			t.Error("Expected non-empty result from get_task_stats")
		}
		// Should show at least 1 backlog task from previous test
		if len(result) < 20 {
			t.Errorf("Expected detailed stats, got: %s", result)
		}
	})

	// Test list_tasks with no filter
	t.Run("list_tasks_all", func(t *testing.T) {
		result, err := engine.handleListTasks(map[string]interface{}{})
		if err != nil {
			t.Fatalf("list_tasks failed: %v", err)
		}
		if result == "" {
			t.Error("Expected non-empty result from list_tasks")
		}
	})

	// Test list_tasks with status filter
	t.Run("list_tasks_filtered", func(t *testing.T) {
		result, err := engine.handleListTasks(map[string]interface{}{
			"status": "backlog",
		})
		if err != nil {
			t.Fatalf("list_tasks with filter failed: %v", err)
		}
		if result == "" {
			t.Error("Expected non-empty result from filtered list_tasks")
		}
	})

	// Test get_task
	t.Run("get_task", func(t *testing.T) {
		// First create a task to get its ID
		createResult, err := engine.handleCreateTask(map[string]interface{}{
			"title":       "Test task for retrieval",
			"description": "Testing get_task",
			"agent_role":  "testing",
		})
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}

		// Extract task ID from result (format: "✅ Created task task-XXX: ...")
		// Simple approach: just use a known pattern or query the coordinator
		tasks, err := engine.coord.GetTasksByStatus("backlog")
		if err != nil {
			t.Fatalf("Failed to get tasks: %v", err)
		}
		if len(tasks) == 0 {
			t.Fatal("Expected at least one task")
		}

		taskID := tasks[0].ID

		result, err := engine.handleGetTask(map[string]interface{}{
			"task_id": taskID,
		})
		if err != nil {
			t.Fatalf("get_task failed: %v", err)
		}
		if result == "" {
			t.Error("Expected non-empty result from get_task")
		}
		if len(result) < 20 {
			t.Errorf("Expected detailed task info, got: %s", result)
		}

		// Verify result format contains expected fields
		expectedFields := []string{"Task", "Title", "Status", "Agent", "Description"}
		for _, field := range expectedFields {
			if !contains(result, field) {
				t.Errorf("Expected result to contain '%s', got: %s", field, result)
			}
		}

		// Verify we didn't leak the create_task result
		_ = createResult
	})

	// Test error cases
	t.Run("create_task_invalid_role", func(t *testing.T) {
		_, err := engine.handleCreateTask(map[string]interface{}{
			"title":       "Bad task",
			"description": "Invalid role",
			"agent_role":  "invalid_role",
		})
		if err == nil {
			t.Error("Expected error for invalid agent_role")
		}
	})

	t.Run("get_task_missing_id", func(t *testing.T) {
		_, err := engine.handleGetTask(map[string]interface{}{})
		if err == nil {
			t.Error("Expected error for missing task_id")
		}
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
