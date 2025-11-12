package tools

import (
	"context"
	"errors"
)

var (
	// ErrPermissionDenied is returned when a tool operation is not allowed
	ErrPermissionDenied = errors.New("permission denied")

	// ErrInvalidPath is returned when a file path is invalid or unsafe
	ErrInvalidPath = errors.New("invalid or unsafe path")

	// ErrCommandFailed is returned when a command execution fails
	ErrCommandFailed = errors.New("command execution failed")
)

// SafetyLevel defines the level of safety checks for tool execution
type SafetyLevel int

const (
	SafetyOff    SafetyLevel = 0 // No safety checks (dangerous)
	SafetyLow    SafetyLevel = 1 // Basic validation only
	SafetyMedium SafetyLevel = 2 // Require confirmation for writes
	SafetyHigh   SafetyLevel = 3 // Strict validation, read-only by default
)

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Success bool        `json:"success"`
	Output  string      `json:"output,omitempty"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Tool is the interface that all tools must implement
type Tool interface {
	// Name returns the name of the tool
	Name() string

	// Description returns a human-readable description of what the tool does
	Description() string

	// Execute runs the tool with the given parameters
	Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error)

	// Validate checks if the parameters are valid before execution
	Validate(params map[string]interface{}) error

	// RequiresConfirmation returns true if this tool needs user confirmation at the current safety level
	RequiresConfirmation(level SafetyLevel) bool
}

// Executor manages and executes tools with safety checks
type Executor struct {
	tools       map[string]Tool
	safetyLevel SafetyLevel
	workDir     string // Base working directory for file operations
}

// NewExecutor creates a new tool executor
func NewExecutor(workDir string, safetyLevel SafetyLevel) *Executor {
	return &Executor{
		tools:       make(map[string]Tool),
		safetyLevel: safetyLevel,
		workDir:     workDir,
	}
}

// Register registers a tool with the executor
func (e *Executor) Register(tool Tool) {
	e.tools[tool.Name()] = tool
}

// Execute executes a tool by name with the given parameters
func (e *Executor) Execute(ctx context.Context, toolName string, params map[string]interface{}) (*ToolResult, error) {
	tool, ok := e.tools[toolName]
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "tool not found: " + toolName,
		}, errors.New("tool not found")
	}

	// Validate parameters
	if err := tool.Validate(params); err != nil {
		return &ToolResult{
			Success: false,
			Error:   "validation failed: " + err.Error(),
		}, err
	}

	// Execute the tool
	return tool.Execute(ctx, params)
}

// GetTool returns a registered tool by name
func (e *Executor) GetTool(name string) (Tool, bool) {
	tool, ok := e.tools[name]
	return tool, ok
}

// ListTools returns the names of all registered tools
func (e *Executor) ListTools() []string {
	names := make([]string, 0, len(e.tools))
	for name := range e.tools {
		names = append(names, name)
	}
	return names
}

// SetSafetyLevel updates the safety level
func (e *Executor) SetSafetyLevel(level SafetyLevel) {
	e.safetyLevel = level
}

// GetSafetyLevel returns the current safety level
func (e *Executor) GetSafetyLevel() SafetyLevel {
	return e.safetyLevel
}

// WorkDir returns the base working directory
func (e *Executor) WorkDir() string {
	return e.workDir
}
