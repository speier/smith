package agent

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/speier/smith/internal/coordinator"
)

type Config struct {
	Role        string
	TaskID      string
	ProjectPath string
	APIKey      string // For future LLM integration
}

type Agent struct {
	config Config
	coord  *coordinator.Coordinator
}

func New(config Config) *Agent {
	return &Agent{
		config: config,
		coord:  coordinator.New(config.ProjectPath),
	}
}

func (a *Agent) Execute() error {
	fmt.Printf("📚 Loading role instructions for %s agent...\n", a.config.Role)

	// 1. Load role-specific instructions from AGENTS.md
	rolePrompt, err := a.loadRolePrompt()
	if err != nil {
		return fmt.Errorf("failed to load role: %w", err)
	}

	// 2. Load task details from TODO.md
	taskDetails, err := a.loadTaskDetails()
	if err != nil {
		return fmt.Errorf("failed to load task: %w", err)
	}

	// 3. Claim task and lock files (update COMMS.md)
	if err := a.coord.ClaimTask(a.config.TaskID, a.config.Role); err != nil {
		fmt.Printf("⚠️  Could not claim task: %v\n", err)
	}

	// 4. Execute LLM agent loop with tools
	fmt.Println("🧠 Starting agent reasoning loop...")

	result, err := a.runAgentLoop(rolePrompt, taskDetails)
	if err != nil {
		return fmt.Errorf("agent execution failed: %w", err)
	}

	// 5. Update TODO.md with completion status
	fmt.Printf("📝 Task result: %s\n", result)

	return nil
}

func (a *Agent) loadRolePrompt() (string, error) {
	path := filepath.Join(a.config.ProjectPath, "AGENTS.md")
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// TODO: Parse AGENTS.md and extract section for this role
	// For now, return full file
	return string(content), nil
}

func (a *Agent) loadTaskDetails() (string, error) {
	path := filepath.Join(a.config.ProjectPath, "TODO.md")
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	// TODO: Parse TODO.md and extract specific task
	return string(content), nil
}

func (a *Agent) runAgentLoop(systemPrompt, task string) (string, error) {
	// TODO: Implement LLM integration when provider is chosen
	// This will call the LLM API with:
	// - systemPrompt (role instructions from AGENTS.md)
	// - task (task details)
	// - tools (read_file, write_file, run_command)
	//
	// For now, return a mock response

	fmt.Println("🤖 Agent would process task here with LLM")
	fmt.Printf("   Role: %s\n", a.config.Role)
	fmt.Printf("   Task: %s\n", a.config.TaskID)

	return "Task completed (LLM integration pending)", nil
}

// Tool execution helpers (for future LLM integration)
func (a *Agent) executeTool(name string, input interface{}) string {
	inputMap, ok := input.(map[string]interface{})
	if !ok {
		return "Error: invalid tool input"
	}

	switch name {
	case "read_file":
		path := filepath.Join(a.config.ProjectPath, inputMap["path"].(string))
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Sprintf("Error reading file: %v", err)
		}
		return string(content)

	case "write_file":
		path := filepath.Join(a.config.ProjectPath, inputMap["path"].(string))
		content := inputMap["content"].(string)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Sprintf("Error writing file: %v", err)
		}
		return "File written successfully"

	case "run_command":
		cmd := exec.Command("sh", "-c", inputMap["command"].(string))
		cmd.Dir = a.config.ProjectPath
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Sprintf("Error: %v\nOutput: %s", err, output)
		}
		return string(output)

	default:
		return fmt.Sprintf("Unknown tool: %s", name)
	}
}
