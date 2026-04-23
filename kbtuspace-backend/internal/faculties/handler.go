package faculties

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetAllFaculties godoc
// @Summary     Get list of faculties
// @Description Returns a list of all available faculties
// @Tags        faculties
// @Produce     json
// @Success     200 {array} models.Faculty
// @Failure     500 {object} map[string]interface{}
// @Router      /faculties [get]
func (h *Handler) GetAllFaculties(c *gin.Context) {
	fs, err := h.service.GetAllFaculties(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch faculties"})
		return
	}
	c.JSON(http.StatusOK, fs)
}
