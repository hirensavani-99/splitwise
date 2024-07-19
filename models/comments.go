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
	AddedAt   time.Time `json:"created_at"`
}

func (comment Comment) GetAddedAt() time.Time {
	return comment.AddedAt
}

func (comment *Comment) Get(db *sql.DB) ([]Comment, error) {
	var comments []Comment
	rows, err := db.Query(QueryToGetCommentsByExpenseId, comment.ExpenseID)
	if err != nil {
		return nil, WrapError(err, ErrExecutingQuery)
	}

	for rows.Next() {
		err := rows.Scan(&comment.ID, &comment.ExpenseID, &comment.UserID, &comment.Content, &comment.AddedAt)
		if err != nil {
			return nil, WrapError(err, ErrScaningRow)
		}
		comments = append(comments, *comment)
	}
	SortByTime(comments)

	return comments, nil
}

// Saving comments
func (comment *Comment) Save(db *sql.DB) error {

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
