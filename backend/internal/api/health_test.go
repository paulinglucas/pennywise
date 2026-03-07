package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/api"
)

func TestGetHealth_Healthy(t *testing.T) {
	_, router := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.HealthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, api.Healthy, resp.Status)
	assert.Equal(t, api.Connected, resp.Database)
	assert.Equal(t, "1.0.0", resp.Version)
	assert.GreaterOrEqual(t, resp.UptimeSeconds, 0)
	assert.False(t, resp.Timestamp.IsZero())
}

func TestGetHealth_NoAuthRequired(t *testing.T) {
	_, router := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
