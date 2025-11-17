package runtime

import "github.com/speier/smith/pkg/lotus/vdom"

// DevToolsProvider is the interface for DevTools integration
type DevToolsProvider interface {
	Log(format string, args ...interface{})
	Render() *vdom.Element
	Enable()
	Disable()
	IsEnabled() bool
}

// HMRManager is the interface for HMR integration
type HMRManager interface {
	Start() error
	Stop() error
	SetCleanupHandler(func())
	SetExitHandler(func())
	ExecRestart(statePath string) error
}

// Global factories (set by devtools package to avoid import cycle)
var devToolsFactory func() DevToolsProvider                    //nolint:unused // Set by devtools package init()
var hmrFactory func(App, DevToolsProvider) (HMRManager, error) //nolint:unused // Set by devtools package init()

// SetDevToolsFactory registers the DevTools constructor
func SetDevToolsFactory(factory func() DevToolsProvider) {
	devToolsFactory = factory
}

// SetHMRFactory registers the HMR constructor
func SetHMRFactory(factory func(App, DevToolsProvider) (HMRManager, error)) {
	hmrFactory = factory
}

// wrapWithDevTools wraps the app element with DevTools panel based on position
func wrapWithDevTools(app *vdom.Element, devToolsPanel *vdom.Element, devTools DevToolsProvider) *vdom.Element {
	// Type assert to get position (if available)
	type positionGetter interface {
		GetPosition() string
	}

	position := "right" // Default
	if pg, ok := devTools.(positionGetter); ok {
		position = pg.GetPosition()
	}

	// Style DevTools panel with border and dark background
	// Put overflow on the inner content, border on the outer box
	scrollableContent := vdom.Box(devToolsPanel).
		WithStyle("background-color", "#1a1a1a").
		WithOverflow("auto")

	styledDevTools := vdom.Box(scrollableContent).
		WithBorderStyle(vdom.BorderStyleRounded)

	switch position {
	case "right":
		// Right side: HStack with app (60%) + devtools (40%)
		// Apply flex-grow directly to preserve app's internal flex behavior
		return vdom.HStack(
			app.Clone().WithFlexGrow(3),              // 60% (3/5) - app maintains internal flex
			vdom.Box(styledDevTools).WithFlexGrow(2), // 40% (2/5) - wrap to prevent mutation
		)

	case "bottom":
		// Bottom: VStack with app (70%) + devtools (30%)
		return vdom.VStack(
			app.Clone().WithFlexGrow(7),              // 70% - app maintains internal flex
			vdom.Box(styledDevTools).WithFlexGrow(3), // 30% - wrap to prevent mutation
		)

	case "left":
		// Left side: HStack with devtools (40%) + app (60%)
		return vdom.HStack(
			vdom.Box(styledDevTools).WithFlexGrow(2), // 40% (2/5) - wrap to prevent mutation
			app.Clone().WithFlexGrow(3),              // 60% (3/5) - app maintains internal flex
		)

	default:
		// Fallback to right
		return vdom.HStack(
			app.Clone().WithFlexGrow(3),
			vdom.Box(styledDevTools).WithFlexGrow(2),
		)
	}
}
