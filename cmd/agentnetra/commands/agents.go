// Package commands — agents command.
//
// agentnetra agents
//
// Lists AI agents stored in the local database.
package commands

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newAgentsCmd creates the "agents" sub-command.
func newAgentsCmd(deps Deps) *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "agents",
		Short: "List detected AI agents",
		Long: `List AI agents that have been detected by AgentNetra.

By default, only currently running agents are shown. Use --all to include
agents that are no longer running.

Run 'agentnetra scan' first to populate the database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if all {
				list, err := deps.AgentList.ListAll(ctx)
				if err != nil {
					deps.Log.Error("agents: list all failed", zap.Error(err))
					return fmt.Errorf("list agents: %w", err)
				}

				if len(list) == 0 {
					fmt.Println("No agents found. Run 'agentnetra scan' to discover agents.")
					return nil
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
				fmt.Fprintln(w, "ID\tAGENT\tPID\tSTATUS\tDISCOVERED AT\tCOMMAND")
				fmt.Fprintln(w, "--\t-----\t---\t------\t-------------\t-------")
				for _, a := range list {
					c := a.Command
					if len(c) > 50 {
						c = c[:47] + "..."
					}
					fmt.Fprintf(w, "%d\t%s\t%d\t%s\t%s\t%s\n",
						a.ID, a.Name, a.PID, a.Status,
						a.DiscoveredAt.Format("2006-01-02 15:04:05"), c)
				}
				_ = w.Flush()

				fmt.Printf("\nTotal: %d agent(s).\n", len(list))
			} else {
				list, err := deps.AgentList.ListRunning(ctx)
				if err != nil {
					deps.Log.Error("agents: list running failed", zap.Error(err))
					return fmt.Errorf("list running agents: %w", err)
				}

				if len(list) == 0 {
					fmt.Println("No running agents found. Run 'agentnetra scan' to detect active agents.")
					return nil
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
				fmt.Fprintln(w, "ID\tAGENT\tPID\tSTATUS\tDISCOVERED AT\tCOMMAND")
				fmt.Fprintln(w, "--\t-----\t---\t------\t-------------\t-------")
				for _, a := range list {
					c := a.Command
					if len(c) > 50 {
						c = c[:47] + "..."
					}
					fmt.Fprintf(w, "%d\t%s\t%d\t%s\t%s\t%s\n",
						a.ID, a.Name, a.PID, a.Status,
						a.DiscoveredAt.Format("2006-01-02 15:04:05"), c)
				}
				_ = w.Flush()

				fmt.Printf("\nRunning: %d agent(s). Use --all to include stopped agents.\n", len(list))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "Include stopped agents")
	return cmd
}
