package render

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// DiffRenderer implements differential rendering with synchronized output
// Three-strategy rendering approach:
// 1. First render: output all lines without clearing scrollback
// 2. Width changed or change above viewport: clear screen and full re-render
// 3. Normal update: move cursor to first changed line, clear to end, render changed lines
type DiffRenderer struct {
	previousLines  []string
	previousWidth  int
	previousHeight int
	cursorRow      int // Track where cursor is (0-indexed)
	cache          *RenderCache

	// Debug stats
	FullRenders    int
	PartialRenders int
	SkippedRenders int
}

// RenderCache stores cached component renders
type RenderCache struct {
	cache map[string]CachedRender
}

// CachedRender stores a cached render result
type CachedRender struct {
	Hash   string
	Output string
}

// NewDiffRenderer creates a new differential renderer
func NewDiffRenderer() *DiffRenderer {
	return &DiffRenderer{
		previousLines:  make([]string, 0),
		previousWidth:  0,
		previousHeight: 0,
		cursorRow:      0,
		cache:          NewRenderCache(),
	}
}

// NewRenderCache creates a new render cache
func NewRenderCache() *RenderCache {
	return &RenderCache{
		cache: make(map[string]CachedRender),
	}
}

// GetCached returns cached render if available
func (rc *RenderCache) GetCached(id string, contentHash string) (string, bool) {
	if cached, ok := rc.cache[id]; ok {
		if cached.Hash == contentHash {
			return cached.Output, true
		}
	}
	return "", false
}

// SetCached stores a render in the cache
func (rc *RenderCache) SetCached(id string, contentHash string, output string) {
	rc.cache[id] = CachedRender{
		Hash:   contentHash,
		Output: output,
	}
}

// Clear clears the cache
func (rc *RenderCache) Clear() {
	rc.cache = make(map[string]CachedRender)
}

// HashContent creates a hash of content for cache invalidation
func HashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes for compact hash
}

// RenderDiff performs differential rendering
// Returns the ANSI output needed to update the screen
func (dr *DiffRenderer) RenderDiff(content string, width, height int) string {
	newLines := strings.Split(content, "\n")

	// Trim trailing empty lines to avoid unnecessary scrolling
	for len(newLines) > 0 && strings.TrimSpace(newLines[len(newLines)-1]) == "" {
		newLines = newLines[:len(newLines)-1]
	}
	if len(newLines) == 0 {
		newLines = []string{""}
	}

	// First render - output everything without clearing scrollback
	// Wrapped in synchronized output for flicker-free rendering
	if len(dr.previousLines) == 0 {
		dr.FullRenders++
		buffer := &strings.Builder{}
		buffer.WriteString("\x1b[?2026h") // Begin synchronized output (CSI 2026)
		for i, line := range newLines {
			if i > 0 {
				buffer.WriteString("\r\n")
			}
			buffer.WriteString(line)
		}
		buffer.WriteString("\x1b[?2026l") // End synchronized output
		dr.previousLines = newLines
		dr.previousWidth = width
		dr.previousHeight = height
		dr.cursorRow = len(newLines) - 1
		return buffer.String()
	}

	// Width changed - full re-render (strategy 2)
	if dr.previousWidth != width {
		dr.FullRenders++
		buffer := &strings.Builder{}
		buffer.WriteString("\x1b[?2026h")          // Begin synchronized output
		buffer.WriteString("\x1b[3J\x1b[2J\x1b[H") // Clear scrollback, screen, and home
		for i, line := range newLines {
			if i > 0 {
				buffer.WriteString("\r\n")
			}
			buffer.WriteString(line)
		}
		buffer.WriteString("\x1b[?2026l") // End synchronized output
		dr.previousLines = newLines
		dr.previousWidth = width
		dr.previousHeight = height
		dr.cursorRow = len(newLines) - 1
		return buffer.String()
	}

	// Find first changed line
	firstChanged := -1

	maxLines := max(len(newLines), len(dr.previousLines))
	for i := 0; i < maxLines; i++ {
		oldLine := ""
		if i < len(dr.previousLines) {
			oldLine = dr.previousLines[i]
		}
		newLine := ""
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine != newLine {
			if firstChanged == -1 {
				firstChanged = i
				break // We only need the first changed line
			}
		}
	}

	// No changes - return empty string
	if firstChanged == -1 {
		dr.SkippedRenders++
		return ""
	}

	// Check if first changed line is above the viewport
	// If cursor is at cursorRow, viewport shows lines [cursorRow - height + 1, cursorRow]
	viewportTop := dr.cursorRow - height + 1
	if viewportTop < 0 {
		viewportTop = 0
	}

	// If change is above viewport, do full re-render (strategy 2)
	if firstChanged < viewportTop {
		dr.FullRenders++
		buffer := &strings.Builder{}
		buffer.WriteString("\x1b[?2026h")          // Begin synchronized output
		buffer.WriteString("\x1b[3J\x1b[2J\x1b[H") // Clear scrollback, screen, and home
		for i, line := range newLines {
			if i > 0 {
				buffer.WriteString("\r\n")
			}
			buffer.WriteString(line)
		}
		buffer.WriteString("\x1b[?2026l") // End synchronized output
		dr.previousLines = newLines
		dr.previousWidth = width
		dr.previousHeight = height
		dr.cursorRow = len(newLines) - 1
		return buffer.String()
	}

	// Normal update - differential rendering (strategy 3)
	dr.PartialRenders++
	buffer := &strings.Builder{}
	buffer.WriteString("\x1b[?2026h") // Begin synchronized output

	// Move cursor to first changed line
	lineDiff := firstChanged - dr.cursorRow
	if lineDiff > 0 {
		fmt.Fprintf(buffer, "\x1b[%dB", lineDiff) // Move down
	} else if lineDiff < 0 {
		fmt.Fprintf(buffer, "\x1b[%dA", -lineDiff) // Move up
	}

	buffer.WriteString("\r")     // Move to column 0
	buffer.WriteString("\x1b[J") // Clear from cursor to end of screen

	// Render from first changed line to end
	for i := firstChanged; i < len(newLines); i++ {
		if i > firstChanged {
			buffer.WriteString("\r\n")
		}
		buffer.WriteString(newLines[i])
	}

	buffer.WriteString("\x1b[?2026l") // End synchronized output

	// Update state
	dr.cursorRow = len(newLines) - 1
	dr.previousLines = newLines
	dr.previousWidth = width
	dr.previousHeight = height

	return buffer.String()
}

// Reset resets the differential renderer state
func (dr *DiffRenderer) Reset() {
	dr.previousLines = make([]string, 0)
	dr.previousWidth = 0
	dr.previousHeight = 0
	dr.cursorRow = 0
	dr.cache.Clear()
}

// GetStats returns rendering statistics for debugging
func (dr *DiffRenderer) GetStats() (full, partial, skipped int) {
	return dr.FullRenders, dr.PartialRenders, dr.SkippedRenders
}
