package render

import (
	"testing"
)

func TestComputeDiff_EmptyBuffers(t *testing.T) {
	old := NewBuffer(10, 5)
	new := NewBuffer(10, 5)

	diff := ComputeDiff(old, new)

	// Empty buffers should have no diff regions
	if len(diff.Regions) != 0 {
		t.Errorf("Expected 0 diff regions for identical empty buffers, got %d", len(diff.Regions))
	}
	if diff.FullRedraw {
		t.Error("Expected FullRedraw=false for identical buffers")
	}
}

func TestComputeDiff_SingleCellChange(t *testing.T) {
	old := NewBuffer(10, 5)
	new := NewBuffer(10, 5)

	// Change one cell
	new.Set(3, 2, Cell{Char: 'X', Style: Style{Bold: true}})

	diff := ComputeDiff(old, new)

	// Should have exactly 1 region
	if len(diff.Regions) != 1 {
		t.Fatalf("Expected 1 diff region, got %d", len(diff.Regions))
	}

	// Verify region
	r := diff.Regions[0]
	if r.Y != 2 {
		t.Errorf("Expected Y=2, got %d", r.Y)
	}
	if r.X != 3 {
		t.Errorf("Expected X=3, got %d", r.X)
	}
	if r.Width != 1 {
		t.Errorf("Expected Width=1, got %d", r.Width)
	}
	if r.Height != 1 {
		t.Errorf("Expected Height=1, got %d", r.Height)
	}
}

func TestComputeDiff_ConsecutiveCellsOnSameLine(t *testing.T) {
	old := NewBuffer(10, 5)
	new := NewBuffer(10, 5)

	// Change 3 consecutive cells
	new.Set(2, 1, Cell{Char: 'A'})
	new.Set(3, 1, Cell{Char: 'B'})
	new.Set(4, 1, Cell{Char: 'C'})

	diff := ComputeDiff(old, new)

	// Should merge into 1 region
	if len(diff.Regions) != 1 {
		t.Fatalf("Expected 1 merged region, got %d", len(diff.Regions))
	}

	r := diff.Regions[0]
	if r.Y != 1 {
		t.Errorf("Expected Y=1, got %d", r.Y)
	}
	if r.X != 2 {
		t.Errorf("Expected X=2, got %d", r.X)
	}
	if r.Width != 3 { // 3 cells wide
		t.Errorf("Expected Width=3, got %d", r.Width)
	}
}

func TestComputeDiff_MultipleLines(t *testing.T) {
	old := NewBuffer(10, 5)
	new := NewBuffer(10, 5)

	// Change cells on different lines
	new.Set(0, 0, Cell{Char: 'A'})
	new.Set(0, 2, Cell{Char: 'B'})
	new.Set(0, 4, Cell{Char: 'C'})

	diff := ComputeDiff(old, new)

	// Should have 3 separate regions (different lines)
	if len(diff.Regions) != 3 {
		t.Fatalf("Expected 3 diff regions, got %d", len(diff.Regions))
	}

	// Verify they're on correct lines
	expectedLines := map[int]bool{0: true, 2: true, 4: true}
	for _, r := range diff.Regions {
		if !expectedLines[r.Y] {
			t.Errorf("Unexpected region at line %d", r.Y)
		}
		delete(expectedLines, r.Y)
	}

	if len(expectedLines) > 0 {
		t.Errorf("Missing regions for lines: %v", expectedLines)
	}
}

func TestComputeDiff_StyleChangeOnly(t *testing.T) {
	old := NewBuffer(10, 5)
	new := NewBuffer(10, 5)

	// Both have 'A', but different styles
	old.Set(5, 2, Cell{Char: 'A', Style: Style{Bold: false}})
	new.Set(5, 2, Cell{Char: 'A', Style: Style{Bold: true}})

	diff := ComputeDiff(old, new)

	// Style change should trigger diff
	if len(diff.Regions) != 1 {
		t.Fatalf("Expected 1 diff region for style change, got %d", len(diff.Regions))
	}

	r := diff.Regions[0]
	if r.Y != 2 || r.X != 5 || r.Width != 1 {
		t.Errorf("Expected region at (5,2) with width 1, got (%d, %d) width %d", r.X, r.Y, r.Width)
	}
}

func TestComputeDiff_ColorChangeOnly(t *testing.T) {
	old := NewBuffer(10, 5)
	new := NewBuffer(10, 5)

	// Same char, different colors
	old.Set(3, 1, Cell{Char: 'X', Style: Style{FgColor: "red"}})
	new.Set(3, 1, Cell{Char: 'X', Style: Style{FgColor: "blue"}})

	diff := ComputeDiff(old, new)

	// Color change should trigger diff
	if len(diff.Regions) != 1 {
		t.Fatalf("Expected 1 diff region for color change, got %d", len(diff.Regions))
	}
}

