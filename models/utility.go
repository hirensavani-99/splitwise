package models

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"strings"

	"hirensavani.com/db"
)

// Check expense with expense id exists or not
func IsExpense(db *sql.DB, expenseId int64) bool {
	var exists bool
	err := db.QueryRow(QueryToCheckIsExpenseExists, expenseId).Scan(&exists)

	if err != nil {
		return false
	}
	return exists
}

// Check user is part of group or not
func userInGroup(db *sql.DB, userId int64, groupId int64) (bool, error) {
	var exists bool
	err := db.QueryRow(QueryToCheckIsMemberOfGroup, userId, groupId).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

// Calculate Balance to add fo each user
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

			balances = append(balances, NewBalances(
				expense.AddedBy,
				userID,
				expense.Groupid,
				amountPerUser))

		}
	}

	return balances, payerPayBackAmount
}

// Claculate net balance for each user
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
		err := debt.Save(db.DB, true)
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

func DeleteUnnecessaryBalances(keepRecords []Balances, groupId int64) error {
	// Construct the SQL query
	query := QueryToDeleteUnnecessaryBalances

	var records []string
	for _, record := range keepRecords {
		tuple := fmt.Sprintf("(%d,%d,%d,%f)", record.FromUserID, record.ToUserID, record.GroupId, record.Amount)
		records = append(records, tuple)
	}

	// If there is no balances which needs to be deleted
	if len(records) == 0 {
		return nil
	}
	query = query + "(" + strings.Join(records, ",") + ");"
	// Execute the DELETE statement

	_, err := db.DB.Exec(query, groupId)
	if err != nil {
		return fmt.Errorf("failed to delete records: %w", err)
	}

	return nil
}

func SortByTime[T TimeSortable](listOfItem []T) {
	sort.SliceStable(listOfItem, func(i, j int) bool {
		return !listOfItem[i].GetAddedAt().Before(listOfItem[j].GetAddedAt())
	})
}
