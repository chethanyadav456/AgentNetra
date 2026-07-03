// Package commands — scan command.
//
// agentnetra scan
//
// Performs a full system scan for AI agents, persists results to the local
// database, and prints a formatted summary table to stdout.
package commands

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newScanCmd creates the "scan" sub-command.
func newScanCmd(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Scan the system for running AI agents",
		Long: `Perform a full system scan to detect running AI agents.

AgentNetra will enumerate all running processes and identify known AI
coding agents (Claude Code, Cursor, Gemini CLI, Aider, Codex CLI, OpenHands).

Detected agents are stored in the local database. Run 'agentnetra agents'
to list them after a scan.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			deps.Log.Info("scan: initiating system scan")

			result, err := deps.ScanService.Scan(ctx)
			if err != nil {
				deps.Log.Error("scan: failed", zap.Error(err))
				return fmt.Errorf("scan failed: %w", err)
			}

			if len(result.Agents) == 0 {
				fmt.Println("No AI agents found on this machine.")
				return nil
			}

			// Print results table.
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "AGENT\tPID\tPARENT PID\tSTATUS\tDISCOVERED AT\tCOMMAND")
			fmt.Fprintln(w, "-----\t---\t----------\t------\t-------------\t-------")

			for _, a := range result.Agents {
				cmd := a.Command
				if len(cmd) > 60 {
					cmd = cmd[:57] + "..."
				}
				fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%s\t%s\n",
					a.Name,
					a.PID,
					a.ParentPID,
					a.Status,
					a.DiscoveredAt.Format("2006-01-02 15:04:05"),
					cmd,
				)
			}
			_ = w.Flush()

			fmt.Printf("\nScan complete. Found %d agent(s). Events: %d.\n",
				len(result.Agents), len(result.Events))
			return nil
		},
	}
}
