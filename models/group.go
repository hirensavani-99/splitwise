package models

import (
	"fmt"

	"hirensavani.com/db"
)

type Groups struct {
	ID           int64
	Name         string
	Description  string
	UserIds      []int64 `json:"userIds" binding:"required"`
	SimplifyDebt bool
}

func (g *Groups) Save() (int64, error) {

	var groupId int64

	stmt, err := db.DB.Prepare(QueryToSaveGroup)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(g.Name, g.Description, g.SimplifyDebt).Scan(&groupId)
	if err != nil {
		return 0, err
	}

	return groupId, nil

}

func (g *Groups) AddMember(userId int64) error {

	stmt, err := db.DB.Prepare(QueryToAddGroupMember)

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
