// Package discovery implements the AI agent detection pipeline for AgentNetra.
// It takes a list of raw OS process metadata objects (models.ProcessInfo) and
// determines which ones correspond to known AI coding agents using a
// signature-based matching strategy.
//
// Detection pipeline:
//  1. Normalise the process name and executable path.
//  2. Load the signature registry.
//  3. Attempt each registered detector against the process.
//  4. On a match, classify and return an Agent record.
package discovery

import (
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/chethanyadav456/agentnetra/pkg/constants"
	"github.com/chethanyadav456/agentnetra/pkg/logger"
	"github.com/chethanyadav456/agentnetra/pkg/models"
)

// Detector is a function type that receives normalised process information
// and returns true when the process matches a particular AI agent.
type Detector func(name, exe, cmdline string) bool

// Engine runs the discovery pipeline over a slice of ProcessInfo and
// produces a slice of Agent records for any matches found.
type Engine struct {
	log       *logger.Logger
	detectors map[string]Detector
}

// NewEngine constructs a discovery Engine pre-loaded with all built-in
// agent detectors. The detectors map keys are the agent type names defined
// in pkg/constants (AgentTypeClaude, etc.).
func NewEngine(log *logger.Logger) *Engine {
	e := &Engine{
		log:       log,
		detectors: make(map[string]Detector),
	}
	e.registerBuiltins()
	return e
}

// registerBuiltins registers all out-of-the-box agent detectors.
func (e *Engine) registerBuiltins() {
	e.Register(constants.AgentTypeClaude, DetectClaude)
	e.Register(constants.AgentTypeCursor, DetectCursor)
	e.Register(constants.AgentTypeGemini, DetectGemini)
	e.Register(constants.AgentTypeAider, DetectAider)
	e.Register(constants.AgentTypeCodex, DetectCodex)
	e.Register(constants.AgentTypeOpenHands, DetectOpenHands)
}

// Register adds or replaces a Detector for the given agent type name.
// This allows external callers to extend the engine without modifying
// the built-in detector set.
func (e *Engine) Register(agentType string, d Detector) {
	e.detectors[agentType] = d
}

// Discover runs all registered detectors against the provided process list
// and returns Agent records for every matching process. It is safe to call
// concurrently from multiple goroutines.
func (e *Engine) Discover(procs []models.ProcessInfo) []models.Agent {
	now := time.Now().UTC()
	var agents []models.Agent

	for _, proc := range procs {
		name := normaliseName(proc.Name)
		// Pass the full lowercased executable path so detectors can match on
		// path components (e.g. "cursor.app", "node_modules/codex").
		exe := normaliseName(proc.ExecutablePath)
		cmdline := strings.ToLower(proc.CommandLine)

		for agentType, detect := range e.detectors {
			if detect(name, exe, cmdline) {
				agent := models.Agent{
					PID:            proc.PID,
					Name:           agentType,
					ExecutablePath: proc.ExecutablePath,
					Command:        proc.CommandLine,
					ParentPID:      proc.ParentPID,
					Status:         constants.AgentStatusRunning,
					DiscoveredAt:   now,
				}

				e.log.Info("discovery: agent detected",
					zap.String("agent_type", agentType),
					zap.Int32("pid", proc.PID),
					zap.String("name", proc.Name),
				)

				agents = append(agents, agent)
				break // A process can only match one agent type.
			}
		}
	}

	e.log.Debug("discovery: scan complete",
		zap.Int("processes_scanned", len(procs)),
		zap.Int("agents_found", len(agents)),
	)

	return agents
}

// normaliseName lowercases and trims whitespace from a process name or
// executable basename so detectors can do simple string comparisons.
func normaliseName(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
