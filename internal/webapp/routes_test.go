package webapp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoutes(t *testing.T) {
	app := setupTestApp() // Initialize the app with routes

	// Test /api/healthcheck route
	t.Run("Healthcheck", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/healthcheck", nil)
		rec := httptest.NewRecorder()

		handler := app.Routes() // This will apply all middlewares and routes

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	// Test NotFound handler
	t.Run("NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/nonexistent", nil)
		rec := httptest.NewRecorder()

		handler := app.Routes() // This will apply all middlewares and routes

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	// Test MethodNotAllowed handler
	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/healthcheck", nil) // Using POST instead of GET
		rec := httptest.NewRecorder()

		handler := app.Routes() // This will apply all middlewares and routes

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
		}
	})
}
