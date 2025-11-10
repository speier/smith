package renderer

import (
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/speier/smith/pkg/lotus/core"
)

// BufferedRenderer renders using double buffering and diffing
type BufferedRenderer struct {
	width    int
	height   int
	current  *Buffer // Current screen state
	previous *Buffer // Previous screen state
}

// NewBufferedRenderer creates a new buffered renderer
func NewBufferedRenderer(width, height int) *BufferedRenderer {
	return &BufferedRenderer{
		width:    width,
		height:   height,
		current:  NewBuffer(width, height),
		previous: nil, // No previous state yet
	}
}

// Render renders the layout tree and returns minimal ANSI diff
func (r *BufferedRenderer) Render(root *core.Node, forceFullRender bool) string {
	// Clear current buffer
	r.current.Clear()

	// Render tree into buffer
	r.renderNodeToBuffer(root)

	// Generate diff or full render
	var output string
	if forceFullRender || r.previous == nil {
		output = r.current.FullRender()
	} else {
		output = r.current.Diff(r.previous)
	}

	// Save current as previous for next render
	r.previous = r.current.Clone()

	return output
}

// Resize updates buffer dimensions
func (r *BufferedRenderer) Resize(width, height int) {
	if r.width != width || r.height != height {
		r.width = width
		r.height = height
		r.current = NewBuffer(width, height)
		r.previous = nil // Force full redraw on next render
	}
}

func (r *BufferedRenderer) renderNodeToBuffer(node *core.Node) {
	if node == nil {
		return
	}

	// Render border if present
	if node.Styles.Border {
		r.renderBorder(node)
	}

	// Render content
	switch node.Type {
	case "text":
		r.renderText(node)
	case "input":
		r.renderInput(node)
	default:
		// Render children
		for _, child := range node.Children {
			r.renderNodeToBuffer(child)
		}
	}
}

func (r *BufferedRenderer) renderBorder(node *core.Node) {
	// Safety check: ensure border can fit
	if node.Width < 2 || node.Height < 2 {
		return
	}
	if node.X < 0 || node.Y < 0 {
		return
	}
	// Don't reject - clip to buffer bounds instead

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

	style := getANSIColor(node.Styles.Color)

	// Calculate safe width for horizontal borders
	safeWidth := node.Width
	if node.X+safeWidth > r.width {
		safeWidth = r.width - node.X
	}

	// Top border
	if safeWidth >= 2 {
		line := topLeft + strings.Repeat(horizontal, safeWidth-2) + topRight
		r.current.Set(node.X, node.Y, line, style)
	}

	// Side borders
	for i := 1; i < node.Height-1; i++ {
		if node.Y+i >= r.height {
			break
		}
		// Left border
		r.current.Set(node.X, node.Y+i, vertical, style)
		// Right border - ensure it's within bounds
		rightX := node.X + safeWidth - 1
		if rightX < r.width {
			r.current.Set(rightX, node.Y+i, vertical, style)
		}
	}

	// Bottom border
	if node.Y+node.Height-1 < r.height && safeWidth >= 2 {
		line := bottomLeft + strings.Repeat(horizontal, safeWidth-2) + bottomRight
		r.current.Set(node.X, node.Y+node.Height-1, line, style)
	}
}

func (r *BufferedRenderer) renderText(node *core.Node) {
	if node.Content == "" {
		return
	}

	colorStyle := getANSIColor(node.Styles.Color)
	bgStyle := getANSIBgColor(node.Styles.BgColor)
	style := colorStyle + bgStyle

	// Calculate text position based on alignment
	x := node.X
	text := node.Content
	availableWidth := node.Width

	// Handle alignment
	switch node.Styles.TextAlign {
	case "center":
		// Strip ANSI codes before calculating width
		plainText := ansiRegex.ReplaceAllString(text, "")
		textWidth := runewidth.StringWidth(plainText)
		padding := (availableWidth - textWidth) / 2
		if padding > 0 {
			x += padding
		}
	case "right":
		// Strip ANSI codes before calculating width
		plainText := ansiRegex.ReplaceAllString(text, "")
		textWidth := runewidth.StringWidth(plainText)
		padding := availableWidth - textWidth
		if padding > 0 {
			x += padding
		}
	}

	r.current.Set(x, node.Y, text, style)
}

func (r *BufferedRenderer) renderInput(node *core.Node) {
	prompt := node.Attributes["prompt"]
	if prompt == "" {
		prompt = "> "
	}

	style := getANSIColor(node.Styles.Color)

	// Position for input
	x := node.X
	if node.Parent != nil && node.Parent.Styles.Border {
		x += 1 + node.Parent.Styles.PaddingLeft
	}

	// Render prompt
	r.current.Set(x, node.Y, prompt, style)

	// Input content is in child text nodes, they'll be rendered separately
}
