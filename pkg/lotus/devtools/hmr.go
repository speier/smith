package devtools

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/speier/smith/pkg/lotus/runtime"
)

// HMR manages hot module reloading via process restart
type HMR struct {
	app            runtime.App
	devtools       *DevTools
	watcher        *Watcher
	workingDir     string
	buildTarget    string // The package/directory to build (e.g., "./pkg/lotus/examples/chat")
	enabled        bool
	cleanupHandler func()            // Called before restart to cleanup resources (e.g., restore terminal)
	exitHandler    func()            // Called to trigger app exit for restart
	fileHashes     map[string]string // Track file hashes to detect actual changes
}

func init() {
	// Register HMR factory with runtime to avoid import cycles
	runtime.SetHMRFactory(func(app runtime.App, dt runtime.DevToolsProvider) (runtime.HMRManager, error) {
		// Convert DevToolsProvider to *DevTools
		var devtools *DevTools
		if dt != nil {
			// Type assertion - we know it's our DevTools
			devtools, _ = dt.(*DevTools)
		}
		return NewHMR(app, devtools)
	})
}

// NewHMR creates a new HMR manager
func NewHMR(app runtime.App, devtools *DevTools) (*HMR, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Detect the build target from the running executable
	buildTarget := detectBuildTarget()

	hmr := &HMR{
		app:         app,
		devtools:    devtools,
		workingDir:  wd,
		buildTarget: buildTarget,
		enabled:     true,
		fileHashes:  make(map[string]string),
	}

	if devtools != nil {
		devtools.Log("🎯 Build target: %s", buildTarget)
	}

	return hmr, nil
}

// detectBuildTarget detects what package/directory should be built for HMR
func detectBuildTarget() string {
	// Check environment variable (for advanced usage)
	if target := os.Getenv("LOTUS_BUILD_TARGET"); target != "" {
		return target
	}

	// Default: build current directory
	// User should cd into the package directory before running
	return "."
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
		// Check for stateful components without IDs on startup
		h.checkStatefulIDs()
	}

	return nil
}

// onFileChange is called when a file changes (debounced)
func (h *HMR) onFileChange(filename string) {
	// Check if file actually changed by comparing hash
	newHash, err := fileHash(filename)
	if err != nil {
		// File might have been deleted or unreadable
		return
	}

	oldHash, exists := h.fileHashes[filename]
	if exists && oldHash == newHash {
		// File content unchanged, ignore
		return
	}

	// Update hash and trigger rebuild
	h.fileHashes[filename] = newHash

	if h.devtools != nil {
		relPath, _ := filepath.Rel(h.workingDir, filename)
		h.devtools.Log("🔥 HMR: %s changed", relPath)
	}

	// Trigger rebuild and restart
	h.restart()
}

// fileHash computes SHA256 hash of a file
func fileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// restart saves state, rebuilds, and restarts the process
func (h *HMR) restart() {
	startTime := time.Now()

	// 1. Check for stateful components without IDs and warn
	if h.devtools != nil {
		h.checkStatefulIDs()
	}

	// 2. Save state automatically (library handles it!)
	statePath := fmt.Sprintf("/tmp/lotus-state-%d.json", os.Getpid())
	if err := runtime.SaveAppState(h.app, statePath); err == nil {
		if h.devtools != nil {
			h.devtools.Log("💾 Saved state")
		}
	}

	// 2. Rebuild
	if h.devtools != nil {
		h.devtools.Log("⚡ Rebuilding %s...", h.buildTarget)
	}

	// Use 'go build' to rebuild the target package
	buildCmd := exec.Command("go", "build", "-o", "/tmp/lotus-hmr-app", h.buildTarget)
	buildCmd.Dir = h.workingDir

	// Capture build output
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		if h.devtools != nil {
			h.devtools.Log("❌ Build failed:")
			// Log build errors line by line
			lines := string(output)
			for i, line := range splitLines(lines) {
				if i < 10 { // Limit to first 10 lines
					h.devtools.Log("   %s", line)
				}
			}
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
		h.devtools.Log("✅ Rebuilt in %dms - restarting...", elapsed.Milliseconds())
	}

	// 5. Trigger app exit (which will cleanup terminal properly via defers)
	// Then the runtime will call execRestart()
	if h.exitHandler != nil {
		h.exitHandler()
	} else {
		// Fallback: just exit and let user restart manually
		os.Exit(0)
	}
}

// SetCleanupHandler sets a callback to cleanup resources before restart
func (h *HMR) SetCleanupHandler(handler func()) {
	h.cleanupHandler = handler
}

// SetExitHandler sets a callback to trigger app exit for restart
func (h *HMR) SetExitHandler(handler func()) {
	h.exitHandler = handler
}

// checkStatefulIDs warns about Stateful components without IDs
func (h *HMR) checkStatefulIDs() {
	element := h.app.Render(runtime.Context{})
	components := runtime.CollectStatefulComponents(element)

	var missingIDs []string
	for _, comp := range components {
		if comp.GetID() == "" {
			// Get component type name
			compType := fmt.Sprintf("%T", comp)
			missingIDs = append(missingIDs, compType)
		}
	}

	if len(missingIDs) > 0 {
		h.devtools.Log("⚠️ Stateful components without ID")
		h.devtools.Log("   (state won't persist):")
		for _, compType := range missingIDs {
			h.devtools.Log("   - %s", compType)
		}
	}
}

// ExecRestart replaces the current process with the rebuilt binary using syscall.Exec
// This should be called after the app has cleanly exited and terminal is restored
func (h *HMR) ExecRestart(statePath string) error {
	binaryPath := "/tmp/lotus-hmr-app"

	// Verify the new binary exists
	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("rebuilt binary not found: %w", err)
	}

	// Prepare environment with state path
	env := os.Environ()
	env = append(env, fmt.Sprintf("LOTUS_STATE_PATH=%s", statePath))

	// Use syscall.Exec to replace current process with new binary
	// This is Unix-specific but works on macOS and Linux
	return syscall.Exec(binaryPath, []string{binaryPath}, env)
}

// Stop stops the HMR watcher
func (h *HMR) Stop() error {
	if h.watcher != nil {
		return h.watcher.Stop()
	}
	return nil
}

// splitLines splits a string into lines
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
