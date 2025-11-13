package primitives

import (
	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// ScrollView provides viewport-based scrolling for content larger than visible area
type ScrollView struct {
	// Component metadata
	ID string

	// Content to scroll
	Content *vdom.Element

	// Scroll state
	ScrollY int // Vertical scroll offset (in lines)
	ScrollX int // Horizontal scroll offset (in columns)

	// Visible area (set by layout)
	Width  int
	Height int

	// Auto-scroll settings
	AutoScroll bool // Auto-scroll to bottom when content changes

	// Callbacks
	OnScroll func(x, y int)

	// Internal state
	contentHeight int // Total content height in lines
	contentWidth  int // Total content width in columns
}

// NewScrollView creates a new scroll view
func NewScrollView() *ScrollView {
	return &ScrollView{
		Width:      80,
		Height:     24,
		AutoScroll: false,
	}
}

// WithID sets the component ID
func (s *ScrollView) WithID(id string) *ScrollView {
	s.ID = id
	return s
}

// WithContent sets the content to display
func (s *ScrollView) WithContent(content *vdom.Element) *ScrollView {
	s.Content = content
	return s
}

// WithAutoScroll enables auto-scrolling to bottom
func (s *ScrollView) WithAutoScroll(enable bool) *ScrollView {
	s.AutoScroll = enable
	return s
}

// WithSize sets the visible viewport size
func (s *ScrollView) WithSize(width, height int) *ScrollView {
	s.Width = width
	s.Height = height
	return s
}

// WithOnScroll sets the scroll callback
func (s *ScrollView) WithOnScroll(callback func(int, int)) *ScrollView {
	s.OnScroll = callback
	return s
}

// GetScrollOffset returns the current scroll position (for buffer clipping)
func (s *ScrollView) GetScrollOffset() (int, int) {
	return s.ScrollX, s.ScrollY
}

// GetViewportSize returns the viewport dimensions
func (s *ScrollView) GetViewportSize() (int, int) {
	return s.Width, s.Height
}

// Render returns the scroll view element
// The actual clipping happens in the layout renderer via GetScrollOffset/GetViewportSize
func (s *ScrollView) Render() *vdom.Element {
	if s.Content == nil {
		return vdom.Text("")
	}

	// Return content wrapped in a box
	// The layout renderer will check for the ScrollViewInterface and apply clipping
	return vdom.Box(s.Content).WithID(s.ID)
}

// ScrollUp scrolls up by N lines
func (s *ScrollView) ScrollUp(lines int) {
	s.ScrollY -= lines
	if s.ScrollY < 0 {
		s.ScrollY = 0
	}
	s.emitScroll()
}

// ScrollDown scrolls down by N lines
func (s *ScrollView) ScrollDown(lines int) {
	s.ScrollY += lines
	maxScroll := s.contentHeight - s.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if s.ScrollY > maxScroll {
		s.ScrollY = maxScroll
	}
	s.emitScroll()
}

// ScrollLeft scrolls left by N columns
func (s *ScrollView) ScrollLeft(cols int) {
	s.ScrollX -= cols
	if s.ScrollX < 0 {
		s.ScrollX = 0
	}
	s.emitScroll()
}

// ScrollRight scrolls right by N columns
func (s *ScrollView) ScrollRight(cols int) {
	s.ScrollX += cols
	maxScroll := s.contentWidth - s.Width
	if maxScroll < 0 {
		maxScroll = 0
	}
	if s.ScrollX > maxScroll {
		s.ScrollX = maxScroll
	}
	s.emitScroll()
}

// ScrollToTop scrolls to the top
func (s *ScrollView) ScrollToTop() {
	s.ScrollY = 0
	s.emitScroll()
}

// ScrollToBottom scrolls to the bottom
func (s *ScrollView) ScrollToBottom() {
	maxScroll := s.contentHeight - s.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	s.ScrollY = maxScroll
	s.emitScroll()
}

// PageUp scrolls up one page
func (s *ScrollView) PageUp() {
	s.ScrollUp(s.Height)
}

// PageDown scrolls down one page
func (s *ScrollView) PageDown() {
	s.ScrollDown(s.Height)
}

// CanScrollUp returns true if can scroll up
func (s *ScrollView) CanScrollUp() bool {
	return s.ScrollY > 0
}

// CanScrollDown returns true if can scroll down
func (s *ScrollView) CanScrollDown() bool {
	return s.ScrollY < (s.contentHeight - s.Height)
}

// SetContentSize updates the content dimensions (called by layout)
func (s *ScrollView) SetContentSize(width, height int) {
	s.contentWidth = width
	s.contentHeight = height

	// Auto-scroll to bottom if enabled
	if s.AutoScroll {
		s.ScrollToBottom()
	}
}

// emitScroll triggers the OnScroll callback
func (s *ScrollView) emitScroll() {
	if s.OnScroll != nil {
		s.OnScroll(s.ScrollX, s.ScrollY)
	}
}

// HandleKey processes keyboard events for scrolling
func (s *ScrollView) HandleKey(event tty.KeyEvent) bool {
	// Check for arrow keys via escape sequences
	switch event.Code {
	case tty.SeqUp:
		s.ScrollUp(1)
		return true
	case tty.SeqDown:
		s.ScrollDown(1)
		return true
	case tty.SeqLeft:
		s.ScrollLeft(1)
		return true
	case tty.SeqRight:
		s.ScrollRight(1)
		return true
	case "[5~": // Page Up
		s.PageUp()
		return true
	case "[6~": // Page Down
		s.PageDown()
		return true
	case tty.SeqHome, tty.SeqHome2:
		s.ScrollToTop()
		return true
	case tty.SeqEnd, tty.SeqEnd2:
		s.ScrollToBottom()
		return true
	}
	return false
}

// Focusable interface implementation

// HandleKeyEvent implements Focusable interface
func (s *ScrollView) HandleKeyEvent(event tty.KeyEvent) bool {
	return s.HandleKey(event)
}

// IsFocusable implements Focusable interface
func (s *ScrollView) IsFocusable() bool {
	return true
}

// IsNode implements vdom.Node interface
func (s *ScrollView) IsNode() {}

// State persistence

// GetID returns the component ID
func (s *ScrollView) GetID() string {
	return s.ID
}

// SaveState saves scroll position
func (s *ScrollView) SaveState() map[string]interface{} {
	return map[string]interface{}{
		"scrollX": float64(s.ScrollX),
		"scrollY": float64(s.ScrollY),
	}
}

// LoadState restores scroll position
func (s *ScrollView) LoadState(state map[string]interface{}) error {
	if x, ok := state["scrollX"].(float64); ok {
		s.ScrollX = int(x)
	}
	if y, ok := state["scrollY"].(float64); ok {
		s.ScrollY = int(y)
	}
	return nil
}
