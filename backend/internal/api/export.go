package api

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/middleware"
	"github.com/jamespsullivan/pennywise/internal/models"
)

func emptyFilter() queries.TransactionFilter {
	return queries.TransactionFilter{}
}

type exportData struct {
	user      *models.User
	accounts  []models.Account
	txns      []models.Transaction
	assets    []models.Asset
	goals     []models.Goal
	recurring []models.RecurringTransaction
	alerts    []models.Alert
}

func (h *AppHandler) fetchExportData(ctx context.Context, userID string) (*exportData, string, bool) {
	user, err := h.users.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, "Failed to get user", false
	}

	accounts, _, err := h.accounts.List(ctx, userID, 1, 10000)
	if err != nil {
		return nil, "Failed to export accounts", false
	}

	txns, _, err := h.transactions.List(ctx, userID, emptyFilter(), 1, 10000)
	if err != nil {
		return nil, "Failed to export transactions", false
	}

	assets, _, err := h.assets.List(ctx, userID, 1, 10000)
	if err != nil {
		return nil, "Failed to export assets", false
	}

	goals, _, err := h.goals.List(ctx, userID, 1, 10000)
	if err != nil {
		return nil, "Failed to export goals", false
	}

	recurring, _, err := h.recurring.List(ctx, userID, 1, 10000)
	if err != nil {
		return nil, "Failed to export recurring", false
	}

	alerts, _, err := h.alerts.List(ctx, userID, 1, 10000)
	if err != nil {
		return nil, "Failed to export alerts", false
	}

	return &exportData{
		user:      user,
		accounts:  accounts,
		txns:      txns,
		assets:    assets,
		goals:     goals,
		recurring: recurring,
		alerts:    alerts,
	}, "", true
}

func buildExportResponse(data *exportData) ExportResponse {
	accountData := make([]AccountResponse, len(data.accounts))
	for i, a := range data.accounts {
		accountData[i] = accountToResponse(a)
	}

	txnData := make([]TransactionResponse, len(data.txns))
	for i, t := range data.txns {
		txnData[i] = transactionToResponse(t)
	}

	assetData := make([]AssetResponse, len(data.assets))
	for i, a := range data.assets {
		assetData[i] = assetToResponse(a)
	}

	goalData := make([]GoalResponse, len(data.goals))
	for i, g := range data.goals {
		goalData[i] = goalToResponse(g)
	}

	recurringData := make([]RecurringResponse, len(data.recurring))
	for i, rt := range data.recurring {
		recurringData[i] = recurringToResponse(rt)
	}

	alertData := make([]AlertResponse, len(data.alerts))
	for i, a := range data.alerts {
		alertData[i] = alertToResponse(a)
	}

	return ExportResponse{
		User:                  userToResponse(*data.user),
		Accounts:              accountData,
		Transactions:          txnData,
		Assets:                assetData,
		Goals:                 goalData,
		RecurringTransactions: recurringData,
		Alerts:                alertData,
		ExportedAt:            time.Now(),
	}
}

func (h *AppHandler) ExportData(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	data, errMsg, ok := h.fetchExportData(r.Context(), userID)
	if !ok {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, errMsg, requestID)
		return
	}

	WriteJSON(w, http.StatusOK, buildExportResponse(data))
}

func (h *AppHandler) ExportCsv(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	requestID := middleware.GetRequestID(r.Context())

	txns, _, err := h.transactions.List(r.Context(), userID, emptyFilter(), 1, 100000)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, INTERNALERROR, "Failed to export transactions", requestID)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=transactions.csv")
	w.WriteHeader(http.StatusOK)

	writer := csv.NewWriter(w)
	defer writer.Flush()

	header := []string{"id", "account_id", "type", "category", "amount", "currency", "date", "notes", "is_recurring", "tags"}
	_ = writer.Write(header)

	for _, t := range txns {
		_ = writer.Write(transactionToCsvRow(t))
	}
}

func transactionToCsvRow(t models.Transaction) []string {
	notes := ""
	if t.Notes != nil {
		notes = *t.Notes
	}
	tags := ""
	for j, tag := range t.Tags {
		if j > 0 {
			tags += ";"
		}
		tags += tag
	}
	return []string{
		t.ID,
		t.AccountID,
		t.Type,
		t.Category,
		fmt.Sprintf("%.2f", t.Amount),
		t.Currency,
		t.Date.Format("2006-01-02"),
		notes,
		fmt.Sprintf("%t", t.IsRecurring),
		tags,
	}
}

func userToResponse(u models.User) UserResponse {
	return UserResponse{
		Id:    ParseID(u.ID),
		Email: u.Email,
		Name:  u.Name,
	}
}
