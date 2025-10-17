package storage

import (
	"context"
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

	// Verify tables exist by using Store methods
	ctx := context.Background()

	// Test events table
	_, err = db.QueryEvents(ctx, EventFilter{})
	if err != nil {
		t.Errorf("events table not accessible: %v", err)
	}

	// Test agents table
	_, err = db.ListAgents(ctx, nil)
	if err != nil {
		t.Errorf("agents table not accessible: %v", err)
	}

	// Test tasks table
	_, err = db.ListTasks(ctx, nil)
	if err != nil {
		t.Errorf("tasks table not accessible: %v", err)
	}

	// Test locks table
	_, err = db.GetLocks(ctx)
	if err != nil {
		t.Errorf("locks table not accessible: %v", err)
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
