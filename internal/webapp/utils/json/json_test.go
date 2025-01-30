package json

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	w := httptest.NewRecorder()
	data := Envelope{"message": "success"}
	headers := http.Header{}
	headers.Set("X-Custom-Header", "test")

	err := Write(w, http.StatusOK, data, headers)
	if err != nil {
		t.Fatalf("Write returned an unexpected error: %v", err)
	}

	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	if res.Header.Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", res.Header.Get("Content-Type"))
	}

	if res.Header.Get("X-Custom-Header") != "test" {
		t.Errorf("expected custom header test, got %s", res.Header.Get("X-Custom-Header"))
	}

	var body Envelope
	err = json.NewDecoder(res.Body).Decode(&body)
	if err != nil {
		t.Fatalf("error decoding response body: %v", err)
	}

	if body["message"] != "success" {
		t.Errorf("expected message 'success', got %v", body["message"])
	}
}

func TestRead_ValidJSON(t *testing.T) {
	reqBody := `{"name": "test"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	var data struct {
		Name string `json:"name"`
	}

	err := Read(w, req, &data)
	if err != nil {
		t.Fatalf("Read returned an unexpected error: %v", err)
	}

	if data.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", data.Name)
	}
}

func TestRead_InvalidJSON(t *testing.T) {
	reqBody := `{invalid}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	var data struct {
		Name string `json:"name"`
	}

	err := Read(w, req, &data)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}

	if !strings.Contains(err.Error(), "badly-formed JSON") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRead_UnknownField(t *testing.T) {
	reqBody := `{"unknown": "value"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	var data struct {
		Name string `json:"name"`
	}

	err := Read(w, req, &data)
	if err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}

	if !strings.Contains(err.Error(), "unknown key") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRead_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	var data struct {
		Name string `json:"name"`
	}

	err := Read(w, req, &data)
	if err == nil {
		t.Fatal("expected error for empty body, got nil")
	}

	if err.Error() != "body must not be empty" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRead_MaxBytes(t *testing.T) {
	largeJSON := `{"data": "` + strings.Repeat("x", 1048577) + `"}` // 1MB + 1 byte
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(largeJSON))
	req.Header.Set("Content-Type", "application/json")

	err := Read(httptest.NewRecorder(), req, &map[string]interface{}{})
	if err == nil || err.Error() != "body must not be larger than 1048576 bytes" {
		t.Errorf("unexpected error message: %v", err)
	}
}
