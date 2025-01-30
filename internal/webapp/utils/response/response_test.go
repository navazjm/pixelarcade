package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"log/slog"

	pa_json "github.com/navazjm/pixelarcade/internal/webapp/utils/json"
)

func TestLogError(t *testing.T) {
	// Create a mock logger using the standard log/slog package.
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	r := httptest.NewRequest(http.MethodGet, "/test-uri", nil)
	r.RequestURI = "/test-uri"

	// Capture the output by overriding the logger's output.
	LogError(r, logger, fmt.Errorf("some error"))
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test-uri", nil)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Simulate a bad request error response
	Error(w, r, logger, http.StatusBadRequest, "bad request error")

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}

	var env pa_json.Envelope
	err := json.NewDecoder(resp.Body).Decode(&env)
	if err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if env["error"] != "bad request error" {
		t.Errorf("expected error message 'bad request error', got '%s'", env["error"])
	}
}

func TestServerError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test-uri", nil)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Log and return a server error
	ServerError(w, r, logger, fmt.Errorf("some error"))

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	var env pa_json.Envelope
	err := json.NewDecoder(resp.Body).Decode(&env)
	if err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if env["error"] != "the server encountered a problem and could not process your request" {
		t.Errorf("expected error message 'the server encountered a problem and could not process your request', got '%s'", env["error"])
	}
}

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	NotFound(w, r, logger)

	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}

	var env pa_json.Envelope
	err := json.NewDecoder(resp.Body).Decode(&env)
	if err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if env["error"] != "the requested resource could not be found" {
		t.Errorf("expected error message 'the requested resource could not be found', got '%s'", env["error"])
	}
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/test-uri", nil)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	err := fmt.Errorf("invalid input")

	BadRequest(w, r, logger, err)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}

	var env pa_json.Envelope
	err = json.NewDecoder(resp.Body).Decode(&env)
	if err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if env["error"] != "invalid input" {
		t.Errorf("expected error message 'invalid input', got '%s'", env["error"])
	}
}

func TestFailedValidation(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/test-uri", nil)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	errors := map[string]string{
		"field1": "cannot be empty",
		"field2": "invalid format",
	}

	FailedValidation(w, r, logger, errors)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, resp.StatusCode)
	}

	var env pa_json.Envelope
	err := json.NewDecoder(resp.Body).Decode(&env)
	if err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if fmt.Sprintf("%v", env["error"]) != fmt.Sprintf("%v", errors) {
		t.Errorf("expected errors map '%v', got '%v'", errors, env["error"])
	}
}

func TestPermissionDenied(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test-uri", nil)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	PermissionDenied(w, r, logger)

	resp := w.Result()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, resp.StatusCode)
	}

	var env pa_json.Envelope
	err := json.NewDecoder(resp.Body).Decode(&env)
	if err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if env["error"] != "you do not have permission to access this resource" {
		t.Errorf("expected error message 'you do not have permission to access this resource', got '%s'", env["error"])
	}
}

func TestOriginNotAllowed(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/test-uri", nil)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	origin := "http://example.com"

	OriginNotAllowed(w, r, logger, origin)

	resp := w.Result()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, resp.StatusCode)
	}

	var env pa_json.Envelope
	err := json.NewDecoder(resp.Body).Decode(&env)
	if err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if env["error"] != fmt.Sprintf("request origin '%s' is not allowed", origin) {
		t.Errorf("expected error message 'request origin '%s' is not allowed', got '%s'", origin, env["error"])
	}
}
