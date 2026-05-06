package auth

import (
	"testing"
)

func TestHashPassword_RoundTrip(t *testing.T) {
	hash, err := HashPassword("correcthorsebatterystaple")
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	if err := VerifyPassword(hash, "correcthorsebatterystaple"); err != nil {
		t.Errorf("VerifyPassword correct: %v", err)
	}
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	hash, err := HashPassword("correctpassword")
	if err != nil {
		t.Fatal(err)
	}
	err = VerifyPassword(hash, "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
