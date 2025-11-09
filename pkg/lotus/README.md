# Lotus UI Package

A minimal HTML/CSS-like markup system for building terminal UIs in Go.

## Features

- **HTML-like markup** for defining UI structure
- **CSS-like styling** for visual properties
- **Flexbox layout** for responsive designs
- **Border rendering** with multiple styles
- **Color support** via ANSI escape codes
- **Declarative API** - describe what you want, not how to render it

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/speier/smith/pkg/lotus"
)

func main() {
    markup := `
        <box id="container">
            <box id="header">My App</box>
            <box id="content"></box>
            <box id="prompt">
                <input prompt="> " />
            </box>
        </box>
    `

    css := `
        #container {
            display: flex;
            flex-direction: column;
            height: 100%;
        }
        #header { height: 3; text-align: center; color: #0f0; }
        #content { flex: 1; }
        #prompt { height: 5; border: 1px solid; padding: 1; }
    `

    ui := lotus.New(markup, css, 100, 40)
    fmt.Print(ui.RenderToTerminal())
}
```

## Supported HTML Elements

- `<box>` - Container element (like `<div>`)
- `<text>` - Text content
- `<input>` - Input field with prompt

### Attributes

- `id="name"` - Element identifier
- `class="name1 name2"` - CSS classes
- `prompt="> "` - Input prompt (for `<input>` only)

## Supported CSS Properties

### Layout
- `width: 100% | 50 | auto`
- `height: 100% | 10 | auto`
- `display: block | flex`
- `flex-direction: column | row`
- `flex: 0 | 1 | 2` - Flex grow factor
- `position: static | fixed`

### Spacing
- `padding: 1` - All sides
- `padding: 1 2` - Vertical horizontal
- `padding: 1 2 3 4` - Top right bottom left
- `margin: 1` - Same syntax as padding

### Visual
- `color: #0f0 | #00ff00` - Text color (hex)
- `background-color: #000` - Background color (hex)
- `border: 1px solid` - Border (must include "solid")
- `border-style: single | rounded | double` - Border characters
- `text-align: left | center | right`

### Positioning (fixed only)
- `top: 5` - Distance from top
- `bottom: 5` - Distance from bottom
- `left: 5` - Distance from left
- `right: 5` - Distance from right

## CSS Selectors

```css
/* Element selector */
box { color: #fff; }

/* Class selector */
.container { display: flex; }

/* ID selector */
#prompt { height: 5; }
```

## Layout System

### Block Layout
Default layout mode. Children stack vertically.

```html
<box>
    <box>First</box>
    <box>Second</box>
</box>
```

### Flex Layout
Use `display: flex` for flexible layouts.

```html
<box id="container">
    <box id="sidebar"></box>
    <box id="main"></box>
</box>
```

```css
#container { display: flex; flex-direction: row; }
#sidebar { width: 20; }
#main { flex: 1; } /* Takes remaining space */
```

### Fixed Positioning
Position elements relative to terminal edges.

```css
#prompt {
    position: fixed;
    bottom: 0;
    height: 5;
}
```

## Color Support

Supports basic hex colors mapped to ANSI 256 colors:

- `#0f0` / `#00ff00` - Bright green
- `#0ff` / `#00ffff` - Bright cyan
- `#fff` / `#ffffff` - Bright white
- `#f00` / `#ff0000` - Bright red
- `#ff0` / `#ffff00` - Bright yellow
- `#00f` / `#0000ff` - Bright blue
- `#444` - Dark gray
- `#888` - Light gray

## API Reference

### Creating a UI

```go
// New creates a UI from markup and CSS
ui := lotus.New(markup, css, width, height)
```

### Rendering

```go
// Render to terminal
output := ui.RenderToTerminal()
fmt.Print(output)
```

### Finding Elements

```go
// Find element by ID
element := ui.FindByID("prompt")
```

### Reflowing

```go
// Recompute layout with new dimensions
ui.Reflow(newWidth, newHeight)
```

## Performance

Lotus automatically caches parsed CSS for performance. CSS parsing only happens once per unique CSS string, dramatically improving performance for applications that re-render frequently (e.g., interactive TUIs).

**Benchmark Results:**
```
BenchmarkCachedCSS-8      862592    1394 ns/op    3576 B/op    28 allocs/op
BenchmarkUncachedCSS-8    398514    2919 ns/op    6336 B/op    61 allocs/op
```

**~2x faster** with caching enabled (default).

### Disabling Cache (for debugging)

```go
// Disable caching globally (useful for debugging CSS)
lotus.SetCacheEnabled(false)

// Re-enable
lotus.SetCacheEnabled(true)

// Clear cache manually
lotus.ClearCache()
```

**Note:** Caching is enabled by default. You typically don't need to manage it manually.


## Examples

See `examples/basic/main.go` for a complete example.

Run it:
```bash
go run ./pkg/lotus/examples/basic/main.go
```

## Architecture

Lotus is a **terminal-only** UI library with clean separation of concerns:

- `layout/` - Pure flexbox layout engine (Yoga-inspired, render-agnostic)
- `parser/` - HTML/CSS-like markup parsing
- `render/terminal/` - ANSI terminal renderer
- `terminal/` - Terminal I/O (keyboard, screen management)
- `ui.go` - High-level API

**Why "Lotus"?**

The name reflects the architecture: a layout engine grounded in terminal rendering. Like a lotus rooted in mud but reaching upward, Lotus has a clean layout core with terminal-specific rendering on top.

### FAQ

**Can Lotus render to HTML or GUI?**

No, Lotus is terminal-only. The name refers to the architecture (grounded layout engine, terminal rendering), not multiple output targets. If you need web UI, use actual HTML/CSS. If you need desktop GUI, use a proper GUI framework.

The architecture separation (layout vs rendering) is valuable even with only terminal output - it keeps the codebase clean and testable.


## Current Limitations

- No text wrapping (text must fit in box)
- Limited color palette (basic ANSI colors only)
- No scrolling (yet)
- No event handling (yet)
- No animations
- Single-line text only

## Future Enhancements

- [ ] Browser preview server
- [ ] Text wrapping and overflow
- [ ] Scrollable regions
- [ ] Event handling (input, clicks)
- [ ] More border styles
- [ ] Gradient backgrounds
- [ ] Media queries for terminal size
- [ ] Component library

## License

MIT
