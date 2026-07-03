// Package services implements the business logic layer of AgentNetra.
// It orchestrates the process engine, discovery engine, and storage layer
// to provide high-level operations like ScanForAgents that the CLI can call
// without knowing about the individual subsystems.
//
// The service layer is the only place that combines multiple internal packages.
package services

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/chethanyadav456/agentnetra/internal/discovery"
	"github.com/chethanyadav456/agentnetra/internal/process"
	"github.com/chethanyadav456/agentnetra/internal/storage"
	"github.com/chethanyadav456/agentnetra/pkg/constants"
	"github.com/chethanyadav456/agentnetra/pkg/logger"
	"github.com/chethanyadav456/agentnetra/pkg/models"
)

// ScanService orchestrates a full agent discovery scan. It enumerates
// processes, identifies AI agents, persists new records, and generates
// lifecycle events.
type ScanService struct {
	log       *logger.Logger
	enumerator *process.Enumerator
	engine    *discovery.Engine
	agents    *storage.AgentRepository
	sessions  *storage.SessionRepository
	events    *storage.EventRepository
}

// NewScanService constructs a ScanService with all required dependencies.
func NewScanService(
	log *logger.Logger,
	enumerator *process.Enumerator,
	engine *discovery.Engine,
	agents *storage.AgentRepository,
	sessions *storage.SessionRepository,
	events *storage.EventRepository,
) *ScanService {
	return &ScanService{
		log:       log,
		enumerator: enumerator,
		engine:    engine,
		agents:    agents,
		sessions:  sessions,
		events:    events,
	}
}

// Scan performs a full system scan:
//  1. Enumerates all running processes via the process engine.
//  2. Passes the process list through the discovery engine.
//  3. For each discovered agent:
//     a. Creates or updates the agent record in the database.
//     b. Opens a new session if this is the first time we see it.
//     c. Emits an agent_discovered or agent_started event.
//  4. Marks previously-running agents as stopped if they are no longer found.
//
// Scan returns a ScanResult summarising everything found and every event emitted.
func (s *ScanService) Scan(ctx context.Context) (models.ScanResult, error) {
	result := models.ScanResult{ScannedAt: time.Now().UTC()}

	s.log.Info("scan: starting agent discovery scan")

	// --- Step 1: Enumerate OS processes. ---
	procs, err := s.enumerator.Processes()
	if err != nil {
		return result, fmt.Errorf("scan: enumerate processes: %w", err)
	}
	s.log.Debug("scan: processes enumerated", zap.Int("count", len(procs)))

	// --- Step 2: Discover AI agents. ---
	discovered := s.engine.Discover(procs)
	s.log.Info("scan: discovery complete", zap.Int("agents_found", len(discovered)))

	// --- Step 3: Persist each discovered agent. ---
	for _, agent := range discovered {
		persisted, evt, err := s.persistAgent(ctx, agent)
		if err != nil {
			s.log.Warn("scan: failed to persist agent",
				zap.String("name", agent.Name),
				zap.Int32("pid", agent.PID),
				zap.Error(err),
			)
			continue
		}
		result.Agents = append(result.Agents, persisted)
		if evt != nil {
			result.Events = append(result.Events, *evt)
		}
	}

	// --- Step 4: Mark vanished agents as stopped. ---
	if err := s.reconcileStoppedAgents(ctx, discovered, &result); err != nil {
		s.log.Warn("scan: reconcile stopped agents failed", zap.Error(err))
	}

	s.log.Info("scan: scan complete",
		zap.Int("agents_persisted", len(result.Agents)),
		zap.Int("events_generated", len(result.Events)),
	)

	return result, nil
}

// persistAgent upserts a discovered agent into the database and emits the
// appropriate lifecycle event. It returns the persisted agent and optionally
// an event record.
func (s *ScanService) persistAgent(ctx context.Context, agent models.Agent) (models.Agent, *models.Event, error) {
	// Check whether we already have a record for this PID.
	existing, err := s.agents.GetByPID(ctx, agent.PID)
	if err == nil && existing.ID != 0 {
		// Agent already known. Update status to running if it was stopped.
		if existing.Status != constants.AgentStatusRunning {
			if err := s.agents.UpdateStatus(ctx, existing.ID, constants.AgentStatusRunning); err != nil {
				return models.Agent{}, nil, fmt.Errorf("persistAgent: update status: %w", err)
			}
			existing.Status = constants.AgentStatusRunning

			// Open a new session.
			sess := models.Session{AgentID: existing.ID, StartedAt: time.Now().UTC()}
			if _, err := s.sessions.Create(ctx, sess); err != nil {
				s.log.Warn("persistAgent: create session failed", zap.Error(err))
			}

			// Emit agent_started event.
			evt, err := s.emitEvent(ctx, existing.ID, constants.EventAgentStarted)
			if err != nil {
				s.log.Warn("persistAgent: emit event failed", zap.Error(err))
			}
			return existing, &evt, nil
		}
		// Agent was already running — no event needed.
		return existing, nil, nil
	}

	// New agent — create record.
	created, err := s.agents.Create(ctx, agent)
	if err != nil {
		return models.Agent{}, nil, fmt.Errorf("persistAgent: create agent: %w", err)
	}

	// Open initial session.
	sess := models.Session{AgentID: created.ID, StartedAt: created.DiscoveredAt}
	if _, err := s.sessions.Create(ctx, sess); err != nil {
		s.log.Warn("persistAgent: create initial session failed", zap.Error(err))
	}

	// Emit agent_discovered event.
	evt, err := s.emitEvent(ctx, created.ID, constants.EventAgentDiscovered)
	if err != nil {
		s.log.Warn("persistAgent: emit discovered event failed", zap.Error(err))
	}

	return created, &evt, nil
}

