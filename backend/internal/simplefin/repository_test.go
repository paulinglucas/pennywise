package simplefin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/db"
)

func setupTestDB(t *testing.T) *SQLiteSimplefinRepository {
	t.Helper()
	database, err := db.Open(":memory:")
	require.NoError(t, err)
	require.NoError(t, db.Migrate(database))
	t.Cleanup(func() { _ = database.Close() })

	ctx := context.Background()
	_, err = database.ExecContext(ctx, `INSERT INTO users (id, email, name, password_hash) VALUES ('u1', 'test@test.com', 'Test', 'hash')`)
	require.NoError(t, err)

	_, err = database.ExecContext(ctx, `INSERT INTO accounts (id, user_id, name, institution, account_type) VALUES
		('a1', 'u1', 'Checking', 'Bank', 'checking'),
		('a2', 'u1', 'Savings', 'Bank', 'savings')`)
	require.NoError(t, err)

	_, err = database.ExecContext(ctx, `INSERT INTO assets (id, user_id, account_id, name, asset_type, current_value) VALUES
		('asset1', 'u1', 'a1', 'Checking Balance', 'liquid', 1000.00),
		('asset2', 'u1', 'a2', 'Savings Balance', 'liquid', 5000.00)`)
	require.NoError(t, err)

	return NewSimplefinRepository(database)
}

func TestSaveAndGetConnection(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	err := repo.SaveConnection(ctx, "u1", "encrypted-access-url")
	require.NoError(t, err)

	conn, err := repo.GetConnection(ctx, "u1")
	require.NoError(t, err)
	require.NotNil(t, conn)
	assert.Equal(t, "u1", conn.UserID)
	assert.Equal(t, "encrypted-access-url", conn.AccessURL)
	assert.Nil(t, conn.LastSyncAt)
	assert.Nil(t, conn.SyncError)
}

func TestSaveConnectionUpserts(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	err := repo.SaveConnection(ctx, "u1", "first-url")
	require.NoError(t, err)

	err = repo.SaveConnection(ctx, "u1", "second-url")
	require.NoError(t, err)

	conn, err := repo.GetConnection(ctx, "u1")
	require.NoError(t, err)
	assert.Equal(t, "second-url", conn.AccessURL)
}

func TestGetConnectionNotFound(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	conn, err := repo.GetConnection(ctx, "u1")
	require.NoError(t, err)
	assert.Nil(t, conn)
}

func TestDeleteConnection(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	err := repo.SaveConnection(ctx, "u1", "url")
	require.NoError(t, err)

	err = repo.DeleteConnection(ctx, "u1")
	require.NoError(t, err)

	conn, err := repo.GetConnection(ctx, "u1")
	require.NoError(t, err)
	assert.Nil(t, conn)
}

func TestLinkAndUnlinkAccount(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	err := repo.LinkAccount(ctx, "u1", "a1", "sfin-001")
	require.NoError(t, err)

	accounts, err := repo.GetLinkedAccounts(ctx, "u1")
	require.NoError(t, err)
	require.Len(t, accounts, 1)
	assert.Equal(t, "a1", accounts[0].AccountID)
	assert.Equal(t, "sfin-001", accounts[0].SimplefinID)
	assert.Equal(t, "Checking", accounts[0].AccountName)

	err = repo.UnlinkAccount(ctx, "u1", "a1")
	require.NoError(t, err)

	accounts, err = repo.GetLinkedAccounts(ctx, "u1")
	require.NoError(t, err)
	assert.Empty(t, accounts)
}

func TestUpdateSyncStatus(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	err := repo.SaveConnection(ctx, "u1", "url")
	require.NoError(t, err)

	err = repo.UpdateSyncSuccess(ctx, "u1")
	require.NoError(t, err)

	conn, err := repo.GetConnection(ctx, "u1")
	require.NoError(t, err)
	assert.NotNil(t, conn.LastSyncAt)
	assert.Nil(t, conn.SyncError)
}

func TestUpdateSyncError(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	err := repo.SaveConnection(ctx, "u1", "url")
	require.NoError(t, err)

	syncErr := "connection failed"
	err = repo.UpdateSyncError(ctx, "u1", syncErr)
	require.NoError(t, err)

	conn, err := repo.GetConnection(ctx, "u1")
	require.NoError(t, err)
	require.NotNil(t, conn.SyncError)
	assert.Equal(t, syncErr, *conn.SyncError)
}

func TestGetAssetForAccount(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	asset, err := repo.GetAssetForAccount(ctx, "u1", "a1")
	require.NoError(t, err)
	require.NotNil(t, asset)
	assert.Equal(t, "asset1", asset.ID)
	assert.Equal(t, 1000.0, asset.CurrentValue)
}

