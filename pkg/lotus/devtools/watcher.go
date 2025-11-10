package devtools

import (
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
	// Watch directory
	if err := w.watcher.Add(dir); err != nil {
		return err
	}

	// Start event processing
	go w.processEvents()

	return nil
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
