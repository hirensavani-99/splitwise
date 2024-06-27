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

	err = stmt.QueryRow(g.Name, g.Description).Scan(&g.ID)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	return nil

}
