package auth

import (
	"time"

	"github.com/navazjm/pixelarcade/internal/webapp/utils/validator"
)

type Role struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
	IsActive  bool      `json:"is_active"`
	Name      string    `json:"name"`
}

// roles value should match that of their ID -> Reference "/migrations/000001_create_auth_roles_table.up.sql"
type RoleID int16

const (
	RoleBasic RoleID = iota + 1
	RoleAdmin
)

func ValidateRole(v *validator.Validator, role *Role) {
	v.Check(role.Name != "", "name", "must be provided")
	v.Check(len(role.Name) <= 500, "name", "must not be more than 500 bytes long")
}
