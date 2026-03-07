package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/api"
)

func TestExportData_EmptyUser(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodGet, "/api/v1/export", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.ExportResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "james@example.com", resp.User.Email)
	assert.Len(t, resp.Accounts, 0)
	assert.Len(t, resp.Transactions, 0)
	assert.Len(t, resp.Assets, 0)
	assert.Len(t, resp.Goals, 0)
	assert.Len(t, resp.RecurringTransactions, 0)
	assert.Len(t, resp.Alerts, 0)
	assert.False(t, resp.ExportedAt.IsZero())
}

func TestExportData_WithData(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	accountBody := `{"name":"Checking","account_type":"checking","institution":"Test Bank","balance":5000,"currency":"USD"}`
	accountReq := authedRequest(http.MethodPost, "/api/v1/accounts", accountBody, cookie)
	accountRec := httptest.NewRecorder()
	router.ServeHTTP(accountRec, accountReq)
	require.Equal(t, http.StatusCreated, accountRec.Code)

	var acct api.AccountResponse
	require.NoError(t, json.Unmarshal(accountRec.Body.Bytes(), &acct))

	txnBody := fmt.Sprintf(`{"account_id":"%s","type":"expense","category":"food","amount":25.50,"date":"2026-03-01"}`, acct.Id)
	txnReq := authedRequest(http.MethodPost, "/api/v1/transactions", txnBody, cookie)
	txnRec := httptest.NewRecorder()
	router.ServeHTTP(txnRec, txnReq)
	require.Equal(t, http.StatusCreated, txnRec.Code)

	req := authedRequest(http.MethodGet, "/api/v1/export", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.ExportResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Accounts, 1)
	assert.Len(t, resp.Transactions, 1)
}

func TestExportData_NoAuth_Returns401(t *testing.T) {
	_, router := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/export", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestExportCsv_Headers(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodGet, "/api/v1/export/csv", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/csv", rec.Header().Get("Content-Type"))
	assert.Contains(t, rec.Header().Get("Content-Disposition"), "transactions.csv")

	body := rec.Body.String()
	lines := strings.Split(strings.TrimSpace(body), "\n")
	require.GreaterOrEqual(t, len(lines), 1)
	assert.Equal(t, "id,account_id,type,category,amount,currency,date,notes,is_recurring,tags", lines[0])
}

