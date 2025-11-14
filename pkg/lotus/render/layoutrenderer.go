package render

import (
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/speier/smith/pkg/lotus/layout"
	"github.com/speier/smith/pkg/lotus/style"
)

// LayoutRenderer converts a LayoutBox tree into a Buffer
type LayoutRenderer struct{}

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

	// Check if this is a ScrollView component - handle specially
	if box.Node.Element != nil && box.Node.Element.Component != nil {
		// Try to get ScrollView interface for viewport clipping
		type ScrollViewInterface interface {
			GetScrollOffset() (int, int)
			GetViewportSize() (int, int)
		}
		if sv, ok := box.Node.Element.Component.(ScrollViewInterface); ok {
			// Render children to temporary buffer
			tempBuf := NewBuffer(buf.Width, buf.Height)

			// Render children with coordinates adjusted relative to ScrollView origin
			offsetX := box.X
			offsetY := box.Y

			for _, child := range box.Children {
				// Recursively adjust entire tree to relative coordinates
				adjustedChild := lr.adjustLayoutTree(child, offsetX, offsetY)
				lr.renderBox(tempBuf, adjustedChild)
			}

			// Get scroll offset
			scrollX, scrollY := sv.GetScrollOffset()

			// Clip the temp buffer to viewport and copy to main buffer
			_, viewportHeight := sv.GetViewportSize()
			clipped := tempBuf.Clip(scrollX, scrollY, box.Width, viewportHeight)

			// Copy clipped buffer to main buffer at box position
			for y := 0; y < clipped.Height; y++ {
				for x := 0; x < clipped.Width; x++ {
					cell := clipped.Get(x, y)
					buf.Set(box.X+x, box.Y+y, cell)
				}
			}
			return
		}
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

		// Write line to buffer
		for _, ch := range line {
			charWidth := runewidth.RuneWidth(ch)
			// Check if there's enough space for the full character width
			if x+charWidth > contentX+contentWidth {
				break // Don't exceed content width
			}
			buf.Set(x, y, Cell{Char: ch, Style: bufStyle})

			// For wide characters (emoji, CJK), mark the next cell as occupied
			// Use a zero-width space (U+200B) as a placeholder
			if charWidth == 2 {
				buf.Set(x+1, y, Cell{Char: '\u200B', Style: bufStyle})
			}

			// Increment by actual character width (emojis = 2, normal = 1)
			x += charWidth
		}
	}
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

// displayWidth calculates the display width of a string (for now, just rune count)
func (lr *LayoutRenderer) displayWidth(s string) int {
	return runewidth.StringWidth(s)
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
