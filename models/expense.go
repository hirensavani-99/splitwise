package models

import "time"

type Expense struct {
	ID              int64
	Groupid         int64
	AddedBy         int64
	Description     string
	AddedAt         time.Time
	Amount          float64
	currency        string
	category        string
	IsRecurring     bool
	RecurringPeriod string
	Notes           string
	Tags            []string
	AddTo           map[int64]string
	spiltType       string
}
