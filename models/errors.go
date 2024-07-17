package models

import "fmt"

const (
	ErrExecutingQuery  = `An Error ocurred while Executting Query : %w`
	ErrScaningRow      = `An Error ocurred while Scanning row : %w`
	ErrGettingGroupId  = `An Error ocurred while getting groupids by userids : %w`
	ErrGettingExpenses = `An Error ocurred while getting Expenses : %w`
	ErrExpNotExists    = `An Error occured while checking Expense exists or not`
)

func WrapError(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}

func WrapErrMessage(msg string) error {
	return fmt.Errorf("%s", msg)
}