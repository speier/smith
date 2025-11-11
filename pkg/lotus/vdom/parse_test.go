package vdom

import "testing"

func TestMarkup(t *testing.T) {
	tests := []struct {
		name   string
		markup string
		want   string // Just check it doesn't crash
	}{
		{
			name:   "simple box",
			markup: `<box id="test">Hello</box>`,
		},
		{
			name:   "box with class",
			markup: `<box class="message user">Content</box>`,
		},
		{
			name:   "nested boxes",
			markup: `<box id="outer"><box id="inner">Text</box></box>`,
		},
		{
			name:   "text only",
			markup: `Hello World`,
		},
		{
			name:   "empty",
			markup: ``,
		},
		{
			name:   "style attribute with CSS",
			markup: `<box style="width: 80; color: #fff; padding: 2">styled</box>`,
		},
		{
			name:   "mixed attribute and style",
			markup: `<box width="100" style="color: red; border: 1">mixed</box>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elem := Markup(tt.markup)
			if elem == nil {
				t.Error("Markup returned nil")
			}
		})
	}
}

func TestParseInlineCSS(t *testing.T) {
	tests := []struct {
		name   string
		css    string
		expect map[string]string
	}{
		{
			name:   "single property",
			css:    "width: 80",
			expect: map[string]string{"width": "80"},
		},
		{
			name:   "multiple properties",
			css:    "width: 80; color: #fff; padding: 2",
			expect: map[string]string{"width": "80", "color": "#fff", "padding": "2"},
		},
		{
			name:   "trailing semicolon",
			css:    "width: 80;",
			expect: map[string]string{"width": "80"},
		},
		{
			name:   "extra whitespace",
			css:    "  width : 80 ;  color : red  ",
			expect: map[string]string{"width": "80", "color": "red"},
		},
		{
			name:   "empty",
			css:    "",
			expect: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &parser{}
			styles := make(map[string]string)
			p.parseInlineCSS(tt.css, styles)

			if len(styles) != len(tt.expect) {
				t.Errorf("got %d styles, want %d", len(styles), len(tt.expect))
			}

			for k, v := range tt.expect {
				if styles[k] != v {
					t.Errorf("styles[%q] = %q, want %q", k, styles[k], v)
				}
			}
		})
	}
}
