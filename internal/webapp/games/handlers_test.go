package games

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/navazjm/pixelarcade/internal/webapp/auth"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/database"
	pa_json "github.com/navazjm/pixelarcade/internal/webapp/utils/json"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/param"

	"github.com/DATA-DOG/go-sqlmock"
)

// func TestPostGameHandler(t *testing.T) {}

func TestGetGamesHandler(t *testing.T) {
	t.Run("SUCCESS Retrieved games list", func(t *testing.T) {
		service, mock := newMockService(t)
		now := time.Now()

		rows := sqlmock.NewRows([]string{
			"id", "created_at", "updated_at", "version", "is_active",
			"name", "description", "logo", "src", "controls", "has_score",
		}).
			AddRow(1, now, now, 1, true, "Game One", "Desc One", "logo1.png", "src1", "controls1", true).
			AddRow(2, now, now, 1, true, "Game Two", "Desc Two", "logo2.png", "src2", "controls2", false)

		mock.ExpectQuery("SELECT .* FROM games_list ORDER BY name").WillReturnRows(rows)

		req := httptest.NewRequest(http.MethodGet, "/api/games", nil)
		w := httptest.NewRecorder()

		service.GetGamesHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var response pa_json.Envelope
		err := json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		games, ok := response["games"].([]interface{})
		if !ok || len(games) != 2 {
			t.Errorf("expected 2 games, got %v", games)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %v", err)
		}
	})

	t.Run("ERROR DB error fetching games list", func(t *testing.T) {
		service, mock := newMockService(t)
		mock.ExpectQuery("SELECT .* FROM games_list ORDER BY name").
			WillReturnError(database.ErrMockDatabase)

		req := httptest.NewRequest(http.MethodGet, "/api/games", nil)
		w := httptest.NewRecorder()

		service.GetGamesHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %v", err)
		}
	})
}

// Test successful retrieval of a game by ID
func TestGetGameByIDHandler(t *testing.T) {
	gameID := int64(1)
	endpoint := fmt.Sprintf("/api/games/%d", gameID)

	t.Run("SUCCESS Retrieved game by ID", func(t *testing.T) {
		service, mock := newMockService(t)
		now := time.Now()

		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "created_at", "updated_at", "version", "is_active",
				"name", "description", "logo", "src", "controls", "has_score",
			}).AddRow(
				gameID, now, now, 1, true, "Game One", "Description One", "logo1.png", "src1", "controls1", true,
			))

		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req = param.InjectID(req, gameID) // Simulate extracting ID from URL
		w := httptest.NewRecorder()

		service.GetGameByIDHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var response pa_json.Envelope
		err := json.NewDecoder(resp.Body).Decode(&response)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		game, ok := response["game"].(map[string]interface{})
		if !ok || int64(game["id"].(float64)) != gameID {
			t.Errorf("expected game ID %d, got %v", gameID, game)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %v", err)
		}
	})

	t.Run("ERROR Param read game ID", func(t *testing.T) {
		service, _ := newMockService(t)
		req := httptest.NewRequest(http.MethodGet, "/api/games/abc", nil)
		w := httptest.NewRecorder()

		service.GetGameByIDHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}

	})

	t.Run("ERROR DB error game not found", func(t *testing.T) {
		service, mock := newMockService(t)
		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnError(database.ErrRecordNotFound)

		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req = param.InjectID(req, gameID)
		w := httptest.NewRecorder()

		service.GetGameByIDHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %v", err)
		}
	})

	t.Run("ERROR DB error fetching game by ID", func(t *testing.T) {
		service, mock := newMockService(t)
		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnError(database.ErrMockDatabase)

		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req = param.InjectID(req, gameID)
		w := httptest.NewRecorder()

		service.GetGameByIDHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %v", err)
		}
	})
}

// func TestUpdateGameByIDHandler(t *testing.T) {}

// func TestDeleteGameByIDHandler(t *testing.T) {}

