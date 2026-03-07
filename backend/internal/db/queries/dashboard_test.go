package queries_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/db/queries"
)

const dashboardUserID = "usr00001-0000-0000-0000-000000000001"

func setupDashboardTestDB(t *testing.T) (*queries.DashboardRepository, func(sql string, args ...interface{})) {
	t.Helper()
	database := setupTestDB(t)
	repo := queries.NewDashboardRepository(database)

	exec := func(sql string, args ...interface{}) {
		_, err := database.ExecContext(context.Background(), sql, args...)
		require.NoError(t, err)
	}

	exec(`INSERT INTO accounts (id, user_id, name, institution, account_type, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"acct0001-0000-0000-0000-000000000001", dashboardUserID, "Checking", "Bank", "checking", "USD")

	return repo, exec
}

func TestGetNetWorth_NoData(t *testing.T) {
	repo, _ := setupDashboardTestDB(t)

	result, err := repo.GetNetWorth(context.Background(), dashboardUserID)
	require.NoError(t, err)
	assert.Equal(t, 0.0, result.AssetTotal)
	assert.Equal(t, 0.0, result.DebtTotal)
}

func TestGetNetWorth_WithAssets(t *testing.T) {
	repo, exec := setupDashboardTestDB(t)

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "Brokerage", "brokerage", 50000, "USD")
	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00002-0000-0000-0000-000000000002", dashboardUserID, "401k", "retirement_401k", 30000, "USD")

	result, err := repo.GetNetWorth(context.Background(), dashboardUserID)
	require.NoError(t, err)
	assert.Equal(t, 80000.0, result.AssetTotal)
}

func TestGetNetWorth_WithDebts(t *testing.T) {
	repo, exec := setupDashboardTestDB(t)

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "Brokerage", "brokerage", 100000, "USD")

	exec(`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"goal0001-0000-0000-0000-000000000001", dashboardUserID, "Mortgage Payoff", "debt_payoff", 200000, 150000, 1)

	result, err := repo.GetNetWorth(context.Background(), dashboardUserID)
	require.NoError(t, err)
	assert.Equal(t, 100000.0, result.AssetTotal)
	assert.Equal(t, 150000.0, result.DebtTotal)
}

