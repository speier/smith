// Package lotus provides a React-like terminal UI framework with Virtual DOM diffing.
//
// Lotus brings the best of modern web development to the terminal:
// declarative components, Virtual DOM performance, CSS-like styling, and
// flexbox layout. It's the fastest, smallest, and most intuitive Go TUI library.
//
// # Quick Start
//
// Build terminal UIs the React way:
//
//	type ChatApp struct {
//	    messages []string
//	}
//
//	func (app *ChatApp) Render() *lotus.Element {
//	    return lotus.VStack(
//	        lotus.Text("💬 Chat Room"),
//	        lotus.NewTextInput("input"),
//	    ).Render()
//	}
//
//	func main() {
//	    lotus.Run("chat", &ChatApp{})
//	}
//
// # Three APIs, One Engine
//
// Choose your style - all produce the same Virtual DOM tree:
//
// 1. JSX-like Markup (Simple)
//
//	markup := `<box direction="column"><text>Hello</text></box>`
//	ui := lotus.NewUI(markup, "", width, height)
//
// 2. React Helpers (Recommended)
//
//	elem := lotus.VStack(
//	    lotus.Text("Hello"),
//	    lotus.HStack(
//	        lotus.Text("Left"),
//	        lotus.Text("Right"),
//	    ),
//	).Render()
//
// 3. Type-Safe Builders (Advanced)
//
//	elem := lotus.Box("container",
//	    lotus.Text("Hello"),
//	).Direction(lotus.Column).Render()
//
// # Virtual DOM Performance
//
// Lotus uses React-like Virtual DOM diffing for optimal performance:
//
//   - 245ns per text update (10-40x faster than alternatives)
//   - 48-192 bytes per frame (10-20x less memory)
//   - Automatic CSS caching (0 allocations when cached)
//
// # Components
//
// Built-in components: TextInput, MessageList, InputBox, Panel, Header,
// ProgressBar, Menu, Dialog, Tabs.
//
// # Performance
//
// CSS parsing is automatically cached for performance (~2x speedup).
// Disable caching for debugging:
//
//	lotus.SetCacheEnabled(false)
package lotus

import (
	"fmt"

	_ "github.com/speier/smith/pkg/lotus/devtools" // Register DevTools factory
	"github.com/speier/smith/pkg/lotus/components"
	"github.com/speier/smith/pkg/lotus/core"
	"github.com/speier/smith/pkg/lotus/layout"
	"github.com/speier/smith/pkg/lotus/reconciler"
	"github.com/speier/smith/pkg/lotus/runtime"
) // Component re-exports - users only need to import lotus
type (
	// Focusable is the interface for components that can receive focus and keyboard input
	Focusable = components.Focusable

	// TextInput is an interactive text input component with editing capabilities
	TextInput = components.TextInput
	// TextInputProps configures TextInput component (placeholder, callbacks, etc.)
	TextInputProps = components.TextInputProps

	// MessageList displays a scrollable list of messages
	MessageList = components.MessageList
	// Message represents a single message in a MessageList
	Message = components.Message

	// InputBox combines a label with a TextInput component
	InputBox = components.InputBox

	// Panel is a bordered container component
	Panel = components.Panel

	// Header is a styled header component
	Header = components.Header

	// ProgressBar shows progress visualization
	ProgressBar = components.ProgressBar

	// Menu is an interactive menu component
	Menu = components.Menu
	// MenuItem represents a single item in a Menu
	MenuItem = components.MenuItem

	// Dialog shows modal dialog boxes
	Dialog = components.Dialog

	// Tabs provides tabbed interface navigation
	Tabs = components.Tabs
	// Tab represents a single tab in Tabs
	Tab = components.Tab
)

// Component constructors
var (
	// NewTextInput creates a new text input component
	NewTextInput = components.NewTextInput
	// NewMessageList creates a new message list component
	NewMessageList = components.NewMessageList
	// NewInputBox creates a new input box component
	NewInputBox = components.NewInputBox
	// NewPanel creates a new panel component
	NewPanel = components.NewPanel
	// NewHeader creates a new header component
	NewHeader = components.NewHeader
	// NewProgressBar creates a new progress bar component
	NewProgressBar = components.NewProgressBar
	// NewMenu creates a new menu component
	NewMenu = components.NewMenu
	// NewDialog creates a new dialog component
	NewDialog = components.NewDialog
	// NewTabs creates a new tabs component
	NewTabs = components.NewTabs
)

