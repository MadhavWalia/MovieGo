package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

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


	// Checking if the "X-Version" header is provided and if it matches the current version of the movie record
	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(movie.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
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


// listMoviesHandler for the "GET /v1/movies" endpoint
func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an input struct to hold the expected data from the client (Resquest DTO)
	var input struct {
		Title string
		Genres []string
		data.Filters
	}


	// Validating the query string parameters
	v := validator.New()
	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}


	// Retriving the movies from the database, based on the filters
	movies, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}


	// Return a 200 OK status code along with the movie data
	err = app.writeJson(w, http.StatusOK, envelope{"movies": movies}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

