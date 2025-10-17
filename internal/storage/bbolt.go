package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/speier/smith/internal/scaffold"
	"go.etcd.io/bbolt"
)

// BoltDB wraps the BBolt database connection
type BoltDB struct {
	*bbolt.DB
	path string
}

// Bucket names
var (
	EventsBucket    = []byte("events")
	FileLocksBucket = []byte("file_locks")
	TasksBucket     = []byte("tasks")
	AgentsBucket    = []byte("agents")
	SessionsBucket  = []byte("sessions")
	SequenceBucket  = []byte("sequences")
)

// Note: Task, Agent, Event, FileLock types are now defined in interfaces.go

// InitProjectStorage initializes the .smith directory and BBolt database.
// This is the primary storage backend using BBolt for lock-free concurrent access.
func InitProjectStorage(projectRoot string) (Store, error) {
	smithDir := filepath.Join(projectRoot, ".smith")

	// Create .smith/ directory if it doesn't exist
	if err := os.MkdirAll(smithDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .smith directory: %w", err)
	}

	// Initialize database
	dbPath := filepath.Join(smithDir, "smith.db")
	db, err := initBoltDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Create default project files (config, .gitignore)
	if err := scaffold.InitProjectFiles(smithDir); err != nil {
		return nil, err
	}

	// Return wrapped Store interface
	return NewBoltStore(db.DB), nil
}

// initBoltDatabase creates and initializes the BBolt database with buckets
func initBoltDatabase(dbPath string) (*BoltDB, error) {
	// Open database
	boltDB, err := bbolt.Open(dbPath, 0600, &bbolt.Options{
		Timeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create buckets
	err = boltDB.Update(func(tx *bbolt.Tx) error {
		buckets := [][]byte{
			EventsBucket,
			FileLocksBucket,
			TasksBucket,
			AgentsBucket,
			SessionsBucket,
			SequenceBucket,
		}
		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists(bucket); err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
		}
		return nil
	})
	if err != nil {
		boltDB.Close()
		return nil, err
	}

	return &BoltDB{
		DB:   boltDB,
		path: dbPath,
	}, nil
}

// Path returns the file path of the database
func (db *BoltDB) Path() string {
	return db.path
}

// Helper function to get next sequence number
func (db *BoltDB) nextSequence(tx *bbolt.Tx, key string) (uint64, error) {
	b := tx.Bucket(SequenceBucket)
	seq, _ := b.NextSequence()
	return seq, nil
}

// Helper functions for JSON encoding/decoding
func encodeJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func decodeJSON(data []byte, v interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("empty data")
	}
	return json.Unmarshal(data, v)
}
