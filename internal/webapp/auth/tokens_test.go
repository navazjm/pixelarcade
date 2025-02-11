package auth

import (
	"testing"
	"time"

	"github.com/navazjm/pixelarcade/internal/webapp/utils/validator"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		userID  int64
		ttl     time.Duration
		scope   string
		wantErr bool
	}{
		{userID: 1, ttl: time.Hour, scope: ScopeAuthentication, wantErr: false},
		{userID: 2, ttl: time.Minute * 30, scope: ScopeAuthentication, wantErr: false},
		{userID: 3, ttl: -time.Minute * 5, scope: ScopeAuthentication, wantErr: true}, // Invalid TTL
	}

	for _, tt := range tests {
		t.Run("Testing token generation", func(t *testing.T) {
			token, err := generateToken(tt.userID, tt.ttl, tt.scope)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
				if token != nil {
					t.Errorf("Expected nil token, but got %v", token)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if token == nil {
					t.Error("Expected token, but got nil")
				}
				if token.UserID != tt.userID {
					t.Errorf("Expected UserID %d, but got %d", tt.userID, token.UserID)
				}
				if token.Scope != tt.scope {
					t.Errorf("Expected Scope %s, but got %s", tt.scope, token.Scope)
				}
				if !time.Now().Before(token.Expiry) {
					t.Errorf("Expected token to expire in the future, but got %v", token.Expiry)
				}
				if len(token.Plaintext) != 26 {
					t.Errorf("Expected token plaintext to have length 26, but got %d", len(token.Plaintext))
				}
				if len(token.Hash) != 32 {
					t.Errorf("Expected token hash to have length 32, but got %d", len(token.Hash))
				}
			}
		})
	}
}

func TestValidateTokenPlaintext(t *testing.T) {
	v := validator.New()

	tests := []struct {
		tokenPlaintext string
		wantErr        bool
	}{
		{"ABCDEFGHIJKLMNOPQRSTUVWXYZ", false},    // Valid token
		{"", true},                               // Empty token
		{"short", true},                          // Invalid token length
		{"ABCDEFGHIJKLMNOPQRSTUVWXYZ1234", true}, // Invalid token length
	}

	for _, tt := range tests {
		t.Run("Testing token validation", func(t *testing.T) {
			v.ResetErrors()
			ValidateTokenPlaintext(v, tt.tokenPlaintext)

			if tt.wantErr {
				if v.Valid() {
					t.Error("Expected validation error, but got none")
				}
			} else {
				if !v.Valid() {
					t.Errorf("Expected no validation errors, but got %v", v.Errors)
				}
			}
		})
	}
}
