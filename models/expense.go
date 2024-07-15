package models

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"
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
	err := db.QueryRow(QueryToCheckIsMemberOfGroup, userId, groupId).Scan(&exists)

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

	err := db.QueryRow(QueryToGetExistingBalances, bal.FromUserID, bal.ToUserID, bal.GroupId).Scan(&from_user_id, &to_user_id, &amount, &group_id)
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

		amount = bal.Amount

		updateQuery = QueryToUpdateBalances
	} else {

		if bal.FromUserID == from_user_id && bal.ToUserID == to_user_id {

			amount += bal.Amount
			updateQuery = QueryToUpdateBalances
			//if fromuserid is different -> sub
		} else if bal.FromUserID == to_user_id && bal.ToUserID == from_user_id {

			amount -= bal.Amount

			if amount < 0 {

				amount = -amount
				updateQuery = QueryToUpdateBalancesData
			} else {
				updateQuery = QueryToUpdateBalances
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

		res, err := getBalanacesForGroup(ex.Groupid)

		res = append(res, debts...)

		if err != nil {
			return fmt.Errorf("error geathering balances : %w", err)
		}
		netBalances := calculateNetBalances(res)

		creditors, debtors := separateDebtorsAndCreditors(netBalances)

		balances := minimizeTransactions(debtors, creditors, netBalances, ex.Groupid)

		err = deleteOtherRecords(balances, ex.Groupid)
		if err != nil {
			log.Fatalf("Error deleting balances: %v", err)
		}
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

func getBalanacesForGroup(groupid int64) ([]Balances, error) {
	balance := Balances{}
	balances := []Balances{}

	query := QueryToGetBalanceByGrouId

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

func minimizeTransactions(debtors []int64, creditors []int64, netBalances map[int64]float64, groupId int64) []Balances {

	i, j := 0, 0

	overAllBalances := []Balances{}

	for i < len(creditors) && j < len(debtors) {
		debtAmount := -netBalances[debtors[j]]
		creditAmount := netBalances[creditors[i]]

		settleAmount := math.Min(debtAmount, creditAmount)

		netBalances[debtors[j]] += settleAmount
		netBalances[creditors[i]] -= settleAmount

		var debt Balances
		debt.FromUserID = creditors[i]
		debt.ToUserID = debtors[j]
		debt.GroupId = groupId
		debt.Amount = settleAmount

		overAllBalances = append(overAllBalances, debt)

		fmt.Printf("User %d pays User %d: %.2f\n", creditors[i], debtors[j], settleAmount)
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
	return overAllBalances
}

func deleteOtherRecords(keepRecords []Balances, groupId int64) error {
	// Construct the SQL query
	query := QueryToDeleteUnnecessaryBalances

	var records []string
	for _, record := range keepRecords {
		tuple := fmt.Sprintf("(%d,%d,%d,%f)", record.FromUserID, record.ToUserID, record.GroupId, record.Amount)
		records = append(records, tuple)
	}

	query = query + "(" + strings.Join(records, ",") + ")"
	// Execute the DELETE statement

	_, err := db.DB.Exec(QueryToDeleteUnnecessaryBalances, groupId)
	if err != nil {
		return fmt.Errorf("failed to delete records: %w", err)
	}

	return nil
}
