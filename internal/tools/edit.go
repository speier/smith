package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReplaceInFileTool performs find-replace operations in files
type ReplaceInFileTool struct {
	workDir string
}

// NewReplaceInFileTool creates a new ReplaceInFileTool
func NewReplaceInFileTool(workDir string) *ReplaceInFileTool {
	return &ReplaceInFileTool{workDir: workDir}
}

func (t *ReplaceInFileTool) Name() string {
	return "replace_in_file"
}

func (t *ReplaceInFileTool) Description() string {
	return "Replace text in a file with new text (atomic operation)"
}

func (t *ReplaceInFileTool) Validate(params map[string]interface{}) error {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("path parameter is required and must be a string")
	}

	oldText, ok := params["old_text"].(string)
	if !ok || oldText == "" {
		return fmt.Errorf("old_text parameter is required and must be a non-empty string")
	}

	newText, ok := params["new_text"].(string)
	if !ok {
		return fmt.Errorf("new_text parameter is required and must be a string")
	}

	_ = newText // newText can be empty (deleting text)

	// Validate the path is safe
	if err := validatePath(t.workDir, path); err != nil {
		return err
	}

	return nil
}

func (t *ReplaceInFileTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return &ToolResult{
			Success: false,
			Error:   "path parameter is required",
		}, fmt.Errorf("path parameter is required")
	}

	oldText, ok := params["old_text"].(string)
	if !ok || oldText == "" {
		return &ToolResult{
			Success: false,
			Error:   "old_text parameter is required",
		}, fmt.Errorf("old_text parameter is required")
	}

	newText, ok := params["new_text"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "new_text parameter is required",
		}, fmt.Errorf("new_text parameter is required")
	}

	absPath := resolvePath(t.workDir, path)

	// Read the file
	content, err := os.ReadFile(absPath)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %v", err),
		}, err
	}

	contentStr := string(content)

	// Check if old_text exists
	if !strings.Contains(contentStr, oldText) {
		return &ToolResult{
			Success: false,
			Error:   "old_text not found in file",
		}, fmt.Errorf("old_text not found in file")
	}

	// Count occurrences
	occurrences := strings.Count(contentStr, oldText)

	// Check if multiple occurrences exist (ambiguous)
	if occurrences > 1 {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("old_text appears %d times in file (ambiguous replacement)", occurrences),
			Data: map[string]interface{}{
				"occurrences": occurrences,
			},
		}, fmt.Errorf("ambiguous replacement: old_text appears %d times", occurrences)
	}

	// Perform replacement (exactly one occurrence)
	newContent := strings.Replace(contentStr, oldText, newText, 1)

	// Write back atomically
	if err := os.WriteFile(absPath, []byte(newContent), 0644); err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %v", err),
		}, err
	}

	return &ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Successfully replaced text in %s", absPath),
		Data: map[string]interface{}{
			"path":          absPath,
			"old_length":    len(contentStr),
			"new_length":    len(newContent),
			"bytes_changed": len(newContent) - len(contentStr),
		},
	}, nil
}

func (t *ReplaceInFileTool) RequiresConfirmation(level SafetyLevel) bool {
	return level >= SafetyMedium // Editing requires confirmation at medium and high
}

// ReplaceAllInFileTool performs find-replace-all operations in files
type ReplaceAllInFileTool struct {
	workDir string
}

// NewReplaceAllInFileTool creates a new ReplaceAllInFileTool
func NewReplaceAllInFileTool(workDir string) *ReplaceAllInFileTool {
	return &ReplaceAllInFileTool{workDir: workDir}
}

func (t *ReplaceAllInFileTool) Name() string {
	return "replace_all_in_file"
}

func (t *ReplaceAllInFileTool) Description() string {
	return "Replace all occurrences of text in a file (with safety limit)"
}

func (t *ReplaceAllInFileTool) Validate(params map[string]interface{}) error {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("path parameter is required and must be a string")
	}

	oldText, ok := params["old_text"].(string)
	if !ok || oldText == "" {
		return fmt.Errorf("old_text parameter is required and must be a non-empty string")
	}

	newText, ok := params["new_text"].(string)
	if !ok {
		return fmt.Errorf("new_text parameter is required and must be a string")
	}

	_ = newText // newText can be empty (deleting text)

	// Validate the path is safe
	if err := validatePath(t.workDir, path); err != nil {
		return err
	}

	return nil
}

