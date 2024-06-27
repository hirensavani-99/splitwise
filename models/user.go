package models

import (
	"fmt"

	"hirensavani.com/db"
	"hirensavani.com/utils"
)

// User model
type User struct {
	ID       int64
	Email    string
	Password string
}

// Save method to save user to the database
func (u *User) Save() error {
	// PostgreSQL uses $1, $2 for placeholders
	query := "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id"

	// Debugging output for SQL query and user data
	fmt.Printf("Executing query: %s with email: %s and hashed password\n", query, u.Email)

	// Prepare the statement
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}
	defer stmt.Close()

	// Hash the password
	hashedPassword, err := utils.HashPassword(u.Password)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	// Execute the query and get the returned id
	var userID int64
	err = stmt.QueryRow(u.Email, hashedPassword).Scan(&userID)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	// Set the user ID
	u.ID = userID
	return nil
}
