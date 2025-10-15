package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/speier/smith/internal/coordinator"
	"github.com/speier/smith/internal/eventbus"
)

// TestingAgent writes tests for code
type TestingAgent struct {
	*BaseAgent
}

// NewTestingAgent creates a new testing agent
func NewTestingAgent(cfg Config) *TestingAgent {
	cfg.Role = eventbus.RoleTesting
	return &TestingAgent{
		BaseAgent: NewBaseAgent(cfg),
	}
}

// Execute implements the Agent interface
func (a *TestingAgent) Execute(ctx context.Context, task *coordinator.Task) (string, error) {
	// If engine is available, use it for LLM-powered test generation
	if a.Engine() != nil {
		return a.Engine().ExecuteTask(ctx, "testing", task.Title, task.Description)
	}

	// Fallback: Simulate testing work
	// This path is used in tests without LLM
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(150 * time.Millisecond):
		// Work completed
	}

	result := fmt.Sprintf("Tested: %s - %s", task.Title, task.Description)
	return result, nil
}

// Start begins the testing agent work loop
func (a *TestingAgent) Start(ctx context.Context) error {
	return a.StartLoop(ctx, a.Execute)
}
