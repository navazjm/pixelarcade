package param

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// Retrieve the "id" URL parameter from the current request context, then convert it to
// an integer and return it. If the operation isn't successful, return 0 and an error.
func ReadID(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

// Injects an ID into the request's context for testing.
func InjectID(r *http.Request, id int64) *http.Request {
	params := httprouter.Params{httprouter.Param{Key: "id", Value: strconv.FormatInt(id, 10)}}
	ctx := context.WithValue(r.Context(), httprouter.ParamsKey, params)
	return r.WithContext(ctx)
}
