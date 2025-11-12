package tools

import (
	"context"
	"testing"
	"time"
)

func TestRunCommandTool(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		params      map[string]interface{}
		shouldError bool
		expectMsg   string
	}{
		{
			name:      "simple_echo",
			params:    map[string]interface{}{"command": "echo hello"},
			expectMsg: "hello",
		},
		{
			name:   "pwd_command",
			params: map[string]interface{}{"command": "pwd"},
		},
		{
			name:   "multi_command",
			params: map[string]interface{}{"command": "echo foo && echo bar"},
		},
		{
			name:        "failing_command",
			params:      map[string]interface{}{"command": "exit 1"},
			shouldError: true,
		},
		{
			name:        "missing_command",
			params:      map[string]interface{}{},
			shouldError: true,
		},
		{
			name:        "empty_command",
			params:      map[string]interface{}{"command": ""},
			shouldError: true,
		},
		{
			name:        "dangerous_rm_rf",
			params:      map[string]interface{}{"command": "rm -rf /"},
			shouldError: true,
		},
		{
			name:        "dangerous_dd",
			params:      map[string]interface{}{"command": "dd if=/dev/zero of=/dev/sda"},
			shouldError: true,
		},
		{
			name:        "fork_bomb",
			params:      map[string]interface{}{"command": ":(){ :|:& };:"},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewRunCommandTool(tempDir, 5*time.Second)

			// Validate
			err := tool.Validate(tt.params)
			if tt.shouldError && err == nil && (tt.name == "dangerous_rm_rf" || tt.name == "dangerous_dd" || tt.name == "fork_bomb" || tt.name == "missing_command" || tt.name == "empty_command") {
				t.Error("expected validation error, got nil")
				return
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
				return
			}

			if tt.shouldError && (tt.name == "dangerous_rm_rf" || tt.name == "dangerous_dd" || tt.name == "fork_bomb" || tt.name == "missing_command" || tt.name == "empty_command") {
				return // Skip execution for validation failures
			}

			// Execute
			result, err := tool.Execute(context.Background(), tt.params)

			if tt.shouldError {
				if err == nil {
					t.Error("expected execution error, got nil")
				}
				if result == nil || result.Success {
					t.Error("expected unsuccessful result")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil || !result.Success {
					t.Errorf("expected successful result, got: %+v", result)
				}
				if tt.expectMsg != "" && !contains(result.Output, tt.expectMsg) {
					t.Errorf("expected output containing '%s', got '%s'", tt.expectMsg, result.Output)
				}
			}
		})
	}
}

func TestRunCommandTool_Timeout(t *testing.T) {
	tempDir := t.TempDir()

	// Create tool with very short timeout
	tool := NewRunCommandTool(tempDir, 100*time.Millisecond)

	params := map[string]interface{}{
		"command": "sleep 5", // Sleep longer than timeout
	}

	// Validate should pass
	if err := tool.Validate(params); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	// Execute should timeout
	result, err := tool.Execute(context.Background(), params)

	if err == nil {
		t.Error("expected timeout error, got nil")
	}

	if result == nil || result.Success {
		t.Error("expected unsuccessful result due to timeout")
	}
}

func TestRunCommandTool_WorkingDirectory(t *testing.T) {
	tempDir := t.TempDir()

	tool := NewRunCommandTool(tempDir, 5*time.Second)

	params := map[string]interface{}{
		"command": "pwd",
	}

	result, err := tool.Execute(context.Background(), params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("expected successful result")
	}

	// Output should contain temp directory path
	if !contains(result.Output, tempDir) {
		t.Errorf("expected output to contain %s, got %s", tempDir, result.Output)
	}
}

func TestRunCommandTool_Metadata(t *testing.T) {
	tool := NewRunCommandTool("/tmp", 5*time.Second)

	if tool.Name() != "run_command" {
		t.Errorf("expected name 'run_command', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}

	if !tool.RequiresConfirmation(SafetyMedium) {
		t.Error("run_command should require confirmation at medium safety")
	}

	if !tool.RequiresConfirmation(SafetyHigh) {
		t.Error("run_command should require confirmation at high safety")
	}

	if tool.RequiresConfirmation(SafetyLow) {
		t.Error("run_command should not require confirmation at low safety")
	}
}

func TestRunCommandTool_DefaultTimeout(t *testing.T) {
	// Create tool without explicit timeout
	tool := NewRunCommandTool("/tmp", 0)

	if tool.timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", tool.timeout)
	}
}

func TestRunCommandTool_ExitCode(t *testing.T) {
	tempDir := t.TempDir()
	tool := NewRunCommandTool(tempDir, 5*time.Second)

	tests := []struct {
		name         string
		command      string
		expectExitOk bool
	}{
		{"success", "exit 0", true},
		{"failure", "exit 1", false},
		{"failure_42", "exit 42", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := map[string]interface{}{"command": tt.command}

			result, err := tool.Execute(context.Background(), params)

			if tt.expectExitOk {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !result.Success {
					t.Error("expected successful result")
				}
			} else {
				if err == nil {
					t.Error("expected error for non-zero exit code")
				}
				if result.Success {
					t.Error("expected unsuccessful result")
				}
			}

			// Check exit code in data
			if data, ok := result.Data.(map[string]interface{}); ok {
				if exitCode, ok := data["exitCode"].(int); ok {
					if tt.expectExitOk && exitCode != 0 {
						t.Errorf("expected exit code 0, got %d", exitCode)
					}
					if !tt.expectExitOk && exitCode == 0 {
						t.Error("expected non-zero exit code")
					}
				}
			}
		})
	}
}