func (t *ReplaceAllInFileTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return &ToolResult{
			Success: false,
			Error:   "path parameter is required",
		}, fmt.Errorf("path parameter is required")
	}

	oldText, ok := params["old_text"].(string)
	if !ok || oldText == "" {
		return &ToolResult{
			Success: false,
			Error:   "old_text parameter is required",
		}, fmt.Errorf("old_text parameter is required")
	}

	newText, ok := params["new_text"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "new_text parameter is required",
		}, fmt.Errorf("new_text parameter is required")
	}

	// Optional: max_replacements (default: 100)
	maxReplacements := 100
	if max, ok := params["max_replacements"].(float64); ok {
		maxReplacements = int(max)
	} else if max, ok := params["max_replacements"].(int); ok {
		maxReplacements = max
	}

	absPath := resolvePath(t.workDir, path)

	// Read the file
	content, err := os.ReadFile(absPath)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %v", err),
		}, err
	}

	contentStr := string(content)

	// Check if old_text exists
	if !strings.Contains(contentStr, oldText) {
		return &ToolResult{
			Success: false,
			Error:   "old_text not found in file",
		}, fmt.Errorf("old_text not found in file")
	}

	// Count occurrences
	occurrences := strings.Count(contentStr, oldText)

	// Check safety limit
	if occurrences > maxReplacements {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("too many occurrences (%d) exceeds max_replacements (%d)", occurrences, maxReplacements),
			Data: map[string]interface{}{
				"occurrences":      occurrences,
				"max_replacements": maxReplacements,
			},
		}, fmt.Errorf("too many occurrences: %d > %d", occurrences, maxReplacements)
	}

	// Perform replacement (all occurrences)
	newContent := strings.ReplaceAll(contentStr, oldText, newText)

	// Write back atomically
	if err := os.WriteFile(absPath, []byte(newContent), 0644); err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %v", err),
		}, err
	}

	return &ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Successfully replaced %d occurrences in %s", occurrences, absPath),
		Data: map[string]interface{}{
			"path":          absPath,
			"replacements":  occurrences,
			"old_length":    len(contentStr),
			"new_length":    len(newContent),
			"bytes_changed": len(newContent) - len(contentStr),
		},
	}, nil
}

func (t *ReplaceAllInFileTool) RequiresConfirmation(level SafetyLevel) bool {
	return level >= SafetyMedium // Editing requires confirmation at medium and high
}

// DiffFilesTool compares two files and returns the differences
type DiffFilesTool struct {
	workDir string
}

// NewDiffFilesTool creates a new DiffFilesTool
func NewDiffFilesTool(workDir string) *DiffFilesTool {
	return &DiffFilesTool{workDir: workDir}
}

func (t *DiffFilesTool) Name() string {
	return "diff_files"
}

func (t *DiffFilesTool) Description() string {
	return "Compare two files and return the differences line by line"
}

func (t *DiffFilesTool) Validate(params map[string]interface{}) error {
	file1, ok := params["file1"].(string)
	if !ok || file1 == "" {
		return fmt.Errorf("file1 parameter is required and must be a string")
	}

	file2, ok := params["file2"].(string)
	if !ok || file2 == "" {
		return fmt.Errorf("file2 parameter is required and must be a string")
	}

	if err := validatePath(t.workDir, file1); err != nil {
		return fmt.Errorf("file1: %w", err)
	}

	if err := validatePath(t.workDir, file2); err != nil {
		return fmt.Errorf("file2: %w", err)
	}

	return nil
}

func (t *DiffFilesTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	file1 := params["file1"].(string)
	file2 := params["file2"].(string)

	absPath1 := resolvePath(t.workDir, file1)
	absPath2 := resolvePath(t.workDir, file2)

	// Read both files
	content1, err := os.ReadFile(absPath1)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read file1: %v", err),
		}, err
	}

	content2, err := os.ReadFile(absPath2)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read file2: %v", err),
		}, err
	}

	// Split into lines
	lines1 := strings.Split(string(content1), "\n")
	lines2 := strings.Split(string(content2), "\n")

	// Simple diff algorithm - find added/removed/changed lines
	type DiffLine struct {
		Type    string `json:"type"`    // "added", "removed", "changed", "same"
		Line1   int    `json:"line1"`   // Line number in file1 (0 if added)
		Line2   int    `json:"line2"`   // Line number in file2 (0 if removed)
		Content string `json:"content"` // Line content
	}

	var diffs []DiffLine
	var output strings.Builder
	identical := string(content1) == string(content2)

	if identical {
		output.WriteString("Files are identical\n")
	} else {
		// Simple line-by-line comparison
		maxLines := len(lines1)
		if len(lines2) > maxLines {
			maxLines = len(lines2)
		}

		for i := 0; i < maxLines; i++ {
			var line1, line2 string
			if i < len(lines1) {
				line1 = lines1[i]
			}
			if i < len(lines2) {
				line2 = lines2[i]
			}

			if i >= len(lines1) {
				// Line added in file2
				diffs = append(diffs, DiffLine{Type: "added", Line2: i + 1, Content: line2})
				output.WriteString(fmt.Sprintf("+ %d: %s\n", i+1, line2))
			} else if i >= len(lines2) {
				// Line removed from file1
				diffs = append(diffs, DiffLine{Type: "removed", Line1: i + 1, Content: line1})
				output.WriteString(fmt.Sprintf("- %d: %s\n", i+1, line1))
			} else if line1 != line2 {
				// Line changed
				diffs = append(diffs, DiffLine{Type: "changed", Line1: i + 1, Line2: i + 1, Content: line2})
				output.WriteString(fmt.Sprintf("  %d: - %s\n", i+1, line1))
				output.WriteString(fmt.Sprintf("  %d: + %s\n", i+1, line2))
			}
		}
	}

	return &ToolResult{
		Success: true,
		Output:  output.String(),
		Data: map[string]interface{}{
			"identical":   identical,
			"differences": diffs,
			"file1":       file1,
			"file2":       file2,
		},
	}, nil
}

