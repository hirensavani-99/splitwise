package models

import (
	"database/sql"
	"fmt"
)

type Balances struct {
	FromUserID int64
	ToUserID   int64
	GroupId    int64
	Amount     float64
}

func (bal *Balances) Get(db *sql.DB, userID int64) (map[int64]float64, error) {
	balances := make(map[int64]float64)
	query := `	
		SELECT from_user_id, to_user_id, amount
		FROM Balances
		WHERE from_user_id = $1 OR to_user_id = $1
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query balances: %w", err)
	}
	defer rows.Close()

	for rows.Next() {

		if err := rows.Scan(&bal.FromUserID, &bal.ToUserID, &bal.Amount); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if bal.FromUserID != userID {
			balances[bal.FromUserID] = -bal  .Amount
		} else {
			balances[bal.ToUserID] = bal.Amount
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return balances, nil
}
