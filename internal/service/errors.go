package service

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrForbidden          = errors.New("forbidden")
	ErrConflict           = errors.New("conflict")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInsufficientFunds  = errors.New("insufficient funds")
)
