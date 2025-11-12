package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/speier/smith/pkg/agent/coordinator"
	"github.com/speier/smith/internal/engine"
	"github.com/speier/smith/internal/eventbus"
)

// Agent represents a background worker that processes tasks
type Agent interface {
	// Role returns the agent's role (planning, implementation, testing, review)
	Role() eventbus.AgentRole

	// Execute performs the task and returns the result or error
	Execute(ctx context.Context, task *coordinator.Task) (result string, err error)

	// Start begins the agent's work loop
	Start(ctx context.Context) error

	// Stop gracefully stops the agent
	Stop() error
}

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	ID           string
	role         eventbus.AgentRole
	coord        coordinator.Coordinator
	registry     coordinator.Registry // Now uses interface instead of concrete type
	engine       *engine.Engine
	pollInterval time.Duration
	stopChan     chan struct{}
	stopped      bool
}

// Config for creating an agent
type Config struct {
	AgentID      string
	Role         eventbus.AgentRole
	Coordinator  coordinator.Coordinator
	Registry     coordinator.Registry // Now uses interface instead of concrete type
	Engine       *engine.Engine
	PollInterval time.Duration
}

// NewBaseAgent creates a new base agent
func NewBaseAgent(cfg Config) *BaseAgent {
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 100 * time.Millisecond
	}

	return &BaseAgent{
		ID:           cfg.AgentID,
		role:         cfg.Role,
		coord:        cfg.Coordinator,
		registry:     cfg.Registry,
		engine:       cfg.Engine,
		pollInterval: cfg.PollInterval,
		stopChan:     make(chan struct{}),
	}
}

// Engine returns the agent's engine for LLM-powered work
func (a *BaseAgent) Engine() *engine.Engine {
	return a.engine
}

// Role returns the agent's role
func (a *BaseAgent) Role() eventbus.AgentRole {
	return a.role
}

// Start begins the agent work loop
// This is the core of the background agent system:
// 1. Poll for available tasks
// 2. Claim a task
// 3. Execute the task (implemented by specific agent types)
// 4. Complete or fail the task
func (a *BaseAgent) StartLoop(ctx context.Context, executor func(context.Context, *coordinator.Task) (string, error)) error {
	// Register with the coordinator (convert eventbus.AgentRole to coordinator.AgentRole)
	if err := a.registry.Register(ctx, a.ID, coordinator.AgentRole(a.role), 0); err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}

	// Agent started (logging removed to avoid TUI contamination)

	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Agent stopping due to context cancellation (logging removed)
			return ctx.Err()

		case <-a.stopChan:
			// Agent stopping due to stop signal (logging removed)
			return nil

		case <-ticker.C:
			// Heartbeat
			_ = a.registry.Heartbeat(ctx, a.ID)

			// Poll for available tasks
			tasks, err := a.coord.GetAvailableTasks()
			if err != nil {
				// Failed to get tasks (logging removed)
				continue
			}

			// Filter tasks by role (only pick up tasks for our role)
			var myTasks []coordinator.Task
			for _, task := range tasks {
				if task.Role == "" || task.Role == string(a.role) {
					myTasks = append(myTasks, task)
				}
			}

			if len(myTasks) == 0 {
				continue // No tasks available for this agent
			}

			// Claim the first available task
			task := myTasks[0]
			if err := a.coord.ClaimTask(task.ID, a.ID); err != nil {
				// Failed to claim task (logging removed)
				continue
			}

			// Task claimed successfully (logging removed)

			// Get full task details
			fullTask, err := a.coord.GetTask(task.ID)
			if err != nil {
				// Failed to get task details (logging removed)
				continue
			}

			// Query recent tasks for context (agent memory)
			recentTasks, err := a.coord.GetRecentTasks(ctx, string(a.role), 5)
			if err != nil {
				// Failed to get recent tasks for context (logging removed)
				// Non-critical, continue without context
				recentTasks = nil
			}

			// Add recent learnings to task context if available
			if len(recentTasks) > 0 {
				contextInfo := "\n\n📚 Recent learnings from past tasks:\n"
				for i, rt := range recentTasks {
					if rt.Learnings != "" {
						contextInfo += fmt.Sprintf("%d. From '%s': %s\n", i+1, rt.Title, rt.Learnings)
					}
					if len(rt.Blockers) > 0 {
						contextInfo += fmt.Sprintf("   ⚠️ Known blockers: %v\n", rt.Blockers)
					}
				}
				// Append context to task description for agent to consider
				fullTask.Description += contextInfo
			}

			// Execute the task
			result, err := executor(ctx, fullTask)
			if err != nil {
				// Task failed (logging removed to avoid TUI contamination)

				// Extract learnings from error if possible
				learnings := a.extractLearnings(err.Error())
				blockers := a.extractBlockers(err.Error())

				opts := []coordinator.TaskOption{}
				if learnings != "" {
					opts = append(opts, coordinator.WithLearnings(learnings))
				}
				if len(blockers) > 0 {
					opts = append(opts, coordinator.WithBlockers(blockers...))
				}

				_ = a.coord.FailTask(task.ID, err.Error(), opts...)
				continue
			}

			// Extract learnings from result
			learnings := a.extractLearnings(result)
			approaches := a.extractApproaches(result)

			// Complete the task with learnings
			// Task completed successfully (logging removed to avoid TUI contamination)

			opts := []coordinator.TaskOption{}
			if learnings != "" {
				opts = append(opts, coordinator.WithLearnings(learnings))
			}
			if len(approaches) > 0 {
				opts = append(opts, coordinator.WithTriedApproaches(approaches...))
			}

			_ = a.coord.CompleteTask(task.ID, result, opts...)
		}
	}
}

