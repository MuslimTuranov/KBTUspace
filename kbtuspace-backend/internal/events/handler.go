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
			if fid, exists := c.Get("facultyID"); exists {
				if id, ok := fid.(int); ok {
					facultyID = &id
				}
			}
		}
	}

	events, err := h.service.GetAll(facultyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	if events == nil {
		events = []models.Post{}
	}

	c.JSON(http.StatusOK, events)
}

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

	event, err := h.service.GetByID(id, role)
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
