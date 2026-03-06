package middleware_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/api"
	"github.com/jamespsullivan/pennywise/internal/middleware"
)

func setupValidationHandler(t *testing.T) http.Handler {
	t.Helper()
	validator, err := middleware.Validation(api.OpenAPISpec, "/api/v1")
	require.NoError(t, err)

	handler := validator(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	return handler
}

func TestValidation_InvalidBody_Returns400(t *testing.T) {
	handler := setupValidationHandler(t)

	body := `{"invalid": true}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp map[string]map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_FAILED", resp["error"]["code"])
}

func TestValidation_ValidBody_PassesThrough(t *testing.T) {
	handler := setupValidationHandler(t)

	body := `{"name":"Test","institution":"Bank","account_type":"checking"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestValidation_UnknownRoute_PassesThrough(t *testing.T) {
	handler := setupValidationHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/unknown/route", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestValidation_ErrorResponseFormat(t *testing.T) {
	handler := setupValidationHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var resp map[string]map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "code")
	assert.Contains(t, resp["error"], "message")
	assert.Contains(t, resp["error"], "request_id")
}
