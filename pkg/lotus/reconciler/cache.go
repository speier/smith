package reconciler

import (
	"sync"

	"github.com/speier/smith/pkg/lotus/parser"
)

var (
	// cssCache stores parsed CSS styles
	// Key: CSS string, Value: parsed styles
	cssCache     = make(map[string]map[string]map[string]string)
	cacheMutex   sync.RWMutex
	cacheEnabled = true
)

// SetEnabled enables or disables CSS caching globally.
// Disabling is useful for debugging CSS parsing issues.
// Default: enabled
func SetEnabled(enabled bool) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	cacheEnabled = enabled
	if !enabled {
		// Clear cache when disabled
		cssCache = make(map[string]map[string]map[string]string)
	}
}

// Clear empties the CSS engine.
func Clear() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	clear(cssCache)
}

// Size returns the number of entries in the cache (for testing).
func Size() int {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	return len(cssCache)
}

// GetStyles returns cached styles or parses and caches them
func GetStyles(css string) map[string]map[string]string {
	if !cacheEnabled {
		return parser.ParseCSS(css)
	}

	// Try read lock first
	cacheMutex.RLock()
	if styles, ok := cssCache[css]; ok {
		cacheMutex.RUnlock()
		return styles
	}
	cacheMutex.RUnlock()

	// Not in cache, parse and store
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Double-check after acquiring write lock
	if styles, ok := cssCache[css]; ok {
		return styles
	}

	// Parse and cache
	styles := parser.ParseCSS(css)
	cssCache[css] = styles
	return styles
}
