package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/database"
)

func TestRegisterNewUserHandler(t *testing.T) {
	authService, mock := newMockService(t)

	now := time.Now()
	mock.ExpectQuery("INSERT INTO auth_users").
		WithArgs("mike", "mike@test.com", sqlmock.AnyArg(), defaultProfilePicture, "N/A", true, false).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version", "role_id"}).
			AddRow(1, now, now, 1, 1))

	reqBody := map[string]any{
		"name":     "mike",
		"email":    "mike@test.com",
		"password": "SecurePass123!",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authService.RegisterNewUserHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestRegisterUser_EmailTaken(t *testing.T) {
	authService, mock := newMockService(t)

	// Simulate a unique constraint violation error on email
	mock.ExpectQuery("INSERT INTO auth_users").
		WithArgs("mike", "mike@test.com", sqlmock.AnyArg(), defaultProfilePicture, "N/A", true, false).
		WillReturnError(fmt.Errorf("pq: duplicate key value violates unique constraint \"auth_users_email_key\""))

	reqBody := map[string]any{
		"name":     "mike",
		"email":    "mike@test.com",
		"password": "SecurePass123!",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authService.RegisterNewUserHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestLoginUser_ValidLogin(t *testing.T) {
	authService, mock := newMockService(t)

	var password password
	err := password.Set("SecurePass123!")
	if err != nil {
		t.Errorf("failed to hash password: %s", err.Error())
	}

	mock.ExpectQuery("SELECT .* FROM auth_users WHERE email = ?").
		WithArgs("mike@test.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "version", "is_active", "email", "name",
			"profile_picture", "password", "provider", "role_id", "is_verified",
		}).AddRow(
			1, time.Now(), time.Now(), 1, true, "mike@test.com", "Mike",
			"default_profile_pic.jpg", password.hash, "N/A", 1, false,
		))

	mock.ExpectExec("INSERT INTO auth_tokens").
		WithArgs(sqlmock.AnyArg(), 1, sqlmock.AnyArg(), "authentication").
		WillReturnResult(sqlmock.NewResult(1, 1)) // Simulating an insert with 1 affected row

	reqBody := map[string]any{
		"email":    "mike@test.com",
		"password": *password.plaintext,
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authService.LoginUserHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestLoginUser_MissingEmail(t *testing.T) {
	authService, _ := newMockService(t)

	reqBody := map[string]any{
		"password": "SecurePass123!",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authService.LoginUserHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, resp.StatusCode)
	}
}

func TestLoginUser_MissingPassword(t *testing.T) {
	authService, _ := newMockService(t)

	reqBody := map[string]any{
		"email": "mike@test.com",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authService.LoginUserHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, resp.StatusCode)
	}
}

func TestLoginUser_InvalidEmailFormat(t *testing.T) {
	authService, _ := newMockService(t)

	reqBody := map[string]any{
		"email":    "invalid-email",
		"password": "SecurePass123!",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authService.LoginUserHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, resp.StatusCode)
	}
}

func TestLoginUser_UserNotFound(t *testing.T) {
	authService, mock := newMockService(t)

	mock.ExpectQuery("SELECT .* FROM auth_users WHERE email = ?").
		WithArgs("nonexistent@test.com").
		WillReturnError(database.ErrRecordNotFound)

	reqBody := map[string]any{
		"email":    "nonexistent@test.com",
		"password": "SecurePass123!",
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authService.LoginUserHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}

func TestLoginUser_PasswordMismatch(t *testing.T) {
	authService, mock := newMockService(t)

	var password password
	err := password.Set("DiffPass123!")
	if err != nil {
		t.Errorf("failed to hash password: %s", err.Error())
	}

	mock.ExpectQuery("SELECT .* FROM auth_users WHERE email = ?").
		WithArgs("mike@test.com").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "version", "is_active", "email", "name",
			"profile_picture", "password", "provider", "role_id", "is_verified",
		}).AddRow(
			1, time.Now(), time.Now(), 1, true, "mike@test.com", "Mike",
			"default_profile_pic.jpg", password.hash, "N/A", 1, false,
		))

	reqBody := map[string]any{
		"email":    "mike@test.com",
		"password": "SecurePass123!", // Incorrect password
	}
	jsonData, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	authService.LoginUserHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unmet expectations: %v", err)
	}
}
