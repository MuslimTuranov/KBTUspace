package events

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kbtuspace-backend/internal/middleware"
	"kbtuspace-backend/internal/models"
	"kbtuspace-backend/pkg/cache"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type mockEventsCache struct {
	post          *models.Post
	postHit       bool
	posts         []models.Post
	postsHit      bool
	deleted       []string
	deletedPrefix []string
}

func (m *mockEventsCache) SetPost(key string, value *models.Post) error {
	m.post = value
	return nil
}

func (m *mockEventsCache) GetPost(key string) (*models.Post, bool, error) {
	return m.post, m.postHit, nil
}

func (m *mockEventsCache) SetPosts(key string, value []models.Post) error {
	m.posts = value
	return nil
}

func (m *mockEventsCache) GetPosts(key string) ([]models.Post, bool, error) {
	return m.posts, m.postsHit, nil
}

func (m *mockEventsCache) Delete(key string) error {
	m.deleted = append(m.deleted, key)
	return nil
}

func (m *mockEventsCache) DeletePrefix(prefix string) error {
	m.deletedPrefix = append(m.deletedPrefix, prefix)
	return nil
}

func setupEventsTest(t *testing.T, c cache.PostsCache) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)
	service := NewService(repo, c)

	return service, mock, func() {
		sqlxDB.Close()
		db.Close()
	}
}

func intPtr(v int) *int {
	return &v
}

func futureDate() string {
	return time.Now().Add(48 * time.Hour).Format(time.RFC3339)
}

func eventRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id",
		"author_id",
		"author_email",
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
	})
}

func addEventRow(rows *sqlmock.Rows, id int, authorID int, facultyID any, scope string, status string, eventDate time.Time, currentCount int) *sqlmock.Rows {
	return rows.AddRow(
		id,
		authorID,
		"author@kbtu.kz",
		facultyID,
		"Event title",
		"Event description",
		nil,
		false,
		scope,
		status,
		nil,
		nil,
		nil,
		eventDate,
		"Auditorium",
		100,
		currentCount,
		time.Now(),
		time.Now(),
	)
}

func validCreateEventInput(scope string) models.CreateEventInput {
	return models.CreateEventInput{
		Title:       "Event title",
		Description: "Event description",
		EventDate:   futureDate(),
		Location:    "Auditorium",
		Capacity:    100,
		Scope:       scope,
	}
}

func validUpdateEventInput(scope string) models.UpdateEventInput {
	return models.UpdateEventInput{
		Title:       "Event title",
		Description: "Event description",
		EventDate:   futureDate(),
		Location:    "Auditorium",
		Capacity:    100,
		Scope:       scope,
	}
}

func TestParseEventDateRFC3339YYYYMMDDDDMMYYYY(t *testing.T) {
	tests := []string{
		"2026-05-12T10:00:00Z",
		"2026-05-12",
		"12.05.2026",
	}

	for _, value := range tests {
		_, err := parseEventDate(value)
		if err != nil {
			t.Fatalf("expected valid date %s, got %v", value, err)
		}
	}
}

func TestParseEventDateEmptyInvalid(t *testing.T) {
	tests := []string{"", "abc", "2026/05/12"}

	for _, value := range tests {
		_, err := parseEventDate(value)
		if !errors.Is(err, ErrInvalidEventDate) {
			t.Fatalf("expected ErrInvalidEventDate, got %v", err)
		}
	}
}

