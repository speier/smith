package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GetGitStatusTool gets the git working tree status
type GetGitStatusTool struct {
	workDir string
}

// NewGetGitStatusTool creates a new GetGitStatusTool
func NewGetGitStatusTool(workDir string) *GetGitStatusTool {
	return &GetGitStatusTool{workDir: workDir}
}

func (t *GetGitStatusTool) Name() string {
	return "get_git_status"
}

func (t *GetGitStatusTool) Description() string {
	return "Get git working tree status (modified, added, deleted, untracked files and current branch)"
}

func (t *GetGitStatusTool) Validate(params map[string]interface{}) error {
	// No parameters required
	return nil
}

func (t *GetGitStatusTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	// Get current branch
	branchCmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	branchCmd.Dir = t.workDir
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get current branch (not a git repo?): %v", err),
		}, err
	}
	branch := strings.TrimSpace(string(branchOutput))

	// Get status with porcelain format for easier parsing
	statusCmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	statusCmd.Dir = t.workDir
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get git status: %v", err),
		}, err
	}

	// Parse porcelain output
	var modified, added, deleted, untracked []string
	lines := strings.Split(string(statusOutput), "\n")
	
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		
		status := line[0:2]
		file := strings.TrimSpace(line[3:])
		
		switch {
		case status == "??" :
			untracked = append(untracked, file)
		case status[0] == 'M' || status[1] == 'M':
			modified = append(modified, file)
		case status[0] == 'A' || status[1] == 'A':
			added = append(added, file)
		case status[0] == 'D' || status[1] == 'D':
			deleted = append(deleted, file)
		}
	}

	// Build output
	var output strings.Builder
	output.WriteString(fmt.Sprintf("Branch: %s\n", branch))
	
	if len(modified)+len(added)+len(deleted)+len(untracked) == 0 {
		output.WriteString("Working tree clean\n")
	} else {
		if len(modified) > 0 {
			output.WriteString(fmt.Sprintf("\nModified (%d):\n", len(modified)))
			for _, f := range modified {
				output.WriteString(fmt.Sprintf("  M %s\n", f))
			}
		}
		if len(added) > 0 {
			output.WriteString(fmt.Sprintf("\nAdded (%d):\n", len(added)))
			for _, f := range added {
				output.WriteString(fmt.Sprintf("  A %s\n", f))
			}
		}
		if len(deleted) > 0 {
			output.WriteString(fmt.Sprintf("\nDeleted (%d):\n", len(deleted)))
			for _, f := range deleted {
				output.WriteString(fmt.Sprintf("  D %s\n", f))
			}
		}
		if len(untracked) > 0 {
			output.WriteString(fmt.Sprintf("\nUntracked (%d):\n", len(untracked)))
			for _, f := range untracked {
				output.WriteString(fmt.Sprintf("  ? %s\n", f))
			}
		}
	}

	return &ToolResult{
		Success: true,
		Output:  output.String(),
		Data: map[string]interface{}{
			"branch":     branch,
			"modified":   modified,
			"added":      added,
			"deleted":    deleted,
			"untracked":  untracked,
			"clean":      len(modified)+len(added)+len(deleted)+len(untracked) == 0,
		},
	}, nil
}

func (t *GetGitStatusTool) RequiresConfirmation(level SafetyLevel) bool {
	return false // Reading only, no confirmation needed
}

// GetGitDiffTool gets the git diff for files
type GetGitDiffTool struct {
	workDir string
}

// NewGetGitDiffTool creates a new GetGitDiffTool
func NewGetGitDiffTool(workDir string) *GetGitDiffTool {
	return &GetGitDiffTool{workDir: workDir}
}

func (t *GetGitDiffTool) Name() string {
	return "get_git_diff"
}

func (t *GetGitDiffTool) Description() string {
	return "Get git diff for changed files (optionally for a specific file)"
}

func (t *GetGitDiffTool) Validate(params map[string]interface{}) error {
	// file parameter is optional
	if file, ok := params["file"].(string); ok && file != "" {
		if err := validatePath(t.workDir, file); err != nil {
			return err
		}
	}
	
	return nil
}

func (t *GetGitDiffTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	args := []string{"diff"}
	
	// Add file if specified
	if file, ok := params["file"].(string); ok && file != "" {
		args = append(args, file)
	}
	
	// Add --cached flag if requested
	if staged, ok := params["staged"].(bool); ok && staged {
		args = []string{"diff", "--cached"}
		if file, ok := params["file"].(string); ok && file != "" {
			args = append(args, file)
		}
	}

	diffCmd := exec.CommandContext(ctx, "git", args...)
	diffCmd.Dir = t.workDir
	diffOutput, err := diffCmd.Output()
	if err != nil {
		// git diff returns non-zero if there are differences, but that's not an error
		if exitErr, ok := err.(*exec.ExitError); ok {
			diffOutput = exitErr.Stderr
		} else {
			return &ToolResult{
				Success: false,
				Error:   fmt.Sprintf("failed to get git diff: %v", err),
			}, err
		}
	}

	output := string(diffOutput)
	if output == "" {
		output = "No differences found"
	}

	return &ToolResult{
		Success: true,
		Output:  output,
		Data: map[string]interface{}{
			"diff":      output,
			"has_diff":  output != "No differences found",
		},
	}, nil
}

func (t *GetGitDiffTool) RequiresConfirmation(level SafetyLevel) bool {
	return false // Reading only, no confirmation needed
}
