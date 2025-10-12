package config

// ProviderPreset defines the default configuration for a provider
type ProviderPreset struct {
	Provider     string
	DefaultModel string
	Agents       map[string]AgentConfig
	Capabilities ProviderCapabilities
}

// ProviderCapabilities describes what a provider supports
type ProviderCapabilities struct {
	SupportsReasoning bool
	MaxContextTokens  int
	HasMiniVariant    bool
	SupportsStreaming bool
}

// GetProviderPresets returns all available provider presets
func GetProviderPresets() map[string]ProviderPreset {
	return map[string]ProviderPreset{
		"copilot": {
			Provider:     "copilot",
			DefaultModel: "gpt-4o",
			Agents: map[string]AgentConfig{
				"planning": {
					Model:     "gpt-4o",
					AutoLevel: "high",
					Reasoning: "high",
				},
				"implementation": {
					Model:     "gpt-4o",
					AutoLevel: "medium",
					Reasoning: "medium",
				},
				"testing": {
					Model:     "gpt-4o-mini",
					AutoLevel: "low",
					Reasoning: "low",
				},
				"review": {
					Model:     "gpt-4o",
					AutoLevel: "high",
					Reasoning: "high",
				},
			},
			Capabilities: ProviderCapabilities{
				SupportsReasoning: true,
				MaxContextTokens:  128000,
				HasMiniVariant:    true,
				SupportsStreaming: true,
			},
		},
		"openrouter": {
			Provider:     "openrouter",
			DefaultModel: "anthropic/claude-3.5-sonnet",
			Agents: map[string]AgentConfig{
				"planning": {
					Model:     "anthropic/claude-3.5-sonnet",
					AutoLevel: "high",
					Reasoning: "high",
				},
				"implementation": {
					Model:     "openai/gpt-4o",
					AutoLevel: "medium",
					Reasoning: "medium",
				},
				"testing": {
					Model:     "openai/gpt-4o-mini",
					AutoLevel: "low",
					Reasoning: "low",
				},
				"review": {
					Model:     "anthropic/claude-3.5-sonnet",
					AutoLevel: "high",
					Reasoning: "high",
				},
			},
			Capabilities: ProviderCapabilities{
				SupportsReasoning: true,
				MaxContextTokens:  200000,
				HasMiniVariant:    true,
				SupportsStreaming: true,
			},
		},
		"openai": {
			Provider:     "openai",
			DefaultModel: "gpt-4o",
			Agents: map[string]AgentConfig{
				"planning": {
					Model:     "gpt-4o",
					AutoLevel: "high",
					Reasoning: "high",
				},
				"implementation": {
					Model:     "gpt-4o",
					AutoLevel: "medium",
					Reasoning: "medium",
				},
				"testing": {
					Model:     "gpt-4o-mini",
					AutoLevel: "low",
					Reasoning: "low",
				},
				"review": {
					Model:     "gpt-4o",
					AutoLevel: "high",
					Reasoning: "high",
				},
			},
			Capabilities: ProviderCapabilities{
				SupportsReasoning: true,
				MaxContextTokens:  128000,
				HasMiniVariant:    true,
				SupportsStreaming: true,
			},
		},
	}
}

// GetProviderPreset returns a specific provider preset
func GetProviderPreset(providerID string) (ProviderPreset, bool) {
	presets := GetProviderPresets()
	preset, exists := presets[providerID]
	return preset, exists
}

// GetAvailableProviders returns a list of available provider IDs
func GetAvailableProviders() []string {
	return []string{"copilot", "openrouter", "openai"}
}
