package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// RunCommandTool executes a shell command
type RunCommandTool struct {
	workDir string
	timeout time.Duration
}

// NewRunCommandTool creates a new RunCommandTool
func NewRunCommandTool(workDir string, timeout time.Duration) *RunCommandTool {
	if timeout == 0 {
		timeout = 30 * time.Second // Default timeout
	}
	return &RunCommandTool{
		workDir: workDir,
		timeout: timeout,
	}
}

func (t *RunCommandTool) Name() string {
	return "run_command"
}

func (t *RunCommandTool) Description() string {
	return "Execute a shell command in the working directory"
}

func (t *RunCommandTool) Validate(params map[string]interface{}) error {
	command, ok := params["command"].(string)
	if !ok || command == "" {
		return fmt.Errorf("command parameter is required and must be a string")
	}

	// Basic validation: no obviously dangerous commands
	dangerous := []string{"rm -rf /", "dd if=", ":(){ :|:& };:", "mkfs"}
	cmdLower := strings.ToLower(command)
	for _, pattern := range dangerous {
		if strings.Contains(cmdLower, pattern) {
			return fmt.Errorf("potentially dangerous command detected")
		}
	}

	return nil
}

func (t *RunCommandTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	command := params["command"].(string)

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	// Execute command
	cmd := exec.CommandContext(execCtx, "sh", "-c", command)
	cmd.Dir = t.workDir

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	if err != nil {
		return &ToolResult{
			Success: false,
			Output:  outputStr,
			Error:   fmt.Sprintf("command failed: %v", err),
			Data: map[string]interface{}{
				"command":  command,
				"exitCode": cmd.ProcessState.ExitCode(),
			},
		}, ErrCommandFailed
	}

	return &ToolResult{
		Success: true,
		Output:  outputStr,
		Data: map[string]interface{}{
			"command":  command,
			"exitCode": 0,
		},
	}, nil
}

func (t *RunCommandTool) RequiresConfirmation(level SafetyLevel) bool {
	return level >= SafetyMedium // Command execution requires confirmation at medium and high
}
