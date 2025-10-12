package prompts

import (
	_ "embed"
	"fmt"

	"gopkg.in/yaml.v3"
)

//go:embed prompts.yaml
var promptsYAML []byte

// AgentPrompts represents the complete agent prompts configuration
type AgentPrompts struct {
	Version string                 `yaml:"version"`
	Agents  map[string]AgentPrompt `yaml:"agents"`
}

// AgentPrompt represents the prompt configuration for a single agent
type AgentPrompt struct {
	Name         string `yaml:"name"`
	Description  string `yaml:"description"`
	SystemPrompt string `yaml:"systemPrompt"`
}

// LoadPrompts loads the embedded agent prompts
func LoadPrompts() (*AgentPrompts, error) {
	var prompts AgentPrompts
	if err := yaml.Unmarshal(promptsYAML, &prompts); err != nil {
		return nil, fmt.Errorf("failed to parse agent prompts: %w", err)
	}
	return &prompts, nil
}

// GetPrompt returns the prompt for a specific agent role
func (ap *AgentPrompts) GetPrompt(role string) (*AgentPrompt, error) {
	prompt, exists := ap.Agents[role]
	if !exists {
		return nil, fmt.Errorf("no prompt found for agent role: %s", role)
	}
	return &prompt, nil
}

// GetSystemPrompt returns just the system prompt for an agent role
func (ap *AgentPrompts) GetSystemPrompt(role string) (string, error) {
	prompt, err := ap.GetPrompt(role)
	if err != nil {
		return "", err
	}
	return prompt.SystemPrompt, nil
}
