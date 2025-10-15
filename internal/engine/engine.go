package engine

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/speier/smith/internal/coordinator"
	"github.com/speier/smith/internal/llm"
	"github.com/speier/smith/internal/orchestrator"
)

// Engine is the core Smith system
// It's frontend-agnostic - can be used by CLI, web UI, API, etc.
type Engine struct {
	coord       *coordinator.FileCoordinator
	orc         *orchestrator.Orchestrator
	llm         llm.Provider
	projectPath string

	// Conversation state
	conversationHistory []Message
	pendingPlan         *Plan
}

type Message struct {
	Role    string // "user", "assistant", "system"
	Content string
}

type Plan struct {
	Title       string
	Description string
	Tasks       []Task
	Confirmed   bool
}

type Task struct {
	ID          string
	Title       string
	Description string
	Tags        []string
}

type Config struct {
	ProjectPath string
	LLMProvider llm.Provider
}

// New creates a new Smith engine instance
func New(cfg Config) (*Engine, error) {
	// Use default copilot provider if not specified
	if cfg.LLMProvider == nil {
		cfg.LLMProvider = llm.NewCopilotProvider()
	}

	coord := coordinator.New(cfg.ProjectPath)

	return &Engine{
		llm:         cfg.LLMProvider,
		coord:       coord,
		projectPath: cfg.ProjectPath,
	}, nil
}

// getSystemPrompt returns the system prompt with tool usage instructions
func (e *Engine) getSystemPrompt() string {
	return `You are Smith, an AI-powered development assistant.

You can help developers by:
- Creating and editing files (write_file, read_file, edit_file)
- Running commands (run_command)
- Exploring the project structure (list_files)

**File Operations Best Practices:**
1. Always read_file before editing to understand current content
2. For multi-file projects, create all necessary files
3. After making changes, use run_command to test (build, run tests, etc.)

**Workflow Example:**
User: "Create a hello world Go program"
1. Use write_file to create main.go with the code
2. Use run_command to build: "go build ."
3. Use run_command to run: "./main" or "go run main.go"

Be conversational and helpful. Confirm what you're doing before executing commands.
Keep responses concise but friendly.`
}

// getTools returns the available tools for the LLM
func (e *Engine) getTools() []llm.Tool {
	return []llm.Tool{
		{
			Name:        "write_file",
			Description: "Create or overwrite a file with the specified content",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file relative to project root",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Complete content to write to the file",
					},
				},
				"required": []string{"file_path", "content"},
			},
		},
		{
			Name:        "read_file",
			Description: "Read the contents of a file",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file relative to project root",
					},
				},
				"required": []string{"file_path"},
			},
		},
		{
			Name:        "edit_file",
			Description: "Edit a file by replacing old_content with new_content",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file relative to project root",
					},
					"old_content": map[string]interface{}{
						"type":        "string",
						"description": "Exact text to replace (must match exactly)",
					},
					"new_content": map[string]interface{}{
						"type":        "string",
						"description": "New text to replace old_content with",
					},
				},
				"required": []string{"file_path", "old_content", "new_content"},
			},
		},
		{
			Name:        "list_files",
			Description: "List files and directories in a path",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"directory": map[string]interface{}{
						"type":        "string",
						"description": "Directory path relative to project root (default: .)",
					},
				},
			},
		},
		{
			Name:        "run_command",
			Description: "Execute a shell command in the project directory",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "Shell command to execute",
					},
					"working_dir": map[string]interface{}{
						"type":        "string",
						"description": "Working directory relative to project root (default: .)",
					},
				},
				"required": []string{"command"},
			},
		},
	}
}

// Chat sends a message and gets a response
// This is the main interface for any frontend
func (e *Engine) Chat(userMessage string) (string, error) {
	// Add user message to history
	e.conversationHistory = append(e.conversationHistory, Message{
		Role:    "user",
		Content: userMessage,
	})

	// Process with LLM (or mock for now)
	response := e.processMessage(userMessage)

	// Add assistant response to history
	e.conversationHistory = append(e.conversationHistory, Message{
		Role:    "assistant",
		Content: response,
	})

	return response, nil
}

