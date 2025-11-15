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

	"github.com/mattn/go-runewidth"
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
	// Resolve node's own width/height (may be constrained)
	nodeWidth := resolveDimension(node.Style.Width, containerWidth)
	if nodeWidth <= 0 {
		nodeWidth = containerWidth
	}

	nodeHeight := resolveDimension(node.Style.Height, containerHeight)
	if nodeHeight <= 0 {
		nodeHeight = containerHeight
	}

	box := &LayoutBox{
		X:      0,
		Y:      0,
		Width:  nodeWidth,
		Height: nodeHeight,
		Node:   node,
	}

	// Layout children
	layoutChildren(box, node)

	// Shrink-wrap flex columns when content exceeds container height
	// This fixes overflow:auto scroll calculation for dynamically growing content
	// Only when: no explicit height, is flex column, has children, container has size
	if resolveDimension(node.Style.Height, containerHeight) <= 0 &&
		node.Style.Display == "flex" && node.Style.FlexDir == "column" &&
		containerHeight > 0 && len(box.Children) > 0 {
		// Calculate actual content height from children
		maxY := 0
		for _, child := range box.Children {
			childBottom := child.Y + child.Height
			if childBottom > maxY {
				maxY = childBottom
			}
		}
		// Add bottom padding/border
		actualHeight := maxY - box.Y + node.Style.PaddingBottom
		if node.Style.Border {
			actualHeight++ // Bottom border
		}
		// Only grow beyond container height (for scroll calculation)
		// Don't shrink below container height (preserves fill behavior)
		if actualHeight > box.Height {
			box.Height = actualHeight
		}
	}

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
			box.Children = layoutFlexRow(node.Children, contentX, contentY, contentWidth, contentHeight, node.Style)
		} else {
			box.Children = layoutFlexColumn(node.Children, contentX, contentY, contentWidth, contentHeight, node.Style)
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
func layoutFlexRow(children []*style.StyledNode, x, y, width, height int, parentStyle style.ComputedStyle) []*LayoutBox {
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
				intrinsicWidth := ComputeIntrinsicWidth(child)
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
				childWidth = ComputeIntrinsicWidth(child)
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

		// Check for auto margins - centers the child horizontally
		marginLeft := child.Style.MarginLeft
		if child.Style.MarginLeftAuto && child.Style.MarginRightAuto {
			// Both margins auto - center the child
			availableSpace := width - childWidth
			if availableSpace > 0 {
				marginLeft = availableSpace / 2
			}
		}

		// Create layout box for this child
		childBox := &LayoutBox{
			X:      currentX + marginLeft,
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

	// Phase 3: Apply justify-content
	switch parentStyle.JustifyContent {
	case "center":
		// Calculate total used width
		totalUsed := currentX - x
		if totalUsed < width {
			offset := (width - totalUsed) / 2
			// Shift all boxes to the right
			for _, box := range boxes {
				box.X += offset
			}
		}
	case "flex-end":
		// Calculate total used width
		totalUsed := currentX - x
		if totalUsed < width {
			offset := width - totalUsed
			// Shift all boxes to the right
			for _, box := range boxes {
				box.X += offset
			}
		}
	}

	return boxes
}

// layoutFlexColumn handles vertical flexbox layout
func layoutFlexColumn(children []*style.StyledNode, x, y, width, height int, parentStyle style.ComputedStyle) []*LayoutBox {
	// Phase 1: Calculate fixed and flexible space
	var totalFixed int
	var totalFlexGrow int

	// Add gap spacing: (n-1) gaps between n children
	if len(children) > 1 && parentStyle.Gap > 0 {
		totalFixed += (len(children) - 1) * parentStyle.Gap
	}

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
				// No explicit height and no flex-grow - use intrinsic height WITH width for wrapping
				// We know the width at this point, so calculate wrapped height accurately
				intrinsic := computeIntrinsicHeightWithWidth(child, width)
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

		// Calculate child width FIRST - respect align-items (cross-axis in column)
		childWidth := resolveDimension(child.Style.Width, width)

		if childWidth <= 0 {
			// No explicit width - check align-items behavior
			// Use parent's align-items to determine cross-axis sizing
			alignItems := parentStyle.AlignItems
			if child.Style.AlignSelf != "" && child.Style.AlignSelf != "stretch" {
				alignItems = child.Style.AlignSelf
			}

			if alignItems == "stretch" {
				// Stretch to fill parent width (default behavior)
				childWidth = width - child.Style.MarginLeft - child.Style.MarginRight
			} else {
				// Use intrinsic width for non-stretch alignment
				childWidth = ComputeIntrinsicWidth(child)
			}
		}

		// NOW calculate child height (using width for wrapping calculation)
		var childHeight int
		if flexGrow > 0 && totalFlexGrow > 0 {
			// Flexible child - distribute remaining space proportionally
			childHeight = (remainingHeight * flexGrow) / totalFlexGrow
		} else {
			// Fixed height
			childHeight = resolveDimension(child.Style.Height, height)
			if childHeight <= 0 {
				// Use intrinsic height WITH width for accurate wrapping calculation
				childHeight = computeIntrinsicHeightWithWidth(child, childWidth)
			}
		}

		// Calculate child X position
		childX := x + child.Style.MarginLeft

		// Apply align-items to position child horizontally (cross-axis)
		alignItems := child.Style.AlignSelf
		if alignItems == "" || alignItems == "auto" {
			alignItems = parentStyle.AlignItems
		}

		// Check if child has explicit width
		hasExplicitWidth := resolveDimension(child.Style.Width, width) > 0

		switch alignItems {
		case "stretch":
			// Only stretch if no explicit width
			if !hasExplicitWidth {
				childWidth = width - child.Style.MarginLeft - child.Style.MarginRight
			}
			childX = x + child.Style.MarginLeft
		case "center":
			// Center horizontally
			availableSpace := width - childWidth - child.Style.MarginLeft - child.Style.MarginRight
			if availableSpace > 0 {
				childX = x + child.Style.MarginLeft + (availableSpace / 2)
			}
		case "flex-end":
			// Align to right
			childX = x + width - childWidth - child.Style.MarginRight
		case "flex-start":
			// Align to left (default)
			childX = x + child.Style.MarginLeft
		default:
			// CSS default is stretch, so use that if alignItems is somehow empty
			if !hasExplicitWidth {
				childWidth = width - child.Style.MarginLeft - child.Style.MarginRight
			}
			childX = x + child.Style.MarginLeft
		}

		// Create layout box for this child
		childBox := &LayoutBox{
			X:      childX,
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

		// Add gap spacing after this child (except for last child)
		if i < len(children)-1 && parentStyle.Gap > 0 {
			currentY += parentStyle.Gap
		}
	}

	// Phase 3: Apply justify-content for column direction
	switch parentStyle.JustifyContent {
	case "center":
		// Calculate total used height
		totalUsed := currentY - y
		if totalUsed < height {
			offset := (height - totalUsed) / 2
			// Shift all boxes down
			for _, box := range boxes {
				box.Y += offset
			}
		}
	case "flex-end":
		// Calculate total used height
		totalUsed := currentY - y
		if totalUsed < height {
			offset := height - totalUsed
			// Shift all boxes down
			for _, box := range boxes {
				box.Y += offset
			}
		}
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
// For text nodes with WordWrap enabled, this calculates the height AFTER wrapping
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

// computeIntrinsicHeightWithWidth calculates height for text that will wrap to a given width
func computeIntrinsicHeightWithWidth(node *style.StyledNode, availableWidth int) int {
	if node.Element == nil || node.Element.Type != 1 || node.Element.Text == "" {
		return computeIntrinsicHeight(node)
	}

	// Calculate content width after padding/border
	contentWidth := availableWidth - node.Style.PaddingLeft - node.Style.PaddingRight
	if node.Style.Border {
		contentWidth -= 2
	}

	if contentWidth <= 0 || !node.Style.WordWrap {
		// Fall back to simple line counting
		lines := strings.Split(node.Element.Text, "\n")
		return len(lines) + node.Style.PaddingTop + node.Style.PaddingBottom +
			node.Style.MarginTop + node.Style.MarginBottom
	}

	// Calculate wrapped line count
	wrappedLines := wrapTextForHeight(node.Element.Text, contentWidth)

	contentHeight := len(wrappedLines)
	totalHeight := contentHeight +
		node.Style.PaddingTop + node.Style.PaddingBottom +
		node.Style.MarginTop + node.Style.MarginBottom

	if node.Style.Border {
		totalHeight += 2
	}

	return totalHeight
}

// wrapTextForHeight calculates how many lines text will wrap to at given width
func wrapTextForHeight(text string, width int) []string {
	if width <= 0 {
		return []string{}
	}

	var lines []string
	inputLines := strings.Split(text, "\n")

	for _, inputLine := range inputLines {
		if inputLine == "" {
			lines = append(lines, "")
			continue
		}

		currentLine := ""
		currentWidth := 0

		words := strings.Fields(inputLine)
		for i, word := range words {
			wordWidth := visibleLen(word)

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

			if i == len(words)-1 && currentLine != "" {
				lines = append(lines, currentLine)
			}
		}
	}

	return lines
}

// ComputeIntrinsicWidth calculates the natural width of a node based on its content
func ComputeIntrinsicWidth(node *style.StyledNode) int {
	contentWidth := 1 // default minimum

	// For text nodes, width is the length of the longest line
	// TextElement type is 1 (from vdom package)
	if node.Element != nil && node.Element.Type == 1 && node.Element.Text != "" {
		lines := strings.Split(node.Element.Text, "\n")
		maxLen := 0
		for _, line := range lines {
			// Use visible length to exclude ANSI escape codes
			visLen := visibleLen(line)
			if visLen > maxLen {
				maxLen = visLen
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
				total += ComputeIntrinsicWidth(child)
			}
			contentWidth = total
		} else {
			// For column, take maximum width
			maxWidth := 1
			for _, child := range node.Children {
				w := ComputeIntrinsicWidth(child)
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

// visibleLen returns the visible character count (excluding ANSI escape codes)
func visibleLen(s string) int {
	count := 0
	inEscape := false

	for _, r := range s { // Iterate over runes, not bytes!
		if r == '\033' { // ESC character
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' { // End of ANSI sequence
				inEscape = false
			}
			continue
		}
		// Use RuneWidth to account for wide characters (emojis = 2, normal = 1)
		count += runewidth.RuneWidth(r)
	}

	return count
}
