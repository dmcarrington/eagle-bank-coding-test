package service

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrForbidden          = errors.New("forbidden")
	ErrConflict           = errors.New("conflict")
	ErrEmailTaken         = errors.New("email address is already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInsufficientFunds  = errors.New("insufficient funds")
)
