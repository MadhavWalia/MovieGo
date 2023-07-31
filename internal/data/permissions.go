package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// Defining a permissions type for storing the permissions for a user
type Permissions []string


// Method for checking if the permissions contains a specific permission
func (p Permissions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}

	return false
}


// Defining a Permissions Model to hold the connection pool
type PermissionModel struct {
	DB *sql.DB
}


// Method for retrieving all permissions for a specific user
func (m PermissionModel) GetAllForUser(userID int64) (Permissions, error) {
	// Defining the SQL query for retrieving the permissions for a specific user
	query := `
		SELECT permissions.code
		FROM permissions
		INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
		INNER JOIN users ON users_permissions.user_id = users.id
		WHERE users.id = $1`

	
	// Defining a context with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()


	// Executing the query and returning the result set or an error
	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()


	// Declaring a Permissions slice to hold the permissions
	var permissions Permissions


	// Looping through the result set and appending the permissions to the slice
	for rows.Next() {
		var permission string

		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}

		permissions = append(permissions, permission)
	}
	// Checking for errors from iterating over the result set
	if err = rows.Err(); err != nil {
		return nil, err
	}


	// Returning the permissions
	return permissions, nil
}


// Method for granting permissions to a user
func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	// Defining the SQL query for inserting the permissions for a specific user
	query := `
		INSERT INTO users_permissions
		SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)`


	// Defining a context with a 3 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()


	// Executing the query and returning the result set or an error
	_, err := m.DB.ExecContext(ctx, query, userID, pq.Array(codes))
	return err
}