package posts

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kbtuspace-backend/internal/models"
	"kbtuspace-backend/pkg/cache"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type mockPostsCache struct {
	post          *models.Post
	postHit       bool
	posts         []models.Post
	postsHit      bool
	deleted       []string
	deletedPrefix []string
	setPostCalled bool
	setListCalled bool
}

func (m *mockPostsCache) SetPost(key string, value *models.Post) error {
	m.setPostCalled = true
	m.post = value
	return nil
}

func (m *mockPostsCache) GetPost(key string) (*models.Post, bool, error) {
	return m.post, m.postHit, nil
}

func (m *mockPostsCache) SetPosts(key string, value []models.Post) error {
	m.setListCalled = true
	m.posts = value
	return nil
}

func (m *mockPostsCache) GetPosts(key string) ([]models.Post, bool, error) {
	return m.posts, m.postsHit, nil
}

func (m *mockPostsCache) Delete(key string) error {
	m.deleted = append(m.deleted, key)
	return nil
}

func (m *mockPostsCache) DeletePrefix(prefix string) error {
	m.deletedPrefix = append(m.deletedPrefix, prefix)
	return nil
}

func setupPostsTest(t *testing.T, c cache.PostsCache) (*Service, sqlmock.Sqlmock, func()) {
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

func strPtr(v string) *string {
	return &v
}

func postRows() *sqlmock.Rows {
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

func addPostRow(rows *sqlmock.Rows, id int, authorID int, facultyID any, pinned bool, scope string, status string, createdAt time.Time) *sqlmock.Rows {
	return rows.AddRow(
		id,
		authorID,
		"author@kbtu.kz",
		facultyID,
		"Title",
		"Content text",
		nil,
		pinned,
		scope,
		status,
		nil,
		nil,
		nil,
		nil,
		nil,
		0,
		0,
		createdAt,
		createdAt,
	)
}

func TestResolveModerationFacultyPostDefault(t *testing.T) {
	actorFacultyID := 1

	facultyID, scope, status, _, _, _, err := resolveModeration(
		"student",
		&actorFacultyID,
		nil,
		"",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if scope != models.ContentScopeFaculty {
		t.Fatalf("expected faculty scope, got %s", scope)
	}

	if status != models.ContentStatusApproved {
		t.Fatalf("expected approved status, got %s", status)
	}

	if facultyID == nil || *facultyID != actorFacultyID {
		t.Fatal("expected actor faculty id")
	}
}

func TestResolveModerationStudentOrganizerFacultyApproved(t *testing.T) {
	roles := []string{"student", "organizer"}

	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			actorFacultyID := 1

			_, scope, status, _, _, _, err := resolveModeration(
				role,
				&actorFacultyID,
				nil,
				models.ContentScopeFaculty,
			)

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if scope != models.ContentScopeFaculty {
				t.Fatalf("expected faculty scope, got %s", scope)
			}

			if status != models.ContentStatusApproved {
				t.Fatalf("expected approved, got %s", status)
			}
		})
	}
}

func TestResolveModerationActorWithoutFacultyCannotCreateFacultyPost(t *testing.T) {
	_, _, _, _, _, _, err := resolveModeration(
		"student",
		nil,
		nil,
		models.ContentScopeFaculty,
	)

	if !errors.Is(err, ErrFacultyRequired) {
		t.Fatalf("expected ErrFacultyRequired, got %v", err)
	}
}

