package games

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/database"
)

//==============================================================================
//
// Test CRUD Games
//
//==============================================================================

func TestInsertGame(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	model := Model{DB: mockDB}

	game := &Game{
		Name:        "New Game",
		Description: "A test game",
		Logo:        "logo.png",
		Src:         "src",
		Controls:    "WASD",
		HasScore:    true,
		IsActive:    true,
	}

	// Test Case 1: Successful insertion
	mock.ExpectQuery("INSERT INTO games_list .* RETURNING id, created_at, updated_at, version").
		WithArgs(game.Name, game.Description, game.Logo, game.Src, game.Controls, game.HasScore, game.IsActive).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version"}).
			AddRow(1, time.Now(), time.Now(), 1))

	err = model.InsertGame(game)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if game.ID != 1 {
		t.Errorf("expected game ID to be 1, got %d", game.ID)
	}

	// Test Case 2: Database error during insertion
	mock.ExpectQuery("INSERT INTO games_list .* RETURNING id, created_at, updated_at, version").
		WillReturnError(sql.ErrConnDone)

	err = model.InsertGame(game)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Test Case 3: Row scan error
	mock.ExpectQuery("INSERT INTO games_list .* RETURNING id, created_at, updated_at, version").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version"}).
			AddRow("invalid_id", time.Now(), time.Now(), 1)) // id should be an integer

	err = model.InsertGame(game)
	if err == nil {
		t.Errorf("expected an error due to row scan failure, got nil")
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetGames(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	model := Model{DB: mockDB}

	// Test Case 1: Valid select with multiple games
	mock.ExpectQuery("SELECT .* FROM games_list ORDER BY name").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version", "is_active", "name", "description", "logo", "src", "controls", "has_score"}).
			AddRow(1, time.Now(), time.Now(), 1, true, "Game One", "Description One", "logo1.png", "src1", "WASD", true).
			AddRow(2, time.Now(), time.Now(), 1, true, "Game Two", "Description Two", "logo2.png", "src2", "Arrow Keys", false))

	games, err := model.GetGames()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(games) != 2 {
		t.Errorf("expected 2 games, got %d", len(games))
	}

	if games[0].Name != "Game One" || games[1].Name != "Game Two" {
		t.Errorf("unexpected game names: %v, %v", games[0].Name, games[1].Name)
	}

	// Test Case 2: No games found
	mock.ExpectQuery("SELECT .* FROM games_list ORDER BY name").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version", "is_active", "name", "description", "logo", "src", "controls", "has_score"}))

	games, err = model.GetGames()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(games) != 0 {
		t.Errorf("expected 0 games, got %d", len(games))
	}

	// Test Case 3: Database error
	mock.ExpectQuery("SELECT .* FROM games_list ORDER BY name").
		WillReturnError(sql.ErrConnDone)

	_, err = model.GetGames()
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Test Case 4: Row scan error
	mock.ExpectQuery("SELECT .* FROM games_list ORDER BY name").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version", "is_active", "name", "description", "logo", "src", "controls", "has_score"}).
			AddRow(1, time.Now(), time.Now(), 1, true, "Game One", "Description One", "logo1.png", "src1", "WASD", "invalid_bool")) // has_score should be boolean

	_, err = model.GetGames()
	if err == nil {
		t.Errorf("expected an error due to row scan failure, got nil")
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetGameByID(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	model := Model{DB: mockDB}

	// Test Case 1: Valid game retrieval
	mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version", "is_active", "name", "description", "logo", "src", "controls", "has_score"}).
			AddRow(1, time.Now(), time.Now(), 1, true, "Game One", "Description One", "logo1.png", "src1", "WASD", true))

	game, err := model.GetGameByID(1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if game.ID != 1 {
		t.Errorf("expected game ID to be 1, got %d", game.ID)
	}

	// Test Case 2: Invalid ID (less than 1)
	_, err = model.GetGameByID(0)
	if err != database.ErrRecordNotFound {
		t.Errorf("expected record not found error, got %v", err)
	}

	// Test Case 3: Game ID not found
	mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
		WithArgs(99).
		WillReturnError(sql.ErrNoRows)

	_, err = model.GetGameByID(99)
	if err != database.ErrRecordNotFound {
		t.Errorf("expected record not found error, got %v", err)
	}

	// Test Case 4: Database error
	mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
		WithArgs(2).
		WillReturnError(sql.ErrConnDone)

	_, err = model.GetGameByID(2)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Test Case 5: Row scan error
	mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
		WithArgs(3).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version", "is_active", "name", "description", "logo", "src", "controls", "has_score"}).
			AddRow(3, time.Now(), time.Now(), 1, true, "Game Three", "Description Three", "logo3.png", "src3", "Arrow Keys", "invalid_bool")) // has_score should be boolean

	_, err = model.GetGameByID(3)
	if err == nil {
		t.Errorf("expected an error due to row scan failure, got nil")
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestExistsGameByID(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	model := Model{DB: mockDB}

	// Test Case 1: Game exists
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM games_list WHERE id = \\$1\\)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := model.ExistsGameByID(1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !exists {
		t.Errorf("expected game to exist, got false")
	}

	// Test Case 2: Game does not exist
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM games_list WHERE id = \\$1\\)").
		WithArgs(2).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err = model.ExistsGameByID(2)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if exists {
		t.Errorf("expected game to not exist, got true")
	}

	// Test Case 3: Database error
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM games_list WHERE id = \\$1\\)").
		WithArgs(3).
		WillReturnError(sql.ErrConnDone)

	_, err = model.ExistsGameByID(3)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestUpdateGameByID(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	model := Model{DB: mockDB}

	// Test Case 1: Successful update
	mock.ExpectQuery(regexp.QuoteMeta(`
        UPDATE games_list
        SET name = $1, description = $2, logo = $3, src = $4, controls = $5, has_score = $6, is_active = $7, version = version + 1, updated_at = now()
        WHERE id = $8 and version = $9
        RETURNING updated_at, version`)).
		WithArgs("Updated Game", "Updated Description", "updated_logo.png", "updated_src", "WASD", true, true, 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"updated_at", "version"}).
			AddRow(time.Now(), 2))

	game := &Game{
		ID:          1,
		Name:        "Updated Game",
		Description: "Updated Description",
		Logo:        "updated_logo.png",
		Src:         "updated_src",
		Controls:    "WASD",
		HasScore:    true,
		IsActive:    true,
		Version:     1,
	}

	err = model.UpdateGameByID(game)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test Case 2: Edit conflict (no rows updated)
	mock.ExpectQuery(regexp.QuoteMeta(`
        UPDATE games_list
        SET name = $1, description = $2, logo = $3, src = $4, controls = $5, has_score = $6, is_active = $7, version = version + 1, updated_at = now()
        WHERE id = $8 and version = $9
        RETURNING updated_at, version`)).
		WithArgs("Outdated Game", "Old Description", "old_logo.png", "old_src", "Arrow Keys", false, false, 2, 1).
		WillReturnError(sql.ErrNoRows)

	game = &Game{
		ID:          2,
		Name:        "Outdated Game",
		Description: "Old Description",
		Logo:        "old_logo.png",
		Src:         "old_src",
		Controls:    "Arrow Keys",
		HasScore:    false,
		IsActive:    false,
		Version:     1,
	}

	err = model.UpdateGameByID(game)
	if err != database.ErrEditConflict {
		t.Errorf("expected edit conflict error, got %v", err)
	}

	// Test Case 3: Database error
	mock.ExpectQuery(regexp.QuoteMeta(`
        UPDATE games_list
        SET name = $1, description = $2, logo = $3, src = $4, controls = $5, has_score = $6, is_active = $7, version = version + 1, updated_at = now()
        WHERE id = $8 and version = $9
        RETURNING updated_at, version`)).
		WithArgs("Broken Game", "Broken Description", "broken_logo.png", "broken_src", "None", false, true, 3, 2).
		WillReturnError(sql.ErrConnDone)

	game = &Game{
		ID:          3,
		Name:        "Broken Game",
		Description: "Broken Description",
		Logo:        "broken_logo.png",
		Src:         "broken_src",
		Controls:    "None",
		HasScore:    false,
		IsActive:    true,
		Version:     2,
	}

	err = model.UpdateGameByID(game)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDeleteGameByID(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	model := Model{DB: mockDB}

	// Test Case 1: Successful deletion
	mock.ExpectExec("DELETE FROM games_list WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = model.DeleteGameByID(1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test Case 2: Record not found (no rows deleted)
	mock.ExpectExec("DELETE FROM games_list WHERE id = \\$1").
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = model.DeleteGameByID(2)
	if err != database.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}

	// Test Case 3: Invalid ID (negative or zero)
	err = model.DeleteGameByID(0)
	if err != database.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound for ID 0, got %v", err)
	}

	err = model.DeleteGameByID(-5)
	if err != database.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound for negative ID, got %v", err)
	}

	// Test Case 4: Database error
	mock.ExpectExec("DELETE FROM games_list WHERE id = \\$1").
		WithArgs(3).
		WillReturnError(sql.ErrConnDone)

	err = model.DeleteGameByID(3)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

