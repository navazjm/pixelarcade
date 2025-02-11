package auth

import (
	"context"
	"net/http"
)

type contextKey string

const (
	CtxKeyUser = contextKey("user")
)

func ContextSetUser(r *http.Request, user *User) *http.Request {
	ctx := context.WithValue(r.Context(), CtxKeyUser, user)
	return r.WithContext(ctx)
}

func ContextGetUser(r *http.Request) *User {
	user, ok := r.Context().Value(CtxKeyUser).(*User)
	if !ok {
		panic("missing user in request context")
	}

	return user
}
