package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/jamespsullivan/pennywise/internal/middleware"
	"github.com/jamespsullivan/pennywise/internal/models"
)

type RecurringRepository interface {
	List(ctx context.Context, userID string, page, perPage int) ([]models.RecurringTransaction, int, error)
	Create(ctx context.Context, rec *models.RecurringTransaction) error
	GetByID(ctx context.Context, userID, id string) (*models.RecurringTransaction, error)
	Update(ctx context.Context, rec *models.RecurringTransaction) (bool, error)
	SoftDelete(ctx context.Context, userID, id string) (bool, error)
}

func (h *AppHandler) ListRecurring(w http.ResponseWriter, r *http.Request, params ListRecurringParams) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	page, perPage := paginationDefaults(params.Page, params.PerPage)

	items, total, err := h.recurring.List(r.Context(), userID, page, perPage)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to list recurring transactions", requestID)
		return
	}

	data := make([]RecurringResponse, len(items))
	for i, rt := range items {
		data[i] = recurringToResponse(rt)
	}

	WriteJSON(w, http.StatusOK, RecurringListResponse{
		Data:       data,
		Pagination: paginationMeta(page, perPage, total),
	})
}

func (h *AppHandler) CreateRecurring(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	var req CreateRecurringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	rec := newRecurringModel(userID, req)

	if err := h.recurring.Create(r.Context(), rec); err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to create recurring transaction", requestID)
		return
	}

	created, err := h.recurring.GetByID(r.Context(), userID, rec.ID)
	if err != nil || created == nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to retrieve recurring transaction", requestID)
		return
	}

	WriteJSON(w, http.StatusCreated, recurringToResponse(*created))
}

func (h *AppHandler) UpdateRecurring(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	rec, err := h.recurring.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get recurring transaction", requestID)
		return
	}
	if rec == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Recurring transaction not found", requestID)
		return
	}

	var req UpdateRecurringRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	applyRecurringUpdates(rec, req)

	if _, err := h.recurring.Update(r.Context(), rec); err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to update recurring transaction", requestID)
		return
	}

	updated, err := h.recurring.GetByID(r.Context(), userID, id.String())
	if err != nil || updated == nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to retrieve recurring transaction", requestID)
		return
	}

	WriteJSON(w, http.StatusOK, recurringToResponse(*updated))
}

func (h *AppHandler) DeleteRecurring(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	found, err := h.recurring.SoftDelete(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to delete recurring transaction", requestID)
		return
	}
	if !found {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Recurring transaction not found", requestID)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func newRecurringModel(userID string, req CreateRecurringRequest) *models.RecurringTransaction {
	currency := USD
	if req.Currency != nil {
		currency = *req.Currency
	}

	return &models.RecurringTransaction{
		ID:             uuid.New().String(),
		UserID:         userID,
		AccountID:      req.AccountId.String(),
		Type:           string(req.Type),
		Category:       req.Category,
		Amount:         float64(req.Amount),
		Currency:       string(currency),
		Frequency:      string(req.Frequency),
		NextOccurrence: req.NextOccurrence.Time,
		IsActive:       true,
	}
}

func recurringToResponse(rt models.RecurringTransaction) RecurringResponse {
	return RecurringResponse{
		Id:             ParseID(rt.ID),
		UserId:         ParseID(rt.UserID),
		AccountId:      ParseID(rt.AccountID),
		Type:           TransactionType(rt.Type),
		Category:       rt.Category,
		Amount:         float32(rt.Amount),
		Currency:       Currency(rt.Currency),
		Frequency:      Frequency(rt.Frequency),
		NextOccurrence: openapi_types.Date{Time: rt.NextOccurrence},
		IsActive:       rt.IsActive,
		CreatedAt:      rt.CreatedAt,
		UpdatedAt:      rt.UpdatedAt,
	}
}

func applyRecurringUpdates(rec *models.RecurringTransaction, req UpdateRecurringRequest) {
	if req.AccountId != nil {
		rec.AccountID = req.AccountId.String()
	}
	if req.Type != nil {
		rec.Type = string(*req.Type)
	}
	if req.Category != nil {
		rec.Category = *req.Category
	}
	if req.Amount != nil {
		rec.Amount = float64(*req.Amount)
	}
	if req.Currency != nil {
		rec.Currency = string(*req.Currency)
	}
	if req.Frequency != nil {
		rec.Frequency = string(*req.Frequency)
	}
	if req.NextOccurrence != nil {
		rec.NextOccurrence = req.NextOccurrence.Time
	}
	if req.IsActive != nil {
		rec.IsActive = *req.IsActive
	}
}
