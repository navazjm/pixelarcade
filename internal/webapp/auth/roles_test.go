package auth

import (
	"testing"

	"github.com/navazjm/pixelarcade/internal/webapp/utils/validator"
)

func TestValidateRole(t *testing.T) {
	v := validator.New()

	role := &Role{Name: ""}
	ValidateRole(v, role)
	if v.Valid() {
		t.Errorf("Expected validation error for empty role name")
	}

	v.ResetErrors()

	longName := string(make([]byte, 501))
	role = &Role{Name: longName}
	ValidateRole(v, role)
	if v.Valid() {
		t.Errorf("Expected validation error for role name exceeding 500 bytes")
	}

	v.ResetErrors()

	role = &Role{Name: "Valid Role"}
	ValidateRole(v, role)
	if !v.Valid() {
		t.Errorf("Did not expect validation error for valid role name")
	}
}
