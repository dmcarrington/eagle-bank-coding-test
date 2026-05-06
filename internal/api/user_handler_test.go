package api_test

import (
	"net/http"
	"testing"
)

// --- Create user ---

func TestCreateUser_Success(t *testing.T) {
	env := newTestEnv(t)
	w := env.do(t, http.MethodPost, "/v1/users", "", map[string]any{
		"name":        "Alice Smith",
		"email":       "alice@example.com",
		"password":    "password123",
		"phoneNumber": "+447700900001",
		"address": map[string]string{
			"line1":    "1 High Street",
			"town":     "London",
			"county":   "Greater London",
			"postcode": "SW1A 1AA",
		},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["id"] == nil || body["id"] == "" {
		t.Error("response missing id field")
	}
	if body["email"] != "alice@example.com" {
		t.Errorf("email = %v, want alice@example.com", body["email"])
	}
	if body["password"] != nil {
		t.Error("password must not be returned in response")
	}
	if body["createdTimestamp"] == nil {
		t.Error("response missing createdTimestamp")
	}
}

func TestCreateUser_MissingFields_Returns400WithDetails(t *testing.T) {
	env := newTestEnv(t)
	// Send only a name — omit email, password, phoneNumber, address
	w := env.do(t, http.MethodPost, "/v1/users", "", map[string]any{
		"name": "Bob",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	body := decodeBody(t, w)
	if body["details"] == nil {
		t.Fatal("response missing details field")
	}
	details, ok := body["details"].([]any)
	if !ok || len(details) == 0 {
		t.Error("details should be a non-empty array")
	}
}

func TestCreateUser_InvalidEmail_Returns400(t *testing.T) {
	env := newTestEnv(t)
	w := env.do(t, http.MethodPost, "/v1/users", "", map[string]any{
		"name":        "Bob",
		"email":       "not-an-email",
		"password":    "password123",
		"phoneNumber": "+447700900002",
		"address": map[string]string{
			"line1": "1 St", "town": "London", "county": "GL", "postcode": "SW1",
		},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
	body := decodeBody(t, w)
	details, _ := body["details"].([]any)
	found := false
	for _, d := range details {
		dm := d.(map[string]any)
		if dm["field"] == "email" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected validation detail for email field; got %v", body["details"])
	}
}

func TestCreateUser_ShortPassword_Returns400(t *testing.T) {
	env := newTestEnv(t)
	w := env.do(t, http.MethodPost, "/v1/users", "", map[string]any{
		"name":        "Bob",
		"email":       "bob@example.com",
		"password":    "short",
		"phoneNumber": "+447700900002",
		"address": map[string]string{
			"line1": "1 St", "town": "London", "county": "GL", "postcode": "SW1",
		},
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "login@example.com", "password123")
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestLogin_BadPassword_Returns401(t *testing.T) {
	env := newTestEnv(t)
	env.mustCreateUser(t, "user@example.com", "correctpassword")

	w := env.do(t, http.MethodPost, "/v1/auth/login", "", map[string]string{
		"email":    "user@example.com",
		"password": "wrongpassword",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestLogin_UnknownEmail_Returns401(t *testing.T) {
	env := newTestEnv(t)
	w := env.do(t, http.MethodPost, "/v1/auth/login", "", map[string]string{
		"email":    "nobody@example.com",
		"password": "password123",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestLogin_MissingFields_Returns400(t *testing.T) {
	env := newTestEnv(t)
	w := env.do(t, http.MethodPost, "/v1/auth/login", "", map[string]string{
		"email": "user@example.com",
	})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

// --- Fetch user ---

func TestFetchUser_Self_Returns200(t *testing.T) {
	env := newTestEnv(t)
	id, token := env.mustCreateUser(t, "fetch@example.com", "password123")

	w := env.do(t, http.MethodGet, "/v1/users/"+id, token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["id"] != id {
		t.Errorf("id = %v, want %s", body["id"], id)
	}
}

func TestFetchUser_OtherUser_Returns403(t *testing.T) {
	env := newTestEnv(t)
	id1, token1 := env.mustCreateUser(t, "alice2@example.com", "password123")
	_ = id1
	id2, _ := env.mustCreateUser(t, "bob2@example.com", "password123")

	// Alice tries to fetch Bob's record
	w := env.do(t, http.MethodGet, "/v1/users/"+id2, token1, nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestFetchUser_NonExistent_Returns404(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "alice3@example.com", "password123")

	w := env.do(t, http.MethodGet, "/v1/users/usr-doesnotexist", token, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestFetchUser_NoAuth_Returns401(t *testing.T) {
	env := newTestEnv(t)
	id, _ := env.mustCreateUser(t, "noauth@example.com", "password123")

	w := env.do(t, http.MethodGet, "/v1/users/"+id, "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

// --- Update user ---

func TestUpdateUser_Self_Returns200(t *testing.T) {
	env := newTestEnv(t)
	id, token := env.mustCreateUser(t, "update@example.com", "password123")

	w := env.do(t, http.MethodPatch, "/v1/users/"+id, token, map[string]any{
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

func TestUpdateUser_OtherUser_Returns403(t *testing.T) {
	env := newTestEnv(t)
	_, token1 := env.mustCreateUser(t, "alice4@example.com", "password123")
	id2, _ := env.mustCreateUser(t, "bob4@example.com", "password123")

	w := env.do(t, http.MethodPatch, "/v1/users/"+id2, token1, map[string]any{
		"name": "Hacked",
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestUpdateUser_NonExistent_Returns404(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "alice5@example.com", "password123")

	w := env.do(t, http.MethodPatch, "/v1/users/usr-ghost", token, map[string]any{
		"name": "Ghost",
	})
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

// --- Delete user ---

func TestDeleteUser_Self_NoAccounts_Returns204(t *testing.T) {
	env := newTestEnv(t)
	id, token := env.mustCreateUser(t, "delete@example.com", "password123")

	w := env.do(t, http.MethodDelete, "/v1/users/"+id, token, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204; body: %s", w.Code, w.Body.String())
	}
}

func TestDeleteUser_Self_WithAccount_Returns409(t *testing.T) {
	env := newTestEnv(t)
	id, token := env.mustCreateUser(t, "deleteaccount@example.com", "password123")

	// Seed an account directly so we don't need account routes yet
	_, err := env.db.DB.Exec(`
		INSERT INTO accounts (account_number, user_id, name, account_type, balance_pence, currency, sort_code, created_timestamp, updated_timestamp)
		VALUES ('01000001', ?, 'Test', 'personal', 0, 'GBP', '10-10-10', '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z')`,
		id)
	if err != nil {
		t.Fatalf("seed account: %v", err)
	}

	w := env.do(t, http.MethodDelete, "/v1/users/"+id, token, nil)
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body: %s", w.Code, w.Body.String())
	}
}

func TestDeleteUser_OtherUser_Returns403(t *testing.T) {
	env := newTestEnv(t)
	_, token1 := env.mustCreateUser(t, "alice6@example.com", "password123")
	id2, _ := env.mustCreateUser(t, "bob6@example.com", "password123")

	w := env.do(t, http.MethodDelete, "/v1/users/"+id2, token1, nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", w.Code)
	}
}

func TestDeleteUser_NonExistent_Returns404(t *testing.T) {
	env := newTestEnv(t)
	_, token := env.mustCreateUser(t, "alice7@example.com", "password123")

	w := env.do(t, http.MethodDelete, "/v1/users/usr-ghost404", token, nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}
