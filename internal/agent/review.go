package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/speier/smith/internal/coordinator"
	"github.com/speier/smith/internal/eventbus"
)

// ReviewAgent reviews code for quality and correctness
type ReviewAgent struct {
	*BaseAgent
}

// NewReviewAgent creates a new review agent
func NewReviewAgent(cfg Config) *ReviewAgent {
	cfg.Role = eventbus.RoleReview
	return &ReviewAgent{
		BaseAgent: NewBaseAgent(cfg),
	}
}

// Execute implements the Agent interface
func (a *ReviewAgent) Execute(ctx context.Context, task *coordinator.Task) (string, error) {
	// If engine is available, use it for LLM-powered code review
	if a.Engine() != nil {
		return a.Engine().ExecuteTask(ctx, "review", task.Title, task.Description)
	}

	// Fallback: Simulate review work
	// This path is used in tests without LLM
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(130 * time.Millisecond):
		// Work completed
	}

	result := fmt.Sprintf("Reviewed: %s - %s", task.Title, task.Description)
	return result, nil
}

// Start begins the review agent work loop
func (a *ReviewAgent) Start(ctx context.Context) error {
	return a.StartLoop(ctx, a.Execute)
}
