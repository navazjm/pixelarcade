package auth

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthenticateNoCookie(t *testing.T) {
	service, _ := newMockService(t)

	// Define the next handler (dummy 200 response)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a request without a cookie
	req := httptest.NewRequest("GET", "http://example.com", nil)
	w := httptest.NewRecorder()

	// Call middleware
	service.Authenticate(next).ServeHTTP(w, req)

	// Expect OK response since no auth token was set
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, w.Result().StatusCode)
	}
}

func TestAuthenticateInvalidToken(t *testing.T) {
	service, mock := newMockService(t)

	// Define the next handler
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Simulate invalid token
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.AddCookie(&http.Cookie{Name: CookieAuthToken, Value: "invalid-token"})
	w := httptest.NewRecorder()

	// Expect DB query for user lookup (mocked to return no results)
	mock.ExpectQuery(`SELECT \* FROM auth_users WHERE token = \$1`).
		WithArgs("invalid-token").
		WillReturnError(sql.ErrNoRows)

	// Call middleware
	service.Authenticate(next).ServeHTTP(w, req)

	// Expect Unauthorized response
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, but got %d", http.StatusUnauthorized, w.Result().StatusCode)
	}
}

func TestRequireAuthenticatedUser(t *testing.T) {
	service, _ := newMockService(t)

	// Define the next handler
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a request with an anonymous user in the context
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	w := httptest.NewRecorder()

	// Simulate anonymous user
	req = ContextSetUser(req, AnonymousUser)

	// Call middleware
	service.RequireAuthenticatedUser(next).ServeHTTP(w, req)

	// Expect Unauthorized response
	if w.Result().StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status code %d, but got %d", http.StatusUnauthorized, w.Result().StatusCode)
	}
}

func TestRequireAuthenticatedUserAuthenticated(t *testing.T) {
	service, _ := newMockService(t)

	// Define the next handler
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a request with an authenticated user in the context
	req := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	w := httptest.NewRecorder()

	// Simulate authenticated user
	user := &User{ID: 1, Name: "John Doe"}
	req = ContextSetUser(req, user)

	// Call middleware
	service.RequireAuthenticatedUser(next).ServeHTTP(w, req)

	// Expect OK response
	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, w.Result().StatusCode)
	}
}
