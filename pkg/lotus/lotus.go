// Package lotus provides a terminal UI library with HTML/CSS-like markup.
//
// Lotus is a declarative UI framework for building terminal applications.
// It features a Yoga-inspired flexbox layout engine, CSS-like styling,
// and multiple APIs to suit different preferences.
//
// # Architecture
//
// Lotus is organized into several layers:
//
//   - layout/    - Pure flexbox layout engine (render-agnostic)
//   - parser/    - HTML/CSS-like markup parsing
//   - render/    - Terminal rendering (ANSI escape codes)
//   - terminal/  - Terminal I/O (keyboard, screen management)
//
// # Usage Patterns
//
// Lotus supports three complementary APIs:
//
// 1. String Markup (Simple)
//
//	markup := `
//	    <box direction="column">
//	        <text>Hello World</text>
//	    </box>
//	`
//	ui := lotus.New(markup, css, width, height)
//
// 2. Helper Functions (Convenient)
//
//	markup := lotus.VStack(
//	    lotus.Text("Hello World"),
//	)
//	ui := lotus.New(markup, css, width, height)
//
// 3. Typed Builders (Type-safe)
//
//	markup := lotus.NewBox().
//	    Direction(lotus.Column).
//	    Children(
//	        lotus.NewText("Hello World"),
//	    ).
//	    ToMarkup()
//	ui := lotus.New(markup, css, width, height)
//
// # Components
//
// Create reusable components by implementing the Component interface:
//
//	type MessageList struct {
//	    messages []Message
//	}
//
//	func (m *MessageList) Render() string {
//	    return lotus.VStack(
//	        // ... render messages
//	    )
//	}
//
// # Markdown Support
//
// Render markdown content with glamour integration:
//
//	content := lotus.Markdown("# Hello\n\n**Bold** text", 80)
//
// # Performance
//
// CSS parsing is automatically cached for performance (~2x speedup).
// Disable caching for debugging:
//
//	lotus.SetCacheEnabled(false)
package lotus

import (
	"github.com/speier/smith/pkg/lotus/cache"
	"github.com/speier/smith/pkg/lotus/layout"
	"github.com/speier/smith/pkg/lotus/parser"
	renderterminal "github.com/speier/smith/pkg/lotus/render/terminal"
	"github.com/speier/smith/pkg/lotus/terminal"
)

// SetCacheEnabled enables or disables CSS caching globally.
// Disabling is useful for debugging CSS parsing issues.
// Default: enabled
func SetCacheEnabled(enabled bool) {
	cache.SetEnabled(enabled)
}

// ClearCache clears the CSS cache.
// Useful for testing or if you want to force re-parsing.
func ClearCache() {
	cache.Clear()
}

// UI represents a complete terminal UI
type UI struct {
	Root   *layout.Node
	Width  int
	Height int
}

// New creates a new terminal UI from markup and CSS
func New(markup, css string, width, height int) *UI {
	// Parse markup
	root := parser.Parse(markup)
	if root == nil {
		root = layout.NewNode("box")
	}

	// Parse and apply styles (with automatic caching)
	styles := cache.GetStyles(css)
	parser.ApplyStyles(root, styles)

	// Create UI
	ui := &UI{
		Root:   root,
		Width:  width,
		Height: height,
	}

	// Compute layout
	layout.Layout(root, width, height)

	return ui
}

// NewFullscreen creates a new terminal UI that auto-detects terminal size
func NewFullscreen(markup, css string) (*UI, error) {
	// Use terminal package for size detection
	term, err := terminal.New()
	if err != nil {
		// Fallback to default size
		return New(markup, css, 100, 40), nil
	}

	width, height := term.Size()
	return New(markup, css, width, height), nil
}

// RenderToTerminal renders the UI to a terminal string
func (ui *UI) RenderToTerminal() string {
	return renderterminal.Render(ui.Root)
}

// FindByID finds a node by its ID
func (ui *UI) FindByID(id string) *layout.Node {
	return ui.Root.FindByID(id)
}

// Reflow recomputes the layout with new dimensions
func (ui *UI) Reflow(width, height int) {
	ui.Width = width
	ui.Height = height
	layout.Layout(ui.Root, width, height)
}
