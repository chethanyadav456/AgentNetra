// Package commands — events command.
//
// agentnetra events
//
// Lists recorded agent lifecycle events stored in the local database.
package commands

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newEventsCmd creates the "events" sub-command.
func newEventsCmd(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "events",
		Short: "List recorded agent lifecycle events",
		Long: `List all recorded lifecycle events for AI agents.

Events capture agent state transitions such as:
  - agent_discovered  — First time an agent is seen
  - agent_started     — Agent transitioned from stopped to running
  - agent_stopped     — Agent process exited
  - agent_restarted   — Agent stopped and restarted

Events are append-only and form a permanent audit log.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			events, err := deps.EventList.ListAll(ctx)
			if err != nil {
				deps.Log.Error("events: list failed", zap.Error(err))
				return fmt.Errorf("list events: %w", err)
			}

			if len(events) == 0 {
				fmt.Println("No events found. Run 'agentnetra scan' to start tracking.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tAGENT ID\tEVENT TYPE\tCREATED AT")
			fmt.Fprintln(w, "--\t--------\t----------\t----------")

			for _, e := range events {
				fmt.Fprintf(w, "%d\t%d\t%s\t%s\n",
					e.ID, e.AgentID, e.EventType,
					e.CreatedAt.Format("2006-01-02 15:04:05"),
				)
			}
			_ = w.Flush()

			fmt.Printf("\nTotal events: %d.\n", len(events))
			return nil
		},
	}
}
