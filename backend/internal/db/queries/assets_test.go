package queries_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/models"
)

const (
	assetTestUserID  = "usr00001-0000-0000-0000-000000000001"
	assetTestUser2ID = "usr00002-0000-0000-0000-000000000002"
)

func setupAssetTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database := setupTestDB(t)
	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		assetTestUser2ID, "alex@example.com", "Alex", "$2a$10$hashedpassword",
	)
	require.NoError(t, err)
	return database
}

func seedAsset(t *testing.T, repo *queries.SQLiteAssetRepository, id, userID, assetType string, value float64) {
	t.Helper()
	err := repo.Create(context.Background(), &models.Asset{
		ID:           id,
		UserID:       userID,
		Name:         "Test Asset " + id,
		AssetType:    assetType,
		CurrentValue: value,
		Currency:     "USD",
	})
	require.NoError(t, err)
}

func TestAssetCreate_And_GetByID(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	metadata := `{"provider":"Fidelity"}`
	err := repo.Create(context.Background(), &models.Asset{
		ID:           "ast-test-001",
		UserID:       assetTestUserID,
		Name:         "Roth IRA",
		AssetType:    "retirement",
		CurrentValue: 50000,
		Currency:     "USD",
		Metadata:     &metadata,
	})
	require.NoError(t, err)

	asset, err := repo.GetByID(context.Background(), assetTestUserID, "ast-test-001")
	require.NoError(t, err)
	require.NotNil(t, asset)
	assert.Equal(t, "ast-test-001", asset.ID)
	assert.Equal(t, "Roth IRA", asset.Name)
	assert.Equal(t, "retirement", asset.AssetType)
	assert.Equal(t, 50000.0, asset.CurrentValue)
	assert.Equal(t, "USD", asset.Currency)
	assert.NotNil(t, asset.Metadata)
	assert.Contains(t, *asset.Metadata, "Fidelity")
}

func TestAssetCreate_CreatesHistoryEntry(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	seedAsset(t, repo, "ast-test-001", assetTestUserID, "liquid", 10000)

	history, err := repo.GetHistory(context.Background(), assetTestUserID, "ast-test-001", nil)
	require.NoError(t, err)
	require.Len(t, history, 1)
	assert.Equal(t, 10000.0, history[0].Value)
}

func TestAssetGetByID_NotFound(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	asset, err := repo.GetByID(context.Background(), assetTestUserID, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, asset)
}

func TestAssetGetByID_WrongUser(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	seedAsset(t, repo, "ast-test-001", assetTestUserID, "liquid", 10000)

	asset, err := repo.GetByID(context.Background(), assetTestUser2ID, "ast-test-001")
	require.NoError(t, err)
	assert.Nil(t, asset)
}

func TestAssetList_FiltersByUser(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	seedAsset(t, repo, "ast-u1-001", assetTestUserID, "liquid", 10000)
	seedAsset(t, repo, "ast-u1-002", assetTestUserID, "retirement", 50000)
	seedAsset(t, repo, "ast-u2-001", assetTestUser2ID, "liquid", 5000)

	assets, total, err := repo.List(context.Background(), assetTestUserID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, assets, 2)
}

func TestAssetList_Pagination(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	for i := range 5 {
		seedAsset(t, repo, fmt.Sprintf("ast-page-%03d", i+1), assetTestUserID, "liquid", float64((i+1)*1000))
	}

	assets, total, err := repo.List(context.Background(), assetTestUserID, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, assets, 2)

	assets3, _, err := repo.List(context.Background(), assetTestUserID, 3, 2)
	require.NoError(t, err)
	assert.Len(t, assets3, 1)
}

