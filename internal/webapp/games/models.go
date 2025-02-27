package games

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/navazjm/pixelarcade/internal/webapp/utils/database"
)

type Model struct {
	DB *sql.DB
}

//==============================================================================
//
// CRUD Games
//
//==============================================================================

func (m Model) InsertGame(game *Game) error {
	query := `
        INSERT INTO games_list (name, description, logo, src, controls, has_score, is_active) 
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, updated_at, version`
	args := []any{game.Name, game.Description, game.Logo, game.Src, game.Controls, game.HasScore, game.IsActive}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&game.ID, &game.CreatedAt, &game.UpdatedAt, &game.Version)
}

func (m Model) GetGames() ([]*Game, error) {
	query := `
        SELECT * 
        FROM games_list
		ORDER BY name`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	games := []*Game{}
	for rows.Next() {
		var game Game
		err := rows.Scan(
			&game.ID,
			&game.CreatedAt,
			&game.UpdatedAt,
			&game.Version,
			&game.IsActive,
			&game.Name,
			&game.Description,
			&game.Logo,
			&game.Src,
			&game.Controls,
			&game.HasScore,
		)
		if err != nil {
			return nil, err
		}
		games = append(games, &game)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return games, nil
}

func (m Model) GetGameByID(id int64) (*Game, error) {
	if id < 1 {
		return nil, database.ErrRecordNotFound
	}

	query := `
        SELECT *
        FROM games_list
        WHERE id = $1`

	var game Game

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&game.ID,
		&game.CreatedAt,
		&game.UpdatedAt,
		&game.Version,
		&game.IsActive,
		&game.Name,
		&game.Description,
		&game.Logo,
		&game.Src,
		&game.Controls,
		&game.HasScore,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, database.ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &game, nil
}

func (m Model) ExistsGameByID(id int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM games_list WHERE id = $1)`
	var exists bool
	err := m.DB.QueryRow(query, id).Scan(&exists)
	return exists, err
}

func (m Model) UpdateGameByID(game *Game) error {
	query := `
        UPDATE games_list
        SET name = $1, description = $2, logo = $3, src = $4, controls = $5, has_score = $6, is_active = $7, version = version + 1, updated_at = now()
        WHERE id = $8 and version = $9
        RETURNING updated_at, version`

	args := []any{
		game.Name,
		game.Description,
		game.Logo,
		game.Src,
		game.Controls,
		game.HasScore,
		game.IsActive,
		game.ID,
		game.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&game.UpdatedAt, &game.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return database.ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m Model) DeleteGameByID(id int64) error {
	if id < 1 {
		return database.ErrRecordNotFound
	}

	query := `
        DELETE FROM games_list
        WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
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

//==============================================================================
//
// CRUD Scores
//
//==============================================================================

func (m Model) InsertScore(score *Score) error {
	query := `
        INSERT INTO games_scores (game_id, user_id, score, is_active)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, updated_at, version`
	args := []any{score.GameID, score.UserID, score.Score, score.IsActive}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&score.ID, &score.CreatedAt, &score.UpdatedAt, &score.Version)
}

func (m *Model) GetScoresByGameID(gameID int64) ([]*Score, error) {
	query := `
        SELECT s.id, s.game_id, s.user_id, u.name, u.profile_picture, s.score, s.created_at, s.updated_at, s.version
        FROM games_scores s
        JOIN auth_user u ON s.user_id = u.uid
        WHERE s.game_id = $1
        ORDER BY s.score DESC
        LIMIT 50`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scores := []*Score{}
	for rows.Next() {
		var score Score
		err := rows.Scan(
			&score.ID,
			&score.GameID,
			&score.UserID,
			&score.UserName,
			&score.UserProfilePicture,
			&score.Score,
			&score.CreatedAt,
			&score.UpdatedAt,
			&score.Version,
		)
		if err != nil {
			return nil, err
		}
		scores = append(scores, &score)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return scores, nil
}

func (m *Model) GetUsersScoresByGameID(gameID int64, userID int64) ([]*Score, error) {
	query := `
        SELECT s.*, u.name, u.profile_picture
        FROM games_scores s
        JOIN auth_users u ON s.user_id = u.uid
        WHERE s.game_id = $1 and s.user_id = $2
        ORDER BY s.score DESC`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, gameID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scores := []*Score{}
	for rows.Next() {
		var score Score
		err := rows.Scan(
			&score.ID,
			&score.CreatedAt,
			&score.UpdatedAt,
			&score.Version,
			&score.IsActive,
			&score.GameID,
			&score.UserID,
			&score.Score,
			&score.UserName,
			&score.UserProfilePicture,
		)
		if err != nil {
			return nil, err
		}
		scores = append(scores, &score)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return scores, nil
}

func (m Model) ExistsScoreByID(id int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM games_scores WHERE id = $1)`
	var exists bool
	err := m.DB.QueryRow(query, id).Scan(&exists)
	return exists, err
}

func (m Model) UpdateScoreByID(score *Score) error {
	query := `
        UPDATE games_scores
        SET score = $1, is_active = $2, version = version + 1, updated_at = now() 
        WHERE id = $3 and version = $4
        RETURNING updated_at, version`

	args := []any{score.Score, score.IsActive, score.ID, score.Version}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&score.UpdatedAt, &score.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return database.ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m Model) DeleteScoreByID(id int64) error {
	if id < 1 {
		return database.ErrRecordNotFound
	}

	query := `
        DELETE FROM games_scores
        WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
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
