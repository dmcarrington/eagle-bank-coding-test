package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

// mustCreateTransaction deposits or withdraws via the API and returns the transaction ID.
func (e *testEnv) mustCreateTransaction(t *testing.T, token, accNum, txType string, amount float64) string {
	t.Helper()
	w := e.do(t, http.MethodPost, "/v1/accounts/"+accNum+"/transactions", token, map[string]any{
		"amount":   amount,
		"currency": "GBP",
		"type":     txType,
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("mustCreateTransaction: status %d, body: %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	return body["id"].(string)
}

// --- Create transaction ---

func TestCreateTransaction_Deposit_Success(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "dep1@example.com", "password123")
	accNum := env.mustCreateAccount(t, token, "Main")

	w := env.do(t, http.MethodPost, "/v1/accounts/"+accNum+"/transactions", token, map[string]any{
		"amount": 100.00, "currency": "GBP", "type": "deposit", "reference": "salary",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["type"] != "deposit" {
		t.Errorf("type = %v, want deposit", body["type"])
	}
	if body["amount"] != 100.00 {
		t.Errorf("amount = %v, want 100", body["amount"])
	}
	if body["reference"] != "salary" {
		t.Errorf("reference = %v, want salary", body["reference"])
	}

	// Balance must have increased
	acc := decodeBody(t, env.do(t, http.MethodGet, "/v1/accounts/"+accNum, token, nil))
	if acc["balance"] != 100.00 {
		t.Errorf("balance after deposit = %v, want 100", acc["balance"])
	}
}

func TestCreateTransaction_Withdrawal_Success(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "wit1@example.com", "password123")
	accNum := env.mustCreateAccount(t, token, "Main")
	env.mustCreateTransaction(t, token, accNum, "deposit", 200.00)

	w := env.do(t, http.MethodPost, "/v1/accounts/"+accNum+"/transactions", token, map[string]any{
		"amount": 75.50, "currency": "GBP", "type": "withdrawal",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", w.Code, w.Body.String())
	}

	acc := decodeBody(t, env.do(t, http.MethodGet, "/v1/accounts/"+accNum, token, nil))
	if acc["balance"] != 124.50 {
		t.Errorf("balance after withdrawal = %v, want 124.50", acc["balance"])
	}
}

func TestCreateTransaction_Withdrawal_InsufficientFunds_Returns422(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "insuf1@example.com", "password123")
	accNum := env.mustCreateAccount(t, token, "Main")
	env.mustCreateTransaction(t, token, accNum, "deposit", 10.00)

	w := env.do(t, http.MethodPost, "/v1/accounts/"+accNum+"/transactions", token, map[string]any{
		"amount": 50.00, "currency": "GBP", "type": "withdrawal",
	})
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422; body: %s", w.Code, w.Body.String())
	}

	// Balance must be unchanged
	acc := decodeBody(t, env.do(t, http.MethodGet, "/v1/accounts/"+accNum, token, nil))
	if acc["balance"] != 10.00 {
		t.Errorf("balance changed after failed withdrawal: %v", acc["balance"])
	}
}

func TestCreateTransaction_OtherUserAccount_Returns403(t *testing.T) {
	env := newTestEnv(t)
	_, token1 := env.mustCreateUser(t, "txother1@example.com", "password123")
	_, token2 := env.mustCreateUser(t, "txother2@example.com", "password123")
	accNum := env.mustCreateAccount(t, token2, "Bob's Account")

	w := env.do(t, http.MethodPost, "/v1/accounts/"+accNum+"/transactions", token1, map[string]any{
		"amount": 10.00, "currency": "GBP", "type": "deposit",
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestCreateTransaction_NonExistentAccount_Returns404(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "txnone@example.com", "password123")

	w := env.do(t, http.MethodPost, "/v1/accounts/01999990/transactions", token, map[string]any{
		"amount": 10.00, "currency": "GBP", "type": "deposit",
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestCreateTransaction_MissingFields_Returns400(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "txbad@example.com", "password123")
	accNum := env.mustCreateAccount(t, token, "Main")

	w := env.do(t, http.MethodPost, "/v1/accounts/"+accNum+"/transactions", token, map[string]any{
		"currency": "GBP",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	body := decodeBody(t, w)
	if body["details"] == nil {
		t.Error("response missing details")
	}
}

func TestCreateTransaction_InvalidType_Returns400(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "txtype@example.com", "password123")
	accNum := env.mustCreateAccount(t, token, "Main")

	w := env.do(t, http.MethodPost, "/v1/accounts/"+accNum+"/transactions", token, map[string]any{
		"amount": 10.00, "currency": "GBP", "type": "transfer",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

// --- List transactions ---

func TestListTransactions_OwnAccount(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "listx1@example.com", "password123")
	accNum := env.mustCreateAccount(t, token, "Main")
	env.mustCreateTransaction(t, token, accNum, "deposit", 50.00)
	env.mustCreateTransaction(t, token, accNum, "deposit", 25.00)

	w := env.do(t, http.MethodGet, "/v1/accounts/"+accNum+"/transactions", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Transactions []map[string]any `json:"transactions"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Transactions) != 2 {
		t.Errorf("got %d transactions, want 2", len(resp.Transactions))
	}
}

func TestListTransactions_OtherUserAccount_Returns403(t *testing.T) {
	env := newTestEnv(t)
	_, token1 := env.mustCreateUser(t, "listx2a@example.com", "password123")
	_, token2 := env.mustCreateUser(t, "listx2b@example.com", "password123")
	accNum := env.mustCreateAccount(t, token2, "Bob")

	w := env.do(t, http.MethodGet, "/v1/accounts/"+accNum+"/transactions", token1, nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestListTransactions_NonExistentAccount_Returns404(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "listx3@example.com", "password123")

	w := env.do(t, http.MethodGet, "/v1/accounts/01999989/transactions", token, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

// --- Fetch transaction ---

func TestFetchTransaction_OwnAccount_Returns200(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "ftx1@example.com", "password123")
	accNum := env.mustCreateAccount(t, token, "Main")
	txnID := env.mustCreateTransaction(t, token, accNum, "deposit", 30.00)

	w := env.do(t, http.MethodGet, "/v1/accounts/"+accNum+"/transactions/"+txnID, token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["id"] != txnID {
		t.Errorf("id = %v, want %s", body["id"], txnID)
	}
	if body["amount"] != 30.00 {
		t.Errorf("amount = %v, want 30", body["amount"])
	}
}

func TestFetchTransaction_OtherUserAccount_Returns403(t *testing.T) {
	env := newTestEnv(t)
	_, token1 := env.mustCreateUser(t, "ftx2a@example.com", "password123")
	_, token2 := env.mustCreateUser(t, "ftx2b@example.com", "password123")
	accNum := env.mustCreateAccount(t, token2, "Bob")
	txnID := env.mustCreateTransaction(t, token2, accNum, "deposit", 10.00)

	w := env.do(t, http.MethodGet, "/v1/accounts/"+accNum+"/transactions/"+txnID, token1, nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestFetchTransaction_NonExistentAccount_Returns404(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "ftx3@example.com", "password123")

	w := env.do(t, http.MethodGet, "/v1/accounts/01999988/transactions/tan-abc123", token, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestFetchTransaction_NonExistentID_Returns404(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "ftx4@example.com", "password123")
	accNum := env.mustCreateAccount(t, token, "Main")
	env.mustCreateTransaction(t, token, accNum, "deposit", 10.00)

	w := env.do(t, http.MethodGet, "/v1/accounts/"+accNum+"/transactions/tan-doesnotexist", token, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestFetchTransaction_TxBelongsToDifferentAccount_Returns404(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "ftx5@example.com", "password123")
	accNum1 := env.mustCreateAccount(t, token, "Account One")
	accNum2 := env.mustCreateAccount(t, token, "Account Two")

	// Create transaction on account 1, then try to fetch it via account 2.
	txnID := env.mustCreateTransaction(t, token, accNum1, "deposit", 10.00)

	w := env.do(t, http.MethodGet, "/v1/accounts/"+accNum2+"/transactions/"+txnID, token, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (txn belongs to different account)", w.Code)
	}
}
