package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"moviego.madhav.net/internal/validator"
)

// Defining a custom error for duplicate email
var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

// Defining an anonymous variable for an empty user struct to differentiate it from an actual erraneous user struct
var AnonymousUser = &User{}

// Defining a User struct to hold the information about a user
type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

// Checking if the user instance is anonymous
func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

// Defining a custom password struct to hold the password in both plain text and hashed format
type password struct {
	plaintext *string
	hash      []byte
}

// Setter Function for the password
func (p *password) Set(plaintext string) error {
	// Generating the hash of the password
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return err
	}

	// Setting the plaintext and hash fields
	p.plaintext = &plaintext
	p.hash = hash

	return nil
}

// Method to check if the password is correct
func (p *password) Matches(plaintext string) (bool, error) {
	// Compare the plaintext with the hash
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintext))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

// Validating thhe email address
func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

// Validating the password
func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")

	// Validate the email address
	ValidateEmail(v, user.Email)

	// If the password plaintext is not nil, validate it
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	// Check if the password hash is present
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

// Defining a UserModel struct to hold the database connection pool
type UserModel struct {
	DB *sql.DB
}

// Insert a new user record into the users table
func (m UserModel) Insert(user *User) error {
	// Defining the SQL query for inserting a new record
	query := `
	INSERT INTO users (name, email, password_hash, activated)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version`

	// Creating a new context with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Creating an args slice to hold the values for the placeholder parameters
	args := []any{user.Name, user.Email, user.Password.hash, user.Activated}

	// Executing the query and storing the result in a new row
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		// If there is a duplicate key error, return the ErrDuplicateEmail custom error
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

// Get a specific user record based on the user email address
func (m UserModel) GetByEmail(email string) (*User, error) {
	// Defining the SQL query for retrieving the user record
	query := `
	SELECT id, created_at, name, email, password_hash, activated, version
	FROM users
	WHERE email = $1`

	// Creating a new context with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Executing the query and storing the result in a new user struct
	var user User
	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		// If there is no matching record, return the ErrRecordNotFound custom error
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

// Update an existing user record in the users table
func (m UserModel) Update(user *User) error {
	// Defining the SQL query for updating the user record
	query := `
	UPDATE users
	SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version`

	// Creating a new context with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Creating an args slice to hold the values for the placeholder parameters
	args := []any{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	// Executing the query and storing the result in a new row
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		// If there is a duplicate key error, return the ErrDuplicateEmail custom error
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

// Retrieving a user record based on the token hash and scope from the tokens table
func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*User, error) {
	// Calculating the hashed version of the plaintext token
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	// Defining the SQL query for retrieving the user record based on the token hash and scope
	query := `
	SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
	FROM users
	INNER JOIN tokens
	ON users.id = tokens.user_id
	WHERE tokens.hash = $1 AND tokens.scope = $2 AND tokens.expiry > $3`

	// Creating an args slice to hold the values for the placeholder parameters
	args := []any{tokenHash[:], tokenScope, time.Now()}

	// Creating a new context with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Executing the query and storing the result in a new user struct
	var user User
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		// If there is no matching record, return the ErrRecordNotFound custom error
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// Returning the user struct
	return &user, nil
}
