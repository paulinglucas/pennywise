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

func createTestAccount(t *testing.T, router http.Handler, cookie *http.Cookie) string {
	t.Helper()
	body := `{"name":"Test Checking","account_type":"checking","institution":"Test Bank","balance":5000,"currency":"USD"}`
	req := authedRequest(http.MethodPost, "/api/v1/accounts", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var resp api.AccountResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	return resp.Id.String()
}

func TestCreateRecurring_ValidRequest(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	accountID := createTestAccount(t, router, cookie)

	body := fmt.Sprintf(`{"account_id":"%s","type":"expense","category":"utilities","amount":150,"currency":"USD","frequency":"monthly","next_occurrence":"2026-04-01"}`, accountID)
	req := authedRequest(http.MethodPost, "/api/v1/recurring", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.RecurringResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "utilities", resp.Category)
	assert.Equal(t, float32(150), resp.Amount)
	assert.Equal(t, api.Monthly, resp.Frequency)
	assert.True(t, resp.IsActive)
	assert.NotEmpty(t, resp.Id)
}

func TestCreateRecurring_MissingFields_Returns400(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"category":"utilities"}`
	req := authedRequest(http.MethodPost, "/api/v1/recurring", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListRecurring_NoAuth_Returns401(t *testing.T) {
	_, router := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/recurring", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestListRecurring_Success(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	accountID := createTestAccount(t, router, cookie)

	for i := 1; i <= 3; i++ {
		body := fmt.Sprintf(`{"account_id":"%s","type":"expense","category":"cat%d","amount":%d,"frequency":"monthly","next_occurrence":"2026-04-01"}`, accountID, i, i*100)
		req := authedRequest(http.MethodPost, "/api/v1/recurring", body, cookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
	}

	req := authedRequest(http.MethodGet, "/api/v1/recurring", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.RecurringListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 3)
	assert.Equal(t, 3, resp.Pagination.Total)
}

func TestListRecurring_Pagination(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	accountID := createTestAccount(t, router, cookie)

	for i := 1; i <= 3; i++ {
		body := fmt.Sprintf(`{"account_id":"%s","type":"expense","category":"cat%d","amount":%d,"frequency":"weekly","next_occurrence":"2026-04-01"}`, accountID, i, i*50)
		req := authedRequest(http.MethodPost, "/api/v1/recurring", body, cookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
	}

	req := authedRequest(http.MethodGet, "/api/v1/recurring?page=1&per_page=2", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.RecurringListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, 3, resp.Pagination.Total)
	assert.Equal(t, 2, resp.Pagination.TotalPages)
}

func TestListRecurring_UserScoping(t *testing.T) {
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		"usr00002-0000-0000-0000-000000000002", "alex@example.com", "Alex", "$2a$10$dummy",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"a0000002-0000-0000-0000-000000000002", "usr00002-0000-0000-0000-000000000002", "Alex Account", "Alex Bank", "checking", "USD",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO recurring_transactions (id, user_id, account_id, type, category, amount, currency, frequency, next_occurrence) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"b0000001-0000-0000-0000-000000000001", "usr00002-0000-0000-0000-000000000002",
		"a0000002-0000-0000-0000-000000000002", "expense", "rent", 1200, "USD", "monthly", "2026-04-01",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/recurring", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.RecurringListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 0)
}

func TestUpdateRecurring_ValidRequest(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	accountID := createTestAccount(t, router, cookie)

	createBody := fmt.Sprintf(`{"account_id":"%s","type":"expense","category":"utilities","amount":150,"frequency":"monthly","next_occurrence":"2026-04-01"}`, accountID)
	createReq := authedRequest(http.MethodPost, "/api/v1/recurring", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.RecurringResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	updateBody := `{"amount":200,"category":"electric"}`
	updateReq := authedRequest(http.MethodPut, fmt.Sprintf("/api/v1/recurring/%s", created.Id), updateBody, cookie)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	assert.Equal(t, http.StatusOK, updateRec.Code)

	var resp api.RecurringResponse
	require.NoError(t, json.Unmarshal(updateRec.Body.Bytes(), &resp))
	assert.Equal(t, float32(200), resp.Amount)
	assert.Equal(t, "electric", resp.Category)
}

func TestUpdateRecurring_AllFields(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	accountID := createTestAccount(t, router, cookie)

	createBody := fmt.Sprintf(`{"account_id":"%s","type":"expense","category":"utilities","amount":150,"frequency":"monthly","next_occurrence":"2026-04-01"}`, accountID)
	createReq := authedRequest(http.MethodPost, "/api/v1/recurring", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.RecurringResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	account2Body := `{"name":"Savings","account_type":"savings","institution":"Other Bank","currency":"USD"}`
	account2Req := authedRequest(http.MethodPost, "/api/v1/accounts", account2Body, cookie)
	account2Rec := httptest.NewRecorder()
	router.ServeHTTP(account2Rec, account2Req)
	require.Equal(t, http.StatusCreated, account2Rec.Code)

	var acct2 api.AccountResponse
	require.NoError(t, json.Unmarshal(account2Rec.Body.Bytes(), &acct2))

	updateBody := fmt.Sprintf(`{"account_id":"%s","type":"deposit","category":"salary","amount":5000,"currency":"EUR","frequency":"weekly","next_occurrence":"2026-05-01","is_active":false}`, acct2.Id)
	updateReq := authedRequest(http.MethodPut, fmt.Sprintf("/api/v1/recurring/%s", created.Id), updateBody, cookie)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	assert.Equal(t, http.StatusOK, updateRec.Code)

	var resp api.RecurringResponse
	require.NoError(t, json.Unmarshal(updateRec.Body.Bytes(), &resp))
	assert.Equal(t, acct2.Id.String(), resp.AccountId.String())
	assert.Equal(t, api.TransactionTypeDeposit, resp.Type)
	assert.Equal(t, "salary", resp.Category)
	assert.Equal(t, float32(5000), resp.Amount)
	assert.Equal(t, api.EUR, resp.Currency)
	assert.Equal(t, api.Weekly, resp.Frequency)
	assert.False(t, resp.IsActive)
}

func TestUpdateRecurring_InvalidJSON_Returns400(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	accountID := createTestAccount(t, router, cookie)

	createBody := fmt.Sprintf(`{"account_id":"%s","type":"expense","category":"utilities","amount":150,"frequency":"monthly","next_occurrence":"2026-04-01"}`, accountID)
	createReq := authedRequest(http.MethodPost, "/api/v1/recurring", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.RecurringResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	updateReq := authedRequest(http.MethodPut, fmt.Sprintf("/api/v1/recurring/%s", created.Id), "not json", cookie)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	assert.Equal(t, http.StatusBadRequest, updateRec.Code)
}

func TestUpdateRecurring_NotFound_Returns404(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"amount":200}`
	req := authedRequest(http.MethodPut, "/api/v1/recurring/00000099-0000-0000-0000-000000000099", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDeleteRecurring_SoftDeletes(t *testing.T) {
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	accountID := createTestAccount(t, router, cookie)

	createBody := fmt.Sprintf(`{"account_id":"%s","type":"expense","category":"utilities","amount":150,"frequency":"monthly","next_occurrence":"2026-04-01"}`, accountID)
	createReq := authedRequest(http.MethodPost, "/api/v1/recurring", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.RecurringResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	deleteReq := authedRequest(http.MethodDelete, fmt.Sprintf("/api/v1/recurring/%s", created.Id), "", cookie)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)

	assert.Equal(t, http.StatusNoContent, deleteRec.Code)

	var deletedAt *string
	err := database.QueryRowContext(context.Background(),
		"SELECT deleted_at FROM recurring_transactions WHERE id = ?",
		created.Id.String(),
	).Scan(&deletedAt)
	require.NoError(t, err)
	assert.NotNil(t, deletedAt)
}

func TestDeleteRecurring_NotFound_Returns404(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodDelete, "/api/v1/recurring/00000099-0000-0000-0000-000000000099", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
