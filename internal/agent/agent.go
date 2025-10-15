package agent

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/speier/smith/internal/coordinator"
	"github.com/speier/smith/internal/engine"
	"github.com/speier/smith/internal/eventbus"
	"github.com/speier/smith/internal/registry"
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
	registry     *registry.Registry
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
	Registry     *registry.Registry
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
	// Register with the coordinator
	if err := a.registry.Register(ctx, a.ID, a.role, 0); err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}

	log.Printf("[%s] Agent started, polling every %v", a.ID, a.pollInterval)

	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Context cancelled, stopping agent", a.ID)
			return ctx.Err()

		case <-a.stopChan:
			log.Printf("[%s] Stop signal received, stopping agent", a.ID)
			return nil

		case <-ticker.C:
			// Heartbeat
			if err := a.registry.Heartbeat(ctx, a.ID); err != nil {
				log.Printf("[%s] Heartbeat failed: %v", a.ID, err)
			}

			// Poll for available tasks
			tasks, err := a.coord.GetAvailableTasks()
			if err != nil {
				log.Printf("[%s] Failed to get tasks: %v", a.ID, err)
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
				log.Printf("[%s] Failed to claim task %s: %v", a.ID, task.ID, err)
				continue
			}

			log.Printf("[%s] Claimed task %s: %s", a.ID, task.ID, task.Title)

			// Get full task details
			fullTask, err := a.coord.GetTask(task.ID)
			if err != nil {
				log.Printf("[%s] Failed to get task details: %v", a.ID, err)
				continue
			}

			// Execute the task
			result, err := executor(ctx, fullTask)
			if err != nil {
				log.Printf("[%s] Task %s failed: %v", a.ID, task.ID, err)
				if failErr := a.coord.FailTask(task.ID, err.Error()); failErr != nil {
					log.Printf("[%s] Failed to mark task as failed: %v", a.ID, failErr)
				}
				continue
			}

			// Complete the task
			log.Printf("[%s] Task %s completed: %s", a.ID, task.ID, result)
			if err := a.coord.CompleteTask(task.ID, result); err != nil {
				log.Printf("[%s] Failed to complete task: %v", a.ID, err)
			}
		}
	}
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

	log.Printf("[%s] Agent stopped", a.ID)
	return nil
}
