package uploads

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

const maxImageSize = 5 << 20

type Handler struct {
	dir string
}

func NewHandler(dir string) *Handler {
	return &Handler{dir: dir}
}

func (h *Handler) UploadImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImageSize)
	if err := c.Request.ParseMultipartForm(maxImageSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image must be 5 MB or smaller"})
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedImage(contentType, ext) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image must be JPG, PNG, GIF, or WebP"})
		return
	}

	if err := os.MkdirAll(h.dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare upload directory"})
		return
	}

	name, err := randomName(ext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate file name"})
		return
	}

	dst, err := os.Create(filepath.Join(h.dir, name))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"url": "/uploads/" + name})
}

func allowedImage(contentType, ext string) bool {
	switch contentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
	default:
		return false
	}

	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func randomName(ext string) (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]) + ext, nil
}
