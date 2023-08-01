package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"moviego.madhav.net/internal/validator"
)

// Define the different scopes of the token (what it can access)
const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

// Defining the token struct to hold the details of the token
type Token struct {
	Plaintext string    `json:"token_plaintext"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

// Function to generate a new token for a user with a specific scope
func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	// Create a token instance to hold the provided details of the token
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// Create a random byte slice to hold the plaintext token
	randomBytes := make([]byte, 16)

	// Reading random bytes from the OS's CSPRNG into the randomBytes slice
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// Encoding the random bytes using base32 encoding to add to the plaintext token
	// This will be sent to the user email
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// Hashing the plaintext token using SHA256 to store in the hash field of the token
	hash := sha256.Sum256([]byte(token.Plaintext))
	// Note: the hash is stored as a byte array, so we need to convert it to a byte slice
	token.Hash = hash[:]

	// Return the token instance
	return token, nil
}

// Function to validate the plaintext token provided by the user
func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

// Defining the TokenModel struct to hold the database pool
type TokenModel struct {
	DB *sql.DB
}

// Method for creating a new token and inserting it into the database
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	// Generating a new token for the user
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	// Insrting the token into the database
	err = m.Insert(token)
	return token, err
}

// Method for inserting a token into the database
func (m TokenModel) Insert(token *Token) error {
	// Defining the SQL query for inserting a new token
	query := `
	INSERT INTO tokens (hash, user_id, expiry, scope)
	VALUES ($1, $2, $3, $4)`

	// Defining the arguments for the SQL query
	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	// Creating a context with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Executing the query using the DB connection pool
	_, err := m.DB.ExecContext(ctx, query, args...)
	return err
}

// Moethod for deleting all tokens for a specific user and scope
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	// Defining the SQL query for deleting all tokens for a specific user and scope
	query := `
	DELETE FROM tokens
	WHERE scope = $1 AND user_id = $2`

	// Creating a context with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Executing the query using the DB connection pool
	_, err := m.DB.ExecContext(ctx, query, scope, userID)
	return err
}
