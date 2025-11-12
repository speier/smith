package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/speier/smith/pkg/agent/coordinator"
	"github.com/speier/smith/internal/eventbus"
)

// PlanningAgent breaks down features into tasks
type PlanningAgent struct {
	*BaseAgent
}

// NewPlanningAgent creates a new planning agent
func NewPlanningAgent(cfg Config) *PlanningAgent {
	cfg.Role = eventbus.RolePlanning
	return &PlanningAgent{
		BaseAgent: NewBaseAgent(cfg),
	}
}

// Execute implements the Agent interface
func (a *PlanningAgent) Execute(ctx context.Context, task *coordinator.Task) (string, error) {
	// If engine is available, use it for LLM-powered planning
	if a.Engine() != nil {
		return a.Engine().ExecuteTask(ctx, "planning", task.Title, task.Description)
	}

	// Fallback: Simulate planning work
	// This path is used in tests without LLM
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(120 * time.Millisecond):
		// Work completed
	}

	result := fmt.Sprintf("Planned: %s - %s", task.Title, task.Description)
	return result, nil
}

// Start begins the planning agent work loop
func (a *PlanningAgent) Start(ctx context.Context) error {
	return a.StartLoop(ctx, a.Execute)
}
