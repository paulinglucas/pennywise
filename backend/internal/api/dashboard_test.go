package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/api"
)

func setupDashboardData(t *testing.T, router http.Handler, cookie *http.Cookie) string {
	t.Helper()
	accountID := createTestAccount(t, router, cookie)

	txns := []struct {
		txnType  string
		category string
		amount   int
		date     string
	}{
		{"deposit", "salary", 5000, "2026-03-01"},
		{"expense", "food", 300, "2026-03-05"},
		{"expense", "housing", 1500, "2026-03-01"},
		{"expense", "utilities", 200, "2026-03-03"},
	}

	for i, txn := range txns {
		body := fmt.Sprintf(`{"account_id":"%s","type":"%s","category":"%s","amount":%d,"date":"%s"}`,
			accountID, txn.txnType, txn.category, txn.amount, txn.date)
		req := authedRequest(http.MethodPost, "/api/v1/transactions", body, cookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code, "transaction %d failed", i)
	}

	assetBody := fmt.Sprintf(`{"name":"Brokerage","asset_type":"brokerage","current_value":50000,"currency":"USD","account_id":"%s"}`, accountID)
	assetReq := authedRequest(http.MethodPost, "/api/v1/assets", assetBody, cookie)
	assetRec := httptest.NewRecorder()
	router.ServeHTTP(assetRec, assetReq)
	require.Equal(t, http.StatusCreated, assetRec.Code)

	return accountID
}

func TestGetDashboard_EmptyUser(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodGet, "/api/v1/dashboard", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.DashboardResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, float32(0), resp.NetWorth)
	assert.Equal(t, float32(0), resp.CashFlowThisMonth)
	assert.Empty(t, resp.SpendingByCategory)
	assert.Empty(t, resp.DebtsSummary)
}

func TestGetDashboard_WithData(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	setupDashboardData(t, router, cookie)

	req := authedRequest(http.MethodGet, "/api/v1/dashboard", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.DashboardResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, float32(53000), resp.NetWorth)
	assert.Equal(t, float32(3000), resp.CashFlowThisMonth)
	assert.Len(t, resp.SpendingByCategory, 3)
}

func TestGetDashboard_SpendingPercentagesSumTo100(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	setupDashboardData(t, router, cookie)

	req := authedRequest(http.MethodGet, "/api/v1/dashboard", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp api.DashboardResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	var totalPct float32
	for _, cat := range resp.SpendingByCategory {
		totalPct += cat.Percentage
	}
	assert.InDelta(t, 100.0, totalPct, 1.0)
}

func TestGetDashboard_NoAuth_Returns401(t *testing.T) {
	_, router := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetDashboard_DebtsSummary(t *testing.T) {
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	accountID := createTestAccount(t, router, cookie)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"cc000001-0000-0000-0000-000000000001", testUserID, "Chase Card", "Chase", "credit_card", "USD",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank, linked_account_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"gl000001-0000-0000-0000-000000000001", testUserID, "Pay off CC", "debt_payoff", 5000, 3000, 1, "cc000001-0000-0000-0000-000000000001",
	)
	require.NoError(t, err)

	txnBody := fmt.Sprintf(`{"account_id":"%s","type":"expense","category":"cc_payment","amount":500,"date":"2026-03-05"}`, accountID)
	txnReq := authedRequest(http.MethodPost, "/api/v1/transactions", txnBody, cookie)
	txnRec := httptest.NewRecorder()
	router.ServeHTTP(txnRec, txnReq)

	req := authedRequest(http.MethodGet, "/api/v1/dashboard", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.DashboardResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.DebtsSummary, 1)
	assert.Equal(t, "Chase Card", resp.DebtsSummary[0].Name)
	assert.Equal(t, float32(3000), resp.DebtsSummary[0].Balance)
}

func TestGetDashboard_SpendingPeriod(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	setupDashboardData(t, router, cookie)

	req := authedRequest(http.MethodGet, "/api/v1/dashboard?spending_period=1y", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.DashboardResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.SpendingByCategory, 3)
}

func TestGetDashboard_DebtsOriginalBalance(t *testing.T) {
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, original_balance) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"cc000002-0000-0000-0000-000000000001", testUserID, "Visa Card", "Visa", "credit_card", "USD", 10000.0,
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank, linked_account_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"gl000002-0000-0000-0000-000000000001", testUserID, "Pay off Visa", "debt_payoff", 10000, 4000, 1, "cc000002-0000-0000-0000-000000000001",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/dashboard", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.DashboardResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.DebtsSummary, 1)
	require.NotNil(t, resp.DebtsSummary[0].OriginalBalance)
	assert.Equal(t, float32(10000), *resp.DebtsSummary[0].OriginalBalance)
}

func TestGetNetWorthHistory_Empty(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodGet, "/api/v1/dashboard/net-worth-history", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.NetWorthHistoryResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.DataPoints, 2)
	assert.Equal(t, float32(0), resp.DataPoints[0].Value)
	assert.Equal(t, float32(0), resp.DataPoints[1].Value)
}

func TestGetNetWorthHistory_WithPeriod(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodGet, "/api/v1/dashboard/net-worth-history?period=all", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetNetWorthHistory_NoAuth_Returns401(t *testing.T) {
	_, router := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/net-worth-history", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
