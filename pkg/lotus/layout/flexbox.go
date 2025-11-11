// Package layout provides a pure flexbox layout engine.
//
// This is independent of vdom, styling, and rendering. It takes styled nodes
// and container dimensions, and returns layout boxes with computed positions/sizes.
//
// This is PURE MATH - no side effects, no mutations, just calculations.
// Similar to Yoga (React Native's layout engine).
package layout

import (
	"strconv"
	"strings"

	"github.com/speier/smith/pkg/lotus/style"
)

// LayoutBox represents a computed layout (position + size)
// This is the OUTPUT of layout calculation
type LayoutBox struct {
	// Computed position (relative to parent)
	X, Y int

	// Computed size
	Width, Height int

	// Reference to styled node
	Node *style.StyledNode

	// Laid out children
	Children []*LayoutBox
}

// Compute performs flexbox layout calculation
// This is a PURE function - no side effects!
//
// Input: StyledNode tree + container dimensions
// Output: LayoutBox tree with computed positions/sizes
func Compute(node *style.StyledNode, containerWidth, containerHeight int) *LayoutBox {
	box := &LayoutBox{
		X:      0,
		Y:      0,
		Width:  containerWidth,
		Height: containerHeight,
		Node:   node,
	}

	// Layout children
	layoutChildren(box, node)

	return box
}

// layoutChildren computes layout for all children
func layoutChildren(box *LayoutBox, node *style.StyledNode) {
	if len(node.Children) == 0 {
		return
	}

	// Calculate content area (after padding and border)
	contentX, contentY, contentWidth, contentHeight := computeContentBox(box, node.Style)

	// Layout based on display mode
	if node.Style.Display == "flex" {
		if node.Style.FlexDir == "row" {
			box.Children = layoutFlexRow(node.Children, contentX, contentY, contentWidth, contentHeight)
		} else {
			box.Children = layoutFlexColumn(node.Children, contentX, contentY, contentWidth, contentHeight)
		}
	} else {
		box.Children = layoutBlock(node.Children, contentX, contentY, contentWidth, contentHeight)
	}
}

// computeContentBox calculates the content area after padding/border
func computeContentBox(box *LayoutBox, style style.ComputedStyle) (x, y, width, height int) {
	x = box.X + style.PaddingLeft + style.MarginLeft
	y = box.Y + style.PaddingTop + style.MarginTop
	width = box.Width - style.PaddingLeft - style.PaddingRight
	height = box.Height - style.PaddingTop - style.PaddingBottom

	// Account for border
	if style.Border {
		x++
		y++
		width -= 2
		height -= 2
	}

	return x, y, width, height
}

// layoutFlexRow handles horizontal flexbox layout
func layoutFlexRow(children []*style.StyledNode, x, y, width, height int) []*LayoutBox {
	// Phase 1: Calculate fixed and flexible space
	var totalFixed int
	var totalFlexGrow int

	for _, child := range children {
		flexGrow := child.Style.FlexGrow

		if flexGrow > 0 {
			totalFlexGrow += flexGrow
			// Margins still count as fixed space
			totalFixed += child.Style.MarginLeft + child.Style.MarginRight
		} else {
			// Fixed width child
			childWidth := resolveDimension(child.Style.Width, width)
			if childWidth > 0 {
				totalFixed += childWidth + child.Style.MarginLeft + child.Style.MarginRight
			} else {
				// No explicit width and no flex-grow - use intrinsic width
				intrinsicWidth := computeIntrinsicWidth(child)
				totalFixed += intrinsicWidth + child.Style.MarginLeft + child.Style.MarginRight
			}
		}
	}

	// Remaining space for flex children
	remainingWidth := width - totalFixed
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	// Phase 2: Lay out each child
	boxes := make([]*LayoutBox, len(children))
	currentX := x

	for i, child := range children {
		flexGrow := child.Style.FlexGrow

		// Calculate child width
		var childWidth int
		if flexGrow > 0 && totalFlexGrow > 0 {
			// Flexible child - distribute remaining space proportionally
			childWidth = (remainingWidth * flexGrow) / totalFlexGrow
		} else {
			// Fixed width
			childWidth = resolveDimension(child.Style.Width, width)
			if childWidth <= 0 {
				// Use intrinsic width based on content
				childWidth = computeIntrinsicWidth(child)
			}
		}

		// Calculate child height - respect align-self
		childHeight := resolveDimension(child.Style.Height, height)
		if childHeight <= 0 {
			if child.Style.AlignSelf == "stretch" {
				// Stretch to fill container height (cross-axis in row)
				childHeight = height - child.Style.MarginTop - child.Style.MarginBottom
			} else {
				// Use intrinsic height
				childHeight = computeIntrinsicHeight(child)
			}
		}

		// Create layout box for this child
		childBox := &LayoutBox{
			X:      currentX + child.Style.MarginLeft,
			Y:      y + child.Style.MarginTop,
			Width:  childWidth,
			Height: childHeight,
			Node:   child,
		}

		// Recursively layout this child's children
		layoutChildren(childBox, child)

		boxes[i] = childBox

		// Move to next position
		currentX += childWidth + child.Style.MarginLeft + child.Style.MarginRight
	}

	return boxes
}

