package tools

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestReadFileTool(t *testing.T) {
	// Create temp directory
	tempDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		params      map[string]interface{}
		shouldError bool
		expectMsg   string
		skipExec    bool
	}{
		{
			name:      "read_existing_file",
			params:    map[string]interface{}{"path": "test.txt"},
			expectMsg: testContent,
		},
		{
			name:        "read_nonexistent_file",
			params:      map[string]interface{}{"path": "nonexistent.txt"},
			shouldError: true,
			skipExec:    false, // Validation passes but execution fails
		},
		{
			name:        "missing_path_param",
			params:      map[string]interface{}{},
			shouldError: true,
		},
		{
			name:        "path_traversal_attempt",
			params:      map[string]interface{}{"path": "../../../etc/passwd"},
			shouldError: true,
			skipExec:    true, // Validation should fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewReadFileTool(tempDir)

			// Validate
			err := tool.Validate(tt.params)
			if tt.skipExec && err == nil {
				t.Error("expected validation error, got nil")
				return
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
				return
			}

			if tt.skipExec {
				return // Skip execution if validation should fail
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
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil || !result.Success {
					t.Error("expected successful result")
				}
				if tt.expectMsg != "" && result.Output != tt.expectMsg {
					t.Errorf("expected output '%s', got '%s'", tt.expectMsg, result.Output)
				}
			}
		})
	}
}

