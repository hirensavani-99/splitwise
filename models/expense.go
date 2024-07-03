package models

import (
	"database/sql"
	"fmt"
	"time"

	"hirensavani.com/db"
)

type Expense struct {
	ID              int64
	Groupid         int64
	AddedBy         int64
	Description     string
	AddedAt         time.Time
	Amount          float64
	currency        string
	category        string
	IsRecurring     bool
	RecurringPeriod string
	Notes           string
	Tags            []string
	AddTo           map[int64]string
	spiltType       string
}

func userInGroup(db *sql.DB, userId int64, groupId int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM group_member WHERE user_id = $1 AND group_id = $2)`
	err := db.QueryRow(query, userId, groupId).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (ex *Expense) Save() error {

	isMember, err := userInGroup(db.DB, ex.AddedBy, ex.Groupid)

	if err != nil {
		return err
	}

	if !isMember {
		return fmt.Errorf("user %d is not member of %d", ex.AddedBy, ex.Groupid)
	}

	query := `
		INSERT INTO expense (
			description, amount, currency, category, added_at, 
			is_recurring, recurring_period, notes, group_id, added_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(ex.Description, ex.Amount, ex.currency, ex.category, ex.AddedAt, ex.IsRecurring, ex.RecurringPeriod, ex.Notes, ex.Groupid, ex.AddedBy)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	return nil

}
