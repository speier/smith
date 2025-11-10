package renderer

import (
	"fmt"
	"strings"

	"github.com/mattn/go-runewidth"
)

// Cell represents a single terminal cell with content and style
type Cell struct {
	Rune  rune
	Style string // ANSI color codes
}

// Buffer represents a 2D terminal screen buffer
type Buffer struct {
	Width  int
	Height int
	Cells  [][]Cell
}

// NewBuffer creates a new terminal buffer
func NewBuffer(width, height int) *Buffer {
	cells := make([][]Cell, height)
	for i := range cells {
		cells[i] = make([]Cell, width)
		// Initialize with spaces
		for j := range cells[i] {
			cells[i][j] = Cell{Rune: ' ', Style: ""}
		}
	}
	return &Buffer{
		Width:  width,
		Height: height,
		Cells:  cells,
	}
}

// Clear resets all cells to empty spaces
func (b *Buffer) Clear() {
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			b.Cells[y][x] = Cell{Rune: ' ', Style: ""}
		}
	}
}

// Set writes a string at the given position with the given style
func (b *Buffer) Set(x, y int, text, style string) {
	if y < 0 || y >= b.Height {
		return
	}

	col := x
	for _, r := range text {
		if col >= b.Width {
			break
		}
		if col >= 0 {
			width := runewidth.RuneWidth(r)
			b.Cells[y][col] = Cell{Rune: r, Style: style}
			col++
			// For wide characters, mark next cell as continuation
			for i := 1; i < width; i++ {
				if col < b.Width {
					b.Cells[y][col] = Cell{Rune: 0, Style: style} // 0 = continuation
					col++
				}
			}
		} else {
			col++
		}
	}
}

// Diff compares this buffer with another and returns ANSI codes to update the screen
// Returns the minimal set of cursor movements and writes needed
func (b *Buffer) Diff(old *Buffer) string {
	if old == nil || old.Width != b.Width || old.Height != b.Height {
		// Full redraw needed
		return b.FullRender()
	}

	var buf strings.Builder
	currentStyle := ""

	for y := 0; y < b.Height; y++ {
		// Find runs of changed cells on this line
		x := 0
		for x < b.Width {
			// Skip unchanged cells
			for x < b.Width && b.Cells[y][x] == old.Cells[y][x] {
				x++
			}
			if x >= b.Width {
				break
			}

			// Found a changed cell - find the end of the changed run
			startX := x
			for x < b.Width && b.Cells[y][x] != old.Cells[y][x] {
				x++
			}

			// Move cursor to start of changed run
			buf.WriteString(fmt.Sprintf("\033[%d;%dH", y+1, startX+1))

			// Write the changed cells
			for i := startX; i < x; i++ {
				cell := b.Cells[y][i]

				// Only output style change if needed
				if cell.Style != currentStyle {
					if currentStyle != "" {
						buf.WriteString("\033[0m") // Reset first
					}
					buf.WriteString(cell.Style)
					currentStyle = cell.Style
				}

				// Output the character
				if cell.Rune != 0 {
					buf.WriteRune(cell.Rune)
				} else {
					// Output space for empty cells or wide char continuations
					buf.WriteRune(' ')
				}
			}
		}
	}

	// Reset style at end
	if currentStyle != "" {
		buf.WriteString("\033[0m")
	}

	return buf.String()
}

// FullRender returns ANSI codes to render the entire buffer from scratch
func (b *Buffer) FullRender() string {
	var buf strings.Builder

	// Clear screen and move to home
	buf.WriteString("\033[2J\033[H")

	currentStyle := ""

	for y := 0; y < b.Height; y++ {
		// Move to start of line
		buf.WriteString(fmt.Sprintf("\033[%d;1H", y+1))

		for x := 0; x < b.Width; x++ {
			cell := b.Cells[y][x]

			// Only output style change if needed
			if cell.Style != currentStyle {
				if currentStyle != "" {
					buf.WriteString("\033[0m")
				}
				buf.WriteString(cell.Style)
				currentStyle = cell.Style
			}

			// Output the character
			if cell.Rune != 0 {
				buf.WriteRune(cell.Rune)
			} else {
				// Output space for empty cells or wide char continuations
				buf.WriteRune(' ')
			}
		}
	}

	// Reset style at end
	if currentStyle != "" {
		buf.WriteString("\033[0m")
	}

	return buf.String()
}

// Clone creates a deep copy of the buffer
func (b *Buffer) Clone() *Buffer {
	clone := NewBuffer(b.Width, b.Height)
	for y := 0; y < b.Height; y++ {
		copy(clone.Cells[y], b.Cells[y])
	}
	return clone
}
