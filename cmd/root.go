package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/vishwaszadte/ding/pkg/daemon"
	"github.com/vishwaszadte/ding/pkg/model"
	"github.com/vishwaszadte/ding/pkg/storage"
	"github.com/vishwaszadte/ding/pkg/timeparser"
)

var (
	flagDBPath  string
	flagMessage string
	flagTitle   string
	flagSound   bool
	flagUrgency string
	flagOpenURL string
	flagTag     string
)

var rootCmd = &cobra.Command{
	Use:   "ding [time-expression] [message]",
	Short: "🔔 ding: Aesthetic cross-platform scheduled desktop notifications",
	Long: `ding is a modern, cross-platform CLI tool for scheduling aesthetic desktop notifications.

Examples:
  ding every hour --message "get up and walk"
  ding today 4pm --message "do this task"
  ding tomorrow 6am --message "something"
  ding in 15m "Tea is ready"
  ding every weekday 9am -m "Team Standup" -t "Daily Sync" --sound`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 && flagMessage == "" {
			return cmd.Help()
		}

		timeExpr, msg := extractTimeAndMessage(args, flagMessage)
		if timeExpr == "" {
			return fmt.Errorf("please provide a time expression (e.g. 'in 15m', 'today 4pm', 'every hour')")
		}
		if msg == "" {
			msg = "Time's up!"
		}

		// Parse the natural language time expression
		result, err := timeparser.Parse(timeExpr, time.Now())
		if err != nil {
			return err
		}

		// Initialize storage
		store, err := storage.NewStore(flagDBPath)
		if err != nil {
			return fmt.Errorf("database error: %w", err)
		}
		defer store.Close()

		title := flagTitle
		if title == "" {
			title = "Reminder"
		}

		urgency := model.Urgency(flagUrgency)
		if urgency == "" {
			urgency = model.UrgencyNormal
		}

		reminder := &model.Reminder{
			Title:           title,
			Message:         msg,
			Expression:      result.OriginalExpr,
			CronExpr:        result.CronExpr,
			IntervalSeconds: result.IntervalSeconds,
			IsRecurring:     result.IsRecurring,
			NextRunAt:       result.NextRunAt,
			Status:          model.StatusActive,
			Urgency:         urgency,
			Sound:           flagSound,
			OpenURL:         flagOpenURL,
			Tag:             flagTag,
		}

		if err := store.CreateReminder(reminder); err != nil {
			return fmt.Errorf("failed to save reminder: %w", err)
		}

		// Auto-spawn the background daemon if not already running!
		if err := daemon.AutoStart(); err != nil {
			fmt.Printf("⚠️ Warning: Could not auto-start background daemon: %v\n", err)
		}

		// Display clean confirmation to user
		printScheduledConfirmation(reminder)
		return nil
	},
}

// Execute is the main entrypoint called by main.go.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagDBPath, "db", "", "Custom path to SQLite database (defaults to ~/.ding/ding.db)")

	rootCmd.Flags().StringVarP(&flagMessage, "message", "m", "", "Reminder message")
	rootCmd.Flags().StringVarP(&flagTitle, "title", "t", "", "Notification title (defaults to 'Reminder')")
	rootCmd.Flags().BoolVar(&flagSound, "sound", true, "Play audio chime on notification")
	rootCmd.Flags().StringVarP(&flagUrgency, "urgency", "u", "normal", "Urgency level: low, normal, critical")
	rootCmd.Flags().StringVar(&flagOpenURL, "open", "", "URL to open on click")
	rootCmd.Flags().StringVar(&flagTag, "tag", "", "Category tag (e.g. 'health', 'work')")
}

// extractTimeAndMessage separates the scheduling expression from the message.
// Handles both:
//   ding every hour --message "walk"   -> time: "every hour", msg: "walk"
//   ding in 15m "Take a break"        -> time: "in 15m",     msg: "Take a break"
//   ding today 4pm "Team sync"        -> time: "today 4pm",  msg: "Team sync"
func extractTimeAndMessage(args []string, flagMsg string) (string, string) {
	if len(args) == 0 {
		return "", flagMsg
	}

	if flagMsg != "" {
		// All positional arguments form the time expression
		return strings.Join(args, " "), flagMsg
	}

	// If the last argument looks like a quoted message or sentence:
	if len(args) >= 2 {
		// Check if everything except the last arg is a valid time expression
		candidateTime := strings.Join(args[:len(args)-1], " ")
		if _, err := timeparser.Parse(candidateTime, time.Now()); err == nil {
			return candidateTime, args[len(args)-1]
		}
	}

	// Try first argument as time (e.g. "15m", "4pm") and rest as message
	if len(args) >= 2 {
		candidateTime := args[0]
		if _, err := timeparser.Parse(candidateTime, time.Now()); err == nil {
			return candidateTime, strings.Join(args[1:], " ")
		}
	}

	// Fallback: all arguments form the time expression
	return strings.Join(args, " "), ""
}

func printScheduledConfirmation(r *model.Reminder) {
	fmt.Println()
	fmt.Printf("  🔔 Scheduled: %s\n", r.Title)
	if r.Message != "" {
		fmt.Printf("  💬 Message:   %q\n", r.Message)
	}
	fmt.Printf("  ⏰ Due:       %s (%s)\n",
		r.NextRunAt.Local().Format("Monday, Jan 02 at 3:04 PM"),
		formatTimeUntil(r.NextRunAt))
	if r.IsRecurring {
		fmt.Printf("  🔁 Repeats:   %s\n", r.Expression)
	}
	fmt.Printf("  🆔 ID:        %s\n", r.ID)
	fmt.Println()
}

func formatTimeUntil(t time.Time) string {
	d := time.Until(t).Round(time.Minute)
	if d <= 0 {
		return "just now"
	}
	return "in " + d.String()
}
