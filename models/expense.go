package models

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"hirensavani.com/db"
)

type TimeSortable interface {
	GetAddedAt() time.Time
}

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
	Comment         []Comment
}

func (ex Expense) GetAddedAt() time.Time {
	return ex.AddedAt
}

func (ex *Expense) Save() error {

	// Validate user membership in group
	isMember, err := userInGroup(db.DB, ex.AddedBy, ex.Groupid)

	if !isMember || err != nil {
		return fmt.Errorf("user %d is not member of %d or system is unable to check your membership.", ex.AddedBy, ex.Groupid)
	}

	// Prepare SQL statment for expense insertion
	stmt, err := db.DB.Prepare(QueryToPostExpense)
	if err != nil {
		return fmt.Errorf("error preparing query: %w", err)
	}
	defer stmt.Close()

	// Calculate debts and adjustment
	debts, adjustment := CalculateBalance(ex)

	// Determine if debt simplification is enabled
	var simplifyDebt bool
	err = db.DB.QueryRow(QueryToGetGroupType, ex.Groupid).Scan(&simplifyDebt)

	if err != nil {
		return fmt.Errorf("error getting group type: %w", err)
	}

	if simplifyDebt {
		// Perform debt simplification concurrently
		balance := Balances{}
		res, err := balance.getBalanacesForGroup(db.DB, ex.Groupid)

		if err != nil {
			return fmt.Errorf("error gathering balances : %w", err)
		}

		// Combine existing balances with new debts
		res = append(res, debts...)

		// Calculate net balances
		netBalances := calculateNetBalances(res)

		// Separate debtors and creditors
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

	var wg sync.WaitGroup
	wg.Add(len(debts) + 1)

	//updating wallete for debtors
	for _, debt := range debts {
		go func(debt Balances) {
			defer wg.Done()

			wallet := &Wallet{}
			err = wallet.Update(db.DB, debt.ToUserID, -debt.Amount)
			if err != nil {
				log.Printf("Error updating wallet for debtor %d: %v", debt.ToUserID, err)
			}
		}(debt)

	}
	go func() {
		defer wg.Done()

		wallet := &Wallet{}
		err := wallet.Update(db.DB, ex.AddedBy, adjustment)
		if err != nil {
			log.Printf("Error updating wallet for creditor %d: %v", ex.AddedBy, err)
		}

	}()

	// Wait for all wallet updates to finish
	wg.Wait()

	_, err = stmt.Exec(ex.Description, ex.Amount, ex.Currency, ex.Category, ex.AddedAt, ex.IsRecurring, ex.RecurringPeriod, ex.Notes, ex.Groupid, ex.AddedBy)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	return nil
}

func (ex *Expense) GetExpenseByGroupId(db *sql.DB) ([]Expense, error) {

	var expenses []Expense

	rows, err := db.Query(QueryToGetExpense, ex.Groupid)

	if err != nil {
		return nil, WrapError(err, ErrExecutingQuery)
	}

	for rows.Next() {
		// ex := Expense{}
		err := rows.Scan(&ex.ID, &ex.Description, &ex.Amount, &ex.Currency, &ex.Category, &ex.AddedAt, &ex.IsRecurring, &ex.RecurringPeriod, &ex.Notes, &ex.Groupid, &ex.AddedBy)

		if err != nil {
			return nil, WrapError(err, ErrScaningRow)
		}

		var comment Comment
		comment.ExpenseID = ex.ID
		comments, err := comment.Get(db)

		if err != nil {
			return nil, WrapError(err, ErrGettingComments)
		}

		ex.Comment = comments
		//connect comments

		expenses = append(expenses, *ex)
	}

	return expenses, nil
}

func GetAllExpense(db *sql.DB, userId int64) ([]Expense, error) {

	var expenses []Expense
	// Find groups where groupmember = userId
	groupIds, err := GetGroupIdsByUserId(db, userId)

	if err != nil {
		return nil, WrapError(err, ErrGettingGroupId)
	}

	if len(groupIds) == 0 {
		return expenses, nil
	}
	// Find all expense atteched to groups

	for _, groupId := range groupIds {
		var expense Expense
		expense.Groupid = groupId
		expensesAttchedToGroupId, err := expense.GetExpenseByGroupId(db)
		if err != nil {
			return nil, WrapError(err, ErrGettingExpenses)
		}
		expenses = append(expenses, expensesAttchedToGroupId...)
	}

	SortByTime(expenses)
	return expenses, nil
}

func (ex *Expense) updateExpense(db *sql.DB) error {

	var expenseToBeUpdated Expense
	//Get Expense
	err := db.QueryRow(QueryToGetExpenseByExpenseId, ex.ID).Scan(&expenseToBeUpdated.ID, &expenseToBeUpdated.Description, &expenseToBeUpdated.Amount, &expenseToBeUpdated.Currency, &expenseToBeUpdated.Category, &expenseToBeUpdated.AddedAt, &expenseToBeUpdated.IsRecurring, &expenseToBeUpdated.RecurringPeriod, &expenseToBeUpdated.Notes, &expenseToBeUpdated.Groupid, &expenseToBeUpdated.AddedBy)

	if err != nil {
		return WrapError(err, ErrGettingExpenses)
	}
	
	
	
	//Update Expense
	_, err = db.Exec()
	//Update data from wallet , Balances and expense it self
	return nil
}
