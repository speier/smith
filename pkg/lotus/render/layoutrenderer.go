package render

import (
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/speier/smith/pkg/lotus/layout"
	"github.com/speier/smith/pkg/lotus/style"
)

// ScrollManager interface for managing scroll state (decoupled from runtime)
type ScrollManager interface {
	GetOffset(id string) (int, int)
	UpdateDimensions(id string, contentW, contentH, viewportW, viewportH int)
}

// LayoutRenderer converts a LayoutBox tree into a Buffer
type LayoutRenderer struct {
	ScrollManager ScrollManager // Optional: for overflow:auto support
}

// NewLayoutRenderer creates a new layout renderer
func NewLayoutRenderer() *LayoutRenderer {
	return &LayoutRenderer{}
}

// RenderToBuffer converts a layout tree into a 2D buffer
func (lr *LayoutRenderer) RenderToBuffer(box *layout.LayoutBox, width, height int) *Buffer {
	buf := NewBuffer(width, height)
	lr.renderBox(buf, box)
	return buf
}

// renderBox renders a single layout box and its children into the buffer
func (lr *LayoutRenderer) renderBox(buf *Buffer, box *layout.LayoutBox) {
	if box == nil || box.Node == nil {
		return
	}

	st := box.Node.Style

	// Convert style to buffer style
	bufStyle := lr.styleToBufferStyle(st)

	// Handle overflow: auto (CSS-like scrolling)
	if st.Overflow == "auto" && box.Node.Element != nil && lr.ScrollManager != nil {
		// Use element path as stable identifier (managed by reconciler)
		scrollID := box.Node.Element.Path
		if scrollID == "" {
			scrollID = "0" // fallback for root
		}

		// Calculate content bounds
		var maxContentWidth, maxContentHeight int

		// Calculate bounds recursively from all descendants, not just immediate children
		var calculateMaxBounds func(*layout.LayoutBox)
		calculateMaxBounds = func(lb *layout.LayoutBox) {
			// Calculate this box's bounds
			childRight := lb.X + lb.Width - box.X
			childBottom := lb.Y + lb.Height - box.Y
			if childRight > maxContentWidth {
				maxContentWidth = childRight
			}
			if childBottom > maxContentHeight {
				maxContentHeight = childBottom
			}

			// Recurse into children
			for _, child := range lb.Children {
				calculateMaxBounds(child)
			}
		}

		// Calculate from all children recursively
		for _, child := range box.Children {
			calculateMaxBounds(child)
		}

		// Update scroll manager with dimensions
		lr.ScrollManager.UpdateDimensions(scrollID, maxContentWidth, maxContentHeight, box.Width, box.Height)

		// Create temp buffer for content
		tempHeight := max(box.Height, maxContentHeight)
		tempBuf := NewBuffer(box.Width, tempHeight)

		// Render children to temp buffer
		for _, child := range box.Children {
			adjustedChild := lr.adjustLayoutTree(child, box.X, box.Y)
			lr.renderBox(tempBuf, adjustedChild)
		}

		// Get scroll offset and clip
		scrollX, scrollY := lr.ScrollManager.GetOffset(scrollID)
		clipped := tempBuf.Clip(scrollX, scrollY, box.Width, box.Height)

		// Copy to main buffer
		for y := 0; y < clipped.Height; y++ {
			for x := 0; x < clipped.Width; x++ {
				cell := clipped.Get(x, y)
				buf.Set(box.X+x, box.Y+y, cell)
			}
		}
		return
	}

	// Render border if present
	if st.Border {
		lr.renderBorder(buf, box, st)
	}

	// Render based on element type
	if box.Node.Element != nil {
		switch box.Node.Element.Type {
		case 1: // TextElement
			lr.renderText(buf, box, bufStyle)
		default:
			// Container - render children
			for _, child := range box.Children {
				lr.renderBox(buf, child)
			}
		}
	} else {
		// No element, just render children
		for _, child := range box.Children {
			lr.renderBox(buf, child)
		}
	}
}

