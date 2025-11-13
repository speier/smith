package render

import (
	"testing"
)

func TestNewBuffer(t *testing.T) {
	buf := NewBuffer(10, 5)

	if buf.Width != 10 {
		t.Errorf("Expected width=10, got %d", buf.Width)
	}
	if buf.Height != 5 {
		t.Errorf("Expected height=5, got %d", buf.Height)
	}

	// Verify default cells are spaces
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.Get(x, y)
			if cell.Char != ' ' {
				t.Errorf("Cell (%d,%d): expected ' ', got '%c'", x, y, cell.Char)
			}
		}
	}
}

func TestBufferSetGet(t *testing.T) {
	buf := NewBuffer(10, 5)

	// Set a cell
	cell := Cell{
		Char: 'A',
		Style: Style{
			FgColor: "red",
			BgColor: "black",
			Bold:    true,
		},
	}

	buf.Set(2, 3, cell)

	// Get the cell back
	got := buf.Get(2, 3)
	if got.Char != 'A' {
		t.Errorf("Expected char 'A', got '%c'", got.Char)
	}
	if got.Style.FgColor != "red" {
		t.Errorf("Expected FgColor 'red', got '%s'", got.Style.FgColor)
	}
	if !got.Style.Bold {
		t.Error("Expected Bold=true")
	}
}

func TestBufferOutOfBounds(t *testing.T) {
	buf := NewBuffer(10, 5)

	// Set/Get outside bounds should not panic
	cell := Cell{Char: 'X'}
	buf.Set(-1, 0, cell)
	buf.Set(0, -1, cell)
	buf.Set(100, 0, cell)
	buf.Set(0, 100, cell)

	// Get outside bounds should return space cell
	got := buf.Get(-1, 0)
	if got.Char != ' ' {
		t.Errorf("Out of bounds Get should return space, got %v", got)
	}
}

func TestBufferClone(t *testing.T) {
	original := NewBuffer(5, 3)
	original.Set(1, 1, Cell{Char: 'A', Style: Style{Bold: true}})
	original.Set(2, 2, Cell{Char: 'B', Style: Style{FgColor: "red"}})

	// Clone the buffer
	clone := original.Clone()

	// Verify dimensions
	if clone.Width != original.Width || clone.Height != original.Height {
		t.Errorf("Clone dimensions don't match: original=%dx%d, clone=%dx%d",
			original.Width, original.Height, clone.Width, clone.Height)
	}

	// Verify cell contents
	for y := 0; y < original.Height; y++ {
		for x := 0; x < original.Width; x++ {
			origCell := original.Get(x, y)
			cloneCell := clone.Get(x, y)
			if origCell != cloneCell {
				t.Errorf("Cell (%d, %d) differs: original=%v, clone=%v",
					x, y, origCell, cloneCell)
			}
		}
	}

	// Modify clone - should not affect original
	clone.Set(1, 1, Cell{Char: 'Z'})
	if original.Get(1, 1).Char == 'Z' {
		t.Error("Modifying clone affected original buffer")
	}
}

func TestBufferClip(t *testing.T) {
	// Create 10x10 buffer with numbered content
	buf := NewBuffer(10, 10)
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			// Use coordinates as identifier
			buf.Set(x, y, Cell{Char: rune('0' + (x % 10))})
		}
	}

	// Clip to 5x5 region starting at (2, 3)
	clipped := buf.Clip(2, 3, 5, 5)

	// Verify clipped dimensions
	if clipped.Width != 5 || clipped.Height != 5 {
		t.Errorf("Clipped buffer: expected 5x5, got %dx%d", clipped.Width, clipped.Height)
	}

	// Verify clipped content
	// Original (2, 3) should be at clipped (0, 0)
	expected := buf.Get(2, 3)
	got := clipped.Get(0, 0)
	if got != expected {
		t.Errorf("Clip(0,0): expected %v, got %v", expected, got)
	}

	// Original (6, 7) should be at clipped (4, 4)
	expected = buf.Get(6, 7)
	got = clipped.Get(4, 4)
	if got != expected {
		t.Errorf("Clip(4,4): expected %v, got %v", expected, got)
	}
}

func TestBufferClipOutOfBounds(t *testing.T) {
	buf := NewBuffer(10, 10)

	// Fill buffer with identifiable content
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			buf.Set(x, y, Cell{Char: rune('A' + x)})
		}
	}

	// Clip region extends beyond buffer
	clipped := buf.Clip(8, 8, 5, 5)

	// Should create 5x5 buffer, but only (0,0) and (1,0), (0,1), (1,1) have content
	if clipped.Width != 5 || clipped.Height != 5 {
		t.Errorf("Clipped buffer: expected 5x5, got %dx%d", clipped.Width, clipped.Height)
	}

	// Verify (0,0) has content from original (8,8)
	if clipped.Get(0, 0).Char == ' ' {
		t.Error("Expected clipped (0,0) to have content from original (8,8)")
	}

	// Verify (4,4) is empty (out of bounds in original)
	if clipped.Get(4, 4).Char != ' ' {
		t.Error("Expected clipped (4,4) to be empty (out of bounds)")
	}
}

