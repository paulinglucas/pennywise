package queries_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/db"
	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/models"
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
	t.Parallel()
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
	t.Parallel()
	database := setupTestDB(t)
	repo := queries.NewUserRepository(database)

	user, err := repo.GetByEmail(context.Background(), "nobody@example.com")

	assert.ErrorIs(t, err, queries.ErrUserNotFound)
	assert.Nil(t, user)
}

func TestGetByID_ExistingUser(t *testing.T) {
	t.Parallel()
	database := setupTestDB(t)
	repo := queries.NewUserRepository(database)

	user, err := repo.GetByID(context.Background(), "usr00001-0000-0000-0000-000000000001")

	require.NoError(t, err)
	assert.Equal(t, "james@example.com", user.Email)
	assert.Equal(t, "James", user.Name)
}

func TestGetByID_NotFound(t *testing.T) {
	t.Parallel()
	database := setupTestDB(t)
	repo := queries.NewUserRepository(database)

	user, err := repo.GetByID(context.Background(), "nonexistent-id")

	assert.ErrorIs(t, err, queries.ErrUserNotFound)
	assert.Nil(t, user)
}

func TestCountUsers(t *testing.T) {
	t.Parallel()
	database := setupTestDB(t)
	repo := queries.NewUserRepository(database)

	count, err := repo.CountUsers(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCreateUser_Success(t *testing.T) {
	t.Parallel()
	database := setupTestDB(t)
	repo := queries.NewUserRepository(database)

	user := &models.User{
		ID:           "usr00002-0000-0000-0000-000000000002",
		Email:        "newuser@example.com",
		Name:         "New User",
		PasswordHash: "$2a$10$somehashedpass",
	}

	err := repo.CreateUser(context.Background(), user)
	require.NoError(t, err)

	count, err := repo.CountUsers(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	found, err := repo.GetByEmail(context.Background(), "newuser@example.com")
	require.NoError(t, err)
	assert.Equal(t, "New User", found.Name)
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	t.Parallel()
	database := setupTestDB(t)
	repo := queries.NewUserRepository(database)

	user := &models.User{
		ID:           "usr00003-0000-0000-0000-000000000003",
		Email:        "james@example.com",
		Name:         "Duplicate",
		PasswordHash: "$2a$10$somehashedpass",
	}

	err := repo.CreateUser(context.Background(), user)
	assert.ErrorIs(t, err, queries.ErrEmailTaken)
}