func TestExportCsv_WithTransactions(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	accountBody := `{"name":"CSV Account","account_type":"checking","institution":"Test Bank","balance":1000,"currency":"USD"}`
	accountReq := authedRequest(http.MethodPost, "/api/v1/accounts", accountBody, cookie)
	accountRec := httptest.NewRecorder()
	router.ServeHTTP(accountRec, accountReq)
	require.Equal(t, http.StatusCreated, accountRec.Code)

	var acct api.AccountResponse
	require.NoError(t, json.Unmarshal(accountRec.Body.Bytes(), &acct))

	txnBody := fmt.Sprintf(`{"account_id":"%s","type":"expense","category":"groceries","amount":42.50,"date":"2026-03-01"}`, acct.Id)
	txnReq := authedRequest(http.MethodPost, "/api/v1/transactions", txnBody, cookie)
	txnRec := httptest.NewRecorder()
	router.ServeHTTP(txnRec, txnReq)
	require.Equal(t, http.StatusCreated, txnRec.Code)

	req := authedRequest(http.MethodGet, "/api/v1/export/csv", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	body := rec.Body.String()
	lines := strings.Split(strings.TrimSpace(body), "\n")
	require.Len(t, lines, 2)
	assert.Contains(t, lines[1], "groceries")
	assert.Contains(t, lines[1], "42.50")
}

func TestExportCsv_WithNotesAndTags(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	accountBody := `{"name":"Tag Account","account_type":"checking","institution":"Test Bank","currency":"USD"}`
	accountReq := authedRequest(http.MethodPost, "/api/v1/accounts", accountBody, cookie)
	accountRec := httptest.NewRecorder()
	router.ServeHTTP(accountRec, accountReq)
	require.Equal(t, http.StatusCreated, accountRec.Code)

	var acct api.AccountResponse
	require.NoError(t, json.Unmarshal(accountRec.Body.Bytes(), &acct))

	txnBody := fmt.Sprintf(`{"account_id":"%s","type":"expense","category":"dining","amount":55,"date":"2026-03-01","notes":"dinner with friends","tags":["food","social"]}`, acct.Id)
	txnReq := authedRequest(http.MethodPost, "/api/v1/transactions", txnBody, cookie)
	txnRec := httptest.NewRecorder()
	router.ServeHTTP(txnRec, txnReq)
	require.Equal(t, http.StatusCreated, txnRec.Code)

	req := authedRequest(http.MethodGet, "/api/v1/export/csv", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	body := rec.Body.String()
	lines := strings.Split(strings.TrimSpace(body), "\n")
	require.Len(t, lines, 2)
	assert.Contains(t, lines[1], "dinner with friends")
	assert.Contains(t, lines[1], "food;social")
}

func TestExportData_IncludesAllEntityTypes(t *testing.T) {
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	accountBody := `{"name":"Full Export","account_type":"checking","institution":"Test Bank","currency":"USD"}`
	accountReq := authedRequest(http.MethodPost, "/api/v1/accounts", accountBody, cookie)
	accountRec := httptest.NewRecorder()
	router.ServeHTTP(accountRec, accountReq)
	require.Equal(t, http.StatusCreated, accountRec.Code)

	var acct api.AccountResponse
	require.NoError(t, json.Unmarshal(accountRec.Body.Bytes(), &acct))

	txnBody := fmt.Sprintf(`{"account_id":"%s","type":"expense","category":"food","amount":10,"date":"2026-03-01"}`, acct.Id)
	txnReq := authedRequest(http.MethodPost, "/api/v1/transactions", txnBody, cookie)
	txnRec := httptest.NewRecorder()
	router.ServeHTTP(txnRec, txnReq)
	require.Equal(t, http.StatusCreated, txnRec.Code)

	goalBody := `{"name":"Save","goal_type":"savings","target_amount":1000}`
	goalReq := authedRequest(http.MethodPost, "/api/v1/goals", goalBody, cookie)
	goalRec := httptest.NewRecorder()
	router.ServeHTTP(goalRec, goalReq)
	require.Equal(t, http.StatusCreated, goalRec.Code)

	assetBody := fmt.Sprintf(`{"name":"Brokerage Fund","asset_type":"brokerage","current_value":500,"currency":"USD","account_id":"%s"}`, acct.Id)
	assetReq := authedRequest(http.MethodPost, "/api/v1/assets", assetBody, cookie)
	assetRec := httptest.NewRecorder()
	router.ServeHTTP(assetRec, assetReq)
	require.Equal(t, http.StatusCreated, assetRec.Code)

	recBody := fmt.Sprintf(`{"account_id":"%s","type":"expense","category":"rent","amount":1500,"frequency":"monthly","next_occurrence":"2026-04-01"}`, acct.Id)
	recReq := authedRequest(http.MethodPost, "/api/v1/recurring", recBody, cookie)
	recRec := httptest.NewRecorder()
	router.ServeHTTP(recRec, recReq)
	require.Equal(t, http.StatusCreated, recRec.Code)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO alerts (id, user_id, alert_type, message, severity, is_read) VALUES (?, ?, ?, ?, ?, ?)`,
		"e0000001-0000-0000-0000-000000000001", testUserID, "budget_exceeded", "Over budget", "warning", 0,
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/export", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.ExportResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "james@example.com", resp.User.Email)
	assert.GreaterOrEqual(t, len(resp.Accounts), 1)
	assert.GreaterOrEqual(t, len(resp.Transactions), 1)
	assert.GreaterOrEqual(t, len(resp.Goals), 1)
	assert.GreaterOrEqual(t, len(resp.Assets), 1)
	assert.GreaterOrEqual(t, len(resp.RecurringTransactions), 1)
	assert.GreaterOrEqual(t, len(resp.Alerts), 1)
}
