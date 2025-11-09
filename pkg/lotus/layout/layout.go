package layout

import (
	"strconv"
	"strings"
)

// Layout computes positions and sizes for all nodes
func Layout(root *Node, width, height int) {
	// Set root dimensions
	root.X = 0
	root.Y = 0
	root.Width = width
	root.Height = height

	// Layout children
	layoutChildren(root)
}

func layoutChildren(node *Node) {
	// Calculate available space for children (account for padding and border)
	innerX := node.X + node.Styles.PaddingLeft
	innerY := node.Y + node.Styles.PaddingTop
	innerWidth := node.Width - node.Styles.PaddingLeft - node.Styles.PaddingRight
	innerHeight := node.Height - node.Styles.PaddingTop - node.Styles.PaddingBottom

	// Account for border
	if node.Styles.Border {
		innerX++
		innerY++
		innerWidth -= 2
		innerHeight -= 2
	}

	// Layout children based on display mode
	if node.Styles.Display == "flex" {
		if node.Styles.FlexDir == "column" || node.Styles.FlexDir == "" {
			layoutFlexColumn(node, innerX, innerY, innerWidth, innerHeight)
		} else {
			layoutFlexRow(node, innerX, innerY, innerWidth, innerHeight)
		}
	} else {
		layoutBlock(node, innerX, innerY, innerWidth, innerHeight)
	}
}

func layoutBlock(parent *Node, x, y, width, height int) {
	currentY := y
	for _, child := range parent.Children {
		// Compute child dimensions
		child.X = x + child.Styles.MarginLeft
		child.Y = currentY + child.Styles.MarginTop
		child.Width = computeDimension(child.Styles.Width, width) - child.Styles.MarginLeft - child.Styles.MarginRight

		childHeight := computeDimension(child.Styles.Height, height)
		if childHeight == 0 {
			childHeight = 1 // min height for text
		}
		child.Height = childHeight

		// Layout this child's children
		layoutChildren(child)

		// Move to next position
		currentY += childHeight + child.Styles.MarginTop + child.Styles.MarginBottom
	}
}

func layoutFlexColumn(parent *Node, x, y, width, height int) {
	// Calculate total fixed height and count flexible children
	var totalFixed int
	var totalFlexGrow float64

	for _, child := range parent.Children {
		flexVal := parseFloat(child.Styles.Flex)

		if flexVal > 0 {
			// Flexible child - will get remaining space
			totalFlexGrow += flexVal
		} else {
			// Fixed child - use explicit height (not "auto")
			if child.Styles.Height != "auto" && child.Styles.Height != "" {
				heightVal := computeDimension(child.Styles.Height, height)
				if heightVal > 0 {
					totalFixed += heightVal
					totalFixed += child.Styles.MarginTop + child.Styles.MarginBottom
				}
			}
		}
	}

	// Remaining space for flexible children
	remainingHeight := height - totalFixed

	// Debug: Uncomment to trace layout
	// fmt.Printf("layoutFlexColumn: height=%d, totalFixed=%d, remaining=%d, totalFlexGrow=%.1f\n",
	//     height, totalFixed, remainingHeight, totalFlexGrow)

	// Layout each child
	currentY := y
	for _, child := range parent.Children {
		// Compute child dimensions
		child.X = x + child.Styles.MarginLeft
		child.Y = currentY + child.Styles.MarginTop
		child.Width = computeDimension(child.Styles.Width, width) - child.Styles.MarginLeft - child.Styles.MarginRight

		flexVal := parseFloat(child.Styles.Flex)

		var childHeight int
		if flexVal > 0 && totalFlexGrow > 0 {
			// Flexible child - distribute remaining space
			childHeight = int(float64(remainingHeight) * (flexVal / totalFlexGrow))
		} else if child.Styles.Height != "auto" && child.Styles.Height != "" {
			// Fixed child - use explicit height
			childHeight = computeDimension(child.Styles.Height, height)
			if childHeight <= 0 {
				childHeight = 1
			}
		} else {
			// Auto height without flex - use minimal size
			childHeight = 1
		}
		child.Height = childHeight

		// Layout this child's children
		layoutChildren(child)

		// Move to next position
		currentY += childHeight + child.Styles.MarginTop + child.Styles.MarginBottom
	}
}

func layoutFlexRow(parent *Node, x, y, width, height int) {
	// Calculate total fixed width and flexible children
	var totalFixed int
	var totalFlexGrow float64

	for _, child := range parent.Children {
		flexVal := parseFloat(child.Styles.Flex)

		if flexVal > 0 {
			// Flexible child - will get remaining space
			totalFlexGrow += flexVal
		} else {
			// Fixed child - use explicit width (not "auto")
			if child.Styles.Width != "auto" && child.Styles.Width != "" {
				widthVal := computeDimension(child.Styles.Width, width)
				if widthVal > 0 {
					totalFixed += widthVal
					totalFixed += child.Styles.MarginLeft + child.Styles.MarginRight
				}
			}
		}
	}

	// Remaining space for flexible children
	remainingWidth := width - totalFixed

	// Layout each child
	currentX := x
	for _, child := range parent.Children {
		// Compute child dimensions
		child.X = currentX + child.Styles.MarginLeft
		child.Y = y + child.Styles.MarginTop
		child.Height = computeDimension(child.Styles.Height, height) - child.Styles.MarginTop - child.Styles.MarginBottom

		flexVal := parseFloat(child.Styles.Flex)

		var childWidth int
		if flexVal > 0 && totalFlexGrow > 0 {
			// Flexible child - distribute remaining space
			childWidth = int(float64(remainingWidth) * (flexVal / totalFlexGrow))
		} else if child.Styles.Width != "auto" && child.Styles.Width != "" {
			// Fixed child - use explicit width
			childWidth = computeDimension(child.Styles.Width, width)
			if childWidth <= 0 {
				childWidth = 1
			}
		} else {
			// Auto width without flex - use minimal size
			childWidth = 1
		}
		child.Width = childWidth

		// Layout this child's children
		layoutChildren(child)

		// Move to next position
		currentX += childWidth + child.Styles.MarginLeft + child.Styles.MarginRight
	}
}

func computeDimension(value string, parent int) int {
	value = strings.TrimSpace(value)

	if value == "auto" || value == "" {
		return parent // Full parent size for auto
	}

	if strings.HasSuffix(value, "%") {
		pct := strings.TrimSuffix(value, "%")
		p, _ := strconv.Atoi(pct)
		return (parent * p) / 100
	}

	// Assume pixels
	i, _ := strconv.Atoi(value)
	return i
}

// parseFloat converts string to float64
func parseFloat(value string) float64 {
	value = strings.TrimSpace(value)
	if value == "" || value == "0" {
		return 0
	}
	f, _ := strconv.ParseFloat(value, 64)
	return f
}
