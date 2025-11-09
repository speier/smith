package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGitignoreContent(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "smith-gitignore-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create .smith directory
	smithDir := filepath.Join(tmpDir, ".smith")
	if err := os.MkdirAll(smithDir, 0755); err != nil {
		t.Fatalf("failed to create .smith dir: %v", err)
	}

	// Initialize project files
	if err := InitProjectFiles(smithDir); err != nil {
		t.Fatalf("InitProjectFiles failed: %v", err)
	}

	// Read .gitignore
	gitignorePath := filepath.Join(smithDir, ".gitignore")
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	contentStr := string(content)

	// Verify all required entries are present
	required := []string{
		"smith.db",
		"config.yaml",
	}

	for _, entry := range required {
		if !strings.Contains(contentStr, entry) {
			t.Errorf(".gitignore missing entry: %s", entry)
		}
	}

	t.Logf(".gitignore content:\n%s", contentStr)
}

func TestInitProjectFilesIdempotent(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "smith-files-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	smithDir := filepath.Join(tmpDir, ".smith")
	if err := os.MkdirAll(smithDir, 0755); err != nil {
		t.Fatalf("failed to create .smith dir: %v", err)
	}

	// Run twice - should be idempotent
	if err := InitProjectFiles(smithDir); err != nil {
		t.Fatalf("first InitProjectFiles failed: %v", err)
	}

	if err := InitProjectFiles(smithDir); err != nil {
		t.Fatalf("second InitProjectFiles failed: %v", err)
	}

	// Verify all files were created
	files := []string{"config.yaml", ".gitignore"}
	for _, file := range files {
		path := filepath.Join(smithDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("file %s was not created", file)
		}
	}
}
