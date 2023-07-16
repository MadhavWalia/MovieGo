package validator

import (
	"regexp"
)


var (
	// Define a regex which can be used to match against valid email addresses
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

//Map of validation errors
type Validator struct {
	Errors map[string]string
}

// Factory method to create a new Validator type
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Checking if there is an error for the given key
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// Add an error message for a given field to the map of errors
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check adds an error message to the map of errors if the condition is false
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}


// In method which checks if a string value is in a list of strings
func In(value string, list ...string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}


// Matches method which checks if a string value matches a regex pattern
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}


// Unique method which checks if all string values in a slice are unique
func Unique(values []string) bool {
	uniqueValues := make(map[string]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}