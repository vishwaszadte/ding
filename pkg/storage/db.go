package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	// Blank import: The underscore `_` tells Go to run this package's `init()`
	// function, which registers "sqlite" as an available driver in `database/sql`.
	_ "modernc.org/sqlite"

	"github.com/vishwaszadte/ding/pkg/model"
)

// Store encapsulates all database operations.
// In Go, structs managing connections hold a pointer to `*sql.DB`.
type Store struct {
	db *sql.DB
}

// NewStore initializes a SQLite database file at `customPath` (or defaults to ~/.ding/ding.db).
// It enables WAL mode for high concurrency and runs migrations.
func NewStore(customPath ...string) (*Store, error) {
	dbPath, err := resolveDBPath(customPath...)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve database path: %w", err)
	}

	// Ensure ~/.ding directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// SQLite PRAGMAs:
	// - WAL (Write-Ahead Logging): Allows concurrent readers without blocking writes.
	// - busy_timeout: Waits up to 5000ms if another process holds a write lock.
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA busy_timeout=5000;",
		"PRAGMA synchronous=NORMAL;",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to execute pragma %s: %w", pragma, err)
		}
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return s, nil
}

// Close closes the database connection pool.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// migrate creates the necessary tables and indexes if they don't exist.
func (s *Store) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS reminders (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		message TEXT NOT NULL,
		expression TEXT NOT NULL,
		cron_expr TEXT,
		interval_seconds INTEGER NOT NULL DEFAULT 0,
		is_recurring INTEGER NOT NULL DEFAULT 0,
		next_run_at DATETIME NOT NULL,
		last_run_at DATETIME,
		status TEXT NOT NULL,
		urgency TEXT NOT NULL,
		sound INTEGER NOT NULL DEFAULT 1,
		open_url TEXT,
		tag TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_reminders_status_next ON reminders(status, next_run_at);

	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		reminder_id TEXT NOT NULL,
		title TEXT NOT NULL,
		message TEXT NOT NULL,
		scheduled_at DATETIME NOT NULL,
		triggered_at DATETIME NOT NULL,
		was_missed INTEGER NOT NULL DEFAULT 0
	);

	CREATE INDEX IF NOT EXISTS idx_history_triggered ON history(triggered_at DESC);
	`
	_, err := s.db.Exec(query)
	return err
}

// CreateReminder inserts a new reminder into the database.
func (s *Store) CreateReminder(r *model.Reminder) error {
	if r.ID == "" {
		// Clean short 8-character ID for easy user typing (e.g. `ding cancel 3a4f89b1`)
		r.ID = uuid.New().String()[:8]
	}
	now := time.Now()
	if r.CreatedAt.IsZero() {
		r.CreatedAt = now
	}
	r.UpdatedAt = now

	query := `
	INSERT INTO reminders (
		id, title, message, expression, cron_expr, interval_seconds,
		is_recurring, next_run_at, last_run_at, status, urgency, sound,
		open_url, tag, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	soundInt := 0
	if r.Sound {
		soundInt = 1
	}
	recurringInt := 0
	if r.IsRecurring {
		recurringInt = 1
	}

	_, err := s.db.Exec(
		query,
		r.ID, r.Title, r.Message, r.Expression, r.CronExpr, r.IntervalSeconds,
		recurringInt, r.NextRunAt.UTC(), r.LastRunAt, string(r.Status),
		string(r.Urgency), soundInt, r.OpenURL, r.Tag, r.CreatedAt.UTC(), r.UpdatedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("could not insert reminder: %w", err)
	}
	return nil
}

// GetReminder retrieves a single reminder by ID.
func (s *Store) GetReminder(id string) (*model.Reminder, error) {
	query := `
	SELECT id, title, message, expression, cron_expr, interval_seconds,
	       is_recurring, next_run_at, last_run_at, status, urgency, sound,
	       open_url, tag, created_at, updated_at
	FROM reminders WHERE id = ?
	`
	row := s.db.QueryRow(query, id)
	return scanReminder(row)
}

