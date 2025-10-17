package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.etcd.io/bbolt"
)

// BoltStore implements the Store interface using BBolt
type BoltStore struct {
	db *bbolt.DB
}

// NewBoltStore creates a BBolt-backed store
func NewBoltStore(db *bbolt.DB) *BoltStore {
	return &BoltStore{db: db}
}

// === EventStore Implementation ===

func (s *BoltStore) SaveEvent(ctx context.Context, event *Event) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(EventsBucket)
		if b == nil {
			return fmt.Errorf("events bucket not found")
		}

		// Get next sequence ID
		seq := tx.Bucket(SequenceBucket)
		if seq == nil {
			return fmt.Errorf("sequence bucket not found")
		}

		id, err := seq.NextSequence()
		if err != nil {
			return fmt.Errorf("failed to get next sequence: %w", err)
		}
		event.ID = int64(id)

		// Set timestamp if not set
		if event.Timestamp.IsZero() {
			event.Timestamp = time.Now()
		}

		// Encode event
		data, err := json.Marshal(event)
		if err != nil {
			return fmt.Errorf("failed to encode event: %w", err)
		}

		// Store by ID
		key := []byte(fmt.Sprintf("%d", event.ID))
		if err := b.Put(key, data); err != nil {
			return fmt.Errorf("failed to store event: %w", err)
		}

		return nil
	})
}

func (s *BoltStore) QueryEvents(ctx context.Context, filter EventFilter) ([]*Event, error) {
	var events []*Event

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(EventsBucket)
		if b == nil {
			return fmt.Errorf("events bucket not found")
		}

		// Iterate all events (BBolt doesn't have SQL-like WHERE, so we filter in Go)
		c := b.Cursor()
		for k, v := c.Last(); k != nil && len(events) < 1000; k, v = c.Prev() {
			var event Event
			if err := json.Unmarshal(v, &event); err != nil {
				return fmt.Errorf("failed to decode event: %w", err)
			}

			// Apply filters
			if len(filter.EventTypes) > 0 {
				found := false
				for _, et := range filter.EventTypes {
					if event.EventType == et {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			if filter.AgentID != nil && event.AgentID != *filter.AgentID {
				continue
			}

			if filter.TaskID != nil {
				if event.TaskID == nil || *event.TaskID != *filter.TaskID {
					continue
				}
			}

			events = append(events, &event)
		}

		return nil
	})

	return events, err
}

// === AgentStore Implementation ===

func (s *BoltStore) RegisterAgent(ctx context.Context, agent *Agent) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(AgentsBucket)
		if b == nil {
			return fmt.Errorf("agents bucket not found")
		}

		// Set timestamps if needed
		if agent.StartedAt.IsZero() {
			agent.StartedAt = time.Now()
		}
		if agent.LastHeartbeat.IsZero() {
			agent.LastHeartbeat = time.Now()
		}

		data, err := json.Marshal(agent)
		if err != nil {
			return fmt.Errorf("failed to encode agent: %w", err)
		}

		if err := b.Put([]byte(agent.ID), data); err != nil {
			return fmt.Errorf("failed to store agent: %w", err)
		}

		return nil
	})
}

func (s *BoltStore) UpdateHeartbeat(ctx context.Context, agentID string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(AgentsBucket)
		if b == nil {
			return fmt.Errorf("agents bucket not found")
		}

		data := b.Get([]byte(agentID))
		if data == nil {
			return fmt.Errorf("agent not found: %s", agentID)
		}

		var agent Agent
		if err := json.Unmarshal(data, &agent); err != nil {
			return fmt.Errorf("failed to decode agent: %w", err)
		}

		agent.LastHeartbeat = time.Now()

		newData, err := json.Marshal(agent)
		if err != nil {
			return fmt.Errorf("failed to encode agent: %w", err)
		}

		if err := b.Put([]byte(agentID), newData); err != nil {
			return fmt.Errorf("failed to update agent: %w", err)
		}

		return nil
	})
}

func (s *BoltStore) UnregisterAgent(ctx context.Context, agentID string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(AgentsBucket)
		if b == nil {
			return fmt.Errorf("agents bucket not found")
		}

		if err := b.Delete([]byte(agentID)); err != nil {
			return fmt.Errorf("failed to delete agent: %w", err)
		}

		return nil
	})
}

