package terminal

import (
	"fmt"
	"strings"

	"github.com/speier/smith/pkg/lotus/layout"
)

// Render renders the node tree to terminal using ANSI codes
func Render(root *layout.Node) string {
	var buf strings.Builder

	// Clear screen
	buf.WriteString("\033[2J\033[H")

	// Render the tree
	renderNode(root, &buf)

	return buf.String()
}

func renderNode(node *layout.Node, buf *strings.Builder) {
	if node == nil {
		return
	}

	// Render border if present
	if node.Styles.Border {
		renderBorder(node, buf)
	}

	// Render content
	switch node.Type {
	case "text":
		renderText(node, buf)
	case "input":
		renderInput(node, buf)
	default:
		// Render children
		for _, child := range node.Children {
			renderNode(child, buf)
		}
	}
}

func renderBorder(node *layout.Node, buf *strings.Builder) {
	var topLeft, topRight, bottomLeft, bottomRight, horizontal, vertical string

	switch node.Styles.BorderChar {
	case "rounded":
		topLeft, topRight = "╭", "╮"
		bottomLeft, bottomRight = "╰", "╯"
		horizontal, vertical = "─", "│"
	case "double":
		topLeft, topRight = "╔", "╗"
		bottomLeft, bottomRight = "╚", "╝"
		horizontal, vertical = "═", "║"
	default: // single
		topLeft, topRight = "┌", "┐"
		bottomLeft, bottomRight = "└", "┘"
		horizontal, vertical = "─", "│"
	}

	// Apply color
	color := getANSIColor(node.Styles.Color)

	// Top border
	moveCursor(node.X, node.Y, buf)
	buf.WriteString(color)
	buf.WriteString(topLeft)
	buf.WriteString(strings.Repeat(horizontal, node.Width-2))
	buf.WriteString(topRight)
	buf.WriteString("\033[0m") // reset

	// Side borders
	for i := 1; i < node.Height-1; i++ {
		// Left border
		moveCursor(node.X, node.Y+i, buf)
		buf.WriteString(color)
		buf.WriteString(vertical)
		buf.WriteString("\033[0m")

		// Right border
		moveCursor(node.X+node.Width-1, node.Y+i, buf)
		buf.WriteString(color)
		buf.WriteString(vertical)
		buf.WriteString("\033[0m")
	}

	// Bottom border
	moveCursor(node.X, node.Y+node.Height-1, buf)
	buf.WriteString(color)
	buf.WriteString(bottomLeft)
	buf.WriteString(strings.Repeat(horizontal, node.Width-2))
	buf.WriteString(bottomRight)
	buf.WriteString("\033[0m")
}

func renderText(node *layout.Node, buf *strings.Builder) {
	if node.Content == "" {
		return
	}

	color := getANSIColor(node.Styles.Color)
	bgColor := getANSIBgColor(node.Styles.BgColor)

	// Calculate text position based on alignment
	x := node.X
	text := node.Content
	availableWidth := node.Width

	if node.Parent != nil && node.Parent.Styles.Border {
		availableWidth = node.Width - 2
		x++
	}

	switch node.Styles.TextAlign {
	case "center":
		padding := (availableWidth - len(text)) / 2
		if padding > 0 {
			x += padding
		}
	case "right":
		padding := availableWidth - len(text)
		if padding > 0 {
			x += padding
		}
	}

	moveCursor(x, node.Y, buf)
	buf.WriteString(color)
	buf.WriteString(bgColor)
	buf.WriteString(text)
	buf.WriteString("\033[0m")
}

func renderInput(node *layout.Node, buf *strings.Builder) {
	prompt := node.Attributes["prompt"]
	if prompt == "" {
		prompt = "> "
	}

	color := getANSIColor(node.Styles.Color)

	// Position cursor for input
	x := node.X
	if node.Parent != nil && node.Parent.Styles.Border {
		x += 1 + node.Parent.Styles.PaddingLeft
	}

	moveCursor(x, node.Y, buf)
	buf.WriteString(color)
	buf.WriteString(prompt)
	buf.WriteString("\033[0m")

	// Position cursor after prompt for input
	moveCursor(x+len(prompt), node.Y, buf)
}

func moveCursor(x, y int, buf *strings.Builder) {
	// ANSI escape: ESC[{y};{x}H (1-indexed)
	fmt.Fprintf(buf, "\033[%d;%dH", y+1, x+1)
}

func getANSIColor(color string) string {
	if color == "" {
		return ""
	}

	// Simple color mapping (could be expanded)
	colorMap := map[string]string{
		"#0f0":    "\033[92m", // bright green
		"#0ff":    "\033[96m", // bright cyan
		"#00ff00": "\033[92m", // bright green
		"#00ffff": "\033[96m", // bright cyan
		"#fff":    "\033[97m", // bright white
		"#ffffff": "\033[97m", // bright white
		"#f00":    "\033[91m", // bright red
		"#ff0000": "\033[91m", // bright red
		"#ff0":    "\033[93m", // bright yellow
		"#ffff00": "\033[93m", // bright yellow
		"#00f":    "\033[94m", // bright blue
		"#0000ff": "\033[94m", // bright blue
		"#444":    "\033[90m", // dark gray
		"#888":    "\033[37m", // light gray
	}

	if code, ok := colorMap[color]; ok {
		return code
	}

	return ""
}

func getANSIBgColor(color string) string {
	if color == "" {
		return ""
	}

	// Simple background color mapping
	colorMap := map[string]string{
		"#000":    "\033[40m",  // black
		"#000000": "\033[40m",  // black
		"#f00":    "\033[101m", // bright red
		"#ff0000": "\033[101m", // bright red
		"#0f0":    "\033[102m", // bright green
		"#00ff00": "\033[102m", // bright green
	}

	if code, ok := colorMap[color]; ok {
		return code
	}

	return ""
}
