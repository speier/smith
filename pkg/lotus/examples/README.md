# Lotus Examples

Learn Lotus TUI framework through working examples.

## 📚 Examples

### 1. [Quickstart](./quickstart) - Hello World
**10 lines** - Your first Lotus app

```bash
cd quickstart && go run main.go
```

**Learn:** Functional components, basic layout, text styling

---

### 2. [Chat](./chat) - Production-Ready Chat UI
**~30 lines** - Complete interactive chat application

```bash
cd chat && go run main.go
```

**Learn:** 
- State management with closures
- Auto-scrolling message history
- Simplified input API (`CreateInput`)
- Responsive flexbox layouts
- Superior rendering performance

**Try:**
- Type any message - echoes back
- Type `long` - stress test with 50 lines (watch auto-scroll!)

---

### 3. [Styling](./styling) - Text & Layout Features
**~100 lines** - Comprehensive styling showcase

```bash
cd styling && go run main.go
```

**Learn:**
- Text styles: bold, italic, underline, strikethrough, dim, reverse
- Colors: hex (`#ff0000`) and named (`bright-cyan`)
- Text alignment, overflow, max lines
- Borders and border colors
- Visibility control

---

## 🎯 Progressive Learning Path

1. **Start with `quickstart`** - Understand the basics
2. **Build the `chat` example** - Learn interactive UIs
3. **Explore `styling`** - Master visual design

## 🚀 Why Lotus?

```go
// Production chat in ~30 lines
lotus.Run(func(ctx lotus.AppContext) *lotus.Element {
    messages := []string{"Welcome!"}
    
    return lotus.VStack(
        lotus.Box(lotus.Text("Chat")).WithBorderStyle(lotus.BorderStyleRounded),
        lotus.Box(
            lotus.NewScrollView().
                WithAutoScroll(true).
                WithContent(lotus.VStack(renderMessages(messages)...)),
        ).WithFlexGrow(1).WithBorderStyle(lotus.BorderStyleRounded),
        lotus.Box(
            lotus.CreateInput("Type...", func(text string) {
                messages = append(messages, text)
                ctx.Rerender()
            }),
        ).WithBorderStyle(lotus.BorderStyleRounded),
    )
})
```

**Outstanding DX:**
- ✅ Functional components - no boilerplate
- ✅ Declarative UI - like React
- ✅ CSS Flexbox layouts - familiar patterns
- ✅ Simple state - closure variables
- ✅ Minimal code - maximum functionality

**Superior Performance:**
- ✅ Buffer-based rendering - only changed cells update
- ✅ 60+ FPS - smooth animations
- ✅ Flicker-free - CSI 2026 synchronized output
- ✅ Low CPU usage - efficient diffing

Perfect for: chat apps, dashboards, logs, developer tools
