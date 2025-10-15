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
	GetTasksByStatus(status string) ([]Task, error)
	GetActiveLocks() ([]Lock, error)
	GetMessages() ([]Message, error)

	// Task lifecycle management
	CreateTask(title, description, role string) (taskID string, err error)
	ClaimTask(taskID, agent string) error
	UpdateTaskStatus(taskID, status string) error
	CompleteTask(taskID, result string) error
	FailTask(taskID, errorMsg string) error
	GetTask(taskID string) (*Task, error)

	// File coordination
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
