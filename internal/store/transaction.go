package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/davidcarrington/eagle-bank/internal/domain"
)

type TransactionStore struct {
	db *sql.DB
}

func NewTransactionStore(db *sql.DB) *TransactionStore {
	return &TransactionStore{db: db}
}

// CreateWithBalanceUpdate atomically inserts a transaction and updates the account
// balance inside a single database transaction. Returns ErrInsufficientFunds if a
// withdrawal would take the balance below zero.
func (s *TransactionStore) CreateWithBalanceUpdate(ctx context.Context, txn *domain.Transaction) error {
	dbtx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			dbtx.Rollback()
		}
	}()

	var currentPence int64
	err = dbtx.QueryRowContext(ctx,
		`SELECT balance_pence FROM accounts WHERE account_number = ?`,
		txn.AccountNumber,
	).Scan(&currentPence)
	if err != nil {
		return err
	}

	var newPence int64
	switch txn.Type {
	case domain.TransactionDeposit:
		newPence = currentPence + int64(txn.Amount)
	case domain.TransactionWithdrawal:
		if int64(txn.Amount) > currentPence {
			err = ErrInsufficientFunds
			return err
		}
		newPence = currentPence - int64(txn.Amount)
	}

	_, err = dbtx.ExecContext(ctx, `
		INSERT INTO transactions (id, account_number, user_id, amount_pence, currency, type, reference, created_timestamp)
		VALUES (?,?,?,?,?,?,?,?)`,
		txn.ID, txn.AccountNumber, txn.UserID,
		int64(txn.Amount), txn.Currency, txn.Type,
		nullString(txn.Reference),
		txn.CreatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	_, err = dbtx.ExecContext(ctx,
		`UPDATE accounts SET balance_pence = ?, updated_timestamp = ? WHERE account_number = ?`,
		newPence, time.Now().UTC().Format(time.RFC3339), txn.AccountNumber,
	)
	if err != nil {
		return err
	}

	return dbtx.Commit()
}

func (s *TransactionStore) ListByAccount(ctx context.Context, accountNumber string) ([]*domain.Transaction, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, account_number, user_id, amount_pence, currency, type, reference, created_timestamp
		FROM transactions WHERE account_number = ?
		ORDER BY created_timestamp ASC`, accountNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []*domain.Transaction
	for rows.Next() {
		t, err := scanTxnRow(rows)
		if err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	return txns, rows.Err()
}

func (s *TransactionStore) Get(ctx context.Context, txnID string) (*domain.Transaction, error) {
	t, err := scanTxnSingle(s.db.QueryRowContext(ctx, `
		SELECT id, account_number, user_id, amount_pence, currency, type, reference, created_timestamp
		FROM transactions WHERE id = ?`, txnID))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return t, err
}

func scanTxnSingle(row *sql.Row) (*domain.Transaction, error) {
	var t domain.Transaction
	var amountPence int64
	var createdStr string
	var reference sql.NullString
	err := row.Scan(
		&t.ID, &t.AccountNumber, &t.UserID,
		&amountPence, &t.Currency, &t.Type, &reference, &createdStr,
	)
	if err != nil {
		return nil, err
	}
	t.Amount = domain.Pence(amountPence)
	t.Reference = reference.String
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	return &t, nil
}

func scanTxnRow(rows *sql.Rows) (*domain.Transaction, error) {
	var t domain.Transaction
	var amountPence int64
	var createdStr string
	var reference sql.NullString
	err := rows.Scan(
		&t.ID, &t.AccountNumber, &t.UserID,
		&amountPence, &t.Currency, &t.Type, &reference, &createdStr,
	)
	if err != nil {
		return nil, err
	}
	t.Amount = domain.Pence(amountPence)
	t.Reference = reference.String
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	return &t, nil
}

