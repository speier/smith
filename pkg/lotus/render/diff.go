package render

// DiffRegion represents a rectangular region that has changed between two buffers
type DiffRegion struct {
	X      int
	Y      int
	Width  int
	Height int
}

// DiffResult contains the differences between two buffers
type DiffResult struct {
	Regions []DiffRegion
	// Track if buffers are completely different (e.g., size changed)
	FullRedraw bool
}

// ComputeDiff compares two buffers and returns the regions that differ
func ComputeDiff(prev, curr *Buffer) *DiffResult {
	if prev == nil {
		return &DiffResult{FullRedraw: true}
	}

	// If dimensions changed, full redraw
	if prev.Width != curr.Width || prev.Height != curr.Height {
		return &DiffResult{FullRedraw: true}
	}

	result := &DiffResult{
		Regions:    make([]DiffRegion, 0),
		FullRedraw: false,
	}

	// Scan buffer line by line to find changed regions
	// We use a simple strategy: find contiguous vertical spans of changes
	for y := 0; y < curr.Height; y++ {
		lineChanged := false
		minX := curr.Width
		maxX := -1

		// Check if this line has any changes
		for x := 0; x < curr.Width; x++ {
			prevCell := prev.Get(x, y)
			currCell := curr.Get(x, y)

			if !cellsEqual(prevCell, currCell) {
				lineChanged = true
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
			}
		}

		// If line changed, create a region for it
		if lineChanged {
			result.Regions = append(result.Regions, DiffRegion{
				X:      minX,
				Y:      y,
				Width:  maxX - minX + 1,
				Height: 1,
			})
		}
	}

	// Merge adjacent horizontal regions to reduce ANSI escape sequences
	result.Regions = mergeRegions(result.Regions)

	return result
}

// cellsEqual checks if two cells are identical
func cellsEqual(a, b Cell) bool {
	return a.Char == b.Char &&
		a.Style.FgColor == b.Style.FgColor &&
		a.Style.BgColor == b.Style.BgColor &&
		a.Style.Bold == b.Style.Bold &&
		a.Style.Italic == b.Style.Italic
}

// mergeRegions merges adjacent or overlapping regions
func mergeRegions(regions []DiffRegion) []DiffRegion {
	if len(regions) <= 1 {
		return regions
	}

	merged := make([]DiffRegion, 0, len(regions))
	current := regions[0]

	for i := 1; i < len(regions); i++ {
		next := regions[i]

		// If next region is on the same line or adjacent line, try to merge
		if next.Y == current.Y || next.Y == current.Y+current.Height {
			// Check if regions overlap or are adjacent horizontally
			if next.Y == current.Y && next.X <= current.X+current.Width {
				// Same line, merge horizontally
				endX := max(current.X+current.Width, next.X+next.Width)
				current.X = min(current.X, next.X)
				current.Width = endX - current.X
			} else if next.Y == current.Y+current.Height && next.X == current.X && next.Width == current.Width {
				// Adjacent lines with same X and Width, merge vertically
				current.Height += next.Height
			} else {
				// Can't merge, save current and start new
				merged = append(merged, current)
				current = next
			}
		} else {
			// Not adjacent, save current and start new
			merged = append(merged, current)
			current = next
		}
	}

	// Add final region
	merged = append(merged, current)

	return merged
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
