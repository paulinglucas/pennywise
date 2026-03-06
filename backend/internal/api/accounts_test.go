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

func loginAndGetCookie(t *testing.T, router http.Handler) *http.Cookie {
	t.Helper()
	body := `{"email":"james@example.com","password":"pennywise"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	cookies := rec.Result().Cookies()
	require.NotEmpty(t, cookies)
	return cookies[0]
}

func authedRequest(method, path string, body string, cookie *http.Cookie) *http.Request {
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req.AddCookie(cookie)
	return req
}

func TestCreateAccount_ValidRequest(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"My Checking","institution":"Chase","account_type":"checking"}`
	req := authedRequest(http.MethodPost, "/api/v1/accounts", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.AccountResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "My Checking", resp.Name)
	assert.Equal(t, "Chase", resp.Institution)
	assert.Equal(t, api.AccountTypeChecking, resp.AccountType)
	assert.Equal(t, api.USD, resp.Currency)
	assert.True(t, resp.IsActive)
	assert.NotEmpty(t, resp.Id)
}

func TestCreateAccount_MissingFields_Returns400(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Missing Type"}`
	req := authedRequest(http.MethodPost, "/api/v1/accounts", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListAccounts_ReturnsOnlyCurrentUserAccounts(t *testing.T) {
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		"usr00002-0000-0000-0000-000000000002", "alex@example.com", "Alex", "$2a$10$dummy",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"acc00001-0000-0000-0000-000000000001", "usr00001-0000-0000-0000-000000000001", "James Checking", "Chase", "checking", "USD", 1,
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"acc00002-0000-0000-0000-000000000002", "usr00002-0000-0000-0000-000000000002", "Alex Savings", "BofA", "savings", "USD", 1,
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/accounts", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AccountListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "James Checking", resp.Data[0].Name)
	assert.Equal(t, 1, resp.Pagination.Total)
}

func TestGetAccount_ReturnsDetail(t *testing.T) {
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"acc00001-0000-0000-0000-000000000001", "usr00001-0000-0000-0000-000000000001", "My Checking", "Chase", "checking", "USD", 1,
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/accounts/acc00001-0000-0000-0000-000000000001", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AccountResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "My Checking", resp.Name)
}

func TestGetAccount_OtherUser_Returns404(t *testing.T) {
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		"usr00002-0000-0000-0000-000000000002", "alex@example.com", "Alex", "$2a$10$dummy",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"acc00002-0000-0000-0000-000000000002", "usr00002-0000-0000-0000-000000000002", "Alex Account", "BofA", "savings", "USD", 1,
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/accounts/acc00002-0000-0000-0000-000000000002", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetAccount_SoftDeleted_Returns404(t *testing.T) {
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		"acc00001-0000-0000-0000-000000000001", "usr00001-0000-0000-0000-000000000001", "Deleted Account", "Chase", "checking", "USD", 1,
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/accounts/acc00001-0000-0000-0000-000000000001", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateAccount_ValidRequest(t *testing.T) {
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"acc00001-0000-0000-0000-000000000001", "usr00001-0000-0000-0000-000000000001", "Old Name", "Old Bank", "checking", "USD", 1,
	)
	require.NoError(t, err)

	body := `{"name":"New Name","institution":"New Bank"}`
	req := authedRequest(http.MethodPut, "/api/v1/accounts/acc00001-0000-0000-0000-000000000001", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AccountResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "New Name", resp.Name)
	assert.Equal(t, "New Bank", resp.Institution)
	assert.Equal(t, api.AccountTypeChecking, resp.AccountType)
}

func TestDeleteAccount_SoftDeletes(t *testing.T) {
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"acc00001-0000-0000-0000-000000000001", "usr00001-0000-0000-0000-000000000001", "To Delete", "Chase", "checking", "USD", 1,
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodDelete, "/api/v1/accounts/acc00001-0000-0000-0000-000000000001", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	var deletedAt *string
	err = database.QueryRowContext(context.Background(),
		"SELECT deleted_at FROM accounts WHERE id = ?",
		"acc00001-0000-0000-0000-000000000001",
	).Scan(&deletedAt)
	require.NoError(t, err)
	assert.NotNil(t, deletedAt)

	getReq := authedRequest(http.MethodGet, "/api/v1/accounts/acc00001-0000-0000-0000-000000000001", "", cookie)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	assert.Equal(t, http.StatusNotFound, getRec.Code)
}

func TestCreateAccount_InvalidJSON_Returns400(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodPost, "/api/v1/accounts", "not json", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp api.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, api.VALIDATIONFAILED, resp.Error.Code)
}

func TestCreateAccount_WithCurrency(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Euro Account","institution":"Deutsche Bank","account_type":"checking","currency":"EUR"}`
	req := authedRequest(http.MethodPost, "/api/v1/accounts", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.AccountResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Euro Account", resp.Name)
	assert.Equal(t, api.EUR, resp.Currency)
}

func TestUpdateAccount_NotFound_Returns404(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Updated"}`
	req := authedRequest(http.MethodPut, "/api/v1/accounts/acc00099-0000-0000-0000-000000000099", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp api.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, api.NOTFOUND, resp.Error.Code)
}

func TestUpdateAccount_InvalidJSON_Returns400(t *testing.T) {
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"acc00010-0000-0000-0000-000000000010", "usr00001-0000-0000-0000-000000000001", "Test Account", "Chase", "checking", "USD", 1,
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodPut, "/api/v1/accounts/acc00010-0000-0000-0000-000000000010", "not json", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp api.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, api.VALIDATIONFAILED, resp.Error.Code)
}

func TestUpdateAccount_AllFields(t *testing.T) {
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"acc00011-0000-0000-0000-000000000011", "usr00001-0000-0000-0000-000000000001", "Original", "Old Bank", "checking", "USD", 1,
	)
	require.NoError(t, err)

	body := `{"name":"Updated Name","institution":"New Bank","account_type":"savings","currency":"EUR","is_active":false}`
	req := authedRequest(http.MethodPut, "/api/v1/accounts/acc00011-0000-0000-0000-000000000011", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AccountResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Updated Name", resp.Name)
	assert.Equal(t, "New Bank", resp.Institution)
	assert.Equal(t, api.AccountTypeSavings, resp.AccountType)
	assert.Equal(t, api.EUR, resp.Currency)
	assert.False(t, resp.IsActive)
}

func TestDeleteAccount_NotFound_Returns404(t *testing.T) {
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodDelete, "/api/v1/accounts/acc00099-0000-0000-0000-000000000099", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var resp api.ErrorResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, api.NOTFOUND, resp.Error.Code)
}

func TestListAccounts_Pagination(t *testing.T) {
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("acc00020-0000-0000-0000-0000000000%02d", i)
		name := fmt.Sprintf("Account %d", i)
		_, err := database.ExecContext(context.Background(),
			`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			id, "usr00001-0000-0000-0000-000000000001", name, "Chase", "checking", "USD", 1,
		)
		require.NoError(t, err)
	}

	req := authedRequest(http.MethodGet, "/api/v1/accounts?page=1&per_page=2", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AccountListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, 3, resp.Pagination.Total)
	assert.Equal(t, 2, resp.Pagination.TotalPages)
	assert.Equal(t, 1, resp.Pagination.Page)
	assert.Equal(t, 2, resp.Pagination.PerPage)
}
