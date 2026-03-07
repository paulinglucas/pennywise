package api

import (
	"context"
	"net/http"

	"github.com/jamespsullivan/pennywise/internal/middleware"
	"github.com/jamespsullivan/pennywise/internal/models"
)

type AlertRepository interface {
	List(ctx context.Context, userID string, page, perPage int) ([]models.Alert, int, error)
	MarkRead(ctx context.Context, userID, id string) (bool, error)
}

func (h *AppHandler) ListAlerts(w http.ResponseWriter, r *http.Request, params ListAlertsParams) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	page, perPage := paginationDefaults(params.Page, params.PerPage)

	items, total, err := h.alerts.List(r.Context(), userID, page, perPage)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to list alerts", requestID)
		return
	}

	data := make([]AlertResponse, len(items))
	for i, a := range items {
		data[i] = alertToResponse(a)
	}

	WriteJSON(w, http.StatusOK, AlertListResponse{
		Data:       data,
		Pagination: paginationMeta(page, perPage, total),
	})
}

func (h *AppHandler) MarkAlertRead(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	found, err := h.alerts.MarkRead(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to mark alert as read", requestID)
		return
	}
	if !found {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Alert not found", requestID)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func alertToResponse(a models.Alert) AlertResponse {
	resp := AlertResponse{
		Id:        ParseID(a.ID),
		UserId:    ParseID(a.UserID),
		AlertType: a.AlertType,
		Message:   a.Message,
		Severity:  AlertSeverity(a.Severity),
		IsRead:    a.IsRead,
		CreatedAt: a.CreatedAt,
	}

	if a.RelatedEntityType != nil {
		resp.RelatedEntityType = a.RelatedEntityType
	}

	if a.RelatedEntityID != nil {
		id := ParseID(*a.RelatedEntityID)
		resp.RelatedEntityId = &id
	}

	return resp
}
