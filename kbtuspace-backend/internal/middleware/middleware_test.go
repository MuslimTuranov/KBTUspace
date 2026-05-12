package middleware

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appjwt "kbtuspace-backend/pkg/jwt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

var testSecret = []byte("12345678901234567890123456789012")

func setupMiddlewareTest(t *testing.T) (*gin.Engine, sqlmock.Sqlmock, func()) {
	t.Helper()

	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	r := gin.New()
	r.GET("/protected", RequireAuth(testSecret, sqlxDB), func(c *gin.Context) {
		userID, _ := c.Get("userID")
		role, _ := c.Get("role")
		facultyID, facultyExists := c.Get("facultyID")

		c.JSON(http.StatusOK, gin.H{
			"userID":        userID,
			"role":          role,
			"facultyID":     facultyID,
			"facultyExists": facultyExists,
		})
	})

	return r, mock, func() {
		sqlxDB.Close()
		db.Close()
	}
}

func generateTestToken(t *testing.T, userID int, role string, facultyID *int) string {
	t.Helper()

	token, err := appjwt.GenerateToken(userID, role, facultyID, testSecret)
	if err != nil {
		t.Fatal(err)
	}

	return token
}

func intPtr(v int) *int {
	return &v
}

func TestRequireAuthNoAuthorization(t *testing.T) {
	r, _, cleanup := setupMiddlewareTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuthInvalidHeaderFormat(t *testing.T) {
	r, _, cleanup := setupMiddlewareTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "WrongFormat token")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuthInvalidToken(t *testing.T) {
	r, _, cleanup := setupMiddlewareTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.value")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuthExpiredToken(t *testing.T) {
	r, _, cleanup := setupMiddlewareTest(t)
	defer cleanup()

	claims := jwt.MapClaims{
		"user_id":    1,
		"role":       "student",
		"faculty_id": 1,
		"exp":        time.Now().Add(-time.Hour).Unix(),
		"iat":        time.Now().Add(-2 * time.Hour).Unix(),
		"nbf":        time.Now().Add(-2 * time.Hour).Unix(),
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(testSecret)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuthUserNotFound(t *testing.T) {
	r, mock, cleanup := setupMiddlewareTest(t)
	defer cleanup()

	token := generateTestToken(t, 1, "student", intPtr(1))

	mock.ExpectQuery(`SELECT id, role, faculty_id, is_banned`).
		WithArgs(1).
		WillReturnError(sql.ErrNoRows)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuthBannedUser(t *testing.T) {
	r, mock, cleanup := setupMiddlewareTest(t)
	defer cleanup()

	token := generateTestToken(t, 1, "student", intPtr(1))

	mock.ExpectQuery(`SELECT id, role, faculty_id, is_banned`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "role", "faculty_id", "is_banned",
		}).AddRow(1, "student", 1, true))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestRequireAuthSetsUserIDRoleFacultyID(t *testing.T) {
	r, mock, cleanup := setupMiddlewareTest(t)
	defer cleanup()

	token := generateTestToken(t, 1, "student", intPtr(2))

	mock.ExpectQuery(`SELECT id, role, faculty_id, is_banned`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "role", "faculty_id", "is_banned",
		}).AddRow(1, "student", 2, false))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequireAuthAdminDoesNotGetFacultyID(t *testing.T) {
	r, mock, cleanup := setupMiddlewareTest(t)
	defer cleanup()

	token := generateTestToken(t, 1, "admin", nil)

	mock.ExpectQuery(`SELECT id, role, faculty_id, is_banned`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "role", "faculty_id", "is_banned",
		}).AddRow(1, "admin", nil, false))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequireRoleAllowsNeededRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	}, RequireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequireRoleRejectsStudentForOrganizerAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/admin", func(c *gin.Context) {
		c.Set("role", "student")
		c.Next()
	}, RequireRole("organizer", "admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestCORSMiddlewareAllowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(CORSMiddleware([]string{"http://localhost:3000"}))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatal("expected allowed origin header")
	}
}

func TestCORSMiddlewarePreflightOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(CORSMiddleware([]string{"http://localhost:3000"}))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestCORSMiddlewareRejectedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(CORSMiddleware([]string{"http://localhost:3000"}))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://evil.com")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "http://evil.com" {
		t.Fatal("rejected origin should not be allowed")
	}
}

//