func TestReadFileTool_Metadata(t *testing.T) {
	tool := NewReadFileTool("/tmp")

	if tool.Name() != "read_file" {
		t.Errorf("expected name 'read_file', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}

	if tool.RequiresConfirmation(SafetyHigh) {
		t.Error("read_file should not require confirmation at any safety level")
	}
}

func TestWriteFileTool(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		params      map[string]interface{}
		shouldError bool
		verifyFile  string
		expectData  string
	}{
		{
			name: "write_new_file",
			params: map[string]interface{}{
				"path":    "output.txt",
				"content": "Test content",
			},
			verifyFile: "output.txt",
			expectData: "Test content",
		},
		{
			name: "write_nested_file",
			params: map[string]interface{}{
				"path":    "subdir/nested.txt",
				"content": "Nested content",
			},
			verifyFile: "subdir/nested.txt",
			expectData: "Nested content",
		},
		{
			name: "overwrite_existing",
			params: map[string]interface{}{
				"path":    "existing.txt",
				"content": "New content",
			},
			verifyFile: "existing.txt",
			expectData: "New content",
		},
		{
			name:        "missing_path",
			params:      map[string]interface{}{"content": "data"},
			shouldError: true,
		},
		{
			name:        "missing_content",
			params:      map[string]interface{}{"path": "file.txt"},
			shouldError: true,
		},
		{
			name: "path_traversal_attempt",
			params: map[string]interface{}{
				"path":    "../../etc/passwd",
				"content": "malicious",
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewWriteFileTool(tempDir)

			// Pre-create file for overwrite test
			if tt.name == "overwrite_existing" {
				existingPath := filepath.Join(tempDir, "existing.txt")
				if err := os.WriteFile(existingPath, []byte("Old content"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
			}

			// Validate
			err := tool.Validate(tt.params)
			if tt.shouldError && err == nil {
				t.Error("expected validation error, got nil")
				return
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
				return
			}

			if tt.shouldError {
				return // Skip execution
			}

			// Execute
			result, err := tool.Execute(context.Background(), tt.params)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result == nil || !result.Success {
				t.Error("expected successful result")
			}

			// Verify file was written
			if tt.verifyFile != "" {
				filePath := filepath.Join(tempDir, tt.verifyFile)
				content, err := os.ReadFile(filePath)
				if err != nil {
					t.Errorf("failed to read written file: %v", err)
				}
				if string(content) != tt.expectData {
					t.Errorf("expected file content '%s', got '%s'", tt.expectData, string(content))
				}
			}
		})
	}
}

func TestWriteFileTool_Metadata(t *testing.T) {
	tool := NewWriteFileTool("/tmp")

	if tool.Name() != "write_file" {
		t.Errorf("expected name 'write_file', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}

	if !tool.RequiresConfirmation(SafetyMedium) {
		t.Error("write_file should require confirmation at medium safety")
	}

	if !tool.RequiresConfirmation(SafetyHigh) {
		t.Error("write_file should require confirmation at high safety")
	}

	if tool.RequiresConfirmation(SafetyLow) {
		t.Error("write_file should not require confirmation at low safety")
	}
}

func TestListFilesTool(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files and directories
	_ = os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("content"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("content"), 0644)
	_ = os.Mkdir(filepath.Join(tempDir, "subdir"), 0755)
	_ = os.WriteFile(filepath.Join(tempDir, "subdir", "nested.txt"), []byte("content"), 0644)

	tests := []struct {
		name        string
		params      map[string]interface{}
		shouldError bool
		expectCount int
	}{
		{
			name:        "list_root",
			params:      map[string]interface{}{"path": "."},
			expectCount: 3, // file1.txt, file2.txt, subdir
		},
		{
			name:        "list_default",
			params:      map[string]interface{}{},
			expectCount: 3,
		},
		{
			name:        "list_subdir",
			params:      map[string]interface{}{"path": "subdir"},
			expectCount: 1, // nested.txt
		},
		{
			name:        "list_nonexistent",
			params:      map[string]interface{}{"path": "nonexistent"},
			shouldError: true,
		},
		{
			name:        "path_traversal",
			params:      map[string]interface{}{"path": "../../../etc"},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewListFilesTool(tempDir)

			// Validate
			err := tool.Validate(tt.params)
			if tt.shouldError && err == nil && tt.name == "path_traversal" {
				t.Error("expected validation error for path traversal, got nil")
				return
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
				return
			}

			if tt.shouldError && tt.name == "path_traversal" {
				return // Skip execution if validation should fail
			}

			// Execute
			result, err := tool.Execute(context.Background(), tt.params)

			if tt.shouldError {
				if err == nil {
					t.Error("expected execution error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil || !result.Success {
					t.Error("expected successful result")
				}

				// Check file count
				if data, ok := result.Data.(map[string]interface{}); ok {
					if count, ok := data["count"].(int); ok {
						if count != tt.expectCount {
							t.Errorf("expected %d files, got %d", tt.expectCount, count)
						}
					}
				}
			}
		})
	}
}

func TestListFilesTool_Metadata(t *testing.T) {
	tool := NewListFilesTool("/tmp")

	if tool.Name() != "list_files" {
		t.Errorf("expected name 'list_files', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}

	if tool.RequiresConfirmation(SafetyHigh) {
		t.Error("list_files should not require confirmation at any safety level")
	}
}

func TestPathValidation(t *testing.T) {
	workDir := "/tmp/test"

	tests := []struct {
		name        string
		path        string
		shouldError bool
	}{
		{"valid_relative", "file.txt", false},
		{"valid_nested", "subdir/file.txt", false},
		{"valid_dot", "./file.txt", false},
		{"traversal_basic", "../etc/passwd", true},
		{"traversal_nested", "subdir/../../etc/passwd", true},
		{"traversal_absolute", "/etc/passwd", true}, // Absolute paths outside workDir should fail
		{"traversal_hidden", "sub/../../../etc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(workDir, tt.path)

			if tt.shouldError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestResolvePath(t *testing.T) {
	workDir := "/tmp/test"

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"relative", "file.txt", "/tmp/test/file.txt"},
		{"nested", "sub/file.txt", "/tmp/test/sub/file.txt"},
		{"absolute", "/etc/passwd", "/etc/passwd"},
		{"dot", "./file.txt", "/tmp/test/file.txt"},
		{"dotdot", "../file.txt", "/tmp/file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := resolvePath(workDir, tt.path)

			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestReadFileLinesTool(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file with multiple lines
	testFile := filepath.Join(tempDir, "lines.txt")
	content := `Line 1
Line 2
Line 3
Line 4
Line 5
Line 6
Line 7
Line 8
Line 9
Line 10`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		params      map[string]interface{}
		shouldError bool
		expectLines int
		expectStart string
		expectEnd   string
	}{
		{
			name: "read_specific_range",
			params: map[string]interface{}{
				"path":       "lines.txt",
				"start_line": 3.0,
				"end_line":   5.0,
			},
			expectLines: 3,
			expectStart: "Line 3",
			expectEnd:   "Line 5",
		},
		{
			name: "read_from_start",
			params: map[string]interface{}{
				"path":       "lines.txt",
				"start_line": 1.0,
				"end_line":   3.0,
			},
			expectLines: 3,
			expectStart: "Line 1",
			expectEnd:   "Line 3",
		},
		{
			name: "read_to_end",
			params: map[string]interface{}{
				"path":       "lines.txt",
				"start_line": 8.0,
				"end_line":   10.0,
			},
			expectLines: 3,
			expectStart: "Line 8",
			expectEnd:   "Line 10",
		},
		{
			name: "read_single_line",
			params: map[string]interface{}{
				"path":       "lines.txt",
				"start_line": 5.0,
				"end_line":   5.0,
			},
			expectLines: 1,
			expectStart: "Line 5",
			expectEnd:   "Line 5",
		},
		{
			name: "read_entire_file",
			params: map[string]interface{}{
				"path": "lines.txt",
			},
			expectLines: 10,
			expectStart: "Line 1",
			expectEnd:   "Line 10",
		},
		{
			name: "beyond_end",
			params: map[string]interface{}{
				"path":       "lines.txt",
				"start_line": 8.0,
				"end_line":   100.0,
			},
			expectLines: 3, // Should cap at line 10
			expectEnd:   "Line 10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewReadFileLinesTool(tempDir)

			result, err := tool.Execute(context.Background(), tt.params)

			if tt.shouldError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil || !result.Success {
					t.Error("expected successful result")
				}

				// Check line count
				if data, ok := result.Data.(map[string]interface{}); ok {
					if linesRead, ok := data["lines_read"].(int); ok {
						if linesRead != tt.expectLines {
							t.Errorf("expected %d lines, got %d", tt.expectLines, linesRead)
						}
					}
				}

				// Check content
				if tt.expectStart != "" && !contains(result.Output, tt.expectStart) {
					t.Errorf("expected output to start with '%s'", tt.expectStart)
				}

				if tt.expectEnd != "" && !contains(result.Output, tt.expectEnd) {
					t.Errorf("expected output to end with '%s'", tt.expectEnd)
				}
			}
		})
	}
}

func TestReadFileLinesTool_Metadata(t *testing.T) {
	tool := NewReadFileLinesTool("/tmp")

	if tool.Name() != "read_file_lines" {
		t.Errorf("expected name 'read_file_lines', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}

	if tool.RequiresConfirmation(SafetyHigh) {
		t.Error("read_file_lines should not require confirmation at any safety level")
	}
}

func TestFileExistsTool(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tempDir, "exists.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create test directory
	testDir := filepath.Join(tempDir, "testdir")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		shouldExist bool
		shouldBeDir bool
	}{
		{"existing_file", "exists.txt", true, false},
		{"existing_dir", "testdir", true, true},
		{"nonexistent", "nothere.txt", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewFileExistsTool(tempDir)

			params := map[string]interface{}{
				"path": tt.path,
			}

			result, err := tool.Execute(context.Background(), params)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !result.Success {
				t.Error("expected successful result")
			}

			data := result.Data.(map[string]interface{})
			exists := data["exists"].(bool)

			if exists != tt.shouldExist {
				t.Errorf("expected exists=%v, got %v", tt.shouldExist, exists)
			}

			if tt.shouldExist {
				isDir := data["is_dir"].(bool)
				if isDir != tt.shouldBeDir {
					t.Errorf("expected is_dir=%v, got %v", tt.shouldBeDir, isDir)
				}
			}
		})
	}
}

func TestFileExistsTool_Metadata(t *testing.T) {
	tool := NewFileExistsTool("/tmp")

	if tool.Name() != "file_exists" {
		t.Errorf("expected name 'file_exists', got '%s'", tool.Name())
	}

	if tool.RequiresConfirmation(SafetyHigh) {
		t.Error("file_exists should not require confirmation")
	}
}

func TestMoveFileTool(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		setup       func() string
		source      string
		dest        string
		shouldError bool
		errorMsg    string
	}{
		{
			name: "move_file",
			setup: func() string {
				path := filepath.Join(tempDir, "move_source.txt")
				_ = os.WriteFile(path, []byte("content"), 0644)
				return path
			},
			source: "move_source.txt",
			dest:   "move_dest.txt",
		},
		{
			name: "rename_file",
			setup: func() string {
				path := filepath.Join(tempDir, "old_name.txt")
				_ = os.WriteFile(path, []byte("content"), 0644)
				return path
			},
			source: "old_name.txt",
			dest:   "new_name.txt",
		},
		{
			name: "move_to_subdir",
			setup: func() string {
				path := filepath.Join(tempDir, "tomove.txt")
				_ = os.WriteFile(path, []byte("content"), 0644)
				return path
			},
			source: "tomove.txt",
			dest:   "subdir/moved.txt",
		},
		{
			name:        "source_not_exists",
			setup:       func() string { return "" },
			source:      "nonexistent.txt",
			dest:        "dest.txt",
			shouldError: true,
			errorMsg:    "does not exist",
		},
		{
			name: "dest_already_exists",
			setup: func() string {
	_ = os.WriteFile(filepath.Join(tempDir, "src.txt"), []byte("source"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "dst.txt"), []byte("dest"), 0644)
				return ""
			},
			source:      "src.txt",
			dest:        "dst.txt",
			shouldError: true,
			errorMsg:    "already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			tool := NewMoveFileTool(tempDir)

			params := map[string]interface{}{
				"source":      tt.source,
				"destination": tt.dest,
			}

			result, err := tool.Execute(context.Background(), params)

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
				if !result.Success {
					t.Errorf("expected successful result, got: %+v", result)
				}

				// Verify source no longer exists
				sourcePath := filepath.Join(tempDir, tt.source)
				if _, err := os.Stat(sourcePath); !os.IsNotExist(err) {
					t.Error("expected source to be moved (not exist)")
				}

				// Verify destination exists
				destPath := filepath.Join(tempDir, tt.dest)
				if _, err := os.Stat(destPath); os.IsNotExist(err) {
					t.Error("expected destination to exist")
				}
			}
		})
	}
}

func TestMoveFileTool_Metadata(t *testing.T) {
	tool := NewMoveFileTool("/tmp")

	if tool.Name() != "move_file" {
		t.Errorf("expected name 'move_file', got '%s'", tool.Name())
	}

	if !tool.RequiresConfirmation(SafetyMedium) {
		t.Error("move_file should require confirmation at medium safety")
	}
}

func TestDeleteFileTool(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		setup       func() string
		path        string
		shouldError bool
		errorMsg    string
	}{
		{
			name: "delete_file",
			setup: func() string {
				path := filepath.Join(tempDir, "delete_me.txt")
	_ = os.WriteFile(path, []byte("content"), 0644)
				return path
			},
			path: "delete_me.txt",
		},
		{
			name: "delete_empty_directory",
			setup: func() string {
				path := filepath.Join(tempDir, "empty_dir")
				_ = os.Mkdir(path, 0755)
				return path
			},
			path: "empty_dir",
		},
		{
			name:        "file_not_exists",
			setup:       func() string { return "" },
			path:        "nonexistent.txt",
			shouldError: true,
			errorMsg:    "does not exist",
		},
		{
			name: "non_empty_directory",
			setup: func() string {
				dirPath := filepath.Join(tempDir, "nonempty_dir")
				_ = os.Mkdir(dirPath, 0755)
	_ = os.WriteFile(filepath.Join(dirPath, "file.txt"), []byte("content"), 0644)
				return dirPath
			},
			path:        "nonempty_dir",
			shouldError: true,
			errorMsg:    "non-empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			tool := NewDeleteFileTool(tempDir)

			params := map[string]interface{}{
				"path": tt.path,
			}

			result, err := tool.Execute(context.Background(), params)

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
				if !result.Success {
					t.Errorf("expected successful result, got: %+v", result)
				}

				// Verify file no longer exists
				absPath := filepath.Join(tempDir, tt.path)
				if _, err := os.Stat(absPath); !os.IsNotExist(err) {
					t.Error("expected file to be deleted")
				}
			}
		})
	}
}

func TestDeleteFileTool_Metadata(t *testing.T) {
	tool := NewDeleteFileTool("/tmp")

	if tool.Name() != "delete_file" {
		t.Errorf("expected name 'delete_file', got '%s'", tool.Name())
	}

	if !tool.RequiresConfirmation(SafetyMedium) {
		t.Error("delete_file should require confirmation at medium safety")
	}
}

func TestAppendToFileTool(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		setup          func() string
		path           string
		content        string
		shouldError    bool
		expectContains string
	}{
		{
			name: "append_to_existing_file",
			setup: func() string {
				path := filepath.Join(tempDir, "append.txt")
	_ = os.WriteFile(path, []byte("line 1\n"), 0644)
				return path
			},
			path:           "append.txt",
			content:        "line 2\n",
			expectContains: "line 1\nline 2\n",
		},
		{
			name:           "append_to_new_file",
			setup:          func() string { return "" },
			path:           "newfile.txt",
			content:        "first line\n",
			expectContains: "first line\n",
		},
		{
			name: "append_empty_string",
			setup: func() string {
				path := filepath.Join(tempDir, "empty_append.txt")
	_ = os.WriteFile(path, []byte("content"), 0644)
				return path
			},
			path:           "empty_append.txt",
			content:        "",
			expectContains: "content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			tool := NewAppendToFileTool(tempDir)

			params := map[string]interface{}{
				"path":    tt.path,
				"content": tt.content,
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
					t.Errorf("expected successful result, got: %+v", result)
				}

				// Verify content
				absPath := filepath.Join(tempDir, tt.path)
				content, err := os.ReadFile(absPath)
				if err != nil {
					t.Fatalf("failed to read result file: %v", err)
				}

				if string(content) != tt.expectContains {
					t.Errorf("expected content '%s', got '%s'", tt.expectContains, string(content))
				}
			}
		})
	}
}

