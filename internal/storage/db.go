package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/speier/smith/internal/scaffold"
	_ "modernc.org/sqlite" // SQLite driver
)

// DB wraps the SQLite database connection
type DB struct {
	*sql.DB
	path string
}

// InitProjectStorage initializes the .smith directory and SQLite database for a project.
// Creates the directory structure, database, and default files if they don't exist.
func InitProjectStorage(projectRoot string) (*DB, error) {
	smithDir := filepath.Join(projectRoot, ".smith")

	// Create .smith/ directory if it doesn't exist
	if err := os.MkdirAll(smithDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .smith directory: %w", err)
	}

	// Initialize database
	dbPath := filepath.Join(smithDir, "smith.db")
	db, err := initDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create default project files (kanban, config, .gitignore)
	if err := scaffold.InitProjectFiles(smithDir); err != nil {
		return nil, err
	}

	return db, nil
}

// initDatabase creates and initializes the SQLite database with schema
func initDatabase(dbPath string) (*DB, error) {
	// Open database connection
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25) // Allow multiple concurrent agents
	sqlDB.SetMaxIdleConns(5)

	// Enable WAL mode for better concurrency
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Enable foreign keys
	if _, err := sqlDB.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	db := &DB{
		DB:   sqlDB,
		path: dbPath,
	}

	// Run migrations
	if err := db.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// runMigrations applies the database schema
func (db *DB) runMigrations() error {
	// Apply main schema
	if _, err := db.Exec(Schema); err != nil {
		return fmt.Errorf("failed to apply schema: %w", err)
	}

	// Apply indexes
	if _, err := db.Exec(Indexes); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

// Path returns the file path of the database
func (db *DB) Path() string {
	return db.path
}
