package devtools

import (
	"fmt"
	"time"

	"github.com/speier/smith/pkg/lotus/components"
	"github.com/speier/smith/pkg/lotus/core"
	"github.com/speier/smith/pkg/lotus/runtime"
)

func init() {
	// Register DevTools factory with runtime to avoid import cycles
	runtime.SetDevToolsFactory(func() runtime.DevToolsProvider {
		return New()
	})
}

// DevTools provides an in-app debug console (like browser DevTools)
type DevTools struct {
	messageList *components.MessageList
	enabled     bool
}

// New creates a new DevTools instance
func New() *DevTools {
	dt := &DevTools{
		messageList: components.NewMessageList("devtools"),
		enabled:     true,
	}

	// Configure for debug display
	dt.messageList.AutoScroll = true

	dt.Log("🛠️  DevTools initialized")

	return dt
}

// Log adds a timestamped debug message (printf-style)
func (dt *DevTools) Log(format string, args ...interface{}) {
	if !dt.enabled {
		return
	}

	timestamp := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf(format, args...)

	dt.messageList.AddMessage("debug", fmt.Sprintf("[%s] %s", timestamp, msg))
}

// GetID returns the component ID
func (dt *DevTools) GetID() string {
	return "devtools-panel"
}

// Render returns the DevTools panel as an Element for integration
func (dt *DevTools) Render() *core.Element {
	if !dt.enabled {
		return nil
	}

	return core.NewComponentElement(dt.messageList)
}

// Enable turns DevTools on
func (dt *DevTools) Enable() {
	dt.enabled = true
	dt.Log("🛠️  DevTools enabled")
}

// Disable turns DevTools off
func (dt *DevTools) Disable() {
	dt.enabled = false
}

// IsEnabled returns whether DevTools is currently enabled
func (dt *DevTools) IsEnabled() bool {
	return dt.enabled
}

// SetDimensions sets the width and height for the MessageList
func (dt *DevTools) SetDimensions(width, height int) {
	dt.messageList.SetDimensions(width, height)
}
