package parser

import (
	"strconv"
	"strings"

	"github.com/speier/smith/pkg/lotus/layout"
)

// Parse parses simple HTML-like markup into a node tree
func Parse(markup string) *layout.Node {
	p := &parser{
		input: strings.TrimSpace(markup),
		pos:   0,
	}
	return p.parse()
}

type parser struct {
	input string
	pos   int
}

func (p *parser) parse() *layout.Node {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return nil
	}

	if p.input[p.pos] != '<' {
		return p.parseText()
	}

	return p.parseElement()
}

func (p *parser) parseElement() *layout.Node {
	// Skip '<'
	p.pos++

	// Check for closing tag
	if p.pos < len(p.input) && p.input[p.pos] == '/' {
		return nil
	}

	// Parse tag name
	tagName := p.parseIdentifier()
	node := layout.NewNode(tagName)

	// Parse attributes
	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) || p.input[p.pos] == '>' || p.input[p.pos] == '/' {
			break
		}

		attrName := p.parseIdentifier()
		p.skipWhitespace()

		if p.pos < len(p.input) && p.input[p.pos] == '=' {
			p.pos++ // skip '='
			p.skipWhitespace()
			attrValue := p.parseAttributeValue()

			// Handle special attributes
			switch attrName {
			case "id":
				node.ID = attrValue
			case "class":
				node.Classes = strings.Split(attrValue, " ")
			default:
				node.Attributes[attrName] = attrValue
			}
		}
	}

	// Check for self-closing tag
	if p.pos < len(p.input) && p.input[p.pos] == '/' {
		p.pos++ // skip '/'
		p.pos++ // skip '>'
		return node
	}

	// Skip '>'
	if p.pos < len(p.input) && p.input[p.pos] == '>' {
		p.pos++
	}

	// Parse children until closing tag
	for p.pos < len(p.input) {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}

		// Check for closing tag
		if p.pos+1 < len(p.input) && p.input[p.pos] == '<' && p.input[p.pos+1] == '/' {
			// Skip closing tag
			p.pos += 2 // skip '</'
			p.parseIdentifier()
			if p.pos < len(p.input) && p.input[p.pos] == '>' {
				p.pos++
			}
			break
		}

		child := p.parse()
		if child != nil {
			node.AddChild(child)
		}
	}

	return node
}

func (p *parser) parseText() *layout.Node {
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] != '<' {
		p.pos++
	}

	text := strings.TrimSpace(p.input[start:p.pos])
	if text == "" {
		return nil
	}

	node := layout.NewNode("text")
	node.Content = text
	return node
}

