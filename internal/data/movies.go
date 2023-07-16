package data

import (
	"time"

	"moviego.madhav.net/internal/validator"
)

// Movie struct which contains information about a movie
type Movie struct {
	ID int64  						// Unique integer ID for the movie
	CreatedAt time.Time  	// Timestamp for when the movie is added to the database
	Title string					// Movie title
	Year int32						// Movie release year
	Runtime int32					// Movie runtime (in minutes)
	Genres []string				// Slice of genres for the movie (romance, comedy, etc.)
	Version int32					// Counter to track the number of updates to the movie
}


//Validate method which validates the movie struct
func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}