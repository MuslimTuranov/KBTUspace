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

// Create godoc
// @Summary     Create a post
// @Description Create a new post (faculty or global). Global posts require admin approval.
// @Tags        posts
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       input body models.CreatePostInput true "Post data"
// @Success     201 {object} map[string]interface{}
// @Success     202 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /posts [post]
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

	statusCode := http.StatusCreated
	message := "Post created successfully"
	if post.Status == models.ContentStatusPending {
		statusCode = http.StatusAccepted
		message = "Global post submitted for admin approval"
	}

	c.JSON(statusCode, gin.H{
		"message": message,
		"post":    post,
	})
}

// GetAll godoc
// @Summary     Get all posts
// @Description Returns posts filtered by faculty or global feed
// @Tags        posts
// @Produce     json
// @Security    BearerAuth
// @Param       global     query bool false "If true, returns global feed"
// @Param       faculty_id query int  false "Filter by faculty ID"
// @Success     200 {array}  models.Post
// @Failure     400 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /posts [get]
func (h *Handler) GetAll(c *gin.Context) {
	var facultyID *int
	globalFeed := c.DefaultQuery("global", "false") == "true"

	roleAny, _ := c.Get("role")
	role, _ := roleAny.(string)

	if !globalFeed {
		if value := c.Query("faculty_id"); value != "" {
			id, err := strconv.Atoi(value)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid faculty_id"})
				return
			}
			facultyID = &id
		} else {
			if fid, exists := c.Get("facultyID"); exists {
				if id, ok := fid.(int); ok {
					facultyID = &id
				}
			}
		}
	}

	posts, err := h.service.GetAll(facultyID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}

	if posts == nil {
		posts = []models.Post{}
	}

	c.JSON(http.StatusOK, posts)
}

// GetByID godoc
// @Summary     Get post by ID
// @Description Returns a single post by its ID
// @Tags        posts
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Post ID"
// @Success     200 {object} models.Post
// @Failure     400 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /posts/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
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

	post, err := h.service.GetByID(id, role, facultyID)
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

// Update godoc
// @Summary     Update a post
// @Description Update an existing post by ID
// @Tags        posts
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id    path int true "Post ID"
// @Param       input body models.UpdatePostInput true "Post update data"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     403 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /posts/{id} [put]
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
		if errors.Is(err, ErrPinForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrFacultyRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

// Delete godoc
// @Summary     Delete a post
// @Description Delete a post by ID (owner or admin only)
// @Tags        posts
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Post ID"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     403 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /posts/{id} [delete]
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

// Pin godoc
// @Summary     Pin or unpin a post
// @Description Toggle pin status of a post (organizer/admin only)
// @Tags        posts
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id    path int true "Post ID"
// @Param       input body models.PinPostInput true "Pin status"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     403 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /posts/{id}/pin [patch]
func (h *Handler) Pin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var input models.PinPostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
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

	if err := h.service.Pin(id, role, facultyID, input.IsPinned); err != nil {
		if errors.Is(err, ErrPinForbidden) || errors.Is(err, ErrInvalidPinScope) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, ErrFacultyRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update pin status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post pin status updated successfully"})
}

// ListPendingGlobal godoc
// @Summary     List pending global posts
// @Description Returns all global posts awaiting admin approval
// @Tags        admin
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array}  models.Post
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/moderation/global-content [get]
func (h *Handler) ListPendingGlobal(c *gin.Context) {
	posts, err := h.service.ListPendingGlobal()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending posts"})
		return
	}

	if posts == nil {
		posts = []models.Post{}
	}

	c.JSON(http.StatusOK, posts)
}

// Approve godoc
// @Summary     Approve a global post
// @Description Admin approves a pending global post
// @Tags        admin
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Post ID"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/posts/{id}/approve [patch]
func (h *Handler) Approve(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	adminIDAny, _ := c.Get("userID")
	adminID, ok := adminIDAny.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.service.Approve(id, adminID); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pending global post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post approved successfully"})
}

// Reject godoc
// @Summary     Reject a global post
// @Description Admin rejects a pending global post with a reason
// @Tags        admin
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id    path int true "Post ID"
// @Param       input body models.RejectContentInput true "Rejection reason"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/posts/{id}/reject [patch]
func (h *Handler) Reject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var input models.RejectContentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if err := h.service.Reject(id, input.Reason); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pending global post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post rejected successfully"})
}

// AdminDelete godoc
// @Summary     Admin delete a post
// @Description Admin forcefully deletes any post by ID
// @Tags        admin
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Post ID"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/posts/{id} [delete]
func (h *Handler) AdminDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	adminIDAny, _ := c.Get("userID")
	adminID, ok := adminIDAny.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.service.Delete(id, adminID, "admin"); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}
