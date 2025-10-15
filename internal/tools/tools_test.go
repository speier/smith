package tools

import (
	"context"
	"errors"
	"testing"
)

func TestNewExecutor(t *testing.T) {
	tests := []struct {
		name        string
		workDir     string
		safetyLevel SafetyLevel
	}{
		{"default", "/tmp/test", SafetyMedium},
		{"high_safety", "/home/user", SafetyHigh},
		{"low_safety", "/var/app", SafetyLow},
		{"safety_off", "/opt/data", SafetyOff},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewExecutor(tt.workDir, tt.safetyLevel)

			if executor == nil {
				t.Fatal("expected executor, got nil")
			}

			if executor.workDir != tt.workDir {
				t.Errorf("expected workDir %s, got %s", tt.workDir, executor.workDir)
			}

			if executor.safetyLevel != tt.safetyLevel {
				t.Errorf("expected safetyLevel %d, got %d", tt.safetyLevel, executor.safetyLevel)
			}

			if executor.tools == nil {
				t.Error("expected tools map to be initialized")
			}

			if len(executor.tools) != 0 {
				t.Errorf("expected empty tools map, got %d tools", len(executor.tools))
			}
		})
	}
}

func TestExecutor_Register(t *testing.T) {
	executor := NewExecutor("/tmp", SafetyMedium)
	tool := &mockTool{name: "test_tool"}

	executor.Register(tool)

	if len(executor.tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(executor.tools))
	}

	registered, ok := executor.GetTool("test_tool")
	if !ok {
		t.Fatal("expected to find registered tool")
	}

	if registered.Name() != "test_tool" {
		t.Errorf("expected tool name 'test_tool', got '%s'", registered.Name())
	}
}

func TestExecutor_Execute(t *testing.T) {
	tests := []struct {
		name        string
		toolName    string
		params      map[string]interface{}
		shouldError bool
		errorMsg    string
	}{
		{
			name:     "success",
			toolName: "test_tool",
			params:   map[string]interface{}{"valid": true},
		},
		{
			name:        "tool_not_found",
			toolName:    "nonexistent",
			params:      map[string]interface{}{},
			shouldError: true,
			errorMsg:    "tool not found",
		},
		{
			name:        "validation_failure",
			toolName:    "test_tool",
			params:      map[string]interface{}{"valid": false},
			shouldError: true,
			errorMsg:    "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewExecutor("/tmp", SafetyMedium)
			executor.Register(&mockTool{name: "test_tool"})

			result, err := executor.Execute(context.Background(), tt.toolName, tt.params)

			if tt.shouldError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if result == nil || result.Success {
					t.Error("expected unsuccessful result")
				}
				if tt.errorMsg != "" && !contains(result.Error, tt.errorMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorMsg, result.Error)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil || !result.Success {
					t.Error("expected successful result")
				}
			}
		})
	}
}

func TestExecutor_ListTools(t *testing.T) {
	executor := NewExecutor("/tmp", SafetyMedium)

	// Start with empty
	tools := executor.ListTools()
	if len(tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(tools))
	}

	// Register some tools
	executor.Register(&mockTool{name: "tool1"})
	executor.Register(&mockTool{name: "tool2"})
	executor.Register(&mockTool{name: "tool3"})

	tools = executor.ListTools()
	if len(tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(tools))
	}

	// Verify all tools are listed
	toolMap := make(map[string]bool)
	for _, name := range tools {
		toolMap[name] = true
	}

	for _, expected := range []string{"tool1", "tool2", "tool3"} {
		if !toolMap[expected] {
			t.Errorf("expected tool '%s' to be listed", expected)
		}
	}
}

func TestExecutor_SafetyLevel(t *testing.T) {
	executor := NewExecutor("/tmp", SafetyMedium)

	if executor.GetSafetyLevel() != SafetyMedium {
		t.Errorf("expected SafetyMedium, got %d", executor.GetSafetyLevel())
	}

	executor.SetSafetyLevel(SafetyHigh)

	if executor.GetSafetyLevel() != SafetyHigh {
		t.Errorf("expected SafetyHigh, got %d", executor.GetSafetyLevel())
	}

	executor.SetSafetyLevel(SafetyOff)

	if executor.GetSafetyLevel() != SafetyOff {
		t.Errorf("expected SafetyOff, got %d", executor.GetSafetyLevel())
	}
}

func TestExecutor_WorkDir(t *testing.T) {
	workDir := "/tmp/test/path"
	executor := NewExecutor(workDir, SafetyMedium)

	if executor.WorkDir() != workDir {
		t.Errorf("expected workDir %s, got %s", workDir, executor.WorkDir())
	}
}

// Mock tool for testing

type mockTool struct {
	name string
}

func (m *mockTool) Name() string {
	return m.name
}

func (m *mockTool) Description() string {
	return "Mock tool for testing"
}

func (m *mockTool) Validate(params map[string]interface{}) error {
	if valid, ok := params["valid"].(bool); ok && !valid {
		return errors.New("validation failed")
	}
	return nil
}

func (m *mockTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	return &ToolResult{
		Success: true,
		Output:  "mock execution successful",
	}, nil
}

func (m *mockTool) RequiresConfirmation(level SafetyLevel) bool {
	return level >= SafetyHigh
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
