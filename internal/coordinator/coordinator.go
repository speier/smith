package coordinator

import (
	"log"
	"os"
)

// New creates a new Coordinator instance
// Returns SQLiteCoordinator if SMITH_USE_SQLITE=true, otherwise FileCoordinator
func New(projectPath string) *FileCoordinator {
	// Feature flag for SQLite implementation
	if os.Getenv("SMITH_USE_SQLITE") == "true" {
		coord, err := NewSQLite(projectPath)
		if err != nil {
			log.Printf("Warning: Failed to create SQLite coordinator: %v, falling back to file-based", err)
			return NewFile(projectPath)
		}

		// Return as FileCoordinator type for now (both implement same methods)
		// TODO: Create interface to avoid this
		_ = coord
		log.Println("Note: SQLite coordinator created but not yet fully integrated")
		// For now, still return FileCoordinator
		return NewFile(projectPath)
	}

	// Default to file-based coordinator
	return NewFile(projectPath)
}
