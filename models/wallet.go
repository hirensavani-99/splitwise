package models

import (
	"fmt"
	"time"

	"hirensavani.com/db"
)

type Wallet struct {
	ID        int64
	UserID    int64
	Balance   float64
	Currency  string
	Balances  map[int64]float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (wallet *Wallet) Save() error {

	query := `
		INSERT INTO wallets (
			user_id, balance, currency
		) VALUES (
			$1, $2, $3
		) RETURNING id`

	stmt, err := db.DB.Prepare(query)
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
