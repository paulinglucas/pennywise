package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jamespsullivan/pennywise/internal/middleware"
)

func (h *AppHandler) PostVitals(w http.ResponseWriter, r *http.Request) {
	var req VitalsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		requestID := middleware.GetRequestID(r.Context())
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	for _, metric := range req.Metrics {
		slog.Info("web vital",
			slog.String("name", metric.Name),
			slog.Float64("value", float64(metric.Value)),
		)
	}

	w.WriteHeader(http.StatusNoContent)
}
