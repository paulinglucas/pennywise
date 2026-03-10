package api_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/api"
)

const txnTestUserID = "usr00001-0000-0000-0000-000000000001"
const txnTestAccountID = "a0000001-0000-0000-0000-000000000001"
const txnTestUser2ID = "usr00002-0000-0000-0000-000000000002"
const txnTestAccount2ID = "a0000002-0000-0000-0000-000000000002"

func setupTransactionTests(t *testing.T) (*sql.DB, http.Handler, *http.Cookie) {
	t.Helper()
	database, router := setupRouter(t)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		txnTestAccountID, txnTestUserID, "Test Checking", "Chase", "checking", "USD", 1,
	)
	require.NoError(t, err)

	cookie := loginAndGetCookie(t, router)
	return database, router, cookie
}

func insertTestTransaction(t *testing.T, db *sql.DB, id, userID, accountID string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, userID, accountID, "expense", "food", 42.99, "USD", "2025-06-15",
	)
	require.NoError(t, err)
}

func TestCreateTransaction_ValidRequest(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	body := fmt.Sprintf(`{"type":"expense","category":"food","amount":42.99,"date":"2025-06-15","account_id":"%s","tags":["lunch","work"]}`, txnTestAccountID)
	req := authedRequest(http.MethodPost, "/api/v1/transactions", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.TransactionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, api.TransactionTypeExpense, resp.Type)
	assert.Equal(t, "food", resp.Category)
	assert.InDelta(t, 42.99, float64(resp.Amount), 0.01)
	assert.Equal(t, api.USD, resp.Currency)
	assert.ElementsMatch(t, []string{"lunch", "work"}, resp.Tags)
	assert.NotEmpty(t, resp.Id)
}

