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

func setupGroupTests(t *testing.T) (http.Handler, *http.Cookie) {
	t.Helper()
	database, router := setupRouter(t)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		txnTestAccountID, txnTestUserID, "Checking", "Chase", "checking", "USD", 1,
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		txnTestAccount2ID, txnTestUserID, "Savings", "Chase", "savings", "USD", 1,
	)
	require.NoError(t, err)

	cookie := loginAndGetCookie(t, router)
	return router, cookie
}

func createTestGroup(t *testing.T, router http.Handler, cookie *http.Cookie) api.TransactionGroupResponse {
	t.Helper()
	body := fmt.Sprintf(`{
		"name": "March Paycheck",
		"members": [
			{"type":"deposit","category":"salary","amount":4000,"date":"2026-03-08","account_id":"%s"},
			{"type":"deposit","category":"401k","amount":500,"date":"2026-03-08","account_id":"%s"}
		]
	}`, txnTestAccountID, txnTestAccount2ID)

	req := authedRequest(http.MethodPost, "/api/v1/transaction-groups", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var resp api.TransactionGroupResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	return resp
}

func TestCreateTransactionGroup_ValidRequest(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)

	resp := createTestGroup(t, router, cookie)

	assert.Equal(t, "March Paycheck", resp.Name)
	assert.Len(t, resp.Members, 2)
	assert.InDelta(t, 4500.0, float64(resp.Total), 0.01)
	assert.NotEmpty(t, resp.Id)
}

func TestCreateTransactionGroup_TooFewMembers_Returns400(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)

	body := fmt.Sprintf(`{
		"name": "Bad Group",
		"members": [
			{"type":"deposit","category":"salary","amount":4000,"date":"2026-03-08","account_id":"%s"}
		]
	}`, txnTestAccountID)

	req := authedRequest(http.MethodPost, "/api/v1/transaction-groups", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetTransactionGroup_Found(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)
	created := createTestGroup(t, router, cookie)

	req := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/transaction-groups/%s", created.Id), "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionGroupResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "March Paycheck", resp.Name)
	assert.Len(t, resp.Members, 2)
}

func TestGetTransactionGroup_NotFound(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)

	req := authedRequest(http.MethodGet, "/api/v1/transaction-groups/00000000-0000-0000-0000-000000000099", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListTransactionGroups(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)
	createTestGroup(t, router, cookie)

	req := authedRequest(http.MethodGet, "/api/v1/transaction-groups", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionGroupListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, 1, resp.Pagination.Total)
}

func TestUpdateTransactionGroup_Name(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)
	created := createTestGroup(t, router, cookie)

	body := `{"name":"April Paycheck"}`
	req := authedRequest(http.MethodPut, fmt.Sprintf("/api/v1/transaction-groups/%s", created.Id), body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionGroupResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "April Paycheck", resp.Name)
	assert.Len(t, resp.Members, 2)
}

func TestUpdateTransactionGroup_WithMembers(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)
	created := createTestGroup(t, router, cookie)
	existingMemberID := created.Members[0].Id.String()

	body := fmt.Sprintf(`{
		"name":"Updated Paycheck",
		"members":[
			{"id":"%s","type":"deposit","category":"salary","amount":4500,"date":"2026-03-08","account_id":"%s"},
			{"type":"deposit","category":"hsa","amount":200,"date":"2026-03-08","account_id":"%s"}
		]
	}`, existingMemberID, txnTestAccountID, txnTestAccount2ID)

	req := authedRequest(http.MethodPut, fmt.Sprintf("/api/v1/transaction-groups/%s", created.Id), body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionGroupResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Updated Paycheck", resp.Name)
	assert.Len(t, resp.Members, 2)
	assert.InDelta(t, 4700.0, float64(resp.Total), 0.01)
}

func TestDeleteTransactionGroup(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)
	created := createTestGroup(t, router, cookie)

	req := authedRequest(http.MethodDelete, fmt.Sprintf("/api/v1/transaction-groups/%s", created.Id), "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	getReq := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/transaction-groups/%s", created.Id), "", cookie)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	assert.Equal(t, http.StatusNotFound, getRec.Code)
}

