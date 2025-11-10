package core

import (
	"fmt"
	"strings"
)

// FlexDirection represents the flex direction
type FlexDirection string

const (
	Column FlexDirection = "column"
	Row    FlexDirection = "row"
)

// BorderStyle represents border style
type BorderStyle string

const (
	BorderSolid   BorderStyle = "solid"
	BorderRounded BorderStyle = "rounded"
	BorderDouble  BorderStyle = "double"
	BorderSingle  BorderStyle = "single"
)

// BoxBuilder provides a fluent API for building boxes
type BoxBuilder struct {
	id          string
	class       string
	direction   FlexDirection
	flex        int
	width       string
	height      string
	padding     string
	margin      string
	border      string
	borderStyle BorderStyle
	color       string
	background  string
	children    []string
}

// NewBox creates a new BoxBuilder
func NewBox() *BoxBuilder {
	return &BoxBuilder{
		children: make([]string, 0),
	}
}

// ID sets the box ID
func (b *BoxBuilder) ID(id string) *BoxBuilder {
	b.id = id
	return b
}

// Class sets the CSS class
func (b *BoxBuilder) Class(class string) *BoxBuilder {
	b.class = class
	return b
}

// Direction sets the flex direction
func (b *BoxBuilder) Direction(dir FlexDirection) *BoxBuilder {
	b.direction = dir
	return b
}

// Flex sets the flex grow factor
func (b *BoxBuilder) Flex(flex int) *BoxBuilder {
	b.flex = flex
	return b
}

// Width sets the width
func (b *BoxBuilder) Width(width string) *BoxBuilder {
	b.width = width
	return b
}

// Height sets the height
func (b *BoxBuilder) Height(height string) *BoxBuilder {
	b.height = height
	return b
}

// Padding sets the padding
func (b *BoxBuilder) Padding(padding string) *BoxBuilder {
	b.padding = padding
	return b
}

// Margin sets the margin
func (b *BoxBuilder) Margin(margin string) *BoxBuilder {
	b.margin = margin
	return b
}

// Border sets the border
func (b *BoxBuilder) Border(border string) *BoxBuilder {
	b.border = border
	return b
}

// BorderStyle sets the border style
func (b *BoxBuilder) BorderStyled(style BorderStyle) *BoxBuilder {
	b.borderStyle = style
	return b
}

// Color sets the text color
func (b *BoxBuilder) Color(color string) *BoxBuilder {
	b.color = color
	return b
}

// Background sets the background color
func (b *BoxBuilder) Background(bg string) *BoxBuilder {
	b.background = bg
	return b
}

// Children sets the children (strings or Components)
func (b *BoxBuilder) Children(children ...interface{}) *BoxBuilder {
	for _, child := range children {
		switch v := child.(type) {
		case string:
			b.children = append(b.children, v)
		case *BoxBuilder:
			b.children = append(b.children, v.ToMarkup())
		case *TextBuilder:
			b.children = append(b.children, v.ToMarkup())
		case Component:
			b.children = append(b.children, v.Render())
		}
	}
	return b
}

// ToMarkup converts the builder to markup string
func (b *BoxBuilder) ToMarkup() string {
	attrs := []string{}

	if b.id != "" {
		attrs = append(attrs, fmt.Sprintf(`id="%s"`, b.id))
	}
	if b.class != "" {
		attrs = append(attrs, fmt.Sprintf(`class="%s"`, b.class))
	}
	if b.direction != "" {
		attrs = append(attrs, fmt.Sprintf(`direction="%s"`, b.direction))
	}
	if b.flex > 0 {
		attrs = append(attrs, fmt.Sprintf(`flex="%d"`, b.flex))
	}
	if b.width != "" {
		attrs = append(attrs, fmt.Sprintf(`width="%s"`, b.width))
	}
	if b.height != "" {
		attrs = append(attrs, fmt.Sprintf(`height="%s"`, b.height))
	}
	if b.padding != "" {
		attrs = append(attrs, fmt.Sprintf(`padding="%s"`, b.padding))
	}
	if b.margin != "" {
		attrs = append(attrs, fmt.Sprintf(`margin="%s"`, b.margin))
	}
	if b.border != "" {
		attrs = append(attrs, fmt.Sprintf(`border="%s"`, b.border))
	}
	if b.borderStyle != "" {
		attrs = append(attrs, fmt.Sprintf(`border-style="%s"`, b.borderStyle))
	}
	if b.color != "" {
		attrs = append(attrs, fmt.Sprintf(`color="%s"`, b.color))
	}
	if b.background != "" {
		attrs = append(attrs, fmt.Sprintf(`background="%s"`, b.background))
	}

	attrString := ""
	if len(attrs) > 0 {
		attrString = " " + strings.Join(attrs, " ")
	}

	childrenString := strings.Join(b.children, "\n")

	return fmt.Sprintf(`<box%s>%s</box>`, attrString, childrenString)
}

// Render implements Component interface
func (b *BoxBuilder) Render() string {
	return b.ToMarkup()
}

// TextBuilder provides a fluent API for building text elements
type TextBuilder struct {
	content    string
	color      string
	background string
	bold       bool
}

// NewText creates a new TextBuilder
func NewText(content string) *TextBuilder {
	return &TextBuilder{
		content: content,
	}
}

// Color sets the text color
func (t *TextBuilder) Color(color string) *TextBuilder {
	t.color = color
	return t
}

// Background sets the background color
func (t *TextBuilder) Background(bg string) *TextBuilder {
	t.background = bg
	return t
}

// Bold makes the text bold
func (t *TextBuilder) Bold() *TextBuilder {
	t.bold = true
	return t
}

// ToMarkup converts the builder to markup string
func (t *TextBuilder) ToMarkup() string {
	attrs := []string{}

	if t.color != "" {
		attrs = append(attrs, fmt.Sprintf(`color="%s"`, t.color))
	}
	if t.background != "" {
		attrs = append(attrs, fmt.Sprintf(`background="%s"`, t.background))
	}
	if t.bold {
		attrs = append(attrs, `bold="true"`)
	}

	attrString := ""
	if len(attrs) > 0 {
		attrString = " " + strings.Join(attrs, " ")
	}

	return fmt.Sprintf(`<text%s>%s</text>`, attrString, t.content)
}

// Render implements Component interface
func (t *TextBuilder) Render() string {
	return t.ToMarkup()
}
