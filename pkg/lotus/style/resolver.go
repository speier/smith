// Package style handles CSS-like style resolution for terminal UIs.
//
// This package is independent of layout and rendering. It takes a vdom.Element
// tree and CSS rules, and produces a StyledNode tree with fully resolved styles.
//
// Similar to browser CSS cascade: defaults → stylesheet → inline styles
package style

import "github.com/speier/smith/pkg/lotus/vdom"

// StyledNode represents an element with fully resolved styles
// This is the output of style resolution, input to layout engine
type StyledNode struct {
	// Reference to original element
	Element *vdom.Element

	// Computed final styles (after cascade, inheritance, defaults)
	Style ComputedStyle

	// Styled children
	Children []*StyledNode
}

// ComputedStyle holds fully resolved CSS-like properties
// All values are normalized (percentages kept as strings, colors as hex, etc.)
type ComputedStyle struct {
	// Box model
	Display string // "block", "flex"
	FlexDir string // "row", "column"
	Width   string // "100%", "50", "auto"
	Height  string // "100%", "auto"
	Flex    string // "1", "0" (legacy shorthand, prefer FlexGrow)

	// Flexbox properties
	FlexGrow   int    // 0 = fixed size, 1+ = grows to fill space (default: 0)
	FlexShrink int    // 0 = can't shrink, 1+ = can shrink (default: 1)
	AlignSelf  string // "stretch" | "flex-start" | "flex-end" | "center" (default: "stretch")

	// Spacing (resolved to integers)
	PaddingTop    int
	PaddingRight  int
	PaddingBottom int
	PaddingLeft   int
	MarginTop     int
	MarginRight   int
	MarginBottom  int
	MarginLeft    int

	// Visual
	Color       string // "#ffffff"
	BgColor     string // "#000000"
	Border      bool
	BorderStyle string // "single", "rounded", "double"
	TextAlign   string // "left", "center", "right"
}

// Resolver handles CSS resolution
type Resolver struct {
	// CSS rules parsed from stylesheets
	rules []Rule
}

// Rule represents a CSS rule (selector + declarations)
type Rule struct {
	Selector     string
	Declarations map[string]string
}

// NewResolver creates a new style resolver
func NewResolver(css string) *Resolver {
	return &Resolver{
		rules: parseCSS(css),
	}
}

// Resolve takes a vdom tree and returns a styled tree
func (r *Resolver) Resolve(elem *vdom.Element) *StyledNode {
	// Compute styles for this element
	style := r.computeStyle(elem)

	// Resolve children recursively
	children := make([]*StyledNode, len(elem.Children))
	for i, child := range elem.Children {
		children[i] = r.Resolve(child)
	}

	return &StyledNode{
		Element:  elem,
		Style:    style,
		Children: children,
	}
}

// computeStyle resolves the final style for an element
// Order: defaults → CSS rules → inline styles
func (r *Resolver) computeStyle(elem *vdom.Element) ComputedStyle {
	// Start with defaults
	style := defaultStyle()

	// Apply matching CSS rules
	for _, rule := range r.rules {
		if r.matches(elem, rule.Selector) {
			applyDeclarations(&style, rule.Declarations)
		}
	}

	// Apply inline styles (highest priority)
	if elem.Props.Styles != nil {
		applyDeclarations(&style, elem.Props.Styles)
	}

	return style
}

// matches checks if element matches a CSS selector
func (r *Resolver) matches(elem *vdom.Element, selector string) bool {
	// Simple selector matching:
	// - "box" matches tag
	// - "#id" matches ID
	// - ".class" matches class
	// TODO: More complex selectors

	if selector == "" {
		return false
	}

	// ID selector
	if selector[0] == '#' {
		return elem.ID == selector[1:]
	}

	// Class selector
	if selector[0] == '.' {
		className := selector[1:]
		for _, c := range elem.Props.Classes {
			if c == className {
				return true
			}
		}
		return false
	}

	// Tag selector
	return elem.Tag == selector
}

