package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)


// routes method which returns a httprouter.Router instance containing the application routes
func(app *application) routes() http.Handler {
	// Initialize the new httprouter router instance
	router := httprouter.New()


	// Set the NotFound and MethodNotAllowed error handlers for the router instance
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)


	// Register the relevant methods, URL patterns and handler functions for our endpoints

	// The Status Healthcheck endpoint
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)


	// CRUD endpoints for the movies resource
	router.HandlerFunc(
		http.MethodPost, 
		"/v1/movies", 
		app.requirePermission("movies:write", app.createMovieHandler),
	)

	router.HandlerFunc(
		http.MethodGet, 
		"/v1/movies/:id", 
		app.requirePermission("movies:read", app.showMovieHandler),
	)

	router.HandlerFunc(
		http.MethodPatch, 
		"/v1/movies/:id", 
		app.requirePermission("movies:write", app.updateMovieHandler),
	)

	router.HandlerFunc(
		http.MethodDelete,
		"/v1/movies/:id",
		app.requirePermission("movies:write", app.deleteMovieHandler),
	)

	router.HandlerFunc(
		http.MethodGet, 
		"/v1/movies", 
		app.requirePermission("movies:read", app.listMoviesHandler),
	)


	// CRUD endpoints for the users resource
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)


	// Authentication and Authorization endpoints
	router.HandlerFunc(
		http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler,
	)


	// Return the httprouter instance
	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))
}