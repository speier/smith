package parser

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		markup   string
		expected string
	}{
		{
			name:     "simple box",
			markup:   `<box id="test">Hello</box>`,
			expected: "box",
		},
		{
			name:     "nested elements",
			markup:   `<box><text>Hello</text></box>`,
			expected: "box",
		},
		{
			name:     "self-closing",
			markup:   `<input prompt="> " />`,
			expected: "input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := Parse(tt.markup)
			if node == nil {
				t.Fatal("Parse returned nil")
			}
			if node.Type != tt.expected {
				t.Errorf("expected type %s, got %s", tt.expected, node.Type)
			}
		})
	}
}

func TestParseCSS(t *testing.T) {
	css := `
		.container {
			width: 100%;
			height: 100%;
			display: flex;
		}
		#prompt {
			height: 5;
			border: 1px solid;
		}
	`

	styles := ParseCSS(css)

	if len(styles) != 2 {
		t.Errorf("expected 2 rules, got %d", len(styles))
	}

	container := styles[".container"]
	if container["width"] != "100%" {
		t.Errorf("expected width 100%%, got %s", container["width"])
	}

	prompt := styles["#prompt"]
	if prompt["height"] != "5" {
		t.Errorf("expected height 5, got %s", prompt["height"])
	}
}

func TestCSSComments(t *testing.T) {
	css := `
		/* This is a comment */
		#header {
			color: #0f0; /* inline comment */
		}
	`

	styles := ParseCSS(css)

	if len(styles) != 1 {
		t.Errorf("expected 1 rule, got %d", len(styles))
	}

	header := styles["#header"]
	if header["color"] != "#0f0" {
		t.Errorf("expected color #0f0, got %s", header["color"])
	}
}

func TestMultiClassSelector(t *testing.T) {
	css := `
		.message { padding: 0 1 0 1; }
		.message.system { text-align: center; color: 10; }
		.message.user { color: 12; }
	`
	markup := `<box class="message system">Centered text</box>`

	styles := ParseCSS(css)
	node := Parse(markup)
	ApplyStyles(node, styles)

	// Should have both "message" and "system" classes
	if len(node.Classes) != 2 {
		t.Errorf("expected 2 classes, got %d", len(node.Classes))
	}

	// Should apply .message styles
	if node.Styles.PaddingLeft != 1 || node.Styles.PaddingRight != 1 {
		t.Errorf("expected padding 1, got left=%d right=%d", node.Styles.PaddingLeft, node.Styles.PaddingRight)
	}

	// Should apply .message.system styles (multi-class selector)
	if node.Styles.TextAlign != "center" {
		t.Errorf("expected text-align center, got %s", node.Styles.TextAlign)
	}
	if node.Styles.Color != "10" {
		t.Errorf("expected color 10, got %s", node.Styles.Color)
	}
}

func TestMultiClassSelectorSpecificity(t *testing.T) {
	css := `
		.message { text-align: left; }
		.message.system { text-align: center; }
	`
	markup := `<box class="message system">Text</box>`

	styles := ParseCSS(css)
	node := Parse(markup)
	ApplyStyles(node, styles)

	// Multi-class selector should override single-class selector
	if node.Styles.TextAlign != "center" {
		t.Errorf("expected text-align center from .message.system, got %s", node.Styles.TextAlign)
	}
}