//==============================================================================
//
// Test CRUD Scores
//
//==============================================================================

func TestInsertScore(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	model := Model{DB: mockDB}

	// Test Case 1: Successful score insertion
	mock.ExpectQuery("INSERT INTO games_scores .* RETURNING id, created_at, updated_at, version").
		WithArgs(1, 42, 1000, true).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version"}).
			AddRow(1, time.Now(), time.Now(), 1))

	score := &Score{GameID: 1, UserID: 42, Score: 1000, IsActive: true}
	err = model.InsertScore(score)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if score.ID != 1 {
		t.Errorf("expected inserted score ID to be 1, got %d", score.ID)
	}

	// Test Case 2: Database error
	mock.ExpectQuery("INSERT INTO games_scores .* RETURNING id, created_at, updated_at, version").
		WithArgs(2, 99, 2000, true).
		WillReturnError(sql.ErrConnDone)

	score = &Score{GameID: 2, UserID: 99, Score: 2000, IsActive: true}
	err = model.InsertScore(score)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Test Case 3: Row scan error
	mock.ExpectQuery("INSERT INTO games_scores .* RETURNING id, created_at, updated_at, version").
		WithArgs(3, 88, -500, true). // Invalid negative score (business logic might reject this)
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version"}).
			AddRow(nil, time.Now(), time.Now(), 1)) // Causes scan error

	score = &Score{GameID: 3, UserID: 88, Score: -500, IsActive: true}
	err = model.InsertScore(score)
	if err == nil {
		t.Errorf("expected an error due to row scan failure, got nil")
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetScoresByGameID(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	model := &Model{DB: mockDB}

	// Test Case 1: Successful retrieval
	mock.ExpectQuery("SELECT s.*, u.* FROM games_scores s JOIN auth_user u ON s.user_id = u.uid WHERE s.game_id = \\$1 ORDER BY s.score DESC LIMIT 50").
		WithArgs(10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "game_id", "user_id", "name", "profile_picture", "score", "created_at", "updated_at", "version"}).
			AddRow(1, 10, 42, "Alice", "alice.png", 5000, time.Now(), time.Now(), 1).
			AddRow(2, 10, 43, "Bob", "bob.png", 3000, time.Now(), time.Now(), 1))

	scores, err := model.GetScoresByGameID(10)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(scores) != 2 {
		t.Errorf("expected 2 scores, got %d", len(scores))
	}

	if scores[0].UserName != "Alice" || scores[1].UserName != "Bob" {
		t.Errorf("unexpected score order or data mismatch")
	}

	// Test Case 2: No scores found (empty result set)
	mock.ExpectQuery("SELECT .* FROM games_scores .* WHERE s.game_id = \\$1 ORDER BY s.score DESC LIMIT 50").
		WithArgs(999).
		WillReturnRows(sqlmock.NewRows([]string{})) // No rows returned

	scores, err = model.GetScoresByGameID(999)
	if err != nil {
		t.Errorf("expected no error for empty result, got %v", err)
	}
	if len(scores) != 0 {
		t.Errorf("expected empty scores, got %d", len(scores))
	}

	// Test Case 3: Database error
	mock.ExpectQuery("SELECT .* FROM games_scores .* WHERE s.game_id = \\$1 ORDER BY s.score DESC LIMIT 50").
		WithArgs(20).
		WillReturnError(sql.ErrConnDone)

	scores, err = model.GetScoresByGameID(20)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}
	if scores != nil {
		t.Errorf("expected nil scores on error, got %v", scores)
	}

	// Test Case 4: Row scan error (corrupted data)
	mock.ExpectQuery("SELECT .* FROM games_scores .* WHERE s.game_id = \\$1 ORDER BY s.score DESC LIMIT 50").
		WithArgs(30).
		WillReturnRows(sqlmock.NewRows([]string{"id", "game_id", "user_id", "name", "profile_picture", "score", "created_at", "updated_at", "version"}).
			AddRow(nil, 30, 99, "Charlie", "charlie.png", 1500, time.Now(), time.Now(), 1)) // `id` is nil, causing scan error

	scores, err = model.GetScoresByGameID(30)
	if err == nil {
		t.Errorf("expected an error due to row scan failure, got nil")
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetUsersScoresByGameID(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	model := &Model{DB: mockDB}

	gameID := int64(10)
	userID := int64(42)

	// Test Case 1: Successful retrieval
	mock.ExpectQuery("SELECT s.*, u.name, u.profile_picture FROM games_scores s JOIN auth_users u ON s.user_id = u.uid WHERE s.game_id = \\$1 and s.user_id = \\$2 ORDER BY s.score DESC").
		WithArgs(gameID, userID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version", "is_active", "game_id", "user_id", "score", "name", "profile_picture"}).
			AddRow(1, time.Now(), time.Now(), 1, true, gameID, userID, 5000, "Alice", "alice.png").
			AddRow(2, time.Now(), time.Now(), 1, true, gameID, userID, 4000, "Alice", "alice.png"))

	scores, err := model.GetUsersScoresByGameID(gameID, userID)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(scores) != 2 {
		t.Errorf("expected 2 scores, got %d", len(scores))
	}

	if scores[0].Score != 5000 || scores[1].Score != 4000 {
		t.Errorf("unexpected score values")
	}

	// Test Case 2: No scores found (empty result set)
	mock.ExpectQuery("SELECT .* FROM games_scores .* WHERE s.game_id = \\$1 and s.user_id = \\$2 ORDER BY s.score DESC").
		WithArgs(999, 999).
		WillReturnRows(sqlmock.NewRows([]string{})) // No rows returned

	scores, err = model.GetUsersScoresByGameID(999, 999)
	if err != nil {
		t.Errorf("expected no error for empty result, got %v", err)
	}
	if len(scores) != 0 {
		t.Errorf("expected empty scores, got %d", len(scores))
	}

	// Test Case 3: Database error
	mock.ExpectQuery("SELECT .* FROM games_scores .* WHERE s.game_id = \\$1 and s.user_id = \\$2 ORDER BY s.score DESC").
		WithArgs(20, 50).
		WillReturnError(sql.ErrConnDone)

	scores, err = model.GetUsersScoresByGameID(20, 50)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}
	if scores != nil {
		t.Errorf("expected nil scores on error, got %v", scores)
	}

	// Test Case 4: Row scan error (corrupted data)
	mock.ExpectQuery("SELECT .* FROM games_scores .* WHERE s.game_id = \\$1 and s.user_id = \\$2 ORDER BY s.score DESC").
		WithArgs(30, 60).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version", "is_active", "game_id", "user_id", "score", "name", "profile_picture"}).
			AddRow(nil, time.Now(), time.Now(), 1, true, 30, 60, 3500, "Charlie", "charlie.png")) // `id` is nil, causing scan error

	scores, err = model.GetUsersScoresByGameID(30, 60)
	if err == nil {
		t.Errorf("expected an error due to row scan failure, got nil")
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestExistsScoreByID(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	model := Model{DB: mockDB}

	// Test Case 1: Score exists
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM games_scores WHERE id = \\$1\\)").
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := model.ExistsScoreByID(1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !exists {
		t.Errorf("expected true, got false")
	}

	// Test Case 2: Score does not exist
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM games_scores WHERE id = \\$1\\)").
		WithArgs(999).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err = model.ExistsScoreByID(999)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if exists {
		t.Errorf("expected false, got true")
	}

	// Test Case 3: Database error
	mock.ExpectQuery("SELECT EXISTS\\(SELECT 1 FROM games_scores WHERE id = \\$1\\)").
		WithArgs(2).
		WillReturnError(sql.ErrConnDone)

	exists, err = model.ExistsScoreByID(2)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}
	if exists {
		t.Errorf("expected false on error, got true")
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestUpdateScoreByID(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	model := Model{DB: mockDB}

	// Test Case 1: Successful update
	mock.ExpectQuery(regexp.QuoteMeta(`
        UPDATE games_scores
        SET score = $1, is_active = $2, version = version + 1, updated_at = now()
        WHERE id = $3 and version = $4
        RETURNING updated_at, version`)).
		WithArgs(100, false, 1, 1).
		WillReturnRows(sqlmock.NewRows([]string{"updated_at", "version"}).
			AddRow(time.Now(), 2))

	score := &Score{
		ID:       1,
		Score:    100,
		IsActive: false,
		Version:  1,
	}

	err = model.UpdateScoreByID(score)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test Case 2: Edit conflict (no rows updated)
	mock.ExpectQuery(regexp.QuoteMeta(`
        UPDATE games_scores
        SET score = $1, is_active = $2, version = version + 1, updated_at = now()
        WHERE id = $3 and version = $4
        RETURNING updated_at, version`)).
		WithArgs(100, false, 1, 1).
		WillReturnError(sql.ErrNoRows)

	score = &Score{
		ID:       1,
		Score:    100,
		IsActive: false,
		Version:  1,
	}

	err = model.UpdateScoreByID(score)
	if err != database.ErrEditConflict {
		t.Errorf("expected edit conflict error, got %v", err)
	}

	// Test Case 3: Database error
	mock.ExpectQuery(regexp.QuoteMeta(`
        UPDATE games_scores
        SET score = $1, is_active = $2, version = version + 1, updated_at = now()
        WHERE id = $3 and version = $4
        RETURNING updated_at, version`)).
		WithArgs(100, false, 3, 2).
		WillReturnError(sql.ErrConnDone)

	score = &Score{
		ID:       3,
		Score:    100,
		IsActive: false,
		Version:  2,
	}

	err = model.UpdateScoreByID(score)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestDeleteScoreByID(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock database: %v", err)
	}
	defer mockDB.Close()

	model := Model{DB: mockDB}

	// Test Case 1: Successful deletion
	mock.ExpectExec("DELETE FROM games_scores WHERE id = \\$1").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = model.DeleteScoreByID(1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Test Case 2: Record not found (no rows deleted)
	mock.ExpectExec("DELETE FROM games_scores WHERE id = \\$1").
		WithArgs(2).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = model.DeleteScoreByID(2)
	if err != database.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}

	// Test Case 3: Invalid ID (negative or zero)
	err = model.DeleteGameByID(0)
	if err != database.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound for ID 0, got %v", err)
	}

	err = model.DeleteScoreByID(-5)
	if err != database.ErrRecordNotFound {
		t.Errorf("expected ErrRecordNotFound for negative ID, got %v", err)
	}

	// Test Case 4: Database error
	mock.ExpectExec("DELETE FROM games_scores WHERE id = \\$1").
		WithArgs(3).
		WillReturnError(sql.ErrConnDone)

	err = model.DeleteScoreByID(3)
	if err != sql.ErrConnDone {
		t.Errorf("expected sql.ErrConnDone, got %v", err)
	}

	// Ensure all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}
