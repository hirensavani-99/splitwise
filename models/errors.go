package models

import "fmt"

const (
	ErrExecutingQuery       = `An Error ocurred while Executting Query : %w`
	ErrScaningRow           = `An Error ocurred while Scanning row : %w`
	ErrGettingGroupId       = `An Error ocurred while getting groupids by userids : %w`
	ErrGettingExpenses      = `An Error ocurred while getting Expenses : %w`
	ErrExpNotExists         = `An Error occurred while checking Expense exists or not`
	ErrGettingComments      = `An Error ocurred while getting comments for expense %w`
	ErrMarshaling           = `An Error ocurred while marshaling to JSON: % w `
	ErrUnMarshaling         = `An Error ocurred while unmarshalling JSON: %w `
	ErrBuildingQuery        = `An Error ocuured while building query %w`
	ErrGettingGroupBalances = `An Error occurred while getting group Balances %w`
	ErrUpdatingWallet       = `An Error occurred while updatiing wallet %w`
)

func WrapError(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}

func WrapErrMessage(msg string) error {
	return fmt.Errorf("%s", msg)
}
