package events

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
// @Summary     Create an event
// @Description Create a new event (organizer/admin only). Global events require admin approval.
// @Tags        events
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       input body models.CreateEventInput true "Event data"
// @Success     201 {object} map[string]interface{}
// @Success     202 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /events/ [post]
func (h *Handler) Create(c *gin.Context) {
	var input models.CreateEventInput
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

	event, err := h.service.Create(userID, role, facultyID, input)
	if err != nil {
		if errors.Is(err, ErrFacultyRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	statusCode := http.StatusCreated
	message := "Event created successfully"
	if event.Status == models.ContentStatusPending {
		statusCode = http.StatusAccepted
		message = "Global event submitted for admin approval"
	}

	c.JSON(statusCode, gin.H{
		"message": message,
		"event":   event,
	})
}

// GetAll godoc
// @Summary     Get all events
// @Description Returns events filtered by faculty or global feed
// @Tags        events
// @Produce     json
// @Security    BearerAuth
// @Param       global     query bool false "If true, returns global feed"
// @Param       faculty_id query int  false "Filter by faculty ID"
// @Success     200 {array}  models.Post
// @Failure     400 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /events [get]
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

	events, err := h.service.GetAll(facultyID, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	if events == nil {
		events = []models.Post{}
	}

	c.JSON(http.StatusOK, events)
}

// GetByID godoc
// @Summary     Get event by ID
// @Description Returns a single event by its ID
// @Tags        events
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Event ID"
// @Success     200 {object} models.Post
// @Failure     400 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /events/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
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

	event, err := h.service.GetByID(id, role, facultyID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch event"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// Update godoc
// @Summary     Update an event
// @Description Update an existing event by ID (organizer/admin only)
// @Tags        events
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id    path int true "Event ID"
// @Param       input body models.UpdateEventInput true "Event update data"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     403 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /events/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var input models.UpdateEventInput
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

	err = h.service.Update(id, role, facultyID, input)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only manage events in your faculty"})
			return
		}
		if errors.Is(err, ErrFacultyRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event updated successfully"})
}

// Delete godoc
// @Summary     Delete an event
// @Description Delete an event by ID (organizer/admin only)
// @Tags        events
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Event ID"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     403 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /events/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
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

	err = h.service.Delete(id, role, facultyID)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only manage events in your faculty"})
			return
		}
		if errors.Is(err, ErrFacultyRequired) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

// Register godoc
// @Summary     Register for an event
// @Description Register the current user for an event
// @Tags        events
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Event ID"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     409 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /events/{id}/register [post]
func (h *Handler) Register(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userIDAny, _ := c.Get("userID")
	userID, ok := userIDAny.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.service.Register(userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		if errors.Is(err, ErrEventFull) {
			c.JSON(http.StatusConflict, gin.H{"error": "Event is full"})
			return
		}
		if errors.Is(err, ErrAlreadyRegistered) {
			c.JSON(http.StatusConflict, gin.H{"error": "Already registered for this event"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register for event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully registered for event"})
}

// CancelRegistration godoc
// @Summary     Cancel event registration
// @Description Cancel the current user's registration for an event
// @Tags        events
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Event ID"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     409 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /events/{id}/register [delete]
func (h *Handler) CancelRegistration(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userIDAny, _ := c.Get("userID")
	userID, ok := userIDAny.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.service.CancelRegistration(userID, eventID); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		case errors.Is(err, ErrNotRegistered):
			c.JSON(http.StatusConflict, gin.H{"error": "Not registered for this event"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel registration"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration cancelled successfully"})
}

// MarkAttendance godoc
// @Summary     Mark user attendance
// @Description Mark a registered user as attended (organizer/admin only)
// @Tags        events
// @Produce     json
// @Security    BearerAuth
// @Param       id     path int true "Event ID"
// @Param       userId path int true "User ID"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /events/{id}/attendance/{userId} [patch]
func (h *Handler) MarkAttendance(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	targetUserID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	actorIDAny, _ := c.Get("userID")
	actorID, ok := actorIDAny.(int)
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

	if err := h.service.MarkAttended(actorID, role, facultyID, targetUserID, eventID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		if errors.Is(err, ErrForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only manage attendance for your own or faculty events"})
			return
		}
		if errors.Is(err, ErrNotRegistered) {
			c.JSON(http.StatusNotFound, gin.H{"error": "User is not registered for this event"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark attendance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Attendance marked successfully"})
}

// ListPendingGlobal godoc
// @Summary     List pending global events
// @Description Returns all global events awaiting admin approval
// @Tags        admin
// @Produce     json
// @Security    BearerAuth
// @Success     200 {array}  models.Post
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/moderation/global-content [get]
func (h *Handler) ListPendingGlobal(c *gin.Context) {
	events, err := h.service.ListPendingGlobal()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending events"})
		return
	}

	if events == nil {
		events = []models.Post{}
	}

	c.JSON(http.StatusOK, events)
}

// Approve godoc
// @Summary     Approve a global event
// @Description Admin approves a pending global event
// @Tags        admin
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Event ID"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/events/{id}/approve [patch]
func (h *Handler) Approve(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
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
			c.JSON(http.StatusNotFound, gin.H{"error": "Pending global event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event approved successfully"})
}

// Reject godoc
// @Summary     Reject a global event
// @Description Admin rejects a pending global event with a reason
// @Tags        admin
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id    path int true "Event ID"
// @Param       input body models.RejectContentInput true "Rejection reason"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/events/{id}/reject [patch]
func (h *Handler) Reject(c *gin.Context) {
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

	if err := h.service.Reject(id, input.Reason); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pending global event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event rejected successfully"})
}

// AdminDelete godoc
// @Summary     Admin delete an event
// @Description Admin forcefully deletes any event by ID
// @Tags        admin
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Event ID"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/events/{id} [delete]
func (h *Handler) AdminDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err := h.service.Delete(id, "admin", nil); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}
