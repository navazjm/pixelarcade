package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContextSetUser(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	user := &User{ID: 1, Name: "John Doe"}
	r = ContextSetUser(r, user)

	ctxUser := r.Context().Value(CtxKeyUser)
	if ctxUser == nil {
		t.Fatal("Expected user in context but got nil")
	}

	retrievedUser, ok := ctxUser.(*User)
	if !ok {
		t.Fatal("Expected value of type *User in context but got different type")
	}

	if retrievedUser.ID != user.ID || retrievedUser.Name != user.Name {
		t.Errorf("Retrieved user does not match: got %+v, want %+v", retrievedUser, user)
	}
}

func TestContextGetUser(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	user := &User{ID: 1, Name: "Jane Doe"}
	r = ContextSetUser(r, user)

	retrievedUser := ContextGetUser(r)
	if retrievedUser == nil {
		t.Fatal("Expected user from context but got nil")
	}

	if retrievedUser.ID != user.ID || retrievedUser.Name != user.Name {
		t.Errorf("Retrieved user does not match: got %+v, want %+v", retrievedUser, user)
	}
}

func TestContextGetUser_PanicOnMissingUser(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Expected panic on missing user in context but did not panic")
		}
	}()

	r := httptest.NewRequest(http.MethodGet, "http://example.com", nil)
	_ = ContextGetUser(r) // Should panic
}
