package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/jamespsullivan/pennywise/internal/middleware"
	"github.com/jamespsullivan/pennywise/internal/models"
)

type TransactionGroupRepository interface {
	Create(ctx context.Context, group *models.TransactionGroup) error
	GetByID(ctx context.Context, userID, id string) (*models.TransactionGroup, error)
	Update(ctx context.Context, group *models.TransactionGroup) (bool, error)
	SoftDelete(ctx context.Context, userID, id string) (bool, error)
	List(ctx context.Context, userID string, page, perPage int) ([]models.TransactionGroup, int, error)
	ListMembers(ctx context.Context, userID, groupID string) ([]models.Transaction, error)
}

func (h *AppHandler) ListTransactionGroups(w http.ResponseWriter, r *http.Request, params ListTransactionGroupsParams) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	page, perPage := paginationDefaults(params.Page, params.PerPage)

	groups, total, err := h.transactionGroups.List(r.Context(), userID, page, perPage)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to list transaction groups", requestID)
		return
	}

	data := make([]TransactionGroupResponse, len(groups))
	for i, g := range groups {
		members, err := h.transactionGroups.ListMembers(r.Context(), userID, g.ID)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to load group members", requestID)
			return
		}
		data[i] = groupToResponse(g, members)
	}

	WriteJSON(w, http.StatusOK, TransactionGroupListResponse{
		Data:       data,
		Pagination: paginationMeta(page, perPage, total),
	})
}

func (h *AppHandler) CreateTransactionGroup(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	var req CreateTransactionGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	if len(req.Members) < 2 {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "A group requires at least 2 members", requestID)
		return
	}

	group := &models.TransactionGroup{
		ID:     uuid.New().String(),
		UserID: userID,
		Name:   req.Name,
	}

	if err := h.transactionGroups.Create(r.Context(), group); err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to create transaction group", requestID)
		return
	}

	memberTxns := make([]models.Transaction, len(req.Members))
	for i, m := range req.Members {
		txn := memberInputToTransaction(userID, group.ID, m)
		if err := h.transactions.Create(r.Context(), &txn, tagsFromPtr(m.Tags)); err != nil {
			WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to create group member", requestID)
			return
		}
		created, err := h.transactions.GetByID(r.Context(), userID, txn.ID)
		if err != nil || created == nil {
			WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to retrieve group member", requestID)
			return
		}
		memberTxns[i] = *created
	}

	fetched, err := h.transactionGroups.GetByID(r.Context(), userID, group.ID)
	if err != nil || fetched == nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to retrieve transaction group", requestID)
		return
	}

	WriteJSON(w, http.StatusCreated, groupToResponse(*fetched, memberTxns))
}

func (h *AppHandler) GetTransactionGroup(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	group, err := h.transactionGroups.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get transaction group", requestID)
		return
	}
	if group == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Transaction group not found", requestID)
		return
	}

	members, err := h.transactionGroups.ListMembers(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to load group members", requestID)
		return
	}

	WriteJSON(w, http.StatusOK, groupToResponse(*group, members))
}

func (h *AppHandler) UpdateTransactionGroup(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	group, err := h.transactionGroups.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get transaction group", requestID)
		return
	}
	if group == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Transaction group not found", requestID)
		return
	}

	var req UpdateTransactionGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	if req.Name != nil {
		group.Name = *req.Name
	}

	if _, err := h.transactionGroups.Update(r.Context(), group); err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to update transaction group", requestID)
		return
	}

	if req.Members != nil {
		if err := h.syncGroupMembers(r.Context(), userID, group.ID, *req.Members); err != nil {
			WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to update group members", requestID)
			return
		}
	}

	members, err := h.transactionGroups.ListMembers(r.Context(), userID, group.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to load group members", requestID)
		return
	}

	updated, err := h.transactionGroups.GetByID(r.Context(), userID, group.ID)
	if err != nil || updated == nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to retrieve transaction group", requestID)
		return
	}

	WriteJSON(w, http.StatusOK, groupToResponse(*updated, members))
}

func (h *AppHandler) DeleteTransactionGroup(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	group, err := h.transactionGroups.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get transaction group", requestID)
		return
	}
	if group == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Transaction group not found", requestID)
		return
	}

	found, err := h.transactionGroups.SoftDelete(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to delete transaction group", requestID)
		return
	}
	if !found {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Transaction group not found", requestID)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AppHandler) syncGroupMembers(ctx context.Context, userID, groupID string, members []TransactionGroupMemberUpdate) error {
	existing, err := h.transactionGroups.ListMembers(ctx, userID, groupID)
	if err != nil {
		return err
	}

	existingByID := make(map[string]models.Transaction, len(existing))
	for _, txn := range existing {
		existingByID[txn.ID] = txn
	}

	seenIDs := make(map[string]bool)
	for _, m := range members {
		if m.Id != nil {
			txnID := m.Id.String()
			seenIDs[txnID] = true
			txn, exists := existingByID[txnID]
			if !exists {
				continue
			}
			applyMemberUpdate(&txn, m)
			if _, err := h.transactions.Update(ctx, &txn, tagsFromPtr(m.Tags)); err != nil {
				return err
			}
		} else {
			newTxn := memberUpdateToTransaction(userID, groupID, m)
			if err := h.transactions.Create(ctx, &newTxn, tagsFromPtr(m.Tags)); err != nil {
				return err
			}
		}
	}

	for _, txn := range existing {
		if !seenIDs[txn.ID] {
			if _, err := h.transactions.SoftDelete(ctx, userID, txn.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

func memberInputToTransaction(userID, groupID string, m TransactionGroupMemberInput) models.Transaction {
	currency := "USD"
	txn := models.Transaction{
		ID:        uuid.New().String(),
		UserID:    userID,
		AccountID: m.AccountId.String(),
		Type:      string(m.Type),
		Category:  m.Category,
		Amount:    float64(m.Amount),
		Currency:  currency,
		Date:      m.Date.Time,
		Notes:     m.Notes,
		GroupID:   &groupID,
	}
	return txn
}

func memberUpdateToTransaction(userID, groupID string, m TransactionGroupMemberUpdate) models.Transaction {
	return models.Transaction{
		ID:        uuid.New().String(),
		UserID:    userID,
		AccountID: m.AccountId.String(),
		Type:      string(m.Type),
		Category:  m.Category,
		Amount:    float64(m.Amount),
		Currency:  "USD",
		Date:      m.Date.Time,
		Notes:     m.Notes,
		GroupID:   &groupID,
	}
}

func applyMemberUpdate(txn *models.Transaction, m TransactionGroupMemberUpdate) {
	txn.Type = string(m.Type)
	txn.Category = m.Category
	txn.Amount = float64(m.Amount)
	txn.AccountID = m.AccountId.String()
	txn.Date = m.Date.Time
	txn.Notes = m.Notes
}

func groupToResponse(group models.TransactionGroup, members []models.Transaction) TransactionGroupResponse {
	memberResponses := make([]TransactionResponse, len(members))
	var total float64
	for i, m := range members {
		memberResponses[i] = transactionToResponse(m)
		total += m.Amount
	}

	return TransactionGroupResponse{
		Id:        ParseID(group.ID),
		UserId:    ParseID(group.UserID),
		Name:      group.Name,
		Total:     float32(total),
		Members:   memberResponses,
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}
}

func tagsFromPtr(tags *[]string) []string {
	if tags == nil {
		return nil
	}
	return *tags
}
