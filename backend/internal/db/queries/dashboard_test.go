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
	t.Parallel()
	repo, _ := setupDashboardTestDB(t)

	result, err := repo.GetNetWorth(context.Background(), dashboardUserID)
	require.NoError(t, err)
	assert.Equal(t, 0.0, result.AssetTotal)
	assert.Equal(t, 0.0, result.CashTotal)
	assert.Equal(t, 0.0, result.DebtTotal)
}

func TestGetNetWorth_WithCashAccounts(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)

	exec(`UPDATE accounts SET current_balance = 4500 WHERE id = ?`, "acct0001-0000-0000-0000-000000000001")

	result, err := repo.GetNetWorth(context.Background(), dashboardUserID)
	require.NoError(t, err)
	assert.Equal(t, 4500.0, result.CashTotal)
}

func TestGetNetWorth_WithAssets(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "Brokerage", "brokerage", 100000, "USD")

	exec(`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, original_balance, current_balance) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"acct0002-0000-0000-0000-000000000002", dashboardUserID, "Mortgage", "Bank", "mortgage", "USD", 200000, 150000)

	result, err := repo.GetNetWorth(context.Background(), dashboardUserID)
	require.NoError(t, err)
	assert.Equal(t, 100000.0, result.AssetTotal)
	assert.Equal(t, 150000.0, result.DebtTotal)
}

func TestGetNetWorth_DebtAccountWithoutGoal(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)

	exec(`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, original_balance) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"acct0003-0000-0000-0000-000000000003", dashboardUserID, "Credit Card", "Bank", "credit_card", "USD", 5000)

	result, err := repo.GetNetWorth(context.Background(), dashboardUserID)
	require.NoError(t, err)
	assert.Equal(t, 5000.0, result.DebtTotal)
}

func TestGetNetWorth_ExcludesSoftDeleted(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)
	since := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

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

	spending, err := repo.GetSpendingByCategory(context.Background(), dashboardUserID, since, until)
	require.NoError(t, err)
	require.Len(t, spending, 2)
	assert.Equal(t, "housing", spending[0].Category)
	assert.Equal(t, 1500.0, spending[0].Amount)
	assert.Equal(t, "food", spending[1].Category)
	assert.Equal(t, 500.0, spending[1].Amount)
}

func TestGetSpendingByCategory_Empty(t *testing.T) {
	t.Parallel()
	repo, _ := setupDashboardTestDB(t)
	since := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	spending, err := repo.GetSpendingByCategory(context.Background(), dashboardUserID, since, until)
	require.NoError(t, err)
	assert.Empty(t, spending)
}

func TestGetDebtsSummary_OriginalBalance(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)
	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	exec(`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, original_balance) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"acct0003-0000-0000-0000-000000000003", dashboardUserID, "Mortgage", "Wells Fargo", "mortgage", "USD", 300000.0)

	exec(`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank, linked_account_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"goal0002-0000-0000-0000-000000000002", dashboardUserID, "Pay mortgage", "debt_payoff", 300000, 250000, 1, "acct0003-0000-0000-0000-000000000003")

	debts, err := repo.GetDebtsSummary(context.Background(), dashboardUserID, now)
	require.NoError(t, err)
	require.Len(t, debts, 1)
	require.NotNil(t, debts[0].OriginalBalance)
	assert.Equal(t, 300000.0, *debts[0].OriginalBalance)
}

func TestGetDebtsSummary(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)
	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	exec(`INSERT INTO accounts (id, user_id, name, institution, account_type, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"acct0002-0000-0000-0000-000000000002", dashboardUserID, "Credit Card", "Chase", "credit_card", "USD")

	exec(`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank, linked_account_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"goal0001-0000-0000-0000-000000000001", dashboardUserID, "Pay off CC", "debt_payoff", 5000, 3000, 1, "acct0002-0000-0000-0000-000000000002")

	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00001-0000-0000-0000-000000000001", dashboardUserID, "acct0002-0000-0000-0000-000000000002",
		"deposit", "transfer", 500, "USD", "2026-03-05")

	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00010-0000-0000-0000-000000000010", dashboardUserID, "acct0002-0000-0000-0000-000000000002",
		"expense", "food", 200, "USD", "2026-03-06")

	debts, err := repo.GetDebtsSummary(context.Background(), dashboardUserID, now)
	require.NoError(t, err)
	require.Len(t, debts, 1)
	assert.Equal(t, "Credit Card", debts[0].Name)
	assert.Equal(t, 3000.0, debts[0].Balance)
	assert.Equal(t, 500.0, debts[0].MonthlyPayment)
}

