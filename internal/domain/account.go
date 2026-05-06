package domain

import "time"

type Account struct {
	AccountNumber string
	UserID        string
	Name          string
	AccountType   string
	Balance       Pence
	Currency      string
	SortCode      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