// SetCacheEnabled enables or disables CSS caching globally.
// Disabling is useful for debugging CSS parsing issues.
// Default: enabled
func SetCacheEnabled(enabled bool) {
	reconciler.SetEnabled(enabled)
}

// ClearCache clears the CSS reconciler.
// Useful for testing or if you want to force re-parsing.
func ClearCache() {
	reconciler.Clear()
}

// NewStyle creates a new Style instance
func NewStyle() layout.Style {
	return layout.NewStyle()
}

// Re-export core types and functions
type (
	// Element represents a virtual DOM element (React pattern).
	// This is a type alias for core.Element to hide internal package structure.
	//
	// Usage in App.Render():
	//
	//	func (app *MyApp) Render() *lotus.Element {
	//	    return lotus.VStack(...).Render()  // Returns *lotus.Element
	//	}
	//
	// Note: *lotus.Element and *core.Element are the exact same type.
	// Your IDE may show a warning about type mismatch - this is a gopls limitation
	// with cross-package type aliases. The Go compiler correctly accepts both.
	// See: https://go.dev/ref/spec#Type_identity
	Element = core.Element

	// Component is the base interface for anything that can render to markup
	// Use this for type-safe component composition
	Component = core.Component

	// ComponentRenderer interface for components that can render themselves
	// Extends Component with GetID() for auto-registration
	ComponentRenderer = reconciler.ComponentRenderer

	// App interface represents a Lotus application (React.Component pattern)
	App = runtime.App

	// RenderFunc is a functional component (React hooks pattern)
	// Allows you to write components as pure functions instead of structs
	RenderFunc = runtime.RenderFunc

	// UI represents a complete terminal UI
	UI = reconciler.UI
)

// Re-export functions
var (
	// NewMarkupElement creates an element from markup string
	NewMarkupElement = core.NewMarkupElement

	// NewComponentElement creates an element from a component
	NewComponentElement = core.NewComponentElement

	// NewContainerElement creates an element with children
	NewContainerElement = core.NewContainerElement

	// Run creates and runs a Lotus terminal app (ReactDOM.render pattern)
	Run = runtime.Run

	// RunWith provides more control over terminal app configuration
	RunWith = runtime.RunWith

	// WithDevTools enables or disables the DevTools panel
	WithDevTools = runtime.WithDevTools

	// WithHMR enables or disables Hot Module Reload
	WithHMR = runtime.WithHMR

	// Render renders markup/CSS with automatic reconciliation
	Render = reconciler.Render

	// RenderWithElement renders using Element tree (enables Virtual DOM diffing)
	RenderWithElement = reconciler.RenderWithElement

	// GetContext returns the UI instance for a context
	GetContext = reconciler.GetContext

	// DestroyContext removes a render context (cleanup)
	DestroyContext = reconciler.DestroyContext

	// RegisterComponent registers a focusable component in a context
	RegisterComponent = reconciler.RegisterComponent

	// SetFocus sets focus in a context
	SetFocus = reconciler.SetFocus

	// New creates a new terminal UI from markup and CSS
	New = reconciler.NewUI

	// NewFullscreen creates a new terminal UI that auto-detects terminal size
	NewFullscreen = reconciler.NewFullscreenUI
)

// Builder helpers - JSX-like functions that return ElementBuilder with fluent style API

// Box creates a box with ID and fluent style API
func Box(id string, children ...interface{}) *ElementBuilder {
	return &ElementBuilder{
		id:       id,
		builder:  core.NewBox().ID(id),
		children: children,
		styles:   make(map[string]string),
	}
}

// VStack creates a vertical stack with fluent style API
func VStack(children ...interface{}) *ElementBuilder {
	return &ElementBuilder{
		builder:  core.NewBox().Direction(core.Column),
		children: children,
		styles:   make(map[string]string),
	}
}

// HStack creates a horizontal stack with fluent style API
func HStack(children ...interface{}) *ElementBuilder {
	return &ElementBuilder{
		builder:  core.NewBox().Direction(core.Row),
		children: children,
		styles:   make(map[string]string),
	}
}

// PanelBox creates a bordered panel with fluent style API
func PanelBox(id string, children ...interface{}) *ElementBuilder {
	return &ElementBuilder{
		id:       id,
		builder:  core.NewBox().ID(id).Border("1px solid"),
		children: children,
		styles:   make(map[string]string),
	}
}

// ElementBuilder wraps BoxBuilder and provides Element-based API with inline styles
type ElementBuilder struct {
	id       string
	builder  *core.BoxBuilder
	children []interface{}
	styles   map[string]string
}

// Style methods (React inline style pattern)

