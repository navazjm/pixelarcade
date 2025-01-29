package webapp

import (
	"net/http"

	"github.com/navazjm/pixelarcade/internal/webapp/utils/json"
	"github.com/navazjm/pixelarcade/internal/webapp/utils/response"
)

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
