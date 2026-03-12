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

func TestCreateGoal_Savings(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Emergency Fund","goal_type":"savings","target_amount":10000}`
	req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Emergency Fund", resp.Name)
	assert.Equal(t, api.Savings, resp.GoalType)
	assert.Equal(t, float32(10000), resp.TargetAmount)
	assert.Equal(t, float32(0), resp.CurrentAmount)
	assert.Equal(t, 1, resp.PriorityRank)
	assert.NotEmpty(t, resp.Id)
}

func TestCreateGoal_DebtPayoff(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Student Loans","goal_type":"debt_payoff","target_amount":25000,"current_amount":5000,"deadline":"2027-06-30"}`
	req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Student Loans", resp.Name)
	assert.Equal(t, api.DebtPayoff, resp.GoalType)
	assert.Equal(t, float32(25000), resp.TargetAmount)
	assert.Equal(t, float32(5000), resp.CurrentAmount)
	assert.NotNil(t, resp.Deadline)
	assert.NotNil(t, resp.RequiredMonthlyContribution)
}

func TestCreateGoal_MissingFields_Returns400(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Incomplete"}`
	req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateGoal_InvalidJSON_Returns400(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodPost, "/api/v1/goals", "not json", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateGoal_AutoPriorityRank(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body1 := `{"name":"Goal 1","goal_type":"savings","target_amount":1000}`
	req1 := authedRequest(http.MethodPost, "/api/v1/goals", body1, cookie)
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusCreated, rec1.Code)

	var g1 api.GoalResponse
	require.NoError(t, json.Unmarshal(rec1.Body.Bytes(), &g1))
	assert.Equal(t, 1, g1.PriorityRank)

	body2 := `{"name":"Goal 2","goal_type":"savings","target_amount":2000}`
	req2 := authedRequest(http.MethodPost, "/api/v1/goals", body2, cookie)
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusCreated, rec2.Code)

	var g2 api.GoalResponse
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &g2))
	assert.Equal(t, 2, g2.PriorityRank)
}

func TestListGoals_PriorityOrder(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	goals := []string{
		`{"name":"Third","goal_type":"savings","target_amount":3000,"priority_rank":3}`,
		`{"name":"First","goal_type":"savings","target_amount":1000,"priority_rank":1}`,
		`{"name":"Second","goal_type":"debt_payoff","target_amount":2000,"priority_rank":2}`,
	}
	for _, body := range goals {
		req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
	}

	req := authedRequest(http.MethodGet, "/api/v1/goals", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.GoalListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 3)
	assert.Equal(t, "First", resp.Data[0].Name)
	assert.Equal(t, "Second", resp.Data[1].Name)
	assert.Equal(t, "Third", resp.Data[2].Name)
	assert.Equal(t, 3, resp.Pagination.Total)
}

func TestListGoals_UserScoping(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		"usr00002-0000-0000-0000-000000000002", "alex@example.com", "Alex", "$2a$10$dummy",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"a0000001-0000-0000-0000-000000000001", "usr00001-0000-0000-0000-000000000001", "My Goal", "savings", 5000, 0, 1,
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"a0000002-0000-0000-0000-000000000002", "usr00002-0000-0000-0000-000000000002", "Alex Goal", "savings", 8000, 0, 1,
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/goals", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.GoalListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 1)
	assert.Equal(t, "My Goal", resp.Data[0].Name)
}

func TestListGoals_Pagination(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	for i := 1; i <= 3; i++ {
		body := fmt.Sprintf(`{"name":"Goal %d","goal_type":"savings","target_amount":%d}`, i, i*1000)
		req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
	}

	req := authedRequest(http.MethodGet, "/api/v1/goals?page=1&per_page=2", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.GoalListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, 3, resp.Pagination.Total)
	assert.Equal(t, 2, resp.Pagination.TotalPages)
}

