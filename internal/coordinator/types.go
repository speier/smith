package coordinator

import "time"

// TaskStats represents statistics about tasks in different states
type TaskStats struct {
	Backlog int
	WIP     int
	Review  int
	Done    int
}

// Task represents a task in the system
type Task struct {
	ID          string
	Title       string
	Description string
	Status      string // backlog, wip, review, done
	Role        string // planning, implementation, testing, review
	AgentID     string
	Result      string   // Output/result from task execution
	Error       string   // Error message if task failed
	Priority    int      // 0=low, 1=medium (default), 2=high
	DependsOn   []string // Task IDs that must be completed first
	StartedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time

	// Agent memory/learnings
	Learnings       string            // What the agent learned
	TriedApproaches []string          // Approaches attempted
	Blockers        []string          // What didn't work
	Notes           map[string]string // Freeform agent notes
}

// Lock represents a file lock held by an agent
type Lock struct {
	Agent  string
	TaskID string
	Files  string
}

// Session represents a work session
type Session struct {
	SessionID  string
	Title      string
	StartedAt  time.Time
	LastActive time.Time
	TaskCount  int
	Status     string
}

// Message represents a message between agents
type Message struct {
	From    string
	To      string
	Subject string
	Body    string
}
