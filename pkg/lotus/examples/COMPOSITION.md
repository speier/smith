# Lotus Component Composition Guide

This guide demonstrates **React-like component composition** in Lotus with **hybrid type safety**.

## Type Safety: The Hybrid Approach

Lotus uses a **hybrid approach** for component types:

- **Accepts `interface{}`** for flexibility (like React/JSX accepting anything)
- **Provides `Component` interface** for type checking when needed
- **Runtime validation** with clear error messages

### Valid Component Types

```go
// ✅ All these are valid components:
lotus.VStack(
    "plain string",                    // string
    lotus.Text("styled text"),         // *lotus.Element
    lotus.HStack(...),                 // *lotus.ElementBuilder
    Card(CardProps{...}),              // interface{} from helper
    myTextInput,                       // lotus.ComponentRenderer
)
```

### Type Safety Helpers

```go
// Check if something is a valid component
if lotus.IsValidComponent(widget) {
    // Safe to use
}

// Validate component type (panics if invalid)
lotus.MustBeComponent(widget)

// Invalid types cause runtime panics with clear messages:
lotus.VStack(42)  // Panic: "Invalid component type: int (must be *Element, *ElementBuilder, ComponentRenderer, or string)"
```

### Why `interface{}` Instead of `Component`?

**Flexibility wins:**
```go
// With interface{} (current) - Easy & flexible ✅
lotus.VStack("Hello", "World")

// With strict Component - Verbose ❌
lotus.VStack(lotus.Text("Hello"), lotus.Text("World"))
```

The hybrid approach gives you **type safety where it matters** without sacrificing **ease of use**.

## Core Principles

### 1. Components are Functions
Just like React functional components, Lotus components are simple functions that return Elements:

```go
// React
function Card({ title, description }) {
  return <div className="card">...</div>
}

// Lotus
func Card(title, description string) interface{} {
  return lotus.VStack(
    lotus.Text(title),
    lotus.Text(description),
  )
}
```

### 2. Props Pattern
Use structs for component props (like React props):

```go
type CardProps struct {
    Title       string
    Description string
    Color       string
}

func Card(props CardProps) interface{} {
    return lotus.VStack(
        lotus.Text(props.Title),
        lotus.Text(props.Description),
    ).Color(props.Color)
}

// Usage
Card(CardProps{
    Title: "Hello",
    Description: "World",
    Color: "blue",
})
```

### 3. Composition Over Inheritance
Build complex UIs by composing simple components:

```go
// Simple components
func Header(text string) interface{} {
    return lotus.Text(text).Color("blue")
}

func Content(text string) interface{} {
    return lotus.Text(text).Padding(1)
}

// Compose them
func Page(title, body string) interface{} {
    return lotus.VStack(
        Header(title),
        Content(body),
    )
}
```

### 4. Children as Variadic Args
Pass children directly in constructors (not via `.Children()` method):

```go
// ✅ Good - children in constructor
lotus.VStack(
    child1,
    child2,
    child3,
)

// ❌ Avoid - mixing patterns
lotus.VStack().Children(child1, child2, child3)

// ✅ Good for dynamic lists
children := []interface{}{child1, child2, child3}
lotus.VStack(children...)
```

### 5. Extract Helper Methods
Break down complex render logic (like React):

```go
type Dashboard struct {
    users []User
}

func (d *Dashboard) Render() *lotus.Element {
    return lotus.VStack(
        d.renderHeader(),
        d.renderUserList(),
        d.renderFooter(),
    ).Render()
}

func (d *Dashboard) renderHeader() interface{} {
    return lotus.Text("Dashboard")
}

func (d *Dashboard) renderUserList() interface{} {
    cards := make([]interface{}, len(d.users))
    for i, user := range d.users {
        cards[i] = UserCard(user)
    }
    return lotus.VStack(cards...)
}
```

## API Consistency

### Constructor Children Pattern

All builders accept children in their constructor:

```go
// Box with ID and children
lotus.Box("mybox", child1, child2)

// VStack with children
lotus.VStack(child1, child2, child3)

// HStack with children
lotus.HStack(child1, child2)

// PanelBox with ID and children
lotus.PanelBox("panel", child1, child2)
```

### Fluent Style API

Chain style methods after construction:

```go
lotus.Box("mybox", children...).
    Height(10).
    Color("blue").
    Border("1px solid").
    Padding(1)
```

### When to Call `.Render()`

- **Top-level Render()**: Always call `.Render()` to return `*Element`
- **Helper methods**: Return `interface{}` for composition (no `.Render()`)
- **Text()**: Already returns `*Element`, no `.Render()` needed

```go
// ✅ Main Render - returns *Element
func (app *App) Render() *lotus.Element {
    return lotus.VStack(
        app.renderHeader(),
    ).Render()  // ← .Render() here!
}

// ✅ Helper - returns interface{} for composition
func (app *App) renderHeader() interface{} {
    return lotus.Text("Header")  // ← NO .Render()
}
```

## Examples

See:
- `chat-tui/main.go` - Chat app with component extraction
- `composition/main.go` - Full composition patterns with props

## Performance Tips

1. **Reuse components** - Don't recreate the same component repeatedly
2. **Extract static parts** - Move unchanging UI to separate functions
3. **Use Props** - Makes components reusable and testable

## Next Steps

- Component state management (hooks-like pattern)
- Virtual DOM diffing (performance optimization)
- Component lifecycle hooks
