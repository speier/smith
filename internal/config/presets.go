package config

// ProviderPreset defines metadata for a provider (no hardcoded models)
type ProviderPreset struct {
	Provider     string
	Capabilities ProviderCapabilities
}

// ProviderCapabilities describes what a provider supports
type ProviderCapabilities struct {
	SupportsReasoning bool
	MaxContextTokens  int
	SupportsStreaming bool
}

// GetProviderPresets returns all available provider metadata
func GetProviderPresets() map[string]ProviderPreset {
	return map[string]ProviderPreset{
		"copilot": {
			Provider: "copilot",
			Capabilities: ProviderCapabilities{
				SupportsReasoning: true,
				MaxContextTokens:  128000,
				SupportsStreaming: true,
			},
		},
		"openrouter": {
			Provider: "openrouter",
			Capabilities: ProviderCapabilities{
				SupportsReasoning: true,
				MaxContextTokens:  200000,
				SupportsStreaming: true,
			},
		},
		"openai": {
			Provider: "openai",
			Capabilities: ProviderCapabilities{
				SupportsReasoning: true,
				MaxContextTokens:  128000,
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