func TestAssetUpdate_ChangesValueAndCreatesHistory(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	seedAsset(t, repo, "ast-test-001", assetTestUserID, "liquid", 10000)

	asset, err := repo.GetByID(context.Background(), assetTestUserID, "ast-test-001")
	require.NoError(t, err)

	prevValue := asset.CurrentValue
	asset.CurrentValue = 12000
	asset.Name = "Updated Asset"

	found, err := repo.Update(context.Background(), asset, prevValue)
	require.NoError(t, err)
	assert.True(t, found)

	updated, err := repo.GetByID(context.Background(), assetTestUserID, "ast-test-001")
	require.NoError(t, err)
	assert.Equal(t, 12000.0, updated.CurrentValue)
	assert.Equal(t, "Updated Asset", updated.Name)

	history, err := repo.GetHistory(context.Background(), assetTestUserID, "ast-test-001", nil)
	require.NoError(t, err)
	assert.Len(t, history, 2)
	assert.Equal(t, 10000.0, history[0].Value)
	assert.Equal(t, 12000.0, history[1].Value)
}

func TestAssetUpdate_SameValueNoNewHistory(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	seedAsset(t, repo, "ast-test-001", assetTestUserID, "liquid", 10000)

	asset, err := repo.GetByID(context.Background(), assetTestUserID, "ast-test-001")
	require.NoError(t, err)

	asset.Name = "Renamed"
	found, err := repo.Update(context.Background(), asset, asset.CurrentValue)
	require.NoError(t, err)
	assert.True(t, found)

	history, err := repo.GetHistory(context.Background(), assetTestUserID, "ast-test-001", nil)
	require.NoError(t, err)
	assert.Len(t, history, 1)
}

func TestAssetUpdate_WrongUser(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	seedAsset(t, repo, "ast-test-001", assetTestUserID, "liquid", 10000)

	found, err := repo.Update(context.Background(), &models.Asset{
		ID:           "ast-test-001",
		UserID:       assetTestUser2ID,
		Name:         "Hacked",
		AssetType:    "liquid",
		CurrentValue: 99999,
		Currency:     "USD",
	}, 10000)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestAssetSoftDelete(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	seedAsset(t, repo, "ast-test-001", assetTestUserID, "liquid", 10000)

	found, err := repo.SoftDelete(context.Background(), assetTestUserID, "ast-test-001")
	require.NoError(t, err)
	assert.True(t, found)

	asset, err := repo.GetByID(context.Background(), assetTestUserID, "ast-test-001")
	require.NoError(t, err)
	assert.Nil(t, asset)
}

func TestAssetSoftDelete_ExcludedFromList(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	seedAsset(t, repo, "ast-test-001", assetTestUserID, "liquid", 10000)
	seedAsset(t, repo, "ast-test-002", assetTestUserID, "retirement", 50000)

	found, err := repo.SoftDelete(context.Background(), assetTestUserID, "ast-test-001")
	require.NoError(t, err)
	assert.True(t, found)

	assets, total, err := repo.List(context.Background(), assetTestUserID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, assets, 1)
	assert.Equal(t, "ast-test-002", assets[0].ID)
}

func TestAssetGetHistory_NonexistentAsset(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	history, err := repo.GetHistory(context.Background(), assetTestUserID, "nonexistent", nil)
	require.NoError(t, err)
	assert.Nil(t, history)
}

func TestAssetGetAllocation(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	seedAsset(t, repo, "ast-liq-001", assetTestUserID, "liquid", 10000)
	seedAsset(t, repo, "ast-ret-001", assetTestUserID, "retirement", 40000)
	seedAsset(t, repo, "ast-re-001", assetTestUserID, "real_estate", 50000)

	alloc, err := repo.GetAllocation(context.Background(), assetTestUserID)
	require.NoError(t, err)
	assert.Len(t, alloc, 3)

	var totalValue float64
	for _, a := range alloc {
		totalValue += a.TotalValue
	}
	assert.Equal(t, 100000.0, totalValue)
}

func TestAssetGetAllocationOverTime(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	seedAsset(t, repo, "ast-liq-001", assetTestUserID, "liquid", 10000)
	seedAsset(t, repo, "ast-ret-001", assetTestUserID, "retirement", 40000)

	snapshots, err := repo.GetAllocationOverTime(context.Background(), assetTestUserID, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, snapshots)
}

func TestAssetWithAccountID(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	accountRepo := queries.NewAccountRepository(database)
	err := accountRepo.Create(context.Background(), &models.Account{
		ID:          "acc-test-001",
		UserID:      assetTestUserID,
		Name:        "Test Account",
		Institution: "Test Bank",
		AccountType: "checking",
		Currency:    "USD",
		IsActive:    true,
	})
	require.NoError(t, err)

	accountID := "acc-test-001"
	err = repo.Create(context.Background(), &models.Asset{
		ID:           "ast-test-001",
		UserID:       assetTestUserID,
		AccountID:    &accountID,
		Name:         "Linked Asset",
		AssetType:    "liquid",
		CurrentValue: 5000,
		Currency:     "USD",
	})
	require.NoError(t, err)

	asset, err := repo.GetByID(context.Background(), assetTestUserID, "ast-test-001")
	require.NoError(t, err)
	require.NotNil(t, asset)
	require.NotNil(t, asset.AccountID)
	assert.Equal(t, "acc-test-001", *asset.AccountID)
}

func TestGetLinkedAccounts_ReturnsAccountWithDebtBalance(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, original_balance) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"acc-mortgage-001", assetTestUserID, "Home Loan", "Bank", "mortgage", "USD", 300000.0,
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, priority_rank, linked_account_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"goal-mortgage-001", assetTestUserID, "Pay off mortgage", "debt_payoff", 300000, 250000, 1, "acc-mortgage-001",
	)
	require.NoError(t, err)

	result, err := repo.GetLinkedAccounts(context.Background(), []string{"acc-mortgage-001"})
	require.NoError(t, err)

	row, ok := result["acc-mortgage-001"]
	require.True(t, ok)
	assert.Equal(t, "Home Loan", row.Name)
	assert.Equal(t, "mortgage", row.AccountType)
	assert.Equal(t, "Bank", row.Institution)
	require.NotNil(t, row.Balance)
	assert.Equal(t, 250000.0, *row.Balance)
}

