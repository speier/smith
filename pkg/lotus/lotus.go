package lotus

import (
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
	// Context provides access to app lifecycle methods in functional components
	// Usage: lotus.Run(func(ctx lotus.Context) *lotus.Element { ... })
	Context = runtime.Context
)

// Runtime functions
var (
	// Run starts a Lotus terminal application
	// Accepts: App interface, functional component, Element, or markup string
	Run = runtime.Run
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
