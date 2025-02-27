package games

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/navazjm/pixelarcade/internal/webapp/auth"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/database"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/json"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/param"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/response"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/validator"
)

// TODO: Post Game Handler, Implement after permissions implementation
// func (s *Service) PostGameHandler(w http.ResponseWriter, r *http.Request) {}

func (s *Service) GetGamesHandler(w http.ResponseWriter, r *http.Request) {
	games, err := s.Models.GetGames()
	if err != nil {
		response.ServerError(w, r, s.Logger, err)
		return
	}

	err = json.WriteResponse(w, http.StatusOK, json.Envelope{"games": games}, nil)
	if err != nil {
		response.ServerError(w, r, s.Logger, err)
	}
}

func (s *Service) GetGameByIDHandler(w http.ResponseWriter, r *http.Request) {
	gameID, err := param.ReadID(r)
	if err != nil {
		response.NotFound(w, r, s.Logger)
		return
	}

	game, err := s.Models.GetGameByID(gameID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			response.NotFound(w, r, s.Logger)
		default:
			response.ServerError(w, r, s.Logger, err)
		}
		return
	}

	err = json.WriteResponse(w, http.StatusOK, json.Envelope{"game": game}, nil)
	if err != nil {
		response.ServerError(w, r, s.Logger, err)
	}
}

// TODO: Update Game Handler, Implement after permissions implementation
// func (s *Service) UpdateGameByIDHandler(w http.ResponseWriter, r *http.Request) {}

// TODO: Delete Game Handler, Implement after permissions implementation
// func (s *Service) DeleteGameByIDHandler(w http.ResponseWriter, r *http.Request) {}

func (s *Service) PostScoreHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.ContextGetUser(r)

	var reqBody struct {
		Score int64 `json:"score"`
	}

	err := json.ReadRequestBody(w, r, &reqBody)
	if err != nil {
		response.BadRequest(w, r, s.Logger, err)
		return
	}

	v := validator.New()
	if ValidateScore(v, reqBody.Score); !v.Valid() {
		response.FailedValidation(w, r, s.Logger, v.Errors)
		return
	}

	gameID, err := param.ReadID(r)
	if err != nil {
		response.NotFound(w, r, s.Logger)
		return
	}

	game, err := s.Models.GetGameByID(gameID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			response.NotFound(w, r, s.Logger)
		default:
			response.ServerError(w, r, s.Logger, err)
		}
		return
	}

	if !game.HasScore {
		err = errors.New(fmt.Sprintf("%s does not track scores", game.Name))
		response.ServerError(w, r, s.Logger, err)
		return
	}

	score := &Score{
		GameID:   game.ID,
		UserID:   user.ID,
		Score:    reqBody.Score,
		IsActive: true,
	}

	err = s.Models.InsertScore(score)
	if err != nil {
		response.ServerError(w, r, s.Logger, err)
		return
	}

	err = json.WriteResponse(w, http.StatusOK, json.Envelope{"score": score}, nil)
	if err != nil {
		response.ServerError(w, r, s.Logger, err)
	}
}

func (s *Service) GetScoresByGameIDHandler(w http.ResponseWriter, r *http.Request) {
	gameID, err := param.ReadID(r)
	if err != nil {
		response.NotFound(w, r, s.Logger)
		return
	}

	game, err := s.Models.GetGameByID(gameID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			response.NotFound(w, r, s.Logger)
		default:
			response.ServerError(w, r, s.Logger, err)
		}
		return
	}

	if !game.HasScore {
		err = errors.New(fmt.Sprintf("%s does not track scores", game.Name))
		response.ServerError(w, r, s.Logger, err)
		return
	}

	scores, err := s.Models.GetScoresByGameID(gameID)
	if err != nil {
		response.ServerError(w, r, s.Logger, err)
		return
	}

	err = json.WriteResponse(w, http.StatusOK, json.Envelope{"scores": scores}, nil)
	if err != nil {
		response.ServerError(w, r, s.Logger, err)
	}
}

func (s *Service) GetUserScoresByGameIDHandler(w http.ResponseWriter, r *http.Request) {
	user := auth.ContextGetUser(r)
	gameID, err := param.ReadID(r)
	if err != nil {
		response.NotFound(w, r, s.Logger)
		return
	}

	game, err := s.Models.GetGameByID(gameID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			response.NotFound(w, r, s.Logger)
		default:
			response.ServerError(w, r, s.Logger, err)
		}
		return
	}

	if !game.HasScore {
		err = errors.New(fmt.Sprintf("%s does not track scores", game.Name))
		response.ServerError(w, r, s.Logger, err)
		return
	}

	scores, err := s.Models.GetUsersScoresByGameID(gameID, user.ID)
	if err != nil {
		response.ServerError(w, r, s.Logger, err)
		return
	}

	err = json.WriteResponse(w, http.StatusOK, json.Envelope{"scores": scores}, nil)
	if err != nil {
		response.ServerError(w, r, s.Logger, err)
	}
}

// TODO: Update Score Handler, Implement after permissions implementation
// func (s *Service) UpdateScoreByIDHandler(w http.ResponseWriter, r *http.Request) {}

// TODO: Delete Score Handler, Implement after permissions implementation
// func (s *Service) DeleteScoreByIDHandler(w http.ResponseWriter, r *http.Request) {}