// ListReminders returns reminders, optionally filtered by status.
func (s *Store) ListReminders(statusFilter ...model.Status) ([]*model.Reminder, error) {
	var query string
	var args []any

	if len(statusFilter) > 0 {
		query = `SELECT id, title, message, expression, cron_expr, interval_seconds,
		                is_recurring, next_run_at, last_run_at, status, urgency, sound,
		                open_url, tag, created_at, updated_at
		         FROM reminders WHERE status = ? ORDER BY next_run_at ASC`
		args = append(args, string(statusFilter[0]))
	} else {
		query = `SELECT id, title, message, expression, cron_expr, interval_seconds,
		                is_recurring, next_run_at, last_run_at, status, urgency, sound,
		                open_url, tag, created_at, updated_at
		         FROM reminders ORDER BY next_run_at ASC`
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query reminders: %w", err)
	}
	// 'defer' ensures rows.Close() is called when ListReminders exits!
	defer rows.Close()

	var results []*model.Reminder
	for rows.Next() {
		r, err := scanReminder(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// GetDueReminders returns active reminders scheduled on or before `now`.
func (s *Store) GetDueReminders(now time.Time) ([]*model.Reminder, error) {
	query := `
	SELECT id, title, message, expression, cron_expr, interval_seconds,
	       is_recurring, next_run_at, last_run_at, status, urgency, sound,
	       open_url, tag, created_at, updated_at
	FROM reminders
	WHERE status = ? AND next_run_at <= ?
	ORDER BY next_run_at ASC
	`
	rows, err := s.db.Query(query, string(model.StatusActive), now.UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to query due reminders: %w", err)
	}
	defer rows.Close()

	var results []*model.Reminder
	for rows.Next() {
		r, err := scanReminder(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// UpdateReminder updates mutable properties of an existing reminder.
func (s *Store) UpdateReminder(r *model.Reminder) error {
	r.UpdatedAt = time.Now()
	query := `
	UPDATE reminders SET
		title = ?, message = ?, expression = ?, cron_expr = ?,
		interval_seconds = ?, is_recurring = ?, next_run_at = ?,
		last_run_at = ?, status = ?, urgency = ?, sound = ?,
		open_url = ?, tag = ?, updated_at = ?
	WHERE id = ?
	`
	soundInt := 0
	if r.Sound {
		soundInt = 1
	}
	recurringInt := 0
	if r.IsRecurring {
		recurringInt = 1
	}

	res, err := s.db.Exec(
		query,
		r.Title, r.Message, r.Expression, r.CronExpr, r.IntervalSeconds,
		recurringInt, r.NextRunAt.UTC(), r.LastRunAt, string(r.Status),
		string(r.Urgency), soundInt, r.OpenURL, r.Tag, r.UpdatedAt.UTC(), r.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update reminder %s: %w", r.ID, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("reminder not found: %s", r.ID)
	}
	return nil
}

// DeleteReminder removes a reminder by ID.
func (s *Store) DeleteReminder(id string) error {
	res, err := s.db.Exec("DELETE FROM reminders WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete reminder %s: %w", id, err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("reminder not found: %s", id)
	}
	return nil
}

// RecordHistory logs a triggered notification.
func (s *Store) RecordHistory(h *model.NotificationHistory) error {
	query := `
	INSERT INTO history (reminder_id, title, message, scheduled_at, triggered_at, was_missed)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	missedInt := 0
	if h.WasMissed {
		missedInt = 1
	}
	_, err := s.db.Exec(query, h.ReminderID, h.Title, h.Message, h.ScheduledAt.UTC(), h.TriggeredAt.UTC(), missedInt)
	return err
}

// ListHistory retrieves recent notification events.
func (s *Store) ListHistory(limit int) ([]*model.NotificationHistory, error) {
	query := `
	SELECT id, reminder_id, title, message, scheduled_at, triggered_at, was_missed
	FROM history
	ORDER BY triggered_at DESC
	LIMIT ?
	`
	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	var logs []*model.NotificationHistory
	for rows.Next() {
		var h model.NotificationHistory
		var missedInt int
		if err := rows.Scan(&h.ID, &h.ReminderID, &h.Title, &h.Message, &h.ScheduledAt, &h.TriggeredAt, &missedInt); err != nil {
			return nil, err
		}
		h.WasMissed = (missedInt == 1)
		logs = append(logs, &h)
	}
	return logs, rows.Err()
}

// rowScanner is an unexported Go interface!
// Both *sql.Row (QueryRow) and *sql.Rows (Query) have a Scan(dest ...any) method.
// By defining this interface, one helper function scanReminder() can read both!
type rowScanner interface {
	Scan(dest ...any) error
}

func scanReminder(scanner rowScanner) (*model.Reminder, error) {
	var r model.Reminder
	var statusStr, urgencyStr string
	var recurringInt, soundInt int
	var cronExpr, openURL, tag sql.NullString
	var lastRunAt sql.NullTime

	err := scanner.Scan(
		&r.ID, &r.Title, &r.Message, &r.Expression, &cronExpr,
		&r.IntervalSeconds, &recurringInt, &r.NextRunAt, &lastRunAt,
		&statusStr, &urgencyStr, &soundInt, &openURL, &tag,
		&r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	r.Status = model.Status(statusStr)
	r.Urgency = model.Urgency(urgencyStr)
	r.IsRecurring = (recurringInt == 1)
	r.Sound = (soundInt == 1)

	if cronExpr.Valid {
		r.CronExpr = cronExpr.String
	}
	if openURL.Valid {
		r.OpenURL = openURL.String
	}
	if tag.Valid {
		r.Tag = tag.String
	}
	if lastRunAt.Valid {
		t := lastRunAt.Time
		r.LastRunAt = &t
	}

	return &r, nil
}

func resolveDBPath(customPath ...string) (string, error) {
	if len(customPath) > 0 && customPath[0] != "" {
		return customPath[0], nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ding", "ding.db"), nil
}
