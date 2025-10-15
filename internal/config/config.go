package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config Design Philosophy:
// - Global config (~/.smith/config.yaml): User's default preferences across all projects
// - Local config (.smith/config.yaml): Per-project overrides
// - Local overrides global (simple merge)
// - First run creates global with user's choices
// - Local only created when user changes something for that project
// - No local = uses global (sensible default)
//
// Flow:
// 1. First time: Prompt for provider/model, save to global AND local
// 2. New project: Uses global automatically
// 3. Change in project: Save to local, global unchanged
// 4. Switch projects: Each uses its own local or falls back to global

// AgentConfig represents configuration for a specific agent
type AgentConfig struct {
	Model     string `yaml:"model,omitempty"`
	AutoLevel string `yaml:"autoLevel,omitempty"`
	Reasoning string `yaml:"reasoning,omitempty"` // low/medium/high
}

// Config represents Smith configuration (both global and local use same structure)
// Global: ~/.smith/config.yaml (user defaults)
// Local: .smith/config.yaml (project overrides)
type Config struct {
	// Provider: github-copilot, openrouter, openai
	// Available providers determined by: auth status + env vars
	Provider string `yaml:"provider,omitempty"`

	// Model: Default model when not specified
	Model string `yaml:"model,omitempty"`

	// Safety level: off, low, medium, high
	SafetyLevel string `yaml:"safety_level,omitempty"`

	// Optional: Per-agent model overrides (advanced usage)
	AgentModels map[string]string `yaml:"agent_models,omitempty"` // planning: claude-3.5-sonnet, etc.

	// Config version
	Version int `yaml:"version,omitempty"`
}

// MergedConfig represents the result of merging global + local configs
type MergedConfig struct {
	Config
	IsLocal bool // True if local config exists and was merged
}

// LoadWithMerge loads both global and local config, merging them appropriately
// Local config overrides global config for any non-empty fields
func LoadWithMerge(projectPath string) (*MergedConfig, error) {
	// Load global config first
	globalCfg, err := loadGlobal()
	if err != nil {
		return nil, fmt.Errorf("loading global config: %w", err)
	}

	// Try to load local config
	localCfg, err := loadLocal(projectPath)
	if err != nil {
		return nil, fmt.Errorf("loading local config: %w", err)
	}

	// Start with global as base
	merged := &MergedConfig{
		Config:  *globalCfg,
		IsLocal: localCfg != nil,
	}

	// Merge local over global (local wins)
	if localCfg != nil {
		if localCfg.Provider != "" {
			merged.Provider = localCfg.Provider
		}
		if localCfg.Model != "" {
			merged.Model = localCfg.Model
		}
		if localCfg.SafetyLevel != "" {
			merged.SafetyLevel = localCfg.SafetyLevel
		}
		if len(localCfg.AgentModels) > 0 {
			if merged.AgentModels == nil {
				merged.AgentModels = make(map[string]string)
			}
			for k, v := range localCfg.AgentModels {
				merged.AgentModels[k] = v
			}
		}
		if localCfg.Version > 0 {
			merged.Version = localCfg.Version
		}
	}

	// Set defaults if still empty
	if merged.SafetyLevel == "" {
		merged.SafetyLevel = "medium"
	}
	if merged.Version == 0 {
		merged.Version = 1
	}

	return merged, nil
}

// loadGlobal loads global config from ~/.smith/config.yaml
func loadGlobal() (*Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, configFileName)

	// Return empty config if file doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			SafetyLevel: "medium",
			Version:     1,
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading global config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing global config: %w", err)
	}

	return &cfg, nil
}

// loadLocal loads local config from .smith/config.yaml in project
func loadLocal(projectPath string) (*Config, error) {
	if projectPath == "" {
		return nil, nil
	}

	configPath := filepath.Join(projectPath, configDirName, configFileName)

	// Return nil if file doesn't exist (not an error)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading local config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing local config: %w", err)
	}

	return &cfg, nil
}

// SaveGlobal saves config to global ~/.smith/config.yaml
func (c *Config) SaveGlobal() error {
	configDir, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, configFileName)

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("writing global config: %w", err)
	}

	return nil
}

// SaveLocal saves config to local .smith/config.yaml in project
func (c *Config) SaveLocal(projectPath string) error {
	if projectPath == "" {
		return fmt.Errorf("project path required for local config")
	}

	configDir := filepath.Join(projectPath, configDirName)
	configPath := filepath.Join(configDir, configFileName)

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("writing local config: %w", err)
	}

	return nil
}

// GetAgentModel returns the model for a specific agent, falling back to default
func (c *Config) GetAgentModel(agentName string) string {
	if c.AgentModels != nil {
		if model, ok := c.AgentModels[agentName]; ok && model != "" {
			return model
		}
	}
	return c.Model
}

// Auth represents stored authentication data
type Auth struct {
	Provider string                 `yaml:"provider"`
	Data     map[string]interface{} `yaml:"data"`
}

const (
	configDirName  = ".smith"
	configFileName = "config.yaml"
	authFileName   = "auth.yaml"
)

// GetConfigDir returns the path to Smith config directory
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, configDirName), nil
}

// EnsureConfigDir creates the config directory if it doesn't exist
func EnsureConfigDir() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("creating config directory: %w", err)
	}

	return configDir, nil
}

// Load loads the configuration from disk, returns default if not found
// Deprecated: Use LoadWithMerge(projectPath) instead for proper global/local merging
func Load() (*Config, error) {
	return loadGlobal()
}

// Save saves the configuration to disk
// Deprecated: Use SaveGlobal() or SaveLocal(projectPath) instead
func (c *Config) Save() error {
	return c.SaveGlobal()
}

// LoadAuth loads authentication data for a provider
func LoadAuth(provider string) (*Auth, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	authPath := filepath.Join(configDir, authFileName)

	// Return nil if file doesn't exist
	if _, err := os.Stat(authPath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(authPath)
	if err != nil {
		return nil, fmt.Errorf("reading auth: %w", err)
	}

	var auth Auth
	if err := yaml.Unmarshal(data, &auth); err != nil {
		return nil, fmt.Errorf("parsing auth: %w", err)
	}

	if auth.Provider != provider {
		return nil, nil // Auth exists but for different provider
	}

	return &auth, nil
}

// SaveAuth saves authentication data for a provider
func SaveAuth(provider string, data map[string]interface{}) error {
	configDir, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	authPath := filepath.Join(configDir, authFileName)

	auth := Auth{
		Provider: provider,
		Data:     data,
	}

	yamlData, err := yaml.Marshal(auth)
	if err != nil {
		return fmt.Errorf("marshaling auth: %w", err)
	}

	if err := os.WriteFile(authPath, yamlData, 0600); err != nil {
		return fmt.Errorf("writing auth: %w", err)
	}

	return nil
}

// ClearAuth removes authentication data
func ClearAuth() error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	authPath := filepath.Join(configDir, authFileName)

	if err := os.Remove(authPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing auth: %w", err)
	}

	return nil
}
