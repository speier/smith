package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFileTool reads the contents of a file
type ReadFileTool struct {
	workDir string
}

// NewReadFileTool creates a new ReadFileTool
func NewReadFileTool(workDir string) *ReadFileTool {
	return &ReadFileTool{workDir: workDir}
}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Read the contents of a file"
}

func (t *ReadFileTool) Validate(params map[string]interface{}) error {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("path parameter is required and must be a string")
	}

	// Validate the path is safe (no path traversal outside workDir)
	if err := validatePath(t.workDir, path); err != nil {
		return err
	}

	return nil
}

func (t *ReadFileTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return &ToolResult{
			Success: false,
			Error:   "path parameter is required and must be a string",
		}, fmt.Errorf("path parameter is required")
	}

	absPath := resolvePath(t.workDir, path)

	content, err := os.ReadFile(absPath)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %v", err),
		}, err
	}

	return &ToolResult{
		Success: true,
		Output:  string(content),
		Data: map[string]interface{}{
			"path": absPath,
			"size": len(content),
		},
	}, nil
}

func (t *ReadFileTool) RequiresConfirmation(level SafetyLevel) bool {
	return false // Reading is safe at all levels
}

// ReadFileLinesTool reads specific lines from a file
type ReadFileLinesTool struct {
	workDir string
}

// NewReadFileLinesTool creates a new ReadFileLinesTool
func NewReadFileLinesTool(workDir string) *ReadFileLinesTool {
	return &ReadFileLinesTool{workDir: workDir}
}

func (t *ReadFileLinesTool) Name() string {
	return "read_file_lines"
}

func (t *ReadFileLinesTool) Description() string {
	return "Read specific lines from a file (for large files, read only what you need)"
}

func (t *ReadFileLinesTool) Validate(params map[string]interface{}) error {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("path parameter is required and must be a string")
	}

	// Validate the path is safe
	if err := validatePath(t.workDir, path); err != nil {
		return err
	}

	// Optional: validate line numbers if provided
	if startLine, ok := params["start_line"].(float64); ok {
		if startLine < 1 {
			return fmt.Errorf("start_line must be >= 1")
		}
	}

	if endLine, ok := params["end_line"].(float64); ok {
		if endLine < 1 {
			return fmt.Errorf("end_line must be >= 1")
		}
	}

	return nil
}

func (t *ReadFileLinesTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return &ToolResult{
			Success: false,
			Error:   "path parameter is required",
		}, fmt.Errorf("path parameter is required")
	}

	absPath := resolvePath(t.workDir, path)

	// Read the entire file
	content, err := os.ReadFile(absPath)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to read file: %v", err),
		}, err
	}

	lines := strings.Split(string(content), "\n")
	totalLines := len(lines)

	// Parse line range (optional parameters)
	startLine := 1
	if sl, ok := params["start_line"].(float64); ok {
		startLine = int(sl)
	} else if sl, ok := params["start_line"].(int); ok {
		startLine = sl
	}

	endLine := totalLines
	if el, ok := params["end_line"].(float64); ok {
		endLine = int(el)
	} else if el, ok := params["end_line"].(int); ok {
		endLine = el
	}

	// Validate range
	if startLine < 1 {
		startLine = 1
	}
	if endLine > totalLines {
		endLine = totalLines
	}
	if startLine > endLine {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("start_line (%d) must be <= end_line (%d)", startLine, endLine),
		}, fmt.Errorf("invalid line range")
	}

	// Extract lines (convert to 0-indexed)
	selectedLines := lines[startLine-1 : endLine]
	output := strings.Join(selectedLines, "\n")

	return &ToolResult{
		Success: true,
		Output:  output,
		Data: map[string]interface{}{
			"path":        absPath,
			"total_lines": totalLines,
			"start_line":  startLine,
			"end_line":    endLine,
			"lines_read":  len(selectedLines),
		},
	}, nil
}