func TestCreateTransaction_MissingFields_Returns400(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	body := `{"type":"expense"}`
	req := authedRequest(http.MethodPost, "/api/v1/transactions", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListTransactions_WithPagination(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	for i := range 5 {
		insertTestTransaction(t, database,
			fmt.Sprintf("b000000%d-0000-0000-0000-000000000001", i+1),
			txnTestUserID, txnTestAccountID)
	}

	req := authedRequest(http.MethodGet, "/api/v1/transactions?page=1&per_page=2", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, 5, resp.Pagination.Total)
	assert.Equal(t, 3, resp.Pagination.TotalPages)
}

func TestListTransactions_FilterByCategory(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"c0000001-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID, "expense", "food", 10.00, "USD", "2025-06-15",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"c0000002-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID, "expense", "transport", 20.00, "USD", "2025-06-15",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/transactions?category=food", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "food", resp.Data[0].Category)
}

func TestGetTransaction_ReturnsDetail(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	txnID := "d0000001-0000-0000-0000-000000000001"
	insertTestTransaction(t, database, txnID, txnTestUserID, txnTestAccountID)

	req := authedRequest(http.MethodGet, "/api/v1/transactions/"+txnID, "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "food", resp.Category)
	assert.InDelta(t, 42.99, float64(resp.Amount), 0.01)
}

func TestGetTransaction_OtherUser_Returns404(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		txnTestUser2ID, "alex@example.com", "Alex", "$2a$10$dummy",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		txnTestAccount2ID, txnTestUser2ID, "Alex Account", "BofA", "savings", "USD", 1,
	)
	require.NoError(t, err)

	txnID := "d0000002-0000-0000-0000-000000000001"
	insertTestTransaction(t, database, txnID, txnTestUser2ID, txnTestAccount2ID)

	req := authedRequest(http.MethodGet, "/api/v1/transactions/"+txnID, "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetTransaction_SoftDeleted_Returns404(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	txnID := "d0000003-0000-0000-0000-000000000001"
	_, err := database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date, deleted_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		txnID, txnTestUserID, txnTestAccountID, "expense", "food", 10, "USD", "2025-06-15",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/transactions/"+txnID, "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateTransaction_ValidRequest(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	txnID := "d0000004-0000-0000-0000-000000000001"
	insertTestTransaction(t, database, txnID, txnTestUserID, txnTestAccountID)

	body := `{"category":"dining","amount":35.00,"tags":["updated"]}`
	req := authedRequest(http.MethodPut, "/api/v1/transactions/"+txnID, body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "dining", resp.Category)
	assert.InDelta(t, 35.00, float64(resp.Amount), 0.01)
	assert.ElementsMatch(t, []string{"updated"}, resp.Tags)
}

func TestUpdateTransaction_CreatesAuditLog(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	txnID := "d0000005-0000-0000-0000-000000000001"
	insertTestTransaction(t, database, txnID, txnTestUserID, txnTestAccountID)

	body := `{"category":"dining"}`
	req := authedRequest(http.MethodPut, "/api/v1/transactions/"+txnID, body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var count int
	err := database.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM audit_log WHERE entity_type = 'transaction' AND entity_id = ? AND action = 'update'", txnID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestDeleteTransaction_SoftDeletes(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	txnID := "d0000006-0000-0000-0000-000000000001"
	insertTestTransaction(t, database, txnID, txnTestUserID, txnTestAccountID)

	req := authedRequest(http.MethodDelete, "/api/v1/transactions/"+txnID, "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	var deletedAt *string
	err := database.QueryRowContext(context.Background(),
		"SELECT deleted_at FROM transactions WHERE id = ?", txnID,
	).Scan(&deletedAt)
	require.NoError(t, err)
	assert.NotNil(t, deletedAt)
}

func TestDeleteTransaction_CreatesAuditLog(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	txnID := "d0000007-0000-0000-0000-000000000001"
	insertTestTransaction(t, database, txnID, txnTestUserID, txnTestAccountID)

	req := authedRequest(http.MethodDelete, "/api/v1/transactions/"+txnID, "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	var count int
	err := database.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM audit_log WHERE entity_type = 'transaction' AND entity_id = ? AND action = 'delete'", txnID,
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestImportTransactions_ValidCSV(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	csvContent := "date,type,category,amount,notes\n2025-06-15,expense,food,42.99,Lunch\n2025-06-16,deposit,salary,5000.00,Paycheck\n"

	var buf strings.Builder
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormField("account_id")
	require.NoError(t, err)
	_, err = part.Write([]byte(txnTestAccountID))
	require.NoError(t, err)

	filePart, err := writer.CreateFormFile("file", "transactions.csv")
	require.NoError(t, err)
	_, err = filePart.Write([]byte(csvContent))
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/import", strings.NewReader(buf.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.ImportResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, 2, resp.Imported)
	assert.Empty(t, resp.Errors)
}

func TestImportTransactions_MalformedCSV(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	csvContent := "date,type,category,amount\n2025-06-15,expense,food,42.99\nbad-date,invalid,food,abc\n"

	var buf strings.Builder
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormField("account_id")
	require.NoError(t, err)
	_, err = part.Write([]byte(txnTestAccountID))
	require.NoError(t, err)

	filePart, err := writer.CreateFormFile("file", "transactions.csv")
	require.NoError(t, err)
	_, err = filePart.Write([]byte(csvContent))
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/import", strings.NewReader(buf.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.ImportResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, 1, resp.Imported)
	assert.NotEmpty(t, resp.Errors)
}

func TestListTransactions_UserScoping(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		txnTestUser2ID, "alex@example.com", "Alex", "$2a$10$dummy",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		txnTestAccount2ID, txnTestUser2ID, "Alex Account", "BofA", "savings", "USD", 1,
	)
	require.NoError(t, err)

	insertTestTransaction(t, database, "e0000001-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID)
	insertTestTransaction(t, database, "e0000002-0000-0000-0000-000000000001", txnTestUser2ID, txnTestAccount2ID)

	req := authedRequest(http.MethodGet, "/api/v1/transactions", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, 1, resp.Pagination.Total)
}

func TestCreateTransaction_CreatesAuditLog(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	body := fmt.Sprintf(`{"type":"expense","category":"food","amount":10,"date":"2025-06-15","account_id":"%s"}`, txnTestAccountID)
	req := authedRequest(http.MethodPost, "/api/v1/transactions", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var count int
	err := database.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM audit_log WHERE entity_type = 'transaction' AND action = 'create'",
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestListTransactions_Search(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date, notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"f0000001-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID, "expense", "food", 42, "USD", "2025-06-15", "Chipotle burrito",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date, notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"f0000002-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID, "expense", "transport", 40, "USD", "2025-06-15", "Gas station",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/transactions?search=chipotle", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "food", resp.Data[0].Category)
}

func TestTransactionWithoutAuth_Returns401(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCreateTransaction_InvalidJSON_Returns400(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	req := authedRequest(http.MethodPost, "/api/v1/transactions", "not json", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateTransaction_WithOptionalFields(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	body := fmt.Sprintf(`{"type":"expense","category":"food","amount":25.50,"date":"2025-06-15","account_id":"%s","currency":"EUR","is_recurring":true,"notes":"test note"}`, txnTestAccountID)
	req := authedRequest(http.MethodPost, "/api/v1/transactions", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.TransactionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, api.EUR, resp.Currency)
	assert.True(t, resp.IsRecurring)
	require.NotNil(t, resp.Notes)
	assert.Equal(t, "test note", *resp.Notes)
}

func TestUpdateTransaction_NotFound_Returns404(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	body := `{"category":"dining"}`
	req := authedRequest(http.MethodPut, "/api/v1/transactions/00000000-aaaa-0000-0000-000000000099", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateTransaction_InvalidJSON_Returns400(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	txnID := "d1000001-0000-0000-0000-000000000001"
	insertTestTransaction(t, database, txnID, txnTestUserID, txnTestAccountID)

	req := authedRequest(http.MethodPut, "/api/v1/transactions/"+txnID, "not json", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateTransaction_AllFields(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	txnID := "d1000002-0000-0000-0000-000000000001"
	insertTestTransaction(t, database, txnID, txnTestUserID, txnTestAccountID)

	body := fmt.Sprintf(`{
		"type":"deposit",
		"category":"salary",
		"amount":100.50,
		"currency":"EUR",
		"date":"2025-07-01",
		"notes":"updated note",
		"account_id":"%s",
		"is_recurring":true
	}`, txnTestAccountID)

	req := authedRequest(http.MethodPut, "/api/v1/transactions/"+txnID, body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, api.TransactionTypeDeposit, resp.Type)
	assert.Equal(t, "salary", resp.Category)
	assert.InDelta(t, 100.50, float64(resp.Amount), 0.01)
	assert.Equal(t, api.EUR, resp.Currency)
	assert.Equal(t, "2025-07-01", resp.Date.Format("2006-01-02"))
	require.NotNil(t, resp.Notes)
	assert.Equal(t, "updated note", *resp.Notes)
	assert.True(t, resp.IsRecurring)
}

func TestDeleteTransaction_NotFound_Returns404(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	req := authedRequest(http.MethodDelete, "/api/v1/transactions/00000000-bbbb-0000-0000-000000000099", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestImportTransactions_MissingAccountID(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	csvContent := "date,type,category,amount\n2025-06-15,expense,food,42.99\n"

	var buf strings.Builder
	writer := multipart.NewWriter(&buf)

	filePart, err := writer.CreateFormFile("file", "transactions.csv")
	require.NoError(t, err)
	_, err = filePart.Write([]byte(csvContent))
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/import", strings.NewReader(buf.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestImportTransactions_MissingFile(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	var buf strings.Builder
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormField("account_id")
	require.NoError(t, err)
	_, err = part.Write([]byte(txnTestAccountID))
	require.NoError(t, err)

	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/import", strings.NewReader(buf.String()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListTransactions_FilterByAccountID(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	secondAccountID := "a1000001-0000-0000-0000-000000000001"
	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		secondAccountID, txnTestUserID, "Savings", "Chase", "savings", "USD", 1,
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"f1000001-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID, "expense", "food", 10, "USD", "2025-06-15",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"f1000002-0000-0000-0000-000000000001", txnTestUserID, secondAccountID, "expense", "food", 20, "USD", "2025-06-15",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/transactions?account_id="+txnTestAccountID, "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, txnTestAccountID, resp.Data[0].AccountId.String())
}

func TestListTransactions_FilterByType(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"f2000001-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID, "expense", "food", 10, "USD", "2025-06-15",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"f2000002-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID, "deposit", "salary", 5000, "USD", "2025-06-15",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/transactions?type=deposit", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, api.TransactionTypeDeposit, resp.Data[0].Type)
}

func TestListTransactions_FilterByDateRange(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"f3000001-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID, "expense", "food", 10, "USD", "2025-01-15",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"f3000002-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID, "expense", "food", 20, "USD", "2025-06-15",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"f3000003-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID, "expense", "food", 30, "USD", "2025-12-15",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/transactions?date_from=2025-03-01&date_to=2025-09-01", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.InDelta(t, 20, float64(resp.Data[0].Amount), 0.01)
}

func TestListTransactions_FilterByAmountRange(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"f4000001-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID, "expense", "food", 5, "USD", "2025-06-15",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"f4000002-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID, "expense", "food", 50, "USD", "2025-06-15",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"f4000003-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID, "expense", "food", 500, "USD", "2025-06-15",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/transactions?amount_min=10&amount_max=100", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.InDelta(t, 50, float64(resp.Data[0].Amount), 0.01)
}

func TestListTransactions_FilterByTags(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"f5000001-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID, "expense", "food", 10, "USD", "2025-06-15",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO transaction_tags (id, transaction_id, tag) VALUES (?, ?, ?)`,
		"tag00001-0000-0000-0000-000000000001", "f5000001-0000-0000-0000-000000000001", "lunch",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"f5000002-0000-0000-0000-000000000001", txnTestUserID, txnTestAccountID, "expense", "transport", 20, "USD", "2025-06-15",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO transaction_tags (id, transaction_id, tag) VALUES (?, ?, ?)`,
		"tag00002-0000-0000-0000-000000000001", "f5000002-0000-0000-0000-000000000001", "commute",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/transactions?tags=lunch", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "food", resp.Data[0].Category)
}

func TestGetTransaction_ReturnsRecurringTransactionId(t *testing.T) {
	t.Parallel()
	database, router, cookie := setupTransactionTests(t)

	recurringID := "a0000099-0000-0000-0000-000000000001"
	txnID := "f6000001-0000-0000-0000-000000000001"

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO recurring_transactions (id, user_id, account_id, type, category, amount, currency, frequency, next_occurrence) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		recurringID, txnTestUserID, txnTestAccountID, "expense", "food", 42.99, "USD", "monthly", "2025-07-15",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date, recurring_transaction_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		txnID, txnTestUserID, txnTestAccountID, "expense", "food", 42.99, "USD", "2025-06-15", recurringID,
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/transactions/"+txnID, "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.TransactionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotNil(t, resp.RecurringTransactionId)
	assert.Equal(t, recurringID, resp.RecurringTransactionId.String())
}

func TestListCategories_ReturnsDistinctCategories(t *testing.T) {
	t.Parallel()
	db, router, cookie := setupTransactionTests(t)

	_, err := db.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn-cat-1", txnTestUserID, txnTestAccountID, "expense", "food", 10, "USD", "2026-01-01",
	)
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(),
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn-cat-2", txnTestUserID, txnTestAccountID, "expense", "rent", 1000, "USD", "2026-01-01",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/categories", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var catResp api.CategoriesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &catResp))
	assert.Equal(t, []string{"food", "rent"}, catResp.Categories)
}

func TestListCategories_Empty(t *testing.T) {
	t.Parallel()
	_, router, cookie := setupTransactionTests(t)

	req := authedRequest(http.MethodGet, "/api/v1/categories", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var catResp api.CategoriesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &catResp))
	assert.Empty(t, catResp.Categories)
}