func TestGetDebtsSummary_ExcludesNonDebtAccounts(t *testing.T) {
	t.Parallel()
	repo, _ := setupDashboardTestDB(t)
	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	debts, err := repo.GetDebtsSummary(context.Background(), dashboardUserID, now)
	require.NoError(t, err)
	assert.Empty(t, debts)
}

func TestGetNetWorthHistory(t *testing.T) {
	t.Parallel()
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
	points, err := repo.GetNetWorthHistory(context.Background(), dashboardUserID, since, true)
	require.NoError(t, err)
	assert.Equal(t, 45000.0, points[0].Value)
	assert.Equal(t, 48000.0, points[1].Value)

	lastPoint := points[len(points)-1]
	today := time.Now().Format("2006-01-02")
	assert.Equal(t, today, lastPoint.Date.Format("2006-01-02"))
	assert.Equal(t, 50000.0, lastPoint.Value)
}

func TestGetNetWorthHistory_MultipleAssetsSameDate(t *testing.T) {
	t.Parallel()
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
	points, err := repo.GetNetWorthHistory(context.Background(), dashboardUserID, since, true)
	require.NoError(t, err)

	assert.Equal(t, "2026-01-01", points[0].Date.Format("2006-01-02"))
	assert.Equal(t, 0.0, points[0].Value)

	assert.Equal(t, "2026-03-01", points[1].Date.Format("2006-01-02"))
	assert.Equal(t, 80000.0, points[1].Value)

	lastPoint := points[len(points)-1]
	assert.Equal(t, 80000.0, lastPoint.Value)
}

func TestGetNetWorthHistory_IncludesCashAndDebts(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)

	exec(`UPDATE accounts SET current_balance = 9000 WHERE id = ?`, "acct0001-0000-0000-0000-000000000001")

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "House", "real_estate", 300000, "USD")

	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000001-0000-0000-0000-000000000001", "ast00001-0000-0000-0000-000000000001", 290000, "2026-01-01")
	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000002-0000-0000-0000-000000000002", "ast00001-0000-0000-0000-000000000001", 295000, "2026-02-01")

	exec(`INSERT INTO account_balance_history (id, account_id, balance, recorded_at) VALUES (?, ?, ?, ?)`,
		"abh00010-0000-0000-0000-000000000010", "acct0001-0000-0000-0000-000000000001", 5000, "2025-12-15")
	exec(`INSERT INTO account_balance_history (id, account_id, balance, recorded_at) VALUES (?, ?, ?, ?)`,
		"abh00011-0000-0000-0000-000000000011", "acct0001-0000-0000-0000-000000000001", 9000, "2026-01-15")

	exec(`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, original_balance, current_balance, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"acct0002-0000-0000-0000-000000000002", dashboardUserID, "Mortgage", "Bank", "mortgage", "USD", 300000, 200000, "2025-01-01")

	exec(`INSERT INTO account_balance_history (id, account_id, balance, recorded_at) VALUES (?, ?, ?, ?)`,
		"abh00001-0000-0000-0000-000000000001", "acct0002-0000-0000-0000-000000000002", 200000, "2025-12-01")

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	points, err := repo.GetNetWorthHistory(context.Background(), dashboardUserID, since, true)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(points), 3)
	assert.Equal(t, "2026-01-01", points[0].Date.Format("2006-01-02"))
	assert.Equal(t, 95000.0, points[0].Value)

	assert.Equal(t, "2026-01-15", points[1].Date.Format("2006-01-02"))
	assert.Equal(t, 99000.0, points[1].Value)

	assert.Equal(t, "2026-02-01", points[2].Date.Format("2006-01-02"))
	assert.Equal(t, 104000.0, points[2].Value)
}

func TestGetNetWorthHistory_DebtConsistentAcrossAllDates(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "Retirement", "retirement", 20000, "USD")

	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000001-0000-0000-0000-000000000001", "ast00001-0000-0000-0000-000000000001", 15000, "2025-01-01")
	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000002-0000-0000-0000-000000000002", "ast00001-0000-0000-0000-000000000001", 18000, "2025-07-01")

	exec(`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, original_balance, current_balance, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"acct0002-0000-0000-0000-000000000002", dashboardUserID, "Car Loan", "Bank", "credit_line", "USD", 10000, 5000, "2025-06-01")

	exec(`INSERT INTO account_balance_history (id, account_id, balance, recorded_at) VALUES (?, ?, ?, ?)`,
		"abh00001-0000-0000-0000-000000000001", "acct0002-0000-0000-0000-000000000002", 5000, "2025-01-01")

	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	points, err := repo.GetNetWorthHistory(context.Background(), dashboardUserID, since, true)
	require.NoError(t, err)

	assert.Equal(t, 10000.0, points[0].Value)
	assert.Equal(t, 13000.0, points[1].Value)
}