// renderText renders text content into the buffer
func (lr *LayoutRenderer) renderText(buf *Buffer, box *layout.LayoutBox, bufStyle Style) {
	if box.Node.Element == nil || box.Node.Element.Text == "" {
		return
	}

	st := box.Node.Style

	// Skip if hidden
	if st.Visibility == "hidden" {
		return
	}

	text := box.Node.Element.Text
	if text == "" {
		return
	}

	// Calculate padding (already resolved as integers)
	paddingTop := st.PaddingTop
	paddingLeft := st.PaddingLeft

	// Calculate content area
	contentX := box.X + paddingLeft
	contentY := box.Y + paddingTop
	contentWidth := box.Width - paddingLeft - st.PaddingRight
	contentHeight := box.Height - paddingTop - st.PaddingBottom

	if contentWidth <= 0 || contentHeight <= 0 {
		return
	}

	// Wrap text to fit content width
	lines := lr.wrapText(text, contentWidth)

	// Apply line clamping if MaxLines is set
	if st.MaxLines > 0 && len(lines) > st.MaxLines {
		// Keep only first MaxLines lines
		lines = lines[:st.MaxLines]

		// Add ellipsis to the last visible line
		lastIdx := len(lines) - 1
		lines[lastIdx] = lines[lastIdx] + "..."
	}

	// Apply text-align for each line
	for i, line := range lines {
		if i >= contentHeight {
			break // Don't exceed content height
		}

		y := contentY + i
		x := contentX

		// Apply text alignment
		switch st.TextAlign {
		case "center":
			lineWidth := lr.displayWidth(line)
			x = contentX + (contentWidth-lineWidth)/2
		case "right":
			lineWidth := lr.displayWidth(line)
			x = contentX + (contentWidth - lineWidth)
		}

		// Parse and render line with ANSI support
		lr.renderLineWithANSI(buf, line, x, y, contentX, contentWidth, bufStyle)
	}
}

// renderLineWithANSI renders a line of text, parsing ANSI escape codes and applying them as styles
func (lr *LayoutRenderer) renderLineWithANSI(buf *Buffer, line string, startX, y, contentX, contentWidth int, baseStyle Style) {
	x := startX
	currentStyle := baseStyle
	i := 0

	for i < len(line) {
		// Check for ANSI escape sequence
		if line[i] == '\033' && i+1 < len(line) && line[i+1] == '[' {
			// Found ANSI sequence, parse it
			j := i + 2
			for j < len(line) && (line[j] < 'A' || line[j] > 'Z') && (line[j] < 'a' || line[j] > 'z') {
				j++
			}
			if j < len(line) {
				// Extract the code
				code := line[i+2 : j]

				// Apply ANSI code to current style
				currentStyle = lr.applyANSICode(code, baseStyle, currentStyle)

				i = j + 1 // Skip past the ANSI sequence
				continue
			}
		}

		// Regular character - render it
		ch, size := decodeRune(line[i:])
		if ch != 0 {
			// Skip variation selectors (U+FE00-U+FE0F, U+E0100-U+E01EF)
			// These are zero-width characters that modify the preceding character
			if (ch >= 0xFE00 && ch <= 0xFE0F) || (ch >= 0xE0100 && ch <= 0xE01EF) {
				i += size
				continue
			}

			charWidth := runewidth.RuneWidth(ch)
			// Check if there's enough space for the full character width
			if x+charWidth > contentX+contentWidth {
				break // Don't exceed content width
			}
			buf.Set(x, y, Cell{Char: ch, Style: currentStyle})

			// For wide characters (emoji, CJK), mark the next cell as occupied
			if charWidth == 2 {
				buf.Set(x+1, y, Cell{Char: '\u200B', Style: currentStyle})
			}

			x += charWidth
		}
		i += size
	}
}

// decodeRune extracts the next rune from a byte slice
func decodeRune(s string) (rune, int) {
	if len(s) == 0 {
		return 0, 0
	}
	// Simple UTF-8 decoding
	for i, r := range s {
		if i == 0 {
			// Calculate size by checking how many continuation bytes follow
			size := 1
			if r >= 0x80 {
				// Multi-byte sequence
				b := s[0]
				if b&0xE0 == 0xC0 {
					size = 2
				} else if b&0xF0 == 0xE0 {
					size = 3
				} else if b&0xF8 == 0xF0 {
					size = 4
				}
			}
			return r, size
		}
	}
	return 0, 0
}

// applyANSICode applies an ANSI escape code to the current style
func (lr *LayoutRenderer) applyANSICode(code string, baseStyle, currentStyle Style) Style {
	// Reset code
	if code == "0" || code == "" {
		return baseStyle
	}

	// Parse color codes
	switch code {
	case "30":
		currentStyle.FgColor = "black"
	case "31":
		currentStyle.FgColor = "red"
	case "32":
		currentStyle.FgColor = "green"
	case "33":
		currentStyle.FgColor = "yellow"
	case "34":
		currentStyle.FgColor = "blue"
	case "35":
		currentStyle.FgColor = "magenta"
	case "36":
		currentStyle.FgColor = "cyan"
	case "37":
		currentStyle.FgColor = "white"
	case "1":
		currentStyle.Bold = true
	case "2":
		currentStyle.Dim = true
	case "3":
		currentStyle.Italic = true
	case "4":
		currentStyle.Underline = true
	}

	return currentStyle
}

