// Package render provides ANSI terminal rendering.
//
// This package is independent of vdom, style, and layout. It takes a LayoutBox
// tree (with computed positions/sizes) and produces ANSI escape sequences.
//
// This is where all terminal-specific output happens - colors, positioning, borders.
package render

import (
	"fmt"
	"strings"

	"github.com/speier/smith/pkg/lotus/layout"
	"github.com/speier/smith/pkg/lotus/style"
)

// Renderer converts a LayoutBox tree to ANSI terminal output
type Renderer struct {
	// Buffer for building output
	buf strings.Builder
}

// New creates a new renderer
func New() *Renderer {
	return &Renderer{}
}

// Render converts a layout tree to ANSI output
func (r *Renderer) Render(layout *layout.LayoutBox) string {
	r.buf.Reset()

	// Clear screen and move cursor to home
	r.buf.WriteString("\033[2J\033[H")

	// Render the tree
	r.renderBox(layout)

	return r.buf.String()
}

// renderBox renders a single layout box and its children
func (r *Renderer) renderBox(box *layout.LayoutBox) {
	if box == nil || box.Node == nil {
		return
	}

	style := box.Node.Style

	// Render border if present
	if style.Border {
		r.renderBorder(box, style)
	}

	// Render based on element type
	if box.Node.Element != nil {
		switch box.Node.Element.Type {
		case 1: // TextElement
			r.renderText(box, style)
		default:
			// Container - render children
			for _, child := range box.Children {
				r.renderBox(child)
			}
		}
	}
}

// renderBorder draws a border around a box
func (r *Renderer) renderBorder(box *layout.LayoutBox, st style.ComputedStyle) {
	if box.Width < 2 || box.Height < 2 {
		return // Too small for border
	}

	if !st.Border {
		return // No border to draw
	}

	// Choose border characters
	var topLeft, topRight, bottomLeft, bottomRight, horizontal, vertical string
	switch st.BorderStyle {
	case "rounded":
		topLeft, topRight = "╭", "╮"
		bottomLeft, bottomRight = "╰", "╯"
		horizontal, vertical = "─", "│"
	case "double":
		topLeft, topRight = "╔", "╗"
		bottomLeft, bottomRight = "╚", "╝"
		horizontal, vertical = "═", "║"
	default: // "single"
		topLeft, topRight = "┌", "┐"
		bottomLeft, bottomRight = "└", "┘"
		horizontal, vertical = "─", "│"
	}

	// Get color (use border-color if set, otherwise use text color)
	color := getANSIColor(st.Color)
	if st.BorderColor != "" {
		color = getANSIColor(st.BorderColor)
	}

	// Top border
	r.moveCursor(box.X, box.Y)
	r.buf.WriteString(color)
	r.buf.WriteString(topLeft)
	if box.Width > 2 {
		r.buf.WriteString(strings.Repeat(horizontal, box.Width-2))
	}
	r.buf.WriteString(topRight)
	r.buf.WriteString("\033[0m")

	// Side borders
	for i := 1; i < box.Height-1; i++ {
		// Left border
		r.moveCursor(box.X, box.Y+i)
		r.buf.WriteString(color)
		r.buf.WriteString(vertical)
		r.buf.WriteString("\033[0m")

		// Right border
		r.moveCursor(box.X+box.Width-1, box.Y+i)
		r.buf.WriteString(color)
		r.buf.WriteString(vertical)
		r.buf.WriteString("\033[0m")
	}

	// Bottom border
	r.moveCursor(box.X, box.Y+box.Height-1)
	r.buf.WriteString(color)
	r.buf.WriteString(bottomLeft)
	if box.Width > 2 {
		r.buf.WriteString(strings.Repeat(horizontal, box.Width-2))
	}
	r.buf.WriteString(bottomRight)
	r.buf.WriteString("\033[0m")
}

// renderText renders text content
func (r *Renderer) renderText(box *layout.LayoutBox, st style.ComputedStyle) {
	if box.Node.Element == nil {
		return
	}

	// Skip if hidden
	if st.Visibility == "hidden" {
		return
	}

	text := box.Node.Element.Text
	if text == "" {
		return
	}

	// Calculate position (accounting for border if present)
	x := box.X
	y := box.Y
	availableWidth := box.Width

	if st.Border {
		x++
		y++
		availableWidth -= 2
	}

	// Handle whitespace modes
	var lines []string
	switch st.WhiteSpace {
	case "pre":
		// Preserve all whitespace and newlines
		lines = strings.Split(text, "\n")
	case "nowrap":
		// No line breaks, single line
		text = strings.ReplaceAll(text, "\n", " ")
		lines = []string{text}
	default: // "normal"
		// Normal wrapping
		lines = strings.Split(text, "\n")
	}

	// Build ANSI style prefix
	stylePrefix := buildStylePrefix(st)
	color := getANSIColor(st.Color)
	bgColor := getANSIBgColor(st.BgColor)

	// Apply line clamping if MaxLines is set
	if st.MaxLines > 0 && len(lines) > st.MaxLines {
		// Keep only first MaxLines lines
		lines = lines[:st.MaxLines]

		// Add ellipsis to the last visible line
		lastIdx := len(lines) - 1
		lines[lastIdx] = lines[lastIdx] + "..."
	}

	for i, line := range lines {
		if i > 0 {
			y++ // Move to next line
		}

		lineX := x

		// Handle text alignment per line (use visible length for ANSI-colored text)
		lineVisibleLen := visibleLen(line)

		// Apply text overflow (ellipsis)
		if lineVisibleLen > availableWidth && st.TextOverflow == "ellipsis" {
			if availableWidth > 3 {
				line = truncateWithEllipsis(line, availableWidth)
				lineVisibleLen = availableWidth
			}
		}

		switch st.TextAlign {
		case "center":
			padding := (availableWidth - lineVisibleLen) / 2
			if padding > 0 {
				lineX += padding
			}
		case "right":
			padding := availableWidth - lineVisibleLen
			if padding > 0 {
				lineX += padding
			}
		}

		// Clip line to available width (using visible length)
		if lineVisibleLen > availableWidth {
			line = clipToWidth(line, availableWidth)
		}

		// Render line with styles
		r.moveCursor(lineX, y)
		r.buf.WriteString(stylePrefix)
		r.buf.WriteString(color)
		r.buf.WriteString(bgColor)
		r.buf.WriteString(line)
		r.buf.WriteString("\033[0m") // Reset all styles
	}
}

