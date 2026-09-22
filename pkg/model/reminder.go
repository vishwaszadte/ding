package model

import (
	"time"
)

// Urgency defines how prominent the desktop notification should be.
type Urgency string

const (
	UrgencyLow      Urgency = "low"
	UrgencyNormal   Urgency = "normal"
	UrgencyCritical Urgency = "critical"
)

// Status represents the lifecycle state of a reminder.
type Status string

const (
	StatusActive    Status = "active"
	StatusPaused    Status = "paused"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
)

// Reminder is the core domain model representing a scheduled notification.
type Reminder struct {
	ID              string     `json:"id" db:"id"`
	Title           string     `json:"title" db:"title"`
	Message         string     `json:"message" db:"message"`
	Expression      string     `json:"expression" db:"expression"`             // e.g. "every 1h", "in 15m"
	CronExpr        string     `json:"cron_expr,omitempty" db:"cron_expr"`     // optional standard cron
	IntervalSeconds int64      `json:"interval_seconds" db:"interval_seconds"` // for recurring interval (e.g. 3600 for 1h)
	IsRecurring     bool       `json:"is_recurring" db:"is_recurring"`
	NextRunAt       time.Time  `json:"next_run_at" db:"next_run_at"`
	LastRunAt       *time.Time `json:"last_run_at,omitempty" db:"last_run_at"` // Pointer allows nil (NULL in DB)
	Status          Status     `json:"status" db:"status"`
	Urgency         Urgency    `json:"urgency" db:"urgency"`
	Sound           bool       `json:"sound" db:"sound"`
	OpenURL         string     `json:"open_url,omitempty" db:"open_url"`
	Tag             string     `json:"tag,omitempty" db:"tag"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// NotificationHistory tracks past triggered events for logging and auditing.
type NotificationHistory struct {
	ID          int64     `json:"id" db:"id"`
	ReminderID  string    `json:"reminder_id" db:"reminder_id"`
	Title       string    `json:"title" db:"title"`
	Message     string    `json:"message" db:"message"`
	ScheduledAt time.Time `json:"scheduled_at" db:"scheduled_at"`
	TriggeredAt time.Time `json:"triggered_at" db:"triggered_at"`
	WasMissed   bool      `json:"was_missed" db:"was_missed"` // true if delivered late after computer woke from sleep
}
