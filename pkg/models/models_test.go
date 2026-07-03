// Package models_test contains unit tests for the AgentNetra domain models.
package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/chethanyadav456/agentnetra/pkg/models"
)

func TestAgentJSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)
	agent := models.Agent{
		ID:             1,
		PID:            12345,
		Name:           "claude",
		ExecutablePath: "/usr/local/bin/claude",
		Command:        "claude --dangerously-skip-permissions",
		ParentPID:      1000,
		Status:         "running",
		DiscoveredAt:   now,
	}

	data, err := json.Marshal(agent)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var got models.Agent
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if got.ID != agent.ID {
		t.Errorf("ID: got %d, want %d", got.ID, agent.ID)
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
	if !got.DiscoveredAt.Equal(agent.DiscoveredAt) {
		t.Errorf("DiscoveredAt: got %v, want %v", got.DiscoveredAt, agent.DiscoveredAt)
	}
}

func TestSessionJSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)
	session := models.Session{
		ID:        1,
		AgentID:   42,
		StartedAt: now,
		EndedAt:   now.Add(10 * time.Minute),
		Duration:  600,
	}

	data, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var got models.Session
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if got.Duration != session.Duration {
		t.Errorf("Duration: got %d, want %d", got.Duration, session.Duration)
	}
}

func TestEventJSONRoundTrip(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC().Truncate(time.Second)
	event := models.Event{
		ID:        1,
		AgentID:   42,
		EventType: "agent_discovered",
		CreatedAt: now,
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	var got models.Event
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if got.EventType != event.EventType {
		t.Errorf("EventType: got %q, want %q", got.EventType, event.EventType)
	}
}

func TestProcessInfoFields(t *testing.T) {
	t.Parallel()

	pi := models.ProcessInfo{
		PID:            999,
		ParentPID:      1,
		Name:           "aider",
		ExecutablePath: "/home/user/.local/bin/aider",
		CommandLine:    "aider --model gpt-4o",
		Username:       "chethan",
	}

	if pi.PID != 999 {
		t.Errorf("PID: got %d, want %d", pi.PID, 999)
	}
	if pi.Name != "aider" {
		t.Errorf("Name: got %q, want %q", pi.Name, "aider")
	}
}

func TestScanResultFields(t *testing.T) {
	t.Parallel()

	result := models.ScanResult{
		Agents:    []models.Agent{{Name: "gemini"}},
		Events:    []models.Event{{EventType: "agent_discovered"}},
		ScannedAt: time.Now(),
	}

	if len(result.Agents) != 1 {
		t.Errorf("Agents length: got %d, want 1", len(result.Agents))
	}
	if len(result.Events) != 1 {
		t.Errorf("Events length: got %d, want 1", len(result.Events))
	}
}
