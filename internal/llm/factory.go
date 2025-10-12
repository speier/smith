package llm

import (
	"fmt"

	"github.com/speier/smith/internal/config"
)

// NewProvider creates a provider based on configuration
func NewProvider() (Provider, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	switch cfg.Provider {
	case "copilot":
		provider := NewCopilotProvider()
		if err := provider.LoadAuth(); err != nil {
			return nil, fmt.Errorf("loading copilot auth: %w", err)
		}
		return provider, nil

	// TODO: Add other providers when implemented
	// case "ollama":
	// case "openai":

	default:
		return nil, fmt.Errorf("unknown provider: %s (only 'copilot' is currently supported)", cfg.Provider)
	}
}