func TestGetAssetForAccountNotFound(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	asset, err := repo.GetAssetForAccount(ctx, "u1", "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, asset)
}

func TestUpdateAssetValue(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	err := repo.UpdateAssetValue(ctx, "asset1", 2500.0)
	require.NoError(t, err)

	asset, err := repo.GetAssetForAccount(ctx, "u1", "a1")
	require.NoError(t, err)
	require.NotNil(t, asset)
	assert.Equal(t, 2500.0, asset.CurrentValue)
}

func TestGetDebtGoalForAccount(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	_, err := repo.db.ExecContext(ctx, `INSERT INTO accounts (id, user_id, name, institution, account_type) VALUES ('cc1', 'u1', 'Credit Card', 'Bank', 'credit_card')`)
	require.NoError(t, err)

	_, err = repo.db.ExecContext(ctx,
		`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank, linked_account_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"goal1", "u1", "Pay CC", "debt_payoff", 5000, 3000, 1, "cc1",
	)
	require.NoError(t, err)

	goal, err := repo.GetDebtGoalForAccount(ctx, "cc1")
	require.NoError(t, err)
	require.NotNil(t, goal)
	assert.Equal(t, "goal1", goal.ID)
	assert.Equal(t, 3000.0, goal.CurrentAmount)
}

func TestGetDebtGoalForAccountNotFound(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	goal, err := repo.GetDebtGoalForAccount(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, goal)
}

func TestUpdateDebtBalance(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	_, err := repo.db.ExecContext(ctx, `INSERT INTO accounts (id, user_id, name, institution, account_type) VALUES ('cc1', 'u1', 'CC', 'Bank', 'credit_card')`)
	require.NoError(t, err)

	_, err = repo.db.ExecContext(ctx,
		`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank, linked_account_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"goal1", "u1", "Pay CC", "debt_payoff", 5000, 3000, 1, "cc1",
	)
	require.NoError(t, err)

	err = repo.UpdateDebtBalance(ctx, "goal1", 2500.0)
	require.NoError(t, err)

	goal, err := repo.GetDebtGoalForAccount(ctx, "cc1")
	require.NoError(t, err)
	require.NotNil(t, goal)
	assert.Equal(t, 2500.0, goal.CurrentAmount)
}

func TestGetAllConnections(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	conns, err := repo.GetAllConnections(ctx)
	require.NoError(t, err)
	assert.Empty(t, conns)

	err = repo.SaveConnection(ctx, "u1", "enc-url")
	require.NoError(t, err)

	conns, err = repo.GetAllConnections(ctx)
	require.NoError(t, err)
	require.Len(t, conns, 1)
	assert.Equal(t, "u1", conns[0].UserID)
}

func TestDeleteConnectionClearsSimplefinIDs(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	err := repo.LinkAccount(ctx, "u1", "a1", "sfin-001")
	require.NoError(t, err)

	err = repo.SaveConnection(ctx, "u1", "enc-url")
	require.NoError(t, err)

	err = repo.DeleteConnection(ctx, "u1")
	require.NoError(t, err)

	linked, err := repo.GetLinkedAccounts(ctx, "u1")
	require.NoError(t, err)
	assert.Empty(t, linked)
}

func TestCreateAccountWithLink_MortgageBackfillsHistory(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	rate := 6.5
	term := 360
	price := 400000.0
	date := "2024-01-15"
	downPct := 20.0

	mortgage := &MortgageFields{
		InterestRate:   &rate,
		LoanTermMonths: &term,
		PurchasePrice:  &price,
		PurchaseDate:   &date,
		DownPaymentPct: &downPct,
	}

	accountID, err := repo.CreateAccountWithLink(ctx, "u1", "Mortgage", "Bank", "mortgage", "USD", "sfin-mort", 310000, mortgage)
	require.NoError(t, err)
	require.NotEmpty(t, accountID)

	var balCount int
	err = repo.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM account_balance_history WHERE account_id = ?`, accountID).Scan(&balCount)
	require.NoError(t, err)
	assert.Greater(t, balCount, 12)

	var firstBal float64
	var firstDate string
	err = repo.db.QueryRowContext(ctx,
		`SELECT balance, DATE(recorded_at) FROM account_balance_history WHERE account_id = ? ORDER BY recorded_at ASC LIMIT 1`,
		accountID).Scan(&firstBal, &firstDate)
	require.NoError(t, err)
	assert.Equal(t, "2024-01-15", firstDate)
	assert.InDelta(t, 320000.0, firstBal, 0.01)

	var assetCount int
	err = repo.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM asset_history ah JOIN assets a ON a.id = ah.asset_id WHERE a.account_id = ?`,
		accountID).Scan(&assetCount)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, assetCount, 2)

	var assetFirstVal float64
	var assetFirstDate string
	err = repo.db.QueryRowContext(ctx,
		`SELECT ah.value, DATE(ah.recorded_at) FROM asset_history ah JOIN assets a ON a.id = ah.asset_id
		 WHERE a.account_id = ? ORDER BY ah.recorded_at ASC LIMIT 1`,
		accountID).Scan(&assetFirstVal, &assetFirstDate)
	require.NoError(t, err)
	assert.Equal(t, "2024-01-15", assetFirstDate)
	assert.InDelta(t, 400000.0, assetFirstVal, 0.01)

	var origBalance float64
	err = repo.db.QueryRowContext(ctx,
		`SELECT original_balance FROM accounts WHERE id = ?`, accountID).Scan(&origBalance)
	require.NoError(t, err)
	assert.InDelta(t, 400000.0, origBalance, 0.01)

	var targetAmount, currentAmount float64
	err = repo.db.QueryRowContext(ctx,
		`SELECT target_amount, current_amount FROM goals WHERE linked_account_id = ?`, accountID).Scan(&targetAmount, &currentAmount)
	require.NoError(t, err)
	assert.InDelta(t, 400000.0, targetAmount, 0.01)
	assert.InDelta(t, 310000.0, currentAmount, 0.01)
}

