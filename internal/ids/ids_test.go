package ids

import (
	"regexp"
	"testing"
)

var (
	userIDRe    = regexp.MustCompile(`^usr-[0-9a-f]{24}$`)
	txIDRe      = regexp.MustCompile(`^tan-[0-9a-f]{24}$`)
	accNumberRe = regexp.MustCompile(`^01\d{6}$`)
)

func TestNewUserID(t *testing.T) {
	for range 100 {
		id := NewUserID()
		if !userIDRe.MatchString(id) {
			t.Errorf("NewUserID() = %q, want match %s", id, userIDRe)
		}
	}
}

func TestNewTransactionID(t *testing.T) {
	for range 100 {
		id := NewTransactionID()
		if !txIDRe.MatchString(id) {
			t.Errorf("NewTransactionID() = %q, want match %s", id, txIDRe)
		}
	}
}

func TestNewAccountNumber(t *testing.T) {
	for range 100 {
		n := NewAccountNumber()
		if !accNumberRe.MatchString(n) {
			t.Errorf("NewAccountNumber() = %q, want match %s", n, accNumberRe)
		}
	}
}
