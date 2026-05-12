package uploads

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupUploadRouter(dir string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	h := NewHandler(dir)

	r.POST("/uploads", h.UploadImage)
	r.Static("/uploads", dir)

	return r
}

func multipartBody(t *testing.T, field, filename, contentType string, data []byte) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {
			`form-data; name="` + field + `"; filename="` + filename + `"`,
		},
		"Content-Type": {contentType},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := part.Write(data); err != nil {
		t.Fatal(err)
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	return body, writer.FormDataContentType()
}

func upload(t *testing.T, r *gin.Engine, field, filename, contentType string, data []byte) *httptest.ResponseRecorder {
	t.Helper()

	body, formType := multipartBody(t, field, filename, contentType, data)

	req := httptest.NewRequest(http.MethodPost, "/uploads", body)
	req.Header.Set("Content-Type", formType)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func parseURL(t *testing.T, body string) string {
	t.Helper()

	var res struct {
		URL string `json:"url"`
	}

	if err := json.Unmarshal([]byte(body), &res); err != nil {
		t.Fatal(err)
	}

	return res.URL
}

func TestUploadSuccessJPGPNGGIFWebP(t *testing.T) {
	tests := []struct {
		filename    string
		contentType string
	}{
		{"image.jpg", "image/jpeg"},
		{"image.png", "image/png"},
		{"image.gif", "image/gif"},
		{"image.webp", "image/webp"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			dir := t.TempDir()
			r := setupUploadRouter(dir)

			w := upload(t, r, "image", tt.filename, tt.contentType, []byte("fake image"))

			if w.Code != http.StatusCreated {
				t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
			}

			url := parseURL(t, w.Body.String())

			if !strings.HasPrefix(url, "/uploads/") {
				t.Fatalf("expected /uploads url, got %s", url)
			}

			path := filepath.Join(dir, strings.TrimPrefix(url, "/uploads/"))

			if _, err := os.Stat(path); err != nil {
				t.Fatalf("expected file saved, got %v", err)
			}
		})
	}
}

func TestUploadNoMultipartFileImage(t *testing.T) {
	dir := t.TempDir()
	r := setupUploadRouter(dir)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("wrong", "value")
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/uploads", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUploadSizeMoreThan5MB(t *testing.T) {
	dir := t.TempDir()
	r := setupUploadRouter(dir)

	data := bytes.Repeat([]byte("a"), maxImageSize+1)

	w := upload(t, r, "image", "big.png", "image/png", data)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUploadInvalidContentType(t *testing.T) {
	dir := t.TempDir()
	r := setupUploadRouter(dir)

	w := upload(t, r, "image", "image.png", "text/plain", []byte("fake"))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUploadInvalidExtension(t *testing.T) {
	dir := t.TempDir()
	r := setupUploadRouter(dir)

	w := upload(t, r, "image", "image.exe", "image/png", []byte("fake"))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUploadCorrectContentTypeWrongExtension(t *testing.T) {
	dir := t.TempDir()
	r := setupUploadRouter(dir)

	w := upload(t, r, "image", "image.txt", "image/jpeg", []byte("fake"))

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestFilenameGeneratedRandomAndKeepsExtension(t *testing.T) {
	dir := t.TempDir()
	r := setupUploadRouter(dir)

	w1 := upload(t, r, "image", "image.png", "image/png", []byte("one"))
	w2 := upload(t, r, "image", "image.png", "image/png", []byte("two"))

	if w1.Code != http.StatusCreated || w2.Code != http.StatusCreated {
		t.Fatalf("expected 201")
	}

	url1 := parseURL(t, w1.Body.String())
	url2 := parseURL(t, w2.Body.String())

	if url1 == url2 {
		t.Fatal("expected random different filenames")
	}

	if !strings.HasSuffix(url1, ".png") || !strings.HasSuffix(url2, ".png") {
		t.Fatal("expected extension .png")
	}
}

func TestUploadDirectoryCreated(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "new-uploads")
	r := setupUploadRouter(dir)

	w := upload(t, r, "image", "image.jpg", "image/jpeg", []byte("fake"))

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("expected upload directory created, got %v", err)
	}
}

func TestReturnsUploadsFileURL(t *testing.T) {
	dir := t.TempDir()
	r := setupUploadRouter(dir)

	w := upload(t, r, "image", "image.webp", "image/webp", []byte("fake"))

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	url := parseURL(t, w.Body.String())

	if !strings.HasPrefix(url, "/uploads/") {
		t.Fatalf("expected /uploads/<file>, got %s", url)
	}
}

func TestStaticUploadsServesFile(t *testing.T) {
	dir := t.TempDir()
	r := setupUploadRouter(dir)

	w := upload(t, r, "image", "image.gif", "image/gif", []byte("static file"))

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	url := parseURL(t, w.Body.String())

	req := httptest.NewRequest(http.MethodGet, url, nil)
	res := httptest.NewRecorder()

	r.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}

	if res.Body.String() != "static file" {
		t.Fatalf("expected static file content, got %s", res.Body.String())
	}
}
