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

const testUserID = "usr00001-0000-0000-0000-000000000001"

func TestListAlerts_Empty(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodGet, "/api/v1/alerts", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AlertListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 0)
	assert.Equal(t, 0, resp.Pagination.Total)
}

func TestListAlerts_ReturnsUnreadOnly(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO alerts (id, user_id, alert_type, message, severity, is_read) VALUES (?, ?, ?, ?, ?, ?)`,
		"c0000001-0000-0000-0000-000000000001", testUserID, "budget_exceeded", "Budget exceeded", "warning", 0,
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO alerts (id, user_id, alert_type, message, severity, is_read) VALUES (?, ?, ?, ?, ?, ?)`,
		"c0000002-0000-0000-0000-000000000002", testUserID, "goal_met", "Goal reached", "info", 1,
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/alerts", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AlertListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "budget_exceeded", resp.Data[0].AlertType)
}

func TestListAlerts_UserScoping(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		"usr00002-0000-0000-0000-000000000002", "alex@example.com", "Alex", "$2a$10$dummy",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO alerts (id, user_id, alert_type, message, severity, is_read) VALUES (?, ?, ?, ?, ?, ?)`,
		"c0000003-0000-0000-0000-000000000003", "usr00002-0000-0000-0000-000000000002", "budget_exceeded", "Alex alert", "warning", 0,
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/alerts", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AlertListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 0)
}

func TestMarkAlertRead_Success(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO alerts (id, user_id, alert_type, message, severity, is_read) VALUES (?, ?, ?, ?, ?, ?)`,
		"c0000004-0000-0000-0000-000000000004", testUserID, "budget_exceeded", "Test alert", "warning", 0,
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodPut, "/api/v1/alerts/c0000004-0000-0000-0000-000000000004/read", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	var isRead bool
	err = database.QueryRowContext(context.Background(),
		"SELECT is_read FROM alerts WHERE id = ?",
		"c0000004-0000-0000-0000-000000000004",
	).Scan(&isRead)
	require.NoError(t, err)
	assert.True(t, isRead)
}

func TestMarkAlertRead_NotFound_Returns404(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodPut, "/api/v1/alerts/00000099-0000-0000-0000-000000000099/read", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListAlerts_WithRelatedEntity(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO alerts (id, user_id, alert_type, message, severity, is_read, related_entity_type, related_entity_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"c0000005-0000-0000-0000-000000000005", testUserID, "budget_exceeded", "Over budget on food", "warning", 0, "transaction", "d0000001-0000-0000-0000-000000000001",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/alerts", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AlertListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	require.NotNil(t, resp.Data[0].RelatedEntityType)
	assert.Equal(t, "transaction", *resp.Data[0].RelatedEntityType)
	require.NotNil(t, resp.Data[0].RelatedEntityId)
	assert.NotEmpty(t, resp.Data[0].RelatedEntityId.String())
}

func TestListAlerts_Pagination(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	for i := 1; i <= 3; i++ {
		_, err := database.ExecContext(context.Background(),
			`INSERT INTO alerts (id, user_id, alert_type, message, severity, is_read) VALUES (?, ?, ?, ?, ?, ?)`,
			fmt.Sprintf("c000000%d-0000-0000-0000-00000000000%d", i+5, i+5), testUserID, "budget_exceeded", fmt.Sprintf("Alert %d", i), "warning", 0,
		)
		require.NoError(t, err)
	}

	req := authedRequest(http.MethodGet, "/api/v1/alerts?page=1&per_page=2", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.AlertListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, 3, resp.Pagination.Total)
	assert.Equal(t, 2, resp.Pagination.TotalPages)
}

func TestListAlerts_NoAuth_Returns401(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