func TestGetGoal_Found(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"My Goal","goal_type":"savings","target_amount":5000}`
	createReq := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.GoalResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	getReq := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/goals/%s", created.Id), "", cookie)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)

	assert.Equal(t, http.StatusOK, getRec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(getRec.Body.Bytes(), &resp))
	assert.Equal(t, "My Goal", resp.Name)
}

func TestGetGoal_NotFound(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodGet, "/api/v1/goals/00000099-0000-0000-0000-000000000099", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetGoal_OtherUser_Returns404(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		"usr00002-0000-0000-0000-000000000002", "alex@example.com", "Alex", "$2a$10$dummy",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"a0000002-0000-0000-0000-000000000002", "usr00002-0000-0000-0000-000000000002", "Alex Goal", "savings", 5000, 0, 1,
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/goals/a0000002-0000-0000-0000-000000000002", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateGoal_ValidRequest(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	createBody := `{"name":"Original","goal_type":"savings","target_amount":5000}`
	createReq := authedRequest(http.MethodPost, "/api/v1/goals", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.GoalResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	updateBody := `{"name":"Updated","current_amount":2500}`
	updateReq := authedRequest(http.MethodPut, fmt.Sprintf("/api/v1/goals/%s", created.Id), updateBody, cookie)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	assert.Equal(t, http.StatusOK, updateRec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(updateRec.Body.Bytes(), &resp))
	assert.Equal(t, "Updated", resp.Name)
	assert.Equal(t, float32(2500), resp.CurrentAmount)
}

func TestUpdateGoal_NotFound_Returns404(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Updated"}`
	req := authedRequest(http.MethodPut, "/api/v1/goals/00000099-0000-0000-0000-000000000099", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestDeleteGoal_SoftDeletes(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	createBody := `{"name":"To Delete","goal_type":"savings","target_amount":5000}`
	createReq := authedRequest(http.MethodPost, "/api/v1/goals", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.GoalResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	deleteReq := authedRequest(http.MethodDelete, fmt.Sprintf("/api/v1/goals/%s", created.Id), "", cookie)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)

	assert.Equal(t, http.StatusNoContent, deleteRec.Code)

	var deletedAt *string
	err := database.QueryRowContext(context.Background(),
		"SELECT deleted_at FROM goals WHERE id = ?",
		created.Id.String(),
	).Scan(&deletedAt)
	require.NoError(t, err)
	assert.NotNil(t, deletedAt)

	getReq := authedRequest(http.MethodGet, fmt.Sprintf("/api/v1/goals/%s", created.Id), "", cookie)
	getRec := httptest.NewRecorder()
	router.ServeHTTP(getRec, getReq)
	assert.Equal(t, http.StatusNotFound, getRec.Code)
}

func TestDeleteGoal_NotFound_Returns404(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodDelete, "/api/v1/goals/00000099-0000-0000-0000-000000000099", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestReorderGoals_Success(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	var ids [3]string
	for i := 1; i <= 3; i++ {
		body := fmt.Sprintf(`{"name":"Goal %d","goal_type":"savings","target_amount":%d}`, i, i*1000)
		req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		var g api.GoalResponse
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &g))
		ids[i-1] = g.Id.String()
	}

	reorderBody := fmt.Sprintf(
		`{"rankings":[{"id":"%s","priority_rank":3},{"id":"%s","priority_rank":1},{"id":"%s","priority_rank":2}]}`,
		ids[0], ids[1], ids[2],
	)
	req := authedRequest(http.MethodPut, "/api/v1/goals/reorder", reorderBody, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.GoalListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 3)
	assert.Equal(t, "Goal 2", resp.Data[0].Name)
	assert.Equal(t, 1, resp.Data[0].PriorityRank)
	assert.Equal(t, "Goal 3", resp.Data[1].Name)
	assert.Equal(t, 2, resp.Data[1].PriorityRank)
	assert.Equal(t, "Goal 1", resp.Data[2].Name)
	assert.Equal(t, 3, resp.Data[2].PriorityRank)
}

