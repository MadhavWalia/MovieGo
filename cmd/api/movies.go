package main

import (
	"fmt"
	"errors"
	"net/http"

	"moviego.madhav.net/internal/data"
	"moviego.madhav.net/internal/validator"
)

// createMovieHandler for the "POST /v1/movies" endpoint
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an input struct to hold the expected data from the client (Resquest DTO)
	var input struct {
		Title *string `json:"title"`
		Year *int32 `json:"year"`
		Runtime *int32 `json:"runtime"`
		Genres []string `json:"genres"`
	}


	// Decode the request body into the input struct
	err := app.readJson(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}


	//Intermediary input for validation
	movie := &data.Movie {
		Title: input.Title,
		Year: input.Year,
		Runtime: input.Runtime,
		Genres: input.Genres,
	}
	// Validate the input
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}


	// Insert the movie into the database using the movie model
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}


	// Add a Location header to the response containing the URL of the new movie
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))


	// Return a 201 Created status code along with the movie data
	err = app.writeJson(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	
}


// showMovieHandler for the "GET /v1/movies/:id" endpoint
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the id from the URL
	id, err := app.readIDParam(r)
	if err != nil{
		app.notFoundResponse(w, r)
		return
	}

	// Retriving the movie record from the database, based on the ID
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return a 200 OK status code along with the movie data
	err = app.writeJson(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}


// updateMovieHandler for the "PATCH /v1/movies/:id" endpoint
func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the id from the URL	
	id, err := app.readIDParam(r)
	if err != nil{
		app.notFoundResponse(w, r)
		return
	}

	// Retriving the movie record from the database, based on the ID
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}


	// Declare an input struct to hold the expected data from the client (Resquest DTO)
	var input struct {
		Title *string `json:"title"`
		Year *int32 `json:"year"`
		Runtime *int32 `json:"runtime"`
		Genres []string `json:"genres"`
	}

	// Decode the request body into the input struct
	err = app.readJson(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}


	// Copy the new data across to the movie record if it is provided
	if input.Title != nil {
		movie.Title = input.Title
	}
	if input.Year != nil {
		movie.Year = input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	// Validate the input
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}


	// Update the movie record in the database
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
			case errors.Is(err, data.ErrEditConflict):
				app.editConflictResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
		}
		return
	}


	// Return a 200 OK status code along with the movie data
	err = app.writeJson(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}


// deleteMovieHandler for the "DELETE /v1/movies/:id" endpoint
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the id from the URL
	id, err := app.readIDParam(r)
	if err != nil{
		app.notFoundResponse(w, r)
		return
	}

	// Delete the movie from the database, based on the ID
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return a 200 OK status code along with a success message
	err = app.writeJson(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

