package simplefin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pennywisecrypto "github.com/jamespsullivan/pennywise/internal/crypto"
	"github.com/jamespsullivan/pennywise/internal/db"
)

func testAccessURL(serverURL string) string {
	parsed, _ := url.Parse(serverURL + "/simplefin")
	parsed.User = url.UserPassword("testuser", "testpass")
	return parsed.String()
}

func setupSyncTest(t *testing.T) (*SQLiteSimplefinRepository, *SyncService, *httptest.Server) {
	t.Helper()

	database, err := db.Open(":memory:")
	require.NoError(t, err)
	require.NoError(t, db.Migrate(database))
	t.Cleanup(func() { _ = database.Close() })

	ctx := context.Background()
	_, err = database.ExecContext(ctx, `INSERT INTO users (id, email, name, password_hash) VALUES ('u1', 'test@test.com', 'Test', 'hash')`)
	require.NoError(t, err)

	_, err = database.ExecContext(ctx, `INSERT INTO accounts (id, user_id, name, institution, account_type, simplefin_id) VALUES
		('a1', 'u1', 'Checking', 'My Bank', 'checking', 'sfin-001'),
		('a2', 'u1', 'Savings', 'My Bank', 'savings', 'sfin-002'),
		('a3', 'u1', 'Unlinked', 'Other Bank', 'checking', NULL)`)
	require.NoError(t, err)

	_, err = database.ExecContext(ctx, `INSERT INTO assets (id, user_id, account_id, name, asset_type, current_value) VALUES
		('asset1', 'u1', 'a1', 'Checking Balance', 'liquid', 1000.00),
		('asset2', 'u1', 'a2', 'Savings Balance', 'liquid', 5000.00)`)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"errors": [],
			"accounts": [
				{"id": "sfin-001", "name": "Checking", "currency": "USD", "balance": "1500.00", "balance-date": 1709856000, "org": {"name": "My Bank", "id": "mybank", "domain": "mybank.com"}},
				{"id": "sfin-002", "name": "Savings", "currency": "USD", "balance": "5200.50", "balance-date": 1709856000, "org": {"name": "My Bank", "id": "mybank", "domain": "mybank.com"}},
				{"id": "sfin-003", "name": "Credit Card", "currency": "USD", "balance": "-450.00", "balance-date": 1709856000, "org": {"name": "My Bank", "id": "mybank", "domain": "mybank.com"}}
			]
		}`))
	}))
	t.Cleanup(server.Close)

	repo := NewSimplefinRepository(database)
	encKey := pennywisecrypto.DeriveKey("test-key")

	client := NewClient(nil)
	svc := NewSyncService(client, repo, encKey)

	return repo, svc, server
}

func TestSyncUpdatesLinkedAssets(t *testing.T) {
	repo, svc, server := setupSyncTest(t)
	ctx := context.Background()

	result, err := svc.SyncUser(ctx, "u1", testAccessURL(server.URL))
	require.NoError(t, err)
	assert.Equal(t, 2, result.Updated)
	assert.Equal(t, 0, result.Errors)

	asset1, err := repo.GetAssetForAccount(ctx, "u1", "a1")
	require.NoError(t, err)
	assert.Equal(t, 1500.0, asset1.CurrentValue)

	asset2, err := repo.GetAssetForAccount(ctx, "u1", "a2")
	require.NoError(t, err)
	assert.Equal(t, 5200.50, asset2.CurrentValue)
}

func TestSyncSkipsUnlinkedAccounts(t *testing.T) {
	_, svc, server := setupSyncTest(t)
	ctx := context.Background()

	result, err := svc.SyncUser(ctx, "u1", testAccessURL(server.URL))
	require.NoError(t, err)
	assert.Equal(t, 2, result.Updated)
}

func TestSyncSkipsUnchangedValues(t *testing.T) {
	repo, svc, server := setupSyncTest(t)
	ctx := context.Background()

	err := repo.UpdateAssetValue(ctx, "asset1", 1500.0)
	require.NoError(t, err)
	err = repo.UpdateAssetValue(ctx, "asset2", 5200.50)
	require.NoError(t, err)

	result, err := svc.SyncUser(ctx, "u1", testAccessURL(server.URL))
	require.NoError(t, err)
	assert.Equal(t, 0, result.Updated)
}

func TestRunSyncForAllUsers(t *testing.T) {
	repo, svc, server := setupSyncTest(t)
	ctx := context.Background()

	encryptedURL, err := pennywisecrypto.Encrypt(svc.encryptionKey, testAccessURL(server.URL))
	require.NoError(t, err)
	err = repo.SaveConnection(ctx, "u1", encryptedURL)
	require.NoError(t, err)

	results := svc.SyncAll(ctx)
	require.Len(t, results, 1)
	assert.Equal(t, "u1", results[0].UserID)
	assert.Equal(t, 2, results[0].Updated)
	assert.NoError(t, results[0].Error)
}

func TestSyncAllDecryptError(t *testing.T) {
	repo, svc, _ := setupSyncTest(t)
	ctx := context.Background()

	err := repo.SaveConnection(ctx, "u1", "not-valid-encrypted-data")
	require.NoError(t, err)

	results := svc.SyncAll(ctx)
	require.Len(t, results, 1)
	assert.Error(t, results[0].Error)
}

func TestSyncAllFetchError(t *testing.T) {
	repo, svc, _ := setupSyncTest(t)
	ctx := context.Background()

	accessURL := "http://testuser:testpass@localhost:1/simplefin" //nolint:gosec // test data
	encrypted, err := pennywisecrypto.Encrypt(svc.encryptionKey, accessURL)
	require.NoError(t, err)
	err = repo.SaveConnection(ctx, "u1", encrypted)
	require.NoError(t, err)

	results := svc.SyncAll(ctx)
	require.Len(t, results, 1)
	assert.Error(t, results[0].Error)
}

func TestSyncUserBadBalance(t *testing.T) {
	database, err := db.Open(":memory:")
	require.NoError(t, err)
	require.NoError(t, db.Migrate(database))
	t.Cleanup(func() { _ = database.Close() })

	ctx := context.Background()
	_, err = database.ExecContext(ctx, `INSERT INTO users (id, email, name, password_hash) VALUES ('u1', 'test@test.com', 'Test', 'hash')`)
	require.NoError(t, err)
	_, err = database.ExecContext(ctx, `INSERT INTO accounts (id, user_id, name, institution, account_type, simplefin_id) VALUES ('a1', 'u1', 'Checking', 'Bank', 'checking', 'sfin-001')`)
	require.NoError(t, err)
	_, err = database.ExecContext(ctx, `INSERT INTO assets (id, user_id, account_id, name, asset_type, current_value) VALUES ('asset1', 'u1', 'a1', 'Checking', 'liquid', 1000.00)`)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[],"accounts":[{"id":"sfin-001","name":"Checking","currency":"USD","balance":"not-a-number","balance-date":1709856000,"org":{"name":"Bank","id":"bank","domain":"bank.com"}}]}`))
	}))
	t.Cleanup(server.Close)

	repo := NewSimplefinRepository(database)
	encKey := pennywisecrypto.DeriveKey("test-key")
	client := NewClient(nil)
	svc := NewSyncService(client, repo, encKey)

	result, err := svc.SyncUser(ctx, "u1", testAccessURL(server.URL))
	require.NoError(t, err)
	assert.Equal(t, 0, result.Updated)
	assert.Equal(t, 1, result.Errors)
}

