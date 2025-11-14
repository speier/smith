package runtime

import "sync"

// scrollManager tracks scroll state for elements with overflow: auto
type scrollManager struct {
	mu      sync.RWMutex
	offsets map[string]*scrollState // key: element ID
}

type scrollState struct {
	offsetX int
	offsetY int
	// Content and viewport dimensions (for boundary checking)
	contentWidth   int
	contentHeight  int
	viewportWidth  int
	viewportHeight int
}

func newScrollManager() *scrollManager {
	return &scrollManager{
		offsets: make(map[string]*scrollState),
	}
}

// GetOffset returns scroll offset for an element (or 0,0 if none)
func (sm *scrollManager) GetOffset(id string) (int, int) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if state, ok := sm.offsets[id]; ok {
		return state.offsetX, state.offsetY
	}
	return 0, 0
}

// SetOffset updates scroll offset for an element
func (sm *scrollManager) SetOffset(id string, x, y int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	state, ok := sm.offsets[id]
	if !ok {
		state = &scrollState{}
		sm.offsets[id] = state
	}

	// Clamp to valid range
	state.offsetX = max(0, min(x, state.contentWidth-state.viewportWidth))
	state.offsetY = max(0, min(y, state.contentHeight-state.viewportHeight))
}

// UpdateDimensions updates content/viewport size for an element
func (sm *scrollManager) UpdateDimensions(id string, contentW, contentH, viewportW, viewportH int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	state, ok := sm.offsets[id]
	if !ok {
		state = &scrollState{}
		sm.offsets[id] = state
	}

	oldContentHeight := state.contentHeight

	state.contentWidth = contentW
	state.contentHeight = contentH
	state.viewportWidth = viewportW
	state.viewportHeight = viewportH

	// Auto-scroll to bottom when content grows (like chat messages)
	// If we were already at the bottom, stay at the bottom
	wasAtBottom := oldContentHeight > 0 && state.offsetY >= oldContentHeight-state.viewportHeight
	if wasAtBottom || oldContentHeight == 0 {
		// Scroll to bottom to show new content
		state.offsetY = max(0, state.contentHeight-state.viewportHeight)
	}

	// Re-clamp offsets after dimension update
	state.offsetX = max(0, min(state.offsetX, state.contentWidth-state.viewportWidth))
	state.offsetY = max(0, min(state.offsetY, state.contentHeight-state.viewportHeight))
}

// ScrollBy adjusts scroll by delta (returns true if changed)
func (sm *scrollManager) ScrollBy(id string, dx, dy int) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	state, ok := sm.offsets[id]
	if !ok {
		return false
	}

	oldX, oldY := state.offsetX, state.offsetY

	// Clamp to valid range
	state.offsetX = max(0, min(state.offsetX+dx, state.contentWidth-state.viewportWidth))
	state.offsetY = max(0, min(state.offsetY+dy, state.contentHeight-state.viewportHeight))

	return state.offsetX != oldX || state.offsetY != oldY
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
