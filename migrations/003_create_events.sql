-- Migration 003: Create events table
-- Events form an append-only audit log of agent lifecycle state transitions.
-- Rows are never updated or deleted in the MVP.

CREATE TABLE IF NOT EXISTS events (
    id         INTEGER  PRIMARY KEY AUTOINCREMENT,
    agent_id   INTEGER  NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    event_type TEXT     NOT NULL,
    created_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_events_agent_id   ON events (agent_id);
CREATE INDEX IF NOT EXISTS idx_events_event_type ON events (event_type);
CREATE INDEX IF NOT EXISTS idx_events_created_at ON events (created_at);
