package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vishwaszadte/ding/pkg/model"
	"github.com/vishwaszadte/ding/pkg/storage"
)

var pauseCmd = &cobra.Command{
	Use:   "pause [id]",
	Short: "Pause a recurring reminder",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := storage.NewStore(flagDBPath)
		if err != nil {
			return fmt.Errorf("database error: %w", err)
		}
		defer store.Close()

		id := args[0]
		r, err := store.GetReminder(id)
		if err != nil {
			return fmt.Errorf("reminder %q not found", id)
		}

		r.Status = model.StatusPaused
		if err := store.UpdateReminder(r); err != nil {
			return err
		}

		fmt.Printf("⏸️ Paused reminder %s (%q)\n", id, r.Title)
		return nil
	},
}

var resumeCmd = &cobra.Command{
	Use:   "resume [id]",
	Short: "Resume a paused reminder",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := storage.NewStore(flagDBPath)
		if err != nil {
			return fmt.Errorf("database error: %w", err)
		}
		defer store.Close()

		id := args[0]
		r, err := store.GetReminder(id)
		if err != nil {
			return fmt.Errorf("reminder %q not found", id)
		}

		r.Status = model.StatusActive
		if err := store.UpdateReminder(r); err != nil {
			return err
		}

		fmt.Printf("▶️ Resumed reminder %s (%q)\n", id, r.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pauseCmd)
	rootCmd.AddCommand(resumeCmd)
}
