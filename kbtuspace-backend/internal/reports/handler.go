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

// CreateReport godoc
// @Summary     Create report
// @Description Create a report for a post/event/content
// @Tags        reports
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       input body models.CreateReportInput true "Report input"
// @Success     201 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     409 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /reports [post]
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

// ListReports godoc
// @Summary     Get reports list
// @Description Get reports filtered by status
// @Tags        admin
// @Produce     json
// @Security    BearerAuth
// @Param       status query string false "Report status (pending/closed/rejected)"
// @Success     200 {array} models.Report
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/reports [get]
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

// CloseReport godoc
// @Summary     Close report
// @Description Admin closes a report with resolution
// @Tags        admin
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id path int true "Report ID"
// @Param       input body models.CloseReportInput true "Close report input"
// @Success     200 {object} map[string]interface{}
// @Failure     400 {object} map[string]interface{}
// @Failure     401 {object} map[string]interface{}
// @Failure     404 {object} map[string]interface{}
// @Failure     500 {object} map[string]interface{}
// @Router      /admin/reports/{id}/close [patch]
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
