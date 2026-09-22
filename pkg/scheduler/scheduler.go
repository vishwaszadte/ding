package scheduler

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vishwaszadte/ding/pkg/model"
	"github.com/vishwaszadte/ding/pkg/notify"
	"github.com/vishwaszadte/ding/pkg/storage"
	"github.com/vishwaszadte/ding/pkg/timeparser"
)

// Scheduler coordinates the background ticker, monitors due reminders in SQLite,
// dispatches desktop notifications, recalculates recurrence, and logs history.
type Scheduler struct {
	store    *storage.Store
	notifier notify.Notifier
}

// NewScheduler creates a new Scheduler instance.
func NewScheduler(store *storage.Store, notifier notify.Notifier) *Scheduler {
	return &Scheduler{
		store:    store,
		notifier: notifier,
	}
}

// Run starts the scheduling loop. It blocks until the context is canceled (e.g. SIGINT/SIGTERM).
// In Go, goroutines and channels with select allow concurrent, non-blocking loops.
func (s *Scheduler) Run(ctx context.Context) error {
	// Check every second for due reminders
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Initial check on startup for any reminders missed while the machine was asleep/off
	s.checkDueReminders(time.Now())

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case t := <-ticker.C:
			s.checkDueReminders(t)
		}
	}
}

// checkDueReminders queries SQLite for active reminders whose next_run_at <= now.
func (s *Scheduler) checkDueReminders(now time.Time) {
	due, err := s.store.GetDueReminders(now)
	if err != nil {
		log.Printf("[scheduler] error querying due reminders: %v", err)
		return
	}

	for _, r := range due {
		s.fireReminder(r, now)
	}
}

// fireReminder dispatches the desktop notification, records history, and updates recurrence.
func (s *Scheduler) fireReminder(r *model.Reminder, now time.Time) {
	// Detect if this notification was missed while the computer was asleep (more than 45s late)
	wasMissed := now.Sub(r.NextRunAt) > 45*time.Second

	attribution := fmt.Sprintf("ding • %s", r.NextRunAt.Local().Format("3:04 PM"))
	if wasMissed {
		attribution = fmt.Sprintf("ding • missed at %s", r.NextRunAt.Local().Format("3:04 PM"))
	}
	if r.Tag != "" {
		tagStr := r.Tag
		if !strings.HasPrefix(tagStr, "#") {
			tagStr = "#" + tagStr
		}
		attribution += " • " + tagStr
	}

	n := notify.Notification{
		ID:          r.ID,
		Title:       r.Title,
		Message:     r.Message,
		Attribution: attribution,
		Sound:       r.Sound,
		Urgency:     r.Urgency,
		OpenURL:     r.OpenURL,
	}

	if err := s.notifier.Send(n); err != nil {
		log.Printf("[scheduler] failed to dispatch notification for reminder %s: %v", r.ID, err)
	} else {
		log.Printf("[scheduler] dispatched reminder %s: %q", r.ID, r.Title)
	}

	// Record to history log
	_ = s.store.RecordHistory(&model.NotificationHistory{
		ReminderID:  r.ID,
		Title:       r.Title,
		Message:     r.Message,
		ScheduledAt: r.NextRunAt,
		TriggeredAt: now,
		WasMissed:   wasMissed,
	})

	lastRun := now
	r.LastRunAt = &lastRun

	if r.IsRecurring {
		// Calculate next recurring time
		nextRun, err := s.calculateNextRecurrence(r, now)
		if err != nil {
			log.Printf("[scheduler] failed to recalculate recurrence for %s: %v", r.ID, err)
			r.Status = model.StatusCompleted
		} else {
			r.NextRunAt = nextRun
		}
	} else {
		r.Status = model.StatusCompleted
	}

	if err := s.store.UpdateReminder(r); err != nil {
		log.Printf("[scheduler] failed to update reminder %s: %v", r.ID, err)
	}
}

func (s *Scheduler) calculateNextRecurrence(r *model.Reminder, now time.Time) (time.Time, error) {
	if r.IntervalSeconds > 0 {
		interval := time.Duration(r.IntervalSeconds) * time.Second
		next := r.NextRunAt.Add(interval)
		// If system was asleep for a long time, advance until in the future
		for !next.After(now) {
			next = next.Add(interval)
		}
		return next, nil
	}

	// Re-evaluate natural language expression (e.g. "every weekday 9am", "every day 8pm")
	result, err := timeparser.Parse(r.Expression, now)
	if err != nil {
		return time.Time{}, err
	}
	return result.NextRunAt, nil
}
