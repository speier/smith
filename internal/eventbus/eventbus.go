package eventbus

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// EventBus handles publishing and subscribing to events via SQLite
type EventBus struct {
	db *sql.DB
}

// New creates a new EventBus instance
func New(db *sql.DB) *EventBus {
	return &EventBus{db: db}
}

// Publish publishes an event to the event bus
func (eb *EventBus) Publish(ctx context.Context, event *Event) error {
	query := `
		INSERT INTO events (agent_id, agent_role, event_type, task_id, file_path, data)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	result, err := eb.db.ExecContext(
		ctx,
		query,
		event.AgentID,
		event.AgentRole,
		event.Type,
		event.TaskID,
		event.FilePath,
		event.Data,
	)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	// Get the inserted event ID
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get event ID: %w", err)
	}
	event.ID = id

	return nil
}

// PublishWithData publishes an event with JSON-encoded data
func (eb *EventBus) PublishWithData(ctx context.Context, agentID string, role AgentRole, eventType EventType, taskID *string, filePath *string, data interface{}) error {
	var jsonData string
	if data != nil {
		bytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal event data: %w", err)
		}
		jsonData = string(bytes)
	}

	event := &Event{
		AgentID:   agentID,
		AgentRole: role,
		Type:      eventType,
		TaskID:    taskID,
		FilePath:  filePath,
		Data:      jsonData,
	}

	return eb.Publish(ctx, event)
}

// Subscribe subscribes to events matching the filter
// Returns a channel that receives events as they are published
func (eb *EventBus) Subscribe(ctx context.Context, filter EventFilter, pollInterval time.Duration) (<-chan Event, error) {
	ch := make(chan Event, 10) // Buffered channel to avoid blocking

	go func() {
		defer close(ch)

		lastID := filter.SinceID
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				events, err := eb.Query(ctx, EventFilter{
					SinceID:    lastID,
					AgentID:    filter.AgentID,
					AgentRole:  filter.AgentRole,
					EventTypes: filter.EventTypes,
					TaskID:     filter.TaskID,
					FilePath:   filter.FilePath,
				})
				if err != nil {
					// Log error but continue polling
					continue
				}

				for _, event := range events {
					select {
					case ch <- event:
						if event.ID > lastID {
							lastID = event.ID
						}
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

// Query queries events matching the filter
func (eb *EventBus) Query(ctx context.Context, filter EventFilter) ([]Event, error) {
	query := `
		SELECT id, timestamp, agent_id, agent_role, event_type, task_id, file_path, data
		FROM events
		WHERE id > ?
	`
	args := []interface{}{filter.SinceID}
	conditions := []string{}

	if filter.AgentID != nil {
		conditions = append(conditions, "agent_id = ?")
		args = append(args, *filter.AgentID)
	}

	if filter.AgentRole != nil {
		conditions = append(conditions, "agent_role = ?")
		args = append(args, *filter.AgentRole)
	}

	if len(filter.EventTypes) > 0 {
		placeholders := make([]string, len(filter.EventTypes))
		for i, et := range filter.EventTypes {
			placeholders[i] = "?"
			args = append(args, et)
		}
		conditions = append(conditions, fmt.Sprintf("event_type IN (%s)", strings.Join(placeholders, ",")))
	}

	if filter.TaskID != nil {
		conditions = append(conditions, "task_id = ?")
		args = append(args, *filter.TaskID)
	}

	if filter.FilePath != nil {
		conditions = append(conditions, "file_path = ?")
		args = append(args, *filter.FilePath)
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY id ASC"

	rows, err := eb.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		var taskID, filePath sql.NullString
		var data sql.NullString

		err := rows.Scan(
			&event.ID,
			&event.Timestamp,
			&event.AgentID,
			&event.AgentRole,
			&event.Type,
			&taskID,
			&filePath,
			&data,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if taskID.Valid {
			event.TaskID = &taskID.String
		}
		if filePath.Valid {
			event.FilePath = &filePath.String
		}
		if data.Valid {
			event.Data = data.String
		}

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating events: %w", err)
	}

	return events, nil
}

// GetLatestEventID returns the ID of the latest event
func (eb *EventBus) GetLatestEventID(ctx context.Context) (int64, error) {
	var id sql.NullInt64
	err := eb.db.QueryRowContext(ctx, "SELECT MAX(id) FROM events").Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest event ID: %w", err)
	}
	if !id.Valid {
		return 0, nil
	}
	return id.Int64, nil
}
