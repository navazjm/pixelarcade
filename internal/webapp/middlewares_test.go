package webapp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func normalizeHeader(header string) string {
	return strings.Join(strings.Fields(header), "")
}

func TestSecureHeaders(t *testing.T) {
	app := setupTestApp()

	// Mock next handler
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := app.secureHeaders(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Assert that headers were set correctly
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Check for specific headers
	expectedCSP := `default-src "self"; connect-src "self"; style-src "self" "unsafe-inline" fonts.googleapis.com; font-src "self" fonts.gstatic.com; img-src "self"; script-src "self"; frame-src "self";`
	actualCSP := rec.Header().Get("Content-Security-Policy")
	if normalizeHeader(actualCSP) != normalizeHeader(expectedCSP) {
		t.Errorf("expected Content-Security-Policy header to be %s, got %s", expectedCSP, rec.Header().Get("Content-Security-Policy"))
	}
	if rec.Header().Get("Strict-Transport-Security") != "max-age=15552000; includeSubDomains" {
		t.Errorf("expected Strict-Transport-Security header to be 'max-age=15552000; includeSubDomains', got %s", rec.Header().Get("Strict-Transport-Security"))
	}
}

func TestLogRequest(t *testing.T) {
	app := setupTestApp()

	// Mock next handler
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := app.logRequest(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// We can't assert on the logger directly, but we can check if the request proceeds
	handler.ServeHTTP(rec, req)

	// Assert the response code is OK
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestRecoverPanic(t *testing.T) {
	app := setupTestApp()

	// Mock next handler that causes a panic
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("Test panic")
	})

	handler := app.recoverPanic(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Assert that the server error response is triggered
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestEnforceCORS(t *testing.T) {
	app := setupTestApp()

	// Mock next handler
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := app.enforceCORS(next)

	// Test a request with a trusted origin
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Assert that CORS headers are set and request is allowed
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Errorf("expected Access-Control-Allow-Origin header to be 'https://example.com', got %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}

	// Test a request with an untrusted origin
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://untrusted.com")
	rec = httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Assert that request is blocked with the appropriate response
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestRateLimit(t *testing.T) {
	app := setupTestApp()

	handler := app.Routes()

	// Test a rate-limited request
	req := httptest.NewRequest(http.MethodGet, "/api/healthcheck", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Assert that the rate limit isn't exceeded
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Simulate multiple requests that exceed the rate limit
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/healthcheck", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if i < 7 {
			// Expecting a successful response before rate-limiting kicks in
			if w.Code != http.StatusOK {
				t.Errorf("expected status OK, got %d", w.Code)
			}
		} else {
			// Expecting 429 status after exceeding rate limit
			if w.Code != http.StatusTooManyRequests {
				t.Errorf("expected status 429, got %d", w.Code)
			}
		}
	}
}
