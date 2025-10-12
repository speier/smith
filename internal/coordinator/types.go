package coordinator

// TaskStats represents statistics about tasks in different states
type TaskStats struct {
	Backlog int
	WIP     int
	Review  int
	Done    int
}

// Task represents a task in the system
type Task struct {
	ID     string
	Title  string
	Status string
	Role   string
}

// Lock represents a file lock held by an agent
type Lock struct {
	Agent  string
	TaskID string
	Files  string
}

// Message represents a message between agents
type Message struct {
	From    string
	To      string
	Subject string
	Body    string
}
