package lotusui

import (
	"fmt"

	"github.com/speier/smith/pkg/lotus/context"
	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// SelectOption represents a single option in the select
type SelectOption struct {
	Label    string
	Value    string
	Disabled bool
}

// Select is a dropdown selection component
type Select struct {
	// Component metadata
	ID string

	// Options
	Options []SelectOption

	// State
	Selected int  // Index of selected option
	Open     bool // Whether dropdown is open

	// Visual style
	Placeholder  string
	MaxHeight    int    // Max height of dropdown (in options)
	SelectedIcon string // Icon for selected option (default: "✓")
	ArrowIcon    string // Icon for dropdown arrow (default: "▼")

	// Behavior
	Disabled bool

	// Callbacks
	OnChange func(context.Context, int, string) // Called with (context, index, value) when selection changes

	// Internal
	focused          bool
	highlightedIndex int // Currently highlighted option in dropdown
}

// NewSelect creates a new select component
func NewSelect() *Select {
	return &Select{
		Placeholder:      "Select an option...",
		MaxHeight:        10,
		SelectedIcon:     "✓",
		ArrowIcon:        "▼",
		highlightedIndex: 0,
	}
}

// WithID sets the component ID
func (s *Select) WithID(id string) *Select {
	s.ID = id
	return s
}

// WithOptions sets the select options
func (s *Select) WithOptions(options []SelectOption) *Select {
	s.Options = options
	return s
}

// WithStringOptions creates options from string array
func (s *Select) WithStringOptions(labels []string) *Select {
	s.Options = make([]SelectOption, len(labels))
	for i, label := range labels {
		s.Options[i] = SelectOption{
			Label: label,
			Value: label,
		}
	}
	return s
}

// WithSelected sets the initially selected option
func (s *Select) WithSelected(index int) *Select {
	if index >= 0 && index < len(s.Options) {
		s.Selected = index
	}
	return s
}

// WithPlaceholder sets the placeholder text
func (s *Select) WithPlaceholder(placeholder string) *Select {
	s.Placeholder = placeholder
	return s
}

// WithMaxHeight sets the max dropdown height
func (s *Select) WithMaxHeight(height int) *Select {
	s.MaxHeight = height
	return s
}

// WithDisabled sets the disabled state
func (s *Select) WithDisabled(disabled bool) *Select {
	s.Disabled = disabled
	return s
}

// WithOnChange sets the change callback
func (s *Select) WithOnChange(callback func(context.Context, int, string)) *Select {
	s.OnChange = callback
	return s
}

// Render returns the select element
func (s *Select) Render() *vdom.Element {
	if s.Open {
		return s.renderOpen()
	}
	return s.renderClosed()
}

// renderClosed renders the select in closed state
func (s *Select) renderClosed() *vdom.Element {
	// Display selected option or placeholder
	displayText := s.Placeholder
	if s.Selected >= 0 && s.Selected < len(s.Options) {
		displayText = s.Options[s.Selected].Label
	}

	// Build display: [displayText] [arrow]
	text := fmt.Sprintf(" %s  %s ", displayText, s.ArrowIcon)

	elem := vdom.Text(text).
		WithBorderStyle(vdom.BorderStyleRounded)

	if s.Disabled {
		elem = elem.WithStyle("color", "#808080")
	} else if s.focused {
		elem = elem.WithStyle("color", "#00ff00")
	}

	return elem
}

// renderOpen renders the select with dropdown open
func (s *Select) renderOpen() *vdom.Element {
	// Build the closed part
	closedPart := s.renderClosed()

	// Build dropdown options
	dropdownOptions := s.renderDropdown()

	// Stack them vertically
	return vdom.VStack(
		closedPart,
		dropdownOptions,
	)
}

// renderDropdown renders the dropdown list
func (s *Select) renderDropdown() *vdom.Element {
	if len(s.Options) == 0 {
		return vdom.Box(vdom.Text(" No options ")).
			WithBorderStyle(vdom.BorderStyleRounded)
	}

	// Determine visible range (for scrolling large lists)
	start := 0
	end := len(s.Options)
	if end > s.MaxHeight {
		// Center highlighted option in view
		start = s.highlightedIndex - s.MaxHeight/2
		if start < 0 {
			start = 0
		}
		end = start + s.MaxHeight
		if end > len(s.Options) {
			end = len(s.Options)
			start = end - s.MaxHeight
			if start < 0 {
				start = 0
			}
		}
	}

	// Build option elements
	optionElements := make([]any, 0, end-start)
	for i := start; i < end; i++ {
		opt := s.Options[i]
		optionElements = append(optionElements, s.renderOption(i, opt))
	}

	return vdom.Box(vdom.VStack(optionElements...)).
		WithBorderStyle(vdom.BorderStyleRounded).
		WithStyle("background-color", "#1a1a1a")
}

// renderOption renders a single option
func (s *Select) renderOption(index int, opt SelectOption) *vdom.Element {
	// Build option text: [icon] label
	icon := "  "
	if index == s.Selected {
		icon = s.SelectedIcon + " "
	}

	text := fmt.Sprintf(" %s%s ", icon, opt.Label)

	elem := vdom.Text(text)

	// Styling
	if opt.Disabled {
		elem = elem.WithStyle("color", "#404040")
	} else if index == s.highlightedIndex {
		// Highlighted option
		elem = elem.WithStyle("color", "#000000").
			WithStyle("background-color", "#00ff00")
	} else if index == s.Selected {
		// Selected option
		elem = elem.WithStyle("color", "#00ff00")
	}

	return elem
}

// Toggle opens/closes the dropdown
func (s *Select) Toggle() {
	if s.Disabled {
		return
	}
	s.Open = !s.Open
	if s.Open {
		// Reset highlighted to selected when opening
		s.highlightedIndex = s.Selected
		if s.highlightedIndex < 0 {
			s.highlightedIndex = 0
		}
	}
}

// Close closes the dropdown
func (s *Select) Close() {
	s.Open = false
}

// SelectHighlighted selects the currently highlighted option
func (s *Select) SelectHighlighted(ctx context.Context) {
	if s.highlightedIndex >= 0 && s.highlightedIndex < len(s.Options) {
		if !s.Options[s.highlightedIndex].Disabled {
			if s.Selected != s.highlightedIndex {
				s.Selected = s.highlightedIndex
				s.emitChange(ctx)
			}
			s.Close()
		}
	}
}

// SelectOption sets the selected option by index
func (s *Select) SelectOption(index int) {
	if s.Disabled {
		return
	}
	if index < 0 || index >= len(s.Options) {
		return
	}
	if s.Options[index].Disabled {
		return
	}
	if s.Selected != index {
		s.Selected = index
	}
}

// HighlightNext moves highlight to next option
func (s *Select) HighlightNext() {
	if !s.Open || len(s.Options) == 0 {
		return
	}

	next := s.highlightedIndex + 1
	if next >= len(s.Options) {
		next = 0 // Wrap
	}

	// Skip disabled
	start := next
	for s.Options[next].Disabled {
		next++
		if next >= len(s.Options) {
			next = 0
		}
		if next == start {
			return // All disabled
		}
	}

	s.highlightedIndex = next
}

// HighlightPrevious moves highlight to previous option
func (s *Select) HighlightPrevious() {
	if !s.Open || len(s.Options) == 0 {
		return
	}

	prev := s.highlightedIndex - 1
	if prev < 0 {
		prev = len(s.Options) - 1 // Wrap
	}

	// Skip disabled
	start := prev
	for s.Options[prev].Disabled {
		prev--
		if prev < 0 {
			prev = len(s.Options) - 1
		}
		if prev == start {
			return // All disabled
		}
	}

	s.highlightedIndex = prev
}

// emitChange triggers the OnChange callback
func (s *Select) emitChange(ctx context.Context) {
	if s.OnChange != nil {
		value := ""
		if s.Selected >= 0 && s.Selected < len(s.Options) {
			value = s.Options[s.Selected].Value
		}
		s.OnChange(ctx, s.Selected, value)
	}
}

// HandleKey processes keyboard events
func (s *Select) HandleKey(event tty.KeyEvent) bool {
	return s.HandleKeyWithContext(context.Context{}, event)
}

// HandleKeyWithContext processes keyboard events with context
func (s *Select) HandleKeyWithContext(ctx context.Context, event tty.KeyEvent) bool {
	if s.Disabled {
		return false
	}

	// Space or Enter to toggle/select
	if event.Key == ' ' || event.IsEnter() {
		if s.Open {
			s.SelectHighlighted(ctx)
		} else {
			s.Toggle()
		}
		return true
	}

	// Escape to close
	if event.Key == tty.KeyEscape {
		if s.Open {
			s.Close()
			return true
		}
	}

	// Arrow keys
	switch event.Code {
	case tty.SeqUp:
		if s.Open {
			s.HighlightPrevious()
		}
		return true
	case tty.SeqDown:
		if !s.Open {
			s.Toggle()
		} else {
			s.HighlightNext()
		}
		return true
	}

	return false
}

// Focusable interface implementation

// HandleKeyEvent implements Focusable interface
func (s *Select) HandleKeyEvent(event tty.KeyEvent) bool {
	return s.HandleKey(event)
}

// IsFocusable implements Focusable interface
func (s *Select) IsFocusable() bool {
	return !s.Disabled
}

// SetFocus sets the focus state
func (s *Select) SetFocus(focused bool) {
	s.focused = focused
	if !focused {
		s.Close() // Close dropdown when losing focus
	}
}

// IsNode implements vdom.Node interface
func (s *Select) IsNode() {}

// State persistence

// GetID returns the component ID
func (s *Select) GetID() string {
	return s.ID
}

// SaveState saves the selected index
func (s *Select) SaveState() map[string]interface{} {
	return map[string]interface{}{
		"selected": float64(s.Selected),
	}
}

// LoadState restores the selected index
func (s *Select) LoadState(state map[string]interface{}) error {
	if selected, ok := state["selected"].(float64); ok {
		s.Selected = int(selected)
	}
	return nil
}
