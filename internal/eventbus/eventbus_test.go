package eventbus

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/speier/smith/internal/storage"
)

func setupTestDB(t *testing.T) (storage.Store, func()) {
	tmpDir, err := os.MkdirTemp("", "smith-eventbus-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	store, err := storage.InitProjectStorage(tmpDir)
	if err != nil {
		t.Fatalf("failed to init storage: %v", err)
	}

	cleanup := func() {
		_ = store.Close()
		_ = os.RemoveAll(tmpDir)
	}

	return store, cleanup
}

func TestPublishAndQuery(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	bus := New(store)
	ctx := context.Background()

	// Publish an event
	taskID := "task-123"
	event := &Event{
		AgentID:   "agent-1",
		AgentRole: RolePlanning,
		Type:      EventTaskStarted,
		TaskID:    &taskID,
	}

	err := bus.Publish(ctx, event)
	if err != nil {
		t.Fatalf("failed to publish event: %v", err)
	}

	if event.ID == 0 {
		t.Error("event ID should be set after publish")
	}

	// Query the event
	events, err := bus.Query(ctx, EventFilter{SinceID: 0})
	if err != nil {
		t.Fatalf("failed to query events: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].AgentID != "agent-1" {
		t.Errorf("expected agent-1, got %s", events[0].AgentID)
	}
}

func TestPublishWithData(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	bus := New(store)
	ctx := context.Background()

	taskID := "task-456"
	filePath := "/path/to/file.go"
	data := map[string]interface{}{
		"message": "locked file for task",
		"timeout": 300,
	}

	err := bus.PublishWithData(ctx, "agent-2", RoleImplementation, EventFileLocked, &taskID, &filePath, data)
	if err != nil {
		t.Fatalf("failed to publish with data: %v", err)
	}

	// Query and verify
	events, err := bus.Query(ctx, EventFilter{SinceID: 0})
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].Data == "" {
		t.Error("event data should not be empty")
	}
}

func TestEventFiltering(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	bus := New(store)
	ctx := context.Background()

	// Publish multiple events
	task1 := "task-1"
	task2 := "task-2"

	events := []Event{
		{AgentID: "agent-1", AgentRole: RolePlanning, Type: EventTaskStarted, TaskID: &task1},
		{AgentID: "agent-2", AgentRole: RoleImplementation, Type: EventTaskStarted, TaskID: &task2},
		{AgentID: "agent-1", AgentRole: RolePlanning, Type: EventTaskCompleted, TaskID: &task1},
	}

	for i := range events {
		if err := bus.Publish(ctx, &events[i]); err != nil {
			t.Fatalf("failed to publish event %d: %v", i, err)
		}
	}

	// Filter by agent
	agentID := "agent-1"
	filtered, err := bus.Query(ctx, EventFilter{
		SinceID: 0,
		AgentID: &agentID,
	})
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}

	if len(filtered) != 2 {
		t.Errorf("expected 2 events for agent-1, got %d", len(filtered))
	}

	// Filter by task
	filtered, err = bus.Query(ctx, EventFilter{
		SinceID: 0,
		TaskID:  &task2,
	})
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("expected 1 event for task-2, got %d", len(filtered))
	}
}

func TestSubscribe(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	bus := New(store)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Subscribe to events
	eventChan, err := bus.Subscribe(ctx, EventFilter{SinceID: 0}, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	// Publish events in background
	go func() {
		time.Sleep(200 * time.Millisecond)
		taskID := "task-async"
		_ = bus.Publish(context.Background(), &Event{
			AgentID:   "agent-async",
			AgentRole: RoleTesting,
			Type:      EventTaskStarted,
			TaskID:    &taskID,
		})
	}()

	// Wait for event
	select {
	case event := <-eventChan:
		if event.AgentID != "agent-async" {
			t.Errorf("expected agent-async, got %s", event.AgentID)
		}
	case <-ctx.Done():
		t.Error("timeout waiting for event")
	}
}