func TestSyncUserLinkedAccountNoAsset(t *testing.T) {
	database, err := db.Open(":memory:")
	require.NoError(t, err)
	require.NoError(t, db.Migrate(database))
	t.Cleanup(func() { _ = database.Close() })

	ctx := context.Background()
	_, err = database.ExecContext(ctx, `INSERT INTO users (id, email, name, password_hash) VALUES ('u1', 'test@test.com', 'Test', 'hash')`)
	require.NoError(t, err)
	_, err = database.ExecContext(ctx, `INSERT INTO accounts (id, user_id, name, institution, account_type, simplefin_id) VALUES ('a1', 'u1', 'Checking', 'Bank', 'checking', 'sfin-001')`)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[],"accounts":[{"id":"sfin-001","name":"Checking","currency":"USD","balance":"1500.00","balance-date":1709856000,"org":{"name":"Bank","id":"bank","domain":"bank.com"}}]}`))
	}))
	t.Cleanup(server.Close)

	repo := NewSimplefinRepository(database)
	encKey := pennywisecrypto.DeriveKey("test-key")
	client := NewClient(nil)
	svc := NewSyncService(client, repo, encKey)

	result, err := svc.SyncUser(ctx, "u1", testAccessURL(server.URL))
	require.NoError(t, err)
	assert.Equal(t, 0, result.Updated)
	assert.Equal(t, 0, result.Errors)
}

func TestSyncUserNoLinkedAccounts(t *testing.T) {
	database, err := db.Open(":memory:")
	require.NoError(t, err)
	require.NoError(t, db.Migrate(database))
	t.Cleanup(func() { _ = database.Close() })

	ctx := context.Background()
	_, err = database.ExecContext(ctx, `INSERT INTO users (id, email, name, password_hash) VALUES ('u1', 'test@test.com', 'Test', 'hash')`)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[],"accounts":[{"id":"sfin-001","name":"Checking","currency":"USD","balance":"1500.00","balance-date":1709856000,"org":{"name":"Bank","id":"bank","domain":"bank.com"}}]}`))
	}))
	t.Cleanup(server.Close)

	repo := NewSimplefinRepository(database)
	encKey := pennywisecrypto.DeriveKey("test-key")
	client := NewClient(nil)
	svc := NewSyncService(client, repo, encKey)

	result, err := svc.SyncUser(ctx, "u1", testAccessURL(server.URL))
	require.NoError(t, err)
	assert.Equal(t, 0, result.Updated)
}
