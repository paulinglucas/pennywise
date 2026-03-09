package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/jamespsullivan/pennywise/internal/middleware"
	"github.com/jamespsullivan/pennywise/internal/models"
)

func (h *AppHandler) ListGoalContributions(w http.ResponseWriter, r *http.Request, id IdParam, params ListGoalContributionsParams) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	goal, err := h.goals.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get goal", requestID)
		return
	}
	if goal == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Goal not found", requestID)
		return
	}

	page, perPage := paginationDefaults(params.Page, params.PerPage)

	contribs, total, err := h.goalContributions.ListByGoal(r.Context(), userID, id.String(), page, perPage)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to list contributions", requestID)
		return
	}

	data := make([]GoalContributionResponse, len(contribs))
	for i, c := range contribs {
		data[i] = contributionToResponse(c)
	}

	WriteJSON(w, http.StatusOK, GoalContributionListResponse{
		Data:       data,
		Pagination: paginationMeta(page, perPage, total),
	})
}

func (h *AppHandler) CreateGoalContribution(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	goal, err := h.goals.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get goal", requestID)
		return
	}
	if goal == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Goal not found", requestID)
		return
	}

	var req CreateGoalContributionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	contributedAt := time.Now()
	if req.ContributedAt != nil {
		contributedAt = req.ContributedAt.Time
	}

	var transactionID *string
	if req.TransactionId != nil {
		txnIDStr := req.TransactionId.String()
		txn, txnErr := h.transactions.GetByID(r.Context(), userID, txnIDStr)
		if txnErr != nil {
			WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to validate transaction", requestID)
			return
		}
		if txn == nil {
			WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Transaction not found", requestID)
			return
		}
		transactionID = &txnIDStr
	}

	contrib := &models.GoalContribution{
		ID:            uuid.New().String(),
		GoalID:        id.String(),
		UserID:        userID,
		Amount:        float64(req.Amount),
		Notes:         req.Notes,
		TransactionID: transactionID,
		ContributedAt: contributedAt,
	}

	if err := h.goalContributions.Create(r.Context(), contrib); err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to create contribution", requestID)
		return
	}

	prevData := goalJSON(goal)
	goal.CurrentAmount += contrib.Amount
	if _, err := h.goals.Update(r.Context(), goal); err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to update goal", requestID)
		return
	}

	updated, err := h.goals.GetByID(r.Context(), userID, id.String())
	if err == nil && updated != nil {
		newData := goalJSON(updated)
		h.recordAudit(r.Context(), userID, "goal", id.String(), "contribute", &prevData, &newData)
	}

	WriteJSON(w, http.StatusCreated, contributionToResponse(*contrib))
}

func (h *AppHandler) DeleteGoalContribution(w http.ResponseWriter, r *http.Request, id IdParam, contributionId openapi_types.UUID) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	goal, err := h.goals.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get goal", requestID)
		return
	}
	if goal == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Goal not found", requestID)
		return
	}

	contrib, err := h.goalContributions.GetByID(r.Context(), userID, contributionId.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get contribution", requestID)
		return
	}
	if contrib == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Contribution not found", requestID)
		return
	}

	found, err := h.goalContributions.Delete(r.Context(), userID, contributionId.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to delete contribution", requestID)
		return
	}
	if !found {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Contribution not found", requestID)
		return
	}

	prevData := goalJSON(goal)
	goal.CurrentAmount -= contrib.Amount
	if goal.CurrentAmount < 0 {
		goal.CurrentAmount = 0
	}
	if _, err := h.goals.Update(r.Context(), goal); err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to update goal", requestID)
		return
	}

	updated, err := h.goals.GetByID(r.Context(), userID, id.String())
	if err == nil && updated != nil {
		newData := goalJSON(updated)
		h.recordAudit(r.Context(), userID, "goal", id.String(), "reverse_contribution", &prevData, &newData)
	}

	w.WriteHeader(http.StatusNoContent)
}

func contributionToResponse(c models.GoalContribution) GoalContributionResponse {
	resp := GoalContributionResponse{
		Id:            ParseID(c.ID),
		GoalId:        ParseID(c.GoalID),
		UserId:        ParseID(c.UserID),
		Amount:        float32(c.Amount),
		Notes:         c.Notes,
		ContributedAt: openapi_types.Date{Time: c.ContributedAt},
		CreatedAt:     c.CreatedAt,
	}
	if c.TransactionID != nil {
		txnID := ParseID(*c.TransactionID)
		resp.TransactionId = &txnID
	}
	return resp
}
