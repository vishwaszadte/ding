package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/vishwaszadte/ding/pkg/model"
	"github.com/vishwaszadte/ding/pkg/storage"
)

var listStatus string

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all scheduled reminders",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := storage.NewStore(flagDBPath)
		if err != nil {
			return fmt.Errorf("database error: %w", err)
		}
		defer store.Close()

		var filter []model.Status
		if listStatus != "" {
			filter = append(filter, model.Status(listStatus))
		}

		reminders, err := store.ListReminders(filter...)
		if err != nil {
			return err
		}

		if len(reminders) == 0 {
			fmt.Println("No reminders found. Schedule one with: ding in 15m \"Take a break\"")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tNEXT RUN\tRECURRING\tSTATUS\tTAG")
		fmt.Fprintln(w, "--\t-----\t--------\t---------\t------\t---")

		for _, r := range reminders {
			recurringStr := "no"
			if r.IsRecurring {
				recurringStr = r.Expression
			}

			tagStr := "-"
			if r.Tag != "" {
				tagStr = "#" + r.Tag
			}

			nextRunStr := r.NextRunAt.Local().Format("Jan 02 3:04 PM")
			if r.Status == model.StatusActive {
				nextRunStr += fmt.Sprintf(" (%s)", formatTimeUntil(r.NextRunAt))
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				r.ID, r.Title, nextRunStr, recurringStr, r.Status, tagStr)
		}
		return w.Flush()
	},
}

func init() {
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "Filter by status: active, paused, completed, cancelled")
	rootCmd.AddCommand(listCmd)
}
