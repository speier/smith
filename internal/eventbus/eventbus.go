package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/speier/smith/internal/storage"
)

// EventBus handles publishing and subscribing to events using storage abstraction
type EventBus struct {
	store storage.EventStore
}

// New creates a new EventBus instance
func New(store storage.EventStore) *EventBus {
	return &EventBus{store: store}
}

// Publish publishes an event to the event bus
func (eb *EventBus) Publish(ctx context.Context, event *Event) error {
	// Convert eventbus.Event to storage.Event
	storageEvent := &storage.Event{
		AgentID:   event.AgentID,
		AgentRole: string(event.AgentRole),
		EventType: string(event.Type),
		TaskID:    event.TaskID,
		FilePath:  event.FilePath,
		Data:      event.Data,
	}

	if err := eb.store.SaveEvent(ctx, storageEvent); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	// Copy back the generated ID
	event.ID = storageEvent.ID
	event.Timestamp = storageEvent.Timestamp

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
	// Convert eventbus filter to storage filter
	var eventTypes []string
	for _, et := range filter.EventTypes {
		eventTypes = append(eventTypes, string(et))
	}

	storageFilter := storage.EventFilter{
		AgentID:    filter.AgentID,
		EventTypes: eventTypes,
		TaskID:     filter.TaskID,
	}

	storageEvents, err := eb.store.QueryEvents(ctx, storageFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}

	// Convert storage.Event to eventbus.Event and apply additional filters
	var events []Event
	for _, se := range storageEvents {
		// Apply SinceID filter (BBolt doesn't support this natively)
		if filter.SinceID > 0 && se.ID <= filter.SinceID {
			continue
		}

		// Apply AgentRole filter if specified
		if filter.AgentRole != nil && se.AgentRole != string(*filter.AgentRole) {
			continue
		}

		// Apply FilePath filter if specified
		if filter.FilePath != nil {
			if se.FilePath == nil || *se.FilePath != *filter.FilePath {
				continue
			}
		}

		event := Event{
			ID:        se.ID,
			Timestamp: se.Timestamp,
			AgentID:   se.AgentID,
			AgentRole: AgentRole(se.AgentRole),
			Type:      EventType(se.EventType),
			TaskID:    se.TaskID,
			FilePath:  se.FilePath,
			Data:      se.Data,
		}

		events = append(events, event)
	}

	return events, nil
}

// GetLatestEventID returns the ID of the latest event
func (eb *EventBus) GetLatestEventID(ctx context.Context) (int64, error) {
	// Query all events and get the max ID (storage interface doesn't provide MAX)
	events, err := eb.store.QueryEvents(ctx, storage.EventFilter{})
	if err != nil {
		return 0, fmt.Errorf("failed to query events: %w", err)
	}

	var maxID int64
	for _, event := range events {
		if event.ID > maxID {
			maxID = event.ID
		}
	}

	return maxID, nil
}