func TestGetNetWorth_ExcludesSoftDeleted(t *testing.T) {
	repo, exec := setupDashboardTestDB(t)

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "Active", "brokerage", 50000, "USD")
	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency, deleted_at) VALUES (?, ?, ?, ?, ?, ?, datetime('now'))`,
		"ast00002-0000-0000-0000-000000000002", dashboardUserID, "Deleted", "brokerage", 30000, "USD")

	result, err := repo.GetNetWorth(context.Background(), dashboardUserID)
	require.NoError(t, err)
	assert.Equal(t, 50000.0, result.AssetTotal)
}

func TestGetCashFlowThisMonth(t *testing.T) {
	repo, exec := setupDashboardTestDB(t)
	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00001-0000-0000-0000-000000000001", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"deposit", "salary", 5000, "USD", "2026-03-01")
	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00002-0000-0000-0000-000000000002", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"expense", "food", 500, "USD", "2026-03-05")
	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00003-0000-0000-0000-000000000003", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"expense", "rent", 1500, "USD", "2026-03-01")

	cashFlow, err := repo.GetCashFlowThisMonth(context.Background(), dashboardUserID, now)
	require.NoError(t, err)
	assert.Equal(t, 3000.0, cashFlow)
}

func TestGetCashFlowThisMonth_ExcludesPreviousMonth(t *testing.T) {
	repo, exec := setupDashboardTestDB(t)
	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00001-0000-0000-0000-000000000001", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"deposit", "salary", 5000, "USD", "2026-02-28")

	cashFlow, err := repo.GetCashFlowThisMonth(context.Background(), dashboardUserID, now)
	require.NoError(t, err)
	assert.Equal(t, 0.0, cashFlow)
}

func TestGetSpendingByCategory(t *testing.T) {
	repo, exec := setupDashboardTestDB(t)
	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00001-0000-0000-0000-000000000001", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"expense", "food", 300, "USD", "2026-03-01")
	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00002-0000-0000-0000-000000000002", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"expense", "food", 200, "USD", "2026-03-10")
	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00003-0000-0000-0000-000000000003", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"expense", "housing", 1500, "USD", "2026-03-01")
	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00004-0000-0000-0000-000000000004", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"deposit", "salary", 5000, "USD", "2026-03-01")

	spending, err := repo.GetSpendingByCategory(context.Background(), dashboardUserID, now)
	require.NoError(t, err)
	require.Len(t, spending, 2)
	assert.Equal(t, "housing", spending[0].Category)
	assert.Equal(t, 1500.0, spending[0].Amount)
	assert.Equal(t, "food", spending[1].Category)
	assert.Equal(t, 500.0, spending[1].Amount)
}

func TestGetSpendingByCategory_Empty(t *testing.T) {
	repo, _ := setupDashboardTestDB(t)
	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	spending, err := repo.GetSpendingByCategory(context.Background(), dashboardUserID, now)
	require.NoError(t, err)
	assert.Empty(t, spending)
}

func TestGetDebtsSummary(t *testing.T) {
	repo, exec := setupDashboardTestDB(t)
	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	exec(`INSERT INTO accounts (id, user_id, name, institution, account_type, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"acct0002-0000-0000-0000-000000000002", dashboardUserID, "Credit Card", "Chase", "credit_card", "USD")

	exec(`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank, linked_account_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"goal0001-0000-0000-0000-000000000001", dashboardUserID, "Pay off CC", "debt_payoff", 5000, 3000, 1, "acct0002-0000-0000-0000-000000000002")

	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00001-0000-0000-0000-000000000001", dashboardUserID, "acct0002-0000-0000-0000-000000000002",
		"expense", "credit_payment", 500, "USD", "2026-03-05")

	debts, err := repo.GetDebtsSummary(context.Background(), dashboardUserID, now)
	require.NoError(t, err)
	require.Len(t, debts, 1)
	assert.Equal(t, "Credit Card", debts[0].Name)
	assert.Equal(t, 3000.0, debts[0].Balance)
	assert.Equal(t, 500.0, debts[0].MonthlyPayment)
}

func TestGetDebtsSummary_ExcludesNonDebtAccounts(t *testing.T) {
	repo, _ := setupDashboardTestDB(t)
	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	debts, err := repo.GetDebtsSummary(context.Background(), dashboardUserID, now)
	require.NoError(t, err)
	assert.Empty(t, debts)
}

func TestGetNetWorthHistory(t *testing.T) {
	repo, exec := setupDashboardTestDB(t)

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "Brokerage", "brokerage", 50000, "USD")

	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000001-0000-0000-0000-000000000001", "ast00001-0000-0000-0000-000000000001", 45000, "2026-01-01")
	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000002-0000-0000-0000-000000000002", "ast00001-0000-0000-0000-000000000001", 48000, "2026-02-01")
	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000003-0000-0000-0000-000000000003", "ast00001-0000-0000-0000-000000000001", 50000, "2026-03-01")

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	points, err := repo.GetNetWorthHistory(context.Background(), dashboardUserID, since)
	require.NoError(t, err)
	require.Len(t, points, 3)
	assert.Equal(t, 45000.0, points[0].Value)
	assert.Equal(t, 48000.0, points[1].Value)
	assert.Equal(t, 50000.0, points[2].Value)
}

func TestGetNetWorthHistory_MultipleAssetsSameDate(t *testing.T) {
	repo, exec := setupDashboardTestDB(t)

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "Brokerage", "brokerage", 50000, "USD")
	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00002-0000-0000-0000-000000000002", dashboardUserID, "401k", "retirement_401k", 30000, "USD")

	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000001-0000-0000-0000-000000000001", "ast00001-0000-0000-0000-000000000001", 50000, "2026-03-01")
	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000002-0000-0000-0000-000000000002", "ast00002-0000-0000-0000-000000000002", 30000, "2026-03-01")

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	points, err := repo.GetNetWorthHistory(context.Background(), dashboardUserID, since)
	require.NoError(t, err)
	require.Len(t, points, 1)
	assert.Equal(t, 80000.0, points[0].Value)
}

func TestPingDB(t *testing.T) {
	repo, _ := setupDashboardTestDB(t)

	err := repo.PingDB(context.Background())
	assert.NoError(t, err)
}
