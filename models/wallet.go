package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type Wallet struct {
	UserID    int64
	Balance   float64
	Currency  string
	Balances  map[int64]map[int64]float64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type SettlementType struct {
	PayeeID int64   `json:"payee_id"`
	PayerID int64   `json:"payer_id"`
	Amount  float64 `json:"amount"`
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

func (settlement *SettlementType) SettleUpWallet(db *sql.DB) error {
	fmt.Println("123")

	// Get balances where user 1 and 2 both exist I will check How much I can settled up rest will be moved for next Group
	rows, err := db.Query(QueryToGetBalancesWhereBothUsersExists, &settlement.PayeeID, &settlement.PayerID)
	if err != nil {
		return WrapError(err, ErrExecutingQuery)
	}
	defer rows.Close() // Ensure rows are closed when the function exits

	// Loop through each row in the result set
	for rows.Next() {

		if settlement.Amount >= 0 {

			var bal Balances // Assuming you have a Balance struct defined
			if err := rows.Scan(&bal.FromUserID, &bal.ToUserID, &bal.GroupId, &bal.Amount); err != nil {
				return WrapError(err, ErrScaningRow)
			}

			settlement.Amount = settlement.Amount - bal.Amount
			bal.FromUserID = settlement.PayerID
			bal.ToUserID = settlement.PayeeID

			fmt.Println(bal)

			balance := Balances{}
			//get Balance For group id
			res, err := balance.getBalanacesForGroup(db, bal.GroupId)

			if err != nil {
				return WrapError(err, ErrGettingGroupBalances)
			}

			// Update wallet
			wallet := &Wallet{}

			// Update wallet for Payer
			err = wallet.Update(db, settlement.PayerID, bal.Amount)

			if err != nil {
				return WrapError(err, ErrUpdatingWallet)
			}

			// Update wallet for payee
			err = wallet.Update(db, settlement.PayeeID, -bal.Amount)

			if err != nil {
				return WrapError(err, ErrUpdatingWallet)
			}

			// Update Balances

			res = append(res, bal)

			fmt.Println(res)

			newbalances := UniqueBalances(res)

			fmt.Println(newbalances)

			// Calculate net balances
			netBalances := calculateNetBalances(newbalances)

			fmt.Println(netBalances)

			// Separate debtors and creditors
			creditors, debtors := separateDebtorsAndCreditors(netBalances)

			fmt.Println(creditors, debtors)

			balances := minimizeTransactions(debtors, creditors, netBalances, bal.GroupId)

			fmt.Println(balances)

			err = DeleteUnnecessaryBalances(balances, bal.GroupId)

			if err != nil {
				log.Fatalf("Error deleting balances: %v", err)
			}
		}

	}

	// Check for errors encountered during iteration
	if err := rows.Err(); err != nil {
		return WrapError(err, "Error Row iteration")
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
