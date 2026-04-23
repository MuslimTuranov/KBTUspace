package users

import (
	"errors"
	"net/http"
	"strconv"

	"kbtuspace-backend/internal/auth"
	"kbtuspace-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// AdminGetAllUsers godoc
// @Summary     Get all users
// @Description Allows admin to get all users
// @Tags        admin
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array} models.User
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/users [get]
func (h *Handler) AdminGetAllUsers(c *gin.Context) {
	users, err := h.service.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	if users == nil {
		users = []models.User{}
	}

	c.JSON(http.StatusOK, users)
}

// GetProfile godoc
// @Summary     Get current user profile
// @Description Returns the profile of the authenticated user
// @Tags        users
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	userIDAny, _ := c.Get("userID")
	userID, ok := userIDAny.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.service.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile godoc
// @Summary     Update current user profile
// @Description Updates the profile of the authenticated user
// @Tags        users
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       input body models.UpdateProfileInput true "Profile update data"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     409 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	userIDAny, _ := c.Get("userID")
	userID, ok := userIDAny.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var input models.UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	user, err := h.service.UpdateProfile(userID, input)
	if err != nil {
		if errors.Is(err, auth.ErrDuplicateEmail) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, auth.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// AdminUpdateUser godoc
// @Summary     Admin update user
// @Description Allows admin to update any user's data (role, ban status, etc.)
// @Tags        admin
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id    path int true "User ID"
// @Param       input body models.AdminUpdateUserInput true "User update data"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     409 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/users/{id} [patch]
func (h *Handler) AdminUpdateUser(c *gin.Context) {
	targetID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var input models.AdminUpdateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	user, err := h.service.AdminUpdateUser(targetID, input)
	if err != nil {
		if errors.Is(err, auth.ErrDuplicateEmail) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, auth.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}
