// Package storage_test contains integration tests for the storage layer.
// Each test opens a fresh in-memory SQLite database so tests are hermetic
// and can run in parallel without file-system side effects.
package storage_test

import (
	"context"
	"testing"
	"time"

	"github.com/chethanyadav456/agentnetra/internal/storage"
	"github.com/chethanyadav456/agentnetra/pkg/logger"
	"github.com/chethanyadav456/agentnetra/pkg/models"
)

// openTestDB opens an in-memory SQLite database for testing.
func openTestDB(t *testing.T) *storage.DB {
	t.Helper()
	log := logger.NewNop()
	db, err := storage.Open(":memory:", log)
	if err != nil {
		t.Fatalf("storage.Open in-memory: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// ---------------------------------------------------------------------------
// AgentRepository tests
// ---------------------------------------------------------------------------

func TestAgentRepository_CreateAndGet(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	repo := storage.NewAgentRepository(db)
	ctx := context.Background()

	agent := models.Agent{
		PID:            12345,
		Name:           "claude",
		ExecutablePath: "/usr/local/bin/claude",
		Command:        "claude --dangerously-skip-permissions",
		ParentPID:      1000,
		Status:         "running",
		DiscoveredAt:   time.Now().UTC().Truncate(time.Second),
	}

	created, err := repo.Create(ctx, agent)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("Create: ID should be non-zero after insert")
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.PID != agent.PID {
		t.Errorf("PID: got %d, want %d", got.PID, agent.PID)
	}
	if got.Name != agent.Name {
		t.Errorf("Name: got %q, want %q", got.Name, agent.Name)
	}
	if got.Status != agent.Status {
		t.Errorf("Status: got %q, want %q", got.Status, agent.Status)
	}
}

func TestAgentRepository_UpdateStatus(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	repo := storage.NewAgentRepository(db)
	ctx := context.Background()

	agent := models.Agent{PID: 1, Name: "gemini", Status: "running", DiscoveredAt: time.Now()}
	created, err := repo.Create(ctx, agent)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := repo.UpdateStatus(ctx, created.ID, "stopped"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID after update: %v", err)
	}
	if got.Status != "stopped" {
		t.Errorf("Status: got %q, want %q", got.Status, "stopped")
	}
}

func TestAgentRepository_GetByPID(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	repo := storage.NewAgentRepository(db)
	ctx := context.Background()

	agent := models.Agent{PID: 9999, Name: "aider", Status: "running", DiscoveredAt: time.Now()}
	if _, err := repo.Create(ctx, agent); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByPID(ctx, 9999)
	if err != nil {
		t.Fatalf("GetByPID: %v", err)
	}
	if got.Name != "aider" {
		t.Errorf("Name: got %q, want %q", got.Name, "aider")
	}
}

func TestAgentRepository_ListRunning(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	repo := storage.NewAgentRepository(db)
	ctx := context.Background()

	now := time.Now()
	agents := []models.Agent{
		{PID: 1, Name: "claude", Status: "running", DiscoveredAt: now},
		{PID: 2, Name: "cursor", Status: "stopped", DiscoveredAt: now},
		{PID: 3, Name: "gemini", Status: "running", DiscoveredAt: now},
	}
	for _, a := range agents {
		if _, err := repo.Create(ctx, a); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	running, err := repo.ListRunning(ctx)
	if err != nil {
		t.Fatalf("ListRunning: %v", err)
	}
	if len(running) != 2 {
		t.Errorf("ListRunning: expected 2, got %d", len(running))
	}
}

// ---------------------------------------------------------------------------
// SessionRepository tests
// ---------------------------------------------------------------------------

func TestSessionRepository_CreateAndGet(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	agentRepo := storage.NewAgentRepository(db)
	sessRepo := storage.NewSessionRepository(db)
	ctx := context.Background()

	a, _ := agentRepo.Create(ctx, models.Agent{PID: 1, Name: "claude", Status: "running", DiscoveredAt: time.Now()})

	sess := models.Session{
		AgentID:   a.ID,
		StartedAt: time.Now().UTC(),
	}

	created, err := sessRepo.Create(ctx, sess)
	if err != nil {
		t.Fatalf("SessionRepo.Create: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("SessionRepo.Create: ID should be non-zero")
	}

	got, err := sessRepo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("SessionRepo.GetByID: %v", err)
	}
	if got.AgentID != a.ID {
		t.Errorf("AgentID: got %d, want %d", got.AgentID, a.ID)
	}
}

func TestSessionRepository_Close(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	agentRepo := storage.NewAgentRepository(db)
	sessRepo := storage.NewSessionRepository(db)
	ctx := context.Background()

	a, _ := agentRepo.Create(ctx, models.Agent{PID: 2, Name: "cursor", Status: "running", DiscoveredAt: time.Now()})

	startTime := time.Now().UTC()
	sess := models.Session{AgentID: a.ID, StartedAt: startTime}
	created, _ := sessRepo.Create(ctx, sess)

	endTime := startTime.Add(5 * time.Minute)
	if err := sessRepo.Close(ctx, created.ID, endTime); err != nil {
		t.Fatalf("SessionRepo.Close: %v", err)
	}

	got, err := sessRepo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("SessionRepo.GetByID after close: %v", err)
	}
	if got.Duration != 300 {
		t.Errorf("Duration: got %d, want 300", got.Duration)
	}
	if got.EndedAt.IsZero() {
		t.Error("EndedAt should be non-zero after Close")
	}
}

// ---------------------------------------------------------------------------
// EventRepository tests
// ---------------------------------------------------------------------------

func TestEventRepository_CreateAndList(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	agentRepo := storage.NewAgentRepository(db)
	evtRepo := storage.NewEventRepository(db)
	ctx := context.Background()

	a, _ := agentRepo.Create(ctx, models.Agent{PID: 3, Name: "aider", Status: "running", DiscoveredAt: time.Now()})

	evts := []models.Event{
		{AgentID: a.ID, EventType: "agent_discovered", CreatedAt: time.Now()},
		{AgentID: a.ID, EventType: "agent_started", CreatedAt: time.Now()},
	}
	for _, e := range evts {
		if _, err := evtRepo.Create(ctx, e); err != nil {
			t.Fatalf("EventRepo.Create: %v", err)
		}
	}

	listed, err := evtRepo.ListByAgent(ctx, a.ID)
	if err != nil {
		t.Fatalf("EventRepo.ListByAgent: %v", err)
	}
	if len(listed) != 2 {
		t.Errorf("ListByAgent: expected 2 events, got %d", len(listed))
	}

	count, err := evtRepo.CountByAgent(ctx, a.ID)
	if err != nil {
		t.Fatalf("EventRepo.CountByAgent: %v", err)
	}
	if count != 2 {
		t.Errorf("CountByAgent: got %d, want 2", count)
	}
}

func TestEventRepository_ListAll(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	agentRepo := storage.NewAgentRepository(db)
	evtRepo := storage.NewEventRepository(db)
	ctx := context.Background()

	a, _ := agentRepo.Create(ctx, models.Agent{PID: 4, Name: "gemini", Status: "running", DiscoveredAt: time.Now()})

	for i := range 3 {
		_, err := evtRepo.Create(ctx, models.Event{
			AgentID:   a.ID,
			EventType: "agent_discovered",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
		})
		if err != nil {
			t.Fatalf("Create event %d: %v", i, err)
		}
	}

	all, err := evtRepo.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("ListAll: expected 3, got %d", len(all))
	}
}