func TestPostScoreHandler(t *testing.T) {
	gameID := int64(1)
	endpoint := fmt.Sprintf("/api/games/%d/scores", gameID)

	t.Run("SUCCESS Score inserted", func(t *testing.T) {
		service, mock := newMockService(t)
		user := &auth.User{ID: 1}
		reqBody := `{"score": 100}`
		now := time.Now()
		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "created_at", "updated_at", "version", "is_active",
				"name", "description", "logo", "src", "controls", "has_score",
			}).AddRow(
				gameID, now, now, 1, true, "Game One", "Description One", "logo1.png", "src1", "controls1", true,
			))

		mock.ExpectQuery("INSERT INTO games_scores").
			WithArgs(gameID, user.ID, 100, true).
			WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version"}).
				AddRow(1, now, now, 1))

		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req = param.InjectID(req, gameID)
		req = auth.ContextSetUser(req, user)
		w := httptest.NewRecorder()

		service.PostScoreHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %v", err)
		}
	})

	t.Run("ERROR Invalid request body", func(t *testing.T) {
		service, _ := newMockService(t)
		user := &auth.User{ID: 1}
		reqBody := `{"hello": "world"}`
		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req = auth.ContextSetUser(req, user)
		w := httptest.NewRecorder()

		service.PostScoreHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, resp.StatusCode)
		}
	})

	t.Run("ERROR Failed score validation check", func(t *testing.T) {
		service, _ := newMockService(t)
		user := &auth.User{ID: 1}
		reqBody := `{"score": -10}` // Invalid score
		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req = auth.ContextSetUser(req, user)
		w := httptest.NewRecorder()

		service.PostScoreHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, resp.StatusCode)
		}
	})

	t.Run("ERROR Param read game ID", func(t *testing.T) {
		service, _ := newMockService(t)
		user := &auth.User{ID: 1}
		reqBody := `{"score": 50}`
		req := httptest.NewRequest(http.MethodPost, "/api/games/invalid/score", strings.NewReader(reqBody))
		req = auth.ContextSetUser(req, user)
		w := httptest.NewRecorder()

		service.PostScoreHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("ERROR DB error game not found", func(t *testing.T) {
		service, mock := newMockService(t)
		user := &auth.User{ID: 1}
		reqBody := `{"score": 50}`
		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnError(database.ErrRecordNotFound)

		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req = param.InjectID(req, gameID)
		req = auth.ContextSetUser(req, user)
		w := httptest.NewRecorder()

		service.PostScoreHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %v", err)
		}
	})

	t.Run("ERROR DB error fetching game by ID", func(t *testing.T) {
		service, mock := newMockService(t)
		user := &auth.User{ID: 1}
		reqBody := `{"score": 50}`
		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnError(database.ErrMockDatabase)

		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req = param.InjectID(req, gameID)
		req = auth.ContextSetUser(req, user)
		w := httptest.NewRecorder()

		service.PostScoreHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %v", err)
		}

	})

	t.Run("ERROR Game does not track score", func(t *testing.T) {
		service, mock := newMockService(t)
		user := &auth.User{ID: 1}
		reqBody := `{"score": 100}`
		now := time.Now()
		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "created_at", "updated_at", "version", "is_active",
				"name", "description", "logo", "src", "controls", "has_score",
			}).AddRow(
				gameID, now, now, 1, true, "Game One", "Description One", "logo1.png", "src1", "controls1", false,
			))

		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req = param.InjectID(req, gameID)
		req = auth.ContextSetUser(req, user)
		w := httptest.NewRecorder()

		service.PostScoreHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})

	t.Run("ERROR Database error inserting score", func(t *testing.T) {
		service, mock := newMockService(t)
		user := &auth.User{ID: 1}
		reqBody := `{"score": 100}`
		now := time.Now()
		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "created_at", "updated_at", "version", "is_active",
				"name", "description", "logo", "src", "controls", "has_score",
			}).AddRow(
				gameID, now, now, 1, true, "Game One", "Description One", "logo1.png", "src1", "controls1", true,
			))

		mock.ExpectExec("INSERT INTO scores").
			WithArgs(gameID, user.ID, 100).
			WillReturnError(database.ErrMockDatabase)

		req := httptest.NewRequest(http.MethodPost, endpoint, strings.NewReader(reqBody))
		req = param.InjectID(req, gameID)
		req = auth.ContextSetUser(req, user)
		w := httptest.NewRecorder()

		service.PostScoreHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})
}

