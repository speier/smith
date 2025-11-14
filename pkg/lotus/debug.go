package lotus

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	debugFile  *os.File
	debugMutex sync.Mutex
	debugPath  = "/tmp/lotus-debug.log"
)

// EnableDebugLog enables debug logging to /tmp/lotus-debug.log
// Call this before lotus.Run() to enable logging
func EnableDebugLog() error {
	debugMutex.Lock()
	defer debugMutex.Unlock()

	if debugFile != nil {
		return nil // Already enabled
	}

	f, err := os.Create(debugPath)
	if err != nil {
		return fmt.Errorf("failed to create debug log: %w", err)
	}

	debugFile = f
	DebugLog("=== Lotus Debug Log Started ===")
	return nil
}

// DebugLog writes a debug message to /tmp/lotus-debug.log
// Safe to call even if debug logging is not enabled (no-op)
func DebugLog(format string, args ...interface{}) {
	debugMutex.Lock()
	defer debugMutex.Unlock()

	if debugFile == nil {
		return // Debug logging not enabled
	}

	timestamp := time.Now().Format("15:04:05.000")
	message := fmt.Sprintf(format, args...)
	_, _ = fmt.Fprintf(debugFile, "[%s] %s\n", timestamp, message)
	_ = debugFile.Sync() // Flush immediately for real-time debugging
}

// CloseDebugLog closes the debug log file
func CloseDebugLog() {
	debugMutex.Lock()
	defer debugMutex.Unlock()

	if debugFile != nil {
		DebugLog("=== Lotus Debug Log Closed ===")
		_ = debugFile.Close()
		debugFile = nil
	}
}

// GetDebugLogPath returns the path to the debug log file
func GetDebugLogPath() string {
	return debugPath
}