func TestComputeDiff_WholeLine(t *testing.T) {
	old := NewBuffer(10, 5)
	new := NewBuffer(10, 5)

	// Change entire line
	for x := 0; x < 10; x++ {
		new.Set(x, 2, Cell{Char: rune('A' + x)})
	}

	diff := ComputeDiff(old, new)

	// Should be 1 region spanning entire line
	if len(diff.Regions) != 1 {
		t.Fatalf("Expected 1 diff region for whole line, got %d", len(diff.Regions))
	}

	r := diff.Regions[0]
	if r.Y != 2 {
		t.Errorf("Expected Y=2, got %d", r.Y)
	}
	if r.X != 0 {
		t.Errorf("Expected X=0, got %d", r.X)
	}
	if r.Width != 10 {
		t.Errorf("Expected Width=10, got %d", r.Width)
	}
}

func TestComputeDiff_NilOldBuffer(t *testing.T) {
	new := NewBuffer(10, 5)
	new.Set(5, 2, Cell{Char: 'X'})

	// Nil old buffer should trigger full redraw
	diff := ComputeDiff(nil, new)

	if !diff.FullRedraw {
		t.Error("Expected FullRedraw=true with nil old buffer")
	}
}

func TestComputeDiff_DifferentSizes(t *testing.T) {
	old := NewBuffer(5, 3)
	new := NewBuffer(10, 5)

	// Different sized buffers should trigger full redraw
	diff := ComputeDiff(old, new)

	if !diff.FullRedraw {
		t.Error("Expected FullRedraw=true for different sized buffers")
	}
}

func TestComputeDiff_FullScreenChange(t *testing.T) {
	old := NewBuffer(5, 3)
	new := NewBuffer(5, 3)

	// Change every cell
	for y := 0; y < 3; y++ {
		for x := 0; x < 5; x++ {
			new.Set(x, y, Cell{Char: rune('A' + x + y*5)})
		}
	}

	diff := ComputeDiff(old, new)

	// Regions may be merged (3 lines could become 1 merged region)
	// At minimum, we should have at least 1 region
	if len(diff.Regions) == 0 {
		t.Fatal("Expected at least 1 diff region for full screen change")
	}

	// Verify all 3 lines are covered
	// Calculate total area covered by all regions
	totalArea := 0
	for _, r := range diff.Regions {
		totalArea += r.Width * r.Height
	}

	// Should cover all 15 cells (5 wide x 3 tall)
	if totalArea < 15 {
		t.Errorf("Expected at least 15 cells covered, got %d", totalArea)
	}
}

func TestDiffCellEquality(t *testing.T) {
	// Test that equal cells don't create diff regions
	old := NewBuffer(5, 5)
	new := NewBuffer(5, 5)

	cell := Cell{
		Char: 'X',
		Style: Style{
			FgColor: "red",
			BgColor: "black",
			Bold:    true,
			Italic:  true,
		},
	}

	// Set same cell in both
	old.Set(2, 2, cell)
	new.Set(2, 2, cell)

	diff := ComputeDiff(old, new)

	// Should have no diff (cells are identical)
	if len(diff.Regions) != 0 {
		t.Errorf("Expected 0 diff regions for identical cells, got %d", len(diff.Regions))
	}
}

func TestComputeDiff_SparseChanges(t *testing.T) {
	old := NewBuffer(20, 10)
	new := NewBuffer(20, 10)

	// Make scattered changes
	new.Set(1, 1, Cell{Char: 'A'})
	new.Set(5, 3, Cell{Char: 'B'})
	new.Set(10, 5, Cell{Char: 'C'})
	new.Set(15, 8, Cell{Char: 'D'})

	diff := ComputeDiff(old, new)

	// Should have 4 regions (one per changed line)
	if len(diff.Regions) != 4 {
		t.Fatalf("Expected 4 diff regions for sparse changes, got %d", len(diff.Regions))
	}

	// Verify each region has height 1
	for i, r := range diff.Regions {
		if r.Height != 1 {
			t.Errorf("Region %d: expected Height=1, got %d", i, r.Height)
		}
		if r.Width < 1 {
			t.Errorf("Region %d: expected Width>=1, got %d", i, r.Width)
		}
	}
}

func TestComputeDiff_PartialLineChange(t *testing.T) {
	old := NewBuffer(20, 5)
	new := NewBuffer(20, 5)

	// Change middle section of a line
	for x := 5; x < 15; x++ {
		new.Set(x, 2, Cell{Char: rune('A' + (x - 5))})
	}

	diff := ComputeDiff(old, new)

	// Should have 1 region
	if len(diff.Regions) != 1 {
		t.Fatalf("Expected 1 diff region, got %d", len(diff.Regions))
	}

	r := diff.Regions[0]
	if r.X != 5 {
		t.Errorf("Expected X=5, got %d", r.X)
	}
	if r.Width != 10 { // cells 5-14 = 10 cells
		t.Errorf("Expected Width=10, got %d", r.Width)
	}
	if r.Y != 2 {
		t.Errorf("Expected Y=2, got %d", r.Y)
	}
}
