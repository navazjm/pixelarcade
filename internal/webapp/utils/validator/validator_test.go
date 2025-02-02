package validator

import (
	"testing"
)

func TestValidator_Valid(t *testing.T) {
	v := New()

	// Check if the validator is valid when no errors are present
	if !v.Valid() {
		t.Errorf("Expected validator to be valid, but got errors")
	}

	// Add an error and check if the validator is invalid
	v.AddError("email", "Invalid email address")
	if v.Valid() {
		t.Errorf("Expected validator to be invalid, but got valid")
	}
}

func TestValidator_AddError(t *testing.T) {
	v := New()

	// Add an error and check if it is in the map
	v.AddError("email", "Invalid email address")
	if v.Errors["email"] != "Invalid email address" {
		t.Errorf("Expected 'Invalid email address' but got %v", v.Errors["email"])
	}

	// Try adding the same error again, it should not overwrite
	v.AddError("email", "Duplicate email error")
	if v.Errors["email"] != "Invalid email address" {
		t.Errorf("Expected 'Invalid email address' but got %v", v.Errors["email"])
	}
}

func TestValidator_RemoveError(t *testing.T) {
	v := New()

	// Add an error and check if it exists
	v.AddError("email", "Invalid email address")
	if v.Errors["email"] != "Invalid email address" {
		t.Errorf("Expected 'Invalid email address' but got %v", v.Errors["email"])
	}

	// Remove the error and check if it is removed
	v.RemoveError("email")
	if _, exists := v.Errors["email"]; exists {
		t.Errorf("Expected 'email' error to be removed, but it still exists")
	}

	// Ensure removing a non-existent error doesn't affect the map
	v.RemoveError("nonexistent")
	if len(v.Errors) != 0 {
		t.Errorf("Expected no errors after removing non-existent key, but got %v", v.Errors)
	}
}

func TestValidator_ResetErrors(t *testing.T) {
	v := New()

	// Add some errors
	v.AddError("email", "Invalid email address")
	v.AddError("username", "Username required")

	// Check that errors are present
	if len(v.Errors) != 2 {
		t.Errorf("Expected 2 errors, but got %v", len(v.Errors))
	}

	// Reset errors and check if the map is cleared
	v.ResetErrors()
	if len(v.Errors) != 0 {
		t.Errorf("Expected no errors after reset, but got %v", len(v.Errors))
	}
}

func TestValidator_Check(t *testing.T) {
	v := New()

	// Test a valid check (should not add an error)
	v.Check(true, "email", "Invalid email address")
	if len(v.Errors) != 0 {
		t.Errorf("Expected no errors, but got %v", v.Errors)
	}

	// Test an invalid check (should add an error)
	v.Check(false, "email", "Invalid email address")
	if v.Errors["email"] != "Invalid email address" {
		t.Errorf("Expected 'Invalid email address' but got %v", v.Errors["email"])
	}
}

func TestPermittedValue(t *testing.T) {
	// Test that PermittedValue correctly checks the value
	if !PermittedValue("apple", "apple", "banana", "cherry") {
		t.Errorf("Expected 'apple' to be permitted")
	}

	if PermittedValue("orange", "apple", "banana", "cherry") {
		t.Errorf("Expected 'orange' to not be permitted")
	}
}

func TestMatches(t *testing.T) {
	// Test email regex matching
	if !Matches("test@example.com", EmailRX) {
		t.Errorf("Expected 'test@example.com' to match the email pattern")
	}

	if Matches("invalid-email", EmailRX) {
		t.Errorf("Expected 'invalid-email' to not match the email pattern")
	}
}

func TestUnique(t *testing.T) {
	// Test unique function
	if !Unique([]int{1, 2, 3, 4, 5}) {
		t.Errorf("Expected values to be unique")
	}

	if Unique([]int{1, 2, 2, 4, 5}) {
		t.Errorf("Expected values to not be unique")
	}
}
