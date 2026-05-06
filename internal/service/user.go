package service

import (
	"context"
	"errors"
	"time"

	"github.com/davidcarrington/eagle-bank/internal/auth"
	"github.com/davidcarrington/eagle-bank/internal/domain"
	"github.com/davidcarrington/eagle-bank/internal/ids"
	"github.com/davidcarrington/eagle-bank/internal/store"
)

type CreateUserInput struct {
	Name        string
	Email       string
	Password    string
	PhoneNumber string
	Address     domain.Address
}

type UpdateUserInput struct {
	Name        *string
	Email       *string
	PhoneNumber *string
	Address     *domain.Address
}

type UserService struct {
	users    *store.UserStore
	accounts *store.AccountStore
}

func NewUserService(users *store.UserStore, accounts *store.AccountStore) *UserService {
	return &UserService{users: users, accounts: accounts}
}

func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*domain.User, error) {
	hash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	u := &domain.User{
		ID:           ids.NewUserID(),
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: hash,
		PhoneNumber:  input.PhoneNumber,
		Address:      input.Address,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Authenticate(ctx context.Context, email, password string) (*domain.User, error) {
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrInvalidCredentials
	}
	if err := auth.VerifyPassword(u.PasswordHash, password); err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	return u, nil
}

func (s *UserService) Get(ctx context.Context, callerID, targetID string) (*domain.User, error) {
	u, err := s.users.FindByID(ctx, targetID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrNotFound
	}
	if callerID != targetID {
		return nil, ErrForbidden
	}
	return u, nil
}

func (s *UserService) Update(ctx context.Context, callerID, targetID string, input UpdateUserInput) (*domain.User, error) {
	u, err := s.users.FindByID(ctx, targetID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrNotFound
	}
	if callerID != targetID {
		return nil, ErrForbidden
	}

	if input.Name != nil {
		u.Name = *input.Name
	}
	if input.Email != nil {
		u.Email = *input.Email
	}
	if input.PhoneNumber != nil {
		u.PhoneNumber = *input.PhoneNumber
	}
	if input.Address != nil {
		u.Address = *input.Address
	}
	u.UpdatedAt = time.Now().UTC()

	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *UserService) Delete(ctx context.Context, callerID, targetID string) error {
	u, err := s.users.FindByID(ctx, targetID)
	if err != nil {
		return err
	}
	if u == nil {
		return ErrNotFound
	}
	if callerID != targetID {
		return ErrForbidden
	}

	count, err := s.accounts.CountByUser(ctx, targetID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrConflict
	}

	return s.users.Delete(ctx, targetID)
}