func TestBufferClipNegativeOffset(t *testing.T) {
	buf := NewBuffer(10, 10)

	// Negative offset should clamp to 0
	clipped := buf.Clip(-5, -3, 5, 5)

	if clipped.Width != 5 || clipped.Height != 5 {
		t.Errorf("Clipped buffer: expected 5x5, got %dx%d", clipped.Width, clipped.Height)
	}

	// Should start from (0, 0)
	expected := buf.Get(0, 0)
	got := clipped.Get(0, 0)
	if got != expected {
		t.Error("Negative offset clip should start at (0,0)")
	}
}

func TestCellEquality(t *testing.T) {
	cell1 := Cell{
		Char: 'A',
		Style: Style{
			FgColor: "red",
			BgColor: "black",
			Bold:    true,
			Italic:  false,
		},
	}

	cell2 := cell1 // Copy

	if cell1 != cell2 {
		t.Error("Identical cells should be equal")
	}

	// Change one field
	cell2.Char = 'B'
	if cell1 == cell2 {
		t.Error("Different chars should make cells unequal")
	}

	// Test color difference
	cell3 := cell1
	cell3.Style.FgColor = "green"
	if cell1 == cell3 {
		t.Error("Different FgColors should make cells unequal")
	}

	// Test style difference
	cell4 := cell1
	cell4.Style.Bold = false
	if cell1 == cell4 {
		t.Error("Different Bold values should make cells unequal")
	}
}

func TestBufferFill(t *testing.T) {
	buf := NewBuffer(5, 3)

	// Fill with specific cell
	fillCell := Cell{
		Char: '.',
		Style: Style{
			FgColor: "gray",
		},
	}

	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			buf.Set(x, y, fillCell)
		}
	}

	// Verify all cells filled
	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			got := buf.Get(x, y)
			if got != fillCell {
				t.Errorf("Cell (%d, %d): expected %v, got %v", x, y, fillCell, got)
			}
		}
	}
}

func TestBufferUnicodeRunes(t *testing.T) {
	buf := NewBuffer(10, 5)

	// Test various Unicode characters
	unicodeRunes := []rune{'😀', '中', 'Ä', '∑', '♠'}

	for i, r := range unicodeRunes {
		buf.Set(i, 0, Cell{Char: r})
	}

	// Verify they're stored correctly
	for i, expected := range unicodeRunes {
		got := buf.Get(i, 0)
		if got.Char != expected {
			t.Errorf("Cell %d: expected rune %c, got %c", i, expected, got.Char)
		}
	}
}

func TestBufferWithStyles(t *testing.T) {
	buf := NewBuffer(10, 5)

	// Set cell with complex styling
	styledCell := Cell{
		Char: 'A',
		Style: Style{
			FgColor: "#FF0000",
			BgColor: "#000000",
			Bold:    true,
			Italic:  true,
		},
	}

	buf.Set(3, 2, styledCell)

	// Verify all style properties
	got := buf.Get(3, 2)
	if got.Char != 'A' {
		t.Errorf("Expected char 'A', got '%c'", got.Char)
	}
	if got.Style.FgColor != "#FF0000" {
		t.Errorf("Expected FgColor '#FF0000', got '%s'", got.Style.FgColor)
	}
	if got.Style.BgColor != "#000000" {
		t.Errorf("Expected BgColor '#000000', got '%s'", got.Style.BgColor)
	}
	if !got.Style.Bold {
		t.Error("Expected Bold=true")
	}
	if !got.Style.Italic {
		t.Error("Expected Italic=true")
	}
}

func TestBufferAsString(t *testing.T) {
	buf := NewBuffer(5, 3)

	// Set some content
	buf.Set(0, 0, Cell{Char: 'H'})
	buf.Set(1, 0, Cell{Char: 'e'})
	buf.Set(2, 0, Cell{Char: 'l'})
	buf.Set(3, 0, Cell{Char: 'l'})
	buf.Set(4, 0, Cell{Char: 'o'})

	buf.Set(0, 1, Cell{Char: 'W'})
	buf.Set(1, 1, Cell{Char: 'o'})
	buf.Set(2, 1, Cell{Char: 'r'})
	buf.Set(3, 1, Cell{Char: 'l'})
	buf.Set(4, 1, Cell{Char: 'd'})

	// Extract first row
	row0 := ""
	for x := 0; x < buf.Width; x++ {
		row0 += string(buf.Get(x, 0).Char)
	}
	if row0 != "Hello" {
		t.Errorf("Row 0: expected 'Hello', got '%s'", row0)
	}

	// Extract second row
	row1 := ""
	for x := 0; x < buf.Width; x++ {
		row1 += string(buf.Get(x, 1).Char)
	}
	if row1 != "World" {
		t.Errorf("Row 1: expected 'World', got '%s'", row1)
	}
}
