package primitives

import (
	"time"

	"github.com/speier/smith/pkg/lotus/vdom"
)

// InputType defines the type of input (like HTML input type attribute)
type InputType string

const (
	InputTypeText     InputType = "text"     // Regular text input (default)
	InputTypePassword InputType = "password" // Masked password input (shows *)
	InputTypeNumber   InputType = "number"   // Numeric input only
)

// CursorStyle defines the visual style of the cursor
type CursorStyle int

const (
	// CursorBlock is a block cursor (█)
	CursorBlock CursorStyle = iota
	// CursorUnderline is an underline cursor (_)
	CursorUnderline
	// CursorBar is a bar cursor (|)
	CursorBar
)

// InputProps represents the props (configuration from parent) for Input
// React pattern: Props are immutable configuration passed by parent
type InputProps struct {
	Placeholder string
	Width       int
	Disabled    bool
	Type        InputType // Input type (text, password, number)
	OnChange    any       // func(string) or func(Context, string)
	OnSubmit    any       // func(string) or func(Context, string)
}

// Input is a single-line text input field (like HTML <input>)
// For multi-line editing, use TextArea
type Input struct {
	// Component metadata
	ID string // Component ID for registration (React key) - MUST be set for state preservation

	// Input type
	Type InputType // Input type (text, password, number)

	// Internal state (private to component)
	Value      string // Current input text (single line)
	CursorPos  int    // Cursor position in the text (0-indexed)
	Scroll     int    // Horizontal scroll offset (for single-line scrolling)
	desiredCol int    // Desired column for vertical navigation (preserved across up/down)
	Focused    bool   // Whether this component has focus (set by runtime)

	// Cursor state (testable!)
	CursorStyle   CursorStyle   // Visual style of cursor
	CursorBlink   bool          // Whether cursor is blinking
	CursorVisible bool          // Current visibility state (for blink cycle)
	BlinkInterval time.Duration // How fast cursor blinks (0 = no blink)
	lastBlinkTime time.Time     // Last blink toggle time

	// Props (configuration from parent)
	Width       int    // Visible width (set automatically from layout)
	Placeholder string // Placeholder text when empty
	Disabled    bool   // If true, component cannot receive focus

	// Event callbacks (React-like props)
	// Supports both func(string) and func(Context, string) signatures
	OnChange any // Called when text changes
	OnSubmit any // Called when Enter is pressed
}

// NewInput creates a new Input with optional ID
// React-like: Input() creates component instance
func NewInput(id ...string) *Input {
	inputID := ""
	if len(id) > 0 {
		inputID = id[0]
	}
	return &Input{
		ID:            inputID,
		Type:          InputTypeText, // Default to text input
		Value:         "",
		CursorPos:     0,
		Scroll:        0,
		Width:         50,    // Default width
		Focused:       false, // Will be set by focus manager
		CursorStyle:   CursorBlock,
		CursorBlink:   true,
		CursorVisible: true,
		BlinkInterval: 500 * time.Millisecond,
		lastBlinkTime: time.Now(),
	}
}

// CreateInput creates a new single-line input with simplified API (like pi-tui)
// Usage: CreateInput(placeholder, onSubmit)
// onSubmit can be func(string) or func(Context, string)
func CreateInput(placeholder string, onSubmit any) *Input {
	return NewInput().
		WithPlaceholder(placeholder).
		WithOnSubmit(onSubmit)
}

// SetProps updates the component props (React pattern)
// Allows parent to configure component behavior
func (t *Input) SetProps(props InputProps) {
	t.Placeholder = props.Placeholder
	if props.Width > 0 {
		t.Width = props.Width
	}
	t.Disabled = props.Disabled
	if props.Type != "" {
		t.Type = props.Type
	}
	t.OnChange = props.OnChange
	t.OnSubmit = props.OnSubmit
}

// Fluent API methods (React-like chaining)

// WithPlaceholder sets the placeholder text and returns the component for chaining
func (t *Input) WithPlaceholder(placeholder string) *Input {
	t.Placeholder = placeholder
	return t
}

// WithWidth sets the width and returns the component for chaining
func (t *Input) WithWidth(width int) *Input {
	t.Width = width
	return t
}

// WithOnChange sets the onChange callback and returns the component for chaining
func (t *Input) WithOnChange(onChange func(string)) *Input {
	t.OnChange = onChange
	return t
}

// WithOnSubmit sets the onSubmit callback and returns the component for chaining
// Supports both func(string) and func(Context, string) signatures
func (t *Input) WithOnSubmit(onSubmit any) *Input {
	t.OnSubmit = onSubmit
	return t
}

// UpdateProps updates callbacks and props from a new component instance
// This is called during reconciliation to keep callbacks fresh while preserving state
func (t *Input) UpdateProps(newComponent vdom.Component) {
	if newInput, ok := newComponent.(*Input); ok {
		// Update callbacks (props that change between renders)
		t.OnSubmit = newInput.OnSubmit
		t.OnChange = newInput.OnChange
		t.Placeholder = newInput.Placeholder
		// Note: We do NOT update Value, CursorPos, Focused - those are state
	}
}

