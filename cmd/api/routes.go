package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)


// routes method which returns a httprouter.Router instance containing the application routes
func(app *application) routes() *httprouter.Router {
	// Initialize the new httprouter router instance
	router := httprouter.New()

	// Set the NotFound and MethodNotAllowed error handlers for the router instance
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Register the relevant methods, URL patterns and handler functions for our endpoints
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	router.HandlerFunc(http.MethodPut, "/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)

	// Return the httprouter instance
	return router
}