package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/julienschmidt/httprouter"
	"moviego.madhav.net/internal/validator"
)

// function to load the environment variables
func loadEnv(key string) (string, error) {
	//Loading the environment variables from the .env file
	err := godotenv.Load(".env")
	if err != nil {
		return "", err
	}

	//Getting the value of the key from the environment variables
	return os.Getenv(key), nil
}


type envelope map[string]any

// method to send a JSON response, with the appropriate status code
func (app *application) writeJson(w http.ResponseWriter, status int, data envelope, headers http.Header) error {

	// Encode the data to JSON, returning the error if there was one
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Append a newline for ease of view
	js = append(js, '\n')

	//Adding the headers to the response
	for key, value := range headers {
		w.Header()[key] = value
	}

	//Setting the content type header to application/json
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}


// method to read the json request body into a destination struct
func (app *application) readJson(w http.ResponseWriter, r *http.Request, dst any) error {

	// Limit the size of the request body to 1MB
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	//Initialize the json.Decoder, and disallow unknown fields
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the request body into the destination struct
	err := dec.Decode(dst)
	if err != nil {
		// Triaging the error
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
			//For a syntax error, log the details and return a bad request error
			case errors.As(err, &syntaxError):
				return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)


			case errors.Is(err, io.ErrUnexpectedEOF):
				return errors.New("body contains badly-formed JSON")
			

			//For an unmarshal type error, log the details and return a bad request error
			case errors.As(err, &unmarshalTypeError):
				if unmarshalTypeError.Field != "" {
					return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
				}
				return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)


			case errors.Is(err, io.EOF):
				return errors.New("body must not be empty")

			
			//For an unknown field error, log the details and return a bad request error
			case strings.HasPrefix(err.Error(), "json: unknown field"):
				fieldname := strings.TrimPrefix(err.Error(), "json: unknown field ")
				return fmt.Errorf("body contains unknown key %s", fieldname)

			
			//For a too large error, log the details and return a bad request error
			case err.Error() == "http: request body too large":
				return fmt.Errorf("body must not be larger than %d bytes", maxBytes)


			//For an invalid unmarshal error, log the details and return a bad request error
			case errors.As(err, &invalidUnmarshalError):
				panic(err)


			//For any other type of error, return the error message as-is
			default:
				return err
		}
	}

	// Checking if body contains only one JSON object
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must contain a JSON object")
	}
	return nil
}


// method to read the id parameter from the URL
func (app *application) readIDParam (r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}


// method to read CSV data from the query string
func (app *application) readCSV(ps url.Values, key string, defaultValue []string) []string {
	// Extract the value from the query string
	csv := ps.Get(key)

	// If no key exists, or the value is empty, return the default value
	if csv == "" {
		return defaultValue
	}

	// Else, split the value into a []string slice
	return strings.Split(csv, ",")
}


// method to read an integer value from the query string and convert it to an int
func (app *application) readInt(ps url.Values, key string, defaultValue int, v *validator.Validator) int {
	// Extract the value from the query string
	s := ps.Get(key)

	// If no key exists, or the value is empty, return the default value
	if s == "" {
		return 0
	}

	// Else, try to convert the value to an int
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return 0
	}

	// Return the integer value
	return i
}


// method to read a string value from the query string
func (app *application) readString(ps url.Values, key string, defaultValue string) string {
	// Extract the value from the query string
	s := ps.Get(key)

	// If no key exists, or the value is empty, return the default value
	if s == "" {
		return defaultValue
	}

	// Return the string value
	return s
}