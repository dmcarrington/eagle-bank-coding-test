package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/davidcarrington/eagle-bank/internal/api"
	"github.com/davidcarrington/eagle-bank/internal/config"
	"github.com/davidcarrington/eagle-bank/internal/store"
)

func init() {
	gin.SetMode(gin.TestMode)
}

var testSecret = []byte("handler-test-secret")

type testEnv struct {
	router *gin.Engine
	db     *store.Store
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	s, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { s.DB.Close() })

	cfg := config.Config{
		JWTSecret: testSecret,
		JWTTTL:    time.Hour,
	}
	r := api.NewRouter(api.Deps{Store: s, Config: cfg})
	return &testEnv{router: r, db: s}
}

func (e *testEnv) do(t *testing.T, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, req)
	return w
}

func (e *testEnv) mustCreateUser(t *testing.T, email, password string) (userID, token string) {
	t.Helper()
	w := e.do(t, http.MethodPost, "/v1/users", "", map[string]any{
		"name":        "Test User",
		"email":       email,
		"password":    password,
		"phoneNumber": "+447700900000",
		"address": map[string]string{
			"line1":    "1 High Street",
			"town":     "London",
			"county":   "Greater London",
			"postcode": "SW1A 1AA",
		},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("mustCreateUser: status %d, body: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	userID = resp["id"].(string)

	lw := e.do(t, http.MethodPost, "/v1/auth/login", "", map[string]string{
		"email":    email,
		"password": password,
	})
	if lw.Code != http.StatusOK {
		t.Fatalf("mustCreateUser login: status %d, body: %s", lw.Code, lw.Body.String())
	}
	var lr map[string]any
	if err := json.NewDecoder(lw.Body).Decode(&lr); err != nil {
		t.Fatal(err)
	}
	token = lr["token"].(string)
	return
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.NewDecoder(w.Body).Decode(&m); err != nil {
		t.Fatalf("decodeBody: %v (body: %s)", err, w.Body.String())
	}
	return m
}
