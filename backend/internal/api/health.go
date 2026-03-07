package api

import (
	"net/http"
	"time"
)

const appVersion = "1.0.0"

func (h *AppHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	dbStatus := Connected
	status := Healthy
	httpStatus := http.StatusOK

	if err := h.dashboard.PingDB(r.Context()); err != nil {
		dbStatus = Disconnected
		status = Unhealthy
		httpStatus = http.StatusServiceUnavailable
	}

	uptime := int(time.Since(h.startTime).Seconds())

	WriteJSON(w, httpStatus, HealthResponse{
		Status:        status,
		Version:       appVersion,
		UptimeSeconds: uptime,
		Database:      dbStatus,
		Timestamp:     time.Now(),
	})
}
