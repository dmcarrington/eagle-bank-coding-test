package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/davidcarrington/eagle-bank/internal/domain"
)

type AccountStore struct {
	db *sql.DB
}

func NewAccountStore(db *sql.DB) *AccountStore {
	return &AccountStore{db: db}
}

func (s *AccountStore) Create(ctx context.Context, a *domain.Account) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO accounts
			(account_number, user_id, name, account_type, balance_pence, currency, sort_code,
			 created_timestamp, updated_timestamp)
		VALUES (?,?,?,?,?,?,?,?,?)`,
		a.AccountNumber, a.UserID, a.Name, a.AccountType,
		int64(a.Balance), a.Currency, a.SortCode,
		a.CreatedAt.UTC().Format(time.RFC3339),
		a.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return ErrDuplicate
	}
	return err
}

func (s *AccountStore) Get(ctx context.Context, accountNumber string) (*domain.Account, error) {
	a, err := s.scanAccount(s.db.QueryRowContext(ctx, `
		SELECT account_number, user_id, name, account_type, balance_pence, currency, sort_code,
		       created_timestamp, updated_timestamp
		FROM accounts WHERE account_number = ?`, accountNumber))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return a, err
}

func (s *AccountStore) ListByUser(ctx context.Context, userID string) ([]*domain.Account, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT account_number, user_id, name, account_type, balance_pence, currency, sort_code,
		       created_timestamp, updated_timestamp
		FROM accounts WHERE user_id = ?
		ORDER BY created_timestamp ASC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*domain.Account
	for rows.Next() {
		a, err := s.scanAccountRow(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

func (s *AccountStore) Update(ctx context.Context, a *domain.Account) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE accounts SET name = ?, account_type = ?, updated_timestamp = ?
		WHERE account_number = ?`,
		a.Name, a.AccountType, a.UpdatedAt.UTC().Format(time.RFC3339), a.AccountNumber,
	)
	return err
}

func (s *AccountStore) Delete(ctx context.Context, accountNumber string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM accounts WHERE account_number = ?`, accountNumber)
	return err
}

func (s *AccountStore) CountByUser(ctx context.Context, userID string) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM accounts WHERE user_id = ?`, userID,
	).Scan(&n)
	return n, err
}

func (s *AccountStore) scanAccount(row *sql.Row) (*domain.Account, error) {
	var a domain.Account
	var balancePence int64
	var createdStr, updatedStr string
	err := row.Scan(
		&a.AccountNumber, &a.UserID, &a.Name, &a.AccountType,
		&balancePence, &a.Currency, &a.SortCode,
		&createdStr, &updatedStr,
	)
	if err != nil {
		return nil, err
	}
	a.Balance = domain.Pence(balancePence)
	a.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	a.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
	return &a, nil
}

func (s *AccountStore) scanAccountRow(rows *sql.Rows) (*domain.Account, error) {
	var a domain.Account
	var balancePence int64
	var createdStr, updatedStr string
	err := rows.Scan(
		&a.AccountNumber, &a.UserID, &a.Name, &a.AccountType,
		&balancePence, &a.Currency, &a.SortCode,
		&createdStr, &updatedStr,
	)
	if err != nil {
		return nil, err
	}
	a.Balance = domain.Pence(balancePence)
	a.CreatedAt, _ = time.Parse(time.RFC3339, createdStr)
	a.UpdatedAt, _ = time.Parse(time.RFC3339, updatedStr)
	return &a, nil
}
