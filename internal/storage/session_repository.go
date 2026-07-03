// Package storage — SessionRepository
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/chethanyadav456/agentnetra/pkg/models"
)

// SessionRepository performs CRUD operations against the sessions table.
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository constructs a SessionRepository backed by the provided DB.
func NewSessionRepository(db *DB) *SessionRepository {
	return &SessionRepository{db: db.Conn()}
}

// Create inserts a new session record and returns it with the assigned ID.
func (r *SessionRepository) Create(ctx context.Context, session models.Session) (models.Session, error) {
	const q = `
		INSERT INTO sessions (agent_id, started_at, ended_at, duration)
		VALUES (?, ?, ?, ?)
	`
	var endedAt any
	if !session.EndedAt.IsZero() {
		endedAt = session.EndedAt.UTC().Format(time.RFC3339Nano)
	}

	result, err := r.db.ExecContext(ctx, q,
		session.AgentID,
		session.StartedAt.UTC().Format(time.RFC3339Nano),
		endedAt,
		session.Duration,
	)
	if err != nil {
		return models.Session{}, fmt.Errorf("session repo: create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return models.Session{}, fmt.Errorf("session repo: get last insert id: %w", err)
	}

	session.ID = id
	return session, nil
}

// Close updates ended_at and duration for the session identified by id.
func (r *SessionRepository) Close(ctx context.Context, id int64, endedAt time.Time) error {
	// Fetch the session to calculate duration.
	sess, err := r.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("session repo: close: fetch session: %w", err)
	}

	duration := int64(endedAt.Sub(sess.StartedAt).Seconds())

	const q = `UPDATE sessions SET ended_at = ?, duration = ? WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, q,
		endedAt.UTC().Format(time.RFC3339Nano),
		duration,
		id,
	); err != nil {
		return fmt.Errorf("session repo: close: update: %w", err)
	}
	return nil
}

// GetByID returns the session with the given primary key.
func (r *SessionRepository) GetByID(ctx context.Context, id int64) (models.Session, error) {
	const q = `
		SELECT id, agent_id, started_at, ended_at, duration
		FROM sessions WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, q, id)
	return scanSession(row)
}

// ListByAgent returns all sessions for the given agent ID, newest first.
func (r *SessionRepository) ListByAgent(ctx context.Context, agentID int64) ([]models.Session, error) {
	const q = `
		SELECT id, agent_id, started_at, ended_at, duration
		FROM sessions WHERE agent_id = ?
		ORDER BY started_at DESC
	`
	rows, err := r.db.QueryContext(ctx, q, agentID)
	if err != nil {
		return nil, fmt.Errorf("session repo: list by agent: %w", err)
	}
	defer rows.Close()
	return scanSessions(rows)
}

// ListAll returns all sessions, newest first.
func (r *SessionRepository) ListAll(ctx context.Context) ([]models.Session, error) {
	const q = `
		SELECT id, agent_id, started_at, ended_at, duration
		FROM sessions ORDER BY started_at DESC
	`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("session repo: list all: %w", err)
	}
	defer rows.Close()
	return scanSessions(rows)
}

// scanSession scans a *sql.Row into a Session.
func scanSession(row *sql.Row) (models.Session, error) {
	var s models.Session
	var startedAt string
	var endedAt sql.NullString

	if err := row.Scan(&s.ID, &s.AgentID, &startedAt, &endedAt, &s.Duration); err != nil {
		return models.Session{}, err
	}
	s.StartedAt, _ = time.Parse(time.RFC3339Nano, startedAt)
	if endedAt.Valid {
		s.EndedAt, _ = time.Parse(time.RFC3339Nano, endedAt.String)
	}
	return s, nil
}

// scanSessions scans *sql.Rows into a slice of Session.
func scanSessions(rows *sql.Rows) ([]models.Session, error) {
	var sessions []models.Session
	for rows.Next() {
		var s models.Session
		var startedAt string
		var endedAt sql.NullString

		if err := rows.Scan(&s.ID, &s.AgentID, &startedAt, &endedAt, &s.Duration); err != nil {
			return nil, fmt.Errorf("session repo: scan row: %w", err)
		}
		s.StartedAt, _ = time.Parse(time.RFC3339Nano, startedAt)
		if endedAt.Valid {
			s.EndedAt, _ = time.Parse(time.RFC3339Nano, endedAt.String)
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("session repo: rows error: %w", err)
	}
	return sessions, nil
}
