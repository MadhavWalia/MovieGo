package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"moviego.madhav.net/internal/validator"
)

// Movie struct which contains information about a movie
type Movie struct {
	ID        int64     // Unique integer ID for the movie
	CreatedAt time.Time // Timestamp for when the movie is added to the database
	Title     *string   // Movie title
	Year      *int32    // Movie release year
	Runtime   *int32    // Movie runtime (in minutes)
	Genres    []string  // Slice of genres for the movie (romance, comedy, etc.)
	Version   int32     // Counter to track the number of updates to the movie
}

// Validate method which validates the movie struct
func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(*movie.Title != "", "title", "must be provided")
	v.Check(len(*movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(*movie.Year != 0, "year", "must be provided")
	v.Check(*movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(*movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(*movie.Runtime != 0, "runtime", "must be provided")
	v.Check(*movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}

// Wrapper around the sql.DB connection pool
type MovieModel struct {
	DB *sql.DB
}

// CRUD OPERATIONS for the MovieModel

// Insert a new movie record into the movies table
func (m MovieModel) Insert(movie *Movie) error {
	// Defining the SQL query for inserting a new record
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	// Creating an args slice to store the values for the placeholder parameters
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Creating a new context with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Executing the query using the DB connection pool
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Get a specific movie based on its id
func (m MovieModel) Get(id int64) (*Movie, error) {
	// Validating the id parameter
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Defining the SQL query for retrieving the movie record
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id = $1`

	// Declaring a movie struct to hold the data returned by the query
	var movie Movie

	// Creating a new context with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Executing the query using the DB connection pool
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)

	// Handling the errors
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// Returning the movie struct
	return &movie, nil
}

// Update a specific movie based on its id
func (m MovieModel) Update(movie *Movie) error {
	// Defining the SQL query for updating the movie record
	query := `
		UPDATE movies
		SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
		WHERE id = $5 AND version = $6
		RETURNING version`

	// Creating an args slice to store the values for the placeholder parameters
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	// Creating a new context with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Executing the query using the DB connection pool
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict

		default:
			return err
		}
	}

	return nil
}

// Delete a specific movie based on its id
func (m MovieModel) Delete(id int64) error {
	// Validating the id parameter
	if id < 1 {
		return ErrRecordNotFound
	}

	// Defining the SQL query for deleting the movie record
	query := `
		DELETE FROM movies
		WHERE id = $1`

	// Creating a new context with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Executing the query using the DB connection pool
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// Checking if the movie record was found
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	// Returning nil if the movie record was found
	return nil
}

// List all movies in the database
func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	// Defining the SQL query for retrieving the movie records
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	// Creating a new context with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Creating an args slice to store the values for the placeholder parameters
	args := []any{title, pq.Array(genres), filters.limit(), filters.offset()}

	// Executing the query using the DB connection pool
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	// Closing the rows object when we return from the function
	defer rows.Close()

	// Declaring a slice to hold the movie records and the total number of records
	totalRecords := 0
	movies := []*Movie{}

	// Looping through the rows in the result set
	for rows.Next() {
		// Initializing an empty movie struct
		var movie Movie

		// Scanning the values from each row into the movie struct
		err := rows.Scan(
			&totalRecords,
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		// Appending the movie struct to the slice
		movies = append(movies, &movie)
	}

	// Handling the errors encountered during the rows.Next() loop
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Declaring a metadata struct to hold the metadata for the response
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	// Returning the slice of movies
	return movies, metadata, nil
}

// CRUD OPERATIONS for the MockMovieModel

// Mock Movie Model for testing
type MockMovieModel struct{}

// CRUD OPERATIONS for the MockMovieModel

// Insert a new movie record into the movies table
func (m MockMovieModel) Insert(movie *Movie) error {
	return nil
}

// Get a specific movie based on its id
func (m MockMovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

// Update a specific movie based on its id
func (m MockMovieModel) Update(movie *Movie) error {
	return nil
}

// Delete a specific movie based on its id
func (m MockMovieModel) Delete(id int64) error {
	return nil
}

// List all movies in the database
func (m MockMovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	return nil, Metadata{}, nil
}
