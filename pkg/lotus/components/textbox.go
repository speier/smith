package components

import (
	"strings"

	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// TextBox is a scrollable, multi-line text viewer (read-only)
// This is a minimal primitive - just displays lines with scrolling
type TextBox struct {
	ID string // Component ID

	// Content
	Lines []string // Lines of text to display

	// Dimensions (set by layout or manually)
	Width  int // Visible width (0 = auto)
	Height int // Visible height (0 = auto)

	// Scrolling state
	ScrollY    int  // Vertical scroll position (line number)
	ScrollX    int  // Horizontal scroll position (for long lines)
	AutoScroll bool // Auto-scroll to bottom when new lines added

	// Appearance
	WordWrap bool // Whether to wrap long lines (not implemented yet)

	// Focus
	Focusable bool // Whether component can receive keyboard focus (for scrolling)

	// Mouse support
	MouseEnabled bool
}

// NewTextBox creates a new scrollable text box
func NewTextBox(id ...string) *TextBox {
	boxID := ""
	if len(id) > 0 {
		boxID = id[0]
	}
	return &TextBox{
		ID:           boxID,
		Lines:        []string{},
		Width:        0,
		Height:       0,
		ScrollY:      0,
		ScrollX:      0,
		AutoScroll:   false,
		WordWrap:     false,
		Focusable:    false, // Not focusable by default (read-only viewer)
		MouseEnabled: true,
	}
}

// SetContent replaces all content with new lines
func (tb *TextBox) SetContent(lines []string) {
	tb.Lines = lines
	if tb.AutoScroll {
		tb.ScrollToBottom()
	}
}

// AppendLine adds a line to the end
func (tb *TextBox) AppendLine(line string) {
	tb.Lines = append(tb.Lines, line)
	if tb.AutoScroll {
		tb.ScrollToBottom()
	}
}

// AppendLines adds multiple lines to the end
func (tb *TextBox) AppendLines(lines []string) {
	tb.Lines = append(tb.Lines, lines...)
	if tb.AutoScroll {
		tb.ScrollToBottom()
	}
}

// Clear removes all content
func (tb *TextBox) Clear() {
	tb.Lines = []string{}
	tb.ScrollY = 0
	tb.ScrollX = 0
}

// ScrollUp scrolls up one line
func (tb *TextBox) ScrollUp() {
	if tb.ScrollY > 0 {
		tb.ScrollY--
	}
}

// ScrollDown scrolls down one line
func (tb *TextBox) ScrollDown() {
	maxScroll := tb.getMaxScrollY()
	if tb.ScrollY < maxScroll {
		tb.ScrollY++
	}
}

// ScrollPageUp scrolls up one page
func (tb *TextBox) ScrollPageUp() {
	pageSize := tb.Height
	if pageSize < 1 {
		pageSize = 10
	}
	tb.ScrollY -= pageSize
	if tb.ScrollY < 0 {
		tb.ScrollY = 0
	}
}

// ScrollPageDown scrolls down one page
func (tb *TextBox) ScrollPageDown() {
	pageSize := tb.Height
	if pageSize < 1 {
		pageSize = 10
	}
	tb.ScrollY += pageSize
	maxScroll := tb.getMaxScrollY()
	if tb.ScrollY > maxScroll {
		tb.ScrollY = maxScroll
	}
}

// ScrollToTop scrolls to the beginning
func (tb *TextBox) ScrollToTop() {
	tb.ScrollY = 0
}

// ScrollToBottom scrolls to the end
func (tb *TextBox) ScrollToBottom() {
	tb.ScrollY = tb.getMaxScrollY()
}

// getMaxScrollY calculates the maximum vertical scroll position
func (tb *TextBox) getMaxScrollY() int {
	height := tb.Height
	if height < 1 {
		height = 10 // Default height
	}

	maxScroll := len(tb.Lines) - height
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// GetVisibleLines returns the lines currently visible in the viewport
// This is 100% testable without rendering!
func (tb *TextBox) GetVisibleLines() []string {
	if len(tb.Lines) == 0 {
		return []string{}
	}

	height := tb.Height
	if height < 1 {
		height = 10 // Default
	}

	// Clamp scroll position
	maxScroll := tb.getMaxScrollY()
	if tb.ScrollY > maxScroll {
		tb.ScrollY = maxScroll
	}
	if tb.ScrollY < 0 {
		tb.ScrollY = 0
	}

	// Calculate visible range
	start := tb.ScrollY
	end := tb.ScrollY + height
	if end > len(tb.Lines) {
		end = len(tb.Lines)
	}

	visible := tb.Lines[start:end]

	// Apply horizontal scroll if needed
	if tb.ScrollX > 0 {
		scrolled := make([]string, len(visible))
		for i, line := range visible {
			if tb.ScrollX < len(line) {
				scrolled[i] = line[tb.ScrollX:]
			} else {
				scrolled[i] = ""
			}
		}
		return scrolled
	}

	return visible
}

// Render generates the Element for the text box
func (tb *TextBox) Render() *vdom.Element {
	visibleLines := tb.GetVisibleLines()

	if len(visibleLines) == 0 {
		return vdom.Box(vdom.Text(""))
	}

	// Build element for each line
	elements := make([]any, len(visibleLines))
	for i, line := range visibleLines {
		// Trim/pad to width if specified
		displayLine := line
		if tb.Width > 0 && len(line) > tb.Width {
			displayLine = line[:tb.Width]
		}
		elements[i] = vdom.Box(vdom.Text(displayLine))
	}

	return vdom.VStack(elements...).WithID(tb.ID)
}

// --- Focusable interface (for keyboard scrolling) ---

// HandleKeyEvent handles keyboard events for scrolling
func (tb *TextBox) HandleKeyEvent(event tty.KeyEvent) bool {
	switch event.Code {
	case tty.SeqUp:
		tb.ScrollUp()
		return true
	case tty.SeqDown:
		tb.ScrollDown()
		return true
	case tty.SeqHome:
		tb.ScrollToTop()
		return true
	case tty.SeqEnd:
		tb.ScrollToBottom()
		return true
	}

	// Also support vi-style navigation
	if event.IsPrintable() {
		switch event.Char {
		case "k": // Up
			tb.ScrollUp()
			return true
		case "j": // Down
			tb.ScrollDown()
			return true
		case "u": // Page up
			tb.ScrollPageUp()
			return true
		case "d": // Page down
			tb.ScrollPageDown()
			return true
		case "g": // Top
			tb.ScrollToTop()
			return true
		case "G": // Bottom
			tb.ScrollToBottom()
			return true
		}
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
	tb.AutoScroll = enabled
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
