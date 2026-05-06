package auth

import (
	"strings"
	"testing"
	"time"
)

var testSecret = []byte("test-secret-key")

func TestSignToken_RoundTrip(t *testing.T) {
	token, expiresAt, err := SignToken(testSecret, "usr-abc123", time.Hour)
	if err != nil {
		t.Fatalf("SignToken error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if time.Until(expiresAt) < 59*time.Minute {
		t.Errorf("expiresAt too soon: %v", expiresAt)
	}

	claims, err := ParseToken(testSecret, token)
	if err != nil {
		t.Fatalf("ParseToken error: %v", err)
	}
	if claims.UserID != "usr-abc123" {
		t.Errorf("UserID = %q, want usr-abc123", claims.UserID)
	}
}

func TestParseToken_TamperedSignature(t *testing.T) {
	token, _, err := SignToken(testSecret, "usr-abc", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	// Flip the last character to tamper with the signature.
	tampered := token[:len(token)-1] + "X"
	if tampered[len(tampered)-1] == token[len(token)-1] {
		tampered = token[:len(token)-1] + "Y"
	}
	_, err = ParseToken(testSecret, tampered)
	if err != ErrTokenInvalid {
		t.Errorf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestParseToken_ExpiredToken(t *testing.T) {
	token, _, err := SignToken(testSecret, "usr-abc", -time.Second)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ParseToken(testSecret, token)
	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestParseToken_AlgNone(t *testing.T) {
	// Craft a token with alg:none by replacing the header portion.
	// A real alg:none token has header {"alg":"none"}, empty signature.
	algNoneToken := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ1c3ItYWJjIn0."
	_, err := ParseToken(testSecret, algNoneToken)
	if err != ErrTokenInvalid {
		t.Errorf("expected ErrTokenInvalid for alg:none, got %v", err)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, _, err := SignToken(testSecret, "usr-abc", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ParseToken([]byte("different-secret"), token)
	if err != ErrTokenInvalid {
		t.Errorf("expected ErrTokenInvalid, got %v", err)
	}
}

func TestSignToken_ContainsThreeParts(t *testing.T) {
	token, _, err := SignToken(testSecret, "usr-abc", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("expected 3 JWT parts, got %d", len(parts))
	}
}
