package models

import (
	"fmt"

	"hirensavani.com/db"
	"hirensavani.com/utils"
)

type Saver interface {
	save() (int64, error)
}

// User model
type User struct {
	ID       int64
	Email    string
	Password string
}

// Save method to save user to the database
func (u *User) Save() (int64, error) {

	// Prepare the statement
	stmt, err := db.DB.Prepare(QueryToSaveUser)
	if err != nil {
		return 0, fmt.Errorf("error preparing query: %w", err)
	}
	defer stmt.Close()

	// Hash the password
	hashedPassword, err := utils.HashPassword(u.Password)
	if err != nil {
		return 0, fmt.Errorf("error hashing password: %w", err)
	}

	// Execute the query and get the returned id
	var userID int64
	err = stmt.QueryRow(u.Email, hashedPassword).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("error executing query: %w", err)
	}

	// Set the user ID
	u.ID = userID
	return userID, nil
}
