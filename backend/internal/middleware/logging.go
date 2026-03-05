package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseCapture struct {
	http.ResponseWriter
	statusCode int
}

func (rc *responseCapture) WriteHeader(code int) {
	rc.statusCode = code
	rc.ResponseWriter.WriteHeader(code)
}

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			capture := &responseCapture{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(capture, r)

			duration := time.Since(start)
			requestID := GetRequestID(r.Context())
			userID := GetUserID(r.Context())

			attrs := []slog.Attr{
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", capture.statusCode),
				slog.Int64("duration_ms", duration.Milliseconds()),
			}

			if userID != "" {
				attrs = append(attrs, slog.String("user_id", userID))
			}

			args := make([]any, len(attrs))
			for i, attr := range attrs {
				args[i] = attr
			}

			if capture.statusCode >= 500 {
				logger.Error("request completed", args...)
			} else {
				logger.Info("request completed", args...)
			}
		})
	}
}
