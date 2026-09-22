package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/vishwaszadte/ding/pkg/model"
	"github.com/vishwaszadte/ding/pkg/notify"
)

var (
	testTitle   string
	testSound   bool
	testUrgency string
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Send an immediate notification to test your OS desktop notifications and sound",
	RunE: func(cmd *cobra.Command, args []string) error {
		notifier, err := notify.NewNotifier()
		if err != nil {
			return fmt.Errorf("could not create notifier: %w", err)
		}

		urgency := model.Urgency(testUrgency)
		if urgency == "" {
			urgency = model.UrgencyNormal
		}

		now := time.Now().Format("3:04 PM")
		fmt.Println("🔔 Dispatching modern desktop notification...")

		err = notifier.Send(notify.Notification{
			ID:          "test-preview",
			Title:       testTitle,
			Message:     "Ding is up and running! Your scheduled reminders will appear here.",
			Attribution: fmt.Sprintf("ding • %s • #test", now),
			Sound:       testSound,
			Urgency:     urgency,
		})

		if err != nil {
			return fmt.Errorf("failed to dispatch notification: %w", err)
		}

		fmt.Println("✅ Notification dispatched! Check the corner of your screen.")
		return nil
	},
}

func init() {
	testCmd.Flags().StringVarP(&testTitle, "title", "t", "Ding Notification", "Custom notification title")
	testCmd.Flags().BoolVar(&testSound, "sound", true, "Play audio chime with notification")
	testCmd.Flags().StringVar(&testUrgency, "urgency", "normal", "Urgency level: low, normal, critical")
	rootCmd.AddCommand(testCmd)
}
