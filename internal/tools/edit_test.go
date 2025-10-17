package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReplaceInFileTool(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file with content
	testFile := filepath.Join(tempDir, "test.go")
	initialContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name         string
		params       map[string]interface{}
		shouldError  bool
		expectInFile string
		errorMsg     string
	}{
		{
			name: "simple_replacement",
			params: map[string]interface{}{
				"path":     "test.go",
				"old_text": `fmt.Println("Hello, World!")`,
				"new_text": `fmt.Println("Hello, Smith!")`,
			},
			expectInFile: `fmt.Println("Hello, Smith!")`,
		},
		{
			name: "multiline_replacement",
			params: map[string]interface{}{
				"path": "test.go",
				"old_text": `func main() {
	fmt.Println("Hello, Smith!")
}`,
				"new_text": `func main() {
	fmt.Println("Greetings, Agent Smith!")
	fmt.Println("Welcome to the Matrix.")
}`,
			},
			expectInFile: `fmt.Println("Greetings, Agent Smith!")`,
		},
		{
			name: "delete_text",
			params: map[string]interface{}{
				"path":     "test.go",
				"old_text": `	fmt.Println("Welcome to the Matrix.")` + "\n",
				"new_text": "",
			},
			expectInFile: `fmt.Println("Greetings, Agent Smith!")`,
		},
		{
			name: "text_not_found",
			params: map[string]interface{}{
				"path":     "test.go",
				"old_text": "NonexistentText",
				"new_text": "Something",
			},
			shouldError: true,
			errorMsg:    "not found",
		},
		{
			name: "ambiguous_replacement",
			params: map[string]interface{}{
				"path":     "test.go",
				"old_text": "fmt",
				"new_text": "FMT",
			},
			shouldError: true,
			errorMsg:    "ambiguous",
		},
		{
			name: "missing_path",
			params: map[string]interface{}{
				"old_text": "foo",
				"new_text": "bar",
			},
			shouldError: true,
		},
		{
			name: "missing_old_text",
			params: map[string]interface{}{
				"path":     "test.go",
				"new_text": "bar",
			},
			shouldError: true,
		},
		{
			name: "missing_new_text",
			params: map[string]interface{}{
				"path":     "test.go",
				"old_text": "foo",
			},
			shouldError: true,
		},
		{
			name: "path_traversal",
			params: map[string]interface{}{
				"path":     "../../../etc/passwd",
				"old_text": "root",
				"new_text": "hacked",
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewReplaceInFileTool(tempDir)

			// Validate
			err := tool.Validate(tt.params)
			if tt.shouldError && err == nil && (tt.name == "missing_path" || tt.name == "missing_old_text" || tt.name == "missing_new_text" || tt.name == "path_traversal") {
				t.Error("expected validation error, got nil")
				return
			}

			if tt.shouldError && (tt.name == "missing_path" || tt.name == "missing_old_text" || tt.name == "missing_new_text" || tt.name == "path_traversal") {
				return // Skip execution for validation failures
			}

			// Execute
			result, err := tool.Execute(context.Background(), tt.params)

			if tt.shouldError {
				if err == nil {
					t.Error("expected execution error, got nil")
				}
				if result == nil || result.Success {
					t.Error("expected unsuccessful result")
				}
				if tt.errorMsg != "" && !contains(result.Error, tt.errorMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorMsg, result.Error)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil || !result.Success {
					t.Errorf("expected successful result, got: %+v", result)
				}

				// Verify file content
				if tt.expectInFile != "" {
					content, err := os.ReadFile(testFile)
					if err != nil {
						t.Errorf("failed to read file: %v", err)
					}
					if !contains(string(content), tt.expectInFile) {
						t.Errorf("expected file to contain '%s', got:\n%s", tt.expectInFile, string(content))
					}
				}
			}
		})
	}
}

