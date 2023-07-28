package main

import (
	"context"
	"net/http"

	"moviego.madhav.net/internal/data"
)


// Defining a custom contextKey type to hold the key for the context
type contextKey string


// Defining a custom contextKey for the user key, which will be used to store the user in the context
const userContextKey = contextKey("user")


// Defining a contextSetUser method to store the user in the request context
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}


// Defining a contextGetUser method to retrieve the user from the request context
// WARNING: This method will panic if the user is not in the context
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}

	return user
}