func TestReorderGoals_NonexistentGoal_Returns400(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"rankings":[{"id":"00000099-0000-0000-0000-000000000099","priority_rank":1}]}`
	req := authedRequest(http.MethodPut, "/api/v1/goals/reorder", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestReorderGoals_EmptyRankings_Returns400(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"rankings":[]}`
	req := authedRequest(http.MethodPut, "/api/v1/goals/reorder", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGoalComputedFields_GoalMet(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Met Goal","goal_type":"savings","target_amount":5000,"current_amount":5000}`
	req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotNil(t, resp.OnTrack)
	assert.True(t, *resp.OnTrack)
	require.NotNil(t, resp.RequiredMonthlyContribution)
	assert.Equal(t, float32(0), *resp.RequiredMonthlyContribution)
}

func TestGoalComputedFields_WithDeadline(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Future Goal","goal_type":"savings","target_amount":12000,"current_amount":0,"deadline":"2030-01-01"}`
	req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotNil(t, resp.RequiredMonthlyContribution)
	assert.Greater(t, *resp.RequiredMonthlyContribution, float32(0))
}

func TestCreateGoal_AuditLogEntry(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Audit Test","goal_type":"savings","target_amount":5000}`
	req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))

	var count int
	err := database.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM audit_log WHERE entity_type = 'goal' AND entity_id = ? AND action = 'create'",
		resp.Id.String(),
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestUpdateGoal_AuditLogEntry(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	createBody := `{"name":"Audit Update","goal_type":"savings","target_amount":5000}`
	createReq := authedRequest(http.MethodPost, "/api/v1/goals", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.GoalResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	updateBody := `{"current_amount":2000}`
	updateReq := authedRequest(http.MethodPut, fmt.Sprintf("/api/v1/goals/%s", created.Id), updateBody, cookie)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)
	require.Equal(t, http.StatusOK, updateRec.Code)

	var count int
	err := database.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM audit_log WHERE entity_type = 'goal' AND entity_id = ? AND action = 'update'",
		created.Id.String(),
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestDeleteGoal_AuditLogEntry(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	createBody := `{"name":"Audit Delete","goal_type":"savings","target_amount":5000}`
	createReq := authedRequest(http.MethodPost, "/api/v1/goals", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.GoalResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	deleteReq := authedRequest(http.MethodDelete, fmt.Sprintf("/api/v1/goals/%s", created.Id), "", cookie)
	deleteRec := httptest.NewRecorder()
	router.ServeHTTP(deleteRec, deleteReq)
	require.Equal(t, http.StatusNoContent, deleteRec.Code)

	var count int
	err := database.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM audit_log WHERE entity_type = 'goal' AND entity_id = ? AND action = 'delete'",
		created.Id.String(),
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestGoalComputedFields_PastDeadline_NotOnTrack(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Overdue Goal","goal_type":"savings","target_amount":10000,"current_amount":2000,"deadline":"2020-01-01"}`
	req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotNil(t, resp.OnTrack)
	assert.False(t, *resp.OnTrack)
	assert.Nil(t, resp.RequiredMonthlyContribution)
}

func TestGoalComputedFields_NoDeadline_NoProjection(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Open Ended","goal_type":"savings","target_amount":50000,"current_amount":1000}`
	req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Nil(t, resp.OnTrack)
	assert.Nil(t, resp.RequiredMonthlyContribution)
	assert.Nil(t, resp.ProjectedCompletionDate)
}

func TestGoalComputedFields_ExceededTarget(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Over Target","goal_type":"savings","target_amount":5000,"current_amount":7000}`
	req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotNil(t, resp.OnTrack)
	assert.True(t, *resp.OnTrack)
	require.NotNil(t, resp.RequiredMonthlyContribution)
	assert.Equal(t, float32(0), *resp.RequiredMonthlyContribution)
}

func TestGoalComputedFields_FutureDeadlineZeroCurrent(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"Zero Progress","goal_type":"debt_payoff","target_amount":20000,"current_amount":0,"deadline":"2030-06-01"}`
	req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.NotNil(t, resp.RequiredMonthlyContribution)
	assert.Greater(t, *resp.RequiredMonthlyContribution, float32(0))
	require.NotNil(t, resp.OnTrack)
	assert.False(t, *resp.OnTrack)
	assert.Nil(t, resp.ProjectedCompletionDate)
}

func TestUpdateGoal_InvalidJSON_Returns400(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	createBody := `{"name":"For Update","goal_type":"savings","target_amount":5000}`
	createReq := authedRequest(http.MethodPost, "/api/v1/goals", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.GoalResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	updateReq := authedRequest(http.MethodPut, fmt.Sprintf("/api/v1/goals/%s", created.Id), "not json", cookie)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	assert.Equal(t, http.StatusBadRequest, updateRec.Code)
}

func TestReorderGoals_InvalidJSON_Returns400(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	req := authedRequest(http.MethodPut, "/api/v1/goals/reorder", "not json", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateGoal_WithLinkedAccount(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"a1c0a101-0000-0000-0000-000000000001", "usr00001-0000-0000-0000-000000000001", "Mortgage", "BankCo", "mortgage", "USD",
	)
	require.NoError(t, err)

	body := `{"name":"Pay Off Mortgage","goal_type":"debt_payoff","target_amount":200000,"linked_account_id":"a1c0a101-0000-0000-0000-000000000001"}`
	req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "Pay Off Mortgage", resp.Name)
	require.NotNil(t, resp.LinkedAccountId)
}

func TestGoalComputedFields_ProjectedCompletionDate(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank, deadline, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now', '-6 months'))`,
		"a0000010-0000-0000-0000-000000000010", "usr00001-0000-0000-0000-000000000001",
		"Projectable Goal", "savings", 10000, 5000, 1, "2030-01-01",
	)
	require.NoError(t, err)

	req := authedRequest(http.MethodGet, "/api/v1/goals/a0000010-0000-0000-0000-000000000010", "", cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.NotNil(t, resp.ProjectedCompletionDate)
	assert.NotNil(t, resp.OnTrack)
	assert.NotNil(t, resp.RequiredMonthlyContribution)
}

func TestCreateGoal_EmptyName_Returns400(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	body := `{"name":"","goal_type":"savings","target_amount":5000}`
	req := authedRequest(http.MethodPost, "/api/v1/goals", body, cookie)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateGoal_AllFields(t *testing.T) {
	t.Parallel()
	_, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	createBody := `{"name":"Original","goal_type":"savings","target_amount":5000}`
	createReq := authedRequest(http.MethodPost, "/api/v1/goals", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.GoalResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	updateBody := `{"name":"Renamed","goal_type":"debt_payoff","target_amount":10000,"current_amount":3000,"deadline":"2028-12-31","priority_rank":5}`
	updateReq := authedRequest(http.MethodPut, fmt.Sprintf("/api/v1/goals/%s", created.Id), updateBody, cookie)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	assert.Equal(t, http.StatusOK, updateRec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(updateRec.Body.Bytes(), &resp))
	assert.Equal(t, "Renamed", resp.Name)
	assert.Equal(t, api.DebtPayoff, resp.GoalType)
	assert.Equal(t, float32(10000), resp.TargetAmount)
	assert.Equal(t, float32(3000), resp.CurrentAmount)
	assert.NotNil(t, resp.Deadline)
	assert.Equal(t, 5, resp.PriorityRank)
}

func TestUpdateGoal_WithLinkedAccount(t *testing.T) {
	t.Parallel()
	database, router := setupRouter(t)
	cookie := loginAndGetCookie(t, router)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"a1c0a201-0000-0000-0000-000000000001", "usr00001-0000-0000-0000-000000000001", "Linked Savings", "BankCo", "savings", "USD",
	)
	require.NoError(t, err)

	createBody := `{"name":"Link Test","goal_type":"savings","target_amount":5000}`
	createReq := authedRequest(http.MethodPost, "/api/v1/goals", createBody, cookie)
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)
	require.Equal(t, http.StatusCreated, createRec.Code)

	var created api.GoalResponse
	require.NoError(t, json.Unmarshal(createRec.Body.Bytes(), &created))

	updateBody := `{"linked_account_id":"a1c0a201-0000-0000-0000-000000000001"}`
	updateReq := authedRequest(http.MethodPut, fmt.Sprintf("/api/v1/goals/%s", created.Id), updateBody, cookie)
	updateRec := httptest.NewRecorder()
	router.ServeHTTP(updateRec, updateReq)

	assert.Equal(t, http.StatusOK, updateRec.Code)

	var resp api.GoalResponse
	require.NoError(t, json.Unmarshal(updateRec.Body.Bytes(), &resp))
	require.NotNil(t, resp.LinkedAccountId)
}
