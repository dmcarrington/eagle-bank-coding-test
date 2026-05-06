package ids

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
)

func NewUserID() string {
	return "usr-" + randomHex(12)
}

func NewTransactionID() string {
	return "tan-" + randomHex(12)
}

// NewAccountNumber returns a random account number matching ^01\d{6}$.
// The caller should retry on a UNIQUE constraint violation.
func NewAccountNumber() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		panic(fmt.Sprintf("ids: failed to generate account number: %v", err))
	}
	return fmt.Sprintf("01%06d", n.Int64())
}

func randomHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("ids: failed to read random bytes: %v", err))
	}
	return hex.EncodeToString(b)
}
