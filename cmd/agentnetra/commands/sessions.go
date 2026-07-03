// Package commands — sessions command.
//
// agentnetra sessions
//
// Lists recorded agent sessions stored in the local database.
package commands

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newSessionsCmd creates the "sessions" sub-command.
func newSessionsCmd(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "sessions",
		Short: "List recorded agent sessions",
		Long: `List all recorded sessions for detected AI agents.

A session is opened when an agent is discovered and closed when it stops.
Duration is shown in seconds for open sessions-in-progress and completed ones.

Run 'agentnetra scan' to discover agents and create sessions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			sessions, err := deps.SessionList.ListAll(ctx)
			if err != nil {
				deps.Log.Error("sessions: list failed", zap.Error(err))
				return fmt.Errorf("list sessions: %w", err)
			}

			if len(sessions) == 0 {
				fmt.Println("No sessions found. Run 'agentnetra scan' to start tracking.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tAGENT ID\tSTARTED AT\tENDED AT\tDURATION (s)")
			fmt.Fprintln(w, "--\t--------\t----------\t--------\t-----------")

			for _, s := range sessions {
				endedAt := "-"
				if !s.EndedAt.IsZero() {
					endedAt = s.EndedAt.Format("2006-01-02 15:04:05")
				}
				fmt.Fprintf(w, "%d\t%d\t%s\t%s\t%d\n",
					s.ID, s.AgentID,
					s.StartedAt.Format("2006-01-02 15:04:05"),
					endedAt,
					s.Duration,
				)
			}
			_ = w.Flush()

			fmt.Printf("\nTotal sessions: %d.\n", len(sessions))
			return nil
		},
	}
}