func (s *BoltStore) GetAgent(ctx context.Context, agentID string) (*Agent, error) {
	var agent *Agent

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(AgentsBucket)
		if b == nil {
			return fmt.Errorf("agents bucket not found")
		}

		data := b.Get([]byte(agentID))
		if data == nil {
			return fmt.Errorf("agent not found: %s", agentID)
		}

		agent = &Agent{}
		if err := json.Unmarshal(data, agent); err != nil {
			return fmt.Errorf("failed to decode agent: %w", err)
		}

		return nil
	})

	return agent, err
}

func (s *BoltStore) ListAgents(ctx context.Context, role *string) ([]*Agent, error) {
	var agents []*Agent

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(AgentsBucket)
		if b == nil {
			return fmt.Errorf("agents bucket not found")
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var agent Agent
			if err := json.Unmarshal(v, &agent); err != nil {
				return fmt.Errorf("failed to decode agent: %w", err)
			}

			// Filter by role if specified
			if role != nil && agent.Role != *role {
				continue
			}

			agents = append(agents, &agent)
		}

		return nil
	})

	return agents, err
}

func (s *BoltStore) MarkAgentDead(ctx context.Context, timeout time.Duration) (int, error) {
	count := 0
	deadline := time.Now().Add(-timeout)

	err := s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(AgentsBucket)
		if b == nil {
			return fmt.Errorf("agents bucket not found")
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var agent Agent
			if err := json.Unmarshal(v, &agent); err != nil {
				return fmt.Errorf("failed to decode agent: %w", err)
			}

			if agent.Status == "active" && agent.LastHeartbeat.Before(deadline) {
				agent.Status = "dead"
				data, err := json.Marshal(agent)
				if err != nil {
					return fmt.Errorf("failed to encode agent: %w", err)
				}
				if err := b.Put(k, data); err != nil {
					return fmt.Errorf("failed to update agent: %w", err)
				}
				count++
			}
		}

		return nil
	})

	return count, err
}

// === TaskStore Implementation ===

func (s *BoltStore) CreateTask(ctx context.Context, task *Task) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(TasksBucket)
		if b == nil {
			return fmt.Errorf("tasks bucket not found")
		}

		// Set timestamps if needed
		if task.StartedAt.IsZero() {
			task.StartedAt = time.Now()
		}
		if task.UpdatedAt.IsZero() {
			task.UpdatedAt = time.Now()
		}

		data, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("failed to encode task: %w", err)
		}

		if err := b.Put([]byte(task.TaskID), data); err != nil {
			return fmt.Errorf("failed to store task: %w", err)
		}

		return nil
	})
}

func (s *BoltStore) GetTask(ctx context.Context, taskID string) (*Task, error) {
	var task *Task

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(TasksBucket)
		if b == nil {
			return fmt.Errorf("tasks bucket not found")
		}

		data := b.Get([]byte(taskID))
		if data == nil {
			return fmt.Errorf("task not found: %s", taskID)
		}

		task = &Task{}
		if err := json.Unmarshal(data, task); err != nil {
			return fmt.Errorf("failed to decode task: %w", err)
		}

		return nil
	})

	return task, err
}

func (s *BoltStore) UpdateTask(ctx context.Context, task *Task) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(TasksBucket)
		if b == nil {
			return fmt.Errorf("tasks bucket not found")
		}

		task.UpdatedAt = time.Now()

		data, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("failed to encode task: %w", err)
		}

		if err := b.Put([]byte(task.TaskID), data); err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}

		return nil
	})
}

func (s *BoltStore) ListTasks(ctx context.Context, status *string) ([]*Task, error) {
	var tasks []*Task

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(TasksBucket)
		if b == nil {
			return fmt.Errorf("tasks bucket not found")
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var task Task
			if err := json.Unmarshal(v, &task); err != nil {
				return fmt.Errorf("failed to decode task: %w", err)
			}

			// Filter by status if specified
			if status != nil && task.Status != *status {
				continue
			}

			tasks = append(tasks, &task)
		}

		return nil
	})

	return tasks, err
}

