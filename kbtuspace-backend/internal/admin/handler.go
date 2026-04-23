package admin

import (
	"net/http"
	"strconv"

	"kbtuspace-backend/internal/events"
	"kbtuspace-backend/internal/models"
	"kbtuspace-backend/internal/posts"
	"kbtuspace-backend/internal/reports"
	"kbtuspace-backend/internal/users"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	postService   *posts.Service
	eventService  *events.Service
	userService   *users.Service
	reportService *reports.Service
}

func NewHandler(
	postService *posts.Service,
	eventService *events.Service,
	userService *users.Service,
	reportService *reports.Service,
) *Handler {
	return &Handler{
		postService:   postService,
		eventService:  eventService,
		userService:   userService,
		reportService: reportService,
	}
}

// GetGlobalContent godoc
// @Summary     List pending global posts and events
// @Description Returns a list of global content pending admin approval
// @Tags        admin
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       type query string false "Content type: all, posts, events"
// @Success     200 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/moderation/global-content [get]
func (h *Handler) GetGlobalContent(c *gin.Context) {
	contentType := c.DefaultQuery("type", "all")

	switch contentType {
	case "posts":
		postsList, err := h.postService.ListPendingGlobal()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending global posts"})
			return
		}
		c.JSON(http.StatusOK, postsList)

	case "events":
		eventsList, err := h.eventService.ListPendingGlobal()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending global events"})
			return
		}
		c.JSON(http.StatusOK, eventsList)

	default:
		postsList, postsErr := h.postService.ListPendingGlobal()
		eventsList, eventsErr := h.eventService.ListPendingGlobal()
		if postsErr != nil || eventsErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending global content"})
			return
		}
		if postsList == nil {
			postsList = []models.Post{}
		}
		if eventsList == nil {
			eventsList = []models.Post{}
		}
		c.JSON(http.StatusOK, gin.H{
			"posts":  postsList,
			"events": eventsList,
		})
	}
}

// ApprovePost godoc
// @Summary     Approve a global post
// @Description Admin approves a pending global post
// @Tags        admin
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Post ID"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/posts/{id}/approve [patch]
func (h *Handler) ApprovePost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	adminIDAny, _ := c.Get("userID")
	adminID, _ := adminIDAny.(int)

	if err := h.postService.Approve(id, adminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post approved successfully"})
}

// RejectPost godoc
// @Summary     Reject a global post
// @Description Admin rejects a pending global post with a reason
// @Tags        admin
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Post ID"
// @Param       input body models.RejectContentInput true "Rejection reason"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/posts/{id}/reject [patch]
func (h *Handler) RejectPost(c *gin.Context) {
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

	if err := h.postService.Reject(id, input.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post rejected successfully"})
}

// DeletePost godoc
// @Summary     Admin delete a post
// @Description Admin forcefully deletes any post by ID
// @Tags        admin
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Post ID"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/posts/{id} [delete]
func (h *Handler) DeletePost(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	adminIDAny, _ := c.Get("userID")
	adminID, _ := adminIDAny.(int)

	if err := h.postService.Delete(id, adminID, "admin"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}

// ApproveEvent godoc
// @Summary     Approve a global event
// @Description Admin approves a pending global event
// @Tags        admin
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Event ID"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/events/{id}/approve [patch]
func (h *Handler) ApproveEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	adminIDAny, _ := c.Get("userID")
	adminID, _ := adminIDAny.(int)

	if err := h.eventService.Approve(id, adminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event approved successfully"})
}

// RejectEvent godoc
// @Summary     Reject a global event
// @Description Admin rejects a pending global event
// @Tags        admin
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Event ID"
// @Param       input body models.RejectContentInput true "Rejection reason"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/events/{id}/reject [patch]
func (h *Handler) RejectEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var input models.RejectContentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	if err := h.eventService.Reject(id, input.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event rejected successfully"})
}

// DeleteEvent godoc
// @Summary     Admin delete an event
// @Description Admin forcefully deletes any event by ID
// @Tags        admin
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Event ID"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/events/{id} [delete]
func (h *Handler) DeleteEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	adminIDAny, _ := c.Get("userID")
	adminID, _ := adminIDAny.(int)

	if err := h.eventService.Delete(id, "admin", &adminID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}