func TestAppendToFileTool_Metadata(t *testing.T) {
	tool := NewAppendToFileTool("/tmp")

	if tool.Name() != "append_to_file" {
		t.Errorf("expected name 'append_to_file', got '%s'", tool.Name())
	}

	if !tool.RequiresConfirmation(SafetyMedium) {
		t.Error("append_to_file should require confirmation at medium safety")
	}
}

func TestFindFilesByPatternTool(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	_ = os.WriteFile(filepath.Join(tempDir, "test1.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "test2.go"), []byte("package test"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "readme.md"), []byte("# README"), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "large.txt"), make([]byte, 1000), 0644)
	_ = os.WriteFile(filepath.Join(tempDir, "small.txt"), []byte("tiny"), 0644)

	tests := []struct {
		name        string
		params      map[string]interface{}
		expectCount int
		expectFiles []string
	}{
		{
			name:        "find_all_go_files",
			params:      map[string]interface{}{"pattern": "*.go"},
			expectCount: 2,
			expectFiles: []string{"test1.go", "test2.go"},
		},
		{
			name:        "find_all_files",
			params:      map[string]interface{}{},
			expectCount: 5,
		},
		{
			name: "find_by_extension",
			params: map[string]interface{}{
				"extensions": []interface{}{".md"},
			},
			expectCount: 1,
			expectFiles: []string{"readme.md"},
		},
		{
			name: "find_small_files",
			params: map[string]interface{}{
				"max_size": 100.0,
			},
			expectCount: 4, // All except large.txt
		},
		{
			name: "find_large_files",
			params: map[string]interface{}{
				"min_size": 500.0,
			},
			expectCount: 1,
			expectFiles: []string{"large.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewFindFilesByPatternTool(tempDir)

			result, err := tool.Execute(context.Background(), tt.params)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !result.Success {
				t.Error("expected successful result")
			}

			data := result.Data.(map[string]interface{})
			count := data["count"].(int)

			if count != tt.expectCount {
				t.Errorf("expected %d files, got %d", tt.expectCount, count)
			}

			// Check specific files if provided
			if len(tt.expectFiles) > 0 {
				files := data["files"].([]map[string]interface{})
				foundFiles := make(map[string]bool)
				for _, f := range files {
					foundFiles[f["path"].(string)] = true
				}

				for _, expected := range tt.expectFiles {
					if !foundFiles[expected] {
						t.Errorf("expected to find file '%s'", expected)
					}
				}
			}
		})
	}
}

func TestFindFilesByPatternTool_Metadata(t *testing.T) {
	tool := NewFindFilesByPatternTool("/tmp")

	if tool.Name() != "find_files_by_pattern" {
		t.Errorf("expected name 'find_files_by_pattern', got '%s'", tool.Name())
	}

	if tool.RequiresConfirmation(SafetyHigh) {
		t.Error("find_files_by_pattern should not require confirmation")
	}
}