func TestReplaceInFileTool_Metadata(t *testing.T) {
	tool := NewReplaceInFileTool("/tmp")

	if tool.Name() != "replace_in_file" {
		t.Errorf("expected name 'replace_in_file', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}

	if !tool.RequiresConfirmation(SafetyMedium) {
		t.Error("replace_in_file should require confirmation at medium safety")
	}

	if !tool.RequiresConfirmation(SafetyHigh) {
		t.Error("replace_in_file should require confirmation at high safety")
	}

	if tool.RequiresConfirmation(SafetyLow) {
		t.Error("replace_in_file should not require confirmation at low safety")
	}
}

func TestReplaceInFileTool_AtomicOperation(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "atomic.txt")

	content := "Line 1\nLine 2\nLine 3\n"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool := NewReplaceInFileTool(tempDir)

	// This should fail and leave file unchanged
	params := map[string]interface{}{
		"path":     "atomic.txt",
		"old_text": "NonExistent",
		"new_text": "Changed",
	}

	result, err := tool.Execute(context.Background(), params)

	if err == nil {
		t.Error("expected error for nonexistent text")
	}

	if result.Success {
		t.Error("expected unsuccessful result")
	}

	// Verify file is unchanged
	newContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(newContent) != content {
		t.Errorf("file should be unchanged after failed replacement")
	}
}

func TestReplaceInFileTool_NonexistentFile(t *testing.T) {
	tempDir := t.TempDir()
	tool := NewReplaceInFileTool(tempDir)

	params := map[string]interface{}{
		"path":     "nonexistent.txt",
		"old_text": "foo",
		"new_text": "bar",
	}

	result, err := tool.Execute(context.Background(), params)

	if err == nil {
		t.Error("expected error for nonexistent file")
	}

	if result == nil || result.Success {
		t.Error("expected unsuccessful result")
	}
}

func TestReplaceInFileTool_EmptyNewText(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "delete.txt")

	content := "Keep this\nDelete this line\nKeep this too"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool := NewReplaceInFileTool(tempDir)

	params := map[string]interface{}{
		"path":     "delete.txt",
		"old_text": "Delete this line\n",
		"new_text": "",
	}

	result, err := tool.Execute(context.Background(), params)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Error("expected successful result")
	}

	// Verify deletion
	newContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	expected := "Keep this\nKeep this too"
	if string(newContent) != expected {
		t.Errorf("expected '%s', got '%s'", expected, string(newContent))
	}
}

func TestReplaceAllInFileTool(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file with multiple occurrences
	testFile := filepath.Join(tempDir, "refactor.go")
	initialContent := `package main

func oldName() {
	oldName := "value"
	oldName += " more"
	return oldName
}

func otherFunc() {
	x := oldName()
	return x
}
`
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name            string
		params          map[string]interface{}
		shouldError     bool
		expectCount     int
		expectInFile    string
		notExpectInFile string
		errorMsg        string
	}{
		{
			name: "replace_all_occurrences",
			params: map[string]interface{}{
				"path":     "refactor.go",
				"old_text": "oldName",
				"new_text": "newName",
			},
			expectCount:     5,
			expectInFile:    "newName",
			notExpectInFile: "oldName",
		},
		{
			name: "replace_with_limit",
			params: map[string]interface{}{
				"path":             "refactor.go",
				"old_text":         "newName",
				"new_text":         "betterName",
				"max_replacements": 10.0,
			},
			expectCount:  5,
			expectInFile: "betterName",
		},
		{
			name: "exceed_limit",
			params: map[string]interface{}{
				"path":             "refactor.go",
				"old_text":         " ",
				"new_text":         "_",
				"max_replacements": 5.0,
			},
			shouldError: true,
			errorMsg:    "too many occurrences",
		},
		{
			name: "text_not_found",
			params: map[string]interface{}{
				"path":     "refactor.go",
				"old_text": "nonexistent",
				"new_text": "something",
			},
			shouldError: true,
			errorMsg:    "not found",
		},
		{
			name: "delete_all",
			params: map[string]interface{}{
				"path":     "refactor.go",
				"old_text": "\t",
				"new_text": "",
			},
			expectCount:     0, // Tabs will be removed
			notExpectInFile: "\t",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewReplaceAllInFileTool(tempDir)

			result, err := tool.Execute(context.Background(), tt.params)

			if tt.shouldError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if result == nil || result.Success {
					t.Error("expected unsuccessful result")
				}
				if tt.errorMsg != "" && !contains(result.Error, tt.errorMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorMsg, result.Error)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil || !result.Success {
					t.Errorf("expected successful result, got: %+v", result)
				}

				// Check replacement count
				if tt.expectCount > 0 {
					if data, ok := result.Data.(map[string]interface{}); ok {
						if count, ok := data["replacements"].(int); ok {
							if count != tt.expectCount {
								t.Errorf("expected %d replacements, got %d", tt.expectCount, count)
							}
						}
					}
				}

				// Verify file content
				content, err := os.ReadFile(testFile)
				if err != nil {
					t.Errorf("failed to read file: %v", err)
				}

				if tt.expectInFile != "" && !contains(string(content), tt.expectInFile) {
					t.Errorf("expected file to contain '%s'", tt.expectInFile)
				}

				if tt.notExpectInFile != "" && contains(string(content), tt.notExpectInFile) {
					t.Errorf("expected file NOT to contain '%s'", tt.notExpectInFile)
				}
			}
		})
	}
}