func (t *ReadFileLinesTool) RequiresConfirmation(level SafetyLevel) bool {
	return false // Reading is safe at all levels
}

// WriteFileTool writes content to a file
type WriteFileTool struct {
	workDir string
}

// NewWriteFileTool creates a new WriteFileTool
func NewWriteFileTool(workDir string) *WriteFileTool {
	return &WriteFileTool{workDir: workDir}
}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return "Write content to a file"
}

func (t *WriteFileTool) Validate(params map[string]interface{}) error {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("path parameter is required and must be a string")
	}

	content, ok := params["content"].(string)
	if !ok {
		return fmt.Errorf("content parameter is required and must be a string")
	}

	_ = content // content is validated but not used here

	// Validate the path is safe
	if err := validatePath(t.workDir, path); err != nil {
		return err
	}

	return nil
}

func (t *WriteFileTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return &ToolResult{
			Success: false,
			Error:   "path parameter is required and must be a string",
		}, fmt.Errorf("path parameter is required")
	}

	content, ok := params["content"].(string)
	if !ok {
		return &ToolResult{
			Success: false,
			Error:   "content parameter is required and must be a string",
		}, fmt.Errorf("content parameter is required")
	}

	absPath := resolvePath(t.workDir, path)

	// Create directory if it doesn't exist
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create directory: %v", err),
		}, err
	}

	// Write the file
	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write file: %v", err),
		}, err
	}

	return &ToolResult{
		Success: true,
		Output:  fmt.Sprintf("File written successfully: %s", absPath),
		Data: map[string]interface{}{
			"path": absPath,
			"size": len(content),
		},
	}, nil
}

func (t *WriteFileTool) RequiresConfirmation(level SafetyLevel) bool {
	return level >= SafetyMedium // Writing requires confirmation at medium and high
}

// ListFilesTool lists files in a directory
type ListFilesTool struct {
	workDir string
}

// NewListFilesTool creates a new ListFilesTool
func NewListFilesTool(workDir string) *ListFilesTool {
	return &ListFilesTool{workDir: workDir}
}

func (t *ListFilesTool) Name() string {
	return "list_files"
}

func (t *ListFilesTool) Description() string {
	return "List files in a directory"
}

func (t *ListFilesTool) Validate(params map[string]interface{}) error {
	path, ok := params["path"].(string)
	if !ok {
		path = "." // Default to current directory
	}

	// Validate the path is safe
	if err := validatePath(t.workDir, path); err != nil {
		return err
	}

	return nil
}

func (t *ListFilesTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path, ok := params["path"].(string)
	if !ok {
		path = "."
	}

	absPath := resolvePath(t.workDir, path)

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to list directory: %v", err),
		}, err
	}

	files := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, map[string]interface{}{
			"name":  entry.Name(),
			"isDir": entry.IsDir(),
			"size":  info.Size(),
		})
	}

	return &ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Found %d entries in %s", len(files), absPath),
		Data: map[string]interface{}{
			"path":  absPath,
			"files": files,
			"count": len(files),
		},
	}, nil
}

func (t *ListFilesTool) RequiresConfirmation(level SafetyLevel) bool {
	return false // Listing is safe at all levels
}

// Helper functions for path validation

func validatePath(workDir, path string) error {
	// If path is absolute and not under workDir, reject it
	if filepath.IsAbs(path) {
		cleanPath := filepath.Clean(path)
		relPath, err := filepath.Rel(workDir, cleanPath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			return ErrInvalidPath
		}
	}

	absPath := resolvePath(workDir, path)

	// Ensure the resolved path is within workDir
	relPath, err := filepath.Rel(workDir, absPath)
	if err != nil {
		return ErrInvalidPath
	}

	// Check for path traversal attempts
	if strings.HasPrefix(relPath, "..") || strings.Contains(relPath, string(filepath.Separator)+"..") {
		return ErrInvalidPath
	}

	return nil
}

