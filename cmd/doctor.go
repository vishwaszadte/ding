package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/vishwaszadte/ding/pkg/daemon"
	"github.com/vishwaszadte/ding/pkg/notify"
	"github.com/vishwaszadte/ding/pkg/storage"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose system health, daemon status, and notification permissions",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🔍 Running ding health check...")
		fmt.Println()

		// 1. Operating system
		fmt.Printf("  💻 OS / Arch:           %s / %s\n", runtime.GOOS, runtime.GOARCH)

		// 2. Storage
		store, err := storage.NewStore(flagDBPath)
		if err != nil {
			fmt.Printf("  ❌ Database:            Error: %v\n", err)
		} else {
			reminders, _ := store.ListReminders()
			fmt.Printf("  ✅ Database:            Connected (%d total reminders)\n", len(reminders))
			store.Close()
		}

		// 3. Daemon
		running, pid := daemon.IsRunning()
		if running {
			fmt.Printf("  ✅ Background Daemon:   Running (PID: %d)\n", pid)
		} else {
			fmt.Println("  ⚠️  Background Daemon:   Not running (will auto-start when scheduling)")
		}

		// 4. Notifier Subsystem
		_, err = notify.NewNotifier()
		if err != nil {
			fmt.Printf("  ❌ Notification System: Error: %v\n", err)
		} else {
			fmt.Println("  ✅ Notification System: Ready (Native Desktop Engine with 🔔 Bell Emoji)")
		}

		fmt.Println()
		fmt.Println("Everything looks great! You're ready to schedule notifications.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