func TestGetNetWorthHistory_AnchorPoint(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "Brokerage", "brokerage", 60000, "USD")

	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000001-0000-0000-0000-000000000001", "ast00001-0000-0000-0000-000000000001", 50000, "2026-01-01")
	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000002-0000-0000-0000-000000000002", "ast00001-0000-0000-0000-000000000001", 55000, "2026-02-01")

	since := time.Date(2026, 2, 8, 0, 0, 0, 0, time.UTC)
	points, err := repo.GetNetWorthHistory(context.Background(), dashboardUserID, since, true)
	require.NoError(t, err)

	require.True(t, len(points) >= 2)
	assert.Equal(t, "2026-02-08", points[0].Date.Format("2006-01-02"))
	assert.Equal(t, 55000.0, points[0].Value)
}

func TestGetNetWorthHistory_AlwaysIncludesToday(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "Brokerage", "brokerage", 75000, "USD")

	since := time.Now().AddDate(0, 0, -7)
	points, err := repo.GetNetWorthHistory(context.Background(), dashboardUserID, since, true)
	require.NoError(t, err)
	require.Len(t, points, 2)

	sinceStr := since.Format("2006-01-02")
	assert.Equal(t, sinceStr, points[0].Date.Format("2006-01-02"))

	today := time.Now().Format("2006-01-02")
	assert.Equal(t, today, points[1].Date.Format("2006-01-02"))
	assert.Equal(t, 75000.0, points[1].Value)
}

func TestGetNetWorth_UsesCurrentBalance(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)

	exec(`UPDATE accounts SET current_balance = 3500 WHERE id = ?`, "acct0001-0000-0000-0000-000000000001")

	result, err := repo.GetNetWorth(context.Background(), dashboardUserID)
	require.NoError(t, err)
	assert.Equal(t, 3500.0, result.CashTotal)
}

func TestGetCashFlowThisMonth_ExcludesTransfers(t *testing.T) {
	t.Parallel()
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
		"transfer", "credit_card_payment", 1200, "USD", "2026-03-10")

	cashFlow, err := repo.GetCashFlowThisMonth(context.Background(), dashboardUserID, now)
	require.NoError(t, err)
	assert.Equal(t, 4500.0, cashFlow)
}

func TestGetSpendingByCategory_ExcludesTransfers(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)
	since := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00001-0000-0000-0000-000000000001", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"expense", "food", 300, "USD", "2026-03-01")
	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00002-0000-0000-0000-000000000002", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"transfer", "credit_card_payment", 1200, "USD", "2026-03-10")

	spending, err := repo.GetSpendingByCategory(context.Background(), dashboardUserID, since, until)
	require.NoError(t, err)
	require.Len(t, spending, 1)
	assert.Equal(t, "food", spending[0].Category)
	assert.Equal(t, 300.0, spending[0].Amount)
}

func TestGetNetWorthHistory_CashUsesBalanceHistory(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)

	exec(`UPDATE accounts SET current_balance = 3500 WHERE id = ?`, "acct0001-0000-0000-0000-000000000001")

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "Brokerage", "brokerage", 50000, "USD")

	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000001-0000-0000-0000-000000000001", "ast00001-0000-0000-0000-000000000001", 50000, "2026-03-01")

	exec(`INSERT INTO account_balance_history (id, account_id, balance, recorded_at) VALUES (?, ?, ?, ?)`,
		"abh00001-0000-0000-0000-000000000001", "acct0001-0000-0000-0000-000000000001", 5000, "2026-03-01")
	exec(`INSERT INTO account_balance_history (id, account_id, balance, recorded_at) VALUES (?, ?, ?, ?)`,
		"abh00002-0000-0000-0000-000000000002", "acct0001-0000-0000-0000-000000000001", 3500, "2026-03-10")

	since := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	points, err := repo.GetNetWorthHistory(context.Background(), dashboardUserID, since, true)
	require.NoError(t, err)

	assert.Equal(t, "2026-03-01", points[0].Date.Format("2006-01-02"))
	assert.Equal(t, 55000.0, points[0].Value)

	assert.Equal(t, "2026-03-10", points[1].Date.Format("2006-01-02"))
	assert.Equal(t, 53500.0, points[1].Value)
}

