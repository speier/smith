package vdom

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// Markup converts JSX-like markup to vdom Elements with optional data substitution
// Based on the original lotus parser but adapted for clean vdom architecture
//
// Supported syntax:
//
//	<box id="name">content</box>
//	<box class="message user">text</box>
//	<vstack>...</vstack>   (vertical flex container)
//	<hstack>...</hstack>   (horizontal flex container)
//	<text>Hello</text>
//
// Data substitution (auto-detects syntax):
//
// Simple indexed placeholders (fast, no template overhead):
//
//	Markup(`<box>{0}</box>`, "Hello")                    // → <box>Hello</box>
//	Markup(`<box>{0} {1}</box>`, "Hello", "World")       // → <box>Hello World</box>
//
// Go template syntax (powerful, for complex data):
//
//	user := struct{ Name string }{"Alice"}
//	Markup(`<box>{{.Name}}</box>`, user)                 // → <box>Alice</box>
//	Markup(`{{if .Show}}<box>{{.Text}}</box>{{end}}`, data)
//
// Auto-detection: Contains "{{" → uses text/template, otherwise uses {0}, {1}
//
// Supported attributes:
//
//	id="name"                        - Element ID
//	class="cls1 cls2"                - CSS classes (space-separated)
//	style="width: 80; color: red"    - Inline CSS (sugar syntax)
//
// Attribute shortcuts (map directly to styles):
//
//	width="80"             - Element width
//	height="24"            - Element height
//	flex="1"               - Flex grow value
//	color="#fff"           - Text color
//	background="#000"      - Background color
//	border="1"             - Border (any value enables border)
//	border-style="rounded" - Border style
//	padding="2"            - Padding (supports 1-4 values like CSS)
//	margin="1"             - Margin (supports 1-4 values like CSS)
//	direction="column"     - Flex direction (sets display:flex automatically)
func Markup(markup string, data ...any) *Element {
	// Auto-detect template syntax: if contains {{...}}, use text/template
	if strings.Contains(markup, "{{") {
		return executeTemplate(markup, data)
	}

	// Otherwise use simple {0}, {1} indexed replacement
	if len(data) > 0 {
		for i, val := range data {
			placeholder := fmt.Sprintf("{%d}", i)
			markup = strings.ReplaceAll(markup, placeholder, fmt.Sprint(val))
		}
	}

	p := &parser{
		input: strings.TrimSpace(markup),
		pos:   0,
	}
	return p.parse()
}

// executeTemplate processes markup using Go's text/template
func executeTemplate(markup string, data []any) *Element {
	// Determine template data
	var templateData any
	if len(data) == 0 {
		templateData = nil
	} else if len(data) == 1 {
		// Single data parameter - use as-is
		templateData = data[0]
	} else {
		// Multiple parameters - create map with indexes
		m := make(map[string]any)
		for i, val := range data {
			m[fmt.Sprintf("%d", i)] = val
		}
		templateData = m
	}

	// Execute template
	tmpl, err := template.New("markup").Parse(markup)
	if err != nil {
		// If template fails, return error as text
		return Text(fmt.Sprintf("Template error: %v", err))
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return Text(fmt.Sprintf("Template execution error: %v", err))
	}

	// Parse the result
	p := &parser{
		input: strings.TrimSpace(buf.String()),
		pos:   0,
	}
	return p.parse()
}

type parser struct {
	input string
	pos   int
}

func (p *parser) parse() *Element {
	p.skipWhitespace()
	if p.pos >= len(p.input) {
		return Text("")
	}

	if p.input[p.pos] != '<' {
		return p.parseText()
	}

	return p.parseElement()
}

func (p *parser) parseElement() *Element {
	// Skip '<'
	p.pos++

	// Check for closing tag
	if p.pos < len(p.input) && p.input[p.pos] == '/' {
		return nil
	}

	// Parse tag name
	tagName := p.parseIdentifier()

	// Create props
	props := Props{
		Attributes: make(map[string]string),
		Styles:     make(map[string]string),
		Classes:    []string{},
	}

	// Auto-apply flexbox for stack elements (like old parser)
	switch tagName {
	case "vstack":
		props.Styles["display"] = "flex"
		props.Styles["flex-direction"] = "column"
		tagName = "box" // vstack is just a box with column flex
	case "hstack", "hbox":
		props.Styles["display"] = "flex"
		props.Styles["flex-direction"] = "row"
		tagName = "box" // hstack is just a box with row flex
	}

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

			// Handle special attributes (like old parser)
			switch attrName {
			case "id":
				props.Attributes["id"] = attrValue
			case "class":
				props.Classes = append(props.Classes, strings.Fields(attrValue)...)
			case "style":
				// Parse CSS syntax: "width: 80; color: #fff; padding: 2"
				p.parseInlineCSS(attrValue, props.Styles)
			case "direction":
				// Map direction to flex styles
				props.Styles["display"] = "flex"
				if attrValue == "column" {
					props.Styles["flex-direction"] = "column"
				} else {
					props.Styles["flex-direction"] = "row"
				}
			case "width", "height", "flex", "color", "background", "border", "border-style", "padding", "margin":
				// Map attributes to inline styles
				props.Styles[attrName] = attrValue
			default:
				props.Attributes[attrName] = attrValue
			}
		}
	}

	// Check for self-closing tag
	if p.pos < len(p.input) && p.input[p.pos] == '/' {
		p.pos++ // skip '/'
		if p.pos < len(p.input) && p.input[p.pos] == '>' {
			p.pos++
		}
		return NewElement(tagName, props)
	}

	// Skip '>'
	if p.pos < len(p.input) && p.input[p.pos] == '>' {
		p.pos++
	}

	// Parse children until closing tag
	var children []*Element
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
			children = append(children, child)
		}
	}

	return NewElement(tagName, props, children...)
}

func (p *parser) parseText() *Element {
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] != '<' {
		p.pos++
	}

	text := strings.TrimSpace(p.input[start:p.pos])
	if text == "" {
		return nil
	}

	return Text(text)
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

// parseInlineCSS parses CSS syntax: "width: 80; color: #fff; padding: 2"
func (p *parser) parseInlineCSS(css string, styles map[string]string) {
	css = strings.TrimSpace(css)
	if css == "" {
		return
	}

	// Split by semicolon
	declarations := strings.Split(css, ";")
	for _, decl := range declarations {
		decl = strings.TrimSpace(decl)
		if decl == "" {
			continue
		}

		// Split property: value
		parts := strings.SplitN(decl, ":", 2)
		if len(parts) != 2 {
			continue
		}

		property := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if property != "" && value != "" {
			styles[property] = value
		}
	}
}

func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

func isAlphaNum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}
