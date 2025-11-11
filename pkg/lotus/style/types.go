package style

// BorderStyle represents the visual style of borders
type BorderStyle string

const (
	BorderStyleSingle  BorderStyle = "single"  // ┌─┐ Default box drawing
	BorderStyleRounded BorderStyle = "rounded" // ╭─╮ Smooth corners
	BorderStyleDouble  BorderStyle = "double"  // ╔═╗ Double lines
	BorderStyleNone    BorderStyle = "none"    // No border
)

// AlignSelf represents cross-axis alignment for flex items
type AlignSelf string

const (
	AlignSelfStretch   AlignSelf = "stretch"    // Fill cross-axis (default)
	AlignSelfFlexStart AlignSelf = "flex-start" // Align to start
	AlignSelfFlexEnd   AlignSelf = "flex-end"   // Align to end
	AlignSelfCenter    AlignSelf = "center"     // Center on cross-axis
)

// TextAlign represents horizontal text alignment
type TextAlign string

const (
	TextAlignLeft   TextAlign = "left"   // Left-aligned (default)
	TextAlignCenter TextAlign = "center" // Centered
	TextAlignRight  TextAlign = "right"  // Right-aligned
)
