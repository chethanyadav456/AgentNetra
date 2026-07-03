// Package models defines the core domain models shared across all AgentNetra
// packages. These structs represent the canonical data shapes for agents,
// sessions, and events — the three primary entities of the platform.
package models

import "time"

// Agent represents a detected AI agent process on the host machine.
// It holds both the real-time process metadata and the persisted record
// identity assigned by the storage layer.
type Agent struct {
	// ID is the database primary key. Zero means the record is not yet persisted.
	ID int64 `json:"id"`

	// PID is the operating system process identifier.
	PID int32 `json:"pid"`

	// Name is the human-readable agent type name (e.g., "claude", "cursor").
	Name string `json:"name"`

	// ExecutablePath is the absolute path to the agent binary on disk.
	ExecutablePath string `json:"executable_path"`

	// Command is the full command line string used to launch the process.
	Command string `json:"command"`

	// ParentPID is the PID of the parent process that spawned this agent.
	ParentPID int32 `json:"parent_pid"`

	// Status is the current lifecycle state of the agent (running/stopped/unknown).
	Status string `json:"status"`

	// DiscoveredAt is the wall-clock time when AgentNetra first observed this agent.
	DiscoveredAt time.Time `json:"discovered_at"`
}

// Session represents a continuous run interval of an agent.
// A new session is opened each time an agent is discovered as running,
// and closed when it transitions to stopped.
type Session struct {
	// ID is the database primary key.
	ID int64 `json:"id"`

	// AgentID is the foreign key linking this session to its parent Agent record.
	AgentID int64 `json:"agent_id"`

	// StartedAt is the time this session began.
	StartedAt time.Time `json:"started_at"`

	// EndedAt is the time this session ended. Zero value means still running.
	EndedAt time.Time `json:"ended_at"`

	// Duration is the session length in seconds. Populated on session close.
	Duration int64 `json:"duration"`
}

// Event represents a discrete lifecycle occurrence observed for an agent.
// Events form an immutable append-only audit log.
type Event struct {
	// ID is the database primary key.
	ID int64 `json:"id"`

	// AgentID is the foreign key linking this event to its originating Agent.
	AgentID int64 `json:"agent_id"`

	// EventType is the classification string (e.g., "agent_discovered").
	// Use the constants.Event* constants to set this field.
	EventType string `json:"event_type"`

	// CreatedAt is the time the event was generated.
	CreatedAt time.Time `json:"created_at"`
}

// ProcessInfo holds raw process metadata collected from the operating system.
// This is an intermediate value object used by the process engine and
// consumed by the discovery engine. It is never persisted directly.
type ProcessInfo struct {
	// PID is the operating system process identifier.
	PID int32 `json:"pid"`

	// ParentPID is the PID of the parent process.
	ParentPID int32 `json:"parent_pid"`

	// Name is the short process name as reported by the OS (e.g., "node", "python3").
	Name string `json:"name"`

	// ExecutablePath is the resolved path to the binary on disk.
	ExecutablePath string `json:"executable_path"`

	// CommandLine is the full invocation string including all arguments.
	CommandLine string `json:"command_line"`

	// Username is the OS user that owns the process.
	Username string `json:"username"`
}

// ScanResult is the output produced by a single scan run. It aggregates
// all discovered agents together with any events generated during the scan.
type ScanResult struct {
	// Agents is the list of AI agents found during this scan.
	Agents []Agent `json:"agents"`

	// Events is the list of lifecycle events produced during this scan.
	Events []Event `json:"events"`

	// ScannedAt is the time the scan was performed.
	ScannedAt time.Time `json:"scanned_at"`
}
