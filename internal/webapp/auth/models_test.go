package auth

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/database"
)

func TestInsertUser(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	model := Model{DB: mockDB}
	user := &User{
		Name:           "John Doe",
		Email:          "john@example.com",
		Password:       password{hash: []byte("hashedpassword")},
		ProfilePicture: "",
		Provider:       "local",
		IsActive:       true,
		IsVerified:     false,
	}

	// Test Case 1: Valid insert
	mock.ExpectQuery("INSERT INTO auth_users").
		WithArgs(user.Name, user.Email, user.Password.hash, sqlmock.AnyArg(), user.Provider, user.IsActive, user.IsVerified).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version", "role_id"}).
			AddRow(1, time.Now(), time.Now(), 1, 2))

	err = model.InsertUser(user)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if user.ID != 1 {
		t.Errorf("expected user ID to be 1, got %d", user.ID)
	}

	// Test Case 2: Duplicate Email
	user = &User{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: password{hash: []byte("hashedpassword")},
	}

	mock.ExpectQuery("INSERT INTO auth_users").
		WithArgs(user.Name, user.Email, user.Password.hash, sqlmock.AnyArg(), user.Provider, user.IsActive, user.IsVerified).
		WillReturnError(errors.New(`pq: duplicate key value violates unique constraint "auth_users_email_key"`))

	err = model.InsertUser(user)
	if err != database.ErrDuplicateEmail {
		t.Errorf("expected duplicate email error, got %v", err)
	}

	// Test Case 3: Database error
	mock.ExpectQuery("INSERT INTO auth_users").
		WithArgs(user.Name, user.Email, user.Password.hash, sqlmock.AnyArg(), user.Provider, user.IsActive, user.IsVerified).
		WillReturnError(sql.ErrConnDone)

	err = model.InsertUser(user)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetUserByID(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	model := Model{DB: mockDB}

	// Test Case 1: Valid select
	mock.ExpectQuery("SELECT .* FROM auth_users WHERE id = ?").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version", "is_active", "email", "name", "profile_picture", "password_hash", "provider", "role_id", "is_verified"}).
			AddRow(1, time.Now(), time.Now(), 1, true, "john@example.com", "John Doe", "profile.jpg", []byte("hashedpassword"), "local", 2, false))

	user, err := model.GetUserByID(1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if user.ID != 1 {
		t.Errorf("expected user ID to be 1, got %d", user.ID)
	}

	// Test Case 2: User ID not found
	mock.ExpectQuery("SELECT .* FROM auth_users WHERE id = ?").
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	_, err = model.GetUserByID(99)
	if err != database.ErrRecordNotFound {
		t.Errorf("expected record not found error, got %v", err)
	}

	// Test Case 3: Database error
	mock.ExpectQuery("SELECT .* FROM auth_users WHERE id = ?").
		WithArgs(99).
		WillReturnError(sql.ErrConnDone)

	_, err = model.GetUserByID(99)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetUserByEmail(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	model := Model{DB: mockDB}

	// Test Case 1: Valid select
	mock.ExpectQuery("SELECT .* FROM auth_users WHERE email = ?").
		WithArgs("john@example.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version", "is_active", "email", "name", "profile_picture", "password_hash", "provider", "role_id", "is_verified"}).
			AddRow(1, time.Now(), time.Now(), 1, true, "john@example.com", "John Doe", "profile.jpg", []byte("hashedpassword"), "local", 2, false))

	user, err := model.GetUserByEmail("john@example.com")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if user.ID != 1 {
		t.Errorf("expected user ID to be 1, got %d", user.ID)
	}

	// Test Case 2: User ID not found
	mock.ExpectQuery("SELECT .* FROM auth_users WHERE email = ?").
		WithArgs("test@email.com").
		WillReturnError(sql.ErrNoRows)

	_, err = model.GetUserByEmail("test@email.com")
	if err != database.ErrRecordNotFound {
		t.Errorf("expected record not found error, got %v", err)
	}

	// Test Case 3: Database error
	mock.ExpectQuery("SELECT .* FROM auth_users WHERE email = ?").
		WithArgs("john@test.com").
		WillReturnError(sql.ErrConnDone)

	_, err = model.GetUserByEmail("john@test.com")
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetUserFromToken(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	model := Model{DB: db}
	tokenScope := ScopeAuthentication
	tokenPlaintext := "testtoken"
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	// Get the current time and a tolerance (e.g., 1 millisecond)
	currentTime := time.Now()

	// Test Case 1: Successful user retrieval
	mock.ExpectQuery(`SELECT au\..* FROM auth_users as au INNER JOIN auth_tokens as at ON au.id = at.user_id WHERE at.hash = \$1 AND at.scope = \$2 AND at.expiry > \$3`).
		WithArgs(tokenHash[:], tokenScope, sqlmock.AnyArg()). // Use sqlmock.AnyArg() for time argument
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "version", "is_active", "email", "name", "profile_picture", "password_hash", "provider", "role_id", "is_verified",
		}).
			AddRow(1, currentTime, currentTime, 1, true, "test@example.com", "Test User", "profile.jpg", "hashedpassword", "provider", 1, true)) // Simulating a user row returned

	user, err := model.GetUserFromToken(tokenScope, tokenPlaintext)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if user == nil {
		t.Errorf("expected user, got nil")
	}

	// Test Case 2: No rows found (token does not exist)
	mock.ExpectQuery(`SELECT au\..* FROM auth_users as au INNER JOIN auth_tokens as at ON au.id = at.user_id WHERE at.hash = \$1 AND at.scope = \$2 AND at.expiry > \$3`).
		WithArgs(tokenHash[:], tokenScope, sqlmock.AnyArg()).
		WillReturnError(sql.ErrNoRows)

	user, err = model.GetUserFromToken(tokenScope, tokenPlaintext)
	if err != database.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}

	// Test Case 3: Database error
	mock.ExpectQuery(`SELECT au\..* FROM auth_users as au INNER JOIN auth_tokens as at ON au.id = at.user_id WHERE at.hash = \$1 AND at.scope = \$2 AND at.expiry > \$3`).
		WithArgs(tokenHash[:], tokenScope, sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)

	user, err = model.GetUserFromToken(tokenScope, tokenPlaintext)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}
	if user != nil {
		t.Errorf("expected nil user, got %v", user)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestUpdateUserByID(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	model := Model{DB: db}

	user := &User{
		ID:             1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Version:        1,
		IsActive:       true,
		Email:          "user@example.com",
		Name:           "Test User",
		ProfilePicture: "profile.jpg",
		Password:       password{hash: []byte("hashedpassword")},
		Provider:       "local",
		RoleID:         2,
		IsVerified:     true,
	}

	// Test Case 1: Valid update
	mock.ExpectQuery(`UPDATE auth_users`).
		WithArgs(user.IsActive, user.Email, user.Name, user.ProfilePicture, user.Password.hash, user.Provider, user.RoleID, user.IsVerified, user.ID, user.Version).
		WillReturnRows(sqlmock.NewRows([]string{"created_at", "updated_at", "version"}).
			AddRow(user.CreatedAt, user.UpdatedAt, user.Version+1))

	err = model.UpdateUserByID(user)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test Case 2: Email fails unique constraint
	mock.ExpectQuery(`UPDATE auth_users`).
		WithArgs(user.IsActive, user.Email, user.Name, user.ProfilePicture, user.Password.hash, user.Provider, user.RoleID, user.IsVerified, user.ID, user.Version).
		WillReturnError(errors.New(`pq: duplicate key value violates unique constraint "auth_users_email_key"`))

	err = model.UpdateUserByID(user)
	if err != database.ErrDuplicateEmail {
		t.Errorf("expected ErrDuplicateEmail, got %v", err)
	}

	// Test Case 3: Edit conflict (no matching row)
	mock.ExpectQuery(`UPDATE auth_users`).
		WithArgs(user.IsActive, user.Email, user.Name, user.ProfilePicture, user.Password.hash, user.Provider, user.RoleID, user.IsVerified, user.ID, user.Version).
		WillReturnError(sql.ErrNoRows)

	err = model.UpdateUserByID(user)
	if err != database.ErrEditConflict {
		t.Errorf("expected ErrEditConflict, got %v", err)
	}

	// Test Case 4: Database error
	mock.ExpectQuery(`UPDATE auth_users`).
		WithArgs(user.IsActive, user.Email, user.Name, user.ProfilePicture, user.Password.hash, user.Provider, user.RoleID, user.IsVerified, user.ID, user.Version).
		WillReturnError(sql.ErrConnDone)

	err = model.UpdateUserByID(user)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDeleteUserByID(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	model := Model{DB: db}

	userID := int64(1)

	// Test Case 1: Valid deletion
	mock.ExpectExec(`DELETE FROM auth_users WHERE id = ?`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

	err = model.DeleteUserByID(userID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test Case 2: Record not found (no rows deleted)
	mock.ExpectExec(`DELETE FROM auth_users WHERE id = ?`).
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // No rows affected

	err = model.DeleteUserByID(userID)
	if err != database.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}

	// Test Case 4: Database error
	mock.ExpectExec(`DELETE FROM auth_users WHERE id = ?`).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	err = model.DeleteUserByID(userID)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestInsertToken(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	model := Model{DB: db}

	token := &Token{
		Hash:   []byte("samplehash"),
		UserID: 1,
		Expiry: time.Now().Add(24 * time.Hour),
		Scope:  "authentication",
	}

	// Test Case 1: Successful token insertion
	mock.ExpectExec(`INSERT INTO auth_tokens \(hash, user_id, expiry, scope\) VALUES \(\$1, \$2, \$3, \$4\)`).
		WithArgs(token.Hash, token.UserID, token.Expiry, token.Scope).
		WillReturnResult(sqlmock.NewResult(1, 1)) // Simulates successful insertion

	err = model.InsertToken(token)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test Case 2: Database error
	mock.ExpectExec(`INSERT INTO auth_tokens \(hash, user_id, expiry, scope\) VALUES \(\$1, \$2, \$3, \$4\)`).
		WithArgs(token.Hash, token.UserID, token.Expiry, token.Scope).
		WillReturnError(sql.ErrConnDone) // Simulate a database connection issue

	err = model.InsertToken(token)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDeleteAllTokensForUser(t *testing.T) {
	// Create a mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	model := Model{DB: db}
	scope := "authentication"
	userID := int64(1)

	// Test Case 1: Successful deletion
	mock.ExpectExec(`DELETE FROM auth_tokens WHERE scope = \$1 AND user_id = \$2`).
		WithArgs(scope, userID).
		WillReturnResult(sqlmock.NewResult(0, 2)) // Simulating 2 rows affected

	err = model.DeleteAllTokensForUser(scope, userID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test Case 2: No rows affected (record not found)
	mock.ExpectExec(`DELETE FROM auth_tokens WHERE scope = \$1 AND user_id = \$2`).
		WithArgs(scope, userID).
		WillReturnResult(sqlmock.NewResult(0, 0)) // Simulating no rows affected

	err = model.DeleteAllTokensForUser(scope, userID)
	if err != database.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}

	// Test Case 3: Database error
	mock.ExpectExec(`DELETE FROM auth_tokens WHERE scope = \$1 AND user_id = \$2`).
		WithArgs(scope, userID).
		WillReturnError(sql.ErrConnDone) // Simulating a database error

	err = model.DeleteAllTokensForUser(scope, userID)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
