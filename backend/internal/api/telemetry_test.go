package api_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostVitals_ValidRequest(t *testing.T) {
	_, router := setupRouter(t)

	body := `{"metrics":[{"name":"LCP","value":2500.5},{"name":"FID","value":100},{"name":"CLS","value":0.1}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/vitals", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestPostVitals_NoAuthRequired(t *testing.T) {
	_, router := setupRouter(t)

	body := `{"metrics":[{"name":"LCP","value":1200}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/vitals", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestPostVitals_WithMetricID(t *testing.T) {
	_, router := setupRouter(t)

	body := `{"metrics":[{"id":"v1-abc","name":"LCP","value":1500}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/vitals", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestPostVitals_EmptyMetrics(t *testing.T) {
	_, router := setupRouter(t)

	body := `{"metrics":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/vitals", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}