// extractLearnings parses learnings from task result/error
// Looks for patterns like "Learning:", "Learned:", "Key insight:"
func (a *BaseAgent) extractLearnings(text string) string {
	// Simple pattern matching - look for common learning indicators
	patterns := []string{
		"Learning:",
		"Learned:",
		"Key insight:",
		"Takeaway:",
		"Note:",
	}

	for _, pattern := range patterns {
		if idx := indexOf(text, pattern); idx >= 0 {
			// Extract text after the pattern
			start := idx + len(pattern)
			end := indexOfAny(text[start:], []string{"\n", ".", ";"})
			if end > 0 {
				return trim(text[start : start+end])
			}
			return trim(text[start:])
		}
	}

	return "" // No explicit learning found
}

// extractApproaches parses tried approaches from result
func (a *BaseAgent) extractApproaches(text string) []string {
	// Look for "Tried:", "Attempted:", "Used approach:"
	patterns := []string{
		"Tried:",
		"Attempted:",
		"Used approach:",
		"Approach:",
	}

	for _, pattern := range patterns {
		if idx := indexOf(text, pattern); idx >= 0 {
			start := idx + len(pattern)
			end := indexOfAny(text[start:], []string{"\n\n", "."})
			if end > 0 {
				return []string{trim(text[start : start+end])}
			}
			return []string{trim(text[start:])}
		}
	}

	return nil
}

// extractBlockers parses blockers from error messages
func (a *BaseAgent) extractBlockers(errMsg string) []string {
	// Look for common blocker indicators
	if errMsg == "" {
		return nil
	}

	// For now, just return the error as a blocker
	// Could be enhanced with pattern matching
	return []string{errMsg}
}

// Helper functions for string parsing
func indexOf(text, substr string) int {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func indexOfAny(text string, substrs []string) int {
	minIdx := -1
	for _, substr := range substrs {
		idx := indexOf(text, substr)
		if idx >= 0 && (minIdx < 0 || idx < minIdx) {
			minIdx = idx
		}
	}
	return minIdx
}

func trim(s string) string {
	// Simple trim implementation
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}

	return s[start:end]
}

// Stop gracefully stops the agent
func (a *BaseAgent) Stop() error {
	if a.stopped {
		return nil
	}
	a.stopped = true
	close(a.stopChan)

	// Unregister from coordinator
	ctx := context.Background()
	if err := a.registry.Unregister(ctx, a.ID); err != nil {
		return fmt.Errorf("failed to unregister: %w", err)
	}

	// Agent stopped (logging removed)
	return nil
}