func TestResolveModerationGlobalNonAdminPending(t *testing.T) {
	actorFacultyID := 1

	facultyID, scope, status, approvedBy, approvedAt, _, err := resolveModeration(
		"student",
		&actorFacultyID,
		nil,
		models.ContentScopeGlobal,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if facultyID != nil {
		t.Fatal("expected nil faculty for global post")
	}

	if scope != models.ContentScopeGlobal {
		t.Fatalf("expected global scope, got %s", scope)
	}

	if status != models.ContentStatusPending {
		t.Fatalf("expected pending status, got %s", status)
	}

	if approvedBy != nil || approvedAt != nil {
		t.Fatal("non-admin global post must not be approved")
	}
}

func TestResolveModerationGlobalAdminApproved(t *testing.T) {
	facultyID, scope, status, _, approvedAt, _, err := resolveModeration(
		"admin",
		nil,
		nil,
		models.ContentScopeGlobal,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if facultyID != nil {
		t.Fatal("expected nil faculty for global post")
	}

	if scope != models.ContentScopeGlobal {
		t.Fatalf("expected global scope, got %s", scope)
	}

	if status != models.ContentStatusApproved {
		t.Fatalf("expected approved status, got %s", status)
	}

	if approvedAt == nil {
		t.Fatal("expected approved_at")
	}
}

func TestCreateFacultyPost(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	actorFacultyID := 1

	mock.ExpectQuery(`INSERT INTO posts`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, time.Now(), time.Now()))

	post, err := service.Create(10, "student", &actorFacultyID, models.CreatePostInput{
		Title:   "Title",
		Content: "Content text",
		Scope:   models.ContentScopeFaculty,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if post.ID != 1 {
		t.Fatalf("expected post id 1, got %d", post.ID)
	}

	if post.Status != models.ContentStatusApproved {
		t.Fatalf("expected approved, got %s", post.Status)
	}

	if post.Scope != models.ContentScopeFaculty {
		t.Fatalf("expected faculty, got %s", post.Scope)
	}
}

func TestCreateGlobalPendingPost(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	actorFacultyID := 1

	mock.ExpectQuery(`INSERT INTO posts`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, time.Now(), time.Now()))

	post, err := service.Create(10, "organizer", &actorFacultyID, models.CreatePostInput{
		Title:   "Title",
		Content: "Content text",
		Scope:   models.ContentScopeGlobal,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if post.Status != models.ContentStatusPending {
		t.Fatalf("expected pending, got %s", post.Status)
	}

	if post.Scope != models.ContentScopeGlobal {
		t.Fatalf("expected global, got %s", post.Scope)
	}
}

func TestGetAllFacultyFeedReturnsApprovedOwnFaculty(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	facultyID := 1
	now := time.Now()

	rows := postRows()
	addPostRow(rows, 1, 10, facultyID, false, models.ContentScopeFaculty, models.ContentStatusApproved, now)

	mock.ExpectQuery(`SELECT p.id, p.author_id`).
		WithArgs(facultyID).
		WillReturnRows(rows)

	posts, err := service.GetAll(&facultyID, "student", false)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}

	if posts[0].FacultyID == nil || *posts[0].FacultyID != facultyID {
		t.Fatal("expected own faculty post")
	}
}

func TestGetAllGlobalFeedReturnsApprovedGlobalPosts(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	now := time.Now()

	rows := postRows()
	addPostRow(rows, 1, 10, nil, false, models.ContentScopeGlobal, models.ContentStatusApproved, now)

	mock.ExpectQuery(`SELECT p.id, p.author_id`).
		WillReturnRows(rows)

	posts, err := service.GetAll(nil, "student", true)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}

	if posts[0].Scope != models.ContentScopeGlobal {
		t.Fatalf("expected global post, got %s", posts[0].Scope)
	}
}

func TestGetAllSortingPinnedFirstThenCreatedAtDesc(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	facultyID := 1
	newer := time.Now()
	older := newer.Add(-time.Hour)

	rows := postRows()
	addPostRow(rows, 1, 10, facultyID, true, models.ContentScopeFaculty, models.ContentStatusApproved, older)
	addPostRow(rows, 2, 11, facultyID, false, models.ContentScopeFaculty, models.ContentStatusApproved, newer)

	mock.ExpectQuery(`ORDER BY p.is_pinned DESC, p.created_at DESC`).
		WithArgs(facultyID).
		WillReturnRows(rows)

	posts, err := service.GetAll(&facultyID, "student", false)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}

	if !posts[0].IsPinned {
		t.Fatal("expected pinned post first")
	}
}

func TestGetByIDAdminSeesUnapproved(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	now := time.Now()

	rows := postRows()
	addPostRow(rows, 1, 10, nil, false, models.ContentScopeGlobal, models.ContentStatusPending, now)

	mock.ExpectQuery(`SELECT p.id, p.author_id`).
		WithArgs(1).
		WillReturnRows(rows)

	post, err := service.GetByID(1, "admin", nil)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if post.Status != models.ContentStatusPending {
		t.Fatalf("expected pending, got %s", post.Status)
	}
}

func TestGetByIDStudentSeesApprovedGlobal(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	facultyID := 1
	now := time.Now()

	rows := postRows()
	addPostRow(rows, 1, 10, nil, false, models.ContentScopeGlobal, models.ContentStatusApproved, now)

	mock.ExpectQuery(`SELECT p.id, p.author_id`).
		WithArgs(1, &facultyID).
		WillReturnRows(rows)

	post, err := service.GetByID(1, "student", &facultyID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if post.Scope != models.ContentScopeGlobal {
		t.Fatalf("expected global, got %s", post.Scope)
	}
}

func TestGetByIDStudentSeesApprovedFacultyOnlyOwnFaculty(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	facultyID := 1
	now := time.Now()

	rows := postRows()
	addPostRow(rows, 1, 10, facultyID, false, models.ContentScopeFaculty, models.ContentStatusApproved, now)

	mock.ExpectQuery(`SELECT p.id, p.author_id`).
		WithArgs(1, &facultyID).
		WillReturnRows(rows)

	post, err := service.GetByID(1, "student", &facultyID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if post.FacultyID == nil || *post.FacultyID != facultyID {
		t.Fatal("expected own faculty post")
	}
}

func TestGetByIDStudentCannotSeeOtherFacultyPendingRejected(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	facultyID := 1

	mock.ExpectQuery(`SELECT p.id, p.author_id`).
		WithArgs(1, &facultyID).
		WillReturnError(sql.ErrNoRows)

	_, err := service.GetByID(1, "student", &facultyID)

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestUpdateAuthorCanUpdateOwnPost(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	facultyID := 1

	mock.ExpectQuery(`WITH target AS`).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("updated"))

	err := service.Update(1, 10, "student", &facultyID, models.UpdatePostInput{
		Title:   "Updated title",
		Content: "Updated content",
		Scope:   models.ContentScopeFaculty,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUpdateOtherAuthorCannotUpdate(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	facultyID := 1

	mock.ExpectQuery(`WITH target AS`).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("forbidden"))

	err := service.Update(1, 99, "student", &facultyID, models.UpdatePostInput{
		Title:   "Updated title",
		Content: "Updated content",
		Scope:   models.ContentScopeFaculty,
	})

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdateStudentCannotPinThroughUpdate(t *testing.T) {
	service, _, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	facultyID := 1

	err := service.Update(1, 10, "student", &facultyID, models.UpdatePostInput{
		Title:    "Updated title",
		Content:  "Updated content",
		Scope:    models.ContentScopeFaculty,
		IsPinned: true,
	})

	if !errors.Is(err, ErrPinForbidden) {
		t.Fatalf("expected ErrPinForbidden, got %v", err)
	}
}

func TestUpdateGlobalNonAdminBecomesPending(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	facultyID := 1

	mock.ExpectQuery(`WITH target AS`).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("updated"))

	err := service.Update(1, 10, "organizer", &facultyID, models.UpdatePostInput{
		Title:   "Updated title",
		Content: "Updated content",
		Scope:   models.ContentScopeGlobal,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteAuthorCanDelete(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	mock.ExpectQuery(`WITH target AS`).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("deleted"))

	err := service.Delete(1, 10, "student")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteAdminCanDeleteAny(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	mock.ExpectQuery(`WITH target AS`).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("deleted"))

	err := service.Delete(1, 99, "admin")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDeleteOtherUserCannotDelete(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	mock.ExpectQuery(`WITH target AS`).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("forbidden"))

	err := service.Delete(1, 99, "student")

	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestPinOrganizerAdminCanPinApprovedFacultyPost(t *testing.T) {
	tests := []struct {
		role      string
		facultyID *int
	}{
		{"organizer", intPtr(1)},
		{"admin", nil},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			service, mock, cleanup := setupPostsTest(t, nil)
			defer cleanup()

			mock.ExpectQuery(`WITH target AS`).
				WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("updated"))

			err := service.Pin(1, tt.role, tt.facultyID, true)

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestPinStudentCannotPin(t *testing.T) {
	service, _, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	facultyID := 1

	err := service.Pin(1, "student", &facultyID, true)

	if !errors.Is(err, ErrPinForbidden) {
		t.Fatalf("expected ErrPinForbidden, got %v", err)
	}
}

func TestPinOrganizerCannotPinGlobalOrOtherFaculty(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	facultyID := 1

	mock.ExpectQuery(`WITH target AS`).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("forbidden"))

	err := service.Pin(1, "organizer", &facultyID, true)

	if !errors.Is(err, ErrInvalidPinScope) {
		t.Fatalf("expected ErrInvalidPinScope, got %v", err)
	}
}

func TestApproveAdminApprovesOnlyPendingGlobalPost(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	mock.ExpectExec(`UPDATE posts`).
		WithArgs(1, 99).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := service.Approve(1, 99)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	service2, mock2, cleanup2 := setupPostsTest(t, nil)
	defer cleanup2()

	mock2.ExpectExec(`UPDATE posts`).
		WithArgs(2, 99).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = service2.Approve(2, 99)

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestRejectAdminRejectsOnlyPendingGlobalPost(t *testing.T) {
	service, mock, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	mock.ExpectExec(`UPDATE posts`).
		WithArgs(1, "bad content").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := service.Reject(1, "bad content")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	service2, mock2, cleanup2 := setupPostsTest(t, nil)
	defer cleanup2()

	mock2.ExpectExec(`UPDATE posts`).
		WithArgs(2, "bad content").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = service2.Reject(2, "bad content")

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestRejectReasonRequiredHandlerValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service, _, cleanup := setupPostsTest(t, nil)
	defer cleanup()

	handler := NewHandler(service)

	r := gin.New()
	r.PATCH("/posts/:id/reject", handler.Reject)

	req := httptest.NewRequest(http.MethodPatch, "/posts/1/reject", bytes.NewBuffer([]byte(`{"reason":""}`)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCacheCreateUpdateDeletePinApproveRejectInvalidatesKeys(t *testing.T) {
	t.Run("create approved invalidates list", func(t *testing.T) {
		mc := &mockPostsCache{}
		service, mock, cleanup := setupPostsTest(t, mc)
		defer cleanup()

		facultyID := 1

		mock.ExpectQuery(`INSERT INTO posts`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
				AddRow(1, time.Now(), time.Now()))

		_, err := service.Create(10, "student", &facultyID, models.CreatePostInput{
			Title:   "Title",
			Content: "Content text",
			Scope:   models.ContentScopeFaculty,
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(mc.deletedPrefix) == 0 {
			t.Fatal("expected list cache invalidation")
		}
	})

	t.Run("update invalidates item and list", func(t *testing.T) {
		mc := &mockPostsCache{}
		service, mock, cleanup := setupPostsTest(t, mc)
		defer cleanup()

		facultyID := 1

		mock.ExpectQuery(`WITH target AS`).
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("updated"))

		err := service.Update(1, 10, "student", &facultyID, models.UpdatePostInput{
			Title:   "Updated title",
			Content: "Updated content",
			Scope:   models.ContentScopeFaculty,
		})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(mc.deleted) == 0 || len(mc.deletedPrefix) == 0 {
			t.Fatal("expected item and list cache invalidation")
		}
	})

	t.Run("delete invalidates item and list", func(t *testing.T) {
		mc := &mockPostsCache{}
		service, mock, cleanup := setupPostsTest(t, mc)
		defer cleanup()

		mock.ExpectQuery(`WITH target AS`).
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("deleted"))

		err := service.Delete(1, 10, "student")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(mc.deleted) == 0 || len(mc.deletedPrefix) == 0 {
			t.Fatal("expected item and list cache invalidation")
		}
	})

	t.Run("pin invalidates item and list", func(t *testing.T) {
		mc := &mockPostsCache{}
		service, mock, cleanup := setupPostsTest(t, mc)
		defer cleanup()

		facultyID := 1

		mock.ExpectQuery(`WITH target AS`).
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("updated"))

		err := service.Pin(1, "organizer", &facultyID, true)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(mc.deleted) == 0 || len(mc.deletedPrefix) == 0 {
			t.Fatal("expected item and list cache invalidation")
		}
	})

	t.Run("approve invalidates item and list", func(t *testing.T) {
		mc := &mockPostsCache{}
		service, mock, cleanup := setupPostsTest(t, mc)
		defer cleanup()

		mock.ExpectExec(`UPDATE posts`).
			WithArgs(1, 99).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := service.Approve(1, 99)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(mc.deleted) == 0 || len(mc.deletedPrefix) == 0 {
			t.Fatal("expected item and list cache invalidation")
		}
	})

	t.Run("reject invalidates item and list", func(t *testing.T) {
		mc := &mockPostsCache{}
		service, mock, cleanup := setupPostsTest(t, mc)
		defer cleanup()

		mock.ExpectExec(`UPDATE posts`).
			WithArgs(1, "bad content").
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := service.Reject(1, "bad content")

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(mc.deleted) == 0 || len(mc.deletedPrefix) == 0 {
			t.Fatal("expected item and list cache invalidation")
		}
	})
}

func TestCacheGetByIDUsesCacheOnlyIfAccessAllowed(t *testing.T) {
	t.Run("allowed cached post returned", func(t *testing.T) {
		facultyID := 1

		mc := &mockPostsCache{
			postHit: true,
			post: &models.Post{
				ID:        1,
				FacultyID: &facultyID,
				Scope:     models.ContentScopeFaculty,
				Status:    models.ContentStatusApproved,
			},
		}

		service, _, cleanup := setupPostsTest(t, mc)
		defer cleanup()

		post, err := service.GetByID(1, "student", &facultyID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if post.ID != 1 {
			t.Fatalf("expected cached post id 1, got %d", post.ID)
		}
	})

	t.Run("forbidden cached post ignored and repo used", func(t *testing.T) {
		studentFacultyID := 1
		otherFacultyID := 2
		now := time.Now()

		mc := &mockPostsCache{
			postHit: true,
			post: &models.Post{
				ID:        1,
				FacultyID: &otherFacultyID,
				Scope:     models.ContentScopeFaculty,
				Status:    models.ContentStatusApproved,
			},
		}

		service, mock, cleanup := setupPostsTest(t, mc)
		defer cleanup()

		rows := postRows()
		addPostRow(rows, 1, 10, studentFacultyID, false, models.ContentScopeFaculty, models.ContentStatusApproved, now)

		mock.ExpectQuery(`SELECT p.id, p.author_id`).
			WithArgs(1, &studentFacultyID).
			WillReturnRows(rows)

		post, err := service.GetByID(1, "student", &studentFacultyID)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if post.FacultyID == nil || *post.FacultyID != studentFacultyID {
			t.Fatal("expected repo post from allowed faculty")
		}
	})
}

//
