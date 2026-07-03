-- Migration 001: Create agents table
-- Records every AI agent discovered on the host machine.

CREATE TABLE IF NOT EXISTS agents (
    id            INTEGER  PRIMARY KEY AUTOINCREMENT,
    pid           INTEGER  NOT NULL,
    name          TEXT     NOT NULL,
    executable_path TEXT   NOT NULL DEFAULT '',
    command       TEXT     NOT NULL DEFAULT '',
    parent_pid    INTEGER  NOT NULL DEFAULT 0,
    status        TEXT     NOT NULL DEFAULT 'running',
    discovered_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- Index on PID for fast lookups during scan reconciliation.
CREATE INDEX IF NOT EXISTS idx_agents_pid    ON agents (pid);
-- Index on name to quickly filter by agent type.
CREATE INDEX IF NOT EXISTS idx_agents_name   ON agents (name);
-- Index on status to list only running agents efficiently.
CREATE INDEX IF NOT EXISTS idx_agents_status ON agents (status);