func TestCreateAccountWithLink_MortgageNoFieldsSkipsBackfill(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	mortgage := &MortgageFields{}

	accountID, err := repo.CreateAccountWithLink(ctx, "u1", "Mortgage", "Bank", "mortgage", "USD", "sfin-mort2", 250000, mortgage)
	require.NoError(t, err)

	var balCount int
	err = repo.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM account_balance_history WHERE account_id = ?`, accountID).Scan(&balCount)
	require.NoError(t, err)
	assert.Equal(t, 1, balCount)
}

func TestReconstructBalanceHistory(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	_, err := repo.db.ExecContext(ctx, `INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES
		('t1', 'u1', 'a1', 'deposit', 'income', 5000, 'USD', '2026-01-01'),
		('t2', 'u1', 'a1', 'expense', 'food', 200, 'USD', '2026-01-05'),
		('t3', 'u1', 'a1', 'expense', 'gas', 50, 'USD', '2026-01-05'),
		('t4', 'u1', 'a1', 'transfer', 'transfer', 1000, 'USD', '2026-01-10'),
		('t5', 'u1', 'a1', 'deposit', 'income', 5000, 'USD', '2026-01-15')`)
	require.NoError(t, err)

	currentBalance := 8750.0

	err = repo.ReconstructBalanceHistory(ctx, "u1", "a1", currentBalance)
	require.NoError(t, err)

	rows, err := repo.db.QueryContext(ctx,
		`SELECT balance, DATE(recorded_at) FROM account_balance_history
		 WHERE account_id = 'a1' ORDER BY recorded_at ASC`)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	type entry struct {
		balance float64
		date    string
	}
	var entries []entry
	for rows.Next() {
		var e entry
		require.NoError(t, rows.Scan(&e.balance, &e.date))
		entries = append(entries, e)
	}
	require.NoError(t, rows.Err())

	require.Len(t, entries, 4)
	assert.Equal(t, "2026-01-01", entries[0].date)
	assert.InDelta(t, 5000.0, entries[0].balance, 0.01)
	assert.Equal(t, "2026-01-05", entries[1].date)
	assert.InDelta(t, 4750.0, entries[1].balance, 0.01)
	assert.Equal(t, "2026-01-10", entries[2].date)
	assert.InDelta(t, 3750.0, entries[2].balance, 0.01)
	assert.Equal(t, "2026-01-15", entries[3].date)
	assert.InDelta(t, 8750.0, entries[3].balance, 0.01)
}

func TestReconstructBalanceHistory_Idempotent(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	_, err := repo.db.ExecContext(ctx, `INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES
		('t1', 'u1', 'a1', 'deposit', 'income', 3000, 'USD', '2026-01-01'),
		('t2', 'u1', 'a1', 'expense', 'food', 500, 'USD', '2026-01-10')`)
	require.NoError(t, err)

	_, err = repo.db.ExecContext(ctx,
		`INSERT INTO account_balance_history (id, account_id, balance, recorded_at) VALUES ('abh-init', 'a1', 2500, datetime('now'))`)
	require.NoError(t, err)

	err = repo.ReconstructBalanceHistory(ctx, "u1", "a1", 2500)
	require.NoError(t, err)

	var count1 int
	err = repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM account_balance_history WHERE account_id = 'a1'`).Scan(&count1)
	require.NoError(t, err)

	err = repo.ReconstructBalanceHistory(ctx, "u1", "a1", 2500)
	require.NoError(t, err)

	var count2 int
	err = repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM account_balance_history WHERE account_id = 'a1'`).Scan(&count2)
	require.NoError(t, err)

	assert.Equal(t, count1, count2)
}

func TestReconstructBalanceHistory_NoTransactions(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	err := repo.ReconstructBalanceHistory(ctx, "u1", "a1", 1000)
	require.NoError(t, err)

	var count int
	err = repo.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM account_balance_history WHERE account_id = 'a1'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}
