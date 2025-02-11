package auth

import (
	"testing"

	"github.com/navazjm/pixelarcade/internal/webapp/utils/validator"
)

func TestUser_IsAnonymous(t *testing.T) {
	tests := []struct {
		user     *User
		expected bool
	}{
		{AnonymousUser, true},
		{&User{}, false},
	}

	for _, test := range tests {
		t.Run(test.user.Email, func(t *testing.T) {
			result := test.user.IsAnonymous()
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}

func TestValidateUser_Valid(t *testing.T) {
	v := validator.New()
	user := &User{
		Name:  "John Doe",
		Email: "johndoe@example.com",
	}

	// Mock password hash
	err := user.Password.Set("validpassword123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ValidateUser(v, user)
	if !v.Valid() {
		t.Errorf("expected validation to be valid, but got errors: %v", v.Errors)
	}
}

func TestValidateUser_Invalid(t *testing.T) {
	v := validator.New()
	validPwPlaintext := "validpassword123"
	invalidPwPlaintextShort := "short"
	emptyPwPlaintext := ""

	tests := []struct {
		user     *User
		hasError bool
	}{
		{
			user: &User{
				Name:     "",
				Email:    "johndoe@example.com",
				Password: password{plaintext: &validPwPlaintext},
			},
			hasError: true,
		},
		{
			user: &User{
				Name:     "John Doe",
				Email:    "invalid-email",
				Password: password{plaintext: &validPwPlaintext},
			},
			hasError: true,
		},
		{
			user: &User{
				Name:     "John Doe",
				Email:    "johndoe@example.com",
				Password: password{plaintext: &invalidPwPlaintextShort},
			},
			hasError: true,
		},
		{
			user: &User{
				Name:     "John Doe",
				Email:    "johndoe@example.com",
				Password: password{plaintext: &emptyPwPlaintext},
			},
			hasError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.user.Email, func(t *testing.T) {
			err := test.user.Password.Set(*test.user.Password.plaintext)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			v.ResetErrors()
			ValidateUser(v, test.user)

			if test.hasError && v.Valid() {
				t.Errorf("expected validation errors, but validation passed")
			}
			if !test.hasError && !v.Valid() {
				t.Errorf("expected validation to pass, but got errors: %v", v.Errors)
			}
		})
	}
}

func TestPassword_Set(t *testing.T) {
	password := &password{}
	err := password.Set("testpassword123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if password.plaintext == nil {
		t.Errorf("expected password.plaintext to be non-nil")
	}
	if len(password.hash) == 0 {
		t.Errorf("expected password.hash to be non-empty")
	}
}

func TestPassword_Matches(t *testing.T) {
	password := &password{}
	err := password.Set("testpassword123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("Correct password", func(t *testing.T) {
		match, err := password.Matches("testpassword123")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !match {
			t.Errorf("expected password to match")
		}
	})

	t.Run("Incorrect password", func(t *testing.T) {
		match, err := password.Matches("wrongpassword")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if match {
			t.Errorf("expected password not to match")
		}
	})
}
