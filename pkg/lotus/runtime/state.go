package runtime

import (
	"encoding/json"
	"os"

	"github.com/speier/smith/pkg/lotus/vdom"
)

// Stateful is an interface for components that can save/restore state
type Stateful interface {
	SaveState() map[string]interface{}
	LoadState(map[string]interface{}) error
	GetID() string
}

// SaveAppState saves app state to a JSON file for HMR
func SaveAppState(app App, path string) error {
	state := map[string]interface{}{
		"version": "1.0",
	}

	// Traverse the element tree and collect state from stateful components
	element := app.Render()
	components := collectStatefulComponents(element)

	if len(components) > 0 {
		componentStates := make(map[string]interface{})
		for _, comp := range components {
			if comp.GetID() != "" {
				componentStates[comp.GetID()] = comp.SaveState()
			}
		}
		state["components"] = componentStates
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// LoadAppState loads app state from a JSON file for HMR
func LoadAppState(app App, path string) error {
	// Check if state file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // No state to restore
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	// Restore component state
	if componentStates, ok := state["components"].(map[string]interface{}); ok {
		element := app.Render()
		components := collectStatefulComponents(element)

		for _, comp := range components {
			if comp.GetID() != "" {
				if compState, exists := componentStates[comp.GetID()]; exists {
					if stateMap, ok := compState.(map[string]interface{}); ok {
						_ = comp.LoadState(stateMap)
					}
				}
			}
		}
	}

	return nil
}

// CollectStatefulComponents traverses the element tree and collects stateful components
// Exported for DevTools to check for missing IDs
func CollectStatefulComponents(element *vdom.Element) []Stateful {
	return collectStatefulComponents(element)
}

// collectStatefulComponents traverses the element tree and collects stateful components
func collectStatefulComponents(element *vdom.Element) []Stateful {
	var components []Stateful

	if element == nil {
		return components
	}

	// Check if this element wraps a stateful component
	if element.Component != nil {
		if stateful, ok := element.Component.(Stateful); ok {
			components = append(components, stateful)
		}
	}

	// Recurse into children
	for _, child := range element.Children {
		components = append(components, collectStatefulComponents(child)...)
	}

	return components
}
