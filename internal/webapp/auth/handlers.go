package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/navazjm/pixelarcade/internal/webapp/utils/database"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/json"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/response"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/validator"
)

func (as *Service) RegisterNewUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.ReadRequestBody(w, r, &input)
	if err != nil {
		response.BadRequest(w, r, as.Logger, err)
		return
	}

	user := &User{
		Email:          input.Email,
		IsActive:       true,
		IsVerified:     false,
		Name:           input.Name,
		ProfilePicture: defaultProfilePicture,
		Provider:       "N/A",
		RoleID:         RoleBasic,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		response.ServerError(w, r, as.Logger, err)
		return
	}

	v := validator.New()

	if ValidateUser(v, user); !v.Valid() {
		response.FailedValidation(w, r, as.Logger, v.Errors)
		return
	}

	err = as.Models.InsertUser(user)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			response.FailedValidation(w, r, as.Logger, v.Errors)
		default:
			response.ServerError(w, r, as.Logger, err)
		}
		return
	}

	err = json.WriteResponse(w, http.StatusCreated, json.Envelope{"user": user}, nil)
	if err != nil {
		response.ServerError(w, r, as.Logger, err)
	}
}

func (as *Service) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.ReadRequestBody(w, r, &input)
	if err != nil {
		response.BadRequest(w, r, as.Logger, err)
		return
	}

	v := validator.New()
	ValidateEmail(v, input.Email)
	ValidatePasswordPlaintextEmpty(v, input.Password)
	if !v.Valid() {
		response.FailedValidation(w, r, as.Logger, v.Errors)
		return
	}

	user, err := as.Models.GetUserByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrRecordNotFound):
			response.InvalidCredentials(w, r, as.Logger)
		default:
			response.ServerError(w, r, as.Logger, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		response.ServerError(w, r, as.Logger, err)
		return
	}

	if !match {
		response.InvalidCredentials(w, r, as.Logger)
		return
	}

	token, err := as.Models.NewToken(user.ID, 24*time.Hour, ScopeAuthentication)
	if err != nil {
		response.ServerError(w, r, as.Logger, err)
		return
	}

	// Set token as an HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     CookieAuthToken,
		Value:    token.Plaintext,
		HttpOnly: true,
		Secure:   false, // Use true for HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		Expires:  token.Expiry,
	})

	err = json.WriteResponse(w, http.StatusCreated, json.Envelope{"user": user}, nil)
	if err != nil {
		response.ServerError(w, r, as.Logger, err)
	}
}

func (as *Service) GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	user := ContextGetUser(r)

	err := json.WriteResponse(w, http.StatusOK, json.Envelope{"user": user}, nil)
	if err != nil {
		response.ServerError(w, r, as.Logger, err)
	}
}

func (as *Service) UpdateCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	user := ContextGetUser(r)

	var input struct {
		Email          *string `json:"email"`
		Name           *string `json:"name"`
		ProfilePicture *string `json:"profile_picture"`
		Password       *string `json:"password"`
		Provider       *string `json:"provider"`
		RoleID         *RoleID `json:"role_id"`
		IsActive       *bool   `json:"is_active"`
	}

	err := json.ReadRequestBody(w, r, &input)
	if err != nil {
		response.BadRequest(w, r, as.Logger, err)
		return
	}

	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Name != nil {
		user.Name = *input.Name
	}
	if input.ProfilePicture != nil {
		user.ProfilePicture = *input.ProfilePicture
	}
	if input.RoleID != nil {
		user.RoleID = *input.RoleID
	}
	if input.Password != nil {
		err = user.Password.Set(*input.Password)
		if err != nil {
			response.ServerError(w, r, as.Logger, err)
			return
		}
	}
	if input.Provider != nil {
		user.Provider = *input.Provider
	}
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
	}

	v := validator.New()
	if ValidateUser(v, user); !v.Valid() {
		response.FailedValidation(w, r, as.Logger, v.Errors)
		return
	}

	err = as.Models.UpdateUserByID(user)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			response.FailedValidation(w, r, as.Logger, v.Errors)
		default:
			response.ServerError(w, r, as.Logger, err)
		}
		return
	}

	err = json.WriteResponse(w, http.StatusOK, json.Envelope{"user": user}, nil)
	if err != nil {
		response.ServerError(w, r, as.Logger, err)
	}
}

func (as *Service) LogoutUserHandler(w http.ResponseWriter, r *http.Request) {
	user := ContextGetUser(r)
	err := as.Models.DeleteAllTokensForUser(ScopeAuthentication, user.ID)
	if err != nil {
		response.ServerError(w, r, as.Logger, err)
		return
	}

	// Set the cookie with an expired date to remove it
	http.SetCookie(w, &http.Cookie{
		Name:     CookieAuthToken,
		Value:    "",
		Expires:  time.Unix(0, 0), // Expire in the past (January 1, 1970)
		HttpOnly: true,
		Secure:   false, // Use true if your app is served over HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	err = json.WriteResponse(w, http.StatusOK, json.Envelope{"message": "tokens were successfully deleted"}, nil)
	if err != nil {
		response.ServerError(w, r, as.Logger, err)
	}
}
