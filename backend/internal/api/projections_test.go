package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/api"
)

func TestComputeProjection_BasicScenarios(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	setupDashboardData(t, router, cookie)

	body := `{"years_to_project":10}`
	req := authedRequest(http.MethodPost, "/api/v1/projections", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.ProjectionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Scenarios, 3)

	scenarioNames := make(map[api.Scenario]bool)
	for _, s := range resp.Scenarios {
		scenarioNames[s.Scenario] = true
		assert.NotEmpty(t, s.DataPoints)
	}
	assert.True(t, scenarioNames[api.Best])
	assert.True(t, scenarioNames[api.Average])
	assert.True(t, scenarioNames[api.Worst])
}

func TestComputeProjection_BestGrowsFastest(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	setupDashboardData(t, router, cookie)

	body := `{"years_to_project":10}`
	req := authedRequest(http.MethodPost, "/api/v1/projections", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp api.ProjectionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	finalValues := make(map[api.Scenario]float32)
	for _, s := range resp.Scenarios {
		lastPoint := s.DataPoints[len(s.DataPoints)-1]
		finalValues[s.Scenario] = lastPoint.Value
	}

	assert.Greater(t, finalValues[api.Best], finalValues[api.Average])
	assert.Greater(t, finalValues[api.Average], finalValues[api.Worst])
}

func TestComputeProjection_WithCustomReturnRate(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	setupDashboardData(t, router, cookie)

	body := `{"years_to_project":5,"return_rate":12}`
	req := authedRequest(http.MethodPost, "/api/v1/projections", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.ProjectionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Scenarios, 3)
}

func TestComputeProjection_WithOneTimeEvents(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	setupDashboardData(t, router, cookie)

	body := `{"years_to_project":5,"one_time_events":[{"amount":50000,"date":"2027-06-01","type":"windfall"},{"amount":10000,"date":"2028-01-01","type":"expense"}]}`
	req := authedRequest(http.MethodPost, "/api/v1/projections", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.ProjectionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Scenarios, 3)
}

func TestComputeProjection_StartValueMatchesNetWorth(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	setupDashboardData(t, router, cookie)

	dashReq := authedRequest(http.MethodGet, "/api/v1/dashboard", "", cookie)
	dashRec := httptest.NewRecorder()
	router.ServeHTTP(dashRec, dashReq)
	var dashResp api.DashboardResponse
	require.NoError(t, json.Unmarshal(dashRec.Body.Bytes(), &dashResp))

	body := `{"years_to_project":1}`
	req := authedRequest(http.MethodPost, "/api/v1/projections", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp api.ProjectionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	for _, scenario := range resp.Scenarios {
		firstValue := scenario.DataPoints[0].Value
		assert.InDelta(t, dashResp.NetWorth, firstValue, 1.0,
			"scenario %s should start at current net worth", scenario.Scenario)
	}
}

func TestComputeProjection_DebtPayoffFreedCashFlow(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	setupDashboardData(t, router, cookie)

	body := `{"years_to_project":20}`
	req := authedRequest(http.MethodPost, "/api/v1/projections", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.ProjectionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	for _, s := range resp.Scenarios {
		lastValue := s.DataPoints[len(s.DataPoints)-1].Value
		firstValue := s.DataPoints[0].Value
		assert.Greater(t, lastValue, firstValue, "scenario %s should grow over 20 years", s.Scenario)
	}
}

func TestComputeProjection_NoAuth_Returns401(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	body := `{"years_to_project":10}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projections", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestComputeProjection_InvalidBody_Returns400(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `not json`
	req := authedRequest(http.MethodPost, "/api/v1/projections", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestComputeProjection_MissingRequiredField_Returns400(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{}`
	req := authedRequest(http.MethodPost, "/api/v1/projections", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestComputeProjection_WithMonthlySavingsAdjustment(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	setupDashboardData(t, router, cookie)

	body := `{"years_to_project":5,"monthly_savings_adjustment":50}`
	req := authedRequest(http.MethodPost, "/api/v1/projections", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.ProjectionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Scenarios, 3)
}

func TestComputeProjection_NegativeCashFlow(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	accountID := createTestAccount(t, router, cookie)

	txnBody := fmt.Sprintf(`{"account_id":"%s","type":"expense","category":"rent","amount":5000,"date":"2026-03-01"}`, accountID)
	txnReq := authedRequest(http.MethodPost, "/api/v1/transactions", txnBody, cookie)
	txnRec := httptest.NewRecorder()
	router.ServeHTTP(txnRec, txnReq)
	require.Equal(t, http.StatusCreated, txnRec.Code)

	body := `{"years_to_project":5}`
	req := authedRequest(http.MethodPost, "/api/v1/projections", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.ProjectionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Scenarios, 3)
}

func TestComputeProjection_NegativeMonthlySavingsAdjustment(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)
	setupDashboardData(t, router, cookie)

	body := `{"years_to_project":5,"monthly_savings_adjustment":-30}`
	req := authedRequest(http.MethodPost, "/api/v1/projections", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.ProjectionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Scenarios, 3)
}

func TestComputeProjection_MillionaireDateSet(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	accountID := createTestAccount(t, router, cookie)

	assetBody := fmt.Sprintf(`{"name":"Big Portfolio","asset_type":"brokerage","current_value":900000,"currency":"USD","account_id":"%s"}`, accountID)
	assetReq := authedRequest(http.MethodPost, "/api/v1/assets", assetBody, cookie)
	assetRec := httptest.NewRecorder()
	router.ServeHTTP(assetRec, assetReq)
	require.Equal(t, http.StatusCreated, assetRec.Code)

	body := `{"years_to_project":10}`
	req := authedRequest(http.MethodPost, "/api/v1/projections", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp api.ProjectionResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	hasMillionaireDate := false
	for _, s := range resp.Scenarios {
		if s.MillionaireDate != nil {
			hasMillionaireDate = true
		}
	}
	assert.True(t, hasMillionaireDate, "at least one scenario should have a millionaire date")
}
