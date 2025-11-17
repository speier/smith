package context

// Context provides helpers for updating the UI from callbacks
type Context struct {
	RenderCallback func() // Callback to trigger re-render
}

// Update triggers a re-render of the application
// Use this in event callbacks or goroutines after modifying app state
func (ctx Context) Update() {
	if ctx.RenderCallback != nil {
		ctx.RenderCallback()
	}
}
