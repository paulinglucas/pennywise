package observability

import (
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	HTTPRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Number of HTTP requests currently being served.",
	})

	DBQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_query_duration_seconds",
		Help:    "Duration of database queries in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"query"})

	DBConnectionsActive = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_connections_active",
		Help: "Number of active database connections.",
	})

	FailedRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "failed_requests_total",
		Help: "Total number of failed requests.",
	}, []string{"code"})

	ActiveUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_users",
		Help: "Number of users with activity in last 24 hours.",
	})
)

func RecordDBQuery(queryName string, duration time.Duration) {
	DBQueryDuration.WithLabelValues(queryName).Observe(duration.Seconds())
}

func RecordFailedRequest(code string) {
	FailedRequestsTotal.WithLabelValues(code).Inc()
}

type statusCapture struct {
	http.ResponseWriter
	statusCode int
}

func (sc *statusCapture) WriteHeader(code int) {
	sc.statusCode = code
	sc.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		HTTPRequestsInFlight.Inc()
		defer HTTPRequestsInFlight.Dec()

		start := time.Now()
		capture := &statusCapture{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(capture, r)

		duration := time.Since(start)
		status := strconv.Itoa(capture.statusCode)

		HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration.Seconds())
	})
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func LocalhostOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
		if host != "localhost" && host != "127.0.0.1" && host != "::1" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
