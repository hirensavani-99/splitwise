package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Wallet struct {
	UserID    int64
	Balance   float64
	Currency  string
	Balances  map[int64]float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Balances struct {
	FromUserID int64
	ToUserID   int64
	GroupId    int64
	Amount     float64
}

func (wallet *Wallet) Save(db *sql.DB) error {

	query := `
		INSERT INTO wallets (
			user_id, balance, currency
		) VALUES (
			$1, $2, $3
		)`
	fmt.Println(wallet)
	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(wallet.UserID, wallet.Balance, wallet.Currency)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	return nil

}

func (wallet *Wallet) Get(db *sql.DB, userID int64) error {

	query := `
		SELECT user_id, balance, currency, createdAt, updatedAt
		FROM wallets
		WHERE user_id = $1
	`

	row := db.QueryRow(query, userID)
	err := row.Scan(&wallet.UserID, &wallet.Balance, &wallet.Currency, &wallet.CreatedAt, &wallet.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("wallet not found for user_id: %d", userID)
		}
		return fmt.Errorf("failed to query wallet: %w", err)
	}

	// Initialize the Balances map
	wallet.Balances, err = getBalancesForUser(db, userID)
	if err != nil {
		return fmt.Errorf("failed to get balances for user: %w", err)
	}

	return nil
}

func getBalancesForUser(db *sql.DB, userID int64) (map[int64]float64, error) {
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
		var fromUser int64
		var toUser int64
		var amount float64
		if err := rows.Scan(&fromUser, &toUser, &amount); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if fromUser != userID {
			balances[fromUser] = -amount
		} else {
			balances[toUser] = amount
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return balances, nil
}

func updateWallet(db *sql.DB, userid int64, adjustment float64) error {
	var currentBalance float64
	querySelect := `SELECT BALANCE FROM Wallets WHERE USER_ID=$1`
	err := db.QueryRow(querySelect, userid).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("failed to get current balance: %w", err)
	}

	totalWalletBalance := currentBalance + adjustment
	//Update wallete
	updatedWallet := `UPDATE Wallets SET BALANCE=$2 WHERE USER_ID=$1`

	_, err = db.Exec(updatedWallet, userid, totalWalletBalance)

	if err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}
	return nil
}

func NewWallet(userId int64, balance float64, currency string) Wallet {
	now := time.Now()
	return Wallet{
		UserID:    userId,
		Balance:   balance,
		Currency:  currency,
		CreatedAt: now,
		UpdatedAt: now,
	}

}