func TestCreateOrganizerAdminCreatesFacultyEvent(t *testing.T) {
	tests := []struct {
		role      string
		facultyID *int
		input     models.CreateEventInput
	}{
		{
			role:      "organizer",
			facultyID: intPtr(1),
			input:     validCreateEventInput(models.ContentScopeFaculty),
		},
		{
			role:      "admin",
			facultyID: nil,
			input: func() models.CreateEventInput {
				input := validCreateEventInput(models.ContentScopeFaculty)
				input.FacultyID = intPtr(1)
				return input
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			service, mock, cleanup := setupEventsTest(t, nil)
			defer cleanup()

			mock.ExpectQuery(`INSERT INTO posts`).
				WillReturnRows(sqlmock.NewRows([]string{
					"id", "created_at", "updated_at",
				}).AddRow(1, time.Now(), time.Now()))

			event, err := service.Create(10, tt.role, tt.facultyID, tt.input)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if event.Scope != models.ContentScopeFaculty {
				t.Fatalf("expected faculty scope, got %s", event.Scope)
			}

			if event.Status != models.ContentStatusApproved {
				t.Fatalf("expected approved, got %s", event.Status)
			}
		})
	}
}

func TestCreateStudentCannotCreateEventThroughProtectedOrganizerRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.POST("/events", func(c *gin.Context) {
		c.Set("role", "student")
		c.Next()
	}, middleware.RequireRole("organizer", "admin"), func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer([]byte(`{}`)))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestCreatePastDateRejected(t *testing.T) {
	service, _, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	facultyID := 1
	input := validCreateEventInput(models.ContentScopeFaculty)
	input.EventDate = time.Now().Add(-time.Hour).Format(time.RFC3339)

	_, err := service.Create(10, "organizer", &facultyID, input)

	if !errors.Is(err, ErrEventDateInPast) {
		t.Fatalf("expected ErrEventDateInPast, got %v", err)
	}
}

func TestCreateCapacityMinMaxHandlerValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service, _, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	handler := NewHandler(service)

	r := gin.New()
	r.POST("/events", func(c *gin.Context) {
		c.Set("userID", 1)
		c.Set("role", "organizer")
		c.Set("facultyID", 1)
		c.Next()
	}, handler.Create)

	bodies := [][]byte{
		[]byte(`{"title":"Event title","description":"Event description","event_date":"` + futureDate() + `","location":"Auditorium","capacity":0,"scope":"faculty"}`),
		[]byte(`{"title":"Event title","description":"Event description","event_date":"` + futureDate() + `","location":"Auditorium","capacity":10001,"scope":"faculty"}`),
	}

	for _, body := range bodies {
		req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	}
}

func TestCreateGlobalOrganizerPending(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	facultyID := 1

	mock.ExpectQuery(`INSERT INTO posts`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at",
		}).AddRow(1, time.Now(), time.Now()))

	event, err := service.Create(10, "organizer", &facultyID, validCreateEventInput(models.ContentScopeGlobal))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if event.Status != models.ContentStatusPending {
		t.Fatalf("expected pending, got %s", event.Status)
	}

	if event.Scope != models.ContentScopeGlobal {
		t.Fatalf("expected global, got %s", event.Scope)
	}
}

func TestCreateGlobalAdminApproved(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	mock.ExpectQuery(`INSERT INTO posts`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at",
		}).AddRow(1, time.Now(), time.Now()))

	event, err := service.Create(99, "admin", nil, validCreateEventInput(models.ContentScopeGlobal))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if event.Status != models.ContentStatusApproved {
		t.Fatalf("expected approved, got %s", event.Status)
	}

	if event.ApprovedBy == nil || *event.ApprovedBy != 99 {
		t.Fatal("expected approved_by admin id")
	}
}

