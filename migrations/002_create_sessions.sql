-- Migration 002: Create sessions table
-- A session represents a continuous execution interval for a single agent.
-- A new session row is inserted each time an agent transitions to "running"
-- and its ended_at / duration are populated when the agent stops.

CREATE TABLE IF NOT EXISTS sessions (
    id         INTEGER  PRIMARY KEY AUTOINCREMENT,
    agent_id   INTEGER  NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    started_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    ended_at   DATETIME,              -- NULL means the session is still open.
    duration   INTEGER  NOT NULL DEFAULT 0  -- seconds
);

CREATE INDEX IF NOT EXISTS idx_sessions_agent_id ON sessions (agent_id);
