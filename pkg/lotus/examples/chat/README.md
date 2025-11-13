# Chat Example

Production-ready chat UI in ~30 lines of code.

Showcases Lotus's outstanding DX:
- ✅ Functional components - no boilerplate
- ✅ State management - simple closure variables
- ✅ Auto-scrolling messages - built-in ScrollView
- ✅ Simplified input API - `CreateInput(placeholder, onSubmit)`
- ✅ Responsive layout - CSS Flexbox-style
- ✅ Superior rendering - buffer-based diffing

## Run it

```bash
go run main.go
```

Try typing:
- Any message - echoes back
- `long` - stress test with 50 lines (watch auto-scroll!)

## Code highlights

### Functional component
```go
lotus.Run(func(ctx lotus.AppContext) *lotus.Element {
    messages := []string{"Welcome!"}  // State in closure
    return lotus.VStack(...)
})
```

### Auto-scrolling messages
```go
lotus.NewScrollView().
    WithAutoScroll(true).
    WithContent(lotus.VStack(messageElements...))
```

### Simplified input
```go
lotus.CreateInput("Type...", func(text string) {
    messages = append(messages, text)
    ctx.Rerender()  // Trigger update
})
```

## Architecture

```
┌─────────────────────┐
│   Functional        │ ← Your code (30 lines)
│   Component         │
├─────────────────────┤
│   VStack/Box/Input  │ ← Lotus primitives
├─────────────────────┤
│   Layout Engine     │ ← CSS Flexbox
├─────────────────────┤
│   Buffer Renderer   │ ← Diff-based ANSI
└─────────────────────┘
```

## Performance

Lotus renders via buffer diffing - only changed cells update:
- 60+ FPS on modern terminals
- Minimal CPU usage
- Flicker-free (CSI 2026 synchronized output)

Perfect for real-time chat, logs, dashboards.
