package models

import (
	"fmt"

	"hirensavani.com/db"
)

type Groups struct {
	ID          int64
	Name        string
	Description string
	Members     []User
}

func (g *Groups) Save() error {

	query := "INSERT INTO groups (name,description) VALUES ($1, $2) RETURNING id"

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(g.Name, g.Description)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	return nil

}

func (g *Groups) AddMember(userId int64) error {

	query := "INSERT INTO group_member (group_id,user_id) VALUES ($1, $2)"

	stmt, err := db.DB.Prepare(query)

	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(g.ID, userId)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	return nil
}
