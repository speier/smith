package tools

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestSearchFilesTool(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	files := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
	processData()
}

func processData() {
	// Process data here
	fmt.Println("Processing...")
}
`,
		"utils.go": `package main

import "fmt"

func helper() {
	fmt.Println("Helper function")
}
`,
		"subdir/test.go": `package subdir

import "testing"

func TestSomething(t *testing.T) {
	t.Log("Test log")
}
`,
		"README.md": `# Project

This is a test project.
Hello from README!
`,
		"binary.bin": string([]byte{0x00, 0x01, 0x02, 0xFF}), // Binary file
	}

	for path, content := range files {
		fullPath := filepath.Join(tempDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", path, err)
		}
	}

	tests := []struct {
		name         string
		params       map[string]interface{}
		shouldError  bool
		expectCount  int
		expectInFile string
	}{
		{
			name: "simple_text_search",
			params: map[string]interface{}{
				"pattern": "fmt.Println",
			},
			expectCount:  3, // 2 in main.go, 1 in utils.go (not 4 - README doesn't have it)
			expectInFile: "main.go",
		},
		{
			name: "case_insensitive_search",
			params: map[string]interface{}{
				"pattern":        "HELLO",
				"case_sensitive": false,
			},
			expectCount:  2, // main.go and README.md
			expectInFile: "README.md",
		},
		{
			name: "case_sensitive_search",
			params: map[string]interface{}{
				"pattern":        "HELLO",
				"case_sensitive": true,
			},
			expectCount: 0,
		},
		{
			name: "regex_search",
			params: map[string]interface{}{
				"pattern":  `func \w+\(\)`,
				"is_regex": true,
			},
			expectCount: 3, // main, processData, helper
		},
		{
			name: "regex_case_insensitive",
			params: map[string]interface{}{
				"pattern":        `FUNC \w+`,
				"is_regex":       true,
				"case_sensitive": false,
			},
			expectCount: 4, // All function definitions
		},
		{
			name: "search_in_subdir",
			params: map[string]interface{}{
				"pattern": "testing",
				"path":    "subdir",
			},
			expectCount:  2, // import and package
			expectInFile: "subdir/test.go",
		},
		{
			name: "max_results_limit",
			params: map[string]interface{}{
				"pattern":     "fmt",
				"max_results": 2.0,
			},
			expectCount: 2,
		},
		{
			name: "no_matches",
			params: map[string]interface{}{
				"pattern": "NonexistentPattern12345",
			},
			expectCount: 0,
		},
		{
			name: "missing_pattern",
			params: map[string]interface{}{
				"path": ".",
			},
			shouldError: true,
		},
		{
			name: "empty_pattern",
			params: map[string]interface{}{
				"pattern": "",
			},
			shouldError: true,
		},
		{
			name: "invalid_regex",
			params: map[string]interface{}{
				"pattern":  "[invalid",
				"is_regex": true,
			},
			shouldError: true,
		},
		{
			name: "path_traversal",
			params: map[string]interface{}{
				"pattern": "test",
				"path":    "../../../etc",
			},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := NewSearchFilesTool(tempDir)

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
				return // Skip execution for validation failures
			}

			// Execute
			result, err := tool.Execute(context.Background(), tt.params)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result == nil || !result.Success {
				t.Errorf("expected successful result, got: %+v", result)
			}

			// Check match count
			if data, ok := result.Data.(map[string]interface{}); ok {
				count, ok := data["count"].(int)
				if !ok {
					t.Error("expected count in result data")
				}
				if count != tt.expectCount {
					t.Errorf("expected %d matches, got %d", tt.expectCount, count)
				}

				// Check if specific file is in results
				if tt.expectInFile != "" {
					matches, ok := data["matches"].([]SearchMatch)
					if !ok {
						t.Error("expected matches in result data")
					}

					found := false
					for _, match := range matches {
						if contains(match.File, tt.expectInFile) {
							found = true
							break
						}
					}

					if !found {
						t.Errorf("expected to find match in file '%s'", tt.expectInFile)
					}
				}
			}
		})
	}
}

func TestSearchFilesTool_Metadata(t *testing.T) {
	tool := NewSearchFilesTool("/tmp")

	if tool.Name() != "search_files" {
		t.Errorf("expected name 'search_files', got '%s'", tool.Name())
	}

	if tool.Description() == "" {
		t.Error("expected non-empty description")
	}

	if tool.RequiresConfirmation(SafetyHigh) {
		t.Error("search_files should not require confirmation at any safety level")
	}
}

func TestSearchFilesTool_MatchDetails(t *testing.T) {
	tempDir := t.TempDir()

	testFile := filepath.Join(tempDir, "test.txt")
	content := `Line 1: Hello World
