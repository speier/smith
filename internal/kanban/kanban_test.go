package kanban

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	// Create a temporary kanban.md file
	tmpDir := t.TempDir()
	kanbanPath := filepath.Join(tmpDir, "kanban.md")

	content := `# Agent Kanban Board

## Backlog
<!-- Tasks waiting to be picked up -->

- [ ] task-001: Implement feature A
- [ ] task-002: Fix bug B

## WIP
<!-- Work in progress - tasks currently being worked on -->

- [x] task-003: Working on feature C
- [ ] task-004: Refactoring module D

## Review
<!-- Tasks pending review -->

- [x] task-005: Code review needed

## Done
<!-- Completed tasks -->

- [x] task-006: Initial setup
- [x] task-007: Documentation written
`

	err := os.WriteFile(kanbanPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test kanban file: %v", err)
	}

	// Parse the file
	board, err := Parse(kanbanPath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify Backlog
	if len(board.Backlog) != 2 {
		t.Errorf("Expected 2 backlog tasks, got %d", len(board.Backlog))
	}
	if board.Backlog[0].ID != "task-001" {
		t.Errorf("Expected ID 'task-001', got '%s'", board.Backlog[0].ID)
	}
	if board.Backlog[0].Title != "Implement feature A" {
		t.Errorf("Expected title 'Implement feature A', got '%s'", board.Backlog[0].Title)
	}
	if board.Backlog[0].Checked {
		t.Error("Expected unchecked task")
	}
	if board.Backlog[0].Status != "backlog" {
		t.Errorf("Expected status 'backlog', got '%s'", board.Backlog[0].Status)
	}

	// Verify WIP
	if len(board.WIP) != 2 {
		t.Errorf("Expected 2 WIP tasks, got %d", len(board.WIP))
	}
	if !board.WIP[0].Checked {
		t.Error("Expected checked task")
	}
	if board.WIP[0].ID != "task-003" {
		t.Errorf("Expected ID 'task-003', got '%s'", board.WIP[0].ID)
	}

	// Verify Review
	if len(board.Review) != 1 {
		t.Errorf("Expected 1 review task, got %d", len(board.Review))
	}

	// Verify Done
	if len(board.Done) != 2 {
		t.Errorf("Expected 2 done tasks, got %d", len(board.Done))
	}

	// Test AllTasks
	allTasks := board.AllTasks()
	if len(allTasks) != 7 {
		t.Errorf("Expected 7 total tasks, got %d", len(allTasks))
	}
}

func TestParseWithoutIDs(t *testing.T) {
	// Create a kanban with tasks without IDs
	tmpDir := t.TempDir()
	kanbanPath := filepath.Join(tmpDir, "kanban.md")

	content := `# Agent Kanban Board

## Backlog

- [ ] Implement feature without ID
- [ ] Another task

## WIP

## Review

## Done
`

	err := os.WriteFile(kanbanPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test kanban file: %v", err)
	}

	board, err := Parse(kanbanPath)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(board.Backlog) != 2 {
		t.Errorf("Expected 2 backlog tasks, got %d", len(board.Backlog))
	}

	// Should have empty ID but valid title
	if board.Backlog[0].ID != "" {
		t.Errorf("Expected empty ID, got '%s'", board.Backlog[0].ID)
	}
	if board.Backlog[0].Title != "Implement feature without ID" {
		t.Errorf("Expected full text as title, got '%s'", board.Backlog[0].Title)
	}
}

func TestWriteToFile(t *testing.T) {
	// Create a board
	board := &Board{
		Backlog: []Task{
			{ID: "task-001", Title: "First task", Status: "backlog", Checked: false},
			{ID: "task-002", Title: "Second task", Status: "backlog", Checked: false},
		},
		WIP: []Task{
			{ID: "task-003", Title: "Working on this", Status: "wip", Checked: true},
		},
		Review: []Task{},
		Done: []Task{
			{ID: "task-004", Title: "Completed", Status: "done", Checked: true},
		},
	}

	// Write to file
	tmpDir := t.TempDir()
	kanbanPath := filepath.Join(tmpDir, "kanban.md")

	err := board.WriteToFile(kanbanPath)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	// Read back and parse
	parsed, err := Parse(kanbanPath)
	if err != nil {
		t.Fatalf("Parse after write failed: %v", err)
	}

	// Verify it matches
	if len(parsed.Backlog) != 2 {
		t.Errorf("Expected 2 backlog tasks after roundtrip, got %d", len(parsed.Backlog))
	}
	if len(parsed.WIP) != 1 {
		t.Errorf("Expected 1 WIP task after roundtrip, got %d", len(parsed.WIP))
	}
	if len(parsed.Done) != 1 {
		t.Errorf("Expected 1 done task after roundtrip, got %d", len(parsed.Done))
	}

	if parsed.Backlog[0].ID != "task-001" {
		t.Errorf("Expected ID preserved, got '%s'", parsed.Backlog[0].ID)
	}
	if parsed.WIP[0].Checked != true {
		t.Error("Expected checked status preserved")
	}
}

func TestParseNonexistentFile(t *testing.T) {
	_, err := Parse("/nonexistent/path/kanban.md")
	if err == nil {
		t.Error("Expected error when parsing nonexistent file")
	}
}
