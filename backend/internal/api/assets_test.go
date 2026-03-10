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

const assetUserID = "usr00001-0000-0000-0000-000000000001"

func TestCreateAsset_ValidRequest(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Roth IRA","asset_type":"retirement","current_value":50000}`
	req := authedRequest(http.MethodPost, "/api/v1/assets", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.AssetResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Roth IRA", resp.Name)
	assert.Equal(t, api.AssetTypeRetirement, resp.AssetType)
	assert.Equal(t, float32(50000), resp.CurrentValue)
	assert.Equal(t, api.USD, resp.Currency)
	assert.NotEmpty(t, resp.Id)
}

func TestCreateAsset_WithMetadata(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Bitcoin Wallet","asset_type":"speculative","current_value":25000,"metadata":{"wallet_addresses":[{"address":"bc1q...","chain":"bitcoin"}]}}`
	req := authedRequest(http.MethodPost, "/api/v1/assets", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.AssetResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Bitcoin Wallet", resp.Name)
	assert.NotNil(t, resp.Metadata)
	meta := *resp.Metadata
	assert.Contains(t, meta, "wallet_addresses")
}

func TestCreateAsset_MissingFields_Returns400(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Missing Type"}`
	req := authedRequest(http.MethodPost, "/api/v1/assets", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListAssets_ReturnsOnlyCurrentUserAssets(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		"usr00002-0000-0000-0000-000000000002", "alex@example.com", "Alex", "$2a$10$dummy",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"a0000001-0000-0000-0000-000000000001", assetUserID, "My Roth IRA", "retirement", 50000, "USD",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"a0000002-0000-0000-0000-000000000002", "usr00002-0000-0000-0000-000000000002", "Alex House", "real_estate", 300000, "USD",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/assets", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AssetListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "My Roth IRA", resp.Data[0].Name)
	assert.Equal(t, 1, resp.Pagination.Total)
	assert.Equal(t, float32(50000), resp.Summary.TotalValue)
	assert.Len(t, resp.Summary.Allocation, 1)
}

func TestGetAsset_ReturnsDetailWithHistory(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Test Asset","asset_type":"liquid","current_value":10000}`
	createReq := authedRequest(http.MethodPost, "/api/v1/assets", body, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.AssetResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	getReq := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/assets/%s", created.Id), "", cookie)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	assert.Equal(t, http.StatusOK, getRec.Code)

	var resp api.AssetResponse
	require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &resp))
	assert.Equal(t, "Test Asset", resp.Name)
	assert.NotNil(t, resp.History)
	assert.Len(t, *resp.History, 1)
}

func TestGetAsset_OtherUser_Returns404(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		"usr00002-0000-0000-0000-000000000002", "alex@example.com", "Alex", "$2a$10$dummy",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"a0000002-0000-0000-0000-000000000002", "usr00002-0000-0000-0000-000000000002", "Alex Asset", "liquid", 5000, "USD",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/assets/a0000002-0000-0000-0000-000000000002", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetAsset_SoftDeleted_Returns404(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency, deleted_at) VALUES (?, ?, ?, ?, ?, ?, datetime('now'))`,
		"a0000001-0000-0000-0000-000000000001", assetUserID, "Deleted Asset", "liquid", 1000, "USD",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/assets/a0000001-0000-0000-0000-000000000001", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateAsset_ValidRequest(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	createBody := `{"name":"Original","asset_type":"liquid","current_value":10000}`
	createReq := authedRequest(http.MethodPost, "/api/v1/assets", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.AssetResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	updateBody := `{"name":"Updated","current_value":15000}`
	updateReq := authedRequest(http.MethodPut, fmt.Sprintf("/api/v1/assets/%s", created.Id), updateBody, cookie)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	assert.Equal(t, http.StatusOK, updateRec.Code)

	var resp api.AssetResponse
	require.NoError(t, json.Unmarshal(updateRec.Body.Bytes(), &resp))
	assert.Equal(t, "Updated", resp.Name)
	assert.Equal(t, float32(15000), resp.CurrentValue)
}

func TestUpdateAsset_CreatesHistoryOnValueChange(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	createBody := `{"name":"History Test","asset_type":"liquid","current_value":10000}`
	createReq := authedRequest(http.MethodPost, "/api/v1/assets", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.AssetResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	updateBody := `{"current_value":12000}`
	updateReq := authedRequest(http.MethodPut, fmt.Sprintf("/api/v1/assets/%s", created.Id), updateBody, cookie)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)
	require.Equal(t, http.StatusOK, updateRec.Code)

	historyReq := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/assets/%s/history", created.Id), "", cookie)
	historyRec := httptest.NewRecorder()
	router.ServeHTTP(historyRec, historyReq)

	assert.Equal(t, http.StatusOK, historyRec.Code)

	var histResp api.AssetHistoryResponse
	require.NoError(t, json.Unmarshal(historyRec.Body.Bytes(), &histResp))
	assert.Len(t, histResp.Entries, 2)
	assert.Equal(t, float32(10000), histResp.Entries[0].Value)
	assert.Equal(t, float32(12000), histResp.Entries[1].Value)
}

func TestUpdateAsset_NotFound_Returns404(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Updated"}`
	req := authedRequest(http.MethodPut, "/api/v1/assets/00000099-0000-0000-0000-000000000099", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDeleteAsset_SoftDeletes(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	createBody := `{"name":"To Delete","asset_type":"liquid","current_value":5000}`
	createReq := authedRequest(http.MethodPost, "/api/v1/assets", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.AssetResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	deleteReq := authedRequest(http.MethodDelete, fmt.Sprintf("/api/v1/assets/%s", created.Id), "", cookie)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)

	assert.Equal(t, http.StatusNoContent, deleteRec.Code)

	var deletedAt *string
	err := database.QueryRowContext(context.Background(),
		"SELECT deleted_at FROM assets WHERE id = ?",
		created.Id.String(),
	).Scan(&deletedAt)
	require.NoError(t, err)
	assert.NotNil(t, deletedAt)

	getReq := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/assets/%s", created.Id), "", cookie)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	assert.Equal(t, http.StatusNotFound, getRec.Code)
}

func TestDeleteAsset_NotFound_Returns404(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodDelete, "/api/v1/assets/00000099-0000-0000-0000-000000000099", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetAssetHistory_NotFound_Returns404(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodGet, "/api/v1/assets/00000099-0000-0000-0000-000000000099/history", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetAssetAllocation(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	assets := []struct {
		name      string
		assetType string
		value     float64
	}{
		{"Checking", "liquid", 10000},
		{"Roth IRA", "retirement", 40000},
		{"House", "real_estate", 50000},
	}

	for _, a := range assets {
		body := fmt.Sprintf(`{"name":"%s","asset_type":"%s","current_value":%f}`, a.name, a.assetType, a.value)
		createReq := authedRequest(http.MethodPost, "/api/v1/assets", body, cookie)
		createRec := httptest.NewRecorder()
		router.ServeHTTP(createRec, createReq)
		require.Equal(t, http.StatusCreated, createRec.Code)
	}

	req := authedRequest(http.MethodGet, "/api/v1/assets/allocation", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AllocationResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.Snapshots)
}

func TestListAssets_PortfolioSummary(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	assets := []struct {
		name      string
		assetType string
		value     float64
	}{
		{"Checking", "liquid", 10000},
		{"Savings", "liquid", 20000},
		{"401k", "retirement", 70000},
	}

	for _, a := range assets {
		body := fmt.Sprintf(`{"name":"%s","asset_type":"%s","current_value":%f}`, a.name, a.assetType, a.value)
		createReq := authedRequest(http.MethodPost, "/api/v1/assets", body, cookie)
		createRec := httptest.NewRecorder()
		router.ServeHTTP(createRec, createReq)
		require.Equal(t, http.StatusCreated, createRec.Code)
	}

	req := authedRequest(http.MethodGet, "/api/v1/assets", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AssetListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, float32(100000), resp.Summary.TotalValue)
	assert.Len(t, resp.Summary.Allocation, 2)

	var pctSum float32
	for _, a := range resp.Summary.Allocation {
		pctSum += a.Percentage
	}
	assert.InDelta(t, 100.0, float64(pctSum), 0.1)
}

func TestCreateAsset_AuditLogEntry(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Audit Test","asset_type":"liquid","current_value":5000}`
	req := authedRequest(http.MethodPost, "/api/v1/assets", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp api.AssetResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	var count int
	err := database.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM audit_log WHERE entity_type = 'asset' AND entity_id = ? AND action = 'create'",
		resp.Id.String(),
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestUpdateAsset_AuditLogEntry(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	createBody := `{"name":"Audit Update","asset_type":"liquid","current_value":5000}`
	createReq := authedRequest(http.MethodPost, "/api/v1/assets", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.AssetResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	updateBody := `{"current_value":7000}`
	updateReq := authedRequest(http.MethodPut, fmt.Sprintf("/api/v1/assets/%s", created.Id), updateBody, cookie)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)
	require.Equal(t, http.StatusOK, updateRec.Code)

	var count int
	err := database.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM audit_log WHERE entity_type = 'asset' AND entity_id = ? AND action = 'update'",
		created.Id.String(),
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestDeleteAsset_AuditLogEntry(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	createBody := `{"name":"Audit Delete","asset_type":"liquid","current_value":5000}`
	createReq := authedRequest(http.MethodPost, "/api/v1/assets", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.AssetResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	deleteReq := authedRequest(http.MethodDelete, fmt.Sprintf("/api/v1/assets/%s", created.Id), "", cookie)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)
	require.Equal(t, http.StatusNoContent, deleteRec.Code)

	var count int
	err := database.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM audit_log WHERE entity_type = 'asset' AND entity_id = ? AND action = 'delete'",
		created.Id.String(),
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCreateAsset_InvalidJSON_Returns400(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodPost, "/api/v1/assets", "not json", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestListAssets_Pagination(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	for i := 1; i <= 3; i++ {
		body := fmt.Sprintf(`{"name":"Asset %d","asset_type":"liquid","current_value":%d}`, i, i*1000)
		createReq := authedRequest(http.MethodPost, "/api/v1/assets", body, cookie)
		createRec := httptest.NewRecorder()
		router.ServeHTTP(createRec, createReq)
		require.Equal(t, http.StatusCreated, createRec.Code)
	}

	req := authedRequest(http.MethodGet, "/api/v1/assets?page=1&per_page=2", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AssetListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, 3, resp.Pagination.Total)
	assert.Equal(t, 2, resp.Pagination.TotalPages)
}

func TestListAssets_LinkedAccountIncluded(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	mortgageAccountID := "accf0001-0000-0000-0000-000000000001"
	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, original_balance) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		mortgageAccountID, assetUserID, "Rocket Mortgage", "Rocket Mortgage", "mortgage", "USD", 300000.0,
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank, linked_account_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"g1f00001-0000-0000-0000-000000000001", assetUserID, "Pay off mortgage", "debt_payoff", 300000, 280000, 1, mortgageAccountID,
	)
	require.NoError(t, err)

	body := fmt.Sprintf(`{"name":"Home","asset_type":"real_estate","current_value":350000,"account_id":"%s"}`, mortgageAccountID)
	createReq := authedRequest(http.MethodPost, "/api/v1/assets", body, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	req := authedRequest(http.MethodGet, "/api/v1/assets", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AssetListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)

	asset := resp.Data[0]
	require.NotNil(t, asset.LinkedAccount)
	assert.Equal(t, "Rocket Mortgage", asset.LinkedAccount.Name)
	assert.Equal(t, api.AccountTypeMortgage, asset.LinkedAccount.AccountType)
	require.NotNil(t, asset.LinkedAccount.Balance)
	assert.Equal(t, float32(280000), *asset.LinkedAccount.Balance)
	require.NotNil(t, asset.LinkedAccount.Institution)
	assert.Equal(t, "Rocket Mortgage", *asset.LinkedAccount.Institution)
}

func TestGetAsset_LinkedAccountIncluded(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	accountID := "accf0002-0000-0000-0000-000000000001"
	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		accountID, assetUserID, "Fidelity Roth IRA", "Fidelity", "retirement_roth_ira", "USD",
	)
	require.NoError(t, err)

	body := fmt.Sprintf(`{"name":"Roth IRA","asset_type":"retirement","current_value":50000,"account_id":"%s"}`, accountID)
	createReq := authedRequest(http.MethodPost, "/api/v1/assets", body, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.AssetResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	getReq := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/assets/%s", created.Id), "", cookie)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	assert.Equal(t, http.StatusOK, getRec.Code)

	var resp api.AssetResponse
	require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &resp))
	require.NotNil(t, resp.LinkedAccount)
	assert.Equal(t, "Fidelity Roth IRA", resp.LinkedAccount.Name)
	assert.Equal(t, api.AccountTypeRetirementRothIra, resp.LinkedAccount.AccountType)
	assert.Nil(t, resp.LinkedAccount.Balance)
}

func TestListAssets_NoLinkedAccountWhenNotLinked(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Standalone","asset_type":"liquid","current_value":5000}`
	createReq := authedRequest(http.MethodPost, "/api/v1/assets", body, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	req := authedRequest(http.MethodGet, "/api/v1/assets", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp api.AssetListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	assert.Nil(t, resp.Data[0].LinkedAccount)
}