func TestDeleteTransactionGroup_CascadesToMembers(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)
	created := createTestGroup(t, router, cookie)
	memberID := created.Members[0].Id.String()

	req := authedRequest(http.MethodDelete, fmt.Sprintf("/api/v1/transaction-groups/%s", created.Id), "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNoContent, rec.Code)

	getReq := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/transactions/%s", memberID), "", cookie)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	assert.Equal(t, http.StatusNotFound, getRec.Code)
}

func TestCreateTransactionGroup_InvalidJSON_Returns400(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)

	req := authedRequest(http.MethodPost, "/api/v1/transaction-groups", "not json", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateTransactionGroup_NotFound_Returns404(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)

	body := `{"name":"Nope"}`
	req := authedRequest(http.MethodPut, "/api/v1/transaction-groups/00000000-0000-0000-0000-000000000099", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateTransactionGroup_InvalidJSON_Returns400(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)
	created := createTestGroup(t, router, cookie)

	req := authedRequest(http.MethodPut, fmt.Sprintf("/api/v1/transaction-groups/%s", created.Id), "not json", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDeleteTransactionGroup_NotFound_Returns404(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)

	req := authedRequest(http.MethodDelete, "/api/v1/transaction-groups/00000000-0000-0000-0000-000000000099", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestCreateTransactionGroup_EmptyMembers_Returns400(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)

	body := `{"name":"Empty Group","members":[]}`
	req := authedRequest(http.MethodPost, "/api/v1/transaction-groups", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListTransactionGroups_Empty(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)

	req := authedRequest(http.MethodGet, "/api/v1/transaction-groups", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionGroupListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 0)
	assert.Equal(t, 0, resp.Pagination.Total)
}

func TestListTransactions_FilterByGroupID(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)
	created := createTestGroup(t, router, cookie)

	req := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/transactions?group_id=%s", created.Id), "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 2)
}

func TestCreateTransactionGroup_WithTags(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)

	body := fmt.Sprintf(`{
		"name": "Tagged Group",
		"members": [
			{"type":"deposit","category":"salary","amount":4000,"date":"2026-03-08","account_id":"%s","tags":["income","paycheck"]},
			{"type":"deposit","category":"401k","amount":500,"date":"2026-03-08","account_id":"%s","tags":["retirement"]}
		]
	}`, txnTestAccountID, txnTestAccount2ID)

	req := authedRequest(http.MethodPost, "/api/v1/transaction-groups", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp api.TransactionGroupResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Tagged Group", resp.Name)
	assert.Len(t, resp.Members, 2)

	foundIncomeTags := false
	foundRetirementTag := false
	for _, m := range resp.Members {
		for _, tag := range m.Tags {
			if tag == "income" {
				foundIncomeTags = true
			}
			if tag == "retirement" {
				foundRetirementTag = true
			}
		}
	}
	assert.True(t, foundIncomeTags)
	assert.True(t, foundRetirementTag)
}

func TestUpdateTransactionGroup_WithNewMemberTags(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)
	created := createTestGroup(t, router, cookie)

	body := fmt.Sprintf(`{
		"members":[
			{"type":"deposit","category":"salary","amount":4500,"date":"2026-03-08","account_id":"%s","tags":["monthly"]},
			{"type":"deposit","category":"bonus","amount":1000,"date":"2026-03-08","account_id":"%s","tags":["quarterly"]}
		]
	}`, txnTestAccountID, txnTestAccount2ID)

	req := authedRequest(http.MethodPut, fmt.Sprintf("/api/v1/transaction-groups/%s", created.Id), body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionGroupResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Members, 2)
	assert.InDelta(t, 5500.0, float64(resp.Total), 0.01)
}

func TestTransactionResponse_IncludesGroupID(t *testing.T) {
	t.Parallel()
	router, cookie := setupGroupTests(t)
	created := createTestGroup(t, router, cookie)
	memberID := created.Members[0].Id.String()

	req := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/transactions/%s", memberID), "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotNil(t, resp.GroupId)
	assert.Equal(t, created.Id.String(), resp.GroupId.String())
}
