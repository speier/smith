package lotus

import (
	"github.com/speier/smith/pkg/lotus/commands"
	"github.com/speier/smith/pkg/lotus/primitives"
	"github.com/speier/smith/pkg/lotus/runtime"
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
	// Context provides access to app lifecycle methods
	// Usage: lotus.Run(func(ctx lotus.Context) *MyApp { ... })
	Context = runtime.Context

	// KeyHandler is a function that handles keyboard events
	// Returns true if the event was handled (stops propagation)
	KeyHandler = runtime.KeyHandler

	// KeyBinding represents a registered keyboard shortcut
	KeyBinding = runtime.KeyBinding
)

// Runtime functions
var (
	// Run starts a Lotus terminal application
	// Accepts: App interface, functional component, Element, or markup string
	Run = runtime.Run

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

// Strings converts []string to []any for use in VStack/HStack
// Strings are auto-converted to Text elements
// Usage: lotus.VStack(lotus.Strings(messages)...)
func Strings(items []string) []any {
	return vdom.Strings(items)
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
	// CreateInput creates a single-line text input field
	// Usage: lotus.CreateInput("placeholder", onSubmit)
	CreateInput = primitives.CreateInput

	// Input is an alias for CreateInput (backward compatibility)
	Input = primitives.CreateInput

	// CreateTextArea creates a multi-line text input field
	// Usage: lotus.CreateTextArea("placeholder", onSubmit)
	CreateTextArea = primitives.CreateTextArea

	// TextArea is an alias for CreateTextArea (backward compatibility)
	TextArea = primitives.CreateTextArea
)

// Command system types
type (
	// Command represents a slash command that can be executed
	Command = commands.Command

	// CommandRegistry manages registered slash commands
	CommandRegistry = commands.CommandRegistry
)

// Command system functions
var (
	// NewCommandRegistry creates a new command registry
	NewCommandRegistry = commands.NewCommandRegistry

	// RegisterGlobalCommand registers a command globally (available in all apps)
	// Usage: lotus.RegisterGlobalCommand("clear", "Clear screen", handler)
	RegisterGlobalCommand = commands.RegisterGlobalCommand

	// GetGlobalCommands returns the global command registry
	GetGlobalCommands = commands.GetGlobalCommands

	// SetGlobalCommandPrefix sets the command prefix (default: "/")
	// Examples: lotus.SetGlobalCommandPrefix("!") for !help, !stream
	// Examples: lotus.SetGlobalCommandPrefix("@bot ") for @bot help
	SetGlobalCommandPrefix = commands.SetGlobalCommandPrefix
)
