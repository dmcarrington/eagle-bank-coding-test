package domain

import "time"

const (
	TransactionDeposit    = "deposit"
	TransactionWithdrawal = "withdrawal"
)

type Transaction struct {
	ID            string
	AccountNumber string
	UserID        string
	Amount        Pence
	Currency      string
	Type          string
	Reference     string
	CreatedAt     time.Time
}
