package queries_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/db/queries"
)

const alertTestUserID = "usr00001-0000-0000-0000-000000000001"

func TestAlertList_Empty(t *testing.T) {
	t.Parallel()
	database := setupTestDB(t)
	repo := queries.NewAlertRepository(database)

	items, total, err := repo.List(context.Background(), alertTestUserID, 1, 10)
	require.NoError(t, err)
	assert.Empty(t, items)
	assert.Equal(t, 0, total)
}

func TestAlertList_ReturnsUnread(t *testing.T) {
	t.Parallel()
	database := setupTestDB(t)
	repo := queries.NewAlertRepository(database)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO alerts (id, user_id, alert_type, message, severity, is_read) VALUES (?, ?, ?, ?, ?, ?)`,
		"alt00001-0000-0000-0000-000000000001", alertTestUserID, "spending_spike", "Food spending up 20%", "warning", 0,
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO alerts (id, user_id, alert_type, message, severity, is_read) VALUES (?, ?, ?, ?, ?, ?)`,
		"alt00002-0000-0000-0000-000000000002", alertTestUserID, "goal_progress", "Emergency fund 50% complete", "info", 1,
	)
	require.NoError(t, err)

	items, total, err := repo.List(context.Background(), alertTestUserID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, 1, total)
	assert.Equal(t, "spending_spike", items[0].AlertType)
}

func TestAlertList_UserScoping(t *testing.T) {
	t.Parallel()
	database := setupTestDB(t)
	repo := queries.NewAlertRepository(database)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		"usr00002-0000-0000-0000-000000000002", "alex@example.com", "Alex", "$2a$10$dummy",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO alerts (id, user_id, alert_type, message, severity, is_read) VALUES (?, ?, ?, ?, ?, ?)`,
		"alt00001-0000-0000-0000-000000000001", alertTestUserID, "info", "My alert", "info", 0,
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO alerts (id, user_id, alert_type, message, severity, is_read) VALUES (?, ?, ?, ?, ?, ?)`,
		"alt00002-0000-0000-0000-000000000002", "usr00002-0000-0000-0000-000000000002", "info", "Alex alert", "info", 0,
	)
	require.NoError(t, err)

	items, total, err := repo.List(context.Background(), alertTestUserID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, 1, total)
}

func TestAlertMarkRead(t *testing.T) {
	t.Parallel()
	database := setupTestDB(t)
	repo := queries.NewAlertRepository(database)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO alerts (id, user_id, alert_type, message, severity, is_read) VALUES (?, ?, ?, ?, ?, ?)`,
		"alt00001-0000-0000-0000-000000000001", alertTestUserID, "info", "Test alert", "info", 0,
	)
	require.NoError(t, err)

	found, err := repo.MarkRead(context.Background(), alertTestUserID, "alt00001-0000-0000-0000-000000000001")
	require.NoError(t, err)
	assert.True(t, found)

	items, _, err := repo.List(context.Background(), alertTestUserID, 1, 10)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestAlertMarkRead_NotFound(t *testing.T) {
	t.Parallel()
	database := setupTestDB(t)
	repo := queries.NewAlertRepository(database)

	found, err := repo.MarkRead(context.Background(), alertTestUserID, "nonexistent")
	require.NoError(t, err)
	assert.False(t, found)
}

func TestAlertMarkRead_WrongUser(t *testing.T) {
	t.Parallel()
	database := setupTestDB(t)
	repo := queries.NewAlertRepository(database)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		"usr00002-0000-0000-0000-000000000002", "alex@example.com", "Alex", "$2a$10$dummy",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO alerts (id, user_id, alert_type, message, severity, is_read) VALUES (?, ?, ?, ?, ?, ?)`,
		"alt00001-0000-0000-0000-000000000001", alertTestUserID, "info", "Test alert", "info", 0,
	)
	require.NoError(t, err)

	found, err := repo.MarkRead(context.Background(), "usr00002-0000-0000-0000-000000000002", "alt00001-0000-0000-0000-000000000001")
	require.NoError(t, err)
	assert.False(t, found)
}

func TestAlertList_Pagination(t *testing.T) {
	t.Parallel()
	database := setupTestDB(t)
	repo := queries.NewAlertRepository(database)

	for i := 0; i < 3; i++ {
		_, err := database.ExecContext(context.Background(),
			`INSERT INTO alerts (id, user_id, alert_type, message, severity, is_read) VALUES (?, ?, ?, ?, ?, ?)`,
			"alt0000"+string(rune('a'+i))+"-0000-0000-0000-000000000001", alertTestUserID, "info", "Alert", "info", 0,
		)
		require.NoError(t, err)
	}

	items, total, err := repo.List(context.Background(), alertTestUserID, 1, 2)
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, 3, total)
}
