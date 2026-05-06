package store_test

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/davidcarrington/eagle-bank/internal/domain"
	"github.com/davidcarrington/eagle-bank/internal/store"
)

func TestCreateWithBalanceUpdate_Concurrent(t *testing.T) {
	// Use a file-based DB so multiple goroutines share the same store
	// (plain :memory: creates a private DB per connection).
	s, err := store.Open(filepath.Join(t.TempDir(), "concurrent_test.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer s.DB.Close()

	ctx := context.Background()
	userStore := store.NewUserStore(s.DB)
	accountStore := store.NewAccountStore(s.DB)
	txnStore := store.NewTransactionStore(s.DB)

	// Seed a user
	now := time.Now().UTC()
	u := &domain.User{
		ID: "usr-conctest", Name: "Test", Email: "conc@example.com",
		PasswordHash: "x", PhoneNumber: "+441234567890",
		Address:   domain.Address{Line1: "1 St", Town: "London", County: "GL", Postcode: "SW1"},
		CreatedAt: now, UpdatedAt: now,
	}
	if err := userStore.Create(ctx, u); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	// Seed an account with zero balance
	acc := &domain.Account{
		AccountNumber: "01000001", UserID: u.ID, Name: "Test",
		AccountType: "personal", Balance: 0, Currency: "GBP",
		SortCode: "10-10-10", CreatedAt: now, UpdatedAt: now,
	}
	if err := accountStore.Create(ctx, acc); err != nil {
		t.Fatalf("seed account: %v", err)
	}

	// 50 concurrent £1 deposits
	const goroutines = 50
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for i := range goroutines {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			txn := &domain.Transaction{
				ID:            "tan-" + string(rune('a'+i%26)) + string(rune('0'+i/26)) + "x",
				AccountNumber: "01000001",
				UserID:        u.ID,
				Amount:        domain.FromGBPFloat(1.00),
				Currency:      "GBP",
				Type:          domain.TransactionDeposit,
				CreatedAt:     time.Now().UTC(),
			}
			if err := txnStore.CreateWithBalanceUpdate(ctx, txn); err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("deposit goroutine error: %v", err)
	}

	// Balance must be exactly £50 (5000 pence)
	got, err := accountStore.Get(ctx, "01000001")
	if err != nil || got == nil {
		t.Fatalf("get account after concurrent deposits: %v", err)
	}
	want := domain.FromGBPFloat(50.00)
	if got.Balance != want {
		t.Errorf("balance = %v pence, want %v pence (£50.00)", got.Balance, want)
	}
}

func TestCreateWithBalanceUpdate_InsufficientFunds(t *testing.T) {
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.DB.Close()

	ctx := context.Background()
	userStore := store.NewUserStore(s.DB)
	accountStore := store.NewAccountStore(s.DB)
	txnStore := store.NewTransactionStore(s.DB)

	now := time.Now().UTC()
	u := &domain.User{
		ID: "usr-insuf", Name: "Test", Email: "insuf@example.com",
		PasswordHash: "x", PhoneNumber: "+441234567890",
		Address:   domain.Address{Line1: "1 St", Town: "London", County: "GL", Postcode: "SW1"},
		CreatedAt: now, UpdatedAt: now,
	}
	if err := userStore.Create(ctx, u); err != nil {
		t.Fatal(err)
	}

	acc := &domain.Account{
		AccountNumber: "01000002", UserID: u.ID, Name: "Test",
		AccountType: "personal", Balance: domain.FromGBPFloat(10.00),
		Currency: "GBP", SortCode: "10-10-10", CreatedAt: now, UpdatedAt: now,
	}
	if err := accountStore.Create(ctx, acc); err != nil {
		t.Fatal(err)
	}

	txn := &domain.Transaction{
		ID: "tan-insuf01", AccountNumber: "01000002", UserID: u.ID,
		Amount: domain.FromGBPFloat(50.00), Currency: "GBP",
		Type: domain.TransactionWithdrawal, CreatedAt: now,
	}
	err = txnStore.CreateWithBalanceUpdate(ctx, txn)
	if err != store.ErrInsufficientFunds {
		t.Errorf("expected ErrInsufficientFunds, got %v", err)
	}

	// Balance must be unchanged
	got, _ := accountStore.Get(ctx, "01000002")
	if got.Balance != domain.FromGBPFloat(10.00) {
		t.Errorf("balance changed after failed withdrawal: %v", got.Balance)
	}
}
