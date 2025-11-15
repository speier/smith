package primitives

// Context provides helpers for updating the UI from callbacks
// This is an interface to avoid import cycles (runtime implements it)
type Context interface {
	// Update triggers a re-render of the application
	// Use this in event callbacks or goroutines after modifying app state
	Update()

	// Rerender is an alias for Update
	Rerender()
}