func (s *BoltStore) ClaimTask(ctx context.Context, taskID, agentID string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(TasksBucket)
		if b == nil {
			return fmt.Errorf("tasks bucket not found")
		}

		data := b.Get([]byte(taskID))
		if data == nil {
			return fmt.Errorf("task not found: %s", taskID)
		}

		var task Task
		if err := json.Unmarshal(data, &task); err != nil {
			return fmt.Errorf("failed to decode task: %w", err)
		}

		// Only claim if in backlog state
		if task.Status != "backlog" {
			return fmt.Errorf("task %s is not claimable", taskID)
		}

		task.AgentID = agentID
		task.Status = "wip"
		task.UpdatedAt = time.Now()

		newData, err := json.Marshal(task)
		if err != nil {
			return fmt.Errorf("failed to encode task: %w", err)
		}

		if err := b.Put([]byte(taskID), newData); err != nil {
			return fmt.Errorf("failed to update task: %w", err)
		}

		return nil
	})
}

func (s *BoltStore) GetTaskStats(ctx context.Context) (*TaskStats, error) {
	stats := &TaskStats{}

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(TasksBucket)
		if b == nil {
			return fmt.Errorf("tasks bucket not found")
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var task Task
			if err := json.Unmarshal(v, &task); err != nil {
				return fmt.Errorf("failed to decode task: %w", err)
			}

			switch task.Status {
			case "backlog":
				stats.Backlog++
			case "wip":
				stats.WIP++
			case "review":
				stats.Review++
			case "done":
				stats.Done++
			}
		}

		return nil
	})

	return stats, err
}

// === LockStore Implementation ===

func (s *BoltStore) AcquireLocks(ctx context.Context, locks []*FileLock) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(FileLocksBucket)
		if b == nil {
			return fmt.Errorf("file_locks bucket not found")
		}

		// Check for existing locks first
		for _, lock := range locks {
			existing := b.Get([]byte(lock.FilePath))
			if existing != nil {
				return fmt.Errorf("file %s is already locked", lock.FilePath)
			}
		}

		// Acquire all locks atomically
		for _, lock := range locks {
			if lock.LockedAt.IsZero() {
				lock.LockedAt = time.Now()
			}

			data, err := json.Marshal(lock)
			if err != nil {
				return fmt.Errorf("failed to encode lock: %w", err)
			}

			if err := b.Put([]byte(lock.FilePath), data); err != nil {
				return fmt.Errorf("failed to store lock: %w", err)
			}
		}

		return nil
	})
}

func (s *BoltStore) ReleaseLocks(ctx context.Context, agentID string, files []string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(FileLocksBucket)
		if b == nil {
			return fmt.Errorf("file_locks bucket not found")
		}

		if len(files) == 0 {
			// Release all locks for this agent
			c := b.Cursor()
			var toDelete [][]byte
			for k, v := c.First(); k != nil; k, v = c.Next() {
				var lock FileLock
				if err := json.Unmarshal(v, &lock); err != nil {
					return fmt.Errorf("failed to decode lock: %w", err)
				}
				if lock.AgentID == agentID {
					toDelete = append(toDelete, append([]byte(nil), k...)) // Copy key
				}
			}
			for _, key := range toDelete {
				if err := b.Delete(key); err != nil {
					return fmt.Errorf("failed to delete lock: %w", err)
				}
			}
		} else {
			// Release specific files
			for _, file := range files {
				data := b.Get([]byte(file))
				if data != nil {
					var lock FileLock
					if err := json.Unmarshal(data, &lock); err != nil {
						return fmt.Errorf("failed to decode lock: %w", err)
					}
					// Only delete if owned by this agent
					if lock.AgentID == agentID {
						if err := b.Delete([]byte(file)); err != nil {
							return fmt.Errorf("failed to delete lock: %w", err)
						}
					}
				}
			}
		}

		return nil
	})
}

