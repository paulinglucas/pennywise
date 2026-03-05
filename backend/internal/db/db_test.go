package db

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpen(t *testing.T) {
	ctx := context.Background()

	t.Run("creates database with WAL mode", func(t *testing.T) {
		db := openTestDB(t)

		var journalMode string
		err := db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&journalMode)
		require.NoError(t, err)
		assert.Equal(t, "wal", journalMode)
	})

	t.Run("enables foreign keys", func(t *testing.T) {
		db := openTestDB(t)

		var fkEnabled int
		err := db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&fkEnabled)
		require.NoError(t, err)
		assert.Equal(t, 1, fkEnabled)
	})

	t.Run("sets busy timeout", func(t *testing.T) {
		db := openTestDB(t)

		var timeout int
		err := db.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&timeout)
		require.NoError(t, err)
		assert.Equal(t, 5000, timeout)
	})

	t.Run("returns error for invalid path", func(t *testing.T) {
		_, err := Open("/nonexistent/dir/test.db")
		require.Error(t, err)
	})

	t.Run("creates file on disk", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "new.db")

		_, err := os.Stat(path)
		require.True(t, os.IsNotExist(err))

		db, err := Open(path)
		require.NoError(t, err)
		t.Cleanup(func() { _ = db.Close() })

		_, err = os.Stat(path)
		require.NoError(t, err)
	})
}

