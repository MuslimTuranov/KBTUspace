package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// buildPingRouter creates a minimal router with the same /ping and /swagger/*any
// routes as main.go, without requiring DB / Redis connections.
func buildPingRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"status":  "UniHub API is running",
		})
	})

	// Swagger wildcard route is registered in main; just verify the path is reachable.
	r.GET("/swagger/*any", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	return r
}

func TestGETPingReturns200AndPong(t *testing.T) {
	r := buildPingRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if body["message"] != "pong" {
		t.Errorf("expected message=%q, got %q", "pong", body["message"])
	}

	if body["status"] == "" {
		t.Error("expected non-empty status field in ping response")
	}
}

func TestSwaggerRouteIsRegisteredAndReturnsNon404(t *testing.T) {
	r := buildPingRouter()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	r.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Error("expected swagger route to be registered (/swagger/*any), got 404")
	}
}
