package main

import (
	"errors"
	"net/http"
	"time"

	"moviego.madhav.net/internal/data"
	"moviego.madhav.net/internal/validator"
)


// createAuthenticationTokenHandler for the "POST /v1/tokens/authentication" endpoint
func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an input struct to hold the expected data from the client (Resquest DTO)
	var input struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}


	// Decode the request body into the input struct
	err := app.readJson(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}


	// Validate the input
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}


	// Check whether a user exists with the provided email address, if not, then send the 401 Unauthorized response
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}


	// Check if the provided password is correct, if not, then send the 401 Unauthorized response
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Checking if the match is successful
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}


	// Create a new instance of the token model, containing the 24hr expiry time and authentication scope
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}


	// Add the token to the response
	err = app.writeJson(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}