// defaultStyle returns the default computed style
func defaultStyle() ComputedStyle {
	return ComputedStyle{
		Display:     "block",
		FlexDir:     "column",
		Width:       "auto",
		Height:      "auto",
		Flex:        "0",
		FlexGrow:    0,         // Don't grow by default
		FlexShrink:  1,         // Can shrink by default
		AlignSelf:   "stretch", // Stretch cross-axis by default
		BorderStyle: "single",
		TextAlign:   "left",
		Color:       "#ffffff",
		BgColor:     "",
	}
}

// applyDeclarations applies CSS declarations to a style
func applyDeclarations(style *ComputedStyle, decls map[string]string) {
	for key, value := range decls {
		switch key {
		case "display":
			style.Display = value
		case "flex-direction":
			style.FlexDir = value
		case "width":
			style.Width = value
		case "height":
			style.Height = value
		case "flex":
			style.Flex = value
			// Also set FlexGrow for convenience (flex: 1 means grow)
			if value != "0" && value != "" {
				style.FlexGrow = parseInt(value)
			}
		case "flex-grow":
			style.FlexGrow = parseInt(value)
		case "flex-shrink":
			style.FlexShrink = parseInt(value)
		case "align-self":
			style.AlignSelf = value
		case "color":
			style.Color = value
		case "background-color", "bg-color":
			style.BgColor = value
		case "border":
			style.Border = value != "" && value != "none"
			if value == "single" || value == "rounded" || value == "double" {
				style.BorderStyle = value
			}
		case "border-style":
			// Setting border-style automatically enables border (unless "none")
			style.BorderStyle = value
			style.Border = value != "" && value != "none"
		case "text-align":
			style.TextAlign = value
		case "padding":
			parsePadding(value, style)
		case "margin":
			parseMargin(value, style)
		}
	}
}

// parseCSS parses CSS string into rules
// Simple parser for now, can be enhanced later
func parseCSS(css string) []Rule {
	// TODO: Implement CSS parser
	// For now, return empty (inline styles will work)
	return nil
}

// parsePadding parses CSS padding shorthand
// Supports: "5", "5 10", "5 10 15", "5 10 15 20"
func parsePadding(value string, style *ComputedStyle) {
	parseSpacing(value, &style.PaddingTop, &style.PaddingRight, &style.PaddingBottom, &style.PaddingLeft)
}

// parseMargin parses CSS margin shorthand
// Supports: "5", "5 10", "5 10 15", "5 10 15 20"
func parseMargin(value string, style *ComputedStyle) {
	parseSpacing(value, &style.MarginTop, &style.MarginRight, &style.MarginBottom, &style.MarginLeft)
}

// parseSpacing handles CSS shorthand for padding/margin
// CSS rules: 1 value = all, 2 values = vertical horizontal, 3 values = top horizontal bottom, 4 values = top right bottom left
func parseSpacing(value string, top, right, bottom, left *int) {
	parts := splitWhitespace(value)

	switch len(parts) {
	case 1:
		// All sides
		v := parseInt(parts[0])
		*top, *right, *bottom, *left = v, v, v, v
	case 2:
		// Vertical horizontal
		vertical := parseInt(parts[0])
		horizontal := parseInt(parts[1])
		*top, *bottom = vertical, vertical
		*left, *right = horizontal, horizontal
	case 3:
		// Top horizontal bottom
		*top = parseInt(parts[0])
		horizontal := parseInt(parts[1])
		*left, *right = horizontal, horizontal
		*bottom = parseInt(parts[2])
	case 4:
		// Top right bottom left
		*top = parseInt(parts[0])
		*right = parseInt(parts[1])
		*bottom = parseInt(parts[2])
		*left = parseInt(parts[3])
	}
}

// parseInt parses an integer from string, returns 0 on error
func parseInt(s string) int {
	var result int
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			result = result*10 + int(ch-'0')
		} else {
			return 0 // Invalid, return 0
		}
	}
	return result
}

// splitWhitespace splits a string on whitespace
func splitWhitespace(s string) []string {
	var parts []string
	var current string

	for _, ch := range s {
		if ch == ' ' || ch == '\t' || ch == '\n' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}
