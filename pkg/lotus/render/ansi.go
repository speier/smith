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

	// Get color
	color := getANSIColor(st.Color)

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

	// Split text by newlines and render each line
	lines := strings.Split(text, "\n")

	color := getANSIColor(st.Color)
	bgColor := getANSIBgColor(st.BgColor)

	for i, line := range lines {
		if i > 0 {
			y++ // Move to next line
		}

		lineX := x

		// Handle text alignment per line
		switch st.TextAlign {
		case "center":
			padding := (availableWidth - len(line)) / 2
			if padding > 0 {
				lineX += padding
			}
		case "right":
			padding := availableWidth - len(line)
			if padding > 0 {
				lineX += padding
			}
		}

		// Clip line to available width
		if len(line) > availableWidth {
			line = line[:availableWidth]
		}

		// Render line
		r.moveCursor(lineX, y)
		r.buf.WriteString(color)
		r.buf.WriteString(bgColor)
		r.buf.WriteString(line)
		r.buf.WriteString("\033[0m")
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
