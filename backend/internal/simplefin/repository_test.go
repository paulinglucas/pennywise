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
