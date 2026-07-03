// Package constants defines application-wide constants for AgentNetra.
// These constants are shared across all packages and should not be modified
// at runtime.
package constants

// Application metadata constants.
const (
	// AppName is the canonical name of the application.
	AppName = "agentnetra"

	// AppVersion is the current semantic version of the application.
	AppVersion = "0.1.0"

	// AppDescription is the short description shown in CLI help text.
	AppDescription = "Observability, security, and governance platform for autonomous AI agents."
)

// Database constants.
const (
	// DBFileName is the default SQLite database filename.
	DBFileName = "agentnetra.db"

	// MigrationsDir is the embedded path prefix for migration SQL files.
	MigrationsDir = "migrations"
)

// Agent status constants define the lifecycle states of a discovered agent.
const (
	// AgentStatusRunning indicates the agent process is currently active.
	AgentStatusRunning = "running"

	// AgentStatusStopped indicates the agent process has exited.
	AgentStatusStopped = "stopped"

	// AgentStatusUnknown is used when status cannot be determined.
	AgentStatusUnknown = "unknown"
)

// Event type constants define lifecycle events recorded for agents.
const (
	// EventAgentDiscovered is emitted the first time an agent is seen.
	EventAgentDiscovered = "agent_discovered"

	// EventAgentStarted is emitted when a previously stopped agent resumes.
	EventAgentStarted = "agent_started"

	// EventAgentStopped is emitted when a running agent process exits.
	EventAgentStopped = "agent_stopped"

	// EventAgentRestarted is emitted when an agent stops and restarts.
	EventAgentRestarted = "agent_restarted"

	// EventAgentRemoved is emitted when an agent record is purged.
	EventAgentRemoved = "agent_removed"
)

// Known agent type identifiers used in signature matching.
const (
	AgentTypeClaude    = "claude"
	AgentTypeCursor    = "cursor"
	AgentTypeGemini    = "gemini"
	AgentTypeAider     = "aider"
	AgentTypeCodex     = "codex"
	AgentTypeOpenHands = "openhands"
)

// Exit codes returned by the CLI.
const (
	ExitCodeOK    = 0
	ExitCodeError = 1
)
