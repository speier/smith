package lotus

import (
	"github.com/speier/smith/pkg/lotus/context"
	"github.com/speier/smith/pkg/lotus/primitives"
	"github.com/speier/smith/pkg/lotus/runtime"
	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"

	// Import devtools to register factories
	_ "github.com/speier/smith/pkg/lotus/devtools"
)

// Core VDOM types
type (
	Element = vdom.Element
	Node    = vdom.Node
)

// Style types for type-safe styling
type (
	BorderStyle    = vdom.BorderStyle
	AlignSelf      = vdom.AlignSelf
	AlignItems     = vdom.AlignItems
	TextAlign      = vdom.TextAlign
	JustifyContent = vdom.JustifyContent
	Overflow       = vdom.Overflow
)

// Border style constants
const (
	BorderStyleSingle  = vdom.BorderStyleSingle
	BorderStyleRounded = vdom.BorderStyleRounded
	BorderStyleDouble  = vdom.BorderStyleDouble
	BorderStyleNone    = vdom.BorderStyleNone
)

// Align self constants
const (
	AlignSelfStretch   = vdom.AlignSelfStretch
	AlignSelfFlexStart = vdom.AlignSelfFlexStart
	AlignSelfFlexEnd   = vdom.AlignSelfFlexEnd
	AlignSelfCenter    = vdom.AlignSelfCenter
)

// Align items constants
const (
	AlignItemsStretch   = vdom.AlignItemsStretch
	AlignItemsFlexStart = vdom.AlignItemsFlexStart
	AlignItemsFlexEnd   = vdom.AlignItemsFlexEnd
	AlignItemsCenter    = vdom.AlignItemsCenter
)

// Text align constants
const (
	TextAlignLeft   = vdom.TextAlignLeft
	TextAlignCenter = vdom.TextAlignCenter
	TextAlignRight  = vdom.TextAlignRight
)

// Justify content constants
const (
	JustifyContentFlexStart    = vdom.JustifyContentFlexStart
	JustifyContentFlexEnd      = vdom.JustifyContentFlexEnd
	JustifyContentCenter       = vdom.JustifyContentCenter
	JustifyContentSpaceBetween = vdom.JustifyContentSpaceBetween
)

// Overflow constants
const (
	OverflowAuto   = vdom.OverflowAuto
	OverflowHidden = vdom.OverflowHidden
)

// Runtime types
type (
	// App interface for stateful applications
	// Implement this interface to create a Lotus app with state and lifecycle
	// Example:
	//   type MyApp struct { count int }
	//   func (a *MyApp) Render(ctx lotus.Context) *lotus.Element { ... }
	//   lotus.Run(&MyApp{})
	App = runtime.App

	// Context provides access to app lifecycle methods
	// Usage: func(ctx lotus.Context, value string)
	// Always pass context as first parameter (Go convention)
	Context = context.Context

	// KeyHandler is a function that handles keyboard events
	// Returns true if the event was handled (stops propagation)
	KeyHandler = runtime.KeyHandler

	// KeyBinding represents a registered keyboard shortcut
	KeyBinding = runtime.KeyBinding

	// KeyEvent represents a keyboard input event
	KeyEvent = tty.KeyEvent
)

// Key constants for special keys
const (
	KeyCtrlC     = tty.KeyCtrlC
	KeyCtrlD     = tty.KeyCtrlD
	KeyBackspace = tty.KeyBackspace
	KeyEnter     = tty.KeyEnter
	KeyEscape    = tty.KeyEscape
)

// Key sequence constants for arrow keys and special keys
const (
	SeqUp         = tty.SeqUp
	SeqDown       = tty.SeqDown
	SeqLeft       = tty.SeqLeft
	SeqRight      = tty.SeqRight
	SeqHome       = tty.SeqHome
	SeqEnd        = tty.SeqEnd
	SeqDelete     = tty.SeqDelete
	SeqCtrlLeft   = tty.SeqCtrlLeft
	SeqCtrlRight  = tty.SeqCtrlRight
	SeqShiftEnter = tty.SeqShiftEnter
	SeqPasteStart = tty.SeqPasteStart
	SeqPasteEnd   = tty.SeqPasteEnd
)

// Runtime functions
var (
	// Run starts a Lotus terminal application
	// Accepts types implementing App interface
	Run = runtime.Run

	// RunFunc runs a functional component
	// Accepts: func(lotus.Context) *lotus.Element
	RunFunc = runtime.RunFunc

	// RunElement runs a static element
	// Accepts: *lotus.Element
	RunElement = runtime.RunElement

	// RegisterGlobalKey registers a global keyboard shortcut
	// Example: lotus.RegisterGlobalKey('o', true, "Open file", handler) for Ctrl+O
	RegisterGlobalKey = runtime.RegisterGlobalKey

	// RegisterGlobalKeyCode registers a keyboard shortcut by escape sequence
	// Example: lotus.RegisterGlobalKeyCode(tty.SeqF1, "Show help", handler)
	RegisterGlobalKeyCode = runtime.RegisterGlobalKeyCode

	// UnregisterAllGlobalKeys clears all global key handlers
	UnregisterAllGlobalKeys = runtime.UnregisterAllGlobalKeys

	// GetGlobalKeyBindings returns all registered global key bindings
	GetGlobalKeyBindings = runtime.GetGlobalKeyBindings
)

// VDOM element constructors
var (
	// Box creates a container element (like HTML <div>)
	Box = vdom.Box

	// VStack creates a vertical stack container (flexbox column)
	VStack = vdom.VStack

	// HStack creates a horizontal stack container (flexbox row)
	HStack = vdom.HStack

	// Text creates a text element with the given content
	Text = vdom.Text

	// Markup creates an element from a markup string with optional data
	// Usage: Markup("Hello {0}!", name)
	Markup = vdom.Markup
)

// Map transforms a slice into Elements using a mapping function
// Generic wrapper around vdom.Map - works with any type
// Usage: lotus.Map(messages, lotus.Text)
func Map[T any](items []T, fn func(T) *Element) []any {
	return vdom.Map(items, fn)
}

// Input types
type (
	// InputType defines the type of input (like HTML input type attribute)
	InputType = primitives.InputType

	// InputComponent represents a single-line or multi-line text input component
	// Use this type when storing input references in structs
	InputComponent = primitives.Input
)

// Input type constants
const (
	InputTypeText     = primitives.InputTypeText     // Regular text input
	InputTypePassword = primitives.InputTypePassword // Masked password input
	InputTypeNumber   = primitives.InputTypeNumber   // Numeric input only
)

// Input components
var (
	// Input creates a single-line text input field
	// Usage: lotus.Input("placeholder", onSubmit)
	Input = primitives.CreateInput

	// TextArea creates a multi-line text input field
	// Usage: lotus.TextArea("placeholder", onSubmit)
	TextArea = primitives.CreateTextArea
)
