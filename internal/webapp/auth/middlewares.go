package auth

import (
	"errors"
	"net/http"

	"github.com/navazjm/pixelarcade/internal/webapp/utils/database"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/response"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/validator"
)

const (
	CookieAuthToken = "auth_token"
)

func (s *Service) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		cookie, err := r.Cookie(CookieAuthToken)
		if err != nil || cookie == nil {
			// No cookie found, proceed as an anonymous user
			r = ContextSetUser(r, AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		token := cookie.Value

		v := validator.New()
		if ValidateTokenPlaintext(v, token); !v.Valid() {
			// Token is invalid, respond with an error
			response.InvalidAuthenticationToken(w, r, s.Logger)
			return
		}

		// Retrieve the user based on the token
		user, err := s.Models.GetUserFromToken(ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, database.ErrRecordNotFound):
				response.InvalidAuthenticationToken(w, r, s.Logger)
			default:
				response.ServerError(w, r, s.Logger, err)
			}
			return
		}

		// Set the user in the request context
		r = ContextSetUser(r, user)
		s.Logger.Info("set user in context")
		next.ServeHTTP(w, r)
	})
}

func (s *Service) RequireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := ContextGetUser(r)

		if user == AnonymousUser {
			response.AuthenticationRequired(w, r, s.Logger)
			return
		}

		next.ServeHTTP(w, r)
	})
}
