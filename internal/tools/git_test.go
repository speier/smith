package tools

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Helper to initialize a git repo
func initGitRepo(t *testing.T, dir string) {
	t.Helper()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Skipf("Skipping test - git not available: %v", err)
	}

	// Configure git
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = dir
	cmd.Run()
}

func TestGetGitStatusTool(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create initial commit
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content"), 0644)
	cmd := exec.Command("git", "add", "file1.txt")
	cmd.Dir = tempDir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = tempDir
	cmd.Run()

	// Create some changes
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("modified"), 0644) // Modified
	os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("new"), 0644)      // Untracked

	tool := NewGetGitStatusTool(tempDir)

	result, err := tool.Execute(context.Background(), map[string]interface{}{})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("expected successful result")
	}

	data := result.Data.(map[string]interface{})

	modified := data["modified"].([]string)
	if len(modified) != 1 || modified[0] != "file1.txt" {
		t.Errorf("expected 1 modified file (file1.txt), got: %v", modified)
	}

	untracked := data["untracked"].([]string)
	if len(untracked) != 1 || untracked[0] != "file2.txt" {
		t.Errorf("expected 1 untracked file (file2.txt), got: %v", untracked)
	}

	clean := data["clean"].(bool)
	if clean {
		t.Error("expected working tree to not be clean")
	}
}

func TestGetGitStatusTool_CleanRepo(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create initial commit
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content"), 0644)
	cmd := exec.Command("git", "add", "file1.txt")
	cmd.Dir = tempDir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = tempDir
	cmd.Run()

	tool := NewGetGitStatusTool(tempDir)

	result, err := tool.Execute(context.Background(), map[string]interface{}{})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	data := result.Data.(map[string]interface{})
	clean := data["clean"].(bool)

	if !clean {
		t.Error("expected working tree to be clean")
	}

	if !strings.Contains(result.Output, "Working tree clean") {
		t.Error("expected output to mention clean working tree")
	}
}

func TestGetGitStatusTool_Metadata(t *testing.T) {
	tool := NewGetGitStatusTool("/tmp")

	if tool.Name() != "get_git_status" {
		t.Errorf("expected name 'get_git_status', got '%s'", tool.Name())
	}

	if tool.RequiresConfirmation(SafetyHigh) {
		t.Error("get_git_status should not require confirmation")
	}
}

func TestGetGitDiffTool(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create initial commit
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("line 1\n"), 0644)
	cmd := exec.Command("git", "add", "file1.txt")
	cmd.Dir = tempDir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = tempDir
	cmd.Run()

	// Modify file
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("line 1\nline 2\n"), 0644)

	tool := NewGetGitDiffTool(tempDir)

	result, err := tool.Execute(context.Background(), map[string]interface{}{})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("expected successful result")
	}

	data := result.Data.(map[string]interface{})
	hasDiff := data["has_diff"].(bool)

	if !hasDiff {
		t.Error("expected diff to be present")
	}

	if !strings.Contains(result.Output, "+line 2") {
		t.Errorf("expected diff output to show added line, got: %s", result.Output)
	}
}

func TestGetGitDiffTool_SpecificFile(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create initial commit
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content 1"), 0644)
	os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("content 2"), 0644)
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = tempDir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = tempDir
	cmd.Run()

	// Modify both files
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("modified 1"), 0644)
	os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("modified 2"), 0644)

	tool := NewGetGitDiffTool(tempDir)

	// Diff specific file
	result, err := tool.Execute(context.Background(), map[string]interface{}{
		"file": "file1.txt",
	})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("expected successful result")
	}

	// Should contain file1 changes but not file2
	if !strings.Contains(result.Output, "file1.txt") {
		t.Error("expected diff to contain file1.txt")
	}
}

func TestGetGitDiffTool_NoDiff(t *testing.T) {
	tempDir := t.TempDir()
	initGitRepo(t, tempDir)

	// Create initial commit
	os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content"), 0644)
	cmd := exec.Command("git", "add", "file1.txt")
	cmd.Dir = tempDir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = tempDir
	cmd.Run()

	tool := NewGetGitDiffTool(tempDir)

	result, err := tool.Execute(context.Background(), map[string]interface{}{})

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	data := result.Data.(map[string]interface{})
	hasDiff := data["has_diff"].(bool)

	if hasDiff {
		t.Error("expected no diff")
	}

	if !strings.Contains(result.Output, "No differences found") {
		t.Error("expected output to say no differences found")
	}
}

func TestGetGitDiffTool_Metadata(t *testing.T) {
	tool := NewGetGitDiffTool("/tmp")

	if tool.Name() != "get_git_diff" {
		t.Errorf("expected name 'get_git_diff', got '%s'", tool.Name())
	}

	if tool.RequiresConfirmation(SafetyHigh) {
		t.Error("get_git_diff should not require confirmation")
	}
}