func (p *parser) parseIdentifier() string {
	start := p.pos
	for p.pos < len(p.input) && (isAlphaNum(p.input[p.pos]) || p.input[p.pos] == '-') {
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *parser) parseAttributeValue() string {
	if p.pos < len(p.input) && p.input[p.pos] == '"' {
		p.pos++ // skip opening quote
		start := p.pos
		for p.pos < len(p.input) && p.input[p.pos] != '"' {
			p.pos++
		}
		value := p.input[start:p.pos]
		if p.pos < len(p.input) {
			p.pos++ // skip closing quote
		}
		return value
	}

	// Unquoted value
	start := p.pos
	for p.pos < len(p.input) && !isWhitespace(p.input[p.pos]) && p.input[p.pos] != '>' {
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *parser) skipWhitespace() {
	for p.pos < len(p.input) && isWhitespace(p.input[p.pos]) {
		p.pos++
	}
}

func isAlphaNum(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

// stripCSSComments removes /* ... */ style comments from CSS
func stripCSSComments(css string) string {
	result := strings.Builder{}
	inComment := false

	for i := 0; i < len(css); i++ {
		if !inComment && i+1 < len(css) && css[i] == '/' && css[i+1] == '*' {
			inComment = true
			i++ // skip the *
			continue
		}
		if inComment && i+1 < len(css) && css[i] == '*' && css[i+1] == '/' {
			inComment = false
			i++ // skip the /
			continue
		}
		if !inComment {
			result.WriteByte(css[i])
		}
	}

	return result.String()
}

// ParseCSS parses simple CSS-like styles
func ParseCSS(css string) map[string]map[string]string {
	styles := make(map[string]map[string]string)

	css = strings.TrimSpace(css)
	if css == "" {
		return styles
	}

	// Strip CSS comments /* ... */
	css = stripCSSComments(css)

	// Split by closing brace to get rules
	rules := strings.Split(css, "}")
	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		// Split selector and properties
		parts := strings.SplitN(rule, "{", 2)
		if len(parts) != 2 {
			continue
		}

		selector := strings.TrimSpace(parts[0])
		propertiesStr := strings.TrimSpace(parts[1])

		properties := make(map[string]string)

		// Parse properties
		for _, prop := range strings.Split(propertiesStr, ";") {
			prop = strings.TrimSpace(prop)
			if prop == "" {
				continue
			}

			kv := strings.SplitN(prop, ":", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.TrimSpace(kv[1])
				properties[key] = value
			}
		}

		styles[selector] = properties
	}

	return styles
}

// ApplyStyles applies CSS styles to the node tree
func ApplyStyles(root *layout.Node, styles map[string]map[string]string) {
	applyStylesToNode(root, styles)
}

func applyStylesToNode(node *layout.Node, styles map[string]map[string]string) {
	// Apply type selector
	if props, ok := styles[node.Type]; ok {
		applyProperties(node, props)
	}

	// Apply class selectors
	for _, class := range node.Classes {
		if props, ok := styles["."+class]; ok {
			applyProperties(node, props)
		}
	}

	// Apply multi-class selectors (e.g., ".message.system")
	// Check all selectors in styles to see if they match this node's classes
	for selector, props := range styles {
		if strings.HasPrefix(selector, ".") && strings.Contains(selector, ".") {
			// This is a multi-class selector like ".message.system"
			// Extract all class names from the selector
			selectorClasses := strings.Split(selector[1:], ".") // Remove leading "." and split

			// Check if node has ALL the classes in the selector
			allMatch := true
			for _, selectorClass := range selectorClasses {
				found := false
				for _, nodeClass := range node.Classes {
					if nodeClass == selectorClass {
						found = true
						break
					}
				}
				if !found {
					allMatch = false
					break
				}
			}

			if allMatch && len(selectorClasses) > 1 {
				applyProperties(node, props)
			}
		}
	}

	// Apply ID selector
	if node.ID != "" {
		if props, ok := styles["#"+node.ID]; ok {
			applyProperties(node, props)
		}
	}

	// Apply to children
	for _, child := range node.Children {
		applyStylesToNode(child, styles)
	}
}

func applyProperties(node *layout.Node, props map[string]string) {
	for key, value := range props {
		switch key {
		case "width":
			node.Styles.Width = value
		case "height":
			node.Styles.Height = value
		case "display":
			node.Styles.Display = value
		case "flex-direction":
			node.Styles.FlexDir = value
		case "flex":
			node.Styles.Flex = value
		case "position":
			node.Styles.Position = value
		case "color":
			node.Styles.Color = value
		case "background-color":
			node.Styles.BgColor = value
		case "text-align":
			node.Styles.TextAlign = value
		case "border":
			node.Styles.Border = value != "none"
		case "border-style":
			node.Styles.BorderChar = value
		case "padding":
			parsePadding(node, value)
		case "margin":
			parseMargin(node, value)
		case "top":
			node.Styles.Top = parseInt(value)
		case "bottom":
			node.Styles.Bottom = parseInt(value)
		case "left":
			node.Styles.Left = parseInt(value)
		case "right":
			node.Styles.Right = parseInt(value)
		}
	}
}

func parsePadding(node *layout.Node, value string) {
	parts := strings.Fields(value)
	switch len(parts) {
	case 1:
		p := parseInt(parts[0])
		node.Styles.PaddingTop = p
		node.Styles.PaddingRight = p
		node.Styles.PaddingBottom = p
		node.Styles.PaddingLeft = p
	case 2:
		v := parseInt(parts[0])
		h := parseInt(parts[1])
		node.Styles.PaddingTop = v
		node.Styles.PaddingBottom = v
		node.Styles.PaddingLeft = h
		node.Styles.PaddingRight = h
	case 4:
		node.Styles.PaddingTop = parseInt(parts[0])
		node.Styles.PaddingRight = parseInt(parts[1])
		node.Styles.PaddingBottom = parseInt(parts[2])
		node.Styles.PaddingLeft = parseInt(parts[3])
	}
}

func parseMargin(node *layout.Node, value string) {
	parts := strings.Fields(value)
	switch len(parts) {
	case 1:
		m := parseInt(parts[0])
		node.Styles.MarginTop = m
		node.Styles.MarginRight = m
		node.Styles.MarginBottom = m
		node.Styles.MarginLeft = m
	case 2:
		v := parseInt(parts[0])
		h := parseInt(parts[1])
		node.Styles.MarginTop = v
		node.Styles.MarginBottom = v
		node.Styles.MarginLeft = h
		node.Styles.MarginRight = h
	case 4:
		node.Styles.MarginTop = parseInt(parts[0])
		node.Styles.MarginRight = parseInt(parts[1])
		node.Styles.MarginBottom = parseInt(parts[2])
		node.Styles.MarginLeft = parseInt(parts[3])
	}
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
