package posts

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"kbtuspace-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	var input models.CreatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	userIDAny, _ := c.Get("userID")
	userID, ok := userIDAny.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	roleAny, _ := c.Get("role")
	role, ok := roleAny.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid role"})
		return
	}

	var facultyID *int
	if facultyIDAny, exists := c.Get("facultyID"); exists {
		if value, ok := facultyIDAny.(int); ok {
			facultyID = &value
		}
	}

	post, err := h.service.Create(userID, role, facultyID, input)
	if err != nil {
		if errors.Is(err, ErrFacultyRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	c.JSON(http.StatusCreated, post)
}

func (h *Handler) GetAll(c *gin.Context) {
	var facultyID *int
	globalFeed := c.DefaultQuery("global", "false") == "true"

	if !globalFeed {
		if value := c.Query("faculty_id"); value != "" {
			id, err := strconv.Atoi(value)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid faculty_id"})
				return
			}
			facultyID = &id
		} else {
			// If no faculty_id provided, get from user context
			if fid, exists := c.Get("facultyID"); exists {
				if id, ok := fid.(int); ok {
					facultyID = &id
				}
			}
		}
	}

	posts, err := h.service.GetAll(facultyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}

	if posts == nil {
		posts = []models.Post{}
	}

	c.JSON(http.StatusOK, posts)
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	post, err := h.service.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch post"})
		return
	}

	c.JSON(http.StatusOK, post)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var input models.UpdatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	userIDAny, _ := c.Get("userID")
	roleAny, _ := c.Get("role")

	userID, ok1 := userIDAny.(int)
	role, ok2 := roleAny.(string)

	if !ok1 || !ok2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid auth context"})
		return
	}

	var facultyID *int
	if facultyIDAny, exists := c.Get("facultyID"); exists {
		if value, ok := facultyIDAny.(int); ok {
			facultyID = &value
		}
	}

	err = h.service.Update(id, userID, role, facultyID, input)
	if err != nil {
		if errors.Is(err, ErrPinForbidden) || errors.Is(err, ErrFacultyRequired) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit your own posts"})
			return
		}
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post updated successfully"})
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	userIDAny, _ := c.Get("userID")
	roleAny, _ := c.Get("role")

	userID, ok1 := userIDAny.(int)
	role, ok2 := roleAny.(string)

	if !ok1 || !ok2 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid auth context"})
		return
	}

	err = h.service.Delete(id, userID, role)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own posts"})
			return
		}
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}
