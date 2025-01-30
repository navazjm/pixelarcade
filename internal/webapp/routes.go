package webapp

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"

	"github.com/navazjm/pixelarcade/internal/webapp/utils/json"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/response"
)

func (app *Application) Routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// since url included "api", we return not found handler instead of returning react FE
		if strings.Contains(r.URL.Path, "api") {
			response.NotFound(w, r, app.Logger)
			return
		}
	})

	router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response.MethodNotAllowed(w, r, app.Logger)
	})

	router.HandlerFunc(http.MethodGet, "/api/healthcheck", app.healthcheckHandler)

	return app.recoverPanic(app.secureHeaders(app.logRequest(app.enforceCORS(app.rateLimit(router)))))
}

func (app *Application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	env := json.Envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.Config.Env,
			"version":     app.Config.Version,
		},
	}

	err := json.Write(w, http.StatusOK, env, nil)
	if err != nil {
		response.ServerError(w, r, app.Logger, err)
	}
}
