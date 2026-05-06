package auth

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

var ErrInvalidCredentials = errors.New("invalid credentials")

func HashPassword(plaintext string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func VerifyPassword(hash, plaintext string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return ErrInvalidCredentials
	}
	return err
}
