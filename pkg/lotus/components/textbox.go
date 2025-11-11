package components

import (
	"strings"

	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// TextBox is a convenience wrapper around ScrollView for displaying text lines
// It provides a simple API for appending lines while using ScrollView for scrolling
type TextBox struct {
	ID string // Component ID

	// Content
	Lines []string // Lines of text to display

	// Scrolling (delegated to ScrollView)
	scrollView *ScrollView

	// Dimensions (set by layout or manually)
	Width  int // Visible width (0 = auto)
	Height int // Visible height (0 = auto)

	// Appearance
	WordWrap bool // Whether to wrap long lines (not implemented yet)

	// Focus
	Focusable bool // Whether component can receive keyboard focus (for scrolling)
}

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
		Focusable: true, // Focusable by default for scrolling
	}
	
	// Create internal ScrollView
	tb.scrollView = NewScrollView().WithID(boxID + "-scroll")
	
	return tb
}

// SetContent replaces all content with new lines
func (tb *TextBox) SetContent(lines []string) {
	tb.Lines = lines
}

// AppendLine adds a line to the end
func (tb *TextBox) AppendLine(line string) {
	tb.Lines = append(tb.Lines, line)
}

// AppendLines adds multiple lines to the end
func (tb *TextBox) AppendLines(lines []string) {
	tb.Lines = append(tb.Lines, lines...)
}

// Clear removes all content
func (tb *TextBox) Clear() {
	tb.Lines = []string{}
}

// Scrolling methods - delegate to ScrollView
func (tb *TextBox) ScrollUp() {
	tb.scrollView.ScrollUp(1)
}

func (tb *TextBox) ScrollDown() {
	tb.scrollView.ScrollDown(1)
}

func (tb *TextBox) ScrollPageUp() {
	tb.scrollView.PageUp()
}

func (tb *TextBox) ScrollPageDown() {
	tb.scrollView.PageDown()
}

func (tb *TextBox) ScrollToTop() {
	tb.scrollView.ScrollToTop()
}

func (tb *TextBox) ScrollToBottom() {
	tb.scrollView.ScrollToBottom()
}

// Render generates the Element for the text box
func (tb *TextBox) Render() *vdom.Element {
	if len(tb.Lines) == 0 {
		return vdom.Box(vdom.Text(""))
	}

	// Build content from lines
	elements := make([]any, len(tb.Lines))
	for i, line := range tb.Lines {
		elements[i] = vdom.Text(line)
	}
	content := vdom.VStack(elements...)

	// Update ScrollView settings
	if tb.Width > 0 {
		tb.scrollView.Width = tb.Width
	}
	if tb.Height > 0 {
		tb.scrollView.Height = tb.Height
	}

	// Use ScrollView for rendering with scrolling
	tb.scrollView.Content = content
	
	return tb.scrollView.Render()
}

// --- Focusable interface (for keyboard scrolling) ---

// HandleKeyEvent handles keyboard events - delegates to ScrollView
func (tb *TextBox) HandleKeyEvent(event tty.KeyEvent) bool {
	// Delegate to ScrollView for consistent keyboard shortcuts
	return tb.scrollView.HandleKeyEvent(event)
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