// Height sets height
func (b *ElementBuilder) Height(height int) *ElementBuilder {
	if b.styles == nil {
		b.styles = make(map[string]string)
	}
	b.styles["height"] = fmt.Sprintf("%d", height)
	return b
}

// Flex sets flex grow
func (b *ElementBuilder) Flex(flex int) *ElementBuilder {
	b.builder.Flex(flex)
	if b.styles == nil {
		b.styles = make(map[string]string)
	}
	b.styles["flex"] = fmt.Sprintf("%d", flex)
	return b
}

// Border sets border
func (b *ElementBuilder) Border(border string) *ElementBuilder {
	b.builder.Border(border)
	if b.styles == nil {
		b.styles = make(map[string]string)
	}
	b.styles["border"] = border
	return b
}

// BorderStyle sets border style
func (b *ElementBuilder) BorderStyle(style string) *ElementBuilder {
	if b.styles == nil {
		b.styles = make(map[string]string)
	}
	b.styles["border-style"] = style
	return b
}

// BorderColor sets border color
func (b *ElementBuilder) BorderColor(color string) *ElementBuilder {
	if b.styles == nil {
		b.styles = make(map[string]string)
	}
	b.styles["border-color"] = color
	return b
}

// Padding sets padding
func (b *ElementBuilder) Padding(padding int) *ElementBuilder {
	if b.styles == nil {
		b.styles = make(map[string]string)
	}
	b.styles["padding"] = fmt.Sprintf("%d", padding)
	return b
}

// Color sets text color
func (b *ElementBuilder) Color(color string) *ElementBuilder {
	if b.styles == nil {
		b.styles = make(map[string]string)
	}
	b.styles["color"] = color
	return b
}

// Width sets width
func (b *ElementBuilder) Width(width string) *ElementBuilder {
	if b.styles == nil {
		b.styles = make(map[string]string)
	}
	b.styles["width"] = width
	return b
}

// Children adds child elements
func (b *ElementBuilder) Children(children ...interface{}) *ElementBuilder {
	b.children = append(b.children, children...)
	return b
}

// Key sets a unique key for diffing (like React's key prop)
// Use this for list items to optimize re-rendering
func (b *ElementBuilder) Key(key string) *ElementBuilder {
	b.id = key
	return b
}

// Render converts to Element with inline styles
func (b *ElementBuilder) Render() *Element {
	elements := make([]*Element, 0, len(b.children))
	for _, child := range b.children {
		elements = append(elements, toElement(child))
	}
	markup := b.builder.Children(convertToMarkup(elements)...).Render()

	container := NewContainerElement(elements...)

	// Use provided key/id, or auto-generate if not set
	if b.id != "" {
		container.ID = b.id
	}

	container.Markup = markup
	container.Type = "container-with-markup"
	container.Styles = b.styles
	return container
}

// Text creates a text element
func Text(content string) *Element {
	return NewMarkupElement(core.NewText(content).Render())
}

// Helper: Convert interface{} to Element
func toElement(child interface{}) *Element {
	switch v := child.(type) {
	case *Element:
		return v
	case *ElementBuilder:
		// Render the builder to get the Element
		return v.Render()
	case ComponentRenderer:
		return NewComponentElement(v)
	case string:
		return NewMarkupElement(v)
	default:
		return NewMarkupElement("")
	}
}

// Helper: Convert Elements to markup strings for engine.BoxBuilder
func convertToMarkup(elements []*Element) []interface{} {
	markups := make([]interface{}, len(elements))
	for i, el := range elements {
		markups[i] = el.ToMarkup()
	}
	return markups
}

// IsValidComponent checks if a value can be used as a component
// Valid types: *Element, *ElementBuilder, ComponentRenderer, string, Component
func IsValidComponent(v interface{}) bool {
	switch v.(type) {
	case *Element, *ElementBuilder, ComponentRenderer, string, Component:
		return true
	default:
		return false
	}
}

// MustBeComponent panics if the value is not a valid component
// Useful for catching component type errors early in development
func MustBeComponent(v interface{}) {
	if !IsValidComponent(v) {
		panic(fmt.Sprintf("Invalid component type: %T\n\nValid component types:\n  - *lotus.Element (created with lotus.Text(), lotus.Box(), etc.)\n  - *lotus.ElementBuilder (from lotus.VStack(), lotus.HStack())\n  - lotus.ComponentRenderer (custom components with GetID() and Render())\n  - string (raw markup)\n\nDid you forget to call .Render() on your builder?", v))
	}
}
