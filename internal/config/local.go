package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LocalConfig represents project-level configuration in .smith/config.json
type LocalConfig struct {
	Provider  string                 `json:"provider"`            // Selected provider (required)
	Model     string                 `json:"model"`               // Primary model
	AutoLevel string                 `json:"autoLevel,omitempty"` // Default: "medium"
	Agents    map[string]AgentConfig `json:"agents"`              // Per-agent configuration
	Version   int                    `json:"version"`             // Config schema version
}

const (
	localConfigDir  = ".smith"
	localConfigFile = "config.json"
)

// LoadLocal loads project-level configuration from .smith/config.json
// Returns nil if no config exists (requires provider selection)
func LoadLocal(projectPath string) (*LocalConfig, error) {
	configPath := filepath.Join(projectPath, localConfigDir, localConfigFile)

	// Return nil if file doesn't exist - caller must handle provider selection
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading local config: %w", err)
	}

	var cfg LocalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing local config: %w", err)
	}

	// Validate required fields
	if cfg.Provider == "" {
		return nil, fmt.Errorf("config missing required field: provider")
	}

	// Set defaults if not specified
	if cfg.AutoLevel == "" {
		cfg.AutoLevel = "medium"
	}
	if cfg.Agents == nil {
		cfg.Agents = make(map[string]AgentConfig)
	}
	if cfg.Version == 0 {
		cfg.Version = 1
	}

	return &cfg, nil
}

// CreateFromPreset creates a new local config from a provider preset
func CreateFromPreset(providerID string) (*LocalConfig, error) {
	preset, exists := GetProviderPreset(providerID)
	if !exists {
		return nil, fmt.Errorf("unknown provider: %s", providerID)
	}

	return &LocalConfig{
		Provider:  preset.Provider,
		Model:     preset.DefaultModel,
		AutoLevel: "medium",
		Agents:    preset.Agents,
		Version:   1,
	}, nil
}

// SaveLocal saves project-level configuration to .smith/config.json
func (c *LocalConfig) SaveLocal(projectPath string) error {
	configDir := filepath.Join(projectPath, localConfigDir)
	configPath := filepath.Join(configDir, localConfigFile)

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling local config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("writing local config: %w", err)
	}

	return nil
}

// GetAgentConfig returns configuration for a specific agent with fallbacks
func (c *LocalConfig) GetAgentConfig(agentName string) AgentConfig {
	if agent, ok := c.Agents[agentName]; ok {
		// Fill in defaults from parent config
		if agent.Model == "" {
			agent.Model = c.Model
		}
		if agent.AutoLevel == "" {
			agent.AutoLevel = c.AutoLevel
		}
		return agent
	}

	// Return defaults
	return AgentConfig{
		Model:     c.Model,
		AutoLevel: c.AutoLevel,
	}
}
