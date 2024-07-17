package models

import (
	"database/sql"
	"time"
)

type Comment struct {
	ID        int64     `json:"id"`
	ExpenseID int64     `json:"expense_id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

func (comment Comment) Get() {

}

// Saving comments
func (comment Comment) Save(db *sql.DB) error {

	// Checking expense exist
	isExpenseExists := IsExpense(db, comment.ExpenseID)

	if !isExpenseExists {
		return WrapErrMessage(ErrExpNotExists)
	}

	//Saving Comments for expense
	_, err := db.Exec(QueryToSaveComments, comment.ExpenseID, comment.UserID, comment.Content)

	if err != nil {
		return WrapError(err, ErrExecutingQuery)
	}

	return nil
}

func (comment Comment) update() {

}

func (comment Comment) Delete() {

}
