package models

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strconv"
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

func userInGroup(db *sql.DB, userId int64, groupId int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM group_member WHERE user_id = $1 AND group_id = $2)`
	err := db.QueryRow(query, userId, groupId).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func CalculateBalance(expense *Expense) ([]Balances, float64) {
	payerPayBackAmount := expense.Amount
	totalAmount := expense.Amount
	numUsers := len(expense.AddTo)

	var balances []Balances

	for userId, _ := range expense.AddTo {

		userID, _ := strconv.ParseInt(userId, 10, 64)

		switch expense.SplitType {

		case "equal":
			amountPerUser := totalAmount / float64(numUsers)
			if userID == expense.AddedBy {
				payerPayBackAmount -= amountPerUser
				continue
			}

			balances = append(balances, Balances{
				FromUserID: expense.AddedBy,
				ToUserID:   userID,
				GroupId:    expense.Groupid,
				Amount:     amountPerUser,
			})

		}
	}

	return balances, payerPayBackAmount
}

func insertDebt(db *sql.DB, bal Balances, isCalculated bool) error {
	var amount float64
	var from_user_id int64
	var to_user_id int64
	var group_id int64
	queryToGetExistingRecord := `SELECT from_user_id, to_user_id, amount , group_id FROM BALANCES WHERE ((from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1)) AND group_id=$3;`

	err := db.QueryRow(queryToGetExistingRecord, bal.FromUserID, bal.ToUserID, bal.GroupId).Scan(&from_user_id, &to_user_id, &amount, &group_id)
	if err != nil {
		query := `
        INSERT INTO BALANCES (from_user_id, to_user_id ,group_id, amount, created_at)
        VALUES ($1, $2, $3, $4,$5)
    `

		_, err = db.Exec(query, bal.FromUserID, bal.ToUserID, bal.GroupId, bal.Amount, time.Now())
		if err != nil {
			return err
		}

		return nil
	}

	var updateQuery string
	//If from userId is same -> add

	if isCalculated {
		fmt.Println("---> amount", bal)
		amount = bal.Amount
		if bal.Amount > 0 {
			updateQuery = `UPDATE BALANCES SET amount=$4 WHERE ((from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1)) AND group_id=$3;`
		} else {
			updateQuery = `UPDATE BALANCES SET amount=$4 WHERE ((from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1) AND group_id=$3);`
		}
	} else {

		if bal.FromUserID == from_user_id && bal.ToUserID == to_user_id {

			amount += bal.Amount
			updateQuery = `UPDATE BALANCES SET amount=$4 WHERE ((from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1)) AND group_id=$3;`
			//if fromuserid is different -> sub
		} else if bal.FromUserID == to_user_id && bal.ToUserID == from_user_id {

			amount -= bal.Amount

			if amount < 0 {

				amount = -amount
				updateQuery = `
		UPDATE BALANCES
		SET amount=$4, from_user_id=$2, to_user_id=$1
		WHERE ((from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1) AND group_id=$3);
	`
			} else {
				updateQuery = `UPDATE BALANCES SET amount=$4 WHERE ((from_user_id = $1 AND to_user_id = $2) OR (from_user_id = $2 AND to_user_id = $1) AND group_id=$3);`
			}
		}
	}
	_, err = db.Exec(updateQuery, from_user_id, to_user_id, group_id, amount)

	if err != nil {
		return fmt.Errorf("failed to update debt: %w", err)
	}

	if err != nil {
		return fmt.Errorf("failed to update the balance in wallete: %w", err)
	}

	return nil
}

func updateWallet(userid int64, adjustment float64) error {
	var currentBalance float64
	querySelect := `SELECT BALANCE FROM Wallets WHERE USER_ID=$1`
	err := db.DB.QueryRow(querySelect, userid).Scan(&currentBalance)
	if err != nil {
		return fmt.Errorf("failed to get current balance: %w", err)
	}

	totalWalletBalance := currentBalance + adjustment
	//Update wallete
	updateWallete := `UPDATE Wallets SET BALANCE=$2 WHERE USER_ID=$1`

	_, err = db.DB.Exec(updateWallete, userid, totalWalletBalance)

	if err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}
	return nil
}

func (ex *Expense) Save() error {

	isMember, err := userInGroup(db.DB, ex.AddedBy, ex.Groupid)

	if err != nil {
		return err
	}

	if !isMember {
		return fmt.Errorf("user %d is not member of %d", ex.AddedBy, ex.Groupid)
	}

	query := `
		INSERT INTO expense (
			description, amount, currency, category, added_at,
			is_recurring, recurring_period, notes, group_id, added_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id`

	stmt, err := db.DB.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}
	defer stmt.Close()

	debts, adjustment := CalculateBalance(ex)

	var simplifyDebt bool
	queryToGetGroupType := `Select simplify_debt from groups where id=$1 `

	err = db.DB.QueryRow(queryToGetGroupType, ex.Groupid).Scan(&simplifyDebt)

	if err != nil {
		return fmt.Errorf("error getting group type: %w", err)
	}

	if simplifyDebt {

		res, err := getBalanacesForGroup(ex.Groupid)

		res = append(res, debts...)
		if err != nil {
			return fmt.Errorf("error geathering balances : %w", err)
		}
		netBalances := calculateNetBalances(res)

		debtors, creditors := separateDebtorsAndCreditors(netBalances)

		minimizeTransactions(debtors, creditors, netBalances, ex.Groupid)
	} else {

		for _, debt := range debts {

			err := insertDebt(db.DB, debt, false)
			if err != nil {
				log.Fatalf("Error inserting debt: %v", err)
			}
		}
	}

	//updating wallete for debtors
	for _, debt := range debts {

		err = updateWallet(debt.ToUserID, -debt.Amount)
		if err != nil {
			return fmt.Errorf("error updating wallet for debtors: %w", err)
		}
	}

	//updating wallet for creditors
	err = updateWallet(ex.AddedBy, adjustment)
	if err != nil {
		return fmt.Errorf("error updating wallet: %w", err)
	}

	_, err = stmt.Exec(ex.Description, ex.Amount, ex.Currency, ex.Category, ex.AddedAt, ex.IsRecurring, ex.RecurringPeriod, ex.Notes, ex.Groupid, ex.AddedBy)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	return nil

}

func getBalanacesForGroup(groupid int64) ([]Balances, error) {
	balance := Balances{}
	balances := []Balances{}

	query := `Select from_user_id , to_user_id , group_id , amount from balances where group_id=$1`

	rows, err := db.DB.Query(query, groupid)
	if err != nil {
		return nil, fmt.Errorf("failed to query balances: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&balance.FromUserID, &balance.ToUserID, &balance.GroupId, &balance.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		balances = append(balances, balance)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}
	return balances, nil
}

func calculateNetBalances(transcations []Balances) map[int64]float64 {
	netBalance := make(map[int64]float64)

	for _, t := range transcations {
		netBalance[t.FromUserID] += t.Amount
		netBalance[t.ToUserID] -= t.Amount
	}

	return netBalance
}

func separateDebtorsAndCreditors(netBalances map[int64]float64) ([]int64, []int64) {
	creditors := []int64{}
	debtors := []int64{}

	for userId, amount := range netBalances {
		if amount > 0 {
			creditors = append(creditors, userId)
		} else if amount < 0 {
			debtors = append(debtors, userId)
		}
	}

	sort.Slice(creditors, func(i, j int) bool {
		return netBalances[creditors[i]] > netBalances[creditors[j]]
	})

	sort.Slice(debtors, func(i, j int) bool {
		return netBalances[debtors[i]] < netBalances[debtors[j]]
	})

	return creditors, debtors
}

func minimizeTransactions(debtors []int64, creditors []int64, netBalances map[int64]float64, groupId int64) {
	i, j := 0, 0

	for i < len(creditors) && j < len(debtors) {

		debtAmount := netBalances[debtors[j]]
		creditAmount := netBalances[creditors[i]]

		settleAmount := min(debtAmount, creditAmount)

		netBalances[debtors[j]] += settleAmount
		netBalances[creditors[i]] -= settleAmount

		var debt Balances
		debt.FromUserID = debtors[j]
		debt.ToUserID = creditors[i]
		debt.GroupId = groupId
		debt.Amount = -settleAmount

		fmt.Printf("User %d pays User %d: %.2f\n", debtors[j], creditors[i], settleAmount)
		err := insertDebt(db.DB, debt, true)
		if err != nil {
			log.Fatalf("Error inserting debt: %v", err)
		}

		if netBalances[creditors[i]] == 0 {
			i++
		}

		if netBalances[debtors[j]] == 0 {
			j++
		}
	}
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
