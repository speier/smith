# Lotus 🪷

**React for Terminal UIs** - Build beautiful terminal applications with React-like components and flexbox layout.

[![Go Reference](https://pkg.go.dev/badge/github.com/speier/smith/pkg/lotus.svg)](https://pkg.go.dev/github.com/speier/smith/pkg/lotus)
[![Go Report Card](https://goreportcard.com/badge/github.com/speier/smith)](https://goreportcard.com/report/github.com/speier/smith)

Lotus brings modern web development to the terminal:

- ⚛️ **React-like Components** - Familiar component-based architecture
- 📐 **Flexbox Layout** - CSS-style flexbox (inspired by Yoga)
- 🎨 **Clean Architecture** - Modular pipeline: vdom → style → layout → render
- 🪶 **Lightweight & Fast** - Pure Go, minimal dependencies
- 🔧 **Extractable Modules** - Use vdom, layout, or render independently

## Architecture

Lotus uses a clean separation of concerns:

```
vdom (Virtual DOM) → style (CSS) → layout (Flexbox) → render (ANSI)
```

Each module is independently usable and testable.



## Quick Start

```go
package main

import (
	"github.com/speier/smith/pkg/lotus"
	"github.com/speier/smith/pkg/lotus/components"
)

type ChatApp struct {
	messages []string
}

func (app *ChatApp) Render() *lotus.Element {
	return lotus.VStack(
		lotus.Box(
			lotus.Text("💬 Chat Room"),
		),
		lotus.Box(
			lotus.Text("Messages here..."),
		),
	)
}

func main() {
	app := &ChatApp{
		messages: []string{"Welcome!"},
	}
	lotus.Run(app)
}
```

That's it! No message passing, no update functions, just declarative UI.

## Installation

```bash
go get github.com/speier/smith/pkg/lotus
```



## Core Concepts

### Virtual DOM

Create UI trees with React-like helpers:

```go
import "github.com/speier/smith/pkg/lotus"

// Vertical stack
ui := lotus.VStack(
	lotus.Text("Title"),
	lotus.Box(
		lotus.Text("Content"),
	),
)

// Horizontal stack
row := lotus.HStack(
	lotus.Text("Left"),
	lotus.Text("Right"),
)
```

### Markup Strings

Quick prototyping with markup syntax:

```go
markup := `
	<box direction="column">
		<text>Hello World</text>
	</box>
`
lotus.Run(markup)
```

### Components

Build reusable components with state:

```go
import "github.com/speier/smith/pkg/lotus/components"

input := components.NewTextInput().
	WithPlaceholder("Type here...").
	WithOnSubmit(func(value string) {
		// Handle submission
	})
```

### Flexbox Layout

CSS-like flexbox properties via inline styles:

```go
elem := lotus.Box(
	lotus.Text("Content"),
)
elem.Styles = map[string]string{
	"flex":    "1",
	"width":   "50%",
	"height":  "10",
	"padding": "1",
}
```

## Features

### Core
- ⚡ **Virtual DOM Diffing** - Only update what changed
- 🎨 **Flexbox Layout** - Modern CSS-like layout engine
- 🎯 **Automatic Focus Management** - Tab through components automatically
- 🔄 **Auto Terminal Resize** - Handles window size changes gracefully
- 📦 **Component System** - Reusable, composable UI components

### Built-in Components
- `TextInput` - Full-featured text input with editing, cursor, scrolling
- `TextBox` - Multi-line text display with auto-scroll
- `ProgressBar` - Progress visualization
- `Tabs` - Tabbed interfaces
- `Select` - Dropdown selection
- `Checkbox` - Toggle checkboxes
- `Radio` - Radio button groups
- `Modal` - Modal dialogs
- `ScrollView` - Scrollable containers

### Performance
- **245ns** - Text update with Virtual DOM diffing
- **152ns** - Complex tree diffing (10 elements)
- **0 allocations** - CSS parsing when cached
- **48-192 bytes/frame** - Minimal memory footprint

## Examples

### Simple Text Display

```go
func (app *App) Render() *lotus.Element {
	return lotus.Text("Hello, World!")
}
```

### Interactive Form

```go
import "github.com/speier/smith/pkg/lotus/components"

type FormApp struct {
	nameInput  *components.TextInput
	emailInput *components.TextInput
}

func NewFormApp() *FormApp {
	return &FormApp{
		nameInput:  components.NewTextInput(),
		emailInput: components.NewTextInput(),
	}
}

func (app *FormApp) Render() *lotus.Element {
	return lotus.VStack(
		lotus.Text("Name:"),
		app.nameInput.Render(),
		lotus.Text("Email:"),
		app.emailInput.Render(),
		lotus.Text("Press Tab to switch fields"),
	)
}

func main() {
	lotus.Run(NewFormApp())
}
```

### Dashboard Layout

```go
func (app *DashboardApp) Render() *lotus.Element {
	return lotus.VStack(
		// Header
		lotus.Text("📊 Dashboard"),
		
		// Main content - horizontal split
		lotus.HStack(
			// Sidebar (30% width)
			lotus.VStack(
				lotus.Text("📁 Files"),
				lotus.Text("⚙️ Settings"),
			),
			
			// Main area
			lotus.VStack(
				lotus.Text("Welcome!"),
			),
		),
	)
}
```

See `examples/chat/` for a complete chat application.

## API Reference

### Application Lifecycle

```go
// App interface - implement Render() to describe your UI
type App interface {
	Render() *lotus.Element
}

// Run your app (accepts App, *Element, or markup string)
lotus.Run(app)
```

### Layout Helpers

```go
// Vertical stack (flex-direction: column)
lotus.VStack(children...)

// Horizontal stack (flex-direction: row)
lotus.HStack(children...)

// Box container
lotus.Box(children...)

// Text node
lotus.Text("content")

// Markup parsing
lotus.Markup("<box><text>Hello</text></box>")
```

### Components

All components are in `github.com/speier/smith/pkg/lotus/components`:

```go
input := components.NewTextInput().
	WithPlaceholder("Type here...").
	WithOnSubmit(func(value string) {
		// Handle submission
	})

textbox := components.NewTextBox().
	WithAutoScroll(true)

progress := components.NewProgressBar(50).
	SetValue(0.75)

tabs := components.NewTabs().
	AddTab("Tab 1", content1).
	AddTab("Tab 2", content2)
```

### Styling

Elements support inline styling via the `Styles` map:

```go
elem := lotus.Box(lotus.Text("Content"))
elem.Styles = map[string]string{
	"flex":           "1",
	"width":          "50%",
	"height":         "10",
	"padding":        "1",
	"border":         "1px solid",
	"flex-direction": "column",
}
```

## Performance

### Virtual DOM Diffing

Lotus uses React-like Virtual DOM diffing for optimal performance:

```go
// First render - builds initial tree
app.Render()  // Full render

// Subsequent renders - only updates changes
app.Render()  // Diff + patch (245ns for simple changes)
```

**Benchmarks:**
```
BenchmarkDiff_TextChange        245ns    48 B/op   2 allocs/op
BenchmarkDiff_ComplexTree       152ns    80 B/op   4 allocs/op  
BenchmarkCachedCSS             13.9ns     0 B/op   0 allocs/op
```

### CSS Caching

CSS parsing is automatically cached:

```go
// Disable caching for debugging
lotus.SetCacheEnabled(false)

// Clear cache
lotus.ClearCache()
```
## Architecture

Lotus is organized into clean, focused layers:

```
pkg/lotus/
├── vdom/          # Virtual DOM - Element tree, builders
├── style/         # CSS resolution - Compute final styles
├── layout/        # Flexbox engine - Yoga-inspired layout math
├── render/        # ANSI renderer - Terminal rendering
├── runtime/       # App lifecycle - Event loop, terminal I/O
├── tty/           # Terminal I/O - Keyboard, screen management
├── components/    # Built-in components
└── devtools/      # DevTools & HMR for development
```

**Design Principles:**
- **Separation of concerns** - Each layer is independent and testable
- **Performance** - Virtual DOM diffing + CSS caching
- **Simplicity** - React-like declarative API
- **Clean pipeline** - vdom → style → layout → render

## Testing

Lotus has comprehensive test coverage:

```bash
# Run all tests
go test ./pkg/lotus/...

# Run with coverage
go test -cover ./pkg/lotus/...

# Run benchmarks
go test -bench=. -benchmem ./pkg/lotus/...
```

**Test Stats:**
- ✅ 100% passing tests
- ✅ Comprehensive component tests
- ✅ Full benchmark suite

## Comparison with Other Libraries

### vs BubbleTea

**BubbleTea (Elm Architecture):**
```go
// Message-driven, requires boilerplate
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle each key...
	}
	return m, nil
}
```

**Lotus (React-like):**
```go
// Just describe what you want
func (app *App) Render() *lotus.Element {
	return lotus.VStack(
		lotus.Text("Hello"),
		input,
	)
}
```

**Verdict:** Lotus is simpler and faster (245ns vs 5µs per update).

## FAQ

### Is Lotus production-ready?

Yes! Lotus powers real applications and includes:
- Complete Virtual DOM implementation with diffing
- Comprehensive component library
- Full test coverage with benchmarks
- Clean, maintainable codebase

### How is performance compared to other libraries?

Lotus achieves excellent performance through Virtual DOM diffing:
- **245ns** for typical updates (text changes)
- **152ns** for complex tree diffing
- **0 allocations** when CSS is cached
- **Minimal memory** footprint per frame

### Can I use Lotus for web/GUI?

No, Lotus is terminal-only. The architecture cleanly separates layout from rendering, but only a terminal renderer exists. For web UIs, use HTML/CSS. For desktop GUIs, use proper GUI frameworks.

### What about BubbleTea compatibility?

Lotus and BubbleTea serve different philosophies:
- **BubbleTea:** Elm architecture (message passing, functional)
- **Lotus:** React architecture (declarative, Virtual DOM)

Both are excellent - choose based on your preference!

## Development Tools

Enable DevTools and Hot Module Reload during development:

```bash
LOTUS_DEV=true go run main.go
```

**Features:**
- **DevTools Panel** - In-app debug console (toggle with `Ctrl+T`)
- **Hot Module Reload** - Auto-rebuild and restart on `.go` file changes
- **Build Errors** - Compile errors shown in DevTools panel

The watcher monitors all `.go` files recursively and debounces changes to avoid rebuild spam.

## Roadmap

Lotus is production-ready today. Future enhancements under consideration:

- [ ] Text wrapping and overflow handling
- [ ] Scrollable regions
- [ ] More border styles  
- [ ] Mouse event handling
- [ ] Animation support
- [ ] Extended color palettes
- [ ] Component library expansion

## Contributing

Lotus is part of the [Smith](https://github.com/speier/smith) project. See the main repository for contribution guidelines.

## License

MIT License - see LICENSE file for details.

## Credits

Built with inspiration from:
- **React** - Virtual DOM and declarative API
- **Yoga** - Flexbox layout algorithm  
- **BubbleTea** - Proving TUIs can be delightful

---

**Made with 🪷 by the Smith team**
