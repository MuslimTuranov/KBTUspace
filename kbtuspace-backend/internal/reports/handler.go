package reports

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
	var input models.CreateReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	report, err := h.service.Create(userID, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrTargetNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
		case errors.Is(err, ErrSelfReport):
			c.JSON(http.StatusBadRequest, gin.H{"error": "you cannot report your own content"})
		case errors.Is(err, ErrDuplicatePending):
			c.JSON(http.StatusConflict, gin.H{"error": "you already have a pending report for this content"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create report"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "report created successfully",
		"report":  report,
	})
}

func (h *Handler) List(c *gin.Context) {
	status := c.DefaultQuery("status", models.ReportStatusPending)

	reports, err := h.service.List(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reports"})
		return
	}

	if reports == nil {
		reports = []models.Report{}
	}

	c.JSON(http.StatusOK, reports)
}

func (h *Handler) Close(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report id"})
		return
	}

	var input models.CloseReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}

	adminIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	adminID, ok := adminIDValue.(int)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.service.Close(id, adminID, input); err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "pending report not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "report updated successfully"})
}
