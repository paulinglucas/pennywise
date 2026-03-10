package queries_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/models"
)

func TestAuditLogRecord(t *testing.T) {
	t.Parallel()
	database := setupTestDB(t)
	repo := queries.NewAuditLogRepository(database)

	prevData := `{"amount":10}`
	newData := `{"amount":20}`
	err := repo.Record(context.Background(), &models.AuditLog{
		UserID:       "usr00001-0000-0000-0000-000000000001",
		EntityType:   "transaction",
		EntityID:     "txn-001",
		Action:       "update",
		PreviousData: &prevData,
		NewData:      &newData,
	})
	require.NoError(t, err)

	var count int
	err = database.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM audit_log WHERE entity_type = 'transaction' AND entity_id = 'txn-001'",
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestAuditLogRecord_CreateAction(t *testing.T) {
	t.Parallel()
	database := setupTestDB(t)
	repo := queries.NewAuditLogRepository(database)

	newData := `{"id":"txn-001","amount":42.99}`
	err := repo.Record(context.Background(), &models.AuditLog{
		UserID:     "usr00001-0000-0000-0000-000000000001",
		EntityType: "transaction",
		EntityID:   "txn-001",
		Action:     "create",
		NewData:    &newData,
	})
	require.NoError(t, err)

	var action, storedNewData string
	err = database.QueryRowContext(context.Background(),
		"SELECT action, new_data FROM audit_log WHERE entity_id = 'txn-001'",
	).Scan(&action, &storedNewData)
	require.NoError(t, err)
	assert.Equal(t, "create", action)
	assert.Equal(t, newData, storedNewData)
}
