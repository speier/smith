package lotus

import (
	"github.com/speier/smith/pkg/lotus/primitives"
	"github.com/speier/smith/pkg/lotus/runtime"
	"github.com/speier/smith/pkg/lotus/vdom"

	// Import devtools to register factories
	_ "github.com/speier/smith/pkg/lotus/devtools"
)

// Core types
type (
	Element = vdom.Element
	Node    = vdom.Node
)

// Style types for type-safe styling
type (
	BorderStyle = vdom.BorderStyle
	AlignSelf   = vdom.AlignSelf
	TextAlign   = vdom.TextAlign
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

// Text align constants
const (
	TextAlignLeft   = vdom.TextAlignLeft
	TextAlignCenter = vdom.TextAlignCenter
	TextAlignRight  = vdom.TextAlignRight
)

// Runtime
var (
	Run = runtime.Run
)

// VDOM primitives - JSX-like API
var (
	Box    = vdom.Box
	VStack = vdom.VStack
	HStack = vdom.HStack
	Text   = vdom.Text
	Markup = vdom.Markup
)

// UI Primitives - browser equivalents
type (
	Input      = primitives.Input
	TextArea   = primitives.TextArea
	ScrollView = primitives.ScrollView
)

// Primitive constructor functions
var (
	NewInput      = primitives.NewInput
	NewTextArea   = primitives.NewTextArea
	NewScrollView = primitives.NewScrollView
)
