package runtime

import "github.com/speier/smith/pkg/lotus/vdom"

// Context provides helpers for functional components
type Context struct {
	renderCallback func() // Callback to trigger re-render
}

// Rerender triggers a re-render of the application
// Use this in event callbacks to update the UI after state changes
func (ctx Context) Rerender() {
	if ctx.renderCallback != nil {
		ctx.renderCallback()
	}
}

// Update triggers a re-render (alias for Rerender)
// Use this in goroutines or async callbacks after modifying app state
func (ctx Context) Update() {
	ctx.Rerender()
}

// FunctionalComponent is a function that renders elements with access to Context
type FunctionalComponent func(Context) *vdom.Element

// functionalApp wraps a FunctionalComponent to satisfy App interface
type functionalApp struct {
	renderFn FunctionalComponent
	ctx      Context
}

func (f *functionalApp) Render() *vdom.Element {
	return f.renderFn(f.ctx)
}

func (f *functionalApp) GetContext() Context {
	return f.ctx
}

func (f *functionalApp) SetRenderCallback(callback func()) {
	f.ctx.renderCallback = callback
}