func resolvePath(workDir, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(workDir, path)
}

// FileExistsTool checks if a file or directory exists
type FileExistsTool struct {
	workDir string
}

// NewFileExistsTool creates a new FileExistsTool
func NewFileExistsTool(workDir string) *FileExistsTool {
	return &FileExistsTool{workDir: workDir}
}

func (t *FileExistsTool) Name() string {
	return "file_exists"
}

func (t *FileExistsTool) Description() string {
	return "Check if a file or directory exists and get basic info"
}

func (t *FileExistsTool) Validate(params map[string]interface{}) error {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("path parameter is required and must be a string")
	}

	// Validate the path is safe
	if err := validatePath(t.workDir, path); err != nil {
		return err
	}

	return nil
}

func (t *FileExistsTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return &ToolResult{
			Success: false,
			Error:   "path parameter is required",
		}, fmt.Errorf("path parameter is required")
	}

	absPath := resolvePath(t.workDir, path)

	// Check if path exists
	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return &ToolResult{
			Success: true,
			Output:  fmt.Sprintf("Path does not exist: %s", absPath),
			Data: map[string]interface{}{
				"path":   absPath,
				"exists": false,
			},
		}, nil
	}

	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to check path: %v", err),
		}, err
	}

	// Path exists, return info
	return &ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Path exists: %s", absPath),
		Data: map[string]interface{}{
			"path":     absPath,
			"exists":   true,
			"is_dir":   info.IsDir(),
			"size":     info.Size(),
			"modified": info.ModTime().Unix(),
		},
	}, nil
}

func (t *FileExistsTool) RequiresConfirmation(level SafetyLevel) bool {
	return false // Checking existence is safe at all levels
}

// MoveFileTool moves or renames a file
type MoveFileTool struct {
	workDir string
}

// NewMoveFileTool creates a new MoveFileTool
func NewMoveFileTool(workDir string) *MoveFileTool {
	return &MoveFileTool{workDir: workDir}
}

func (t *MoveFileTool) Name() string {
	return "move_file"
}

func (t *MoveFileTool) Description() string {
	return "Move or rename a file or directory"
}

func (t *MoveFileTool) Validate(params map[string]interface{}) error {
	source, ok := params["source"].(string)
	if !ok || source == "" {
		return fmt.Errorf("source parameter is required and must be a string")
	}

	dest, ok := params["destination"].(string)
	if !ok || dest == "" {
		return fmt.Errorf("destination parameter is required and must be a string")
	}

	// Validate both paths are safe
	if err := validatePath(t.workDir, source); err != nil {
		return fmt.Errorf("invalid source path: %v", err)
	}

	if err := validatePath(t.workDir, dest); err != nil {
		return fmt.Errorf("invalid destination path: %v", err)
	}

	return nil
}

func (t *MoveFileTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	source, ok := params["source"].(string)
	if !ok || source == "" {
		return &ToolResult{
			Success: false,
			Error:   "source parameter is required",
		}, fmt.Errorf("source parameter is required")
	}

	dest, ok := params["destination"].(string)
	if !ok || dest == "" {
		return &ToolResult{
			Success: false,
			Error:   "destination parameter is required",
		}, fmt.Errorf("destination parameter is required")
	}

	sourcePath := resolvePath(t.workDir, source)
	destPath := resolvePath(t.workDir, dest)

	// Check if source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("source file does not exist: %s", sourcePath),
		}, fmt.Errorf("source file does not exist")
	}

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("destination already exists: %s", destPath),
		}, fmt.Errorf("destination already exists")
	}

	// Create destination directory if needed
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to create destination directory: %v", err),
		}, err
	}

	// Perform the move
	if err := os.Rename(sourcePath, destPath); err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to move file: %v", err),
		}, err
	}

	return &ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Successfully moved %s to %s", sourcePath, destPath),
		Data: map[string]interface{}{
			"source":      sourcePath,
			"destination": destPath,
		},
	}, nil
}

