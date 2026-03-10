package simplefin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pennywisecrypto "github.com/jamespsullivan/pennywise/internal/crypto"
	"github.com/jamespsullivan/pennywise/internal/db"
)

func TestSchedulerStartStop(t *testing.T) {
	database, err := db.Open(":memory:")
	require.NoError(t, err)
	require.NoError(t, db.Migrate(database))
	t.Cleanup(func() { _ = database.Close() })

	repo := NewSimplefinRepository(database)
	encKey := pennywisecrypto.DeriveKey("test-key")
	client := NewClient(nil)
	svc := NewSyncService(client, repo, encKey)

	scheduler := NewScheduler(svc, 6)
	assert.NotNil(t, scheduler)

	scheduler.Start()
	scheduler.Stop()
}

func TestExecuteSync(t *testing.T) {
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
		_, _ = w.Write([]byte(`{"errors":[],"accounts":[{"id":"sfin-001","name":"Checking","currency":"USD","balance":"1500.00","balance-date":1709856000,"org":{"name":"Bank","id":"bank","domain":"bank.com"}}]}`))
	}))
	t.Cleanup(server.Close)

	repo := NewSimplefinRepository(database)
	encKey := pennywisecrypto.DeriveKey("test-key")
	client := NewClient(nil)
	svc := NewSyncService(client, repo, encKey)

	accessURL := testAccessURL(server.URL)
	encrypted, err := pennywisecrypto.Encrypt(encKey, accessURL)
	require.NoError(t, err)
	require.NoError(t, repo.SaveConnection(ctx, "u1", encrypted))

	scheduler := NewScheduler(svc, 6)
	scheduler.executeSync()

	asset, err := repo.GetAssetForAccount(ctx, "u1", "a1")
	require.NoError(t, err)
	assert.Equal(t, 1500.0, asset.CurrentValue)
}
