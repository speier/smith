// Package component provides reusable UI components for Lotus applications
package components

import (
	"time"

	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
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

// TextInputProps represents the props (configuration from parent) for TextInput
// React pattern: Props are immutable configuration passed by parent
type TextInputProps struct {
	Placeholder string
	Width       int
	Disabled    bool
	OnChange    func(value string)
	OnSubmit    func(value string)
}

// TextInput is a reusable text input field with cursor navigation and multi-line support
type TextInput struct {
	// Component metadata
	ID string // Component ID for registration (React key) - MUST be set for state preservation

	// Internal state (private to component)
	Value       string // Current input text (may contain newlines for multi-line)
	CursorPos   int    // Cursor position in the full text (0-indexed, counts newlines)
	Scroll      int    // Horizontal scroll offset (for single-line scrolling)
	desiredCol  int    // Desired column for vertical navigation (preserved across up/down)

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
	OnChange func(value string) // Called when text changes
	OnSubmit func(value string) // Called when Enter is pressed
}

// NewTextInput creates a new TextInput with optional ID
// React-like: TextInput() creates component instance
func NewTextInput(id ...string) *TextInput {
	inputID := ""
	if len(id) > 0 {
		inputID = id[0]
	}
	return &TextInput{
		ID:            inputID,
		Value:         "",
		CursorPos:     0,
		Scroll:        0,
		Width:         50, // Default width
		CursorStyle:   CursorBlock,
		CursorBlink:   true,
		CursorVisible: true,
		BlinkInterval: 500 * time.Millisecond,
		lastBlinkTime: time.Now(),
	}
}

// SetProps updates the component props (React pattern)
// Allows parent to configure component behavior
func (t *TextInput) SetProps(props TextInputProps) {
	t.Placeholder = props.Placeholder
	if props.Width > 0 {
		t.Width = props.Width
	}
	t.Disabled = props.Disabled
	t.OnChange = props.OnChange
	t.OnSubmit = props.OnSubmit
}

// Fluent API methods (React-like chaining)

// WithID sets the component ID and returns the component for chaining
// ID is required for HMR state persistence to work
func (t *TextInput) WithID(id string) *TextInput {
	t.ID = id
	return t
}

// WithPlaceholder sets the placeholder text and returns the component for chaining
func (t *TextInput) WithPlaceholder(placeholder string) *TextInput {
	t.Placeholder = placeholder
	return t
}

// WithWidth sets the width and returns the component for chaining
func (t *TextInput) WithWidth(width int) *TextInput {
	t.Width = width
	return t
}

// WithOnChange sets the onChange callback and returns the component for chaining
func (t *TextInput) WithOnChange(onChange func(string)) *TextInput {
	t.OnChange = onChange
	return t
}

// WithOnSubmit sets the onSubmit callback and returns the component for chaining
func (t *TextInput) WithOnSubmit(onSubmit func(string)) *TextInput {
	t.OnSubmit = onSubmit
	return t
}

// WithCursorStyle sets the cursor style and returns the component for chaining
func (t *TextInput) WithCursorStyle(style CursorStyle) *TextInput {
	t.CursorStyle = style
	return t
}

// WithValue sets the initial value and returns the component for chaining
func (t *TextInput) WithValue(value string) *TextInput {
	t.Value = value
	t.CursorPos = len(value)
	return t
}

// applyCursorStyle applies the appropriate cursor styling to a text element
func (t *TextInput) applyCursorStyle(textElem *vdom.Element) *vdom.Element {
	switch t.CursorStyle {
	case CursorBlock:
		// Block cursor: bright background with dark text (inverse video)
		return textElem.
			WithStyle("color", "#000000").
			WithStyle("background-color", "#ffffff")
	case CursorUnderline:
		// Underline cursor: underline decoration
		return textElem.WithStyle("text-decoration", "underline")
	case CursorBar:
		// Bar cursor: render a vertical bar before the character
		// Note: This is handled differently - we'll prefix with "|" 
		return textElem
	default:
		return textElem
	}
}

// Render returns the Element for this component (React-like render)
func (t *TextInput) Render() *vdom.Element {
	// Always show prompt "> " followed by content
	if len(t.Value) == 0 && t.Placeholder != "" {
		// Empty with placeholder: show first char with cursor styling (inverse/underline) + rest
		// This makes the character visible "through" the cursor
		
		// If placeholder has at least one character, show styled first char + rest
		if len(t.Placeholder) > 0 {
			firstChar := string(t.Placeholder[0])
			restOfPlaceholder := t.Placeholder[1:]
			
			// Build cursor element based on style
			var cursorElements []interface{}
			if t.CursorStyle == CursorBar {
				// Bar cursor: show "|" before the character
				cursorElements = []interface{}{
					vdom.Text("> "),
					vdom.Text("|"),
					vdom.Text(firstChar).WithStyle("color", "#808080"),
					vdom.Text(restOfPlaceholder).WithStyle("color", "#808080"),
				}
			} else {
				// Block or Underline: style the first character
				styledChar := vdom.Text(firstChar)
				if t.CursorStyle == CursorBlock {
					// Block cursor on placeholder: keep dark text visible through light background
					styledChar = t.applyCursorStyle(styledChar)
				} else {
					// Underline on placeholder: gray text with underline
					styledChar = styledChar.WithStyle("color", "#808080").WithStyle("text-decoration", "underline")
				}
				
				cursorElements = []interface{}{
					vdom.Text("> "),
					styledChar,
					vdom.Text(restOfPlaceholder).WithStyle("color", "#808080"),
				}
			}
			
			return vdom.Box(
				vdom.HStack(cursorElements...),
			).WithID(t.ID).
				WithStyle("padding", "0 1")
		}
		
		// Placeholder is empty, just show prompt + cursor
		cursorChar := t.GetCursorChar()
		return vdom.Box(
			vdom.HStack(
				vdom.Text("> "),
				vdom.Text(cursorChar),
			),
		).WithID(t.ID).
			WithStyle("padding", "0 1")
	}

	// With text: show prompt + text with inverse video cursor (supports multi-line)
	visible, cursorOffset := t.GetVisible()
	
	// Split visible text into: before cursor, cursor char, after cursor
	var beforeCursor, cursorChar, afterCursor string
	if cursorOffset >= len(visible) {
		// Cursor at end
		beforeCursor = visible
		cursorChar = " " // Show space with cursor
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
		// TODO: Handle cursor position across multiple lines
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
		).WithID(t.ID).
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
		textElements = append(textElements, vdom.Text("|"))
		textElements = append(textElements, vdom.Text(cursorChar))
		if afterCursor != "" {
			textElements = append(textElements, vdom.Text(afterCursor))
		}
	} else {
		// Block or Underline: style the cursor character
		if beforeCursor != "" {
			textElements = append(textElements, vdom.Text(beforeCursor))
		}
		textElements = append(textElements, t.applyCursorStyle(vdom.Text(cursorChar)))
		if afterCursor != "" {
			textElements = append(textElements, vdom.Text(afterCursor))
		}
	}
	
	return vdom.Box(
		vdom.HStack(textElements...),
	).WithID(t.ID).
		WithStyle("padding", "0 1")
}