func (t *MoveFileTool) RequiresConfirmation(level SafetyLevel) bool {
	return level >= SafetyMedium // Moving files requires confirmation at medium and high
}

// DeleteFileTool deletes a file or directory
type DeleteFileTool struct {
	workDir string
}

// NewDeleteFileTool creates a new DeleteFileTool
func NewDeleteFileTool(workDir string) *DeleteFileTool {
	return &DeleteFileTool{workDir: workDir}
}

func (t *DeleteFileTool) Name() string {
	return "delete_file"
}

func (t *DeleteFileTool) Description() string {
	return "Delete a file or empty directory"
}

func (t *DeleteFileTool) Validate(params map[string]interface{}) error {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("path parameter is required and must be a string")
	}

	// Validate the path is safe
	if err := validatePath(t.workDir, path); err != nil {
		return err
	}

	return nil
}

func (t *DeleteFileTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return &ToolResult{
			Success: false,
			Error:   "path parameter is required",
		}, fmt.Errorf("path parameter is required")
	}

	absPath := resolvePath(t.workDir, path)

	// Check if path exists
	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("path does not exist: %s", absPath),
		}, fmt.Errorf("path does not exist")
	}

	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to check path: %v", err),
		}, err
	}

	// Safety check: don't delete directories recursively by default
	if info.IsDir() {
		// Check if directory is empty
		entries, err := os.ReadDir(absPath)
		if err != nil {
			return &ToolResult{
				Success: false,
				Error:   fmt.Sprintf("failed to read directory: %v", err),
			}, err
		}

		if len(entries) > 0 {
			return &ToolResult{
				Success: false,
				Error:   "cannot delete non-empty directory (safety check)",
			}, fmt.Errorf("directory not empty")
		}
	}

	// Delete the file or empty directory
	if err := os.Remove(absPath); err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to delete: %v", err),
		}, err
	}

	fileType := "file"
	if info.IsDir() {
		fileType = "directory"
	}

	return &ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Successfully deleted %s: %s", fileType, absPath),
		Data: map[string]interface{}{
			"path": absPath,
			"type": fileType,
		},
	}, nil
}

func (t *DeleteFileTool) RequiresConfirmation(level SafetyLevel) bool {
	return level >= SafetyMedium // Deleting requires confirmation at medium and high
}

// AppendToFileTool appends content to the end of a file
type AppendToFileTool struct {
	workDir string
}

// NewAppendToFileTool creates a new AppendToFileTool
func NewAppendToFileTool(workDir string) *AppendToFileTool {
	return &AppendToFileTool{workDir: workDir}
}

func (t *AppendToFileTool) Name() string {
	return "append_to_file"
}

func (t *AppendToFileTool) Description() string {
	return "Append content to the end of a file without reading the entire file first"
}

func (t *AppendToFileTool) Validate(params map[string]interface{}) error {
	path, ok := params["path"].(string)
	if !ok || path == "" {
		return fmt.Errorf("path parameter is required and must be a string")
	}

	if err := validatePath(t.workDir, path); err != nil {
		return err
	}

	content, ok := params["content"].(string)
	if !ok {
		return fmt.Errorf("content parameter is required and must be a string")
	}

	// Content can be empty string (valid use case)
	_ = content

	return nil
}

func (t *AppendToFileTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	path := params["path"].(string)
	content := params["content"].(string)

	absPath := resolvePath(t.workDir, path)

	// Open file in append mode, create if doesn't exist
	f, err := os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to open file: %v", err),
		}, err
	}
	defer func() { _ = f.Close() }()

	// Write content
	bytesWritten, err := f.WriteString(content)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to write to file: %v", err),
		}, err
	}

	return &ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Appended %d bytes to %s", bytesWritten, path),
		Data: map[string]interface{}{
			"bytes_written": bytesWritten,
			"path":          path,
		},
	}, nil
}

