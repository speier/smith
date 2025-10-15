package storage

// Schema defines the SQLite database schema for the Smith agent coordination system.
const Schema = `
-- Events table: append-only log of all agent activities
CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    agent_id TEXT NOT NULL,
    agent_role TEXT NOT NULL CHECK(agent_role IN ('coordinator', 'planning', 'implementation', 'testing', 'review')),
    event_type TEXT NOT NULL,
    task_id TEXT,
    file_path TEXT,
    data TEXT
);

-- File locks: active locks on files being worked on
CREATE TABLE IF NOT EXISTS file_locks (
    file_path TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    task_id TEXT NOT NULL,
    locked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (agent_id) REFERENCES agents(agent_id) ON DELETE CASCADE
);

-- Task assignments: tracks which agent is working on which task
CREATE TABLE IF NOT EXISTS task_assignments (
    task_id TEXT PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    agent_id TEXT,
    agent_role TEXT CHECK(agent_role IN ('planning', 'implementation', 'testing', 'review', '') OR agent_role IS NULL),
    status TEXT NOT NULL CHECK(status IN ('backlog', 'wip', 'review', 'done')) DEFAULT 'backlog',
    result TEXT,
    error TEXT,
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at DATETIME,
    FOREIGN KEY (agent_id) REFERENCES agents(agent_id) ON DELETE CASCADE
);

-- Agents registry: tracks all active agents with heartbeat
CREATE TABLE IF NOT EXISTS agents (
    agent_id TEXT PRIMARY KEY,
    agent_role TEXT NOT NULL CHECK(agent_role IN ('coordinator', 'planning', 'implementation', 'testing', 'review')),
    status TEXT NOT NULL CHECK(status IN ('active', 'idle', 'dead')) DEFAULT 'active',
    task_id TEXT,
    pid INTEGER,  -- Process ID for tracking
    started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_heartbeat DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Triggers to auto-update timestamps
CREATE TRIGGER IF NOT EXISTS update_task_assignment_timestamp 
AFTER UPDATE ON task_assignments
FOR EACH ROW
BEGIN
    UPDATE task_assignments SET updated_at = CURRENT_TIMESTAMP WHERE task_id = NEW.task_id;
END;

CREATE TRIGGER IF NOT EXISTS update_agent_heartbeat_on_event
AFTER INSERT ON events
FOR EACH ROW
BEGIN
    UPDATE agents SET last_heartbeat = CURRENT_TIMESTAMP WHERE agent_id = NEW.agent_id;
END;
`

// Indexes for performance optimization
const Indexes = `
CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
CREATE INDEX IF NOT EXISTS idx_events_agent_id ON events(agent_id);
CREATE INDEX IF NOT EXISTS idx_events_task_id ON events(task_id);
CREATE INDEX IF NOT EXISTS idx_events_file_path ON events(file_path);
CREATE INDEX IF NOT EXISTS idx_events_agent_role ON events(agent_role);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(event_type);
CREATE INDEX IF NOT EXISTS idx_agents_role ON agents(agent_role);
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents(status);
CREATE INDEX IF NOT EXISTS idx_task_assignments_status ON task_assignments(status);
`
