package core

// Component represents a reusable UI component that can render itself to markup
type Component interface {
	Render() string
}
