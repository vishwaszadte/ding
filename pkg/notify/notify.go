package notify

import (
	"github.com/vishwaszadte/ding/pkg/model"
)

// Notification represents a visual alert sent to the desktop.
type Notification struct {
	ID          string
	Title       string
	Message     string
	Attribution string // e.g. "ding • 4:00 PM • #health"
	Sound       bool
	Urgency     model.Urgency
	OpenURL     string
	IconPath    string
}

// Notifier is the common interface that every OS implementation satisfies.
// In Go, interfaces are satisfied implicitly (duck typing):
// any struct that implements `Send(Notification) error` is a Notifier.
type Notifier interface {
	Send(n Notification) error
}

// NewNotifier returns the appropriate Notifier for the current operating system.
// This function delegates to newPlatformNotifier(), which is defined separately
// in toast_windows.go, toast_darwin.go, and toast_linux.go using Go build tags!
func NewNotifier() (Notifier, error) {
	return newPlatformNotifier()
}
