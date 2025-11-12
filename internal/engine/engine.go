package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/speier/smith/pkg/agent/coordinator"
	"github.com/speier/smith/pkg/llm"
)

// Engine is the core Smith system
// It's frontend-agnostic - can be used by CLI, web UI, API, etc.
type Engine struct {
	coord       coordinator.Coordinator
	llm         llm.Provider
	projectPath string
	autoLevel   string // Current safety auto-level

	// Approval callback for blocked commands
	approvalCallback func(command, reason string) (approved bool, addToAllowlist bool)

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
	AutoLevel   string // Safety auto-level (low/medium/high)
}

// New creates a new Smith engine instance
func New(cfg Config) (*Engine, error) {
	// Use default copilot provider if not specified
	if cfg.LLMProvider == nil {
		cfg.LLMProvider = llm.NewCopilotProvider()
	}

	// Default to medium if not specified
	autoLevel := cfg.AutoLevel
	if autoLevel == "" {
		autoLevel = "medium"
	}

	coord := coordinator.New(cfg.ProjectPath)

	return &Engine{
		llm:         cfg.LLMProvider,
		coord:       coord,
		projectPath: cfg.ProjectPath,
		autoLevel:   autoLevel,
	}, nil
}

// GetCoordinator returns the coordinator instance for accessing task stats and other coordination features
func (e *Engine) GetCoordinator() coordinator.Coordinator {
	return e.coord
}

