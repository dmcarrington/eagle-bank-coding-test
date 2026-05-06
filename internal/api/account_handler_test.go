package api_test

import (
	"net/http"
	"regexp"
	"testing"
)

var accountNumberRe = regexp.MustCompile(`^01\d{6}$`)

// mustCreateAccount creates an account via the API and returns the accountNumber.
func (e *testEnv) mustCreateAccount(t *testing.T, token, name string) string {
	t.Helper()
	w := e.do(t, http.MethodPost, "/v1/accounts", token, map[string]string{
		"name":        name,
		"accountType": "personal",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("mustCreateAccount: status %d, body: %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	return body["accountNumber"].(string)
}

// --- Create account ---

func TestCreateAccount_Success(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "acc1@example.com", "password123")

	w := env.do(t, http.MethodPost, "/v1/accounts", token, map[string]string{
		"name":        "Main Account",
		"accountType": "personal",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if n, _ := body["accountNumber"].(string); !accountNumberRe.MatchString(n) {
		t.Errorf("accountNumber = %q, want match ^01\\d{6}$", n)
	}
	if body["sortCode"] != "10-10-10" {
		t.Errorf("sortCode = %v, want 10-10-10", body["sortCode"])
	}
	if body["balance"] != 0.0 {
		t.Errorf("balance = %v, want 0", body["balance"])
	}
	if body["currency"] != "GBP" {
		t.Errorf("currency = %v, want GBP", body["currency"])
	}
}

func TestCreateAccount_MissingFields_Returns400(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "acc2@example.com", "password123")

	w := env.do(t, http.MethodPost, "/v1/accounts", token, map[string]string{})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	body := decodeBody(t, w)
	if body["details"] == nil {
		t.Error("response missing details")
	}
}

func TestCreateAccount_InvalidAccountType_Returns400(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "acc3@example.com", "password123")

	w := env.do(t, http.MethodPost, "/v1/accounts", token, map[string]string{
		"name":        "Account",
		"accountType": "business",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestCreateAccount_NoAuth_Returns401(t *testing.T) {
	env := newTestEnv(t)
	w := env.do(t, http.MethodPost, "/v1/accounts", "", map[string]string{
		"name": "Account", "accountType": "personal",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

// --- List accounts ---

func TestListAccounts_OnlyOwn(t *testing.T) {
	env := newTestEnv(t)
	_, token1 := env.mustCreateUser(t, "list1@example.com", "password123")
	_, token2 := env.mustCreateUser(t, "list2@example.com", "password123")

	env.mustCreateAccount(t, token1, "Alice Account 1")
	env.mustCreateAccount(t, token1, "Alice Account 2")
	env.mustCreateAccount(t, token2, "Bob Account")

	w := env.do(t, http.MethodGet, "/v1/accounts", token1, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	accounts, _ := body["accounts"].([]any)
	if len(accounts) != 2 {
		t.Errorf("got %d accounts, want 2 (only Alice's)", len(accounts))
	}
}

func TestListAccounts_EmptyList(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "listempty@example.com", "password123")

	w := env.do(t, http.MethodGet, "/v1/accounts", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := decodeBody(t, w)
	accounts, _ := body["accounts"].([]any)
	if len(accounts) != 0 {
		t.Errorf("expected empty accounts list, got %d", len(accounts))
	}
}

// --- Fetch account ---

func TestFetchAccount_Self_Returns200(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "fetch1@example.com", "password123")
	accNum := env.mustCreateAccount(t, token, "My Account")

	w := env.do(t, http.MethodGet, "/v1/accounts/"+accNum, token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["accountNumber"] != accNum {
		t.Errorf("accountNumber = %v, want %s", body["accountNumber"], accNum)
	}
}

func TestFetchAccount_OtherUser_Returns403(t *testing.T) {
	env := newTestEnv(t)
	_, token1 := env.mustCreateUser(t, "fetch2a@example.com", "password123")
	_, token2 := env.mustCreateUser(t, "fetch2b@example.com", "password123")
	accNum := env.mustCreateAccount(t, token2, "Bob's Account")

	// Alice tries to fetch Bob's account
	w := env.do(t, http.MethodGet, "/v1/accounts/"+accNum, token1, nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestFetchAccount_NonExistent_Returns404(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "fetch3@example.com", "password123")

	w := env.do(t, http.MethodGet, "/v1/accounts/01999999", token, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestFetchAccount_InvalidPattern_Returns400(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "fetch4@example.com", "password123")

	w := env.do(t, http.MethodGet, "/v1/accounts/invalid", token, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

// --- Update account ---

func TestUpdateAccount_Self_Returns200(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "upd1@example.com", "password123")
	accNum := env.mustCreateAccount(t, token, "Original Name")

	w := env.do(t, http.MethodPatch, "/v1/accounts/"+accNum, token, map[string]string{
		"name": "Updated Name",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["name"] != "Updated Name" {
		t.Errorf("name = %v, want 'Updated Name'", body["name"])
	}
}

func TestUpdateAccount_OtherUser_Returns403(t *testing.T) {
	env := newTestEnv(t)
	_, token1 := env.mustCreateUser(t, "upd2a@example.com", "password123")
	_, token2 := env.mustCreateUser(t, "upd2b@example.com", "password123")
	accNum := env.mustCreateAccount(t, token2, "Bob's Account")

	w := env.do(t, http.MethodPatch, "/v1/accounts/"+accNum, token1, map[string]string{
		"name": "Hacked",
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestUpdateAccount_NonExistent_Returns404(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "upd3@example.com", "password123")

	w := env.do(t, http.MethodPatch, "/v1/accounts/01999998", token, map[string]string{
		"name": "Ghost",
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

// --- Delete account ---

func TestDeleteAccount_Self_Returns204(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "del1@example.com", "password123")
	accNum := env.mustCreateAccount(t, token, "To Delete")

	w := env.do(t, http.MethodDelete, "/v1/accounts/"+accNum, token, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body: %s", w.Code, w.Body.String())
	}

	// Confirm it's gone
	w2 := env.do(t, http.MethodGet, "/v1/accounts/"+accNum, token, nil)
	if w2.Code != http.StatusNotFound {
		t.Errorf("after delete: status = %d, want 404", w2.Code)
	}
}

func TestDeleteAccount_OtherUser_Returns403(t *testing.T) {
	env := newTestEnv(t)
	_, token1 := env.mustCreateUser(t, "del2a@example.com", "password123")
	_, token2 := env.mustCreateUser(t, "del2b@example.com", "password123")
	accNum := env.mustCreateAccount(t, token2, "Bob's Account")

	w := env.do(t, http.MethodDelete, "/v1/accounts/"+accNum, token1, nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestDeleteAccount_NonExistent_Returns404(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "del3@example.com", "password123")

	w := env.do(t, http.MethodDelete, "/v1/accounts/01999997", token, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

// --- Regression: delete user with account still returns 409 ---

func TestDeleteUser_WithAccount_StillReturns409(t *testing.T) {
	env := newTestEnv(t)
	id, token := env.mustCreateUser(t, "regressiondel@example.com", "password123")
	env.mustCreateAccount(t, token, "Regression Account")

	w := env.do(t, http.MethodDelete, "/v1/users/"+id, token, nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", w.Code)
	}
}
