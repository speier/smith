package devtools

import (
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches files for changes and triggers callbacks
type Watcher struct {
	watcher    *fsnotify.Watcher
	debounceMs int
	onChange   func(string)
	stopChan   chan bool
}

// NewWatcher creates a new file watcher
func NewWatcher(debounceMs int, onChange func(string)) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		watcher:    fsWatcher,
		debounceMs: debounceMs,
		onChange:   onChange,
		stopChan:   make(chan bool),
	}

	return w, nil
}

// Watch starts watching a directory recursively for .go files
func (w *Watcher) Watch(dir string) error {
	// Watch directory recursively
	if err := w.addRecursive(dir); err != nil {
		return err
	}

	// Start event processing
	go w.processEvents()

	return nil
}

// addRecursive adds a directory and all subdirectories to the watcher
func (w *Watcher) addRecursive(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and common build/vendor directories
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" ||
				name == ".idea" || name == ".vscode" || name[0] == '.' {
				return filepath.SkipDir
			}

			// Add directory to watcher
			if err := w.watcher.Add(path); err != nil {
				// Ignore errors for directories we can't watch
				return nil
			}
		}

		return nil
	})
}

// processEvents handles file system events with debouncing
func (w *Watcher) processEvents() {
	var debounceTimer *time.Timer
	var lastFile string

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			// Only watch .go files
			if filepath.Ext(event.Name) != ".go" {
				continue
			}

			// Only care about write events
			if event.Op&fsnotify.Write != fsnotify.Write {
				continue
			}

			lastFile = event.Name

			// Reset debounce timer
			if debounceTimer != nil {
				debounceTimer.Stop()
			}

			debounceTimer = time.AfterFunc(time.Duration(w.debounceMs)*time.Millisecond, func() {
				if w.onChange != nil {
					w.onChange(lastFile)
				}
			})

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			// Ignore errors for now (could log to DevTools)
			_ = err

		case <-w.stopChan:
			return
		}
	}
}

// Stop stops the watcher
func (w *Watcher) Stop() error {
	close(w.stopChan)
	return w.watcher.Close()
}
