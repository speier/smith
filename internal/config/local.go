package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LocalConfig represents project-level configuration in .smith/config.json
type LocalConfig struct {
	Provider  string                 `json:"provider,omitempty"`  // Override global provider
	Model     string                 `json:"model,omitempty"`     // Override global model
	AutoLevel string                 `json:"autoLevel,omitempty"` // Default: "medium"
	Agents    map[string]AgentConfig `json:"agents,omitempty"`
}

// AgentConfig represents configuration for a specific agent
type AgentConfig struct {
	Model     string `json:"model,omitempty"`
	AutoLevel string `json:"autoLevel,omitempty"`
	Reasoning string `json:"reasoning,omitempty"` // low/medium/high
}

const (
	localConfigDir  = ".smith"
	localConfigFile = "config.json"
)

// LoadLocal loads project-level configuration from .smith/config.json
func LoadLocal(projectPath string) (*LocalConfig, error) {
	configPath := filepath.Join(projectPath, localConfigDir, localConfigFile)

	// Return default if file doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &LocalConfig{
			AutoLevel: "medium", // Default auto-level
			Agents:    make(map[string]AgentConfig),
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading local config: %w", err)
	}

	var cfg LocalConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing local config: %w", err)
	}

	// Set defaults if not specified
	if cfg.AutoLevel == "" {
		cfg.AutoLevel = "medium"
	}
	if cfg.Agents == nil {
		cfg.Agents = make(map[string]AgentConfig)
	}

	return &cfg, nil
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