func TestGetDebtsSummary_TransferCountsAsPayment(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)
	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	exec(`INSERT INTO accounts (id, user_id, name, institution, account_type, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"acct0002-0000-0000-0000-000000000002", dashboardUserID, "Credit Card", "Chase", "credit_card", "USD")

	exec(`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank, linked_account_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"goal0001-0000-0000-0000-000000000001", dashboardUserID, "Pay off CC", "debt_payoff", 5000, 3000, 1, "acct0002-0000-0000-0000-000000000002")

	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00001-0000-0000-0000-000000000001", dashboardUserID, "acct0002-0000-0000-0000-000000000002",
		"deposit", "transfer", 500, "USD", "2026-03-05")
	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00002-0000-0000-0000-000000000002", dashboardUserID, "acct0002-0000-0000-0000-000000000002",
		"expense", "dining", 200, "USD", "2026-03-06")

	debts, err := repo.GetDebtsSummary(context.Background(), dashboardUserID, now)
	require.NoError(t, err)
	require.Len(t, debts, 1)
	assert.Equal(t, 500.0, debts[0].MonthlyPayment)
}

func TestGetDebtsSummary_ExcludesChargesFromPayment(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)
	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	exec(`INSERT INTO accounts (id, user_id, name, institution, account_type, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"acct0002-0000-0000-0000-000000000002", dashboardUserID, "Credit Card", "Chase", "credit_card", "USD")

	exec(`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank, linked_account_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"goal0001-0000-0000-0000-000000000001", dashboardUserID, "Pay off CC", "debt_payoff", 5000, 3000, 1, "acct0002-0000-0000-0000-000000000002")

	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00001-0000-0000-0000-000000000001", dashboardUserID, "acct0002-0000-0000-0000-000000000002",
		"expense", "dining", 300, "USD", "2026-03-05")

	debts, err := repo.GetDebtsSummary(context.Background(), dashboardUserID, now)
	require.NoError(t, err)
	require.Len(t, debts, 1)
	assert.Equal(t, 0.0, debts[0].MonthlyPayment)
}

func TestGetSpendingByCategory_ExcludesExpenseTypedTransfers(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)
	since := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00001-0000-0000-0000-000000000001", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"expense", "groceries", 200, "USD", "2026-03-01")
	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00002-0000-0000-0000-000000000002", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"expense", "transfer", 1500, "USD", "2026-03-10")
	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00003-0000-0000-0000-000000000003", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"expense", "cash", 100, "USD", "2026-03-12")

	spending, err := repo.GetSpendingByCategory(context.Background(), dashboardUserID, since, until)
	require.NoError(t, err)
	require.Len(t, spending, 1)
	assert.Equal(t, "groceries", spending[0].Category)
	assert.Equal(t, 200.0, spending[0].Amount)
}

func TestGetCashFlowThisMonth_ExcludesExpenseTypedTransfers(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)
	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00001-0000-0000-0000-000000000001", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"deposit", "salary", 5000, "USD", "2026-03-01")
	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00002-0000-0000-0000-000000000002", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"expense", "groceries", 300, "USD", "2026-03-05")
	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00003-0000-0000-0000-000000000003", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"expense", "transfer", 1500, "USD", "2026-03-10")

	cashFlow, err := repo.GetCashFlowThisMonth(context.Background(), dashboardUserID, now)
	require.NoError(t, err)
	assert.Equal(t, 4700.0, cashFlow)
}

func TestGetNetWorthHistory_PerAssetCarryForward(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "Brokerage", "brokerage", 50000, "USD")
	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00002-0000-0000-0000-000000000002", dashboardUserID, "Rollover IRA", "rollover_ira", 42000, "USD")

	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000001-0000-0000-0000-000000000001", "ast00001-0000-0000-0000-000000000001", 45000, "2025-06-01")
	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000002-0000-0000-0000-000000000002", "ast00002-0000-0000-0000-000000000002", 40000, "2025-04-01")
	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000003-0000-0000-0000-000000000003", "ast00001-0000-0000-0000-000000000001", 48000, "2025-09-01")

	since := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	points, err := repo.GetNetWorthHistory(context.Background(), dashboardUserID, since, true)
	require.NoError(t, err)

	assert.Equal(t, "2025-01-01", points[0].Date.Format("2006-01-02"))
	assert.Equal(t, 0.0, points[0].Value)

	assert.Equal(t, "2025-04-01", points[1].Date.Format("2006-01-02"))
	assert.Equal(t, 40000.0, points[1].Value)

	assert.Equal(t, "2025-06-01", points[2].Date.Format("2006-01-02"))
	assert.Equal(t, 85000.0, points[2].Value)

	assert.Equal(t, "2025-09-01", points[3].Date.Format("2006-01-02"))
	assert.Equal(t, 88000.0, points[3].Value)
}

