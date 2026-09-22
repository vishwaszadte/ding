package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vishwaszadte/ding/pkg/model"
	"github.com/vishwaszadte/ding/pkg/storage"
)

var cancelAll bool

var cancelCmd = &cobra.Command{
	Use:     "cancel [id]",
	Aliases: []string{"rm", "delete"},
	Short:   "Cancel a scheduled reminder by ID (or --all)",
	Args:    cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := storage.NewStore(flagDBPath)
		if err != nil {
			return fmt.Errorf("database error: %w", err)
		}
		defer store.Close()

		if cancelAll {
			reminders, err := store.ListReminders(model.StatusActive)
			if err != nil {
				return err
			}
			for _, r := range reminders {
				_ = store.DeleteReminder(r.ID)
			}
			fmt.Printf("✅ Cancelled all %d active reminders.\n", len(reminders))
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("please specify a reminder ID to cancel, or use --all")
		}

		id := args[0]
		reminder, err := store.GetReminder(id)
		if err != nil {
			return fmt.Errorf("reminder %q not found", id)
		}

		if err := store.DeleteReminder(id); err != nil {
			return err
		}

		fmt.Printf("✅ Cancelled reminder %s (%q)\n", id, reminder.Title)
		return nil
	},
}

func init() {
	cancelCmd.Flags().BoolVarP(&cancelAll, "all", "a", false, "Cancel all active reminders")
	rootCmd.AddCommand(cancelCmd)
}
