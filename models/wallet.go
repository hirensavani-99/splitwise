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

	stmt, err := db.Prepare(QueryToSaveWallet)
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

	row := db.QueryRow(QueryToGetWalletDataByUserId, userID)
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

	err = db.QueryRow(QueryToGetWalletBalanceByUserId, userid).Scan(&wallet.Balance)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to get current balance: %w", err)
	}

	wallet.Balance += adjustment
	wallet.UpdatedAt = time.Now()
	
	//Update wallet
	_, err = db.Exec(QueryToUpdateWalletBalance, userid, wallet.Balance, wallet.UpdatedAt)

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
