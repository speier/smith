package devtools

import (
	"fmt"
	"time"

	"github.com/speier/smith/pkg/lotus/runtime"
	"github.com/speier/smith/pkg/lotus/vdom"
)

func init() {
	// Register DevTools factory with runtime to avoid import cycles
	runtime.SetDevToolsFactory(func() runtime.DevToolsProvider {
		return New()
	})
}

// DevToolsPosition defines where the DevTools panel is displayed
type DevToolsPosition string

const (
	DevToolsRight  DevToolsPosition = "right"
	DevToolsBottom DevToolsPosition = "bottom"
	DevToolsLeft   DevToolsPosition = "left"
)

// DevTools provides an in-app debug console (like browser DevTools)
type DevTools struct {
	logs       []string
	enabled    bool
	position   DevToolsPosition
	onLogAdded func() // Callback to trigger re-render
}

// New creates a new DevTools instance
func New() *DevTools {
	dt := &DevTools{
		logs:     make([]string, 0),
		enabled:  true,
		position: DevToolsRight, // Default to right side like browser DevTools
	}

	dt.Log("🛠️ DevTools initialized")
	dt.Log("   Ctrl+T: Toggle DevTools | Ctrl+P: Cycle position")

	return dt
}

// Log adds a timestamped debug message (printf-style)
func (dt *DevTools) Log(format string, args ...interface{}) {
	if !dt.enabled {
		return
	}

	timestamp := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf(format, args...)

	dt.logs = append(dt.logs, fmt.Sprintf("[%s] %s", timestamp, msg))

	// Trigger re-render if callback is set
	if dt.onLogAdded != nil {
		dt.onLogAdded()
	}
}

// SetOnLogAdded sets a callback to trigger when logs are added
func (dt *DevTools) SetOnLogAdded(callback func()) {
	dt.onLogAdded = callback
}

// Render returns the DevTools panel as an Element for integration
func (dt *DevTools) Render() *vdom.Element {
	if !dt.enabled {
		return nil
	}

	// Simple rendering of logs as text elements
	children := make([]any, 0, len(dt.logs))
	for _, log := range dt.logs {
		children = append(children, vdom.Text(log))
	}

	return vdom.VStack(children...).WithPadding(1)
}

// Enable turns DevTools on
func (dt *DevTools) Enable() {
	dt.enabled = true
	dt.Log("🛠️ DevTools enabled")
}

// Disable turns DevTools off
func (dt *DevTools) Disable() {
	dt.enabled = false
}

// IsEnabled returns whether DevTools is currently enabled
func (dt *DevTools) IsEnabled() bool {
	return dt.enabled
}

// GetPosition returns the current DevTools panel position
func (dt *DevTools) GetPosition() string {
	return string(dt.position)
}

// SetPosition sets the DevTools panel position
func (dt *DevTools) SetPosition(pos DevToolsPosition) {
	dt.position = pos
	dt.Log("📍 DevTools position: %s", pos)
}

// CyclePosition cycles through available positions (right → bottom → left → right)
func (dt *DevTools) CyclePosition() {
	switch dt.position {
	case DevToolsRight:
		dt.SetPosition(DevToolsBottom)
	case DevToolsBottom:
		dt.SetPosition(DevToolsLeft)
	case DevToolsLeft:
		dt.SetPosition(DevToolsRight)
	default:
		dt.SetPosition(DevToolsRight)
	}
}
