package main

import (
	"fmt"
	"net/http"
)



func (app *application) logError(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url": r.URL.String(),
	})
}



func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	err := app.writeJson(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}


func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}


func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}



func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "The server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}



func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "The requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}



func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}


func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "Unable to update the record due to an edit conflict, please try again"
	app.errorResponse(w, r, http.StatusConflict, message)
}


func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "Rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}


func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "Invalid authentication credentials"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}


func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	// Setting the WWW-Authenticate header to tell the client that it needs to send credentials
	w.Header().Set("WWW-Authenticate", "Bearer")

	message := "Invalid or missing authentication token"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}


func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "You must be authenticated to access this resource"
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}


func (app *application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "Your user account must be activated to access this resource"
	app.errorResponse(w, r, http.StatusForbidden, message)
}