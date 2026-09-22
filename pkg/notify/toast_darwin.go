//go:build darwin

package notify

import (
	"fmt"
	"os/exec"
	"strings"
)

type darwinNotifier struct{}

func newPlatformNotifier() (Notifier, error) {
	return &darwinNotifier{}, nil
}

func (d *darwinNotifier) Send(n Notification) error {
	if n.Title == "" {
		n.Title = "Reminder"
	}
	if !strings.HasPrefix(n.Title, "🔔") {
		n.Title = "🔔 " + n.Title
	}

	soundClause := ""
	if n.Sound {
		soundClause = `sound name "default"`
	}

	subtitleClause := ""
	if n.Attribution != "" {
		subtitleClause = fmt.Sprintf(`subtitle %q`, n.Attribution)
	}

	script := fmt.Sprintf(`display notification %q with title %q %s %s`,
		n.Message, n.Title, subtitleClause, soundClause)

	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}
