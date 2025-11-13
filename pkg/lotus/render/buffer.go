package render

import "strings"

// Cell represents a single terminal cell with character and style information
type Cell struct {
	Char  rune
	Style Style
}

// Style holds the ANSI styling for a cell
type Style struct {
	FgColor       string
	BgColor       string
	Bold          bool
	Italic        bool
	Underline     bool
	Strikethrough bool
	Dim           bool
	Reverse       bool
}

// Buffer represents a 2D grid of terminal cells
type Buffer struct {
	Width  int
	Height int
	cells  [][]Cell
}

// NewBuffer creates a new buffer with the given dimensions
func NewBuffer(width, height int) *Buffer {
	cells := make([][]Cell, height)
	for y := 0; y < height; y++ {
		cells[y] = make([]Cell, width)
		for x := 0; x < width; x++ {
			cells[y][x] = Cell{Char: ' ', Style: Style{}}
		}
	}
	return &Buffer{
		Width:  width,
		Height: height,
		cells:  cells,
	}
}

// Set sets a cell at the given position
func (b *Buffer) Set(x, y int, cell Cell) {
	if x >= 0 && x < b.Width && y >= 0 && y < b.Height {
		b.cells[y][x] = cell
	}
}

// Get retrieves a cell at the given position
func (b *Buffer) Get(x, y int) Cell {
	if x >= 0 && x < b.Width && y >= 0 && y < b.Height {
		return b.cells[y][x]
	}
	return Cell{Char: ' ', Style: Style{}}
}

// Clip returns a new buffer that is a clipped view of this buffer
func (b *Buffer) Clip(x, y, width, height int) *Buffer {
	clipped := NewBuffer(width, height)
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			srcX := x + col
			srcY := y + row
			if srcX >= 0 && srcX < b.Width && srcY >= 0 && srcY < b.Height {
				clipped.Set(col, row, b.Get(srcX, srcY))
			}
		}
	}
	return clipped
}

// Clear fills the entire buffer with spaces
func (b *Buffer) Clear() {
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			b.cells[y][x] = Cell{Char: ' ', Style: Style{}}
		}
	}
}

// WriteString writes a string at the given position with the given style
func (b *Buffer) WriteString(x, y int, text string, style Style) {
	col := x
	for _, ch := range text {
		if ch == '\n' {
			y++
			col = x
			continue
		}
		if col < b.Width && y < b.Height {
			b.Set(col, y, Cell{Char: ch, Style: style})
			col++
		}
	}
}

// ToString converts the buffer to a string representation (for debugging)
func (b *Buffer) ToString() string {
	var sb strings.Builder
	for y := 0; y < b.Height; y++ {
		for x := 0; x < b.Width; x++ {
			cell := b.Get(x, y)
			sb.WriteRune(cell.Char)
		}
		if y < b.Height-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

// Clone creates a deep copy of the buffer
func (b *Buffer) Clone() *Buffer {
	clone := NewBuffer(b.Width, b.Height)
	for y := 0; y < b.Height; y++ {
		copy(clone.cells[y], b.cells[y])
	}
	return clone
}