func TestGetNetWorthHistory_BalanceHistoryDatesCreateDataPoints(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)

	exec(`UPDATE accounts SET current_balance = 4500 WHERE id = ?`, "acct0001-0000-0000-0000-000000000001")

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "Brokerage", "brokerage", 50000, "USD")

	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000001-0000-0000-0000-000000000001", "ast00001-0000-0000-0000-000000000001", 50000, "2026-03-01")

	exec(`INSERT INTO account_balance_history (id, account_id, balance, recorded_at) VALUES (?, ?, ?, ?)`,
		"abh00001-0000-0000-0000-000000000001", "acct0001-0000-0000-0000-000000000001", 5000, "2026-03-01")
	exec(`INSERT INTO account_balance_history (id, account_id, balance, recorded_at) VALUES (?, ?, ?, ?)`,
		"abh00002-0000-0000-0000-000000000002", "acct0001-0000-0000-0000-000000000001", 4800, "2026-03-05")
	exec(`INSERT INTO account_balance_history (id, account_id, balance, recorded_at) VALUES (?, ?, ?, ?)`,
		"abh00003-0000-0000-0000-000000000003", "acct0001-0000-0000-0000-000000000001", 4500, "2026-03-10")

	since := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	points, err := repo.GetNetWorthHistory(context.Background(), dashboardUserID, since, true)
	require.NoError(t, err)

	require.True(t, len(points) >= 3)
	assert.Equal(t, "2026-03-01", points[0].Date.Format("2006-01-02"))
	assert.Equal(t, 55000.0, points[0].Value)

	assert.Equal(t, "2026-03-05", points[1].Date.Format("2006-01-02"))
	assert.Equal(t, 54800.0, points[1].Value)

	assert.Equal(t, "2026-03-10", points[2].Date.Format("2006-01-02"))
	assert.Equal(t, 54500.0, points[2].Value)
}

func TestGetNetWorth_CreditCardExpensesNotInCash(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)

	exec(`UPDATE accounts SET current_balance = 5000 WHERE id = ?`, "acct0001-0000-0000-0000-000000000001")

	exec(`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, current_balance) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"acct0002-0000-0000-0000-000000000002", dashboardUserID, "Amex", "Amex", "credit_card", "USD", 1000)

	result, err := repo.GetNetWorth(context.Background(), dashboardUserID)
	require.NoError(t, err)
	assert.Equal(t, 5000.0, result.CashTotal)
	assert.Equal(t, 1000.0, result.DebtTotal)
}

func TestGetSpendingByCategory_IncludesCreditCardExpenses(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)
	since := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)

	exec(`INSERT INTO accounts (id, user_id, name, institution, account_type, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"acct0002-0000-0000-0000-000000000002", dashboardUserID, "Amex", "Amex", "credit_card", "USD")

	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00001-0000-0000-0000-000000000001", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"expense", "food", 100, "USD", "2026-03-01")
	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00002-0000-0000-0000-000000000002", dashboardUserID, "acct0002-0000-0000-0000-000000000002",
		"expense", "food", 200, "USD", "2026-03-05")

	spending, err := repo.GetSpendingByCategory(context.Background(), dashboardUserID, since, until)
	require.NoError(t, err)
	require.Len(t, spending, 1)
	assert.Equal(t, "food", spending[0].Category)
	assert.Equal(t, 300.0, spending[0].Amount)
}

