// Package storage — EventRepository
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/chethanyadav456/agentnetra/pkg/models"
)

// EventRepository performs insert and query operations against the events table.
// Events are append-only; there are no update or delete operations.
type EventRepository struct {
	db *sql.DB
}

// NewEventRepository constructs an EventRepository backed by the provided DB.
func NewEventRepository(db *DB) *EventRepository {
	return &EventRepository{db: db.Conn()}
}

// Create inserts a new event record and returns it with the assigned ID.
func (r *EventRepository) Create(ctx context.Context, event models.Event) (models.Event, error) {
	const q = `
		INSERT INTO events (agent_id, event_type, created_at)
		VALUES (?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, q,
		event.AgentID,
		event.EventType,
		event.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return models.Event{}, fmt.Errorf("event repo: create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return models.Event{}, fmt.Errorf("event repo: get last insert id: %w", err)
	}

	event.ID = id
	return event, nil
}

// ListByAgent returns all events for the given agent ID, newest first.
func (r *EventRepository) ListByAgent(ctx context.Context, agentID int64) ([]models.Event, error) {
	const q = `
		SELECT id, agent_id, event_type, created_at
		FROM events WHERE agent_id = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, q, agentID)
	if err != nil {
		return nil, fmt.Errorf("event repo: list by agent: %w", err)
	}
	defer rows.Close()
	return scanEvents(rows)
}

// ListAll returns all events ordered by creation time descending.
func (r *EventRepository) ListAll(ctx context.Context) ([]models.Event, error) {
	const q = `
		SELECT id, agent_id, event_type, created_at
		FROM events ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("event repo: list all: %w", err)
	}
	defer rows.Close()
	return scanEvents(rows)
}

// ListByType returns all events of the specified event_type, newest first.
func (r *EventRepository) ListByType(ctx context.Context, eventType string) ([]models.Event, error) {
	const q = `
		SELECT id, agent_id, event_type, created_at
		FROM events WHERE event_type = ?
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, q, eventType)
	if err != nil {
		return nil, fmt.Errorf("event repo: list by type: %w", err)
	}
	defer rows.Close()
	return scanEvents(rows)
}

// scanEvents scans *sql.Rows into a slice of Event.
func scanEvents(rows *sql.Rows) ([]models.Event, error) {
	var events []models.Event
	for rows.Next() {
		var e models.Event
		var createdAt string
		if err := rows.Scan(&e.ID, &e.AgentID, &e.EventType, &createdAt); err != nil {
			return nil, fmt.Errorf("event repo: scan: %w", err)
		}
		e.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("event repo: rows error: %w", err)
	}
	return events, nil
}

// CountByAgent returns the total number of events recorded for the given agent.
func (r *EventRepository) CountByAgent(ctx context.Context, agentID int64) (int64, error) {
	const q = `SELECT COUNT(*) FROM events WHERE agent_id = ?`
	var count int64
	if err := r.db.QueryRowContext(ctx, q, agentID).Scan(&count); err != nil {
		return 0, fmt.Errorf("event repo: count by agent: %w", err)
	}
	return count, nil
}

// newEvent is a helper used internally to construct an Event with current time.
func newEvent(agentID int64, eventType string) models.Event {
	return models.Event{
		AgentID:   agentID,
		EventType: eventType,
		CreatedAt: time.Now().UTC(),
	}
}