func TestGetLinkedAccounts_NonDebtAccountHasNilBalance(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"acc-ira-001", assetTestUserID, "Roth IRA", "Fidelity", "retirement_roth_ira", "USD",
	)
	require.NoError(t, err)

	result, err := repo.GetLinkedAccounts(context.Background(), []string{"acc-ira-001"})
	require.NoError(t, err)

	row, ok := result["acc-ira-001"]
	require.True(t, ok)
	assert.Equal(t, "Roth IRA", row.Name)
	assert.Nil(t, row.Balance)
}

func TestGetLinkedAccounts_EmptyInput(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	result, err := repo.GetLinkedAccounts(context.Background(), []string{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetLinkedAccounts_FallsBackToOriginalBalance(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, original_balance) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"acc-cc-001", assetTestUserID, "Credit Card", "Chase", "credit_card", "USD", 5000.0,
	)
	require.NoError(t, err)

	result, err := repo.GetLinkedAccounts(context.Background(), []string{"acc-cc-001"})
	require.NoError(t, err)

	row, ok := result["acc-cc-001"]
	require.True(t, ok)
	require.NotNil(t, row.Balance)
	assert.Equal(t, 5000.0, *row.Balance)
}

func TestAssetGetHistory_AnchorPoint(t *testing.T) {
	t.Parallel()
	database := setupAssetTestDB(t)
	repo := queries.NewAssetRepository(database)

	seedAsset(t, repo, "ast-anchor-001", assetTestUserID, "retirement", 10000)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"hist-old-001", "ast-anchor-001", 8000.0, "2024-06-01T00:00:00Z",
	)
	require.NoError(t, err)
	_, err = database.ExecContext(context.Background(),
		`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, ?)`,
		"hist-old-002", "ast-anchor-001", 9000.0, "2025-03-01T00:00:00Z",
	)
	require.NoError(t, err)

	since := time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC)
	history, err := repo.GetHistory(context.Background(), assetTestUserID, "ast-anchor-001", &since)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(history), 1)
	assert.Equal(t, 9000.0, history[0].Value)
	assert.Equal(t, "hist-old-002", history[0].ID)
}
