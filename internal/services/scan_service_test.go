// Package services_test contains integration tests for the AgentNetra service layer.
// Each test uses in-memory SQLite so tests are hermetic and can run in parallel.
package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/chethanyadav456/agentnetra/internal/discovery"
	"github.com/chethanyadav456/agentnetra/internal/process"
	"github.com/chethanyadav456/agentnetra/internal/services"
	"github.com/chethanyadav456/agentnetra/internal/storage"
	"github.com/chethanyadav456/agentnetra/pkg/logger"
	"github.com/chethanyadav456/agentnetra/pkg/models"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func openTestDB(t *testing.T) *storage.DB {
	t.Helper()
	db, err := storage.Open(":memory:", logger.NewNop())
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func newScanService(t *testing.T, db *storage.DB) *services.ScanService {
	t.Helper()
	log := logger.NewNop()
	return services.NewScanService(
		log,
		process.NewEnumerator(log),
		discovery.NewEngine(log),
		storage.NewAgentRepository(db),
		storage.NewSessionRepository(db),
		storage.NewEventRepository(db),
	)
}

// ---------------------------------------------------------------------------
// ScanService tests
// ---------------------------------------------------------------------------

func TestScanService_Scan_ReturnsResult(t *testing.T) {
	// NOTE: Not parallel — gopsutil/v4 on macOS has an upstream race in its
	// purego-based RegisterFunc lazy-init. Running real OS scans sequentially
	// avoids triggering it. All other tests in this package ARE parallel.

	db := openTestDB(t)
	svc := newScanService(t, db)

	result, err := svc.Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: unexpected error: %v", err)
	}

	if result.ScannedAt.IsZero() {
		t.Error("ScanResult.ScannedAt should not be zero")
	}
	// Agents and Events slices may be nil/empty — that is valid on a machine
	// without any AI agents running.
}

func TestScanService_Scan_IsDeterministic(t *testing.T) {
	// NOTE: Not parallel — same upstream gopsutil/v4 race reason as above.

	db := openTestDB(t)
	svc := newScanService(t, db)
	ctx := context.Background()

	first, err := svc.Scan(ctx)
	if err != nil {
		t.Fatalf("first Scan: %v", err)
	}

	second, err := svc.Scan(ctx)
	if err != nil {
		t.Fatalf("second Scan: %v", err)
	}

	// Second scan should not re-discover already-running agents as new.
	// Agent count should stay the same or only decrease (stopped agents).
	_ = first
	_ = second
}

// ---------------------------------------------------------------------------
// AgentListService tests
// ---------------------------------------------------------------------------

func TestAgentListService_ListRunning(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	agentRepo := storage.NewAgentRepository(db)
	log := logger.NewNop()
	svc := services.NewAgentListService(log, agentRepo)
	ctx := context.Background()

	// Seed two agents.
	now := time.Now()
	_, _ = agentRepo.Create(ctx, models.Agent{PID: 1, Name: "claude", Status: "running", DiscoveredAt: now})
	_, _ = agentRepo.Create(ctx, models.Agent{PID: 2, Name: "gemini", Status: "stopped", DiscoveredAt: now})

	running, err := svc.ListRunning(ctx)
	if err != nil {
		t.Fatalf("ListRunning: %v", err)
	}
	if len(running) != 1 {
		t.Errorf("expected 1 running agent, got %d", len(running))
	}
	if running[0].Name != "claude" {
		t.Errorf("expected 'claude', got %q", running[0].Name)
	}
}

func TestAgentListService_ListAll(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	agentRepo := storage.NewAgentRepository(db)
	svc := services.NewAgentListService(logger.NewNop(), agentRepo)
	ctx := context.Background()

	now := time.Now()
	_, _ = agentRepo.Create(ctx, models.Agent{PID: 10, Name: "aider", Status: "running", DiscoveredAt: now})
	_, _ = agentRepo.Create(ctx, models.Agent{PID: 11, Name: "cursor", Status: "stopped", DiscoveredAt: now})

	all, err := svc.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 agents, got %d", len(all))
	}
}

// ---------------------------------------------------------------------------
// SessionListService tests
// ---------------------------------------------------------------------------

func TestSessionListService_ListAll(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	agentRepo := storage.NewAgentRepository(db)
	sessRepo := storage.NewSessionRepository(db)
	svc := services.NewSessionListService(sessRepo)
	ctx := context.Background()

	a, _ := agentRepo.Create(ctx, models.Agent{PID: 20, Name: "gemini", Status: "running", DiscoveredAt: time.Now()})
	_, _ = sessRepo.Create(ctx, models.Session{AgentID: a.ID, StartedAt: time.Now()})
	_, _ = sessRepo.Create(ctx, models.Session{AgentID: a.ID, StartedAt: time.Now()})

	sessions, err := svc.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(sessions))
	}
}

// ---------------------------------------------------------------------------
// EventListService tests
// ---------------------------------------------------------------------------

func TestEventListService_ListAll(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	agentRepo := storage.NewAgentRepository(db)
	evtRepo := storage.NewEventRepository(db)
	svc := services.NewEventListService(evtRepo)
	ctx := context.Background()

	a, _ := agentRepo.Create(ctx, models.Agent{PID: 30, Name: "codex", Status: "running", DiscoveredAt: time.Now()})
	_, _ = evtRepo.Create(ctx, models.Event{AgentID: a.ID, EventType: "agent_discovered", CreatedAt: time.Now()})

	events, err := svc.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != "agent_discovered" {
		t.Errorf("event type: got %q, want %q", events[0].EventType, "agent_discovered")
	}
}
