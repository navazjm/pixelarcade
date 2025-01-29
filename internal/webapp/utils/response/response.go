package response

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/navazjm/pixelarcade/internal/webapp/utils/json"
)

func LogError(r *http.Request, logger *slog.Logger, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	logger.Error(err.Error(), "method", method, "uri", uri)
}

func Error(w http.ResponseWriter, r *http.Request, logger *slog.Logger, status int, message any) {
	env := json.Envelope{"error": message}

	err := json.Write(w, status, env, nil)
	if err != nil {
		LogError(r, logger, err)
		w.WriteHeader(500)
	}
}

func ServerError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	LogError(r, logger, err)

	message := "the server encountered a problem and could not process your request"
	Error(w, r, logger, http.StatusInternalServerError, message)
}

func NotFound(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := "the requested resource could not be found"
	Error(w, r, logger, http.StatusNotFound, message)
}

func MethodNotAllowed(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	Error(w, r, logger, http.StatusMethodNotAllowed, message)
}

func BadRequest(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	Error(w, r, logger, http.StatusBadRequest, err.Error())
}

func FailedValidation(w http.ResponseWriter, r *http.Request, logger *slog.Logger, errors map[string]string) {
	Error(w, r, logger, http.StatusUnprocessableEntity, errors)
}

func EditConflict(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := "unable to update the record due to an edit conflict, please try again"
	Error(w, r, logger, http.StatusConflict, message)
}

func RateLimitExceeded(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := "rate limit exceeded"
	Error(w, r, logger, http.StatusTooManyRequests, message)
}

func InvalidCredentials(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := "invalid authentication credentials"
	Error(w, r, logger, http.StatusUnauthorized, message)
}

func InvalidAuthenticationToken(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "invalid or missing authentication token"
	Error(w, r, logger, http.StatusUnauthorized, message)
}

func AuthenticationRequired(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := "you must be authenticated to access this resource"
	Error(w, r, logger, http.StatusUnauthorized, message)
}

func PermissionDenied(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	message := "you do not have permission to access this resource"
	Error(w, r, logger, http.StatusForbidden, message)
}

func OriginNotAllowed(w http.ResponseWriter, r *http.Request, logger *slog.Logger, origin string) {
	message := fmt.Sprintf("request origin '%s' is not allowed", origin)
	Error(w, r, logger, http.StatusForbidden, message)
}