func (t *AppendToFileTool) RequiresConfirmation(level SafetyLevel) bool {
	return level >= SafetyMedium // Modifying files requires confirmation at medium and high
}

// FindFilesByPatternTool finds files matching advanced criteria
type FindFilesByPatternTool struct {
	workDir string
}

// NewFindFilesByPatternTool creates a new FindFilesByPatternTool
func NewFindFilesByPatternTool(workDir string) *FindFilesByPatternTool {
	return &FindFilesByPatternTool{workDir: workDir}
}

func (t *FindFilesByPatternTool) Name() string {
	return "find_files_by_pattern"
}

func (t *FindFilesByPatternTool) Description() string {
	return "Find files matching glob patterns with advanced filtering (size, extension, modification time)"
}

func (t *FindFilesByPatternTool) Validate(params map[string]interface{}) error {
	// Pattern is optional - defaults to searching all files
	if pattern, ok := params["pattern"].(string); ok && pattern != "" {
		// Validate it's not trying to escape workDir
		if strings.Contains(pattern, "..") {
			return fmt.Errorf("pattern cannot contain '..' (path traversal)")
		}
	}

	// Optional: max_size (in bytes)
	if maxSize, ok := params["max_size"].(float64); ok && maxSize < 0 {
		return fmt.Errorf("max_size must be non-negative")
	}

	// Optional: min_size (in bytes)
	if minSize, ok := params["min_size"].(float64); ok && minSize < 0 {
		return fmt.Errorf("min_size must be non-negative")
	}

	return nil
}

func (t *FindFilesByPatternTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	pattern := "*" // Default: all files
	if p, ok := params["pattern"].(string); ok && p != "" {
		pattern = p
	}

	// Optional filters
	var maxSize, minSize int64
	maxSize = -1 // -1 means no limit
	if ms, ok := params["max_size"].(float64); ok {
		maxSize = int64(ms)
	}
	if ms, ok := params["min_size"].(float64); ok {
		minSize = int64(ms)
	}

	extensions := []string{}
	if exts, ok := params["extensions"].([]interface{}); ok {
		for _, ext := range exts {
			if e, ok := ext.(string); ok {
				extensions = append(extensions, e)
			}
		}
	}

	var results []map[string]interface{}
	err := filepath.Walk(t.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Get relative path
		relPath, err := filepath.Rel(t.workDir, path)
		if err != nil {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check pattern match
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil || !matched {
			return nil
		}

		// Check size constraints
		if maxSize >= 0 && info.Size() > maxSize {
			return nil
		}
		if minSize > 0 && info.Size() < minSize {
			return nil
		}

		// Check extension filter
		if len(extensions) > 0 {
			ext := filepath.Ext(path)
			found := false
			for _, e := range extensions {
				if ext == e || ext == "."+e {
					found = true
					break
				}
			}
			if !found {
				return nil
			}
		}

		// Add to results
		results = append(results, map[string]interface{}{
			"path":     relPath,
			"size":     info.Size(),
			"modified": info.ModTime().Format("2006-01-02 15:04:05"),
			"is_dir":   info.IsDir(),
		})

		return nil
	})

	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to search files: %v", err),
		}, err
	}

	output := fmt.Sprintf("Found %d files matching criteria", len(results))
	if len(results) > 0 {
		output += "\n"
		for i, r := range results {
			if i >= 50 { // Limit output
				output += fmt.Sprintf("... and %d more files\n", len(results)-50)
				break
			}
			output += fmt.Sprintf("  %s (%d bytes)\n", r["path"], r["size"])
		}
	}

	return &ToolResult{
		Success: true,
		Output:  output,
		Data: map[string]interface{}{
			"files": results,
			"count": len(results),
		},
	}, nil
}

func (t *FindFilesByPatternTool) RequiresConfirmation(level SafetyLevel) bool {
	return false // Reading only, no confirmation needed
}