// WithCursorStyle sets the cursor style and returns the component for chaining
func (t *Input) WithCursorStyle(style CursorStyle) *Input {
	t.CursorStyle = style
	return t
}

// WithValue sets the initial value and returns the component for chaining
func (t *Input) WithValue(value string) *Input {
	t.Value = value
	t.CursorPos = len(value)
	return t
}

// WithType sets the input type and returns the component for chaining
func (t *Input) WithType(inputType InputType) *Input {
	t.Type = inputType
	return t
}

// getDisplayValue returns the value to display (masked for password type)
func (t *Input) getDisplayValue() string {
	if t.Type == InputTypePassword && len(t.Value) > 0 {
		// Mask password with asterisks
		runes := []rune(t.Value)
		masked := make([]rune, len(runes))
		for i := range runes {
			masked[i] = '*'
		}
		return string(masked)
	}
	return t.Value
}

// renderUnfocused renders the input without cursor (when not focused)
func (t *Input) renderUnfocused() *vdom.Element {
	// Render same structure but without cursor styling
	text := ""
	if len(t.Value) > 0 {
		text = t.getDisplayValue()
	} else if t.Placeholder != "" {
		text = t.Placeholder
	}

	return vdom.Box(
		vdom.HStack(
			vdom.Text("> "),
			vdom.Text(text),
		),
	).
		WithStyle("padding", "0 1")
}

// Render returns the Element for this component (React-like render)
func (t *Input) Render() *vdom.Element {
	// If not focused, render without cursor
	if !t.Focused {
		return t.renderUnfocused()
	}

	// When focused with empty value: show placeholder with cursor over first char
	if len(t.Value) == 0 && t.Placeholder != "" {
		// If placeholder has at least one character, show styled first char + rest
		if len(t.Placeholder) > 0 {
			// Get first rune (character) properly for Unicode support
			runes := []rune(t.Placeholder)
			firstChar := string(runes[0])
			var restOfPlaceholder string
			if len(runes) > 1 {
				restOfPlaceholder = string(runes[1:])
			}

			// Build cursor element based on style
			var cursorElements []interface{}
			if t.CursorStyle == CursorBar {
				// Bar cursor: show "|" before the character
				var barElements []interface{}
				barElements = append(barElements, vdom.Text("> "))
				// Only show bar if cursor is visible
				if t.CursorVisible {
					barElements = append(barElements, vdom.Text("|"))
				}
				barElements = append(barElements,
					vdom.Text(firstChar).WithStyle("color", "#808080"),
					vdom.Text(restOfPlaceholder).WithStyle("color", "#808080"),
				)
				cursorElements = barElements
			} else {
				// Block or Underline: style the first character
				var styledChar *vdom.Element
				if t.CursorVisible {
					// Show cursor styling only when visible
					styledChar = vdom.Text(firstChar).WithStyle("color", "#808080")
					if t.CursorStyle == CursorBlock {
						// Block cursor on placeholder: light background with gray text visible
						styledChar = styledChar.WithStyle("background-color", "#ffffff")
					} else {
						// Underline on placeholder: gray text with underline
						styledChar = styledChar.WithStyle("text-decoration", "underline")
					}
				} else {
					// Cursor not visible: just gray text, no styling
					styledChar = vdom.Text(firstChar).WithStyle("color", "#808080")
				}

				cursorElements = []interface{}{
					vdom.Text("> "),
					styledChar,
					vdom.Text(restOfPlaceholder).WithStyle("color", "#808080"),
				}
			}

			return vdom.Box(
				vdom.HStack(cursorElements...),
			).
				WithStyle("padding", "0 1")
		}

		// Placeholder is empty, just show prompt + cursor
		cursorChar := t.GetCursorChar()
		return vdom.Box(
			vdom.HStack(
				vdom.Text("> "),
				vdom.Text(cursorChar),
			),
		).
			WithStyle("padding", "0 1")
	}

	// With text: show value with cursor (no placeholder suffix)
	visible, cursorOffset := t.GetVisible()

	// Split visible text into: before cursor, cursor char, after cursor
	var beforeCursor, cursorChar, afterCursor string
	cursorAtEnd := cursorOffset >= len(visible)

	if cursorAtEnd {
		// Cursor at end
		beforeCursor = visible
		cursorChar = " " // Space (will be replaced with cursor glyph if visible)
		afterCursor = ""
	} else {
		beforeCursor = visible[:cursorOffset]
		cursorChar = string(visible[cursorOffset])
		if cursorOffset+1 < len(visible) {
			afterCursor = visible[cursorOffset+1:]
		}
	}

	// For multi-line content, render as VStack with prompt on first line
	lines := t.getLines()
	if len(lines) > 1 {
		// Multi-line: first line has prompt, rest are indented
		display := t.GetDisplayWithCursor()
		displayLines := t.getDisplayLines(display)
		children := make([]interface{}, len(displayLines))
		children[0] = vdom.HStack(
			vdom.Text("> "),
			vdom.Text(displayLines[0]),
		)
		for i := 1; i < len(displayLines); i++ {
			children[i] = vdom.HStack(
				vdom.Text("  "), // Indent continuation lines
				vdom.Text(displayLines[i]),
			)
		}
		return vdom.Box(
			vdom.VStack(children...),
		).
			WithStyle("padding", "0 1")
	}

	// Single line: prompt + before + cursor + after
	var textElements []interface{}
	textElements = append(textElements, vdom.Text("> "))

	if t.CursorStyle == CursorBar {
		// Bar cursor: show "|" before the cursor character
		if beforeCursor != "" {
			textElements = append(textElements, vdom.Text(beforeCursor))
		}
		// Only show cursor bar if focused and visible
		if t.Focused && t.CursorVisible {
			textElements = append(textElements, vdom.Text("|"))
		}
		textElements = append(textElements, vdom.Text(cursorChar))
		if afterCursor != "" {
			textElements = append(textElements, vdom.Text(afterCursor))
		}
	} else {
		// Block or Underline: style the cursor character
		if beforeCursor != "" {
			textElements = append(textElements, vdom.Text(beforeCursor))
		}
		// Only show cursor styling if cursor is visible (not in blink-off state)
		if t.Focused && t.CursorVisible {
			// When cursor is at end, show the actual cursor glyph instead of styled space
			if cursorAtEnd {
				// Use actual cursor character (█, _, etc.) for visibility
				cursorGlyph := t.GetCursorChar()
				textElements = append(textElements, vdom.Text(cursorGlyph))
			} else {
				// Cursor over a character: apply inverse styling
				textElements = append(textElements, t.applyCursorStyle(vdom.Text(cursorChar)))
			}
		} else {
			// Cursor not visible: show character normally (or nothing if at end)
			if !cursorAtEnd {
				textElements = append(textElements, vdom.Text(cursorChar))
			}
		}
		if afterCursor != "" {
			textElements = append(textElements, vdom.Text(afterCursor))
		}
	}

	return vdom.Box(
		vdom.HStack(textElements...),
	).
		WithStyle("padding", "0 1")
}