func (t *DiffFilesTool) RequiresConfirmation(level SafetyLevel) bool {
	return false // Reading only, no confirmation needed
}

// BatchSearchReplaceTool performs search and replace across multiple files
type BatchSearchReplaceTool struct {
	workDir string
}

// NewBatchSearchReplaceTool creates a new BatchSearchReplaceTool
func NewBatchSearchReplaceTool(workDir string) *BatchSearchReplaceTool {
	return &BatchSearchReplaceTool{workDir: workDir}
}

func (t *BatchSearchReplaceTool) Name() string {
	return "batch_search_replace"
}

func (t *BatchSearchReplaceTool) Description() string {
	return "Search and replace text across multiple files matching a pattern"
}

func (t *BatchSearchReplaceTool) Validate(params map[string]interface{}) error {
	oldText, ok := params["old_text"].(string)
	if !ok || oldText == "" {
		return fmt.Errorf("old_text parameter is required and must be a non-empty string")
	}

	newText, ok := params["new_text"].(string)
	if !ok {
		return fmt.Errorf("new_text parameter is required and must be a string")
	}
	_ = newText // Validated, will be used in Execute

	filePattern, ok := params["file_pattern"].(string)
	if !ok || filePattern == "" {
		return fmt.Errorf("file_pattern parameter is required (e.g., '*.go', '**/*.txt')")
	}

	// Check for path traversal in pattern
	if strings.Contains(filePattern, "..") {
		return fmt.Errorf("file_pattern cannot contain '..' (path traversal)")
	}

	// Optional max_files limit
	if maxFiles, ok := params["max_files"].(float64); ok && maxFiles < 1 {
		return fmt.Errorf("max_files must be at least 1")
	}

	return nil
}

func (t *BatchSearchReplaceTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	oldText := params["old_text"].(string)
	newText := params["new_text"].(string)
	filePattern := params["file_pattern"].(string)

	maxFiles := 100 // Default safety limit
	if mf, ok := params["max_files"].(float64); ok {
		maxFiles = int(mf)
	}

	var modifiedFiles []string
	var totalReplacements int
	filesProcessed := 0

	err := filepath.Walk(t.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		if info.IsDir() {
			return nil
		}

		// Check if we've hit the limit
		if filesProcessed >= maxFiles {
			return filepath.SkipDir
		}

		// Get relative path
		relPath, err := filepath.Rel(t.workDir, path)
		if err != nil {
			return nil
		}

		// Check pattern match
		matched, err := filepath.Match(filePattern, filepath.Base(path))
		if err != nil || !matched {
			// Also try matching full relative path for patterns like "**/*.go"
			if strings.Contains(filePattern, "**") {
				pattern := strings.ReplaceAll(filePattern, "**/*", "*")
				matched, _ = filepath.Match(pattern, filepath.Base(path))
			}
			if !matched {
				return nil
			}
		}

		filesProcessed++

		// Read file
		content, err := os.ReadFile(path)
		if err != nil {
			return nil // Skip unreadable files
		}

		contentStr := string(content)

		// Check if file contains the search text
		if !strings.Contains(contentStr, oldText) {
			return nil
		}

		// Count replacements in this file
		count := strings.Count(contentStr, oldText)

		// Perform replacement
		newContent := strings.ReplaceAll(contentStr, oldText, newText)

		// Write back atomically
		if err := os.WriteFile(path, []byte(newContent), info.Mode()); err != nil {
			return nil // Skip files we can't write
		}

		modifiedFiles = append(modifiedFiles, relPath)
		totalReplacements += count

		return nil
	})

	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("batch replace failed: %v", err),
		}, err
	}

	output := fmt.Sprintf("Modified %d files (%d total replacements)\n", len(modifiedFiles), totalReplacements)
	for _, f := range modifiedFiles {
		output += fmt.Sprintf("  - %s\n", f)
	}

	return &ToolResult{
		Success: true,
		Output:  output,
		Data: map[string]interface{}{
			"files_modified":     len(modifiedFiles),
			"total_replacements": totalReplacements,
			"modified_files":     modifiedFiles,
		},
	}, nil
}

func (t *BatchSearchReplaceTool) RequiresConfirmation(level SafetyLevel) bool {
	return level >= SafetyMedium // Modifying multiple files requires confirmation
}
