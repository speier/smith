package ui

import (
	"strings"
	"sync"

	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// LineRenderer is a custom renderer for transforming text before display
// Useful for markdown rendering (glamour), syntax highlighting, etc.
type LineRenderer func(text string) string

// Deprecated: Use Box with overflow:auto instead (via WithFlexGrow).
// TextBox is a convenience wrapper for displaying text lines with scrolling.
type TextBox struct {
	ID string // Component ID

	// Content
	Lines []string   // Lines of text to display
	mu    sync.Mutex // Protects Lines for concurrent access (streaming)

	// Rendering
	Renderer LineRenderer // Optional custom renderer (e.g., glamour for markdown)

	// Dimensions (set by layout or manually)
	Width  int // Visible width (0 = auto)
	Height int // Visible height (0 = auto)

	// Appearance
	WordWrap bool // Whether to wrap long lines (not implemented yet)

	// Focus
	Focusable bool // Whether component can receive keyboard focus (for scrolling)

	// Streaming state
	streamBuffer string // Buffer for partial line (streaming mode)

	// Scroll state (for manual scrolling)
	scrollOffset int
}

// Deprecated: Use Box(VStack(lines...)).WithFlexGrow(1) for auto overflow scrolling.
// NewTextBox creates a new scrollable text box
func NewTextBox(id ...string) *TextBox {
	boxID := ""
	if len(id) > 0 {
		boxID = id[0]
	}

	tb := &TextBox{
		ID:        boxID,
		Lines:     []string{},
		Width:     0,
		Height:    0,
		WordWrap:  false,
		Focusable: false, // Not focusable by default (read-only display)
	}

	return tb
}

// SetContent replaces all content with new lines
func (tb *TextBox) SetContent(lines []string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.Lines = lines
}

// AppendLine adds a line to the end (thread-safe for streaming)
func (tb *TextBox) AppendLine(line string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.Lines = append(tb.Lines, line)
}

// AppendLines adds multiple lines to the end (thread-safe)
func (tb *TextBox) AppendLines(lines []string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.Lines = append(tb.Lines, lines...)
}

// AppendText appends text chunk (for streaming)
// Automatically handles newlines - completes current line or adds new lines
// Thread-safe for concurrent streaming from goroutines
func (tb *TextBox) AppendText(text string) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Append to stream buffer
	tb.streamBuffer += text

	// Process complete lines (split by \n)
	for {
		idx := strings.Index(tb.streamBuffer, "\n")
		if idx == -1 {
			break // No complete line yet
		}

		// Extract complete line
		line := tb.streamBuffer[:idx]
		tb.streamBuffer = tb.streamBuffer[idx+1:]

		// Add to lines
		tb.Lines = append(tb.Lines, line)
	}

	// If buffer has content without newline, update last line
	if len(tb.streamBuffer) > 0 && len(tb.Lines) > 0 {
		// Update last line with partial content (for live streaming effect)
		lastIdx := len(tb.Lines) - 1
		tb.Lines[lastIdx] = tb.Lines[lastIdx] + tb.streamBuffer
		tb.streamBuffer = ""
	} else if len(tb.streamBuffer) > 0 {
		// First chunk without newline - start new line
		tb.Lines = append(tb.Lines, tb.streamBuffer)
		tb.streamBuffer = ""
	}
}

// FlushStream completes the current streaming line
func (tb *TextBox) FlushStream() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.streamBuffer != "" {
		tb.Lines = append(tb.Lines, tb.streamBuffer)
		tb.streamBuffer = ""
	}
}

// Clear removes all content
func (tb *TextBox) Clear() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.Lines = []string{}
	tb.streamBuffer = ""
}

// Scrolling methods - no-op (deprecated - use Box with overflow:auto)
func (tb *TextBox) ScrollUp() {
	if tb.scrollOffset > 0 {
		tb.scrollOffset--
	}
}

func (tb *TextBox) ScrollDown() {
	tb.scrollOffset++
}

func (tb *TextBox) ScrollPageUp() {
	tb.scrollOffset -= tb.Height
	if tb.scrollOffset < 0 {
		tb.scrollOffset = 0
	}
}

func (tb *TextBox) ScrollPageDown() {
	tb.scrollOffset += tb.Height
}

func (tb *TextBox) ScrollToTop() {
	tb.scrollOffset = 0
}

func (tb *TextBox) ScrollToBottom() {
	tb.scrollOffset = len(tb.Lines)
}

// Render generates the Element for the text box
func (tb *TextBox) Render() *vdom.Element {
	tb.mu.Lock()
	lines := make([]string, len(tb.Lines))
	copy(lines, tb.Lines)
	tb.mu.Unlock()

	if len(lines) == 0 {
		return vdom.Box(vdom.Text(""))
	}

	// Build content from lines (apply custom renderer if set)
	elements := make([]any, len(lines))
	for i, line := range lines {
		displayLine := line
		if tb.Renderer != nil {
			displayLine = tb.Renderer(line)
		}
		elements[i] = vdom.Text(displayLine)
	}

	// Return VStack with overflow:auto if dimensions are set
	result := vdom.VStack(elements...)
	if tb.Height > 0 {
		return vdom.Box(result).WithFlexGrow(1) // Auto overflow:auto
	}
	return result
}

// --- Focusable interface (for keyboard scrolling) ---

// HandleKeyEvent handles keyboard events
func (tb *TextBox) HandleKeyEvent(event tty.KeyEvent) bool {
	// Simple arrow key scrolling
	switch event.Code {
	case "up":
		tb.ScrollUp()
		return true
	case "down":
		tb.ScrollDown()
		return true
	}
	return false
}

// GetCursorOffset returns 0 (read-only component has no cursor)
func (tb *TextBox) GetCursorOffset() int {
	return 0
}

// IsFocusable returns true if component can receive keyboard focus
func (tb *TextBox) IsFocusable() bool {
	return tb.Focusable
}

// IsNode implements vdom.Node interface
func (tb *TextBox) IsNode() {}

// --- Fluent API ---

// WithLines sets initial content and returns the component
func (tb *TextBox) WithLines(lines []string) *TextBox {
	tb.Lines = lines
	return tb
}

// WithContent sets initial content from a string (split by newlines)
func (tb *TextBox) WithContent(content string) *TextBox {
	tb.Lines = strings.Split(content, "\n")
	return tb
}

// WithAutoScroll enables auto-scrolling to bottom
func (tb *TextBox) WithAutoScroll(enabled bool) *TextBox {
	tb.scrollView.AutoScroll = enabled
	return tb
}

// WithFocusable enables keyboard focus (for scrolling with j/k/g/u/d keys)
func (tb *TextBox) WithFocusable(enabled bool) *TextBox {
	tb.Focusable = enabled
	return tb
}

// WithHeight sets the viewport height
func (tb *TextBox) WithHeight(height int) *TextBox {
	tb.Height = height
	return tb
}

// WithWidth sets the viewport width
func (tb *TextBox) WithWidth(width int) *TextBox {
	tb.Width = width
	return tb
}

// WithRenderer sets a custom line renderer (e.g., glamour for markdown)
func (tb *TextBox) WithRenderer(renderer LineRenderer) *TextBox {
	tb.Renderer = renderer
	return tb
}