func TestReplaceAllInFileTool_Metadata(t *testing.T) {
	tool := NewReplaceAllInFileTool("/tmp")

	if tool.Name() != "replace_all_in_file" {
		t.Errorf("expected name 'replace_all_in_file', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}

	if !tool.RequiresConfirmation(SafetyMedium) {
		t.Error("replace_all_in_file should require confirmation at medium safety")
	}
}

func TestDiffFilesTool(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	file1Path := filepath.Join(tempDir, "file1.txt")
	file2Path := filepath.Join(tempDir, "file2.txt")
	file3Path := filepath.Join(tempDir, "file3.txt")

	_ = os.WriteFile(file1Path, []byte("line 1\nline 2\nline 3\n"), 0644)
	_ = os.WriteFile(file2Path, []byte("line 1\nmodified line 2\nline 3\n"), 0644)
	_ = os.WriteFile(file3Path, []byte("line 1\nline 2\nline 3\n"), 0644)

	tests := []struct {
		name        string
		file1       string
		file2       string
		shouldError bool
		expectDiff  bool
	}{
		{
			name:       "files_different",
			file1:      "file1.txt",
			file2:      "file2.txt",
			expectDiff: true,
		},
		{
			name:       "files_identical",
			file1:      "file1.txt",
			file2:      "file3.txt",
			expectDiff: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewDiffFilesTool(tempDir)

			params := map[string]interface{}{
				"file1": tt.file1,
				"file2": tt.file2,
			}

			result, err := tool.Execute(context.Background(), params)

			if tt.shouldError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !result.Success {
					t.Error("expected successful result")
				}

				data := result.Data.(map[string]interface{})
				identical := data["identical"].(bool)

				if tt.expectDiff && identical {
					t.Error("expected files to be different")
				}
				if !tt.expectDiff && !identical {
					t.Error("expected files to be identical")
				}
			}
		})
	}
}

func TestDiffFilesTool_Metadata(t *testing.T) {
	tool := NewDiffFilesTool("/tmp")

	if tool.Name() != "diff_files" {
		t.Errorf("expected name 'diff_files', got '%s'", tool.Name())
	}

	if tool.RequiresConfirmation(SafetyHigh) {
		t.Error("diff_files should not require confirmation")
	}
}

func TestBatchSearchReplaceTool(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	_ = os.WriteFile(filepath.Join(tempDir, "test1.go"), []byte("package main\nfunc oldName() {}\n"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "test2.go"), []byte("package test\nfunc oldName() {}\n"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "readme.md"), []byte("# oldName"), 0644)

	tests := []struct {
		name           string
		oldText        string
		newText        string
		filePattern    string
		expectModified int
		expectContains map[string]string
	}{
		{
			name:           "replace_in_go_files",
			oldText:        "oldName",
			newText:        "newName",
			filePattern:    "*.go",
			expectModified: 2,
			expectContains: map[string]string{
				"test1.go": "func newName()",
				"test2.go": "func newName()",
			},
		},
		{
			name:           "replace_in_all_files",
			oldText:        "oldName",
			newText:        "newName",
			filePattern:    "*",
			expectModified: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset files for each test
			_ = os.WriteFile(filepath.Join(tempDir, "test1.go"), []byte("package main\nfunc oldName() {}\n"), 0644)
			_ = os.WriteFile(filepath.Join(tempDir, "test2.go"), []byte("package test\nfunc oldName() {}\n"), 0644)
			_ = os.WriteFile(filepath.Join(tempDir, "readme.md"), []byte("# oldName"), 0644)

			tool := NewBatchSearchReplaceTool(tempDir)

			params := map[string]interface{}{
				"old_text":     tt.oldText,
				"new_text":     tt.newText,
				"file_pattern": tt.filePattern,
			}

			result, err := tool.Execute(context.Background(), params)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !result.Success {
				t.Error("expected successful result")
			}

			data := result.Data.(map[string]interface{})
			filesModified := data["files_modified"].(int)

			if filesModified != tt.expectModified {
				t.Errorf("expected %d files modified, got %d", tt.expectModified, filesModified)
			}

			// Verify content changes
			for file, expectedContent := range tt.expectContains {
				content, err := os.ReadFile(filepath.Join(tempDir, file))
				if err != nil {
					t.Errorf("failed to read %s: %v", file, err)
				}
				if !strings.Contains(string(content), expectedContent) {
					t.Errorf("expected %s to contain '%s'", file, expectedContent)
				}
			}
		})
	}
}

func TestBatchSearchReplaceTool_Metadata(t *testing.T) {
	tool := NewBatchSearchReplaceTool("/tmp")

	if tool.Name() != "batch_search_replace" {
		t.Errorf("expected name 'batch_search_replace', got '%s'", tool.Name())
	}

	if !tool.RequiresConfirmation(SafetyMedium) {
		t.Error("batch_search_replace should require confirmation at medium safety")
	}
}