// getDisplayLines splits display text into lines for rendering
func (t *Input) getDisplayLines(display string) []string {
	lines := make([]string, 0)
	start := 0
	for i := 0; i < len(display); i++ {
		if display[i] == '\n' {
			lines = append(lines, display[start:i])
			start = i + 1
		}
	}
	// Add last line
	if start < len(display) {
		lines = append(lines, display[start:])
	} else {
		lines = append(lines, "")
	}
	return lines
}

// GetVisible returns the visible portion of the text and the cursor offset within it
func (t *Input) GetVisible() (visible string, cursorOffset int) {
	visibleWidth := t.Width
	if visibleWidth < 1 {
		visibleWidth = 10 // Minimum width
	}

	endPos := t.Scroll + visibleWidth
	if endPos > len(t.Value) {
		endPos = len(t.Value)
	}

	visible = t.Value[t.Scroll:endPos]
	cursorOffset = t.CursorPos - t.Scroll
	return visible, cursorOffset
}

// GetDisplay returns the text to display (visible portion or placeholder)
func (t *Input) GetDisplay() string {
	visible, _ := t.GetVisible()
	if len(visible) == 0 && t.Placeholder != "" {
		return t.Placeholder
	}
	if len(visible) == 0 {
		return " " // Empty space to maintain layout
	}
	return visible
}

// --- Focusable interface implementation ---

// IsFocusable implements Focusable interface
// Returns true if this component can receive focus
func (t *Input) IsFocusable() bool {
	return !t.Disabled
}

// SetFocusState sets whether this component has focus (called by runtime)
func (t *Input) SetFocusState(focused bool) {
	t.Focused = focused
	if focused {
		// Reset cursor visibility when gaining focus (cursor should always be visible when focused)
		t.CursorVisible = true
		t.lastBlinkTime = time.Now()
	}
}

// IsNode implements vdom.Node interface
func (t *Input) IsNode() {}

// --- State persistence methods (for HMR) ---

// GetID returns the component ID (for Stateful interface)
func (t *Input) GetID() string {
	return t.ID
}

// SaveState returns the component state for HMR persistence
func (t *Input) SaveState() map[string]interface{} {
	return map[string]interface{}{
		"value":     t.Value,
		"cursorPos": t.CursorPos,
		"scroll":    t.Scroll,
	}
}

// LoadState restores the component state from HMR
func (t *Input) LoadState(state map[string]interface{}) error {
	if value, ok := state["value"].(string); ok {
		t.Value = value
	}
	if cursorPos, ok := state["cursorPos"].(float64); ok {
		t.CursorPos = int(cursorPos)
	}
	if scroll, ok := state["scroll"].(float64); ok {
		t.Scroll = int(scroll)
	}
	return nil
}
