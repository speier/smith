package devtools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/speier/smith/pkg/lotus/runtime"
)

// HMR manages hot module reloading via process restart
type HMR struct {
	app        runtime.App
	devtools   *DevTools
	watcher    *Watcher
	workingDir string
	enabled    bool
}

// NewHMR creates a new HMR manager
func NewHMR(app runtime.App, devtools *DevTools) (*HMR, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	hmr := &HMR{
		app:        app,
		devtools:   devtools,
		workingDir: wd,
		enabled:    true,
	}

	return hmr, nil
}

// Start begins watching for file changes
func (h *HMR) Start() error {
	if !h.enabled {
		return nil
	}

	// Create file watcher with 300ms debounce
	watcher, err := NewWatcher(300, h.onFileChange)
	if err != nil {
		return err
	}

	h.watcher = watcher

	// Watch current directory
	if err := watcher.Watch(h.workingDir); err != nil {
		return err
	}

	if h.devtools != nil {
		h.devtools.Log("🔥 HMR enabled - watching *.go files")
	}

	return nil
}

// onFileChange is called when a file changes (debounced)
func (h *HMR) onFileChange(filename string) {
	if h.devtools != nil {
		relPath, _ := filepath.Rel(h.workingDir, filename)
		h.devtools.Log("🔥 HMR: %s changed", relPath)
	}

	// Trigger rebuild and restart
	h.restart()
}

// restart saves state, rebuilds, and restarts the process
func (h *HMR) restart() {
	startTime := time.Now()

	// 1. Save state automatically (library handles it!)
	statePath := fmt.Sprintf("/tmp/lotus-state-%d.json", os.Getpid())
	if err := runtime.SaveAppState(h.app, statePath); err == nil {
		if h.devtools != nil {
			h.devtools.Log("💾 Saved state")
		}
	}

	// 2. Rebuild
	if h.devtools != nil {
		h.devtools.Log("⚡ Rebuilding...")
	}

	// Use 'go build' to rebuild current package
	buildCmd := exec.Command("go", "build", "-o", "/tmp/lotus-hmr-app")
	buildCmd.Dir = h.workingDir
	if err := buildCmd.Run(); err != nil {
		if h.devtools != nil {
			h.devtools.Log("❌ Build failed: %v", err)
		}
		return
	}

	// 3. Prepare restart with state restoration
	restartCmd := exec.Command("/tmp/lotus-hmr-app")
	restartCmd.Dir = h.workingDir
	restartCmd.Stdout = os.Stdout
	restartCmd.Stderr = os.Stderr
	restartCmd.Stdin = os.Stdin

	// Set state path for automatic restoration
	restartCmd.Env = append(os.Environ(), fmt.Sprintf("LOTUS_STATE_PATH=%s", statePath))

	// 4. Log timing
	elapsed := time.Since(startTime)
	if h.devtools != nil {
		h.devtools.Log("✅ Reloaded in %dms", elapsed.Milliseconds())
	}

	// 5. Replace current process
	if err := restartCmd.Start(); err != nil {
		if h.devtools != nil {
			h.devtools.Log("❌ Restart failed: %v", err)
		}
		return
	}

	// Exit current process (new process takes over)
	os.Exit(0)
}

// Stop stops the HMR watcher
func (h *HMR) Stop() error {
	if h.watcher != nil {
		return h.watcher.Stop()
	}
	return nil
}
