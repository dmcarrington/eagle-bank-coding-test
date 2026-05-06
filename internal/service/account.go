package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/davidcarrington/eagle-bank/internal/domain"
	"github.com/davidcarrington/eagle-bank/internal/ids"
	"github.com/davidcarrington/eagle-bank/internal/store"
)

type CreateAccountInput struct {
	Name        string
	AccountType string
}

type UpdateAccountInput struct {
	Name        *string
	AccountType *string
}

type AccountService struct {
	accounts *store.AccountStore
}

func NewAccountService(accounts *store.AccountStore) *AccountService {
	return &AccountService{accounts: accounts}
}

func (s *AccountService) Create(ctx context.Context, userID string, input CreateAccountInput) (*domain.Account, error) {
	now := time.Now().UTC()
	acc := &domain.Account{
		UserID:      userID,
		Name:        input.Name,
		AccountType: input.AccountType,
		Balance:     0,
		Currency:    "GBP",
		SortCode:    "10-10-10",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	for attempt := 0; attempt < 10; attempt++ {
		acc.AccountNumber = ids.NewAccountNumber()
		err := s.accounts.Create(ctx, acc)
		if err == nil {
			return acc, nil
		}
		if !errors.Is(err, store.ErrDuplicate) {
			return nil, err
		}
	}
	return nil, fmt.Errorf("failed to generate a unique account number")
}

func (s *AccountService) List(ctx context.Context, userID string) ([]*domain.Account, error) {
	return s.accounts.ListByUser(ctx, userID)
}

func (s *AccountService) Get(ctx context.Context, callerID, accountNumber string) (*domain.Account, error) {
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
	return acc, nil
}

func (s *AccountService) Update(ctx context.Context, callerID, accountNumber string, input UpdateAccountInput) (*domain.Account, error) {
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

	if input.Name != nil {
		acc.Name = *input.Name
	}
	if input.AccountType != nil {
		acc.AccountType = *input.AccountType
	}
	acc.UpdatedAt = time.Now().UTC()

	if err := s.accounts.Update(ctx, acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *AccountService) Delete(ctx context.Context, callerID, accountNumber string) error {
	acc, err := s.accounts.Get(ctx, accountNumber)
	if err != nil {
		return err
	}
	if acc == nil {
		return ErrNotFound
	}
	if acc.UserID != callerID {
		return ErrForbidden
	}
	return s.accounts.Delete(ctx, accountNumber)
}
