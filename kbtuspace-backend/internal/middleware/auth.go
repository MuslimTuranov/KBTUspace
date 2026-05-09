package middleware

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"kbtuspace-backend/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type authUser struct {
	ID        int    `db:"id"`
	FacultyID *int   `db:"faculty_id"`
	Role      string `db:"role"`
	IsBanned  bool   `db:"is_banned"`
}

func RequireAuth(secretKey []byte, db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		tokenString := parts[1]
		claims, err := jwt.ParseToken(tokenString, secretKey)
		if err != nil {
			ctx := context.Background()
			if errors.Is(err, jwt.ErrExpiredToken) {
				slog.ErrorContext(ctx, "expired token", slog.Any("error", err))
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
				return
			}
			slog.ErrorContext(ctx, "invalid token", slog.Any("error", err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}

		var user authUser
		if err := db.GetContext(c.Request.Context(), &user, `
			SELECT id, role, faculty_id, is_banned
			FROM users
			WHERE id = $1
		`, int(userID)); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
				return
			}
			slog.ErrorContext(c.Request.Context(), "failed to load auth user", slog.Any("error", err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
			return
		}

		if user.IsBanned {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User is banned"})
			return
		}

		c.Set("userID", user.ID)
		c.Set("role", user.Role)

		if user.Role != "admin" && user.FacultyID != nil {
			c.Set("facultyID", *user.FacultyID)
		}

		c.Next()
	}
}

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid role"})
			return
		}

		roleAllowed := false
		for _, role := range allowedRoles {
			if roleStr == role {
				roleAllowed = true
				break
			}
		}

		if !roleAllowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}

		c.Next()
	}
}
