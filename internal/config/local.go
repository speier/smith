package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LocalConfig represents project-level configuration in .smith/config.yaml
type LocalConfig struct {
	Provider  string                 `yaml:"provider"`            // Selected provider (required)
	Model     string                 `yaml:"model"`               // Primary model
	AutoLevel string                 `yaml:"autoLevel,omitempty"` // Default: "medium"
	Agents    map[string]AgentConfig `yaml:"agents"`              // Per-agent configuration
	Version   int                    `yaml:"version"`             // Config schema version
}

const (
	localConfigDir  = ".smith"
	localConfigFile = "config.yaml"
)

// LoadLocal loads project-level configuration from .smith/config.yaml
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
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing local config: %w", err)
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

// CreateFromPreset creates a new local config from a provider (no hardcoded models)
func CreateFromPreset(providerID string) (*LocalConfig, error) {
	preset, exists := GetProviderPreset(providerID)
	if !exists {
		return nil, fmt.Errorf("unknown provider: %s", providerID)
	}

	return &LocalConfig{
		Provider:  preset.Provider,
		Model:     "", // User must select model from provider's API
		AutoLevel: "medium",
		Agents:    make(map[string]AgentConfig), // Empty until models selected
		Version:   1,
	}, nil
}

// SaveLocal saves project-level configuration to .smith/config.yaml
func (c *LocalConfig) SaveLocal(projectPath string) error {
	configDir := filepath.Join(projectPath, localConfigDir)
	configPath := filepath.Join(configDir, localConfigFile)

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
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
