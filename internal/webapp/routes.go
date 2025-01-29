package webapp

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"

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

	return app.recoverPanic(app.secureHeaders(app.logRequest(app.enableCORS(app.validateOrigin(app.rateLimit(router))))))
}
