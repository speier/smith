package primitives

import "unicode"

// This file contains text editing operations for Input component

// InsertChar inserts a character at the cursor position
func (t *Input) InsertChar(ch string) {
	// Validate input based on type
	if t.Type == InputTypeNumber {
		// Only allow digits, decimal point, minus sign
		if len(ch) == 1 {
			r := rune(ch[0])
			// Allow: digits, decimal point, minus (only at start)
			isValid := unicode.IsDigit(r) || ch == "."
			if ch == "-" && t.CursorPos == 0 {
				isValid = true // Allow minus at start
			}
			if !isValid {
				return // Reject invalid character
			}
		}
	}

	t.Value = t.Value[:t.CursorPos] + ch + t.Value[t.CursorPos:]
	t.CursorPos++
	t.adjustScroll()
	t.desiredCol = 0 // Reset desired column on text change
}

// InsertNewline inserts a newline at the cursor position (for multi-line support)
func (t *Input) InsertNewline() {
	t.Value = t.Value[:t.CursorPos] + "\n" + t.Value[t.CursorPos:]
	t.CursorPos++
	t.desiredCol = 0 // Reset desired column on text change
}

// DeleteChar deletes the character before the cursor (backspace)
func (t *Input) DeleteChar() {
	if t.CursorPos > 0 {
		t.Value = t.Value[:t.CursorPos-1] + t.Value[t.CursorPos:]
		t.CursorPos--
		t.adjustScroll()
	}
}

// DeleteForward deletes the character at the cursor (delete key)
func (t *Input) DeleteForward() {
	if t.CursorPos < len(t.Value) {
		t.Value = t.Value[:t.CursorPos] + t.Value[t.CursorPos+1:]
	}
}

// MoveLeft moves the cursor one position to the left
func (t *Input) MoveLeft() {
	if t.CursorPos > 0 {
		t.CursorPos--
		t.adjustScroll()
		t.desiredCol = 0 // Reset desired column on horizontal movement
	}
}

// MoveRight moves the cursor one position to the right
func (t *Input) MoveRight() {
	if t.CursorPos < len(t.Value) {
		t.CursorPos++
		t.adjustScroll()
		t.desiredCol = 0 // Reset desired column on horizontal movement
	}
}

// Home moves the cursor to the beginning
func (t *Input) Home() {
	t.CursorPos = 0
	t.Scroll = 0
}

// End moves the cursor to the end
func (t *Input) End() {
	t.CursorPos = len(t.Value)
	t.adjustScroll()
}

// DeleteToBeginning deletes from cursor to beginning of input (Cmd+Backspace / Ctrl+U)
func (t *Input) DeleteToBeginning() {
	if t.CursorPos > 0 {
		t.Value = t.Value[t.CursorPos:]
		t.CursorPos = 0
		t.Scroll = 0
	}
}

// DeleteToEnd deletes from cursor to end of input (Ctrl+K)
func (t *Input) DeleteToEnd() {
	if t.CursorPos < len(t.Value) {
		t.Value = t.Value[:t.CursorPos]
	}
}

// DeleteWordBackward deletes the word before the cursor (Ctrl+Backspace / Ctrl+W)
func (t *Input) DeleteWordBackward() {
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
func (t *Input) MoveWordLeft() {
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
func (t *Input) MoveWordRight() {
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
func (t *Input) MoveUp() {
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
func (t *Input) MoveDown() {
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

// Clear clears the input and resets cursor
func (t *Input) Clear() {
	t.Value = ""
	t.CursorPos = 0
	t.Scroll = 0
}

// getLines splits the value into lines
func (t *Input) getLines() []string {
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
func (t *Input) getCurrentLineAndCol() (line, col int) {
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
func (t *Input) getLineStart(lineNum int) int {
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