// moveCursor moves the cursor to (x, y) using ANSI escape codes
func (r *Renderer) moveCursor(x, y int) {
	// ANSI escape: ESC[{y};{x}H (1-indexed)
	fmt.Fprintf(&r.buf, "\033[%d;%dH", y+1, x+1)
}

// getANSIColor converts a hex color to ANSI foreground color
func getANSIColor(color string) string {
	if color == "" {
		return ""
	}

	// Basic color mapping (can be enhanced with true color support)
	colorMap := map[string]string{
		"#ffffff": "\033[97m", // bright white
		"#ffff00": "\033[93m", // bright yellow
		"#ff0000": "\033[91m", // bright red
		"#00ff00": "\033[92m", // bright green
		"#0000ff": "\033[94m", // bright blue
		"#ff00ff": "\033[95m", // bright magenta
		"#00ffff": "\033[96m", // bright cyan
		"#000000": "\033[30m", // black
		"#808080": "\033[90m", // gray
	}

	if ansi, ok := colorMap[strings.ToLower(color)]; ok {
		return ansi
	}

	// Default to white
	return "\033[97m"
}

// getANSIBgColor converts a hex color to ANSI background color
func getANSIBgColor(color string) string {
	if color == "" {
		return ""
	}

	// Basic color mapping
	colorMap := map[string]string{
		"#ffffff": "\033[107m", // bright white bg
		"#ffff00": "\033[103m", // bright yellow bg
		"#ff0000": "\033[101m", // bright red bg
		"#00ff00": "\033[102m", // bright green bg
		"#0000ff": "\033[104m", // bright blue bg
		"#ff00ff": "\033[105m", // bright magenta bg
		"#00ffff": "\033[106m", // bright cyan bg
		"#000000": "\033[40m",  // black bg
		"#808080": "\033[100m", // gray bg
	}

	if ansi, ok := colorMap[strings.ToLower(color)]; ok {
		return ansi
	}

	return ""
}

// visibleLen returns the visible character count (excluding ANSI escape codes)
func visibleLen(s string) int {
	count := 0
	inEscape := false

	for _, r := range s { // Iterate over runes, not bytes!
		if r == '\033' { // ESC character
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' { // End of ANSI sequence
				inEscape = false
			}
			continue
		}
		count++
	}

	return count
}

// buildStylePrefix builds ANSI escape codes for text styling
func buildStylePrefix(st style.ComputedStyle) string {
	var codes []string

	// Font weight (bold)
	if st.FontWeight == "bold" {
		codes = append(codes, "1")
	}

	// Opacity (dim)
	if st.Opacity > 0 && st.Opacity < 100 {
		codes = append(codes, "2") // dim
	}

	// Font style (italic)
	if st.FontStyle == "italic" {
		codes = append(codes, "3")
	}

	// Text decoration
	switch st.TextDecoration {
	case "underline":
		codes = append(codes, "4")
	case "strikethrough":
		codes = append(codes, "9")
	}

	// Reverse video
	if st.Reverse {
		codes = append(codes, "7")
	}

	if len(codes) == 0 {
		return ""
	}

	// Build escape sequence: \033[{code1};{code2};...m
	return "\033[" + strings.Join(codes, ";") + "m"
}

// truncateWithEllipsis truncates text to fit width and adds "..." at the end
func truncateWithEllipsis(text string, width int) string {
	if width < 1 {
		return ""
	}
	if width <= 3 {
		return strings.Repeat(".", width)
	}

	// Count runes for proper Unicode handling
	runes := []rune(text)
	if len(runes) <= width-3 {
		return text
	}

	// Truncate and add ellipsis
	return string(runes[:width-3]) + "..."
}

// clipToWidth clips text to fit within the specified width (rune-aware)
func clipToWidth(text string, width int) string {
	if width < 1 {
		return ""
	}

	runes := []rune(text)
	if len(runes) <= width {
		return text
	}

	return string(runes[:width])
}
