package coordinator

import (
	"fmt"
	"log"
)

// Coordinator defines the interface for task and file coordination
type Coordinator interface {
	EnsureDirectories() error
	GetTaskStats() (*TaskStats, error)
	GetAvailableTasks() ([]Task, error)
	GetActiveLocks() ([]Lock, error)
	GetMessages() ([]Message, error)
	ClaimTask(taskID, agent string) error
	LockFiles(taskID, agent string, files []string) error
}

// New creates a new SQLite-based Coordinator instance
func New(projectPath string) Coordinator {
	coord, err := NewSQLite(projectPath)
	if err != nil {
		log.Fatalf("Failed to create coordinator: %v", err)
	}
	return coord
}

// MustNew creates a coordinator or panics
func MustNew(projectPath string) Coordinator {
	coord, err := NewSQLite(projectPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to create coordinator: %v", err))
	}
	return coord
}