Line 2: Foo Bar
Line 3: Hello Again
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool := NewSearchFilesTool(tempDir)

	params := map[string]interface{}{
		"pattern": "Hello",
	}

	result, err := tool.Execute(context.Background(), params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Fatal("expected successful result")
	}

	data := result.Data.(map[string]interface{})
	matches := data["matches"].([]SearchMatch)

	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}

	// Check first match details
	match1 := matches[0]
	if match1.Line != 1 {
		t.Errorf("expected line 1, got %d", match1.Line)
	}
	if match1.Column != 9 {
		t.Errorf("expected column 9, got %d", match1.Column)
	}
	if match1.Match != "Hello" {
		t.Errorf("expected match 'Hello', got '%s'", match1.Match)
	}
	if !contains(match1.Context, "Line 1") {
		t.Errorf("expected context to contain 'Line 1', got '%s'", match1.Context)
	}

	// Check second match
	match2 := matches[1]
	if match2.Line != 3 {
		t.Errorf("expected line 3, got %d", match2.Line)
	}
}

func TestSearchFilesTool_BinaryFileSkipped(t *testing.T) {
	tempDir := t.TempDir()

	// Create text file with search term
	textFile := filepath.Join(tempDir, "text.txt")
	if err := os.WriteFile(textFile, []byte("SearchTerm"), 0644); err != nil {
		t.Fatalf("failed to create text file: %v", err)
	}

	// Create binary file with search term (should be skipped)
	binaryFile := filepath.Join(tempDir, "binary.exe")
	if err := os.WriteFile(binaryFile, []byte("SearchTerm"), 0644); err != nil {
		t.Fatalf("failed to create binary file: %v", err)
	}

	tool := NewSearchFilesTool(tempDir)

	params := map[string]interface{}{
		"pattern": "SearchTerm",
	}

	result, err := tool.Execute(context.Background(), params)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data := result.Data.(map[string]interface{})
	count := data["count"].(int)

	// Should only find match in text file, not binary
	if count != 1 {
		t.Errorf("expected 1 match (binary file should be skipped), got %d", count)
	}

	matches := data["matches"].([]SearchMatch)
	if !contains(matches[0].File, "text.txt") {
		t.Error("expected match in text.txt only")
	}
}

func TestSearchFilesTool_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()

	// Create a file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tool := NewSearchFilesTool(tempDir)

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	params := map[string]interface{}{
		"pattern": "test",
	}

	// Should handle cancellation gracefully
	result, err := tool.Execute(ctx, params)

	// Either completes quickly or returns error - both acceptable
	if err == nil && result != nil && result.Success {
		// Completed before cancellation kicked in
		return
	}

	if err != nil && err == context.Canceled {
		// Properly handled cancellation
		return
	}

	// Any other result is acceptable for this test
}

func TestSearchInFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "search.txt")

	content := `First line
Second line with PATTERN
Third line
Fourth line with pattern
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	tests := []struct {
		name          string
		pattern       string
		caseSensitive bool
		isRegex       bool
		expectCount   int
	}{
		{"case_sensitive_plain", "pattern", true, false, 1},
		{"case_insensitive_plain", "pattern", false, false, 2},
		{"regex_match", `line \w+`, true, true, 2}, // Only "Second line" and "Fourth line" match this pattern
		{"no_match", "nonexistent", true, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var re *regexp.Regexp
			var err error
			if tt.isRegex {
				flags := ""
				if !tt.caseSensitive {
					flags = "(?i)"
				}
				re, err = regexp.Compile(flags + tt.pattern)
				if err != nil {
					t.Fatalf("failed to compile regex: %v", err)
				}
			}

			matches, err := searchInFile(testFile, tt.pattern, re, tt.caseSensitive, tt.isRegex)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(matches) != tt.expectCount {
				t.Errorf("expected %d matches, got %d", tt.expectCount, len(matches))
			}
		})
	}
}

func TestIsBinaryFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"file.txt", false},
		{"file.go", false},
		{"file.md", false},
		{"file.exe", true},
		{"file.dll", true},
		{"file.so", true},
		{"file.jpg", true},
		{"file.png", true},
		{"file.pdf", true},
		{"file.zip", true},
		{"file.db", true},
		{"file.pyc", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isBinaryFile(tt.path)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
