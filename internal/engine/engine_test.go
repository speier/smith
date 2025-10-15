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
