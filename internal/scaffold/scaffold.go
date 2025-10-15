package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultConfig represents the default project configuration structure
type DefaultConfig struct {
	Provider  string                    `yaml:"provider"`
	Model     string                    `yaml:"model"`
	AutoLevel string                    `yaml:"autoLevel"`
	Agents    map[string]AgentConfigDef `yaml:"agents,omitempty"`
	Version   int                       `yaml:"version"`
}

// AgentConfigDef represents per-agent configuration in scaffold
type AgentConfigDef struct {
	Model     string `yaml:"model,omitempty"`
	AutoLevel string `yaml:"autoLevel,omitempty"`
	Reasoning string `yaml:"reasoning,omitempty"`
}

// InitProjectFiles creates the default files for a Smith project
func InitProjectFiles(smithDir string) error {
	// Create default kanban.md if it doesn't exist
	kanbanPath := filepath.Join(smithDir, "kanban.md")
	if err := createDefaultKanban(kanbanPath); err != nil {
		return fmt.Errorf("failed to create kanban: %w", err)
	}

	// Create default config.yaml if it doesn't exist
	configPath := filepath.Join(smithDir, "config.yaml")
	if err := createDefaultConfig(configPath); err != nil {
		return fmt.Errorf("failed to create config: %w", err)
	}

	// Create .gitignore for config.yaml
	gitignorePath := filepath.Join(smithDir, ".gitignore")
	if err := createGitignore(gitignorePath); err != nil {
		return fmt.Errorf("failed to create .gitignore: %w", err)
	}

	return nil
}

// createDefaultKanban creates a default kanban board if it doesn't exist
func createDefaultKanban(path string) error {
	if fileExists(path) {
		return nil
	}

	content := `# Agent Kanban Board

## Backlog
<!-- Tasks waiting to be picked up -->

## In Progress
<!-- Tasks currently being worked on -->

## Review
<!-- Tasks pending review -->

## Done
<!-- Completed tasks -->
`

	return os.WriteFile(path, []byte(content), 0644)
}

// createDefaultConfig creates a default config file if it doesn't exist
func createDefaultConfig(path string) error {
	if fileExists(path) {
		return nil
	}

	// Create config struct
	cfg := DefaultConfig{
		Provider:  "",
		Model:     "",
		AutoLevel: "medium",
		Version:   1,
		// Agents map is optional, omitted by default
	}

	// Marshal to YAML with comments
	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshaling default config: %w", err)
	}

	// Add header comment
	header := `# Smith Project Configuration (.smith/config.yaml)
# This file is gitignored - each developer/project has their own settings

# LLM Provider (required)
# Run '/settings' in the TUI to configure provider and model

`

	footer := `
# Per-agent overrides (optional)
# Uncomment and configure after selecting a provider and model:
#
# agents:
#   planning:
#     model: ""
#     autoLevel: high
#     reasoning: high
#   implementation:
#     model: ""
#     autoLevel: medium
#     reasoning: medium
#   testing:
#     model: ""
#     autoLevel: low
#     reasoning: low
#   review:
#     model: ""
#     autoLevel: high
#     reasoning: high
`

	fullContent := header + string(data) + footer

	return os.WriteFile(path, []byte(fullContent), 0644)
}

// GitignoreContent represents the .gitignore structure
type GitignoreContent struct {
	DBFiles    []string `yaml:"-"` // Not in YAML, just for documentation
	ConfigFile string   `yaml:"-"`
}

// createGitignore creates a .gitignore file for the .smith directory
func createGitignore(path string) error {
	if fileExists(path) {
		return nil
	}

	// Define what should be ignored
	ignoreItems := []string{
		"# Gitignore for .smith directory",
		"",
		"# Local agent runtime state (not committed)",
		"smith.db",
		"smith.db-shm",
		"smith.db-wal",
		"",
		"# User-specific config with API keys",
		"config.yaml",
	}

	content := ""
	for _, line := range ignoreItems {
		content += line + "\n"
	}

	return os.WriteFile(path, []byte(content), 0644)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