// reconcileStoppedAgents compares the current running agents in the database
// against the freshly discovered set. Any agent in the DB with status=running
// that no longer appears in the discovered list is marked as stopped and
// receives an agent_stopped event.
func (s *ScanService) reconcileStoppedAgents(
	ctx context.Context,
	discovered []models.Agent,
	result *models.ScanResult,
) error {
	// Build a PID set for the current scan.
	livePIDs := make(map[int32]struct{}, len(discovered))
	for _, a := range discovered {
		livePIDs[a.PID] = struct{}{}
	}

	// Load all currently-running agents from the database.
	runningInDB, err := s.agents.ListRunning(ctx)
	if err != nil {
		return fmt.Errorf("reconcile: list running agents: %w", err)
	}

	for _, dbAgent := range runningInDB {
		if _, isAlive := livePIDs[dbAgent.PID]; isAlive {
			continue
		}

		// Agent has vanished — mark as stopped.
		if err := s.agents.UpdateStatus(ctx, dbAgent.ID, constants.AgentStatusStopped); err != nil {
			s.log.Warn("reconcile: update status failed",
				zap.Int64("agent_id", dbAgent.ID),
				zap.Error(err),
			)
			continue
		}

		// Emit agent_stopped event.
		evt, err := s.emitEvent(ctx, dbAgent.ID, constants.EventAgentStopped)
		if err != nil {
			s.log.Warn("reconcile: emit stopped event failed", zap.Error(err))
			continue
		}
		result.Events = append(result.Events, evt)

		s.log.Info("scan: agent marked as stopped",
			zap.String("name", dbAgent.Name),
			zap.Int32("pid", dbAgent.PID),
		)
	}

	return nil
}

// emitEvent creates and persists a lifecycle event for the given agent.
func (s *ScanService) emitEvent(ctx context.Context, agentID int64, eventType string) (models.Event, error) {
	evt := models.Event{
		AgentID:   agentID,
		EventType: eventType,
		CreatedAt: time.Now().UTC(),
	}
	return s.events.Create(ctx, evt)
}

// ---------------------------------------------------------------------------
// AgentListService — thin read-only service used by the "agents" CLI command.
// ---------------------------------------------------------------------------

// AgentListService provides read operations over persisted agent records.
type AgentListService struct {
	log    *logger.Logger
	agents *storage.AgentRepository
}

// NewAgentListService constructs an AgentListService.
func NewAgentListService(log *logger.Logger, agents *storage.AgentRepository) *AgentListService {
	return &AgentListService{log: log, agents: agents}
}

// ListRunning returns all agents whose status is "running".
func (s *AgentListService) ListRunning(ctx context.Context) ([]models.Agent, error) {
	return s.agents.ListRunning(ctx)
}

// ListAll returns all agent records regardless of status.
func (s *AgentListService) ListAll(ctx context.Context) ([]models.Agent, error) {
	return s.agents.ListAll(ctx)
}

// ---------------------------------------------------------------------------
// SessionListService
// ---------------------------------------------------------------------------

// SessionListService provides read operations over persisted session records.
type SessionListService struct {
	sessions *storage.SessionRepository
}

// NewSessionListService constructs a SessionListService.
func NewSessionListService(sessions *storage.SessionRepository) *SessionListService {
	return &SessionListService{sessions: sessions}
}

// ListAll returns all recorded sessions newest-first.
func (s *SessionListService) ListAll(ctx context.Context) ([]models.Session, error) {
	return s.sessions.ListAll(ctx)
}

// ---------------------------------------------------------------------------
// EventListService
// ---------------------------------------------------------------------------

// EventListService provides read operations over persisted event records.
type EventListService struct {
	events *storage.EventRepository
}

// NewEventListService constructs an EventListService.
func NewEventListService(events *storage.EventRepository) *EventListService {
	return &EventListService{events: events}
}

// ListAll returns all recorded events newest-first.
func (s *EventListService) ListAll(ctx context.Context) ([]models.Event, error) {
	return s.events.ListAll(ctx)
}
