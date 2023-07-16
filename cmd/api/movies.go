package main

import (
	"fmt"
	"net/http"

	"moviego.madhav.net/internal/data"
	"moviego.madhav.net/internal/validator"
)

// createMovieHandler for the "POST /v1/movies" endpoint
// Adding a placeholder response for now
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
		Year int32 `json:"year"`
		Runtime int32 `json:"runtime"`
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


	// Return a response 
	fmt.Fprintf(w, "%+v\n", input)
}


// showMovieHandler for the "GET /v1/movies/:id" endpoint
// Adding a placeholder response for now
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil{
		app.notFoundResponse(w, r)
		return
	}

	
	fmt.Fprintf(w, "show the details of movie %d\n", id)
}


