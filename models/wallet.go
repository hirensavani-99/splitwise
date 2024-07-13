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

	balances := &Balances{}
	// Initialize the Balances map
	wallet.Balances, err = balances.Get(db, userID)
	if err != nil {
		return fmt.Errorf("failed to get balances for user: %w", err)
	}

	return nil
}

func (wallet *Wallet) Update(db *sql.DB, userid int64, adjustment float64) error {
	tx, err := db.Begin()

	if err != nil {
		return fmt.Errorf("Failed to begin transcations %w", err)
	}

	querySelect := `SELECT BALANCE FROM Wallets WHERE USER_ID=$1`
	err = db.QueryRow(querySelect, userid).Scan(&wallet.Balance)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get current balance: %w", err)
	}

	wallet.Balance += adjustment
	wallet.UpdatedAt = time.Now()
	//Update wallete
	updatedWallet := `UPDATE Wallets SET BALANCE=$2 , updatedAt=$3 WHERE USER_ID=$1`

	_, err = db.Exec(updatedWallet, userid, wallet.Balance, wallet.UpdatedAt)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transcations.")
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