func TestGetAllFacultyGlobalFilters(t *testing.T) {
	t.Run("faculty", func(t *testing.T) {
		service, mock, cleanup := setupEventsTest(t, nil)
		defer cleanup()

		facultyID := 1
		rows := eventRows()
		addEventRow(rows, 1, 10, facultyID, models.ContentScopeFaculty, models.ContentStatusApproved, time.Now().Add(24*time.Hour), 0)

		mock.ExpectQuery(`WHERE event_date IS NOT NULL`).
			WithArgs(sqlmock.AnyArg(), facultyID).
			WillReturnRows(rows)

		events, err := service.GetAll(&facultyID, "student", false)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		if events[0].Scope != models.ContentScopeFaculty {
			t.Fatalf("expected faculty, got %s", events[0].Scope)
		}
	})

	t.Run("global", func(t *testing.T) {
		service, mock, cleanup := setupEventsTest(t, nil)
		defer cleanup()

		rows := eventRows()
		addEventRow(rows, 1, 10, nil, models.ContentScopeGlobal, models.ContentStatusApproved, time.Now().Add(24*time.Hour), 0)

		mock.ExpectQuery(`AND scope = 'global'`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(rows)

		events, err := service.GetAll(nil, "student", true)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(events) != 1 {
			t.Fatalf("expected 1 event, got %d", len(events))
		}

		if events[0].Scope != models.ContentScopeGlobal {
			t.Fatalf("expected global, got %s", events[0].Scope)
		}
	})
}

func TestGetAllShowsEventsNotOlderThan30Days(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	facultyID := 1
	rows := eventRows()
	addEventRow(rows, 1, 10, facultyID, models.ContentScopeFaculty, models.ContentStatusApproved, time.Now().AddDate(0, 0, -29), 0)

	mock.ExpectQuery(`AND event_date >= \$1`).
		WithArgs(sqlmock.AnyArg(), facultyID).
		WillReturnRows(rows)

	events, err := service.GetAll(&facultyID, "student", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
}

func TestGetByIDReturnsIsRegisteredForCurrentUser(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	facultyID := 1
	rows := eventRows()
	addEventRow(rows, 1, 10, facultyID, models.ContentScopeFaculty, models.ContentStatusApproved, time.Now().Add(24*time.Hour), 0)

	mock.ExpectQuery(`SELECT p.id, p.author_id`).
		WithArgs(1).
		WillReturnRows(rows)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	event, err := service.GetByID(1, 5, "student", &facultyID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if event.IsRegistered == nil || *event.IsRegistered != true {
		t.Fatal("expected is_registered true")
	}
}

func TestGetByIDCacheDoesNotStoreUserSpecificIsRegistered(t *testing.T) {
	cacheMock := &mockEventsCache{}

	service, mock, cleanup := setupEventsTest(t, cacheMock)
	defer cleanup()

	facultyID := 1
	rows := eventRows()
	addEventRow(rows, 1, 10, facultyID, models.ContentScopeFaculty, models.ContentStatusApproved, time.Now().Add(24*time.Hour), 0)

	mock.ExpectQuery(`SELECT p.id, p.author_id`).
		WithArgs(1).
		WillReturnRows(rows)

	mock.ExpectQuery(`SELECT EXISTS`).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	event, err := service.GetByID(1, 5, "student", &facultyID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if event.IsRegistered == nil || *event.IsRegistered != true {
		t.Fatal("expected is_registered true")
	}

	if cacheMock.post == nil {
		t.Fatal("expected event saved to cache")
	}

	if cacheMock.post.IsRegistered != nil {
		t.Fatal("cache must not store user-specific is_registered")
	}
}

func TestUpdateOwnerFacultyOrganizerAdminCanUpdate(t *testing.T) {
	tests := []struct {
		name      string
		actorID   int
		role      string
		facultyID *int
	}{
		{"owner", 10, "organizer", intPtr(1)},
		{"faculty organizer", 20, "organizer", intPtr(1)},
		{"admin", 99, "admin", intPtr(1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mock, cleanup := setupEventsTest(t, nil)
			defer cleanup()

			mock.ExpectQuery(`WITH target AS`).
				WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("updated"))

			err := service.Update(1, tt.actorID, tt.role, tt.facultyID, validUpdateEventInput(models.ContentScopeFaculty))
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestUpdateOtherFacultyForbidden(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	facultyID := 1

	mock.ExpectQuery(`WITH target AS`).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("forbidden"))

	err := service.Update(1, 20, "organizer", &facultyID, validUpdateEventInput(models.ContentScopeFaculty))

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdatePastDateForbidden(t *testing.T) {
	service, _, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	facultyID := 1
	input := validUpdateEventInput(models.ContentScopeFaculty)
	input.EventDate = time.Now().Add(-time.Hour).Format(time.RFC3339)

	err := service.Update(1, 10, "organizer", &facultyID, input)

	if !errors.Is(err, ErrEventDateInPast) {
		t.Fatalf("expected ErrEventDateInPast, got %v", err)
	}
}

func TestDeleteOwnerFacultyOrganizerAdminCanDelete(t *testing.T) {
	tests := []struct {
		name      string
		actorID   int
		role      string
		facultyID *int
	}{
		{"owner", 10, "organizer", intPtr(1)},
		{"faculty organizer", 20, "organizer", intPtr(1)},
		{"admin", 99, "admin", intPtr(1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mock, cleanup := setupEventsTest(t, nil)
			defer cleanup()

			mock.ExpectQuery(`WITH target AS`).
				WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("deleted"))

			err := service.Delete(1, tt.actorID, tt.role, tt.facultyID)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestRegisterSuccessIncreasesCurrentCount(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	facultyID := 1

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT capacity, current_count, scope, faculty_id`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"capacity", "current_count", "scope", "faculty_id",
		}).AddRow(10, 0, models.ContentScopeFaculty, facultyID))

	mock.ExpectQuery(`SELECT status FROM registrations`).
		WithArgs(5, 1).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO registrations`).
		WithArgs(5, 1, models.RegistrationStatusRegistered).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`UPDATE posts SET current_count = current_count \+ 1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := service.Register(5, 1, "student", &facultyID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRegisterEventNotFound(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT capacity, current_count, scope, faculty_id`).
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectRollback()

	err := service.Register(5, 1, "student", intPtr(1))

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestRegisterAlreadyRegistered(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	facultyID := 1

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT capacity, current_count, scope, faculty_id`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"capacity", "current_count", "scope", "faculty_id",
		}).AddRow(10, 0, models.ContentScopeFaculty, facultyID))

	mock.ExpectQuery(`SELECT status FROM registrations`).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(models.RegistrationStatusRegistered))

	mock.ExpectRollback()

	err := service.Register(5, 1, "student", &facultyID)

	if !errors.Is(err, ErrAlreadyRegistered) {
		t.Fatalf("expected ErrAlreadyRegistered, got %v", err)
	}
}

func TestRegisterCancelledCanBeActivatedAgain(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	facultyID := 1

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT capacity, current_count, scope, faculty_id`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"capacity", "current_count", "scope", "faculty_id",
		}).AddRow(10, 0, models.ContentScopeFaculty, facultyID))

	mock.ExpectQuery(`SELECT status FROM registrations`).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(models.RegistrationStatusCancelled))

	mock.ExpectExec(`UPDATE registrations SET status`).
		WithArgs(models.RegistrationStatusRegistered, 5, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`UPDATE posts SET current_count = current_count \+ 1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := service.Register(5, 1, "student", &facultyID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRegisterFullEventConflict(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	facultyID := 1

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT capacity, current_count, scope, faculty_id`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"capacity", "current_count", "scope", "faculty_id",
		}).AddRow(10, 10, models.ContentScopeFaculty, facultyID))

	mock.ExpectRollback()

	err := service.Register(5, 1, "student", &facultyID)

	if !errors.Is(err, ErrEventFull) {
		t.Fatalf("expected ErrEventFull, got %v", err)
	}
}

func TestRegisterFacultyEventForbidsCrossFacultyStudent(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	eventFacultyID := 2
	studentFacultyID := 1

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT capacity, current_count, scope, faculty_id`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"capacity", "current_count", "scope", "faculty_id",
		}).AddRow(10, 0, models.ContentScopeFaculty, eventFacultyID))

	mock.ExpectRollback()

	err := service.Register(5, 1, "student", &studentFacultyID)

	if !errors.Is(err, ErrCrossFacultyAccess) {
		t.Fatalf("expected ErrCrossFacultyAccess, got %v", err)
	}
}

func TestRegisterAdminWithoutFacultyRestriction(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	eventFacultyID := 2

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT capacity, current_count, scope, faculty_id`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"capacity", "current_count", "scope", "faculty_id",
		}).AddRow(10, 0, models.ContentScopeFaculty, eventFacultyID))

	mock.ExpectQuery(`SELECT status FROM registrations`).
		WithArgs(5, 1).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(`INSERT INTO registrations`).
		WithArgs(5, 1, models.RegistrationStatusRegistered).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`UPDATE posts SET current_count = current_count \+ 1`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := service.Register(5, 1, "admin", nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCancelRegistrationSuccessDecrementsNotBelowZero(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT 1 FROM posts`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))

	mock.ExpectQuery(`SELECT status FROM registrations`).
		WithArgs(5, 1).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(models.RegistrationStatusRegistered))

	mock.ExpectExec(`UPDATE registrations SET status = 'cancelled'`).
		WithArgs(5, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec(`UPDATE posts SET current_count = GREATEST`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err := service.CancelRegistration(5, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestCancelRegistrationEventNotFound(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT 1 FROM posts`).
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectRollback()

	err := service.CancelRegistration(5, 1)

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestCancelRegistrationNotRegisteredCancelledAttended(t *testing.T) {
	tests := []struct {
		name   string
		status string
		noRow  bool
	}{
		{"not registered", "", true},
		{"cancelled", models.RegistrationStatusCancelled, false},
		{"attended", models.RegistrationStatusAttended, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mock, cleanup := setupEventsTest(t, nil)
			defer cleanup()

			mock.ExpectBegin()

			mock.ExpectQuery(`SELECT 1 FROM posts`).
				WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))

			if tt.noRow {
				mock.ExpectQuery(`SELECT status FROM registrations`).
					WithArgs(5, 1).
					WillReturnError(sql.ErrNoRows)
			} else {
				mock.ExpectQuery(`SELECT status FROM registrations`).
					WithArgs(5, 1).
					WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(tt.status))
			}

			mock.ExpectRollback()

			err := service.CancelRegistration(5, 1)

			if !errors.Is(err, ErrNotRegistered) {
				t.Fatalf("expected ErrNotRegistered, got %v", err)
			}
		})
	}
}

