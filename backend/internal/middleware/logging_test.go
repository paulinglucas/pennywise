package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/middleware"
)

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewJSONHandler(buf, nil))
}

func TestLogging_LogsRequestFields(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := middleware.RequestID(
		middleware.Logging(logger)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
		),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var logEntry map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, "request completed", logEntry["msg"])
	assert.Equal(t, "GET", logEntry["method"])
	assert.Equal(t, "/api/v1/accounts", logEntry["path"])
	assert.Equal(t, float64(200), logEntry["status"])
	assert.NotEmpty(t, logEntry["request_id"])
	assert.Contains(t, logEntry, "duration_ms")
}

func TestLogging_LogsStatusCode(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := middleware.Logging(logger)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var logEntry map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, float64(404), logEntry["status"])
}

func TestLogging_ServerErrorLogsAsError(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := middleware.Logging(logger)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var logEntry map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, "ERROR", logEntry["level"])
}

func TestLogging_IncludesUserID(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	inner := middleware.Logging(logger)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := middleware.WithUserID(req.Context(), "usr-123")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	inner.ServeHTTP(rec, req)

	var logEntry map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	assert.Equal(t, "usr-123", logEntry["user_id"])
}

func TestLogging_OmitsUserIDWhenMissing(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := middleware.Logging(logger)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	var logEntry map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &logEntry))

	_, hasUserID := logEntry["user_id"]
	assert.False(t, hasUserID)
}
