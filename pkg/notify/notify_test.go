package notify

import (
	"testing"
)

func TestNewNotifier(t *testing.T) {
	notifier, err := NewNotifier()
	if err != nil {
		t.Fatalf("expected NewNotifier to succeed, got: %v", err)
	}
	if notifier == nil {
		t.Fatal("expected notifier not to be nil")
	}
}