// getDisplayLines splits display text into lines for rendering
func (t *TextInput) getDisplayLines(display string) []string {
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

// InsertChar inserts a character at the cursor position
func (t *TextInput) InsertChar(ch string) {
	t.Value = t.Value[:t.CursorPos] + ch + t.Value[t.CursorPos:]
	t.CursorPos++
	t.adjustScroll()
	t.desiredCol = 0 // Reset desired column on text change
}

// InsertNewline inserts a newline at the cursor position (for multi-line support)
func (t *TextInput) InsertNewline() {
	t.Value = t.Value[:t.CursorPos] + "\n" + t.Value[t.CursorPos:]
	t.CursorPos++
	t.desiredCol = 0 // Reset desired column on text change
}

// DeleteChar deletes the character before the cursor (backspace)
func (t *TextInput) DeleteChar() {
	if t.CursorPos > 0 {
		t.Value = t.Value[:t.CursorPos-1] + t.Value[t.CursorPos:]
		t.CursorPos--
		t.adjustScroll()
	}
}

// DeleteForward deletes the character at the cursor (delete key)
func (t *TextInput) DeleteForward() {
	if t.CursorPos < len(t.Value) {
		t.Value = t.Value[:t.CursorPos] + t.Value[t.CursorPos+1:]
	}
}

// MoveLeft moves the cursor one position to the left
func (t *TextInput) MoveLeft() {
	if t.CursorPos > 0 {
		t.CursorPos--
		t.adjustScroll()
		t.desiredCol = 0 // Reset desired column on horizontal movement
	}
}

// MoveRight moves the cursor one position to the right
func (t *TextInput) MoveRight() {
	if t.CursorPos < len(t.Value) {
		t.CursorPos++
		t.adjustScroll()
		t.desiredCol = 0 // Reset desired column on horizontal movement
	}
}

// Home moves the cursor to the beginning
func (t *TextInput) Home() {
	t.CursorPos = 0
	t.Scroll = 0
}

// End moves the cursor to the end
func (t *TextInput) End() {
	t.CursorPos = len(t.Value)
	t.adjustScroll()
}

// DeleteToBeginning deletes from cursor to beginning of input (Cmd+Backspace / Ctrl+U)
func (t *TextInput) DeleteToBeginning() {
	if t.CursorPos > 0 {
		t.Value = t.Value[t.CursorPos:]
		t.CursorPos = 0
		t.Scroll = 0
	}
}

// DeleteToEnd deletes from cursor to end of input (Ctrl+K)
func (t *TextInput) DeleteToEnd() {
	if t.CursorPos < len(t.Value) {
		t.Value = t.Value[:t.CursorPos]
	}
}

// DeleteWordBackward deletes the word before the cursor (Ctrl+Backspace / Ctrl+W)
func (t *TextInput) DeleteWordBackward() {
	if t.CursorPos == 0 {
		return
	}

	oldPos := t.CursorPos

	// Skip any spaces at current position
	for t.CursorPos > 0 && t.Value[t.CursorPos-1] == ' ' {
		t.CursorPos--
	}

	// Delete to start of current/previous word
	for t.CursorPos > 0 && t.Value[t.CursorPos-1] != ' ' {
		t.CursorPos--
	}

	// Remove the deleted text
	t.Value = t.Value[:t.CursorPos] + t.Value[oldPos:]
	t.adjustScroll()
}

// MoveWordLeft moves the cursor to the start of the previous word
func (t *TextInput) MoveWordLeft() {
	if t.CursorPos == 0 {
		return
	}

	// Skip any spaces at current position
	for t.CursorPos > 0 && t.Value[t.CursorPos-1] == ' ' {
		t.CursorPos--
	}

	// Move to start of current/previous word
	for t.CursorPos > 0 && t.Value[t.CursorPos-1] != ' ' {
		t.CursorPos--
	}

	t.adjustScroll()
}

// MoveWordRight moves the cursor to the start of the next word
func (t *TextInput) MoveWordRight() {
	if t.CursorPos >= len(t.Value) {
		return
	}

	// Skip any spaces at current position
	for t.CursorPos < len(t.Value) && t.Value[t.CursorPos] == ' ' {
		t.CursorPos++
	}

	// Move to end of current/next word
	for t.CursorPos < len(t.Value) && t.Value[t.CursorPos] != ' ' {
		t.CursorPos++
	}

	t.adjustScroll()
}

// MoveUp moves cursor to previous line (multi-line support)
func (t *TextInput) MoveUp() {
	lines := t.getLines()
	currentLine, col := t.getCurrentLineAndCol()

	// Remember desired column on first vertical move
	if t.desiredCol == 0 {
		t.desiredCol = col
	}

	if currentLine > 0 {
		// Move to previous line, using desired column (or end if line is shorter)
		prevLineStart := t.getLineStart(currentLine - 1)
		prevLineLen := len(lines[currentLine-1])
		targetCol := t.desiredCol
		if targetCol > prevLineLen {
			targetCol = prevLineLen
		}
		t.CursorPos = prevLineStart + targetCol
	}
}

// MoveDown moves cursor to next line (multi-line support)
func (t *TextInput) MoveDown() {
	lines := t.getLines()
	currentLine, col := t.getCurrentLineAndCol()

	// Remember desired column on first vertical move
	if t.desiredCol == 0 {
		t.desiredCol = col
	}

	if currentLine < len(lines)-1 {
		// Move to next line, using desired column (or end if line is shorter)
		nextLineStart := t.getLineStart(currentLine + 1)
		nextLineLen := len(lines[currentLine+1])
		targetCol := t.desiredCol
		if targetCol > nextLineLen {
			targetCol = nextLineLen
		}
		t.CursorPos = nextLineStart + targetCol
	}
}

// getLines splits the value into lines
func (t *TextInput) getLines() []string {
	if t.Value == "" {
		return []string{""}
	}
	lines := make([]string, 0)
	start := 0
	for i, ch := range t.Value {
		if ch == '\n' {
			lines = append(lines, t.Value[start:i])
			start = i + 1
		}
	}
	// Add last line (even if empty)
	lines = append(lines, t.Value[start:])
	return lines
}

// getCurrentLineAndCol returns the current line number and column position
func (t *TextInput) getCurrentLineAndCol() (line, col int) {
	pos := 0
	line = 0
	for i := 0; i < t.CursorPos && i < len(t.Value); i++ {
		if t.Value[i] == '\n' {
			line++
			pos = 0
		} else {
			pos++
		}
	}
	col = pos
	return line, col
}

// getLineStart returns the starting position of a given line number
func (t *TextInput) getLineStart(lineNum int) int {
	if lineNum == 0 {
		return 0
	}
	pos := 0
	currentLine := 0
	for i := 0; i < len(t.Value); i++ {
		if t.Value[i] == '\n' {
			currentLine++
			if currentLine == lineNum {
				return i + 1
			}
		}
		pos = i + 1
	}
	return pos
}

// Clear clears the input and resets cursor
func (t *TextInput) Clear() {
	t.Value = ""
	t.CursorPos = 0
	t.Scroll = 0
}

// GetVisible returns the visible portion of the text and the cursor offset within it
func (t *TextInput) GetVisible() (visible string, cursorOffset int) {
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
func (t *TextInput) GetDisplay() string {
	visible, _ := t.GetVisible()
	if len(visible) == 0 && t.Placeholder != "" {
		return t.Placeholder
	}
	if len(visible) == 0 {
		return " " // Empty space to maintain layout
	}
	return visible
}

// HandleKey handles a key event and returns true if it was handled
// Returns false if the key should be handled by the application (e.g., Enter)
func (t *TextInput) HandleKey(event tty.KeyEvent) bool {
	oldValue := t.Value

	// Printable characters
	if event.IsPrintable() {
		t.InsertChar(event.Char)
		t.emitChange(oldValue)
		return true
	}

	// Backspace
	if event.IsBackspace() {
		t.DeleteChar()
		t.emitChange(oldValue)
		return true
	}

	// Delete key
	if event.Code == tty.SeqDelete {
		t.DeleteForward()
		t.emitChange(oldValue)
		return true
	}

	// Shift+Enter - insert newline (multi-line support)
	if event.Code == tty.SeqShiftEnter {
		t.InsertNewline()
		t.emitChange(oldValue)
		return true
	}

	// Enter key - emit submit event (normal behavior)
	if event.IsEnter() {
		if t.OnSubmit != nil {
			t.OnSubmit(t.Value)
		}
		return false // Let app handle it too
	}

	// Arrow keys
	if event.Code == tty.SeqLeft {
		t.MoveLeft()
		return true
	}

	if event.Code == tty.SeqRight {
		t.MoveRight()
		return true
	}

	if event.Code == tty.SeqUp {
		t.MoveUp()
		return true
	}

	if event.Code == tty.SeqDown {
		t.MoveDown()
		return true
	}

	// Word navigation (Ctrl+Left/Right)
	if event.Code == tty.SeqCtrlLeft {
		t.MoveWordLeft()
		return true
	}

	if event.Code == tty.SeqCtrlRight {
		t.MoveWordRight()
		return true
	}

	// Cmd+Left/Right on Mac (beginning/end of line)
	if event.Code == tty.SeqCmdLeft {
		t.Home()
		return true
	}

	if event.Code == tty.SeqCmdRight {
		t.End()
		return true
	}

	// Home key
	if event.Code == tty.SeqHome {
		t.Home()
		return true
	}

	// End key
	if event.Code == tty.SeqEnd {
		t.End()
		return true
	}

	// Cmd+Backspace (Mac) or Ctrl+U - delete to beginning
	if event.Code == tty.SeqCmdBackspace || event.Key == '\x15' { // Ctrl+U is 0x15
		t.DeleteToBeginning()
		t.emitChange(oldValue)
		return true
	}

	// Ctrl+K - delete to end
	if event.Key == '\x0b' { // Ctrl+K is 0x0b
		t.DeleteToEnd()
		t.emitChange(oldValue)
		return true
	}

	// Ctrl+Backspace or Ctrl+W - delete word backward
	if event.Code == tty.SeqCtrlBackspace || event.Key == '\x17' { // Ctrl+W is 0x17
		t.DeleteWordBackward()
		t.emitChange(oldValue)
		return true
	}

	// Not handled - let application handle it (e.g., Tab, etc.)
	return false
}

// emitChange triggers OnChange callback if value changed
func (t *TextInput) emitChange(oldValue string) {
	if t.Value != oldValue && t.OnChange != nil {
		t.OnChange(t.Value)
	}
}

// adjustScroll adjusts horizontal scroll to keep cursor visible
func (t *TextInput) adjustScroll() {
	visibleWidth := t.Width
	if visibleWidth < 1 {
		visibleWidth = 10
	}

	// If cursor is past the visible area, scroll right
	if t.CursorPos-t.Scroll >= visibleWidth {
		t.Scroll = t.CursorPos - visibleWidth + 1
	}

	// If cursor is before the visible area, scroll left
	if t.CursorPos < t.Scroll {
		t.Scroll = t.CursorPos
	}
}

// --- Focusable interface implementation ---

// HandleKeyEvent implements Focusable interface
// Processes keyboard events when this component has focus
// Returns true if the event was handled, false to bubble up (e.g., Enter key)
func (t *TextInput) HandleKeyEvent(event tty.KeyEvent) bool {
	return t.HandleKey(event)
}

// GetCursorOffset implements Focusable interface
// Returns the cursor position offset within the visible text (including prompt)
func (t *TextInput) GetCursorOffset() int {
	_, offset := t.GetVisible()
	// Add prompt length ("> " = 2 characters)
	return 2 + offset
}

// IsFocusable implements Focusable interface
// Returns true if this component can receive focus
func (t *TextInput) IsFocusable() bool {
	return !t.Disabled
}

// IsNode implements vdom.Node interface
func (t *TextInput) IsNode() {}

// --- Cursor methods (100% testable!) ---

// UpdateCursorBlink updates the cursor blink state based on elapsed time
// Returns true if cursor visibility changed (for re-rendering)
func (t *TextInput) UpdateCursorBlink() bool {
	if !t.CursorBlink || t.BlinkInterval == 0 {
		return false
	}

	now := time.Now()
	if now.Sub(t.lastBlinkTime) >= t.BlinkInterval {
		t.CursorVisible = !t.CursorVisible
		t.lastBlinkTime = now
		return true // Visibility changed
	}

	return false // No change
}

// SetCursorStyle sets the cursor visual style
func (t *TextInput) SetCursorStyle(style CursorStyle) {
	t.CursorStyle = style
}

// SetCursorBlink enables or disables cursor blinking
func (t *TextInput) SetCursorBlink(enabled bool) {
	t.CursorBlink = enabled
	if !enabled {
		t.CursorVisible = true // Always visible if not blinking
	}
}

// SetBlinkInterval sets how fast the cursor blinks
func (t *TextInput) SetBlinkInterval(interval time.Duration) {
	t.BlinkInterval = interval
}

// GetCursorChar returns the character to display for the cursor
func (t *TextInput) GetCursorChar() string {
	if !t.CursorVisible {
		return " " // Invisible during blink
	}

	switch t.CursorStyle {
	case CursorBlock:
		return "█"
	case CursorUnderline:
		return "_"
	case CursorBar:
		return "|"
	default:
		return "█"
	}
}

// GetDisplayWithCursor returns the text with cursor inserted at current position
// This is 100% testable without terminal rendering!
func (t *TextInput) GetDisplayWithCursor() string {
	visible, cursorOffset := t.GetVisible()

	// Cursor character
	cursorChar := t.GetCursorChar()

	// Insert cursor at the offset position
	if cursorOffset < 0 {
		cursorOffset = 0
	}
	if cursorOffset > len(visible) {
		cursorOffset = len(visible)
	}

	if cursorOffset == len(visible) {
		// Cursor at end
		return visible + cursorChar
	}

	// Cursor in middle - replace character at cursor position
	return visible[:cursorOffset] + cursorChar + visible[cursorOffset+1:]
}

// --- State persistence methods (for HMR) ---

// GetID returns the component ID (for Stateful interface)
func (t *TextInput) GetID() string {
	return t.ID
}

// SaveState returns the component state for HMR persistence
func (t *TextInput) SaveState() map[string]interface{} {
	return map[string]interface{}{
		"value":     t.Value,
		"cursorPos": t.CursorPos,
		"scroll":    t.Scroll,
	}
}

// LoadState restores the component state from HMR
func (t *TextInput) LoadState(state map[string]interface{}) error {
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
