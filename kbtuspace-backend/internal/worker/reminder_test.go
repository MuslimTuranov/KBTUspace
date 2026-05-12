package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func setupReminderTest(t *testing.T) (*ReminderWorker, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	worker := NewReminderWorker(sqlxDB)

	return worker, mock, func() {
		_ = sqlxDB.Close()
		_ = db.Close()
	}
}

func TestSendRemindersFindsApprovedEventsIn50To70Minutes(t *testing.T) {
	worker, mock, cleanup := setupReminderTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, title\s+FROM posts\s+WHERE event_date IS NOT NULL\s+AND status = 'approved'\s+AND reminder_sent_at IS NULL\s+AND event_date BETWEEN \$1 AND \$2`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title"}).
				AddRow(1, "Test Event"),
		)

	mock.ExpectQuery(`SELECT user_id FROM registrations WHERE event_id = \$1 AND status = 'registered'`).
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"user_id"}).
				AddRow(10).
				AddRow(11),
		)

	mock.ExpectExec(`UPDATE posts\s+SET reminder_sent_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP\s+WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	worker.sendReminders(context.Background())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSendRemindersIgnoresReminderSentAndUnapprovedEventsByQuery(t *testing.T) {
	worker, mock, cleanup := setupReminderTest(t)
	defer cleanup()

	mock.ExpectQuery(`status = 'approved'\s+AND reminder_sent_at IS NULL`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}))

	worker.sendReminders(context.Background())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSendRemindersUsesOnlyRegisteredRegistrations(t *testing.T) {
	worker, mock, cleanup := setupReminderTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, title`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title"}).
				AddRow(1, "Event"),
		)

	mock.ExpectQuery(`SELECT user_id FROM registrations WHERE event_id = \$1 AND status = 'registered'`).
		WithArgs(1).
		WillReturnRows(
			sqlmock.NewRows([]string{"user_id"}).
				AddRow(10),
		)

	mock.ExpectExec(`UPDATE posts`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	worker.sendReminders(context.Background())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSendRemindersSetsReminderSentAtAfterProcessing(t *testing.T) {
	worker, mock, cleanup := setupReminderTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, title`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "title"}).
				AddRow(1, "Event"),
		)

	mock.ExpectQuery(`SELECT user_id FROM registrations`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

	mock.ExpectExec(`UPDATE posts\s+SET reminder_sent_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP\s+WHERE id = \$1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	worker.sendReminders(context.Background())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestSendRemindersQueryErrorDoesNotPanic(t *testing.T) {
	worker, mock, cleanup := setupReminderTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, title`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(errors.New("db error"))

	worker.sendReminders(context.Background())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestStartStopsOnContextShutdown(t *testing.T) {
	worker, mock, cleanup := setupReminderTest(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())

	mock.ExpectQuery(`SELECT id, title`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title"}))

	done := make(chan struct{})

	go func() {
		worker.Start(ctx)
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after context cancel")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