func TestMarkAttendanceAdminOwnerFacultyOrganizer(t *testing.T) {
	tests := []struct {
		name          string
		actorID       int
		role          string
		actorFaculty  *int
		eventAuthorID int
		eventFaculty  any
	}{
		{"admin", 99, "admin", nil, 10, 2},
		{"owner", 10, "organizer", intPtr(1), 10, 2},
		{"same faculty organizer", 20, "organizer", intPtr(2), 10, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mock, cleanup := setupEventsTest(t, nil)
			defer cleanup()

			mock.ExpectQuery(`SELECT author_id, faculty_id`).
				WithArgs(1).
				WillReturnRows(sqlmock.NewRows([]string{
					"author_id", "faculty_id",
				}).AddRow(tt.eventAuthorID, tt.eventFaculty))

			mock.ExpectExec(`UPDATE registrations`).
				WithArgs(5, 1).
				WillReturnResult(sqlmock.NewResult(0, 1))

			err := service.MarkAttended(tt.actorID, tt.role, tt.actorFaculty, 5, 1)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestMarkAttendanceOtherFacultyForbidden(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	mock.ExpectQuery(`SELECT author_id, faculty_id`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"author_id", "faculty_id",
		}).AddRow(10, 2))

	err := service.MarkAttended(20, "organizer", intPtr(1), 5, 1)

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestMarkAttendanceUserNotRegistered(t *testing.T) {
	service, mock, cleanup := setupEventsTest(t, nil)
	defer cleanup()

	mock.ExpectQuery(`SELECT author_id, faculty_id`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"author_id", "faculty_id",
		}).AddRow(10, 1))

	mock.ExpectExec(`UPDATE registrations`).
		WithArgs(5, 1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := service.MarkAttended(99, "admin", nil, 5, 1)

	if !errors.Is(err, ErrNotRegistered) {
		t.Fatalf("expected ErrNotRegistered, got %v", err)
	}
}

func TestApproveRejectOnlyPendingGlobalEvent(t *testing.T) {
	t.Run("approve", func(t *testing.T) {
		service, mock, cleanup := setupEventsTest(t, nil)
		defer cleanup()

		mock.ExpectExec(`UPDATE posts`).
			WithArgs(1, 99).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := service.Approve(1, 99)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		service2, mock2, cleanup2 := setupEventsTest(t, nil)
		defer cleanup2()

		mock2.ExpectExec(`UPDATE posts`).
			WithArgs(2, 99).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err = service2.Approve(2, 99)

		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected sql.ErrNoRows, got %v", err)
		}
	})

	t.Run("reject", func(t *testing.T) {
		service, mock, cleanup := setupEventsTest(t, nil)
		defer cleanup()

		mock.ExpectExec(`UPDATE posts`).
			WithArgs(1, "bad event").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := service.Reject(1, "bad event")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		service2, mock2, cleanup2 := setupEventsTest(t, nil)
		defer cleanup2()

		mock2.ExpectExec(`UPDATE posts`).
			WithArgs(2, "bad event").
			WillReturnResult(sqlmock.NewResult(0, 0))

		err = service2.Reject(2, "bad event")

		if !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("expected sql.ErrNoRows, got %v", err)
		}
	})
}

