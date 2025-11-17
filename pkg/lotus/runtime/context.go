package runtime

import (
	"github.com/speier/smith/pkg/lotus/context"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// Context is an alias to context.Context
type Context = context.Context

// FunctionalComponent is a function that renders elements with access to Context
type FunctionalComponent func(Context) *vdom.Element

// functionalApp wraps a FunctionalComponent to satisfy App interface
type functionalApp struct {
	renderFn FunctionalComponent
	ctx      Context
}

func (f *functionalApp) Render(ctx Context) *vdom.Element {
	return f.renderFn(ctx)
}

func (f *functionalApp) GetContext() Context {
	return f.ctx
}

func (f *functionalApp) SetRenderCallback(callback func()) {
	f.ctx.RenderCallback = callback
}