// layoutFlexColumn handles vertical flexbox layout
func layoutFlexColumn(children []*style.StyledNode, x, y, width, height int) []*LayoutBox {
	// Phase 1: Calculate fixed and flexible space
	var totalFixed int
	var totalFlexGrow int

	for _, child := range children {
		flexGrow := child.Style.FlexGrow

		if flexGrow > 0 {
			totalFlexGrow += flexGrow
			// Margins still count as fixed space
			totalFixed += child.Style.MarginTop + child.Style.MarginBottom
		} else {
			// Fixed height child
			childHeight := resolveDimension(child.Style.Height, height)
			if childHeight > 0 {
				totalFixed += childHeight + child.Style.MarginTop + child.Style.MarginBottom
			} else {
				// No explicit height and no flex-grow - use intrinsic height
				intrinsic := computeIntrinsicHeight(child)
				totalFixed += intrinsic + child.Style.MarginTop + child.Style.MarginBottom
			}
		}
	}

	// Remaining space for flex children
	remainingHeight := height - totalFixed
	if remainingHeight < 0 {
		remainingHeight = 0
	}

	// Phase 2: Lay out each child
	boxes := make([]*LayoutBox, len(children))
	currentY := y

	for i, child := range children {
		flexGrow := child.Style.FlexGrow

		// Calculate child height
		var childHeight int
		if flexGrow > 0 && totalFlexGrow > 0 {
			// Flexible child - distribute remaining space proportionally
			childHeight = (remainingHeight * flexGrow) / totalFlexGrow
		} else {
			// Fixed height
			childHeight = resolveDimension(child.Style.Height, height)
			if childHeight <= 0 {
				// Use intrinsic height if no explicit height set
				childHeight = computeIntrinsicHeight(child)
			}
		}

		// Calculate child width - respect align-self
		childWidth := resolveDimension(child.Style.Width, width)
		if childWidth <= 0 {
			if child.Style.AlignSelf == "stretch" {
				// Stretch to fill container width (cross-axis in column)
				childWidth = width - child.Style.MarginLeft - child.Style.MarginRight
			} else {
				// Use auto width (full width for now, can add flex-start/end later)
				childWidth = width - child.Style.MarginLeft - child.Style.MarginRight
			}
		}

		// Create layout box for this child
		childBox := &LayoutBox{
			X:      x + child.Style.MarginLeft,
			Y:      currentY + child.Style.MarginTop,
			Width:  childWidth,
			Height: childHeight,
			Node:   child,
		}

		// Recursively layout this child's children
		layoutChildren(childBox, child)

		boxes[i] = childBox

		// Move to next position
		currentY += childHeight + child.Style.MarginTop + child.Style.MarginBottom
	}

	return boxes
}

