package models

import (
	"fmt"
	"log"
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

func (ex *Expense) Save() error {

	isMember, err := userInGroup(db.DB, ex.AddedBy, ex.Groupid)

	if !isMember || err != nil {
		return fmt.Errorf("user %d is not member of %d or system is unable to check your membership.", ex.AddedBy, ex.Groupid)
	}

	stmt, err := db.DB.Prepare(QueryToPostExpense)
	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}
	defer stmt.Close()

	debts, adjustment := CalculateBalance(ex)

	var simplifyDebt bool

	err = db.DB.QueryRow(QueryToGetGroupType, ex.Groupid).Scan(&simplifyDebt)

	if err != nil {
		return fmt.Errorf("error getting group type: %w", err)
	}

	if simplifyDebt {
		balance := Balances{}
		res, err := balance.getBalanacesForGroup(db.DB, ex.Groupid)

		res = append(res, debts...)

		if err != nil {
			return fmt.Errorf("error geathering balances : %w", err)
		}
		netBalances := calculateNetBalances(res)

		creditors, debtors := separateDebtorsAndCreditors(netBalances)

		balances := minimizeTransactions(debtors, creditors, netBalances, ex.Groupid)

		err = DeleteUnnecessaryBalances(balances, ex.Groupid)
		if err != nil {
			log.Fatalf("Error deleting balances: %v", err)
		}
	} else {

		for _, debt := range debts {

			err := debt.Save(db.DB, false)
			if err != nil {
				log.Fatalf("Error inserting debt: %v", err)
			}
		}
	}

	//updating wallete for debtors
	for _, debt := range debts {
		wallet := &Wallet{}

		err = wallet.Update(db.DB, debt.ToUserID, -debt.Amount)
		if err != nil {
			return fmt.Errorf("error updating wallet for debtors: %w", err)
		}
	}
	wallet := &Wallet{}
	//updating wallet for creditors
	err = wallet.Update(db.DB, ex.AddedBy, adjustment)
	if err != nil {
		return fmt.Errorf("error updating wallet: %w", err)
	}

	_, err = stmt.Exec(ex.Description, ex.Amount, ex.Currency, ex.Category, ex.AddedAt, ex.IsRecurring, ex.RecurringPeriod, ex.Notes, ex.Groupid, ex.AddedBy)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	return nil

}