// renderBorder draws a border around a box in the buffer
// adjustLayoutTree recursively adjusts a layout tree to be relative to given offset
func (lr *LayoutRenderer) adjustLayoutTree(box *layout.LayoutBox, offsetX, offsetY int) *layout.LayoutBox {
	adjusted := &layout.LayoutBox{
		X:      box.X - offsetX,
		Y:      box.Y - offsetY,
		Width:  box.Width,
		Height: box.Height,
		Node:   box.Node,
	}

	// Recursively adjust children
	if len(box.Children) > 0 {
		adjusted.Children = make([]*layout.LayoutBox, len(box.Children))
		for i, child := range box.Children {
			adjusted.Children[i] = lr.adjustLayoutTree(child, offsetX, offsetY)
		}
	}

	return adjusted
}

func (lr *LayoutRenderer) renderBorder(buf *Buffer, box *layout.LayoutBox, st style.ComputedStyle) {
	if box.Width < 2 || box.Height < 2 {
		return
	}

	// Use border-color if set, otherwise use text color
	borderStyle := st
	if st.BorderColor != "" {
		borderStyle.Color = st.BorderColor
	}

	bufStyle := lr.styleToBufferStyle(borderStyle)

	// Determine border characters based on border style
	var tl, tr, bl, br, h, v rune
	switch st.BorderStyle {
	case "rounded":
		tl, tr, bl, br = '╭', '╮', '╰', '╯'
		h, v = '─', '│'
	case "double":
		tl, tr, bl, br = '╔', '╗', '╚', '╝'
		h, v = '═', '║'
	default: // "single" or empty
		tl, tr, bl, br = '┌', '┐', '└', '┘'
		h, v = '─', '│'
	}

	// Draw corners
	buf.Set(box.X, box.Y, Cell{Char: tl, Style: bufStyle})
	buf.Set(box.X+box.Width-1, box.Y, Cell{Char: tr, Style: bufStyle})
	buf.Set(box.X, box.Y+box.Height-1, Cell{Char: bl, Style: bufStyle})
	buf.Set(box.X+box.Width-1, box.Y+box.Height-1, Cell{Char: br, Style: bufStyle})

	// Draw horizontal lines
	for x := box.X + 1; x < box.X+box.Width-1; x++ {
		buf.Set(x, box.Y, Cell{Char: h, Style: bufStyle})
		buf.Set(x, box.Y+box.Height-1, Cell{Char: h, Style: bufStyle})
	}

	// Draw vertical lines
	for y := box.Y + 1; y < box.Y+box.Height-1; y++ {
		buf.Set(box.X, y, Cell{Char: v, Style: bufStyle})
		buf.Set(box.X+box.Width-1, y, Cell{Char: v, Style: bufStyle})
	}
}

// wrapText wraps text to fit within the given width
func (lr *LayoutRenderer) wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{}
	}

	var lines []string

	// First split by newlines (preserve explicit line breaks)
	inputLines := strings.Split(text, "\n")

	for _, inputLine := range inputLines {
		if inputLine == "" {
			lines = append(lines, "")
			continue
		}

		// Wrap each line if it's too long
		currentLine := ""
		currentWidth := 0

		words := strings.Fields(inputLine)
		for i, word := range words {
			wordWidth := lr.displayWidth(word)

			// If adding this word exceeds width, start new line
			if currentWidth+wordWidth > width && currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
				currentWidth = wordWidth
			} else {
				if currentLine != "" {
					currentLine += " "
					currentWidth++
				}
				currentLine += word
				currentWidth += wordWidth
			}

			// Handle last word
			if i == len(words)-1 && currentLine != "" {
				lines = append(lines, currentLine)
			}
		}
	}

	return lines
}

// displayWidth calculates the display width of a string, ignoring ANSI escape codes
func (lr *LayoutRenderer) displayWidth(s string) int {
	return runewidth.StringWidth(stripANSI(s))
}

// stripANSI removes ANSI escape sequences from a string
func stripANSI(s string) string {
	result := strings.Builder{}
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
}

// styleToBufferStyle converts a ComputedStyle to a buffer Style
func (lr *LayoutRenderer) styleToBufferStyle(st style.ComputedStyle) Style {
	return Style{
		FgColor:       st.Color,
		BgColor:       st.BgColor,
		Bold:          st.FontWeight == "bold",
		Italic:        st.FontStyle == "italic",
		Underline:     st.TextDecoration == "underline",
		Strikethrough: st.TextDecoration == "strikethrough",
		Dim:           st.Opacity > 0 && st.Opacity < 100,
		Reverse:       st.Reverse,
	}
}
