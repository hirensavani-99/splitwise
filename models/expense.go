package models

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
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
	Currency        string
	Category        string
	IsRecurring     bool
	RecurringPeriod string
	Notes           string
	Tags            []string
	AddTo           map[string]string
	SplitType       string
}

type Balances struct {
	FromUserID int64
	ToUserID   int64
	Amount     float64
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

func CalculateBalance(expense *Expense) []Balances {
	totalAmount := expense.Amount
	numUsers := len(expense.AddTo)

	var balances []Balances

	for userId, _ := range expense.AddTo {
		fmt.Println(userId)
		userID, _ := strconv.ParseInt(userId, 10, 64)
		fmt.Println(expense.SplitType)
		switch expense.SplitType {

		case "equal":
			amountPerUser := totalAmount / float64(numUsers)

			fmt.Println(amountPerUser)
			balances = append(balances, Balances{
				FromUserID: expense.AddedBy,
				ToUserID:   userID,
				Amount:     amountPerUser,
			})

			fmt.Println(balances)

		}
	}

	return balances
}

func insertDebt(db *sql.DB, bal Balances) error {
	var amount float64
	var from_user_id int64
	var to_user_id int64
	queryToGetExistingRecord := `SELECT from_user_id, to_user_id, amount FROM BALANCES WHERE (from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1);`

	err := db.QueryRow(queryToGetExistingRecord, bal.FromUserID, bal.ToUserID).Scan(&from_user_id, &to_user_id, &amount)
	if err != nil {
		query := `
        INSERT INTO BALANCES (from_user_id, to_user_id, amount, created_at)
        VALUES ($1, $2, $3, $4)
    `

		_, err = db.Exec(query, bal.FromUserID, bal.ToUserID, bal.Amount, time.Now())
		if err != nil {
			return err
		}
		return nil
	}

	var updateQuery string
	//If from userId is same -> add
	if bal.FromUserID == from_user_id && bal.ToUserID == to_user_id {
		amount += bal.Amount
		updateQuery = `UPDATE BALANCES SET amount=$3 WHERE (from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1);`
	} else if bal.FromUserID == to_user_id && bal.ToUserID == from_user_id {
		amount -= bal.Amount

		if amount < 0 {
			amount = -amount
			updateQuery = `UPDATE BALANACES SET amount=$3 , from_user_id=$2 , to_user_id=$1 WHERE (from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1);`
		} else {
			updateQuery = `UPDATE BALANCES SET amount=$3 WHERE (from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1);`
		}
	}

	_, err = db.Exec(updateQuery, bal.FromUserID, bal.ToUserID, amount)

	if err != nil {
		return fmt.Errorf("failed to update debt: %w", err)
	}
	//if fromuserid is different -> sub

	return nil
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

	debts := CalculateBalance(ex)

	for _, debt := range debts {
		err := insertDebt(db.DB, debt)
		if err != nil {
			log.Fatalf("Error inserting debt: %v", err)
		}
	}

	_, err = stmt.Exec(ex.Description, ex.Amount, ex.Currency, ex.Category, ex.AddedAt, ex.IsRecurring, ex.RecurringPeriod, ex.Notes, ex.Groupid, ex.AddedBy)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	return nil

}