// getSystemPrompt returns the system prompt with tool usage instructions
func (e *Engine) getSystemPrompt() string {
	return `You are Smith, an AI-powered development assistant with a multi-agent architecture.

**Your Role as Main Coordinator:**
You are the main agent that talks to users. You can delegate work to specialized background agents.

**Intent Classification:**
First, classify the user's request into one of these categories:

1. **Simple Query** - Questions, explanations, advice
   → Respond directly, no tools needed
   Examples: "What's the difference between X and Y?", "Explain how this works"

2. **Direct Execution** - Quick file operations, status checks, simple edits
   → Use tools yourself (read_file, write_file, run_command)
   Examples: "Show me main.go", "Run the tests", "Add a comment to this function"

3. **Complex Feature** - Multi-step implementation, new features, refactoring
   → Break into tasks, delegate to specialized agents
   Examples: "Implement authentication", "Refactor the API layer", "Add logging system"

4. **Specialist Consultation** - Need specific expertise
   → Directly ask the appropriate agent (bypass task queue for quick consultations)
   Examples: "Review this code for bugs", "What's the best way to test this?"

**Available Tools:**

**File Operations:**
- write_file: Create or overwrite files
- read_file: Read file contents
- edit_file: Replace content in files
- list_files: Browse project structure
- run_command: Execute shell commands

**Task Management (Multi-Agent Coordination):**
- create_task: Delegate work to background agents
  - 'implementation' agents: Write code, create features
  - 'testing' agents: Write tests, validate code
- list_tasks: See all tasks and their status
- get_task: Get detailed info about a specific task
- get_task_stats: Get summary of task counts

**Multi-Agent Workflow:**
When a user asks for features that involve coding or testing:
1. Break down the work into tasks
2. Use create_task to delegate to specialized agents
3. Background agents will execute tasks automatically
4. Use get_task_stats or list_tasks to monitor progress
5. Respond to user about what you've delegated

**Routing Guidelines:**
- **Simple Query?** Just answer, no tools
- **Quick check/edit?** Use file/command tools directly
- **Feature to build?** Break into tasks, delegate to agents:
  - Architect (planning): Design and planning tasks
  - Keymaker (implementation): Code implementation
  - Sentinel (testing): Test writing and validation
  - Oracle (review): Code review and quality checks
- **Need expert opinion?** Consult specific agent directly

**Example - Complex Feature:**
User: "Implement JWT authentication and add tests"
→ Intent: Complex Feature
→ Actions:
1. create_task(title="Design JWT auth system", agent_role="architect")
2. create_task(title="Implement JWT middleware", agent_role="keymaker")
3. create_task(title="Test JWT authentication", agent_role="sentinel")
4. create_task(title="Review security", agent_role="oracle")

**Example - Simple Query:**
User: "What's the difference between interfaces and structs in Go?"
→ Intent: Simple Query
→ Action: Answer directly, explain concepts

**Example - Direct Execution:**
User: "Show me the main function"
→ Intent: Direct Execution
→ Action: read_file("main.go")

Be conversational and helpful. Explain what you're doing and why.
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
		// Task Management Tools
		{
			Name:        "create_task",
			Description: "Create a new task for a background agent to execute. Use this to delegate work like implementing features or writing tests. Set priority and dependencies for optimal task ordering.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Short title for the task (e.g., 'Implement JWT auth')",
					},
					"description": map[string]interface{}{
						"type":        "string",
						"description": "Detailed description of what needs to be done",
					},
					"agent_role": map[string]interface{}{
						"type":        "string",
						"description": "Type of agent: 'keymaker' for coding, 'sentinel' for testing, 'architect' for planning, 'oracle' for review",
						"enum":        []string{"keymaker", "sentinel", "architect", "oracle"},
					},
					"priority": map[string]interface{}{
						"type":        "string",
						"description": "Task priority: 'high' (critical/blocking), 'medium' (normal, default), 'low' (nice-to-have)",
						"enum":        []string{"high", "medium", "low"},
					},
					"depends_on": map[string]interface{}{
						"type":        "array",
						"description": "Array of task IDs that must complete before this task (e.g., ['task-001', 'task-002'])",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"required": []string{"title", "description", "agent_role"},
			},
		},
		{
			Name:        "list_tasks",
			Description: "List all tasks, optionally filtered by status",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"status": map[string]interface{}{
						"type":        "string",
						"description": "Filter by status: 'backlog', 'wip', 'review', 'done', or empty for all",
						"enum":        []string{"", "backlog", "wip", "review", "done"},
					},
				},
			},
		},
		{
			Name:        "get_task",
			Description: "Get detailed information about a specific task including its result or error",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"task_id": map[string]interface{}{
						"type":        "string",
						"description": "Task ID (e.g., 'task-001')",
					},
				},
				"required": []string{"task_id"},
			},
		},
		{
			Name:        "get_task_stats",
			Description: "Get summary statistics of all tasks (counts by status)",
			Parameters: map[string]interface{}{
				"type": "object",
			},
		},
		{
			Name:        "consult_agent",
			Description: "Consult a specialized agent for quick advice or analysis. Use this for questions that don't require creating tasks (e.g., code review, design advice, test strategy). The agent responds immediately without going through the task queue.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"agent_role": map[string]interface{}{
						"type":        "string",
						"description": "Which specialist to consult: 'architect' for design/planning, 'keymaker' for implementation help, 'sentinel' for testing strategy, 'oracle' for code review/quality",
						"enum":        []string{"architect", "keymaker", "sentinel", "oracle"},
					},
					"question": map[string]interface{}{
						"type":        "string",
						"description": "The question or request for the specialist agent",
					},
					"context": map[string]interface{}{
						"type":        "string",
						"description": "Optional context (code snippets, file contents, etc.) to help the agent answer",
					},
				},
				"required": []string{"agent_role", "question"},
			},
		},
	}
}

// getAgentTools returns tools available to background agents (no task management)
// This prevents agents from creating infinite delegation loops
func (e *Engine) getAgentTools() []llm.Tool {
	allTools := e.getTools()
	var agentTools []llm.Tool

	// Filter out task management tools
	for _, tool := range allTools {
		if tool.Name != "create_task" &&
			tool.Name != "list_tasks" &&
			tool.Name != "get_task" &&
			tool.Name != "get_task_stats" {
			agentTools = append(agentTools, tool)
		}
	}

	return agentTools
}

// getRoleSystemPrompt returns a role-specific system prompt for background agents
func (e *Engine) getRoleSystemPrompt(role, taskTitle, taskDescription string) string {
	// Common tools section
	toolsSection := `**Available Tools:**
- write_file: Create or overwrite files
- read_file: Read file contents  
- edit_file: Replace content in files
- list_files: Browse project structure
- run_command: Execute shell commands (build, test, etc.)`

	switch role {
	case "keymaker", "implementation":
		return fmt.Sprintf(`You are the Keymaker - a specialized coding agent focused on building features.

**Your Task:**
Title: %s
Description: %s

%s

**Your Role & Responsibilities:**
1. **Understand the Requirement** - Read the task description carefully
2. **Read Existing Code** - Use read_file to understand current codebase
3. **Write Clean Code** - Implement features following best practices
4. **Build & Test** - Use run_command to verify your code compiles
5. **Return Summary** - Describe what you implemented

**Best Practices:**
- Read before writing - understand the existing code structure
- Follow the project's coding style and patterns
- Write idiomatic code for the language (Go, Python, etc.)
- Add comments for complex logic
- Run builds to catch syntax errors
- Keep implementations focused and complete

Be professional, thorough, and detail-oriented. Your code should work correctly.`,
			taskTitle, taskDescription, toolsSection)

	case "sentinel", "testing":
		return fmt.Sprintf(`You are a Sentinel - a specialized testing agent focused on hunting down bugs relentlessly.

**Your Task:**
Title: %s
Description: %s

%s

**Your Role & Responsibilities:**
1. **Analyze the Code** - Use read_file to understand what needs testing
2. **Write Test Cases** - Create unit tests, integration tests as appropriate
3. **Cover Edge Cases** - Think about boundary conditions, errors, edge cases
4. **Run Tests** - Use run_command to execute tests and verify they pass
5. **Return Summary** - Describe test coverage and results

**Best Practices:**
- Read the implementation code first
- Write clear, descriptive test names
- Test happy paths AND error cases
- Aim for high code coverage
- Make tests deterministic (no flaky tests)
- Use table-driven tests where appropriate (especially for Go)
- Run tests to ensure they pass before completing

Be thorough and skeptical. Your tests should catch bugs before production.`,
			taskTitle, taskDescription, toolsSection)

	case "architect", "planning":
		return fmt.Sprintf(`You are the Architect - a specialized planning agent focused on designing elegant solutions.

**Your Task:**
Title: %s
Description: %s

%s

**Your Role & Responsibilities:**
1. **Analyze Requirements** - Understand what the user wants
2. **Review Codebase** - Use read_file to understand current architecture
3. **Break Down Work** - Split large features into smaller, focused tasks
4. **Set Priorities** - Determine urgency: HIGH (critical/blocking), MEDIUM (normal), LOW (nice-to-have)
5. **Identify Dependencies** - Determine which tasks must complete before others
6. **Return Structured Plan** - Provide tasks in this exact format:

**Task Breakdown Format:**
For each task, output:
TASK: <clear title>
DESCRIPTION: <detailed description>
ROLE: <keymaker|sentinel|oracle>
PRIORITY: <HIGH|MEDIUM|LOW>
DEPENDS_ON: <task-001, task-002> (or NONE if no dependencies)
---

**Priority Guidelines:**
- HIGH: Critical bugs, blocking issues, foundational work needed by other tasks
- MEDIUM: Normal features, improvements (default priority)
- LOW: Nice-to-have improvements, refactoring, optimizations

**Dependency Guidelines:**
- Implementation must come before testing
- Testing must come before review
- Foundation/infrastructure before features that use it
- Use task IDs like task-001, task-002 for dependencies

**Best Practices:**
- Read existing code to understand patterns
- Create tasks that are single-responsibility
- Order tasks logically (infrastructure before features)
- Consider testing needs for each task
- Be specific in task descriptions
- Set realistic priorities based on impact and urgency

Be strategic and thoughtful. Your plans guide the entire team.`,
			taskTitle, taskDescription, toolsSection)

	case "oracle", "review":
		return fmt.Sprintf(`You are the Oracle - a specialized code review agent who sees quality and predicts issues.

**Your Task:**
Title: %s
Description: %s

%s

**Your Role & Responsibilities:**
1. **Read the Code** - Use read_file to review implementation
2. **Check Quality** - Look for bugs, anti-patterns, style issues
3. **Verify Tests** - Ensure tests exist and provide good coverage
4. **Run Validation** - Use run_command to build and test
5. **Return Feedback** - Provide constructive review comments

**Best Practices:**
- Be thorough but constructive
- Check for common bugs (nil checks, error handling, race conditions)
- Verify code follows project conventions
- Ensure tests are comprehensive
- Run builds and tests to verify everything works
- Suggest improvements, not just criticisms

Be critical but helpful. Your reviews improve code quality.`,
			taskTitle, taskDescription, toolsSection)

	default:
		// Generic fallback for unknown roles
		return fmt.Sprintf(`You are a specialized development agent executing a task.

**Your Task:**
Title: %s
Description: %s

%s

**Your Job:**
1. Implement exactly what the task describes
2. Use file tools to read/write code
3. Use run_command to test your changes
4. Return a summary of what you did

Be concise and focused. Execute the task completely.`,
			taskTitle, taskDescription, toolsSection)
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

// SetAutoLevel updates the current auto-level
func (e *Engine) SetAutoLevel(level string) {
	e.autoLevel = level
}

// GetAutoLevel returns the current auto-level
func (e *Engine) GetAutoLevel() string {
	return e.autoLevel
}

// SetApprovalCallback sets the callback for command approval requests
// The callback receives (command, reason) and returns (approved, addToAllowlist)
func (e *Engine) SetApprovalCallback(callback func(command, reason string) (bool, bool)) {
	e.approvalCallback = callback
}

// ExecuteTask executes a task using LLM with agent tools (no task management)
// This is used by background agents to implement/test features
func (e *Engine) ExecuteTask(ctx context.Context, role, taskTitle, taskDescription string) (string, error) {
	// Get role-specific system prompt
	systemPrompt := e.getRoleSystemPrompt(role, taskTitle, taskDescription)

	// Build messages with system prompt
	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: fmt.Sprintf("Execute this task: %s", taskDescription)},
	}

	// Get agent tools (no task management)
	tools := e.getAgentTools()

	// Execute with LLM
	var fullResponse strings.Builder
	err := e.llm.ChatStream(messages, tools, func(response *llm.Response) error {
		// Handle tool calls
		if len(response.ToolCalls) > 0 {
			for _, toolCall := range response.ToolCalls {
				result, err := e.executeToolCall(toolCall)
				if err != nil {
					return fmt.Errorf("tool execution failed: %w", err)
				}
				fullResponse.WriteString(fmt.Sprintf("[%s: %s]\n", toolCall.Name, result))
			}
		}

		// Accumulate content
		if response.Content != "" {
			fullResponse.WriteString(response.Content)
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("LLM execution failed: %w", err)
	}

	return fullResponse.String(), nil
}

// GetStatus returns current task statistics
func (e *Engine) GetStatus() string {
	stats, err := e.coord.GetTaskStats()
	if err != nil {
		return fmt.Sprintf("Error getting status: %v", err)
	}

	return fmt.Sprintf(`
📊 Task Status:
   Backlog: %d tasks ready
   WIP:     %d in progress
   Review:  %d awaiting review
   Done:    %d completed
`, stats.Backlog, stats.WIP, stats.Review, stats.Done)
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

	// Safety check - validate command against auto-level rules
	checkResult := IsCommandAllowed(command, e.autoLevel)
	if !checkResult.Allowed {
		// Command blocked - request approval if callback is set
		if e.approvalCallback != nil {
			approved, addToAllowlist := e.approvalCallback(command, checkResult.Reason)
			if !approved {
				return "", fmt.Errorf("command denied by user")
			}
			// If approved and should be added to allowlist
			if addToAllowlist {
				AddToSessionAllowlist(command)
			}
			// Fall through to execute the approved command
		} else {
			// No approval callback - deny immediately
			return "", fmt.Errorf("command blocked by safety rules (%s): %s\nCommand: %s\nReason: %s",
				e.autoLevel, checkResult.Reason, command, checkResult.Reason)
		}
	}

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

// handleCreateTask handles the create_task tool call
func (e *Engine) handleCreateTask(input map[string]interface{}) (string, error) {
	title, ok := input["title"].(string)
	if !ok {
		return "", fmt.Errorf("title is required")
	}

	description, ok := input["description"].(string)
	if !ok {
		return "", fmt.Errorf("description is required")
	}

	agentRole, ok := input["agent_role"].(string)
	if !ok {
		return "", fmt.Errorf("agent_role is required")
	}

	// Validate agent_role
	if agentRole != "implementation" && agentRole != "testing" {
		return "", fmt.Errorf("agent_role must be 'implementation' or 'testing'")
	}

	// Parse optional priority (default: 1=medium)
	priority := 1 // medium
	if priorityStr, ok := input["priority"].(string); ok {
		switch priorityStr {
		case "high", "HIGH":
			priority = 2
		case "low", "LOW":
			priority = 0
		default:
			priority = 1 // medium
		}
	}

	// Parse optional dependencies
	var dependsOn []string
	if deps, ok := input["depends_on"].([]interface{}); ok {
		for _, dep := range deps {
			if depStr, ok := dep.(string); ok && depStr != "" {
				dependsOn = append(dependsOn, depStr)
			}
		}
	}

	// Build task options
	var opts []coordinator.TaskOption
	if priority != 1 { // Only set if non-default
		opts = append(opts, coordinator.WithPriority(priority))
	}
	if len(dependsOn) > 0 {
		opts = append(opts, coordinator.WithDependencies(dependsOn...))
	}

	// Create task via coordinator
	taskID, err := e.coord.CreateTask(title, description, agentRole, opts...)
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	var priorityLabel string
	switch priority {
	case 2:
		priorityLabel = "high"
	case 0:
		priorityLabel = "low"
	default:
		priorityLabel = "medium"
	}

	result := fmt.Sprintf("✅ Created task %s: %s (assigned to %s agent, priority: %s)", taskID, title, agentRole, priorityLabel)
	if len(dependsOn) > 0 {
		result += fmt.Sprintf(", depends on: %v", dependsOn)
	}

	return result, nil
}

// handleListTasks handles the list_tasks tool call
func (e *Engine) handleListTasks(input map[string]interface{}) (string, error) {
	status, _ := input["status"].(string)

	tasks, err := e.coord.GetTasksByStatus(status)
	if err != nil {
		return "", fmt.Errorf("failed to list tasks: %w", err)
	}

	if len(tasks) == 0 {
		if status == "" {
			return "No tasks found", nil
		}
		return fmt.Sprintf("No tasks with status '%s'", status), nil
	}

	var result strings.Builder
	if status == "" {
		result.WriteString(fmt.Sprintf("📋 All tasks (%d):\n", len(tasks)))
	} else {
		result.WriteString(fmt.Sprintf("📋 Tasks with status '%s' (%d):\n", status, len(tasks)))
	}

	for _, task := range tasks {
		statusEmoji := map[string]string{
			"backlog": "📥",
			"wip":     "🔄",
			"review":  "👀",
			"done":    "✅",
		}[task.Status]
		result.WriteString(fmt.Sprintf("  %s %s - %s (%s)\n", statusEmoji, task.ID, task.Title, task.Role))
	}

	return result.String(), nil
}

// handleGetTask handles the get_task tool call
func (e *Engine) handleGetTask(input map[string]interface{}) (string, error) {
	taskID, ok := input["task_id"].(string)
	if !ok {
		return "", fmt.Errorf("task_id is required")
	}

	task, err := e.coord.GetTask(taskID)
	if err != nil {
		return "", fmt.Errorf("failed to get task: %w", err)
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("📝 Task %s:\n", task.ID))
	result.WriteString(fmt.Sprintf("  Title: %s\n", task.Title))
	result.WriteString(fmt.Sprintf("  Status: %s\n", task.Status))
	result.WriteString(fmt.Sprintf("  Agent: %s\n", task.Role))
	result.WriteString(fmt.Sprintf("  Description: %s\n", task.Description))

	if task.Result != "" {
		result.WriteString(fmt.Sprintf("  Result: %s\n", task.Result))
	}

	if task.Error != "" {
		result.WriteString(fmt.Sprintf("  Error: %s\n", task.Error))
	}

	return result.String(), nil
}

// handleGetTaskStats handles the get_task_stats tool call
func (e *Engine) handleGetTaskStats(input map[string]interface{}) (string, error) {
	stats, err := e.coord.GetTaskStats()
	if err != nil {
		return "", fmt.Errorf("failed to get task stats: %w", err)
	}

	return fmt.Sprintf(`📊 Task Statistics:
  Backlog: %d tasks ready
  WIP:     %d in progress
  Review:  %d under review
  Done:    %d completed`, stats.Backlog, stats.WIP, stats.Review, stats.Done), nil
}

// handleConsultAgent handles the consult_agent tool call for direct agent-to-agent communication
func (e *Engine) handleConsultAgent(input map[string]interface{}) (string, error) {
	agentRole, ok := input["agent_role"].(string)
	if !ok {
		return "", fmt.Errorf("agent_role is required")
	}

	question, ok := input["question"].(string)
	if !ok {
		return "", fmt.Errorf("question is required")
	}

	context, _ := input["context"].(string)

	// Build the specialist agent's prompt
	var systemPrompt string
	switch agentRole {
	case "architect":
		systemPrompt = `You are The Architect - a planning and design specialist.
Your role is to provide architectural guidance, design patterns, and strategic planning advice.
Focus on: structure, patterns, dependencies, order of operations, scalability, maintainability.
Be concise but thorough in your analysis.`

	case "keymaker":
		systemPrompt = `You are The Keymaker - an implementation specialist.
Your role is to provide coding advice, suggest implementations, and solve technical challenges.
Focus on: code patterns, best practices, language features, libraries, algorithms.
Be practical and include code examples when helpful.`

	case "sentinel":
		systemPrompt = `You are a Sentinel - a testing and quality specialist.
Your role is to provide testing strategies, identify edge cases, and ensure code reliability.
Focus on: test coverage, edge cases, failure scenarios, testing patterns, validation.
Be thorough and think of what could go wrong.`

	case "oracle":
		systemPrompt = `You are The Oracle - a code review and quality specialist.
Your role is to review code for quality, identify issues, and predict potential problems.
Focus on: bugs, security, performance, maintainability, best practices, code smells.
Be insightful and constructive in your feedback.`

	default:
		return "", fmt.Errorf("unknown agent role: %s", agentRole)
	}

	// Build the consultation message
	var promptBuilder strings.Builder
	promptBuilder.WriteString(question)
	if context != "" {
		promptBuilder.WriteString("\n\nContext:\n")
		promptBuilder.WriteString(context)
	}

	// Create a single-turn conversation for the consultation
	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: promptBuilder.String()},
	}

	// Get response from the specialist agent
	response, err := e.llm.Chat(messages, nil)
	if err != nil {
		return "", fmt.Errorf("consultation failed: %w", err)
	}

	// Format the response with agent identity
	agentNames := map[string]string{
		"architect": "🏛️ The Architect",
		"keymaker":  "🔑 The Keymaker",
		"sentinel":  "🦑 Sentinel",
		"oracle":    "🔮 The Oracle",
	}

	return fmt.Sprintf("%s says:\n\n%s", agentNames[agentRole], response.Content), nil
}

// SessionAllowlistStats represents statistics about the session allowlist
type SessionAllowlistStats struct {
	AllowedCount int
	BlockedCount int
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
	case "create_task":
		return e.handleCreateTask(toolCall.Input)
	case "list_tasks":
		return e.handleListTasks(toolCall.Input)
	case "get_task":
		return e.handleGetTask(toolCall.Input)
	case "get_task_stats":
		return e.handleGetTaskStats(toolCall.Input)
	case "consult_agent":
		return e.handleConsultAgent(toolCall.Input)
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
