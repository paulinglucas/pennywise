package queries_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/db"
	"github.com/jamespsullivan/pennywise/internal/db/queries"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	err = db.Migrate(database)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		"usr00001-0000-0000-0000-000000000001",
		"james@example.com",
		"James",
		"$2a$10$hashedpassword",
	)
	require.NoError(t, err)

	return database
}

func TestGetByEmail_ExistingUser(t *testing.T) {
	database := setupTestDB(t)
	repo := queries.NewUserRepository(database)

	user, err := repo.GetByEmail(context.Background(), "james@example.com")

	require.NoError(t, err)
	assert.Equal(t, "usr00001-0000-0000-0000-000000000001", user.ID)
	assert.Equal(t, "james@example.com", user.Email)
	assert.Equal(t, "James", user.Name)
	assert.Equal(t, "$2a$10$hashedpassword", user.PasswordHash)
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())
}

func TestGetByEmail_NotFound(t *testing.T) {
	database := setupTestDB(t)
	repo := queries.NewUserRepository(database)

	user, err := repo.GetByEmail(context.Background(), "nobody@example.com")

	assert.ErrorIs(t, err, queries.ErrUserNotFound)
	assert.Nil(t, user)
}

func TestGetByID_ExistingUser(t *testing.T) {
	database := setupTestDB(t)
	repo := queries.NewUserRepository(database)

	user, err := repo.GetByID(context.Background(), "usr00001-0000-0000-0000-000000000001")

	require.NoError(t, err)
	assert.Equal(t, "james@example.com", user.Email)
	assert.Equal(t, "James", user.Name)
}

func TestGetByID_NotFound(t *testing.T) {
	database := setupTestDB(t)
	repo := queries.NewUserRepository(database)

	user, err := repo.GetByID(context.Background(), "nonexistent-id")

	assert.ErrorIs(t, err, queries.ErrUserNotFound)
	assert.Nil(t, user)
}
