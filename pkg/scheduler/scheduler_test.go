package scheduler

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/vishwaszadte/ding/pkg/model"
	"github.com/vishwaszadte/ding/pkg/notify"
	"github.com/vishwaszadte/ding/pkg/storage"
)

// mockNotifier captures sent notifications in memory for testing
type mockNotifier struct {
	sent []notify.Notification
}

func (m *mockNotifier) Send(n notify.Notification) error {
	m.sent = append(m.sent, n)
	return nil
}

func TestSchedulerFireDueReminder(t *testing.T) {
	tempDir := t.TempDir()
	store, err := storage.NewStore(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	mock := &mockNotifier{}
	sched := NewScheduler(store, mock)

	// Insert reminder due 1 second ago
	r := &model.Reminder{
		ID:         "sched-1",
		Title:      "Test Alert",
		Message:    "Should fire immediately",
		Expression: "in 1s",
		NextRunAt:  time.Now().Add(-1 * time.Second),
		Status:     model.StatusActive,
		Sound:      false,
	}
	if err := store.CreateReminder(r); err != nil {
		t.Fatalf("failed to create reminder: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Run scheduler in a goroutine
	go func() {
		_ = sched.Run(ctx)
	}()

	// Wait up to 1.5s for scheduler to fire
	time.Sleep(1500 * time.Millisecond)

	if len(mock.sent) == 0 {
		t.Fatal("expected at least 1 notification to be sent")
	}
	if mock.sent[0].Title != "Test Alert" {
		t.Errorf("expected title %q, got %q", "Test Alert", mock.sent[0].Title)
	}

	// Verify status updated to completed
	updated, err := store.GetReminder("sched-1")
	if err != nil {
		t.Fatalf("failed to fetch updated reminder: %v", err)
	}
	if updated.Status != model.StatusCompleted {
		t.Errorf("expected status completed, got %v", updated.Status)
	}
}
