package models

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"hirensavani.com/db"
)

type TimeSortable interface {
	GetAddedAt() time.Time
}

type Expense struct {
	ID              int64             `json:"id"`
	Groupid         int64             `json:"group_id"`
	AddedBy         int64             `json:"added_by"`
	Description     string            `json:"description"`
	AddedAt         time.Time         `json:"added_at"`
	Amount          float64           `json:"amount"`
	Currency        string            `json:"currency"`
	Category        string            `json:"category"`
	IsRecurring     bool              `json:"is_recurring"`
	RecurringPeriod string            `json:"recurring_period"`
	Notes           string            `json:"notes"`
	Tags            []string          `json:"tags"`
	AddTo           map[string]string `json:"add_to"`
	SplitType       string            `json:"split_type"`
	Comment         []Comment         `json:"comments"`
}

func (ex Expense) GetAddedAt() time.Time {
	return ex.AddedAt
}

func (ex *Expense) Save() error {

	// Validate user membership in group
	isMember, err := userInGroup(db.DB, ex.AddedBy, ex.Groupid)

	fmt.Println(ex, err)

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

func UpdateExpense(db *sql.DB, ex map[string]interface{}, expenseId int64) error {

	var expenseToBeUpdated Expense

	//Get Expense
	err := db.QueryRow(QueryToGetExpenseByExpenseId, expenseId).Scan(&expenseToBeUpdated.ID, &expenseToBeUpdated.Description, &expenseToBeUpdated.Amount, &expenseToBeUpdated.Currency, &expenseToBeUpdated.Category, &expenseToBeUpdated.AddedAt, &expenseToBeUpdated.IsRecurring, &expenseToBeUpdated.RecurringPeriod, &expenseToBeUpdated.Notes, &expenseToBeUpdated.Groupid, &expenseToBeUpdated.AddedBy)

	if err != nil {
		return WrapError(err, ErrGettingExpenses)
	}

	fmt.Println(ex)

	query, err := buildUpdateQuery(ex, expenseId)
	if err != nil {

		return WrapError(err, "Error building update query:")
	}

	fmt.Println("Query:", query)

	//Update Expense
	_, err = db.Exec(query)

	fmt.Println(query)
	//Update data from wallet , Balances and expense it self
	if err != nil {
		return WrapError(err, ErrExecutingQuery)
	}
	return nil
}

func buildUpdateQuery(expense map[string]interface{}, expenseId int64) (string, error) {
	query := "UPDATE expense SET "
	var setClauses []string

	for key, value := range expense {
		if key == "ID" || key == "Comment" || key == "Tags" || key == "AddTo" || key == "SplitType" || key == "Groupid" || key == "AddedBy" {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = '%v'", key, value))
	}

	query += strings.Join(setClauses, ", ")
	query += fmt.Sprintf(" WHERE id = %d", expenseId)

	return query, nil
}
