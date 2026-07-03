-- Migration 004: Create schema_migrations tracking table
-- Tracks which migration files have been applied so the runner is idempotent.

CREATE TABLE IF NOT EXISTS schema_migrations (
    version    TEXT     PRIMARY KEY,
    applied_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
