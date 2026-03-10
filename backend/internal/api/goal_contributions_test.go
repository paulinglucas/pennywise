package api_test

import (
	"bytes"
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

func createTestGoal(t *testing.T, router http.Handler, cookie *http.Cookie) api.GoalResponse {
	t.Helper()
	body := `{"name":"Contribution Test Goal","goal_type":"savings","target_amount":10000,"current_amount":1000}`
	req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	return resp
}

func TestCreateGoalContribution_Success(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)

	body := `{"amount":500,"notes":"Monthly savings"}`
	req := authedRequest(http.MethodPost, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.GoalContributionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, float32(500), resp.Amount)
	assert.Equal(t, "Monthly savings", *resp.Notes)
	assert.Equal(t, goal.Id, resp.GoalId)
}

func TestCreateGoalContribution_UpdatesGoalAmount(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)

	body := `{"amount":500}`
	req := authedRequest(http.MethodPost, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	getReq := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/goals/%s", goal.Id), "", cookie)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	require.Equal(t, http.StatusOK, getRec.Code)

	var updated api.GoalResponse
	require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &updated))
	assert.Equal(t, float32(1500), updated.CurrentAmount)
}

func TestCreateGoalContribution_WithDate(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)

	body := `{"amount":250,"contributed_at":"2026-01-15"}`
	req := authedRequest(http.MethodPost, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.GoalContributionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "2026-01-15", resp.ContributedAt.Format("2006-01-02"))
}

func TestCreateGoalContribution_GoalNotFound(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"amount":500}`
	req := authedRequest(http.MethodPost, "/api/v1/goals/00000099-0000-0000-0000-000000000099/contributions", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCreateGoalContribution_InvalidBody(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)

	body := `{"amount":-5}`
	req := authedRequest(http.MethodPost, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateGoalContribution_NoAuth(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)

	body := `{"amount":500}`
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestListGoalContributions_Success(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)

	for i := 0; i < 3; i++ {
		body := fmt.Sprintf(`{"amount":%d}`, (i+1)*100)
		req := authedRequest(http.MethodPost, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), body, cookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
	}

	req := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.GoalContributionListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 3)
	assert.Equal(t, 3, resp.Pagination.Total)
}

func TestListGoalContributions_GoalNotFound(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodGet, "/api/v1/goals/00000099-0000-0000-0000-000000000099/contributions", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListGoalContributions_Pagination(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)

	for i := 0; i < 5; i++ {
		body := fmt.Sprintf(`{"amount":%d}`, (i+1)*100)
		req := authedRequest(http.MethodPost, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), body, cookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
	}

	req := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/goals/%s/contributions?page=1&per_page=2", goal.Id), "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.GoalContributionListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, 5, resp.Pagination.Total)
	assert.Equal(t, 3, resp.Pagination.TotalPages)
}

func TestDeleteGoalContribution_Success(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)

	body := `{"amount":500}`
	createReq := authedRequest(http.MethodPost, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), body, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var contrib api.GoalContributionResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &contrib))

	deleteReq := authedRequest(http.MethodDelete, fmt.Sprintf("/api/v1/goals/%s/contributions/%s", goal.Id, contrib.Id), "", cookie)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)

	assert.Equal(t, http.StatusNoContent, deleteRec.Code)
}

func TestDeleteGoalContribution_ReversesGoalAmount(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)

	body := `{"amount":500}`
	createReq := authedRequest(http.MethodPost, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), body, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var contrib api.GoalContributionResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &contrib))

	deleteReq := authedRequest(http.MethodDelete, fmt.Sprintf("/api/v1/goals/%s/contributions/%s", goal.Id, contrib.Id), "", cookie)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)
	require.Equal(t, http.StatusNoContent, deleteRec.Code)

	getReq := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/goals/%s", goal.Id), "", cookie)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	require.Equal(t, http.StatusOK, getRec.Code)

	var updated api.GoalResponse
	require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &updated))
	assert.Equal(t, float32(1000), updated.CurrentAmount)
}

func TestDeleteGoalContribution_NotFound(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)

	req := authedRequest(http.MethodDelete, fmt.Sprintf("/api/v1/goals/%s/contributions/00000099-0000-0000-0000-000000000099", goal.Id), "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDeleteGoalContribution_GoalNotFound(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodDelete, "/api/v1/goals/00000099-0000-0000-0000-000000000099/contributions/00000088-0000-0000-0000-000000000088", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCreateGoalContribution_AuditLog(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)

	body := `{"amount":500}`
	req := authedRequest(http.MethodPost, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var count int
	err := database.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM audit_log WHERE entity_type = 'goal' AND entity_id = ? AND action = 'contribute'",
		goal.Id.String(),
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCreateGoalContribution_MultipleContributions(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)

	amounts := []int{500, 300, 200}
	for _, amt := range amounts {
		body := fmt.Sprintf(`{"amount":%d}`, amt)
		req := authedRequest(http.MethodPost, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), body, cookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
	}

	getReq := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/goals/%s", goal.Id), "", cookie)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	require.Equal(t, http.StatusOK, getRec.Code)

	var updated api.GoalResponse
	require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &updated))
	assert.Equal(t, float32(2000), updated.CurrentAmount)
}

func createTestTransaction(t *testing.T, router http.Handler, cookie *http.Cookie, accountID string) string {
	t.Helper()
	body := fmt.Sprintf(`{"account_id":"%s","type":"expense","category":"savings","amount":500,"currency":"USD","date":"2026-03-01"}`, accountID)
	req := authedRequest(http.MethodPost, "/api/v1/transactions", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var resp api.TransactionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	return resp.Id.String()
}

func TestCreateGoalContribution_WithTransactionID(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)
	accountID := createTestAccount(t, router, cookie)
	txnID := createTestTransaction(t, router, cookie, accountID)

	body := fmt.Sprintf(`{"amount":500,"transaction_id":"%s"}`, txnID)
	req := authedRequest(http.MethodPost, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.GoalContributionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotNil(t, resp.TransactionId)
	assert.Equal(t, txnID, resp.TransactionId.String())
}

func TestCreateGoalContribution_WithoutTransactionID(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)

	body := `{"amount":500}`
	req := authedRequest(http.MethodPost, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.GoalContributionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Nil(t, resp.TransactionId)
}

func TestCreateGoalContribution_InvalidTransactionID(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)

	body := `{"amount":500,"transaction_id":"00000000-0000-0000-0000-000000000099"}`
	req := authedRequest(http.MethodPost, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateGoalContribution_TransactionIDInListResponse(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	goal := createTestGoal(t, router, cookie)
	accountID := createTestAccount(t, router, cookie)
	txnID := createTestTransaction(t, router, cookie, accountID)

	body := fmt.Sprintf(`{"amount":500,"transaction_id":"%s"}`, txnID)
	req := authedRequest(http.MethodPost, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	listReq := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/goals/%s/contributions", goal.Id), "", cookie)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)
	require.Equal(t, http.StatusOK, listRec.Code)

	var listResp api.GoalContributionListResponse
	require.NoError(t, json.Unmarshal(listRec.Body.Bytes(), &listResp))
	require.Len(t, listResp.Data, 1)
	require.NotNil(t, listResp.Data[0].TransactionId)
	assert.Equal(t, txnID, listResp.Data[0].TransactionId.String())
}
