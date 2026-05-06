package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/davidcarrington/eagle-bank/internal/auth"
)

func init() {
	gin.SetMode(gin.TestMode)
}

var testSecret = []byte("test-secret")

func newTestRouter() *gin.Engine {
	r := gin.New()
	r.GET("/protected", RequireAuth(testSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"userID": CallerID(c)})
	})
	return r
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	r := newTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestRequireAuth_MalformedHeader(t *testing.T) {
	r := newTestRouter()
	cases := []string{"Basic abc", "notabearer", "Bearer"}
	for _, h := range cases {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", h)
		r.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("header %q: status = %d, want 401", h, w.Code)
		}
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	token, _, _ := auth.SignToken(testSecret, "usr-abc", -time.Second)
	r := newTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestRequireAuth_ValidToken_SetsUserID(t *testing.T) {
	token, _, _ := auth.SignToken(testSecret, "usr-xyz789", time.Hour)
	r := newTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "usr-xyz789") {
		t.Errorf("body %q does not contain expected userID", w.Body.String())
	}
}
