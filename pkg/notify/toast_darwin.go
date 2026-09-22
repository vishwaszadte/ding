//go:build darwin

package notify

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/vishwaszadte/ding/pkg/assets"
)

type darwinNotifier struct {
	iconPath string
}

func newPlatformNotifier() (Notifier, error) {
	icon, _, _ := assets.EnsureAssetsWritten()
	return &darwinNotifier{iconPath: icon}, nil
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
