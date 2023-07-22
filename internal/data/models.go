package data

import (
	"database/sql"
	"errors"
)


var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict = errors.New("edit conflict")
)


//Parent Model struct for all the models
type Models struct {
	Movies interface {
		Insert(movie *Movie) error
		Get(id int64) (*Movie, error)
		Update(movie *Movie) error
		Delete(id int64) error
		GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error)
	}
}


// Factory method to create a new Models struct
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}

//Factory method to create a new Mock Movie struct for testing
func NewMockModels() Models {
	return Models{
		Movies: MockMovieModel{},
	}
}