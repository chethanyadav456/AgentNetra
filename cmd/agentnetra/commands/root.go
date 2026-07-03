// Package commands contains the Cobra command definitions for the AgentNetra CLI.
// Each command is a separate function that returns a *cobra.Command. Commands
// receive all their dependencies via the Deps struct so there is no global
// state and the commands are independently testable.
package commands

import (
	"github.com/spf13/cobra"

	"github.com/chethanyadav456/agentnetra/internal/services"
	"github.com/chethanyadav456/agentnetra/pkg/constants"
	"github.com/chethanyadav456/agentnetra/pkg/logger"
)

// Deps bundles all service-layer dependencies that commands need. It is
// constructed once in main and passed into NewRootCmd.
type Deps struct {
	Log         *logger.Logger
	ScanService *services.ScanService
	AgentList   *services.AgentListService
	SessionList *services.SessionListService
	EventList   *services.EventListService
}

// NewRootCmd creates the top-level Cobra command and attaches all sub-commands.
// The returned *cobra.Command is ready to Execute().
func NewRootCmd(deps Deps) *cobra.Command {
	root := &cobra.Command{
		Use:   constants.AppName,
		Short: constants.AppDescription,
		Long: `AgentNetra — See What Your AI Agents Are Really Doing.

An open-source observability, security, and governance platform for
autonomous AI agents. Discover which AI coding agents are running on
your machine, track their sessions, and audit their lifecycle events.

Usage:
  agentnetra scan      Run a full system scan for AI agents
  agentnetra agents    List detected AI agents
  agentnetra sessions  List recorded agent sessions
  agentnetra events    List recorded agent lifecycle events`,

		// Suppress the default "Error: unknown command" message — we print
		// our own and return exit code 1.
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.Version = constants.AppVersion

	// Register sub-commands.
	root.AddCommand(newScanCmd(deps))
	root.AddCommand(newAgentsCmd(deps))
	root.AddCommand(newSessionsCmd(deps))
	root.AddCommand(newEventsCmd(deps))

	return root
}
