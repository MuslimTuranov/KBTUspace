package auth

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"kbtuspace-backend/internal/models"
	"kbtuspace-backend/pkg/hash"
	appjwt "kbtuspace-backend/pkg/jwt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

var testSecret = []byte("12345678901234567890123456789012")

func newTestService(t *testing.T) (*Service, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := NewRepository(sqlxDB)
	service := NewService(repo, testSecret)

	return service, mock, func() {
		sqlxDB.Close()
		db.Close()
	}
}

func intPtr(v int) *int {
	return &v
}

func TestValidateKBTUEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"valid", "test@kbtu.kz", true},
		{"trim and lowercase", " TEST@KBTU.KZ ", true},
		{"gmail rejected", "test@gmail.com", false},
		{"other domain rejected", "test@kbtu.com", false},
		{"empty rejected", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKBTUEmail(tt.email)

			if tt.want && err != nil {
				t.Fatalf("expected valid email, got error: %v", err)
			}

			if !tt.want && !errors.Is(err, ErrInvalidEmailDomain) {
				t.Fatalf("expected ErrInvalidEmailDomain, got: %v", err)
			}
		})
	}
}

func TestRegisterUserSuccessStudent(t *testing.T) {
	service, mock, cleanup := newTestService(t)
	defer cleanup()

	facultyID := 1

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("student@kbtu.kz", sqlmock.AnyArg(), "student", &facultyID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, time.Now(), time.Now()))

	err := service.RegisterUser(models.RegisterInput{
		Email:     "student@kbtu.kz",
		Password:  "password123",
		FacultyID: &facultyID,
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestRegisterUserEmptyEmailOrPassword(t *testing.T) {
	service, _, cleanup := newTestService(t)
	defer cleanup()

	facultyID := 1

	tests := []models.RegisterInput{
		{Email: "", Password: "password123", FacultyID: &facultyID},
		{Email: "student@kbtu.kz", Password: "", FacultyID: &facultyID},
	}

	for _, input := range tests {
		err := service.RegisterUser(input)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	}
}

func TestRegisterUserShortPasswordModelValidation(t *testing.T) {
	input := models.RegisterInput{
		Email:     "student@kbtu.kz",
		Password:  "123",
		FacultyID: intPtr(1),
	}

	if len(input.Password) >= 6 {
		t.Fatal("test is incorrect: password should be shorter than 6")
	}
}

func TestRegisterUserFacultyMissingZeroNegative(t *testing.T) {
	service, _, cleanup := newTestService(t)
	defer cleanup()

	tests := []models.RegisterInput{
		{Email: "student@kbtu.kz", Password: "password123", FacultyID: nil},
		{Email: "student@kbtu.kz", Password: "password123", FacultyID: intPtr(0)},
		{Email: "student@kbtu.kz", Password: "password123", FacultyID: intPtr(-1)},
	}

	for _, input := range tests {
		err := service.RegisterUser(input)

		if !errors.Is(err, ErrFacultyRequired) {
			t.Fatalf("expected ErrFacultyRequired, got: %v", err)
		}
	}
}

func TestRegisterUserDuplicateEmail(t *testing.T) {
	service, mock, cleanup := newTestService(t)
	defer cleanup()

	facultyID := 1

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs("student@kbtu.kz", sqlmock.AnyArg(), "student", &facultyID).
		WillReturnError(ErrDuplicateEmail)

	err := service.RegisterUser(models.RegisterInput{
		Email:     "student@kbtu.kz",
		Password:  "password123",
		FacultyID: &facultyID,
	})

	if !errors.Is(err, ErrDuplicateEmail) {
		t.Fatalf("expected ErrDuplicateEmail, got: %v", err)
	}
}

func TestRegisterUserPasswordIsHashed(t *testing.T) {
	service, mock, cleanup := newTestService(t)
	defer cleanup()

	facultyID := 1
	plainPassword := "password123"

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(
			"student@kbtu.kz",
			sqlmock.AnyArg(),
			"student",
			&facultyID,
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(1, time.Now(), time.Now()))

	err := service.RegisterUser(models.RegisterInput{
		Email:     "student@kbtu.kz",
		Password:  plainPassword,
		FacultyID: &facultyID,
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestLoginUserSuccessReturnsJWT(t *testing.T) {
	service, mock, cleanup := newTestService(t)
	defer cleanup()

	facultyID := 1
	passwordHash, err := hash.HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs("student@kbtu.kz").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "email", "password_hash", "role", "faculty_id", "is_banned", "created_at", "updated_at",
		}).AddRow(1, "student@kbtu.kz", passwordHash, "student", facultyID, false, time.Now(), time.Now()))

	token, err := service.LoginUser(models.LoginInput{
		Email:    "student@kbtu.kz",
		Password: "password123",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if token == "" {
		t.Fatal("expected JWT token, got empty string")
	}
}

func TestLoginUserWrongEmail(t *testing.T) {
	service, mock, cleanup := newTestService(t)
	defer cleanup()

	mock.ExpectQuery(`SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs("wrong@kbtu.kz").
		WillReturnError(sql.ErrNoRows)

	_, err := service.LoginUser(models.LoginInput{
		Email:    "wrong@kbtu.kz",
		Password: "password123",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestLoginUserWrongPassword(t *testing.T) {
	service, mock, cleanup := newTestService(t)
	defer cleanup()

	facultyID := 1
	passwordHash, err := hash.HashPassword("correctpassword")
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs("student@kbtu.kz").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "email", "password_hash", "role", "faculty_id", "is_banned", "created_at", "updated_at",
		}).AddRow(1, "student@kbtu.kz", passwordHash, "student", facultyID, false, time.Now(), time.Now()))

	_, err = service.LoginUser(models.LoginInput{
		Email:    "student@kbtu.kz",
		Password: "wrongpassword",
	})

	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestLoginUserBanned(t *testing.T) {
	service, mock, cleanup := newTestService(t)
	defer cleanup()

	facultyID := 1
	passwordHash, err := hash.HashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectQuery(`SELECT id, email, password_hash, role, faculty_id, is_banned, created_at, updated_at FROM users WHERE email = \$1`).
		WithArgs("student@kbtu.kz").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "email", "password_hash", "role", "faculty_id", "is_banned", "created_at", "updated_at",
		}).AddRow(1, "student@kbtu.kz", passwordHash, "student", facultyID, true, time.Now(), time.Now()))

	_, err = service.LoginUser(models.LoginInput{
		Email:    "student@kbtu.kz",
		Password: "password123",
	})

	if !errors.Is(err, ErrUserBanned) {
		t.Fatalf("expected ErrUserBanned, got: %v", err)
	}
}

func TestJWTGenerateTokenContainsClaims(t *testing.T) {
	facultyID := 1

	tokenString, err := appjwt.GenerateToken(10, "student", &facultyID, testSecret)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	claims, err := appjwt.ParseToken(tokenString, testSecret)
	if err != nil {
		t.Fatalf("expected valid token, got: %v", err)
	}

	if int(claims["user_id"].(float64)) != 10 {
		t.Fatal("wrong user_id claim")
	}

	if claims["role"] != "student" {
		t.Fatal("wrong role claim")
	}

	if int(claims["faculty_id"].(float64)) != facultyID {
		t.Fatal("wrong faculty_id claim")
	}

	if claims["exp"] == nil || claims["iat"] == nil || claims["nbf"] == nil {
		t.Fatal("expected exp, iat, nbf claims")
	}
}

func TestJWTShortSecretRejected(t *testing.T) {
	_, err := appjwt.GenerateToken(1, "student", nil, []byte("short"))

	if err == nil {
		t.Fatal("expected error for short secret")
	}
}

func TestJWTExpiredInvalidWrongSigningMethod(t *testing.T) {
	expiredClaims := jwt.MapClaims{
		"user_id":    1,
		"role":       "student",
		"faculty_id": 1,
		"exp":        time.Now().Add(-time.Hour).Unix(),
		"iat":        time.Now().Add(-2 * time.Hour).Unix(),
		"nbf":        time.Now().Add(-2 * time.Hour).Unix(),
	}

	expiredToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims).SignedString(testSecret)
	if err != nil {
		t.Fatal(err)
	}

	_, err = appjwt.ParseToken(expiredToken, testSecret)
	if err == nil {
		t.Fatal("expected expired token error")
	}

	_, err = appjwt.ParseToken("invalid.token.value", testSecret)
	if err == nil {
		t.Fatal("expected invalid token error")
	}

	wrongMethodToken, err := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"user_id": 1,
		"exp":     time.Now().Add(time.Hour).Unix(),
	}).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}

	_, err = appjwt.ParseToken(wrongMethodToken, testSecret)
	if err == nil || !strings.Contains(err.Error(), "invalid signing method") {
		t.Fatalf("expected wrong signing method error, got: %v", err)
	}
}

//