func TestGetScoresByGameIDHandler(t *testing.T) {
	gameID := int64(1)
	endpoint := fmt.Sprintf("/api/games/%d/scores", gameID)

	t.Run("SUCCESS Retrieved scores by game ID", func(t *testing.T) {
		service, mock := newMockService(t)
		now := time.Now()
		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "created_at", "updated_at", "version", "is_active",
				"name", "description", "logo", "src", "controls", "has_score",
			}).AddRow(
				gameID, now, now, 1, true, "Game One", "Description One", "logo1.png", "src1", "controls1", true,
			))

		mock.ExpectQuery("SELECT s.id, s.game_id, s.user_id, u.name, u.profile_picture, s.score, s.created_at, s.updated_at, s.version FROM games_scores s JOIN auth_user u ON s.user_id = u.uid WHERE s.game_id = \\$1").
			WithArgs(gameID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "game_id", "user_id", "name", "profile_picture", "score", "created_at", "updated_at", "version",
			}).AddRow(
				1, gameID, 2, "User One", "profile1.jpg", 500, now, now, 1,
			).AddRow(
				2, gameID, 3, "User Two", "profile2.jpg", 400, now, now, 1,
			))

		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req = param.InjectID(req, gameID)
		w := httptest.NewRecorder()

		service.GetScoresByGameIDHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var jsonResponse struct {
			Scores []Score `json:"scores"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&jsonResponse); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(jsonResponse.Scores) != 2 {
			t.Errorf("expected 2 scores, got %d", len(jsonResponse.Scores))
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %v", err)
		}
	})

	t.Run("ERROR Param read game ID", func(t *testing.T) {
		service, _ := newMockService(t)
		req := httptest.NewRequest(http.MethodGet, "/api/games/invalid/scores", nil)
		w := httptest.NewRecorder()

		service.GetScoresByGameIDHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("ERROR DB error game not found", func(t *testing.T) {
		service, mock := newMockService(t)
		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnError(database.ErrRecordNotFound) // Simulating a missing game

		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req = param.InjectID(req, gameID)
		w := httptest.NewRecorder()

		service.GetScoresByGameIDHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %v", err)
		}
	})

	t.Run("ERROR DB error fetching game by ID", func(t *testing.T) {
		service, mock := newMockService(t)
		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnError(database.ErrMockDatabase)

		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req = param.InjectID(req, gameID)
		w := httptest.NewRecorder()

		service.GetScoresByGameIDHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %v", err)
		}

	})

	t.Run("ERROR Game does not track scores", func(t *testing.T) {
		service, mock := newMockService(t)
		now := time.Now()
		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "created_at", "updated_at", "version", "is_active",
				"name", "description", "logo", "src", "controls", "has_score",
			}).AddRow(
				gameID, now, now, 1, true, "Game One", "Description One", "logo1.png", "src1", "controls1", false, // `has_score` is false
			))

		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req = param.InjectID(req, gameID)
		w := httptest.NewRecorder()

		service.GetScoresByGameIDHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %v", err)
		}
	})

	t.Run("ERROR DB error fetching scores by game ID", func(t *testing.T) {
		service, mock := newMockService(t)
		now := time.Now()
		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "created_at", "updated_at", "version", "is_active",
				"name", "description", "logo", "src", "controls", "has_score",
			}).AddRow(
				gameID, now, now, 1, true, "Game One", "Description One", "logo1.png", "src1", "controls1", true,
			))

		mock.ExpectQuery("SELECT s.id, s.game_id, s.user_id, u.name, u.profile_picture, s.score, s.created_at, s.updated_at, s.version FROM games_scores s JOIN auth_user u ON s.user_id = u.uid WHERE s.game_id = \\$1").
			WithArgs(gameID).
			WillReturnError(database.ErrMockDatabase)

		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req = param.InjectID(req, gameID)
		w := httptest.NewRecorder()

		service.GetScoresByGameIDHandler(w, req)

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %v", err)
		}
	})
}

func TestGetUserScoresByGameIDHandler(t *testing.T) {
	gameID := int64(1)
	endpoint := fmt.Sprintf("/api/games/%d/scores/user", gameID)

	t.Run("SUCCESS Retrieved user scores", func(t *testing.T) {
		service, mock := newMockService(t)
		user := &auth.User{ID: 1}
		now := time.Now()
		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req = param.InjectID(req, gameID)
		req = auth.ContextSetUser(req, user)
		respRecorder := httptest.NewRecorder()

		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "created_at", "updated_at", "version", "is_active",
				"name", "description", "logo", "src", "controls", "has_score",
			}).AddRow(
				gameID, now, now, 1, true, "Game One", "Description One", "logo1.png", "src1", "controls1", true,
			))

		mock.ExpectQuery("SELECT .* FROM games_scores").
			WithArgs(gameID, user.ID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "created_at", "updated_at", "version", "is_active", "game_id", "user_id", "score", "name", "profile_picture",
			}).AddRow(
				1, now, now, 1, true, gameID, user.ID, 200, "User One", "profile1.png",
			).AddRow(
				2, now, now, 1, true, gameID, user.ID, 150, "User One", "profile1.png",
			))

		service.GetUserScoresByGameIDHandler(respRecorder, req)

		resp := respRecorder.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unmet expectations: %v", err)
		}
	})

	t.Run("ERROR Param read game ID", func(t *testing.T) {
		service, _ := newMockService(t)
		user := &auth.User{ID: 1}
		req := httptest.NewRequest(http.MethodGet, "/api/games/invalid/user/scores", nil)
		req = auth.ContextSetUser(req, user)
		respRecorder := httptest.NewRecorder()

		service.GetUserScoresByGameIDHandler(respRecorder, req)

		resp := respRecorder.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("ERROR DB error game not found", func(t *testing.T) {
		service, mock := newMockService(t)
		user := &auth.User{ID: 1}
		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req = param.InjectID(req, gameID)
		req = auth.ContextSetUser(req, user)
		respRecorder := httptest.NewRecorder()

		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnError(database.ErrRecordNotFound)

		service.GetUserScoresByGameIDHandler(respRecorder, req)

		resp := respRecorder.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
		}
	})

	t.Run("ERROR DB error when fetching game by ID", func(t *testing.T) {
		service, mock := newMockService(t)
		user := &auth.User{ID: 1}
		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req = param.InjectID(req, gameID)
		req = auth.ContextSetUser(req, user)
		respRecorder := httptest.NewRecorder()

		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnError(database.ErrMockDatabase)

		service.GetUserScoresByGameIDHandler(respRecorder, req)

		resp := respRecorder.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})

	t.Run("ERROR Game does not track scores", func(t *testing.T) {
		service, mock := newMockService(t)
		user := &auth.User{ID: 1}
		now := time.Now()
		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req = param.InjectID(req, gameID)
		req = auth.ContextSetUser(req, user)
		respRecorder := httptest.NewRecorder()

		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "created_at", "updated_at", "version", "is_active",
				"name", "description", "logo", "src", "controls", "has_score",
			}).AddRow(
				gameID, now, now, 1, true, "Game One", "Description One", "logo1.png", "src1", "controls1", false,
			))

		service.GetUserScoresByGameIDHandler(respRecorder, req)

		resp := respRecorder.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})

	t.Run("ERROR DB error when fetching user scores", func(t *testing.T) {
		service, mock := newMockService(t)
		user := &auth.User{ID: 1}
		now := time.Now()
		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req = param.InjectID(req, gameID)
		req = auth.ContextSetUser(req, user)
		respRecorder := httptest.NewRecorder()

		mock.ExpectQuery("SELECT .* FROM games_list WHERE id = \\$1").
			WithArgs(gameID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "created_at", "updated_at", "version", "is_active",
				"name", "description", "logo", "src", "controls", "has_score",
			}).AddRow(
				gameID, now, now, 1, true, "Game One", "Description One", "logo1.png", "src1", "controls1", true,
			))

		mock.ExpectQuery("SELECT .* FROM games_scores").
			WithArgs(gameID, user.ID).
			WillReturnError(database.ErrMockDatabase)

		service.GetUserScoresByGameIDHandler(respRecorder, req)

		resp := respRecorder.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})
}
