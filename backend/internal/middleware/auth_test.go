package middleware_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/middleware"
)

var testSecret = []byte("test-secret-key-at-least-32-bytes-long")

type scopesKeyType string

const testScopesKey scopesKeyType = "test.Scopes"

func signTestToken(claims *middleware.Claims) string {
	token, err := middleware.SignToken(claims, testSecret)
	if err != nil {
		panic(err)
	}
	return token
}

func withScopes(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), testScopesKey, []string{})
	return r.WithContext(ctx)
}

func TestAuth_ValidToken_SetsUserID(t *testing.T) {
	t.Parallel()
	claims := middleware.NewClaims("usr-123", "test@example.com", time.Hour)
	tokenStr := signTestToken(claims)

	var capturedUserID string
	handler := middleware.Auth(testSecret, testScopesKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = middleware.GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := withScopes(httptest.NewRequest(http.MethodGet, "/protected", nil))
	req.AddCookie(&http.Cookie{Name: "token", Value: tokenStr})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "usr-123", capturedUserID)
}

func TestAuth_MissingCookie_Returns401(t *testing.T) {
	t.Parallel()
	handler := middleware.Auth(testSecret, testScopesKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := withScopes(httptest.NewRequest(http.MethodGet, "/protected", nil))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var resp map[string]map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "UNAUTHORIZED", resp["error"]["code"])
	assert.Equal(t, "Missing authentication token", resp["error"]["message"])
}

func TestAuth_ExpiredToken_Returns401(t *testing.T) {
	t.Parallel()
	claims := &middleware.Claims{
		UserID: "usr-123",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			Issuer:    "pennywise",
		},
	}
	tokenStr := signTestToken(claims)

	handler := middleware.Auth(testSecret, testScopesKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := withScopes(httptest.NewRequest(http.MethodGet, "/protected", nil))
	req.AddCookie(&http.Cookie{Name: "token", Value: tokenStr})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_MalformedToken_Returns401(t *testing.T) {
	t.Parallel()
	handler := middleware.Auth(testSecret, testScopesKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := withScopes(httptest.NewRequest(http.MethodGet, "/protected", nil))
	req.AddCookie(&http.Cookie{Name: "token", Value: "not-a-valid-jwt"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_WrongSecret_Returns401(t *testing.T) {
	t.Parallel()
	claims := middleware.NewClaims("usr-123", "test@example.com", time.Hour)
	tokenStr := signTestToken(claims)

	wrongSecret := []byte("different-secret-entirely")
	handler := middleware.Auth(wrongSecret, testScopesKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := withScopes(httptest.NewRequest(http.MethodGet, "/protected", nil))
	req.AddCookie(&http.Cookie{Name: "token", Value: tokenStr})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_NoScopes_PassesThrough(t *testing.T) {
	t.Parallel()
	handler := middleware.Auth(testSecret, testScopesKey)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestSignToken_ReturnsValidToken(t *testing.T) {
	t.Parallel()
	claims := middleware.NewClaims("usr-456", "user@example.com", time.Hour)

	tokenStr, err := middleware.SignToken(claims, testSecret)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenStr)
}

func TestNewClaims_SetsFields(t *testing.T) {
	t.Parallel()
	claims := middleware.NewClaims("usr-789", "user@example.com", 24*time.Hour)

	assert.Equal(t, "usr-789", claims.UserID)
	assert.Equal(t, "user@example.com", claims.Email)
	assert.Equal(t, "pennywise", claims.Issuer)
	assert.False(t, claims.ExpiresAt.IsZero())
	assert.True(t, claims.ExpiresAt.After(time.Now()))
}

func TestGetUserID_EmptyWithoutMiddleware(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	id := middleware.GetUserID(req.Context())
	assert.Empty(t, id)
}

func TestJWTAuth_ValidToken_SetsUserID(t *testing.T) {
	t.Parallel()
	claims := middleware.NewClaims("usr-jwt-1", "jwt@example.com", time.Hour)
	tokenStr := signTestToken(claims)

	var capturedUserID string
	handler := middleware.JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID = middleware.GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: tokenStr})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "usr-jwt-1", capturedUserID)
}

func TestJWTAuth_MissingCookie_Returns401(t *testing.T) {
	t.Parallel()
	handler := middleware.JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuth_ExpiredToken_Returns401(t *testing.T) {
	t.Parallel()
	claims := &middleware.Claims{
		UserID: "usr-jwt-2",
		Email:  "jwt@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			Issuer:    "pennywise",
		},
	}
	tokenStr := signTestToken(claims)

	handler := middleware.JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: tokenStr})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTAuth_MalformedToken_Returns401(t *testing.T) {
	t.Parallel()
	handler := middleware.JWTAuth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "token", Value: "garbage-token"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestWithUserID_And_GetUserID(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := middleware.WithUserID(req.Context(), "usr-test")

	assert.Equal(t, "usr-test", middleware.GetUserID(ctx))
}
