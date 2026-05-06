package store

import "errors"

var (
	// ErrDuplicate is returned when an INSERT violates a UNIQUE constraint.
	ErrDuplicate = errors.New("duplicate")
	// ErrInsufficientFunds is returned when a withdrawal would take balance below zero.
	ErrInsufficientFunds = errors.New("insufficient funds")
)
