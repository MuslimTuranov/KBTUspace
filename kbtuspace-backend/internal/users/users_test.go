package users

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kbtuspace-backend/internal/auth"
	"kbtuspace-backend/internal/models"
	"kbtuspace-backend/pkg/hash"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func setupUsersTest(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)
	service := NewService(repo)

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

func boolPtr(v bool) *bool {
	return &v
}

func userRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "email", "password_hash", "role", "faculty_id", "is_banned", "created_at", "updated_at",
	})
}

func TestGetProfileReturnsUser(t *testing.T) {
	service, mock, cleanup := setupUsersTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(userRows().AddRow(
			1,
			"student@kbtu.kz",
			"hash",
			"student",
			2,
			false,
			time.Now(),
			time.Now(),
		))

	user, err := service.GetProfile(1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.ID != 1 {
		t.Fatalf("expected user id 1, got %d", user.ID)
	}

	if user.FacultyID == nil || *user.FacultyID != 2 {
		t.Fatal("expected faculty_id 2")
	}
}

func TestGetProfileAdminFacultyNormalizedNil(t *testing.T) {
	service, mock, cleanup := setupUsersTest(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(userRows().AddRow(
			1,
			"admin@kbtu.kz",
			"hash",
			"admin",
			2,
			false,
			time.Now(),
			time.Now(),
		))

	user, err := service.GetProfile(1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.Role != "admin" {
		t.Fatalf("expected admin role, got %s", user.Role)
	}

	if user.FacultyID != nil {
		t.Fatal("expected admin faculty_id to be nil")
	}
}

func TestUpdateProfileSuccessEmail(t *testing.T) {
	service, mock, cleanup := setupUsersTest(t)
	defer cleanup()

	newEmail := "new@kbtu.kz"

	mock.ExpectQuery(`SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(userRows().AddRow(
			1,
			"old@kbtu.kz",
			"hash",
			"student",
			1,
			false,
			time.Now(),
			time.Now(),
		))

	mock.ExpectQuery(`UPDATE users`).
		WithArgs(1, newEmail).
		WillReturnRows(userRows().AddRow(
			1,
			newEmail,
			"hash",
			"student",
			1,
			false,
			time.Now(),
			time.Now(),
		))

	user, err := service.UpdateProfile(1, models.UpdateProfileInput{
		Email: &newEmail,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.Email != newEmail {
		t.Fatalf("expected email %s, got %s", newEmail, user.Email)
	}
}

func TestUpdateProfileInvalidDomain(t *testing.T) {
	service, _, cleanup := setupUsersTest(t)
	defer cleanup()

	email := "user@gmail.com"

	_, err := service.UpdateProfile(1, models.UpdateProfileInput{
		Email: &email,
	})

	if !errors.Is(err, auth.ErrInvalidEmailDomain) {
		t.Fatalf("expected ErrInvalidEmailDomain, got %v", err)
	}
}

func TestUpdateProfileDuplicateEmail(t *testing.T) {
	service, mock, cleanup := setupUsersTest(t)
	defer cleanup()

	newEmail := "duplicate@kbtu.kz"

	mock.ExpectQuery(`SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(userRows().AddRow(
			1,
			"old@kbtu.kz",
			"hash",
			"student",
			1,
			false,
			time.Now(),
			time.Now(),
		))

	mock.ExpectQuery(`UPDATE users`).
		WithArgs(1, newEmail).
		WillReturnError(&pq.Error{
			Code:    "23505",
			Message: "duplicate key value violates unique constraint users_email_key email",
		})

	_, err := service.UpdateProfile(1, models.UpdateProfileInput{
		Email: &newEmail,
	})

	if !errors.Is(err, auth.ErrDuplicateEmail) {
		t.Fatalf("expected ErrDuplicateEmail, got %v", err)
	}
}

func TestUpdateProfileUserNotFound(t *testing.T) {
	service, mock, cleanup := setupUsersTest(t)
	defer cleanup()

	newEmail := "new@kbtu.kz"

	mock.ExpectQuery(`SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	_, err := service.UpdateProfile(99, models.UpdateProfileInput{
		Email: &newEmail,
	})

	if !errors.Is(err, auth.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestChangePasswordSuccess(t *testing.T) {
	service, mock, cleanup := setupUsersTest(t)
	defer cleanup()

	oldHash, err := hash.HashPassword("oldpassword")
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(userRows().AddRow(
			1,
			"student@kbtu.kz",
			oldHash,
			"student",
			1,
			false,
			time.Now(),
			time.Now(),
		))

	mock.ExpectExec(`UPDATE users`).
		WithArgs(1, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = service.ChangePassword(1, models.ChangePasswordInput{
		CurrentPassword: "oldpassword",
		NewPassword:     "newpassword",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestChangePasswordWrongCurrentPassword(t *testing.T) {
	service, mock, cleanup := setupUsersTest(t)
	defer cleanup()

	oldHash, err := hash.HashPassword("oldpassword")
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(userRows().AddRow(
			1,
			"student@kbtu.kz",
			oldHash,
			"student",
			1,
			false,
			time.Now(),
			time.Now(),
		))

	err = service.ChangePassword(1, models.ChangePasswordInput{
		CurrentPassword: "wrongpassword",
		NewPassword:     "newpassword",
	})

	if !errors.Is(err, auth.ErrInvalidPassword) {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestChangePasswordNewPasswordLessThan6HandlerValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service, _, cleanup := setupUsersTest(t)
	defer cleanup()

	handler := NewHandler(service)

	r := gin.New()
	r.PATCH("/password", func(c *gin.Context) {
		c.Set("userID", 1)
		c.Next()
	}, handler.ChangePassword)

	body := []byte(`{"current_password":"oldpassword","new_password":"123"}`)

	req := httptest.NewRequest(http.MethodPatch, "/password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAdminUpdateUserRoleStudentOrganizerAdmin(t *testing.T) {
	roles := []string{"student", "organizer", "admin"}

	for _, role := range roles {
		t.Run(role, func(t *testing.T) {
			service, mock, cleanup := setupUsersTest(t)
			defer cleanup()

			facultyID := 1

			mock.ExpectQuery(`SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE id = \$1`).
				WithArgs(1).
				WillReturnRows(userRows().AddRow(
					1,
					"user@kbtu.kz",
					"hash",
					"student",
					facultyID,
					false,
					time.Now(),
					time.Now(),
				))

			var updatedFaculty interface{} = facultyID
			if role == "admin" {
				updatedFaculty = nil
			}

			mock.ExpectQuery(`UPDATE users`).
				WithArgs(1, role, updatedFaculty, false).
				WillReturnRows(userRows().AddRow(
					1,
					"user@kbtu.kz",
					"hash",
					role,
					updatedFaculty,
					false,
					time.Now(),
					time.Now(),
				))

			user, err := service.AdminUpdateUser(1, models.AdminUpdateUserInput{
				Role: &role,
			})

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if user.Role != role {
				t.Fatalf("expected role %s, got %s", role, user.Role)
			}

			if role == "admin" && user.FacultyID != nil {
				t.Fatal("expected admin faculty_id nil")
			}
		})
	}
}

func TestAdminUpdateUserAdminRoleClearsFacultyID(t *testing.T) {
	service, mock, cleanup := setupUsersTest(t)
	defer cleanup()

	role := "admin"

	mock.ExpectQuery(`SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE id = \$1`).
		WithArgs(1).
		WillReturnRows(userRows().AddRow(
			1,
			"user@kbtu.kz",
			"hash",
			"student",
			1,
			false,
			time.Now(),
			time.Now(),
		))

	mock.ExpectQuery(`UPDATE users`).
		WithArgs(1, role, nil, false).
		WillReturnRows(userRows().AddRow(
			1,
			"user@kbtu.kz",
			"hash",
			"admin",
			nil,
			false,
			time.Now(),
			time.Now(),
		))

	user, err := service.AdminUpdateUser(1, models.AdminUpdateUserInput{
		Role: &role,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.FacultyID != nil {
		t.Fatal("expected faculty_id nil")
	}
}

func TestAdminUpdateUserFacultyIDLessOrEqualZeroClearsFaculty(t *testing.T) {
	values := []int{0, -1}

	for _, value := range values {
		t.Run("faculty_clear", func(t *testing.T) {
			service, mock, cleanup := setupUsersTest(t)
			defer cleanup()

			mock.ExpectQuery(`SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE id = \$1`).
				WithArgs(1).
				WillReturnRows(userRows().AddRow(
					1,
					"user@kbtu.kz",
					"hash",
					"student",
					1,
					false,
					time.Now(),
					time.Now(),
				))

			mock.ExpectQuery(`UPDATE users`).
				WithArgs(1, "student", nil, false).
				WillReturnRows(userRows().AddRow(
					1,
					"user@kbtu.kz",
					"hash",
					"student",
					nil,
					false,
					time.Now(),
					time.Now(),
				))

			user, err := service.AdminUpdateUser(1, models.AdminUpdateUserInput{
				FacultyID: &value,
			})

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if user.FacultyID != nil {
				t.Fatal("expected faculty_id nil")
			}
		})
	}
}

func TestAdminUpdateUserBanUnban(t *testing.T) {
	tests := []bool{true, false}

	for _, banned := range tests {
		t.Run("ban_unban", func(t *testing.T) {
			service, mock, cleanup := setupUsersTest(t)
			defer cleanup()

			mock.ExpectQuery(`SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE id = \$1`).
				WithArgs(1).
				WillReturnRows(userRows().AddRow(
					1,
					"user@kbtu.kz",
					"hash",
					"student",
					1,
					!banned,
					time.Now(),
					time.Now(),
				))

			mock.ExpectQuery(`UPDATE users`).
				WithArgs(1, "student", 1, banned).
				WillReturnRows(userRows().AddRow(
					1,
					"user@kbtu.kz",
					"hash",
					"student",
					1,
					banned,
					time.Now(),
					time.Now(),
				))

			user, err := service.AdminUpdateUser(1, models.AdminUpdateUserInput{
				IsBanned: &banned,
			})

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if user.IsBanned != banned {
				t.Fatalf("expected banned %v, got %v", banned, user.IsBanned)
			}
		})
	}
}

func TestAdminUpdateUserInvalidRoleHandlerValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service, _, cleanup := setupUsersTest(t)
	defer cleanup()

	handler := NewHandler(service)

	r := gin.New()
	r.PATCH("/admin/users/:id", handler.AdminUpdateUser)

	body := []byte(`{"role":"superadmin"}`)

	req := httptest.NewRequest(http.MethodPatch, "/admin/users/1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

//