func TestMigrate(t *testing.T) {
	ctx := context.Background()

	t.Run("applies all migrations", func(t *testing.T) {
		db := openTestDB(t)

		err := Migrate(db)
		require.NoError(t, err)

		tables := queryTables(t, db)
		expectedTables := []string{
			"users",
			"accounts",
			"transactions",
			"transaction_tags",
			"assets",
			"asset_history",
			"goals",
			"recurring_transactions",
			"alerts",
			"audit_log",
			"failed_requests",
			"schema_migrations",
		}
		for _, expected := range expectedTables {
			assert.Contains(t, tables, expected, "missing table: %s", expected)
		}
	})

	t.Run("is idempotent", func(t *testing.T) {
		db := openTestDB(t)

		require.NoError(t, Migrate(db))
		require.NoError(t, Migrate(db))

		var count int
		err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 9, count)
	})

	t.Run("records applied migrations", func(t *testing.T) {
		db := openTestDB(t)

		require.NoError(t, Migrate(db))

		rows, err := db.QueryContext(ctx, "SELECT name FROM schema_migrations ORDER BY name")
		require.NoError(t, err)
		defer func() { _ = rows.Close() }()

		var names []string
		for rows.Next() {
			var name string
			require.NoError(t, rows.Scan(&name))
			names = append(names, name)
		}
		require.NoError(t, rows.Err())

		assert.Equal(t, []string{
			"001_create_users.sql",
			"002_create_accounts.sql",
			"003_create_transactions.sql",
			"004_create_assets.sql",
			"005_create_goals.sql",
			"006_create_recurring_transactions.sql",
			"007_create_alerts.sql",
			"008_create_audit_log.sql",
			"009_create_failed_requests.sql",
		}, names)
	})

	t.Run("creates required indexes", func(t *testing.T) {
		db := openTestDB(t)
		require.NoError(t, Migrate(db))

		indexes := queryIndexes(t, db)
		expectedIndexes := []string{
			"idx_transactions_user_date",
			"idx_transactions_user_category",
			"idx_transactions_user_account",
			"idx_transactions_deleted",
			"idx_assets_user_type",
			"idx_asset_history_asset_recorded",
			"idx_goals_user_priority",
			"idx_alerts_user_unread",
			"idx_failed_requests_created",
			"idx_audit_entity",
		}
		for _, expected := range expectedIndexes {
			assert.Contains(t, indexes, expected, "missing index: %s", expected)
		}
	})

	t.Run("enforces foreign key constraints", func(t *testing.T) {
		db := openTestDB(t)
		require.NoError(t, Migrate(db))

		_, err := db.ExecContext(ctx, `INSERT INTO accounts (id, user_id, name, institution, account_type)
			VALUES ('acc1', 'nonexistent', 'Test', 'Bank', 'checking')`)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "FOREIGN KEY constraint failed")
	})

	t.Run("users table has correct columns", func(t *testing.T) {
		db := openTestDB(t)
		require.NoError(t, Migrate(db))

		columns := queryColumns(t, db, "users")
		assert.ElementsMatch(t, []string{
			"id", "email", "name", "password_hash", "created_at", "updated_at",
		}, columns)
	})

	t.Run("accounts table has correct columns", func(t *testing.T) {
		db := openTestDB(t)
		require.NoError(t, Migrate(db))

		columns := queryColumns(t, db, "accounts")
		assert.ElementsMatch(t, []string{
			"id", "user_id", "name", "institution", "account_type",
			"currency", "is_active", "created_at", "updated_at", "deleted_at",
		}, columns)
	})

	t.Run("transactions table has correct columns", func(t *testing.T) {
		db := openTestDB(t)
		require.NoError(t, Migrate(db))

		columns := queryColumns(t, db, "transactions")
		assert.ElementsMatch(t, []string{
			"id", "user_id", "account_id", "type", "category", "amount",
			"currency", "date", "notes", "is_recurring", "recurring_transaction_id",
			"created_at", "updated_at", "deleted_at",
		}, columns)
	})

	t.Run("assets table has correct columns", func(t *testing.T) {
		db := openTestDB(t)
		require.NoError(t, Migrate(db))

		columns := queryColumns(t, db, "assets")
		assert.ElementsMatch(t, []string{
			"id", "user_id", "account_id", "name", "asset_type", "current_value",
			"currency", "metadata", "created_at", "updated_at", "deleted_at",
		}, columns)
	})

	t.Run("goals table has correct columns", func(t *testing.T) {
		db := openTestDB(t)
		require.NoError(t, Migrate(db))

		columns := queryColumns(t, db, "goals")
		assert.ElementsMatch(t, []string{
			"id", "user_id", "name", "goal_type", "target_amount", "current_amount",
			"deadline", "linked_account_id", "priority_rank",
			"created_at", "updated_at", "deleted_at",
		}, columns)
	})

	t.Run("recurring_transactions table has correct columns", func(t *testing.T) {
		db := openTestDB(t)
		require.NoError(t, Migrate(db))

		columns := queryColumns(t, db, "recurring_transactions")
		assert.ElementsMatch(t, []string{
			"id", "user_id", "account_id", "type", "category", "amount",
			"currency", "frequency", "next_occurrence", "is_active",
			"created_at", "updated_at", "deleted_at",
		}, columns)
	})

	t.Run("alerts table has correct columns", func(t *testing.T) {
		db := openTestDB(t)
		require.NoError(t, Migrate(db))

		columns := queryColumns(t, db, "alerts")
		assert.ElementsMatch(t, []string{
			"id", "user_id", "alert_type", "message", "severity",
			"is_read", "related_entity_type", "related_entity_id", "created_at",
		}, columns)
	})

	t.Run("audit_log table has correct columns", func(t *testing.T) {
		db := openTestDB(t)
		require.NoError(t, Migrate(db))

		columns := queryColumns(t, db, "audit_log")
		assert.ElementsMatch(t, []string{
			"id", "user_id", "entity_type", "entity_id", "action",
			"previous_data", "new_data", "created_at",
		}, columns)
	})

	t.Run("failed_requests table has correct columns", func(t *testing.T) {
		db := openTestDB(t)
		require.NoError(t, Migrate(db))

		columns := queryColumns(t, db, "failed_requests")
		assert.ElementsMatch(t, []string{
			"id", "request_id", "user_id", "method", "path", "status_code",
			"request_body", "request_headers", "error_code", "error_message",
			"stack_trace", "created_at",
		}, columns)
	})
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func queryTables(t *testing.T, db *sql.DB) []string {
	t.Helper()
	ctx := context.Background()
	rows, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'")
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	var tables []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		tables = append(tables, name)
	}
	require.NoError(t, rows.Err())
	return tables
}

func queryIndexes(t *testing.T, db *sql.DB) []string {
	t.Helper()
	ctx := context.Background()
	rows, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='index' AND name NOT LIKE 'sqlite_%'")
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	var indexes []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		indexes = append(indexes, name)
	}
	require.NoError(t, rows.Err())
	return indexes
}

func queryColumns(t *testing.T, db *sql.DB, table string) []string {
	t.Helper()
	ctx := context.Background()
	rows, err := db.QueryContext(ctx, "PRAGMA table_info("+table+")")
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	var columns []string
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue sql.NullString
		require.NoError(t, rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk))
		columns = append(columns, name)
	}
	require.NoError(t, rows.Err())
	return columns
}
