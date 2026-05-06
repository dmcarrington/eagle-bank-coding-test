package service

import (
	"context"
	"errors"
	"time"

	"github.com/davidcarrington/eagle-bank/internal/domain"
	"github.com/davidcarrington/eagle-bank/internal/ids"
	"github.com/davidcarrington/eagle-bank/internal/store"
)

type CreateTransactionInput struct {
	Amount    domain.Pence
	Currency  string
	Type      string
	Reference string
}

type TransactionService struct {
	accounts *store.AccountStore
	txns     *store.TransactionStore
}

func NewTransactionService(accounts *store.AccountStore, txns *store.TransactionStore) *TransactionService {
	return &TransactionService{accounts: accounts, txns: txns}
}

func (s *TransactionService) Create(ctx context.Context, callerID, accountNumber string, input CreateTransactionInput) (*domain.Transaction, error) {
	acc, err := s.accounts.Get(ctx, accountNumber)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, ErrNotFound
	}
	if acc.UserID != callerID {
		return nil, ErrForbidden
	}

	txn := &domain.Transaction{
		ID:            ids.NewTransactionID(),
		AccountNumber: accountNumber,
		UserID:        callerID,
		Amount:        input.Amount,
		Currency:      input.Currency,
		Type:          input.Type,
		Reference:     input.Reference,
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.txns.CreateWithBalanceUpdate(ctx, txn); err != nil {
		if errors.Is(err, store.ErrInsufficientFunds) {
			return nil, ErrInsufficientFunds
		}
		return nil, err
	}
	return txn, nil
}

func (s *TransactionService) List(ctx context.Context, callerID, accountNumber string) ([]*domain.Transaction, error) {
	acc, err := s.accounts.Get(ctx, accountNumber)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, ErrNotFound
	}
	if acc.UserID != callerID {
		return nil, ErrForbidden
	}
	return s.txns.ListByAccount(ctx, accountNumber)
}

func (s *TransactionService) Get(ctx context.Context, callerID, accountNumber, txnID string) (*domain.Transaction, error) {
	acc, err := s.accounts.Get(ctx, accountNumber)
	if err != nil {
		return nil, err
	}
	if acc == nil {
		return nil, ErrNotFound
	}
	if acc.UserID != callerID {
		return nil, ErrForbidden
	}

	txn, err := s.txns.Get(ctx, txnID)
	if err != nil {
		return nil, err
	}
	// 404 if transaction doesn't exist OR belongs to a different account.
	if txn == nil || txn.AccountNumber != accountNumber {
		return nil, ErrNotFound
	}
	return txn, nil
}
