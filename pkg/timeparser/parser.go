package timeparser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseResult holds the parsed scheduling outcome.
type ParseResult struct {
	OriginalExpr    string
	NextRunAt       time.Time
	IntervalSeconds int64
	IsRecurring     bool
	CronExpr        string
}

// In Go, package-level variables initialized with `var` run before any function.
// `regexp.MustCompile` compiles a regular expression at program startup.
// If the regex is invalid, it panics immediately (helping you catch bugs at startup).
var (
	// Matches relative expressions: "in 15m", "in 2 hours", "in 30s", "10m"
	relativeRegex = regexp.MustCompile(`^(?:in\s+)?(\d+)\s*(s|sec|secs|seconds?|m|min|mins|minutes?|h|hr|hrs|hours?|d|days?)$`)

	// Matches clock times: "today 4pm", "today 16:30", "tomorrow 9am", "4pm"
	clockRegex = regexp.MustCompile(`^(?:(today|tomorrow)\s+)?(\d{1,2})(?::(\d{2}))?\s*(am|pm)?$`)

	// Matches recurring intervals: "every hour", "every 2 hours", "every 30m"
	everyIntervalRegex = regexp.MustCompile(`^every\s+(\d+)?\s*(s|sec|secs|seconds?|m|min|mins|minutes?|h|hr|hrs|hours?|d|days?)$`)

	// Matches recurring days: "every weekday 9am", "every Monday at 10:30am", "every day 8pm"
	everyDayRegex = regexp.MustCompile(`^every\s+(day|weekday|monday|tuesday|wednesday|thursday|friday|saturday|sunday)(?:\s+at)?\s+(\d{1,2})(?::(\d{2}))?\s*(am|pm)?$`)
)

// Parse takes a natural language string and a reference `now` time,
// returning the computed NextRunAt timestamp and recurrence properties.
func Parse(input string, now time.Time) (*ParseResult, error) {
	expr := strings.TrimSpace(strings.ToLower(input))
	if expr == "" {
		return nil, fmt.Errorf("empty time expression")
	}

	// 1. Try relative duration ("in 15m", "2h", "30s")
	if matches := relativeRegex.FindStringSubmatch(expr); len(matches) > 0 {
		qty, _ := strconv.Atoi(matches[1])
		unit := matches[2]
		duration, err := toDuration(qty, unit)
		if err != nil {
			return nil, err
		}
		return &ParseResult{
			OriginalExpr: input,
			NextRunAt:    now.Add(duration),
			IsRecurring:  false,
		}, nil
	}

	// 2. Try recurring simple interval ("every hour", "every 30m")
	if matches := everyIntervalRegex.FindStringSubmatch(expr); len(matches) > 0 {
		qty := 1
		if matches[1] != "" {
			qty, _ = strconv.Atoi(matches[1])
		}
		unit := matches[2]
		duration, err := toDuration(qty, unit)
		if err != nil {
			return nil, err
		}
		return &ParseResult{
			OriginalExpr:    input,
			NextRunAt:       now.Add(duration),
			IntervalSeconds: int64(duration.Seconds()),
			IsRecurring:     true,
		}, nil
	}

	// 3. Try recurring day/weekday ("every weekday 9am", "every day 8pm", "every monday 10:30am")
	if matches := everyDayRegex.FindStringSubmatch(expr); len(matches) > 0 {
		daySpec := matches[1]
		hour, _ := strconv.Atoi(matches[2])
		minute := 0
		if matches[3] != "" {
			minute, _ = strconv.Atoi(matches[3])
		}
		ampm := matches[4]
		hour = normalizeHour(hour, ampm)

		nextTime, cronStr := calculateNextDayRecurrence(daySpec, hour, minute, now)
		return &ParseResult{
			OriginalExpr: input,
			NextRunAt:    nextTime,
			IsRecurring:  true,
			CronExpr:     cronStr,
		}, nil
	}

	// 4. Try clock time ("today 4pm", "tomorrow 9am", "4:30pm")
	if matches := clockRegex.FindStringSubmatch(expr); len(matches) > 0 {
		daySpec := matches[1]
		hour, _ := strconv.Atoi(matches[2])
		minute := 0
		if matches[3] != "" {
			minute, _ = strconv.Atoi(matches[3])
		}
		ampm := matches[4]
		hour = normalizeHour(hour, ampm)

		targetDate := now
		if daySpec == "tomorrow" {
			targetDate = targetDate.AddDate(0, 0, 1)
		}

		targetTime := time.Date(
			targetDate.Year(), targetDate.Month(), targetDate.Day(),
			hour, minute, 0, 0, now.Location(),
		)

		// If time has already passed today and user didn't explicitly say "today",
		// infer they mean tomorrow.
		if targetTime.Before(now) {
			if daySpec == "today" {
				return nil, fmt.Errorf("time %02d:%02d has already passed today", hour, minute)
			}
			targetTime = targetTime.AddDate(0, 0, 1)
		}

		return &ParseResult{
			OriginalExpr: input,
			NextRunAt:    targetTime,
			IsRecurring:  false,
		}, nil
	}

	return nil, fmt.Errorf("could not understand time expression %q. Try 'in 15m', 'today 4pm', 'tomorrow 9am', or 'every 1h'", input)
}

func toDuration(qty int, unit string) (time.Duration, error) {
	switch {
	case strings.HasPrefix(unit, "s"):
		return time.Duration(qty) * time.Second, nil
	case strings.HasPrefix(unit, "m"):
		return time.Duration(qty) * time.Minute, nil
	case strings.HasPrefix(unit, "h"):
		return time.Duration(qty) * time.Hour, nil
	case strings.HasPrefix(unit, "d"):
		return time.Duration(qty) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown time unit: %s", unit)
	}
}

func normalizeHour(hour int, ampm string) int {
	if ampm == "pm" && hour < 12 {
		return hour + 12
	}
	if ampm == "am" && hour == 12 {
		return 0
	}
	return hour
}

func calculateNextDayRecurrence(daySpec string, hour, minute int, now time.Time) (time.Time, string) {
	cron := fmt.Sprintf("%d %d * * *", minute, hour)

	switch daySpec {
	case "day":
		target := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if !target.After(now) {
			target = target.AddDate(0, 0, 1)
		}
		return target, cron

	case "weekday":
		cron = fmt.Sprintf("%d %d * * 1-5", minute, hour)
		target := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		for !target.After(now) || isWeekend(target.Weekday()) {
			target = target.AddDate(0, 0, 1)
		}
		return target, cron

	default: // Specific day of week (e.g. monday)
		targetWeekday := parseWeekday(daySpec)
		cron = fmt.Sprintf("%d %d * * %d", minute, hour, int(targetWeekday))
		target := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		for !target.After(now) || target.Weekday() != targetWeekday {
			target = target.AddDate(0, 0, 1)
		}
		return target, cron
	}
}

func isWeekend(wd time.Weekday) bool {
	return wd == time.Saturday || wd == time.Sunday
}

func parseWeekday(name string) time.Weekday {
	switch strings.ToLower(name) {
	case "sunday":
		return time.Sunday
	case "monday":
		return time.Monday
	case "tuesday":
		return time.Tuesday
	case "wednesday":
		return time.Wednesday
	case "thursday":
		return time.Thursday
	case "friday":
		return time.Friday
	case "saturday":
		return time.Saturday
	default:
		return time.Monday
	}
}