func TestCacheInvalidationCreateUpdateDeleteRegisterCancelApproveReject(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		cacheMock := &mockEventsCache{}
		service, mock, cleanup := setupEventsTest(t, cacheMock)
		defer cleanup()

		mock.ExpectQuery(`INSERT INTO posts`).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "created_at", "updated_at",
			}).AddRow(1, time.Now(), time.Now()))

		_, err := service.Create(99, "admin", nil, validCreateEventInput(models.ContentScopeGlobal))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(cacheMock.deletedPrefix) == 0 {
			t.Fatal("expected list cache invalidation")
		}
	})

	t.Run("update", func(t *testing.T) {
		cacheMock := &mockEventsCache{}
		service, mock, cleanup := setupEventsTest(t, cacheMock)
		defer cleanup()

		mock.ExpectQuery(`WITH target AS`).
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("updated"))

		err := service.Update(1, 99, "admin", nil, validUpdateEventInput(models.ContentScopeGlobal))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(cacheMock.deleted) == 0 || len(cacheMock.deletedPrefix) == 0 {
			t.Fatal("expected cache invalidation")
		}
	})

	t.Run("delete", func(t *testing.T) {
		cacheMock := &mockEventsCache{}
		service, mock, cleanup := setupEventsTest(t, cacheMock)
		defer cleanup()

		mock.ExpectQuery(`WITH target AS`).
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("deleted"))

		err := service.Delete(1, 99, "admin", nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(cacheMock.deleted) == 0 || len(cacheMock.deletedPrefix) == 0 {
			t.Fatal("expected cache invalidation")
		}
	})

	t.Run("approve", func(t *testing.T) {
		cacheMock := &mockEventsCache{}
		service, mock, cleanup := setupEventsTest(t, cacheMock)
		defer cleanup()

		mock.ExpectExec(`UPDATE posts`).
			WithArgs(1, 99).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := service.Approve(1, 99)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(cacheMock.deleted) == 0 || len(cacheMock.deletedPrefix) == 0 {
			t.Fatal("expected cache invalidation")
		}
	})

	t.Run("reject", func(t *testing.T) {
		cacheMock := &mockEventsCache{}
		service, mock, cleanup := setupEventsTest(t, cacheMock)
		defer cleanup()

		mock.ExpectExec(`UPDATE posts`).
			WithArgs(1, "bad event").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := service.Reject(1, "bad event")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(cacheMock.deleted) == 0 || len(cacheMock.deletedPrefix) == 0 {
			t.Fatal("expected cache invalidation")
		}
	})
}

//
