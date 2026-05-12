package reports

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"kbtuspace-backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

type mockReportsCache struct {
	deleted       []string
	deletedPrefix []string
}

func (m *mockReportsCache) SetPost(key string, value *models.Post) error {
	return nil
}

func (m *mockReportsCache) GetPost(key string) (*models.Post, bool, error) {
	return nil, false, nil
}

func (m *mockReportsCache) SetPosts(key string, value []models.Post) error {
	return nil
}

func (m *mockReportsCache) GetPosts(key string) ([]models.Post, bool, error) {
	return nil, false, nil
}

func (m *mockReportsCache) Delete(key string) error {
	m.deleted = append(m.deleted, key)
	return nil
}

func (m *mockReportsCache) DeletePrefix(prefix string) error {
	m.deletedPrefix = append(m.deletedPrefix, prefix)
	return nil
}

func setupReportsTest(t *testing.T, cache *mockReportsCache) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()

	if cache == nil {
		cache = &mockReportsCache{}
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)
	service := NewService(repo, cache)

	return service, mock, func() {
		sqlxDB.Close()
		db.Close()
	}
}

func targetRows(event bool, authorID int) *sqlmock.Rows {
	var eventDate any
	if event {
		eventDate = time.Now()
	}

	return sqlmock.NewRows([]string{
		"id",
		"author_id",
		"faculty_id",
		"title",
		"content",
		"image_url",
		"is_pinned",
		"scope",
		"status",
		"approved_by",
		"approved_at",
		"rejection_reason",
		"event_date",
		"location",
		"capacity",
		"current_count",
		"created_at",
		"updated_at",
	}).AddRow(
		1,
		authorID,
		1,
		"Target title",
		"Target content",
		nil,
		false,
		"faculty",
		"approved",
		nil,
		nil,
		nil,
		eventDate,
		nil,
		0,
		0,
		time.Now(),
		time.Now(),
	)
}

func reportRows(status string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"reporter_id",
		"target_post_id",
		"target_type",
		"reason",
		"status",
		"review_note",
		"reviewed_by",
		"reviewed_at",
		"created_at",
		"updated_at",
		"target_title",
		"target_author_id",
	}).AddRow(
		1,
		1,
		1,
		models.ReportTargetPost,
		"spam",
		status,
		nil,
		nil,
		nil,
		time.Now(),
		time.Now(),
		"Target title",
		2,
	)
}

