package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SearchMatch represents a single search result
type SearchMatch struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Match   string `json:"match"`
	Context string `json:"context"` // The full line containing the match
}

// SearchFilesTool searches for text patterns in files
type SearchFilesTool struct {
	workDir string
}

// NewSearchFilesTool creates a new SearchFilesTool
func NewSearchFilesTool(workDir string) *SearchFilesTool {
	return &SearchFilesTool{workDir: workDir}
}

func (t *SearchFilesTool) Name() string {
	return "search_files"
}

func (t *SearchFilesTool) Description() string {
	return "Search for text patterns in files (supports regex)"
}

func (t *SearchFilesTool) Validate(params map[string]interface{}) error {
	pattern, ok := params["pattern"].(string)
	if !ok || pattern == "" {
		return fmt.Errorf("pattern parameter is required and must be a non-empty string")
	}

	// If is_regex is true, validate the regex pattern
	if isRegex, ok := params["is_regex"].(bool); ok && isRegex {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid regex pattern: %v", err)
		}
	}

	// Validate path if provided
	if path, ok := params["path"].(string); ok && path != "" {
		if err := validatePath(t.workDir, path); err != nil {
			return err
		}
	}

	return nil
}

func (t *SearchFilesTool) Execute(ctx context.Context, params map[string]interface{}) (*ToolResult, error) {
	pattern, ok := params["pattern"].(string)
	if !ok || pattern == "" {
		return &ToolResult{
			Success: false,
			Error:   "pattern parameter is required",
		}, fmt.Errorf("pattern parameter is required")
	}

	// Optional: path to search in (defaults to workDir)
	searchPath := "."
	if path, ok := params["path"].(string); ok && path != "" {
		searchPath = path
	}

	// Optional: regex flag (defaults to false - plain text search)
	isRegex := false
	if regex, ok := params["is_regex"].(bool); ok {
		isRegex = regex
	}

	// Optional: case sensitive flag (defaults to true)
	caseSensitive := true
	if cs, ok := params["case_sensitive"].(bool); ok {
		caseSensitive = cs
	}

	// Optional: max results (defaults to 100)
	maxResults := 100
	if max, ok := params["max_results"].(float64); ok {
		maxResults = int(max)
	} else if max, ok := params["max_results"].(int); ok {
		maxResults = max
	}

	absPath := resolvePath(t.workDir, searchPath)

	// Compile regex if needed
	var re *regexp.Regexp
	var err error
	if isRegex {
		flags := ""
		if !caseSensitive {
			flags = "(?i)"
		}
		re, err = regexp.Compile(flags + pattern)
		if err != nil {
			return &ToolResult{
				Success: false,
				Error:   fmt.Sprintf("invalid regex: %v", err),
			}, err
		}
	}

	// Search for matches
	matches := make([]SearchMatch, 0)
	searchFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't read
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip binary files (simple heuristic: check extension)
		if isBinaryFile(path) {
			return nil
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Search in this file
		fileMatches, err := searchInFile(path, pattern, re, caseSensitive, isRegex)
		if err != nil {
			return nil // Skip files with errors
		}

		matches = append(matches, fileMatches...)

		// Stop if we've reached max results
		if len(matches) >= maxResults {
			return filepath.SkipAll
		}

		return nil
	}

	// Walk the directory tree
	if err := filepath.Walk(absPath, searchFunc); err != nil && err != filepath.SkipAll {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("search failed: %v", err),
		}, err
	}

	// Limit results
	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}

	return &ToolResult{
		Success: true,
		Output:  fmt.Sprintf("Found %d matches", len(matches)),
		Data: map[string]interface{}{
			"matches":     matches,
			"count":       len(matches),
			"search_path": absPath,
			"pattern":     pattern,
		},
	}, nil
}

func (t *SearchFilesTool) RequiresConfirmation(level SafetyLevel) bool {
	return false // Searching is safe at all levels
}

// Helper functions

func searchInFile(path string, pattern string, re *regexp.Regexp, caseSensitive bool, isRegex bool) ([]SearchMatch, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	matches := make([]SearchMatch, 0)
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		var matched bool
		var matchStart int
		var matchText string

		if isRegex {
			// Regex search
			loc := re.FindStringIndex(line)
			if loc != nil {
				matched = true
				matchStart = loc[0]
				matchText = line[loc[0]:loc[1]]
			}
		} else {
			// Plain text search
			searchLine := line
			searchPattern := pattern
			if !caseSensitive {
				searchLine = strings.ToLower(line)
				searchPattern = strings.ToLower(pattern)
			}

			idx := strings.Index(searchLine, searchPattern)
			if idx != -1 {
				matched = true
				matchStart = idx
				matchText = line[idx : idx+len(pattern)]
			}
		}

		if matched {
			matches = append(matches, SearchMatch{
				File:    path,
				Line:    lineNum,
				Column:  matchStart + 1, // 1-indexed
				Match:   matchText,
				Context: strings.TrimSpace(line),
			})
		}
	}

	return matches, scanner.Err()
}

func isBinaryFile(path string) bool {
	// Simple heuristic: check file extension
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := map[string]bool{
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".bin": true, ".dat": true, ".db": true, ".sqlite": true,
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".pdf": true, ".zip": true, ".tar": true, ".gz": true,
		".mp3": true, ".mp4": true, ".avi": true, ".mov": true,
		".o": true, ".a": true, ".pyc": true, ".class": true,
	}
	return binaryExts[ext]
}
