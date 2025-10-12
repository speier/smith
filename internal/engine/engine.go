package engine

import (
	"fmt"
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

	// Convert conversation history to LLM messages
	messages := make([]llm.Message, len(e.conversationHistory))
	for i, msg := range e.conversationHistory {
		messages[i] = llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Stream the response
	var fullResponse strings.Builder
	err := e.llm.ChatStream(messages, nil, func(response *llm.Response) error {
		fullResponse.WriteString(response.Content)
		return callback(response.Content)
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

type Status struct {
	TodoCount       int
	InProgressCount int
	ReviewCount     int
	DoneCount       int
	BlockedCount    int
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
