package storage

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/vishwaszadte/ding/pkg/model"
)

func TestStoreCRUD(t *testing.T) {
	// t.TempDir() automatically creates an isolated temp directory for this test
	// and cleans it up when the test finishes!
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_ding.db")

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("expected store to initialize, got: %v", err)
	}
	defer store.Close()

	reminder := &model.Reminder{
		ID:              "test-123",
		Title:           "Water Break",
		Message:         "Drink 250ml water",
		Expression:      "every 1h",
		IntervalSeconds: 3600,
		IsRecurring:     true,
		NextRunAt:       time.Now().Add(1 * time.Hour),
		Status:          model.StatusActive,
		Urgency:         model.UrgencyNormal,
		Sound:           true,
		Tag:             "health",
	}

	t.Run("Create and Get Reminder", func(t *testing.T) {
		err := store.CreateReminder(reminder)
		if err != nil {
			t.Fatalf("CreateReminder failed: %v", err)
		}

		fetched, err := store.GetReminder("test-123")
		if err != nil {
			t.Fatalf("GetReminder failed: %v", err)
		}
		if fetched.Title != reminder.Title {
			t.Errorf("expected Title %q, got %q", reminder.Title, fetched.Title)
		}
		if fetched.Tag != "health" {
			t.Errorf("expected Tag %q, got %q", "health", fetched.Tag)
		}
	})

	t.Run("List Active Reminders", func(t *testing.T) {
		list, err := store.ListReminders(model.StatusActive)
		if err != nil {
			t.Fatalf("ListReminders failed: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 active reminder, got %d", len(list))
		}
	})

	t.Run("Update Reminder", func(t *testing.T) {
		reminder.Status = model.StatusPaused
		err := store.UpdateReminder(reminder)
		if err != nil {
			t.Fatalf("UpdateReminder failed: %v", err)
		}

		updated, err := store.GetReminder("test-123")
		if err != nil {
			t.Fatalf("GetReminder failed: %v", err)
		}
		if updated.Status != model.StatusPaused {
			t.Errorf("expected status paused, got %v", updated.Status)
		}
	})

	t.Run("Record and List History", func(t *testing.T) {
		historyItem := &model.NotificationHistory{
			ReminderID:  "test-123",
			Title:       "Test Notification",
			Message:     "Delivered",
			ScheduledAt: time.Now().Add(-1 * time.Minute),
			TriggeredAt: time.Now(),
			WasMissed:   false,
		}
		if err := store.RecordHistory(historyItem); err != nil {
			t.Fatalf("RecordHistory failed: %v", err)
		}

		history, err := store.ListHistory(10)
		if err != nil {
			t.Fatalf("ListHistory failed: %v", err)
		}
		if len(history) != 1 {
			t.Fatalf("expected 1 history record, got %d", len(history))
		}
	})

	t.Run("Delete Reminder", func(t *testing.T) {
		err := store.DeleteReminder("test-123")
		if err != nil {
			t.Fatalf("DeleteReminder failed: %v", err)
		}

		_, err = store.GetReminder("test-123")
		if err == nil {
			t.Errorf("expected error fetching deleted reminder, got nil")
		}
	})
}
