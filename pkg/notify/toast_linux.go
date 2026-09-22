//go:build linux

package notify

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/vishwaszadte/ding/pkg/model"
)

type linuxNotifier struct{}

func newPlatformNotifier() (Notifier, error) {
	return &linuxNotifier{}, nil
}

func (l *linuxNotifier) Send(n Notification) error {
	if n.Title == "" {
		n.Title = "Reminder"
	}
	if !strings.HasPrefix(n.Title, "🔔") {
		n.Title = "🔔 " + n.Title
	}

	urgency := "normal"
	if n.Urgency == model.UrgencyLow {
		urgency = "low"
	} else if n.Urgency == model.UrgencyCritical {
		urgency = "critical"
	}

	args := []string{
		"-u", urgency,
		"-a", "ding",
	}

	body := n.Message
	if n.Attribution != "" {
		body = fmt.Sprintf("%s\n<small>%s</small>", n.Message, n.Attribution)
	}

	args = append(args, n.Title, body)

	cmd := exec.Command("notify-send", args...)
	return cmd.Run()
}