func TestGetNetWorthHistory_MultipleAssetsWithDifferentSnapshotFrequencies(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "Brokerage", "brokerage", 55000, "USD")
	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00002-0000-0000-0000-000000000002", dashboardUserID, "Rollover IRA", "rollover_ira", 42000, "USD")
	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00003-0000-0000-0000-000000000003", dashboardUserID, "401k", "retirement_401k", 30000, "USD")

	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000001-0000-0000-0000-000000000001", "ast00001-0000-0000-0000-000000000001", 50000, "2026-01-01")
	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000002-0000-0000-0000-000000000002", "ast00002-0000-0000-0000-000000000002", 40000, "2025-04-01")
	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000003-0000-0000-0000-000000000003", "ast00003-0000-0000-0000-000000000003", 28000, "2026-01-01")
	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000004-0000-0000-0000-000000000004", "ast00001-0000-0000-0000-000000000001", 52000, "2026-02-01")
	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000005-0000-0000-0000-000000000005", "ast00003-0000-0000-0000-000000000003", 29000, "2026-02-01")

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	points, err := repo.GetNetWorthHistory(context.Background(), dashboardUserID, since, true)
	require.NoError(t, err)

	assert.Equal(t, "2026-01-01", points[0].Date.Format("2006-01-02"))
	assert.Equal(t, 118000.0, points[0].Value)

	assert.Equal(t, "2026-02-01", points[1].Date.Format("2006-01-02"))
	assert.Equal(t, 121000.0, points[1].Value)
}

func TestGetNetWorthHistory_MultipleCashSnapshots(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)

	exec(`UPDATE accounts SET current_balance = 3800 WHERE id = ?`, "acct0001-0000-0000-0000-000000000001")

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "Brokerage", "brokerage", 50000, "USD")

	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000001-0000-0000-0000-000000000001", "ast00001-0000-0000-0000-000000000001", 50000, "2026-03-01")

	exec(`INSERT INTO account_balance_history (id, account_id, balance, recorded_at) VALUES (?, ?, ?, ?)`,
		"abh00001-0000-0000-0000-000000000001", "acct0001-0000-0000-0000-000000000001", 5000, "2026-03-01")
	exec(`INSERT INTO account_balance_history (id, account_id, balance, recorded_at) VALUES (?, ?, ?, ?)`,
		"abh00002-0000-0000-0000-000000000002", "acct0001-0000-0000-0000-000000000001", 4800, "2026-03-05")
	exec(`INSERT INTO account_balance_history (id, account_id, balance, recorded_at) VALUES (?, ?, ?, ?)`,
		"abh00003-0000-0000-0000-000000000003", "acct0001-0000-0000-0000-000000000001", 3800, "2026-03-10")

	since := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	points, err := repo.GetNetWorthHistory(context.Background(), dashboardUserID, since, true)
	require.NoError(t, err)

	assert.Equal(t, 55000.0, points[0].Value)
	assert.Equal(t, 54800.0, points[1].Value)
	assert.Equal(t, 53800.0, points[2].Value)
}

func TestGetNetWorthHistory_TodayPointReplacesExisting(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)

	exec(`INSERT INTO assets (id, user_id, name, asset_type, current_value, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"ast00001-0000-0000-0000-000000000001", dashboardUserID, "Brokerage", "brokerage", 60000, "USD")

	today := time.Now().Format("2006-01-02")
	exec(`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"ah000001-0000-0000-0000-000000000001", "ast00001-0000-0000-0000-000000000001", 55000, today)

	since := time.Now().AddDate(0, -1, 0)
	points, err := repo.GetNetWorthHistory(context.Background(), dashboardUserID, since, true)
	require.NoError(t, err)

	lastPoint := points[len(points)-1]
	assert.Equal(t, today, lastPoint.Date.Format("2006-01-02"))
	assert.Equal(t, 60000.0, lastPoint.Value)
}

func TestGetCashFlowThisMonth_ExcludesDepositCategoryTransfer(t *testing.T) {
	t.Parallel()
	repo, exec := setupDashboardTestDB(t)
	now := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	exec(`INSERT INTO accounts (id, user_id, name, institution, account_type, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"acct0002-0000-0000-0000-000000000002", dashboardUserID, "Amex", "Amex", "credit_card", "USD")

	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00001-0000-0000-0000-000000000001", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"deposit", "salary", 5000, "USD", "2026-03-01")
	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00002-0000-0000-0000-000000000002", dashboardUserID, "acct0002-0000-0000-0000-000000000002",
		"deposit", "transfer", 1200, "USD", "2026-03-10")
	exec(`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"txn00003-0000-0000-0000-000000000003", dashboardUserID, "acct0001-0000-0000-0000-000000000001",
		"expense", "food", 300, "USD", "2026-03-05")

	cashFlow, err := repo.GetCashFlowThisMonth(context.Background(), dashboardUserID, now)
	require.NoError(t, err)
	assert.Equal(t, 4700.0, cashFlow)
}

func TestPingDB(t *testing.T) {
	t.Parallel()
	repo, _ := setupDashboardTestDB(t)

	err := repo.PingDB(context.Background())
	assert.NoError(t, err)
}
