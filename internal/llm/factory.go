package llm

import (
	"fmt"

	"github.com/speier/smith/internal/config"
)

// ProviderInfo describes an available provider
type ProviderInfo struct {
	ID           string
	Name         string
	Description  string
	RequiresAuth bool
}

// GetAvailableProviders returns list of all supported providers
func GetAvailableProviders() []ProviderInfo {
	return []ProviderInfo{
		{
			ID:           "copilot",
			Name:         "GitHub Copilot",
			Description:  "Access GPT-4 and O1 models via GitHub Copilot subscription",
			RequiresAuth: true,
		},
		{
			ID:           "openrouter",
			Name:         "OpenRouter",
			Description:  "Access multiple models (Claude, GPT-4, Gemini, Llama) with pay-per-use",
			RequiresAuth: true,
		},
	}
}

// NewProvider creates a provider based on configuration
func NewProvider() (Provider, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	return NewProviderByID(cfg.Provider)
}

// NewProviderByID creates a provider by ID
func NewProviderByID(providerID string) (Provider, error) {
	switch providerID {
	case "copilot":
		provider := NewCopilotProvider()
		// Try to load auth, but don't fail if it doesn't exist
		// Auth will be required when Chat is called
		provider.LoadAuth()
		return provider, nil

	case "openrouter":
		return NewOpenRouterProvider(), nil

	default:
		return nil, fmt.Errorf("unknown provider: %s", providerID)
	}
}
