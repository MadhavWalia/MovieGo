package main

import (
	"errors"
	"net/http"
	"time"

	"moviego.madhav.net/internal/data"
	"moviego.madhav.net/internal/validator"
)

// registerUserHandler for the "POST /v1/users" endpoint
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an input struct to hold the expected data from the client (Resquest DTO)
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode the request body into the input struct
	err := app.readJson(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Intermediary input for validation
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}
	// Use the Set() method to generate and store the hashed and plaintext passwords
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Validate the input
	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the user into the database using the user model
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Adding the movies:read permission to the user as default
	err = app.models.Permissions.AddForUser(user.ID, "movies:read")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Create a new activation token for the user
	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a welcome email to the user as a background task
	app.background(func() {
		// Define the data for the welcome email
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}

		// Sending the welcome email
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	// Return a 201 Created status code along with the user data
	err = app.writeJson(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// activateUserHandler for the "POST /v1/users/activated" endpoint
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	// Declaring an input struct to hold the expected data from the client (Request DTO)
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	// Decoding the request body into the input struct
	err := app.readJson(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validating the input
	v := validator.New()
	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Retrieving the details of the user associated with the token hash and scope
	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Updating the user's activation status to true
	user.Activated = true

	// Updating the user record in the database
	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		// If there is a edit conflict, then we return a 409 Conflict status code
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// If the user is successfully activated, then we delete all the activation tokens for the user
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Return a 200 OK status code along with the user data
	err = app.writeJson(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
