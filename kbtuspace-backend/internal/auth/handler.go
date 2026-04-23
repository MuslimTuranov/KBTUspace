package auth

import (
	"errors"
	"net/http"

	"kbtuspace-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register user
// @Summary     Register a new user
// @Description Creates a new student or user
// @Tags        auth
// @Accept      json
// @Produce     json
// @Param       input body models.RegisterInput true "Registration info"
// @Success     201 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     409 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var input models.RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if err := h.service.RegisterUser(input); err != nil {
		if errors.Is(err, ErrDuplicateEmail) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func (h *Handler) Login(c *gin.Context) {
	var input models.LoginInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	token, err := h.service.LoginUser(input)
	if err != nil {
		if errors.Is(err, ErrUserBanned) {
			c.JSON(http.StatusForbidden, gin.H{"error": "User is banned"})
			return
		}
		if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
	})
}
