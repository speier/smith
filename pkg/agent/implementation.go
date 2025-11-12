package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/speier/smith/pkg/agent/coordinator"
	"github.com/speier/smith/internal/eventbus"
)

// ImplementationAgent implements code based on task descriptions
type ImplementationAgent struct {
	*BaseAgent
}

// NewImplementationAgent creates a new implementation agent
func NewImplementationAgent(cfg Config) *ImplementationAgent {
	cfg.Role = eventbus.RoleImplementation
	return &ImplementationAgent{
		BaseAgent: NewBaseAgent(cfg),
	}
}

// Execute implements the Agent interface
func (a *ImplementationAgent) Execute(ctx context.Context, task *coordinator.Task) (string, error) {
	// If engine is available, use it for LLM-powered implementation
	if a.Engine() != nil {
		return a.Engine().ExecuteTask(ctx, "implementation", task.Title, task.Description)
	}

	// Fallback: Simulate implementation work
	// This path is used in tests without LLM
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(100 * time.Millisecond):
		// Work completed
	}

	result := fmt.Sprintf("Implemented: %s - %s", task.Title, task.Description)
	return result, nil
}

// Start begins the implementation agent work loop
func (a *ImplementationAgent) Start(ctx context.Context) error {
	return a.StartLoop(ctx, a.Execute)
}