func TestCreateReportOnPost(t *testing.T) {
	service, mock, cleanup := setupReportsTest(t, nil)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, author_id`).
		WithArgs(1).
		WillReturnRows(targetRows(false, 2))

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(`INSERT INTO reports`).
		WithArgs(1, 1, models.ReportTargetPost, "spam", models.ReportStatusPending).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at",
		}).AddRow(1, time.Now(), time.Now()))

	report, err := service.Create(1, models.CreateReportInput{
		TargetID:   1,
		TargetType: models.ReportTargetPost,
		Reason:     "spam",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if report.Status != models.ReportStatusPending {
		t.Fatal("expected pending report")
	}
}

func TestCreateReportOnEvent(t *testing.T) {
	service, mock, cleanup := setupReportsTest(t, nil)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, author_id`).
		WithArgs(1).
		WillReturnRows(targetRows(true, 2))

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(`INSERT INTO reports`).
		WithArgs(1, 1, models.ReportTargetEvent, "fake event", models.ReportStatusPending).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at",
		}).AddRow(1, time.Now(), time.Now()))

	_, err := service.Create(1, models.CreateReportInput{
		TargetID:   1,
		TargetType: models.ReportTargetEvent,
		Reason:     "fake event",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCreateTargetNotFound(t *testing.T) {
	service, mock, cleanup := setupReportsTest(t, nil)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, author_id`).
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	_, err := service.Create(1, models.CreateReportInput{
		TargetID:   99,
		TargetType: models.ReportTargetPost,
		Reason:     "spam",
	})

	if !errors.Is(err, ErrTargetNotFound) {
		t.Fatalf("expected ErrTargetNotFound, got %v", err)
	}
}

func TestCreateSelfReportForbidden(t *testing.T) {
	service, mock, cleanup := setupReportsTest(t, nil)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, author_id`).
		WithArgs(1).
		WillReturnRows(targetRows(false, 1))

	_, err := service.Create(1, models.CreateReportInput{
		TargetID:   1,
		TargetType: models.ReportTargetPost,
		Reason:     "spam",
	})

	if !errors.Is(err, ErrSelfReport) {
		t.Fatalf("expected ErrSelfReport, got %v", err)
	}
}

func TestCreateDuplicatePendingReportForbidden(t *testing.T) {
	service, mock, cleanup := setupReportsTest(t, nil)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, author_id`).
		WithArgs(1).
		WillReturnRows(targetRows(false, 2))

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	_, err := service.Create(1, models.CreateReportInput{
		TargetID:   1,
		TargetType: models.ReportTargetPost,
		Reason:     "spam",
	})

	if !errors.Is(err, ErrDuplicatePending) {
		t.Fatalf("expected ErrDuplicatePending, got %v", err)
	}
}

func TestCreateAfterClosedRejectedCanCreateNewPending(t *testing.T) {
	service, mock, cleanup := setupReportsTest(t, nil)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, author_id`).
		WithArgs(1).
		WillReturnRows(targetRows(false, 2))

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.ExpectQuery(`INSERT INTO reports`).
		WithArgs(1, 1, models.ReportTargetPost, "spam", models.ReportStatusPending).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at",
		}).AddRow(1, time.Now(), time.Now()))

	_, err := service.Create(1, models.CreateReportInput{
		TargetID:   1,
		TargetType: models.ReportTargetPost,
		Reason:     "spam",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestListDefaultClosedRejectedAndInvalidFallback(t *testing.T) {
	tests := []struct {
		inputStatus string
		dbStatus    string
	}{
		{"", models.ReportStatusPending},
		{models.ReportStatusClosed, models.ReportStatusClosed},
		{models.ReportStatusRejected, models.ReportStatusRejected},
		{"invalid", models.ReportStatusPending},
	}

	for _, tt := range tests {
		t.Run(tt.inputStatus, func(t *testing.T) {
			service, mock, cleanup := setupReportsTest(t, nil)
			defer cleanup()

			mock.ExpectQuery(`SELECT`).
				WithArgs(tt.dbStatus).
				WillReturnRows(reportRows(tt.dbStatus))

			reports, err := service.List(tt.inputStatus)

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if len(reports) != 1 {
				t.Fatal("expected 1 report")
			}

			if reports[0].Status != tt.dbStatus {
				t.Fatalf("expected %s, got %s", tt.dbStatus, reports[0].Status)
			}
		})
	}
}

func TestCloseStatusClosedDeletesTargetContent(t *testing.T) {
	service, mock, cleanup := setupReportsTest(t, &mockReportsCache{})
	defer cleanup()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT target_post_id, target_type`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"target_post_id", "target_type",
		}).AddRow(1, models.ReportTargetPost))

	mock.ExpectExec(`UPDATE reports`).
		WithArgs(1, models.ReportStatusClosed, "deleted", 99).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`DELETE FROM posts`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := service.Close(1, 99, models.CloseReportInput{
		Status:     models.ReportStatusClosed,
		ReviewNote: "deleted",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCloseStatusRejectedDoesNotDeleteTargetContent(t *testing.T) {
	service, mock, cleanup := setupReportsTest(t, &mockReportsCache{})
	defer cleanup()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT target_post_id, target_type`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"target_post_id", "target_type",
		}).AddRow(1, models.ReportTargetPost))

	mock.ExpectExec(`UPDATE reports`).
		WithArgs(1, models.ReportStatusRejected, "not violation", 99).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := service.Close(1, 99, models.CloseReportInput{
		Status:     models.ReportStatusRejected,
		ReviewNote: "not violation",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCloseReviewNoteCanBeEmpty(t *testing.T) {
	service, mock, cleanup := setupReportsTest(t, &mockReportsCache{})
	defer cleanup()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT target_post_id, target_type`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"target_post_id", "target_type",
		}).AddRow(1, models.ReportTargetPost))

	mock.ExpectExec(`UPDATE reports`).
		WithArgs(1, models.ReportStatusRejected, "", 99).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := service.Close(1, 99, models.CloseReportInput{
		Status:     models.ReportStatusRejected,
		ReviewNote: "",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCloseOnlyPendingReportCanBeClosed(t *testing.T) {
	service, mock, cleanup := setupReportsTest(t, &mockReportsCache{})
	defer cleanup()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT target_post_id, target_type`).
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectRollback()

	err := service.Close(1, 99, models.CloseReportInput{
		Status: models.ReportStatusClosed,
	})

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestCloseCacheInvalidationForDeletedPostAndEvent(t *testing.T) {
	tests := []string{
		models.ReportTargetPost,
		models.ReportTargetEvent,
	}

	for _, targetType := range tests {
		t.Run(targetType, func(t *testing.T) {
			cacheMock := &mockReportsCache{}
			service, mock, cleanup := setupReportsTest(t, cacheMock)
			defer cleanup()

			mock.ExpectBegin()

			mock.ExpectQuery(`SELECT target_post_id, target_type`).
				WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{
					"target_post_id", "target_type",
				}).AddRow(1, targetType))

			mock.ExpectExec(`UPDATE reports`).
				WithArgs(1, models.ReportStatusClosed, "", 99).
				WillReturnResult(sqlmock.NewResult(0, 1))

			mock.ExpectExec(`DELETE FROM posts`).
				WithArgs(1).
				WillReturnResult(sqlmock.NewResult(0, 1))

			mock.ExpectCommit()

			err := service.Close(1, 99, models.CloseReportInput{
				Status:     models.ReportStatusClosed,
				ReviewNote: "",
			})

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if len(cacheMock.deleted) == 0 || len(cacheMock.deletedPrefix) == 0 {
				t.Fatal("expected cache invalidation")
			}
		})
	}
}
