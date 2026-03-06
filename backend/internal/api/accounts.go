package api

import (
	"context"
	"encoding/json"
	"math"
	"net/http"

	"github.com/google/uuid"

	"github.com/jamespsullivan/pennywise/internal/middleware"
	"github.com/jamespsullivan/pennywise/internal/models"
)

type AccountRepository interface {
	List(ctx context.Context, userID string, page, perPage int) ([]models.Account, int, error)
	Create(ctx context.Context, account *models.Account) error
	GetByID(ctx context.Context, userID, id string) (*models.Account, error)
	Update(ctx context.Context, account *models.Account) (bool, error)
	SoftDelete(ctx context.Context, userID, id string) (bool, error)
}

func (h *AppHandler) ListAccounts(w http.ResponseWriter, r *http.Request, params ListAccountsParams) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	page, perPage := paginationDefaults(params.Page, params.PerPage)

	accounts, total, err := h.accounts.List(r.Context(), userID, page, perPage)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to list accounts", requestID)
		return
	}

	data := make([]AccountResponse, len(accounts))
	for i, a := range accounts {
		data[i] = accountToResponse(a)
	}

	WriteJSON(w, http.StatusOK, AccountListResponse{
		Data:       data,
		Pagination: paginationMeta(page, perPage, total),
	})
}

func (h *AppHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	if req.Name == "" || req.Institution == "" || req.AccountType == "" {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "name, institution, and account_type are required", requestID)
		return
	}

	account := newAccountModel(userID, req)

	if err := h.accounts.Create(r.Context(), account); err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to create account", requestID)
		return
	}

	created, err := h.accounts.GetByID(r.Context(), userID, account.ID)
	if err != nil || created == nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to retrieve account", requestID)
		return
	}

	WriteJSON(w, http.StatusCreated, accountToResponse(*created))
}

func (h *AppHandler) GetAccount(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	account, err := h.accounts.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get account", requestID)
		return
	}
	if account == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Account not found", requestID)
		return
	}

	WriteJSON(w, http.StatusOK, accountToResponse(*account))
}

func (h *AppHandler) UpdateAccount(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	account, err := h.accounts.GetByID(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to get account", requestID)
		return
	}
	if account == nil {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Account not found", requestID)
		return
	}

	var req UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, VALIDATIONFAILED, "Invalid request body", requestID)
		return
	}

	applyAccountUpdates(account, req)

	if _, err := h.accounts.Update(r.Context(), account); err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to update account", requestID)
		return
	}

	updated, err := h.accounts.GetByID(r.Context(), userID, id.String())
	if err != nil || updated == nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to retrieve account", requestID)
		return
	}

	WriteJSON(w, http.StatusOK, accountToResponse(*updated))
}

func (h *AppHandler) DeleteAccount(w http.ResponseWriter, r *http.Request, id IdParam) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	found, err := h.accounts.SoftDelete(r.Context(), userID, id.String())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to delete account", requestID)
		return
	}
	if !found {
		WriteError(w, http.StatusNotFound, NOTFOUND, "Account not found", requestID)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func newAccountModel(userID string, req CreateAccountRequest) *models.Account {
	currency := USD
	if req.Currency != nil {
		currency = *req.Currency
	}
	return &models.Account{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        req.Name,
		Institution: req.Institution,
		AccountType: string(req.AccountType),
		Currency:    string(currency),
		IsActive:    true,
	}
}

func accountToResponse(a models.Account) AccountResponse {
	return AccountResponse{
		Id:          ParseID(a.ID),
		UserId:      ParseID(a.UserID),
		Name:        a.Name,
		Institution: a.Institution,
		AccountType: AccountType(a.AccountType),
		Currency:    Currency(a.Currency),
		IsActive:    a.IsActive,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

func applyAccountUpdates(account *models.Account, req UpdateAccountRequest) {
	if req.Name != nil {
		account.Name = *req.Name
	}
	if req.Institution != nil {
		account.Institution = *req.Institution
	}
	if req.AccountType != nil {
		account.AccountType = string(*req.AccountType)
	}
	if req.Currency != nil {
		account.Currency = string(*req.Currency)
	}
	if req.IsActive != nil {
		account.IsActive = *req.IsActive
	}
}

func paginationDefaults(page, perPage *int) (int, int) {
	p := 1
	pp := 20
	if page != nil && *page > 0 {
		p = *page
	}
	if perPage != nil && *perPage > 0 {
		pp = *perPage
		if pp > 100 {
			pp = 100
		}
	}
	return p, pp
}

func paginationMeta(page, perPage, total int) PaginationMeta {
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	return PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}
