package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "unit-test-secret"

func sign(t *testing.T, method jwt.SigningMethod, key any, claims jwt.MapClaims) string {
	t.Helper()
	s, err := jwt.NewWithClaims(method, claims).SignedString(key)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return s
}

func runAuth(authHeader string) (*httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, e := gin.CreateTestContext(w)
	var captured *gin.Context
	e.Use(AuthRequired(testSecret))
	e.GET("/x", func(ctx *gin.Context) {
		captured = ctx
		ctx.Status(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	e.ServeHTTP(w, req)
	_ = c
	return w, captured
}

func TestAuthRequired_MissingHeader(t *testing.T) {
	w, _ := runAuth("")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestAuthRequired_ValidToken_SetsIdentity(t *testing.T) {
	tok := sign(t, jwt.SigningMethodHS256, []byte(testSecret), jwt.MapClaims{
		"user_id": float64(42),
		"role_id": float64(7),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	w, c := runAuth("Bearer " + tok)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if c == nil || c.GetUint("userID") != 42 || c.GetUint("roleID") != 7 {
		t.Fatalf("identity not set: userID=%v roleID=%v", c.GetUint("userID"), c.GetUint("roleID"))
	}
}

func TestAuthRequired_WrongSecret(t *testing.T) {
	tok := sign(t, jwt.SigningMethodHS256, []byte("attacker-secret"), jwt.MapClaims{
		"user_id": float64(1), "exp": time.Now().Add(time.Hour).Unix(),
	})
	w, _ := runAuth("Bearer " + tok)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("forged-secret token must be rejected, got %d", w.Code)
	}
}

func TestAuthRequired_ExpiredToken(t *testing.T) {
	tok := sign(t, jwt.SigningMethodHS256, []byte(testSecret), jwt.MapClaims{
		"user_id": float64(1), "exp": time.Now().Add(-time.Hour).Unix(),
	})
	w, _ := runAuth("Bearer " + tok)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expired token must be rejected, got %d", w.Code)
	}
}

func TestAuthRequired_MissingUserIDClaim(t *testing.T) {
	tok := sign(t, jwt.SigningMethodHS256, []byte(testSecret), jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	w, _ := runAuth("Bearer " + tok)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("token without user_id must be rejected, got %d", w.Code)
	}
}
