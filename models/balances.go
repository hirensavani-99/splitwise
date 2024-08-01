package models

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
)

type Balances struct {
	FromUserID int64
	ToUserID   int64
	GroupId    int64
	Amount     float64
}

func (bal *Balances) Save(db *sql.DB, isCalculated bool) error {
	balance := Balances{}

	err := db.QueryRow(QueryToGetExistingBalances, bal.FromUserID, bal.ToUserID, bal.GroupId).Scan(&balance.FromUserID, &balance.ToUserID, &balance.Amount, &balance.GroupId)
	if err != nil {

		_, err = db.Exec(QueryToPostBalances, bal.FromUserID, bal.ToUserID, bal.GroupId, bal.Amount, time.Now())
		if err != nil {
			return err
		}

		return nil
	}

	var updateQuery string
	//If from userId is same -> add

	if isCalculated {

		balance.Amount = bal.Amount

		updateQuery = QueryToUpdateBalances
	} else {

		if bal.FromUserID == balance.FromUserID && bal.ToUserID == balance.ToUserID {

			balance.Amount += bal.Amount
			updateQuery = QueryToUpdateBalances
			//if fromuserid is different -> sub
		} else if bal.FromUserID == balance.ToUserID && bal.ToUserID == bal.FromUserID {

			balance.Amount -= bal.Amount

			if bal.Amount < 0 {

				balance.Amount = -bal.Amount
				updateQuery = QueryToUpdateBalancesData
			} else {
				updateQuery = QueryToUpdateBalances
			}
		}
	}
	_, err = db.Exec(updateQuery, balance.FromUserID, balance.ToUserID, balance.GroupId, balance.Amount)

	if err != nil {
		return fmt.Errorf("failed to update debt: %w", err)
	}

	if err != nil {
		return fmt.Errorf("failed to update the balance in wallete: %w", err)
	}

	return nil
}

func (bal *Balances) Get(db *sql.DB, userID int64) (map[int64]float64, error) {
	balances := make(map[int64]float64)

	rows, err := db.Query(QueryToGetBalances, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query balances: %w", err)
	}
	defer rows.Close()

	for rows.Next() {

		if err := rows.Scan(&bal.FromUserID, &bal.ToUserID, &bal.Amount); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if bal.FromUserID != userID {
			balances[bal.FromUserID] = -bal.Amount
		} else {
			balances[bal.ToUserID] = bal.Amount
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return balances, nil
}

func (bal *Balances) getBalanacesForGroup(db *sql.DB, groupid int64) ([]Balances, error) {
	balances := []Balances{}

	rows, err := db.Query(QueryToGetBalanceByGrouId, groupid)
	if err != nil {
		return nil, fmt.Errorf("failed to query balances: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&bal.FromUserID, &bal.ToUserID, &bal.GroupId, &bal.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		balances = append(balances, *bal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}
	return balances, nil
}

// Update Balances data
func UpdateBalances(db *sql.DB, AddToDataToBeUpdatedForExpense, updatedAddToDataForExpense *Expense) {

	// balance := Balances{}

	var res []Balances
	// res, err := balance.getBalanacesForGroup(db, AddToDataToBeUpdatedForExpense.Groupid)
	// fmt.Println("-->1", res)
	calculateBalanceToBeRemoved, payerPayBack := CalculateBalance(AddToDataToBeUpdatedForExpense)
	calculateBalanceToAdd, payerGetBack := CalculateBalance(updatedAddToDataForExpense)

	fmt.Println(payerPayBack, payerGetBack)
	res = append(append(res, calculateBalanceToAdd...), calculateBalanceToBeRemoved...)

	newbalances := UniqueBalances(res)

	fmt.Println(newbalances, calculateBalanceToBeRemoved, calculateBalanceToAdd)

	// Calculate net balances
	netBalances := calculateNetBalances(calculateBalanceToAdd)

	var wg sync.WaitGroup
	wg.Add(len(res) + 1)
	fmt.Println(res)

	for _, debt := range res {
		go func(debt Balances) {
			defer wg.Done()

			wallet := &Wallet{}
			err := wallet.Update(db, debt.ToUserID, -debt.Amount)
			if err != nil {
				log.Printf("Error updating wallet for debtor %d: %v", debt.ToUserID, err)
			}
		}(debt)

	}

	// Separate debtors and creditors
	creditors, debtors := separateDebtorsAndCreditors(netBalances)

	balances := minimizeTransactions(debtors, creditors, netBalances, AddToDataToBeUpdatedForExpense.Groupid)

	err := DeleteUnnecessaryBalances(balances, AddToDataToBeUpdatedForExpense.Groupid)

	if err != nil {
		log.Fatalf("Error deleting balances: %v", err)
	}

}

func NewBalances(FromUserID int64, ToUserID int64, GroupId int64, Amount float64) Balances {
	return Balances{
		FromUserID,
		ToUserID,
		GroupId,
		Amount}
}
