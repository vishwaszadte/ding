//go:build windows

package notify

import (
	"encoding/xml"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/vishwaszadte/ding/pkg/assets"
	"github.com/vishwaszadte/ding/pkg/model"
)

type windowsNotifier struct {
	iconPath string
}

func newPlatformNotifier() (Notifier, error) {
	icon, _, _ := assets.EnsureAssetsWritten()
	return &windowsNotifier{
		iconPath: icon,
	}, nil
}

// Send displays a modern Windows 10/11 WinRT toast notification using Fluent Design XML.
func (w *windowsNotifier) Send(n Notification) error {
	if n.Title == "" {
		n.Title = "🔔 Reminder"
	}
	if !strings.HasPrefix(n.Title, "🔔") && !strings.HasPrefix(n.Title, "⏰") {
		n.Title = "🔔 " + n.Title
	}

	icon := n.IconPath
	if icon == "" {
		icon = w.iconPath
	}
	iconURI := "file:///" + filepath.ToSlash(icon)

	attrib := n.Attribution
	if attrib == "" {
		attrib = "ding"
	}

	// Native Windows notification audio chime
	audioXML := `<audio silent="true" />`
	if n.Sound {
		audioXML = `<audio src="ms-winsoundevent:Notification.Default" silent="false" />`
	}

	// Escape strings to prevent XML injection
	var titleEscaped, bodyEscaped, attribEscaped strings.Builder
	_ = xml.EscapeText(&titleEscaped, []byte(n.Title))
	_ = xml.EscapeText(&bodyEscaped, []byte(n.Message))
	_ = xml.EscapeText(&attribEscaped, []byte(attrib))

	scenario := "default"
	if n.Urgency == model.UrgencyCritical {
		scenario = "reminder" // Critical reminders stay on screen longer
	}

	toastXML := fmt.Sprintf(`
<toast scenario="%s">
    <visual>
        <binding template="ToastGeneric">
            <image placement="appLogoOverride" src="%s" />
            <text hint-style="title">%s</text>
            <text hint-style="body">%s</text>
            <text hint-style="attribution">%s</text>
        </binding>
    </visual>
    %s
    <actions>
        <action content="Dismiss" arguments="dismiss" activationType="system" />
    </actions>
</toast>`, scenario, iconURI, titleEscaped.String(), bodyEscaped.String(), attribEscaped.String(), audioXML)

	// PowerShell WinRT ToastNotificationManager script
	psScript := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null

$template = @'
%s
'@

$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
$xml.LoadXml($template)
$toast = [Windows.UI.Notifications.ToastNotification]::new($xml)

$notifier = [Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('{1AC14E77-02E7-4E5D-B744-2EB1AE5198B7}\WindowsPowerShell\v1.0\powershell.exe')
$notifier.Show($toast)
`, toastXML)

	// Run PowerShell hidden in the background without popping a console window
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", psScript)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to show windows toast: %w", err)
	}

	return nil
}
