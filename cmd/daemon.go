package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/vishwaszadte/ding/pkg/daemon"
	"github.com/vishwaszadte/ding/pkg/notify"
	"github.com/vishwaszadte/ding/pkg/scheduler"
	"github.com/vishwaszadte/ding/pkg/storage"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the background scheduling daemon",
}

var daemonRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the daemon scheduler in the foreground (used by background worker)",
	RunE: func(cmd *cobra.Command, args []string) error {
		running, pid := daemon.IsRunning()
		if running && pid != os.Getpid() {
			return fmt.Errorf("daemon is already running (PID: %d)", pid)
		}

		if err := daemon.SavePID(); err != nil {
			return fmt.Errorf("could not record PID: %w", err)
		}
		defer daemon.RemovePID()

		store, err := storage.NewStore(flagDBPath)
		if err != nil {
			return fmt.Errorf("database error: %w", err)
		}
		defer store.Close()

		notifier, err := notify.NewNotifier()
		if err != nil {
			return fmt.Errorf("notifier error: %w", err)
		}

		sched := scheduler.NewScheduler(store, notifier)

		// Listen for OS interrupt signals (Ctrl+C, SIGTERM)
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()

		log.Printf("[daemon] ding background daemon started (PID: %d)", os.Getpid())
		if err := sched.Run(ctx); err != nil && err != context.Canceled {
			return err
		}

		log.Println("[daemon] ding background daemon stopped gracefully")
		return nil
	},
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the background daemon process",
	RunE: func(cmd *cobra.Command, args []string) error {
		running, pid := daemon.IsRunning()
		if running {
			fmt.Printf("Daemon is already running (PID: %d)\n", pid)
			return nil
		}

		if err := daemon.AutoStart(); err != nil {
			return err
		}

		fmt.Println("✅ Background daemon started successfully!")
		return nil
	},
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background daemon process",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := daemon.Stop(); err != nil {
			return err
		}
		fmt.Println("🛑 Background daemon stopped.")
		return nil
	},
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check background daemon status",
	Run: func(cmd *cobra.Command, args []string) {
		running, pid := daemon.IsRunning()
		if running {
			fmt.Printf("✅ Daemon is running (PID: %d)\n", pid)
		} else {
			fmt.Println("⚪ Daemon is stopped")
		}
	},
}

var daemonLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Show recent daemon logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		logPath, err := daemon.LogPath()
		if err != nil {
			return err
		}

		data, err := os.ReadFile(logPath)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No log file found yet.")
				return nil
			}
			return err
		}

		fmt.Print(string(data))
		return nil
	},
}

func init() {
	daemonCmd.AddCommand(daemonRunCmd)
	daemonCmd.AddCommand(daemonStartCmd)
	daemonCmd.AddCommand(daemonStopCmd)
	daemonCmd.AddCommand(daemonStatusCmd)
	daemonCmd.AddCommand(daemonLogsCmd)
	rootCmd.AddCommand(daemonCmd)
}
