package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/vishwaszadte/ding/pkg/storage"
)

var historyLimit int

var historyCmd = &cobra.Command{
	Use:     "history",
	Aliases: []string{"log", "logs"},
	Short:   "View recent notification history and triggered alerts",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := storage.NewStore(flagDBPath)
		if err != nil {
			return fmt.Errorf("database error: %w", err)
		}
		defer store.Close()

		logs, err := store.ListHistory(historyLimit)
		if err != nil {
			return err
		}

		if len(logs) == 0 {
			fmt.Println("No notification history yet.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "TRIGGERED AT\tTITLE\tMESSAGE\tSTATUS")
		fmt.Fprintln(w, "------------\t-----\t-------\t------")

		for _, h := range logs {
			statusStr := "delivered"
			if h.WasMissed {
				statusStr = "missed (delivered on wake)"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				h.TriggeredAt.Local().Format("Jan 02 3:04 PM"),
				h.Title,
				h.Message,
				statusStr,
			)
		}
		return w.Flush()
	},
}

func init() {
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "n", 20, "Number of past notifications to display")
	rootCmd.AddCommand(historyCmd)
}
