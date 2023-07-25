package main

import (
	"errors"
	"net/http"

	"moviego.madhav.net/internal/data"
	"moviego.madhav.net/internal/validator"
)


// registerUserHandler for the "POST /v1/users" endpoint
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	// Declare an input struct to hold the expected data from the client (Resquest DTO)
	var input struct {
		Name string `json:"name"`
		Email string `json:"email"`
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
		Name: input.Name,
		Email: input.Email,
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


	// Return a 201 Created status code along with the user data
	err = app.writeJson(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}