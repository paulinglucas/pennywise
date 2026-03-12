package observability_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jamespsullivan/pennywise/internal/observability"
)

func TestMetricsMiddleware_IncrementsCounter(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := observability.MetricsMiddleware(inner)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMetricsMiddleware_Records500(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	handler := observability.MetricsMiddleware(inner)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/error", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestMetricsHandler_ReturnsPrometheusOutput(t *testing.T) {
	handler := observability.MetricsHandler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assert.Contains(t, body, "http_requests_in_flight")
	assert.Contains(t, body, "db_connections_active")
	assert.Contains(t, body, "active_users")
}

func TestLocalhostOnly_AllowsLocalhost(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := observability.LocalhostOnly(inner)

	tests := []struct {
		host string
	}{
		{"localhost:9090"},
		{"127.0.0.1:8080"},
		{"::1"},
		{"localhost"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.Host = tt.host
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "host %s should be allowed", tt.host)
	}
}

func TestLocalhostOnly_BlocksRemote(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := observability.LocalhostOnly(inner)

	tests := []struct {
		host string
	}{
		{"192.168.1.100:8080"},
		{"example.com"},
		{"10.0.0.1"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.Host = tt.host
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code, "host %s should be blocked", tt.host)
	}
}

func TestRecordDBQuery_DoesNotPanic(t *testing.T) {
	observability.RecordDBQuery("test_query", 100*time.Millisecond)
}

func TestRecordFailedRequest_DoesNotPanic(t *testing.T) {
	observability.RecordFailedRequest("VALIDATION_FAILED")
}

func TestMetricsMiddleware_DefaultStatusIs200(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	handler := observability.MetricsMiddleware(inner)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/default-status", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestMetricsMiddleware_StatusCodes(t *testing.T) {
	codes := []int{
		http.StatusCreated,
		http.StatusNoContent,
		http.StatusBadRequest,
		http.StatusNotFound,
		http.StatusConflict,
		http.StatusServiceUnavailable,
	}

	for _, code := range codes {
		code := code
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(code)
		})
		handler := observability.MetricsMiddleware(inner)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/status-test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, code, rec.Code)
	}
}

func TestMetricsMiddleware_DifferentMethods(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPost,
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
	}

	for _, method := range methods {
		inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := observability.MetricsMiddleware(inner)

		req := httptest.NewRequest(method, "/api/v1/method-test", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestMetricsHandler_ContainsCustomMetrics(t *testing.T) {
	observability.RecordDBQuery("test_get_accounts", 50*time.Millisecond)
	observability.RecordFailedRequest("TEST_ERROR")

	handler := observability.MetricsHandler()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	assert.True(t, strings.Contains(body, "db_query_duration_seconds"))
	assert.True(t, strings.Contains(body, "failed_requests_total"))
}

func TestInitTracer_ReturnsShutdownFunc(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&strings.Builder{}, nil))
	var traceOutput strings.Builder

	shutdown, err := observability.InitTracer(logger, &traceOutput)
	assert.NoError(t, err)
	assert.NotNil(t, shutdown)

	ctx := context.Background()
	err = shutdown(ctx)
	assert.NoError(t, err)
}

func TestStartSpan_ReturnsSpan(t *testing.T) {
	ctx := context.Background()
	spanCtx, span := observability.StartSpan(ctx, "test-span")
	assert.NotNil(t, spanCtx)
	assert.NotNil(t, span)
	span.End()
}

func TestTracer_ReturnsTracer(t *testing.T) {
	tracer := observability.Tracer()
	assert.NotNil(t, tracer)
}
