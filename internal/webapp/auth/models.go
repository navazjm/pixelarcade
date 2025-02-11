package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/navazjm/pixelarcade/internal/webapp/utils/database"
)

const (
	defaultProfilePicture = "link-to-default.jpg"
)

type Model struct {
	DB *sql.DB
}

// CRUD Users

func (m Model) InsertUser(user *User) error {
	query := `
        INSERT INTO auth_users (name, email, password_hash, profile_picture, provider, is_active, is_verified) 
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, updated_at, version, role_id`

	if user.ProfilePicture == "" {
		user.ProfilePicture = defaultProfilePicture
	}

	args := []any{user.Name, user.Email, user.Password.hash, user.ProfilePicture, user.Provider, user.IsActive, user.IsVerified}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.Version, &user.RoleID)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "auth_users_email_key"`:
			return database.ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (m Model) GetUserByID(userID int64) (*User, error) {
	query := `
        SELECT * 
        FROM auth_users
        WHERE id = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Version,
		&user.IsActive,
		&user.Email,
		&user.Name,
		&user.ProfilePicture,
		&user.Password.hash,
		&user.Provider,
		&user.RoleID,
		&user.IsVerified,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, database.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m Model) GetUserByEmail(email string) (*User, error) {
	query := `
        SELECT * 
        FROM auth_users
        WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Version,
		&user.IsActive,
		&user.Email,
		&user.Name,
		&user.ProfilePicture,
		&user.Password.hash,
		&user.Provider,
		&user.RoleID,
		&user.IsVerified,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, database.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m Model) GetUserFromToken(tokenScope, tokenPlaintext string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
        SELECT au.*
        FROM auth_users as au
        INNER JOIN auth_tokens as at
        ON au.id = at.user_id
        WHERE at.hash = $1
        AND at.scope = $2 
        AND at.expiry > $3`

	args := []any{tokenHash[:], tokenScope, time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Version,
		&user.IsActive,
		&user.Email,
		&user.Name,
		&user.ProfilePicture,
		&user.Password.hash,
		&user.Provider,
		&user.RoleID,
		&user.IsVerified,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, database.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (m Model) UpdateUserByID(user *User) error {
	query := `
        UPDATE auth_users 
        SET updated_at = NOW(), version = version + 1, is_active = $1, email = $2, name = $3, profile_picture = $4, password_hash = $5, provider = $6, role_id = $7, is_verified = $8 
        WHERE id = $9 AND version = $10
        RETURNING created_at, updated_at, version`

	args := []any{
		user.IsActive,
		user.Email,
		user.Name,
		user.ProfilePicture,
		user.Password.hash,
		user.Provider,
		user.RoleID,
		user.IsVerified,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.CreatedAt, &user.UpdatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "auth_users_email_key"`:
			return database.ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return database.ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m Model) DeleteUserByID(userID int64) error {
	query := `
        DELETE FROM auth_users 
        WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return database.ErrRecordNotFound
	}

	return nil
}

// Tokens

func (m Model) InsertToken(token *Token) error {
	query := `
        INSERT INTO auth_tokens (hash, user_id, expiry, scope) 
        VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

func (m Model) DeleteAllTokensForUser(scope string, userID int64) error {
	query := `
        DELETE FROM auth_tokens 
        WHERE scope = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, scope, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return database.ErrRecordNotFound
	}

	return nil
}