func (s *BoltStore) GetLocks(ctx context.Context) ([]*FileLock, error) {
	var locks []*FileLock

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(FileLocksBucket)
		if b == nil {
			return fmt.Errorf("file_locks bucket not found")
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var lock FileLock
			if err := json.Unmarshal(v, &lock); err != nil {
				return fmt.Errorf("failed to decode lock: %w", err)
			}
			locks = append(locks, &lock)
		}

		return nil
	})

	return locks, err
}

func (s *BoltStore) GetLocksForAgent(ctx context.Context, agentID string) ([]*FileLock, error) {
	var locks []*FileLock

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(FileLocksBucket)
		if b == nil {
			return fmt.Errorf("file_locks bucket not found")
		}

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var lock FileLock
			if err := json.Unmarshal(v, &lock); err != nil {
				return fmt.Errorf("failed to decode lock: %w", err)
			}
			if lock.AgentID == agentID {
				locks = append(locks, &lock)
			}
		}

		return nil
	})

	return locks, err
}

// === SessionStore Implementation ===

func (s *BoltStore) CreateSession(ctx context.Context, session *Session) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(SessionsBucket)
		if b == nil {
			return fmt.Errorf("sessions bucket not found")
		}

		// Encode session
		data, err := json.Marshal(session)
		if err != nil {
			return fmt.Errorf("failed to encode session: %w", err)
		}

		// Store session
		return b.Put([]byte(session.SessionID), data)
	})
}

func (s *BoltStore) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	var session *Session

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(SessionsBucket)
		if b == nil {
			return fmt.Errorf("sessions bucket not found")
		}

		data := b.Get([]byte(sessionID))
		if data == nil {
			return fmt.Errorf("session not found: %s", sessionID)
		}

		session = &Session{}
		return json.Unmarshal(data, session)
	})

	return session, err
}

func (s *BoltStore) UpdateSession(ctx context.Context, session *Session) error {
	return s.CreateSession(ctx, session) // Same as create - upsert
}

func (s *BoltStore) ListSessions(ctx context.Context, limit int) ([]*Session, error) {
	var sessions []*Session

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(SessionsBucket)
		if b == nil {
			return fmt.Errorf("sessions bucket not found")
		}

		// Collect all sessions
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var session Session
			if err := json.Unmarshal(v, &session); err != nil {
				continue // Skip corrupted entries
			}
			sessions = append(sessions, &session)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by LastActive (most recent first)
	// Simple bubble sort for small lists
	for i := 0; i < len(sessions)-1; i++ {
		for j := i + 1; j < len(sessions); j++ {
			if sessions[j].LastActive.After(sessions[i].LastActive) {
				sessions[i], sessions[j] = sessions[j], sessions[i]
			}
		}
	}

	// Apply limit
	if limit > 0 && len(sessions) > limit {
		sessions = sessions[:limit]
	}

	return sessions, nil
}

func (s *BoltStore) ArchiveSession(ctx context.Context, sessionID string) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(SessionsBucket)
		if b == nil {
			return fmt.Errorf("sessions bucket not found")
		}

		// Get existing session
		data := b.Get([]byte(sessionID))
		if data == nil {
			return fmt.Errorf("session not found: %s", sessionID)
		}

		var session Session
		if err := json.Unmarshal(data, &session); err != nil {
			return fmt.Errorf("failed to decode session: %w", err)
		}

		// Update status
		session.Status = "archived"

		// Re-encode and save
		updatedData, err := json.Marshal(session)
		if err != nil {
			return fmt.Errorf("failed to encode session: %w", err)
		}

		return b.Put([]byte(sessionID), updatedData)
	})
}

func (s *BoltStore) GetSessionTasks(ctx context.Context, sessionID string) ([]*Task, error) {
	var tasks []*Task

	err := s.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(TasksBucket)
		if b == nil {
			return fmt.Errorf("tasks bucket not found")
		}

		// Iterate through all tasks
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var task Task
			if err := json.Unmarshal(v, &task); err != nil {
				continue // Skip corrupted entries
			}

			// Filter by session
			if task.SessionID == sessionID {
				tasks = append(tasks, &task)
			}
		}

		return nil
	})

	return tasks, err
}

// Close closes the database
func (s *BoltStore) Close() error {
	return s.db.Close()
}

// Helper function to convert string to int64
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}