// ChatStream sends a message and streams the response
func (e *Engine) ChatStream(userMessage string, callback func(string) error) error {
	// Add user message to history
	e.conversationHistory = append(e.conversationHistory, Message{
		Role:    "user",
		Content: userMessage,
	})

	// Convert conversation history to LLM messages with system prompt
	messages := []llm.Message{
		{
			Role:    "system",
			Content: e.getSystemPrompt(),
		},
	}
	for _, msg := range e.conversationHistory {
		messages = append(messages, llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Get tools
	tools := e.getTools()

	// Stream the response
	var fullResponse strings.Builder
	err := e.llm.ChatStream(messages, tools, func(response *llm.Response) error {
		// Handle tool calls
		if len(response.ToolCalls) > 0 {
			for _, toolCall := range response.ToolCalls {
				result, err := e.executeToolCall(toolCall)
				if err != nil {
					return fmt.Errorf("tool execution failed: %w", err)
				}
				// Send tool result back to callback
				toolMsg := fmt.Sprintf("\n[%s: %s]\n", toolCall.Name, result)
				if err := callback(toolMsg); err != nil {
					return err
				}
				fullResponse.WriteString(result)
			}
		}

		// Stream regular content
		if response.Content != "" {
			fullResponse.WriteString(response.Content)
			return callback(response.Content)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Add complete assistant response to history
	e.conversationHistory = append(e.conversationHistory, Message{
		Role:    "assistant",
		Content: fullResponse.String(),
	})

	return nil
}

// GetConversationHistory returns the full conversation
func (e *Engine) GetConversationHistory() []Message {
	return e.conversationHistory
}

// ClearConversation clears the conversation history
func (e *Engine) ClearConversation() {
	e.conversationHistory = []Message{}
	e.pendingPlan = nil
}

// GetStatus returns current task statistics
func (e *Engine) GetStatus() string {
	stats, err := e.coord.GetTaskStats()
	if err != nil {
		return fmt.Sprintf("Error getting status: %v", err)
	}

	return fmt.Sprintf(`
📊 Task Status:
   Todo:   %d tasks ready
   WIP:    %d in progress
   Review: %d awaiting review
   Done:   %d completed
`, stats.Available, stats.InProgress, stats.Blocked, stats.Done)
}

// handleWriteFile handles the write_file tool call
func (e *Engine) handleWriteFile(input map[string]interface{}) (string, error) {
	filePath, ok := input["file_path"].(string)
	if !ok {
		return "", fmt.Errorf("file_path is required")
	}

	content, ok := input["content"].(string)
	if !ok {
		return "", fmt.Errorf("content is required")
	}

	// Build full path (relative to project root)
	fullPath := filepath.Join(e.projectPath, filePath)

	// TODO: Add safety checks based on auto-level
	// For MVP, just create the file

	// Ensure parent directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("✅ Created: %s", filePath), nil
}

// handleReadFile handles the read_file tool call
func (e *Engine) handleReadFile(input map[string]interface{}) (string, error) {
	filePath, ok := input["file_path"].(string)
	if !ok {
		return "", fmt.Errorf("file_path is required")
	}

	// Build full path (relative to project root)
	fullPath := filepath.Join(e.projectPath, filePath)

	// Read file
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// handleEditFile handles the edit_file tool call
func (e *Engine) handleEditFile(input map[string]interface{}) (string, error) {
	filePath, ok := input["file_path"].(string)
	if !ok {
		return "", fmt.Errorf("file_path is required")
	}

	oldContent, ok := input["old_content"].(string)
	if !ok {
		return "", fmt.Errorf("old_content is required")
	}

	newContent, ok := input["new_content"].(string)
	if !ok {
		return "", fmt.Errorf("new_content is required")
	}

	// Build full path (relative to project root)
	fullPath := filepath.Join(e.projectPath, filePath)

	// Read current file content
	currentContent, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Replace old content with new content
	updatedContent := strings.Replace(string(currentContent), oldContent, newContent, 1)

	// Check if replacement was made
	if updatedContent == string(currentContent) {
		return "", fmt.Errorf("old_content not found in file")
	}

	// Write updated content
	if err := os.WriteFile(fullPath, []byte(updatedContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return fmt.Sprintf("✅ Edited: %s", filePath), nil
}

// handleListFiles handles the list_files tool call
func (e *Engine) handleListFiles(input map[string]interface{}) (string, error) {
	dirPath, _ := input["directory"].(string)
	if dirPath == "" {
		dirPath = "."
	}

	// Build full path (relative to project root)
	fullPath := filepath.Join(e.projectPath, dirPath)

	// Read directory
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read directory: %w", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("📁 %s:\n", dirPath))
	for _, entry := range entries {
		if entry.IsDir() {
			result.WriteString(fmt.Sprintf("  📂 %s/\n", entry.Name()))
		} else {
			result.WriteString(fmt.Sprintf("  📄 %s\n", entry.Name()))
		}
	}

	return result.String(), nil
}

// handleRunCommand handles the run_command tool call
func (e *Engine) handleRunCommand(input map[string]interface{}) (string, error) {
	command, ok := input["command"].(string)
	if !ok {
		return "", fmt.Errorf("command is required")
	}

	workingDir, _ := input["working_dir"].(string)
	if workingDir == "" {
		workingDir = e.projectPath
	} else {
		workingDir = filepath.Join(e.projectPath, workingDir)
	}

	// TODO: Add safety checks - require confirmation for certain commands
	// For MVP, execute directly

	// Use sh -c to execute command (works on Unix-like systems)
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = workingDir

	output, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			return "", fmt.Errorf("failed to execute command: %w", err)
		}
	}

	return fmt.Sprintf("Exit code: %d\n%s", exitCode, string(output)), nil
}

type Status struct {
	TodoCount       int
	InProgressCount int
	ReviewCount     int
	DoneCount       int
	BlockedCount    int
}

// executeToolCall executes a tool call and returns the result
func (e *Engine) executeToolCall(toolCall llm.ToolCall) (string, error) {
	switch toolCall.Name {
	case "write_file":
		return e.handleWriteFile(toolCall.Input)
	case "read_file":
		return e.handleReadFile(toolCall.Input)
	case "edit_file":
		return e.handleEditFile(toolCall.Input)
	case "list_files":
		return e.handleListFiles(toolCall.Input)
	case "run_command":
		return e.handleRunCommand(toolCall.Input)
	default:
		return "", fmt.Errorf("unknown tool: %s", toolCall.Name)
	}
}

// GetBacklog returns all tasks across all states
func (e *Engine) GetBacklog() ([]TaskInfo, error) {
	// TODO: Implement - read from .smith/backlog/* directories
	return nil, fmt.Errorf("not implemented")
}

type TaskInfo struct {
	ID          string
	Title       string
	Description string
	Status      string // "todo", "in-progress", "review", "done"
	Tags        []string
}

// CommitPlan commits the pending plan to backlog
func (e *Engine) CommitPlan() error {
	if e.pendingPlan == nil {
		return fmt.Errorf("no pending plan to commit")
	}

	// Ensure directories exist
	if err := e.coord.EnsureDirectories(); err != nil {
		return fmt.Errorf("creating directories: %w", err)
	}

	// TODO: Save plan to .smith/backlog/todo/ as markdown files
	// For now, just clear it
	e.pendingPlan = nil

	return nil
}

// HasPendingPlan checks if there's a plan waiting for confirmation
func (e *Engine) HasPendingPlan() bool {
	return e.pendingPlan != nil
}

// GetPendingPlan returns the current pending plan
func (e *Engine) GetPendingPlan() *Plan {
	return e.pendingPlan
}

// processMessage handles the conversation logic
// This is where the magic happens - LLM integration, plan creation, etc.
func (e *Engine) processMessage(input string) string {
	// Convert conversation history to LLM messages
	messages := make([]llm.Message, len(e.conversationHistory))
	for i, msg := range e.conversationHistory {
		messages[i] = llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Call LLM
	response, err := e.llm.Chat(messages, nil)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	return response.Content
}
