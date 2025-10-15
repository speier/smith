package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitProjectStorage(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "smith-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize storage
	db, err := InitProjectStorage(tmpDir)
	if err != nil {
		t.Fatalf("InitProjectStorage failed: %v", err)
	}
	defer db.Close()

	// Verify .smith directory was created
	smithDir := filepath.Join(tmpDir, ".smith")
	if _, err := os.Stat(smithDir); os.IsNotExist(err) {
		t.Error(".smith directory was not created")
	}

	// Verify smith.db was created
	dbPath := filepath.Join(smithDir, "smith.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("smith.db was not created")
	}

	// Verify config.yaml was created
	configPath := filepath.Join(smithDir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config.yaml was not created")
	}

	// Verify .gitignore was created
	gitignorePath := filepath.Join(smithDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Error(".gitignore was not created")
	}

	// Verify tables were created by querying them
	tables := []string{"events", "file_locks", "task_assignments", "agents"}
	for _, table := range tables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
		if err != nil {
			t.Errorf("table %s does not exist or is invalid: %v", table, err)
		}
	}
}

func TestInitProjectStorageIdempotent(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "smith-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize storage twice
	db1, err := InitProjectStorage(tmpDir)
	if err != nil {
		t.Fatalf("first InitProjectStorage failed: %v", err)
	}
	db1.Close()

	db2, err := InitProjectStorage(tmpDir)
	if err != nil {
		t.Fatalf("second InitProjectStorage failed: %v", err)
	}
	defer db2.Close()

	// Should succeed without errors (idempotent)
}
