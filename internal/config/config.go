package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the Smith configuration
type Config struct {
	Provider     string            `json:"provider"`     // "copilot", "ollama", "openai"
	Model        string            `json:"model"`        // Model identifier
	ProviderOpts map[string]string `json:"providerOpts"` // Provider-specific options
}

// Auth represents stored authentication data
type Auth struct {
	Provider string                 `json:"provider"`
	Data     map[string]interface{} `json:"data"`
}

const (
	configDirName  = ".smith"
	configFileName = "config.json"
	authFileName   = "auth.json"
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
func Load() (*Config, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, configFileName)

	// Return default config if file doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{
			Provider:     "copilot",
			Model:        "gpt-4o",
			ProviderOpts: make(map[string]string),
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	return &cfg, nil
}

// Save saves the configuration to disk
func (c *Config) Save() error {
	configDir, err := EnsureConfigDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(configDir, configFileName)

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
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
	if err := json.Unmarshal(data, &auth); err != nil {
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

	jsonData, err := json.MarshalIndent(auth, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling auth: %w", err)
	}

	if err := os.WriteFile(authPath, jsonData, 0600); err != nil {
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
