package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrInvalidToken         = errors.New("invalid token")
	ErrExpiredToken         = errors.New("token has expired")
)

const (
	TokenExpireDuration = 72 * time.Hour
)

func GenerateToken(userID int, role string, facultyID *int, secretKey []byte) (string, error) {
	if len(secretKey) < 32 {
		return "", errors.New("secret key must be at least 32 bytes")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"user_id":    userID,
		"role":       role,
		"faculty_id": facultyID,
		"exp":        now.Add(TokenExpireDuration).Unix(),
		"iat":        now.Unix(),
		"nbf":        now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

func ParseToken(tokenString string, secretKey []byte) (jwt.MapClaims, error) {
	if len(secretKey) < 32 {
		return nil, errors.New("secret key must be at least 32 bytes")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, ErrExpiredToken
		}
	}

	return claims, nil
}
