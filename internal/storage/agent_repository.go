// Package storage — AgentRepository
//
// AgentRepository provides Create, Update, and query methods for the agents
// table. All methods accept a context.Context so callers can impose deadlines.
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/chethanyadav456/agentnetra/pkg/models"
)

// AgentRepository performs CRUD operations against the agents table.
type AgentRepository struct {
	db *sql.DB
}

// NewAgentRepository constructs an AgentRepository backed by the provided DB.
func NewAgentRepository(db *DB) *AgentRepository {
	return &AgentRepository{db: db.Conn()}
}

// Create inserts a new agent record and populates the ID field of the returned
// Agent with the auto-assigned primary key.
func (r *AgentRepository) Create(ctx context.Context, agent models.Agent) (models.Agent, error) {
	const q = `
		INSERT INTO agents (pid, name, executable_path, command, parent_pid, status, discovered_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.ExecContext(ctx, q,
		agent.PID,
		agent.Name,
		agent.ExecutablePath,
		agent.Command,
		agent.ParentPID,
		agent.Status,
		agent.DiscoveredAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return models.Agent{}, fmt.Errorf("agent repo: create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return models.Agent{}, fmt.Errorf("agent repo: get last insert id: %w", err)
	}

	agent.ID = id
	return agent, nil
}

// UpdateStatus changes the status field of the agent identified by id.
func (r *AgentRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	const q = `UPDATE agents SET status = ? WHERE id = ?`
	if _, err := r.db.ExecContext(ctx, q, status, id); err != nil {
		return fmt.Errorf("agent repo: update status: %w", err)
	}
	return nil
}

// GetByID returns the agent with the given primary key, or sql.ErrNoRows when
// no matching record exists.
func (r *AgentRepository) GetByID(ctx context.Context, id int64) (models.Agent, error) {
	const q = `
		SELECT id, pid, name, executable_path, command, parent_pid, status, discovered_at
		FROM agents WHERE id = ?
	`
	row := r.db.QueryRowContext(ctx, q, id)
	return scanAgent(row)
}

// GetByPID returns the most recently discovered agent matching the given PID,
// or sql.ErrNoRows when none exists.
func (r *AgentRepository) GetByPID(ctx context.Context, pid int32) (models.Agent, error) {
	const q = `
		SELECT id, pid, name, executable_path, command, parent_pid, status, discovered_at
		FROM agents WHERE pid = ?
		ORDER BY discovered_at DESC LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, q, pid)
	return scanAgent(row)
}

// ListRunning returns all agents whose status is "running".
func (r *AgentRepository) ListRunning(ctx context.Context) ([]models.Agent, error) {
	const q = `
		SELECT id, pid, name, executable_path, command, parent_pid, status, discovered_at
		FROM agents WHERE status = 'running'
		ORDER BY discovered_at DESC
	`
	return r.queryAgents(ctx, q)
}

// ListAll returns all agent records ordered by discovery time descending.
func (r *AgentRepository) ListAll(ctx context.Context) ([]models.Agent, error) {
	const q = `
		SELECT id, pid, name, executable_path, command, parent_pid, status, discovered_at
		FROM agents
		ORDER BY discovered_at DESC
	`
	return r.queryAgents(ctx, q)
}

// queryAgents executes a query that returns agent rows and scans them.
func (r *AgentRepository) queryAgents(ctx context.Context, query string, args ...any) ([]models.Agent, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("agent repo: query: %w", err)
	}
	defer rows.Close()

	var agents []models.Agent
	for rows.Next() {
		agent, err := scanAgentRow(rows)
		if err != nil {
			return nil, fmt.Errorf("agent repo: scan: %w", err)
		}
		agents = append(agents, agent)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("agent repo: rows error: %w", err)
	}
	return agents, nil
}

// scanAgent scans a *sql.Row into a models.Agent.
func scanAgent(row *sql.Row) (models.Agent, error) {
	var a models.Agent
	var discoveredAt string
	err := row.Scan(
		&a.ID, &a.PID, &a.Name, &a.ExecutablePath,
		&a.Command, &a.ParentPID, &a.Status, &discoveredAt,
	)
	if err != nil {
		return models.Agent{}, err
	}
	a.DiscoveredAt, _ = time.Parse(time.RFC3339Nano, discoveredAt)
	return a, nil
}

// scanAgentRow scans a *sql.Rows into a models.Agent.
func scanAgentRow(rows *sql.Rows) (models.Agent, error) {
	var a models.Agent
	var discoveredAt string
	err := rows.Scan(
		&a.ID, &a.PID, &a.Name, &a.ExecutablePath,
		&a.Command, &a.ParentPID, &a.Status, &discoveredAt,
	)
	if err != nil {
		return models.Agent{}, err
	}
	a.DiscoveredAt, _ = time.Parse(time.RFC3339Nano, discoveredAt)
	return a, nil
}
