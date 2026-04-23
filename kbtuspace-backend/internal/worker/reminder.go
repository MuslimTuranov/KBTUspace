package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
)

type ReminderWorker struct {
	db *sqlx.DB
}

func NewReminderWorker(db *sqlx.DB) *ReminderWorker {
	return &ReminderWorker{db: db}
}

func (w *ReminderWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	w.sendReminders(ctx)

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "reminder worker: shutting down")
			return
		case <-ticker.C:
			w.sendReminders(ctx)
		}
	}
}

type upcomingEvent struct {
	ID    int    `db:"id"`
	Title string `db:"title"`
}

type registrant struct {
	UserID int `db:"user_id"`
}

func (w *ReminderWorker) sendReminders(ctx context.Context) {
	now := time.Now()
	windowStart := now.Add(50 * time.Minute)
	windowEnd := now.Add(70 * time.Minute)

	var upcoming []upcomingEvent
	err := w.db.SelectContext(ctx, &upcoming, `
		SELECT id, title
		FROM posts
		WHERE event_date IS NOT NULL
		  AND status = 'approved'
		  AND reminder_sent_at IS NULL
		  AND event_date BETWEEN $1 AND $2
	`, windowStart, windowEnd)
	if err != nil {
		slog.ErrorContext(ctx, "reminder worker: failed to query upcoming events", slog.Any("error", err))
		return
	}

	slog.InfoContext(ctx, "reminder worker: tick", slog.Int("events_in_window", len(upcoming)))

	if len(upcoming) == 0 {
		return
	}

	for _, event := range upcoming {
		var registrants []registrant
		if err := w.db.SelectContext(ctx, &registrants, `
			SELECT user_id FROM registrations WHERE event_id = $1 AND status = 'registered'
		`, event.ID); err != nil {
			slog.ErrorContext(ctx, "reminder worker: failed to query registrants",
				slog.Int("event_id", event.ID), slog.Any("error", err))
			continue
		}

		slog.InfoContext(ctx, "reminder worker: sending reminders",
			slog.Int("event_id", event.ID),
			slog.String("title", event.Title),
			slog.Int("registrant_count", len(registrants)),
		)

		for _, r := range registrants {
			slog.InfoContext(ctx, "reminder worker: event starts in ~1 hour",
				slog.Int("user_id", r.UserID),
				slog.Int("event_id", event.ID),
				slog.String("event_title", event.Title),
			)
		}

		if _, err := w.db.ExecContext(ctx, `
			UPDATE posts
			SET reminder_sent_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`, event.ID); err != nil {
			slog.ErrorContext(ctx, "reminder worker: failed to mark reminder sent",
				slog.Int("event_id", event.ID), slog.Any("error", err))
		}
	}
}