// layoutBlock handles block (non-flex) layout
func layoutBlock(children []*style.StyledNode, x, y, width, height int) []*LayoutBox {
	boxes := make([]*LayoutBox, len(children))
	currentY := y

	for i, child := range children {
		// Calculate child dimensions
		childWidth := resolveDimension(child.Style.Width, width)
		if childWidth <= 0 {
			childWidth = width
		}

		childHeight := resolveDimension(child.Style.Height, height)
		if childHeight <= 0 {
			// Use intrinsic height for block children
			childHeight = computeIntrinsicHeight(child)
		}

		// Create layout box
		childBox := &LayoutBox{
			X:      x + child.Style.MarginLeft,
			Y:      currentY + child.Style.MarginTop,
			Width:  childWidth - child.Style.MarginLeft - child.Style.MarginRight,
			Height: childHeight,
			Node:   child,
		}

		// Recursively layout children
		layoutChildren(childBox, child)

		boxes[i] = childBox

		// Move to next position
		currentY += childHeight + child.Style.MarginTop + child.Style.MarginBottom
	}

	return boxes
}

// resolveDimension converts CSS dimension to pixels
func resolveDimension(value string, containerSize int) int {
	value = strings.TrimSpace(value)

	if value == "" || value == "auto" {
		return 0 // auto means "use available space"
	}

	// Percentage
	if strings.HasSuffix(value, "%") {
		pct := strings.TrimSuffix(value, "%")
		p, _ := strconv.Atoi(pct)
		return (containerSize * p) / 100
	}

	// Pixels (or unitless number)
	num, _ := strconv.Atoi(value)
	return num
}

// computeIntrinsicHeight calculates the natural height of a node based on its content
func computeIntrinsicHeight(node *style.StyledNode) int {
	contentHeight := 1 // default minimum

	// For text nodes, height is based on number of lines
	// TextElement type is 1 (from vdom package)
	if node.Element != nil && node.Element.Type == 1 && node.Element.Text != "" {
		lines := strings.Split(node.Element.Text, "\n")
		contentHeight = len(lines)
	} else if len(node.Children) > 0 {
		// For containers, sum children's intrinsic heights (if vstack) or take max (if hstack)
		if node.Style.FlexDir == "column" {
			total := 0
			for _, child := range node.Children {
				total += computeIntrinsicHeight(child)
			}
			contentHeight = total
		} else {
			// For row, take maximum height
			maxHeight := 1
			for _, child := range node.Children {
				h := computeIntrinsicHeight(child)
				if h > maxHeight {
					maxHeight = h
				}
			}
			contentHeight = maxHeight
		}
	}

	// Add padding and border to content height
	totalHeight := contentHeight +
		node.Style.PaddingTop + node.Style.PaddingBottom +
		node.Style.MarginTop + node.Style.MarginBottom

	// Add border space (2 lines for top + bottom border)
	if node.Style.Border {
		totalHeight += 2
	}

	return totalHeight
}

// computeIntrinsicWidth calculates the natural width of a node based on its content
func computeIntrinsicWidth(node *style.StyledNode) int {
	contentWidth := 1 // default minimum

	// For text nodes, width is the length of the longest line
	// TextElement type is 1 (from vdom package)
	if node.Element != nil && node.Element.Type == 1 && node.Element.Text != "" {
		lines := strings.Split(node.Element.Text, "\n")
		maxLen := 0
		for _, line := range lines {
			if len(line) > maxLen {
				maxLen = len(line)
			}
		}
		if maxLen > 0 {
			contentWidth = maxLen
		}
	} else if len(node.Children) > 0 {
		// For containers, sum children's intrinsic widths (if hstack) or take max (if vstack)
		if node.Style.FlexDir == "row" {
			total := 0
			for _, child := range node.Children {
				total += computeIntrinsicWidth(child)
			}
			contentWidth = total
		} else {
			// For column, take maximum width
			maxWidth := 1
			for _, child := range node.Children {
				w := computeIntrinsicWidth(child)
				if w > maxWidth {
					maxWidth = w
				}
			}
			contentWidth = maxWidth
		}
	}

	// Add padding and border to content width
	totalWidth := contentWidth +
		node.Style.PaddingLeft + node.Style.PaddingRight +
		node.Style.MarginLeft + node.Style.MarginRight

	// Add border space (2 columns for left + right border)
	if node.Style.Border {
		totalWidth += 2
	}

	return totalWidth
}
