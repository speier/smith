package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

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

## WIP
<!-- Work in progress - tasks currently being worked on -->

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

	content := `# Smith Configuration
# This file is gitignored - each developer has their own settings

# LLM Provider settings
llm:
  provider: copilot  # copilot, openrouter
  model: gpt-4o      # Model to use
  
  # OpenRouter settings (if using openrouter)
  # openrouter:
  #   api_key: your-api-key-here
  #   model: anthropic/claude-3.5-sonnet

# Agent settings
agents:
  max_concurrent: 10  # Maximum number of concurrent agents
  heartbeat_interval: 10s  # How often agents send heartbeats
  timeout: 5m  # Agent timeout (considered dead if no heartbeat)

# Safety settings
safety:
  level: auto  # auto, safe, yolo
  require_approval: false  # Require human approval for certain actions
`

	return os.WriteFile(path, []byte(content), 0644)
}

// createGitignore creates a .gitignore file for the .smith directory
func createGitignore(path string) error {
	if fileExists(path) {
		return nil
	}

	content := `# Gitignore for .smith directory

# Local agent runtime state (not committed)
smith.db
smith.db-shm
smith.db-wal

# User-specific config with API keys
config.yaml
`

	return os.WriteFile(path, []byte(content), 0644)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
