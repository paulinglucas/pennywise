package queries_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/models"
)

func setupAccountTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database := setupTestDB(t)
	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		"usr00002-0000-0000-0000-000000000002", "alex@example.com", "Alex", "$2a$10$hashedpassword",
	)
	require.NoError(t, err)
	return database
}

func seedAccount(t *testing.T, repo *queries.SQLiteAccountRepository, id, userID string) {
	t.Helper()
	err := repo.Create(context.Background(), &models.Account{
		ID:          id,
		UserID:      userID,
		Name:        "Test Checking",
		Institution: "Test Bank",
		AccountType: "checking",
		Currency:    "USD",
		IsActive:    true,
	})
	require.NoError(t, err)
}

func TestAccountCreate_And_GetByID(t *testing.T) {
	t.Parallel()
	database := setupAccountTestDB(t)
	repo := queries.NewAccountRepository(database)

	seedAccount(t, repo, "acc-test-001", "usr00001-0000-0000-0000-000000000001")

	account, err := repo.GetByID(context.Background(), "usr00001-0000-0000-0000-000000000001", "acc-test-001")
	require.NoError(t, err)
	assert.Equal(t, "acc-test-001", account.ID)
	assert.Equal(t, "Test Checking", account.Name)
	assert.Equal(t, "Test Bank", account.Institution)
	assert.Equal(t, "checking", account.AccountType)
	assert.Equal(t, "USD", account.Currency)
	assert.True(t, account.IsActive)
}

func TestAccountGetByID_NotFound(t *testing.T) {
	t.Parallel()
	database := setupAccountTestDB(t)
	repo := queries.NewAccountRepository(database)

	account, err := repo.GetByID(context.Background(), "usr00001-0000-0000-0000-000000000001", "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, account)
}

func TestAccountGetByID_WrongUser(t *testing.T) {
	t.Parallel()
	database := setupAccountTestDB(t)
	repo := queries.NewAccountRepository(database)

	seedAccount(t, repo, "acc-test-001", "usr00001-0000-0000-0000-000000000001")

	account, err := repo.GetByID(context.Background(), "other-user", "acc-test-001")
	require.NoError(t, err)
	assert.Nil(t, account)
}

func TestAccountList_FiltersByUser(t *testing.T) {
	t.Parallel()
	database := setupAccountTestDB(t)
	repo := queries.NewAccountRepository(database)

	seedAccount(t, repo, "acc-user1-001", "usr00001-0000-0000-0000-000000000001")
	seedAccount(t, repo, "acc-user1-002", "usr00001-0000-0000-0000-000000000001")
	seedAccount(t, repo, "acc-user2-001", "usr00002-0000-0000-0000-000000000002")

	accounts, total, err := repo.List(context.Background(), "usr00001-0000-0000-0000-000000000001", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, accounts, 2)
}

func TestAccountList_Pagination(t *testing.T) {
	t.Parallel()
	database := setupAccountTestDB(t)
	repo := queries.NewAccountRepository(database)

	for i := range 5 {
		seedAccount(t, repo, fmt.Sprintf("acc-page-%03d", i+1), "usr00001-0000-0000-0000-000000000001")
	}

	accounts, total, err := repo.List(context.Background(), "usr00001-0000-0000-0000-000000000001", 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, accounts, 2)

	accounts2, _, err := repo.List(context.Background(), "usr00001-0000-0000-0000-000000000001", 2, 2)
	require.NoError(t, err)
	assert.Len(t, accounts2, 2)

	accounts3, _, err := repo.List(context.Background(), "usr00001-0000-0000-0000-000000000001", 3, 2)
	require.NoError(t, err)
	assert.Len(t, accounts3, 1)
}

func TestAccountUpdate(t *testing.T) {
	t.Parallel()
	database := setupAccountTestDB(t)
	repo := queries.NewAccountRepository(database)

	seedAccount(t, repo, "acc-test-001", "usr00001-0000-0000-0000-000000000001")

	account, err := repo.GetByID(context.Background(), "usr00001-0000-0000-0000-000000000001", "acc-test-001")
	require.NoError(t, err)

	account.Name = "Updated Name"
	account.Institution = "Updated Bank"
	found, err := repo.Update(context.Background(), account)
	require.NoError(t, err)
	assert.True(t, found)

	updated, err := repo.GetByID(context.Background(), "usr00001-0000-0000-0000-000000000001", "acc-test-001")
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated Bank", updated.Institution)
}

func TestAccountUpdate_WrongUser(t *testing.T) {
	t.Parallel()
	database := setupAccountTestDB(t)
	repo := queries.NewAccountRepository(database)

	seedAccount(t, repo, "acc-test-001", "usr00001-0000-0000-0000-000000000001")

	found, err := repo.Update(context.Background(), &models.Account{
		ID:          "acc-test-001",
		UserID:      "other-user",
		Name:        "Hacked",
		Institution: "Hacked Bank",
		AccountType: "checking",
		Currency:    "USD",
		IsActive:    true,
	})
	require.NoError(t, err)
	assert.False(t, found)
}

func TestAccountSoftDelete(t *testing.T) {
	t.Parallel()
	database := setupAccountTestDB(t)
	repo := queries.NewAccountRepository(database)

	seedAccount(t, repo, "acc-test-001", "usr00001-0000-0000-0000-000000000001")

	found, err := repo.SoftDelete(context.Background(), "usr00001-0000-0000-0000-000000000001", "acc-test-001")
	require.NoError(t, err)
	assert.True(t, found)

	account, err := repo.GetByID(context.Background(), "usr00001-0000-0000-0000-000000000001", "acc-test-001")
	require.NoError(t, err)
	assert.Nil(t, account)
}

func TestAccountSoftDelete_ExcludedFromList(t *testing.T) {
	t.Parallel()
	database := setupAccountTestDB(t)
	repo := queries.NewAccountRepository(database)

	seedAccount(t, repo, "acc-test-001", "usr00001-0000-0000-0000-000000000001")
	seedAccount(t, repo, "acc-test-002", "usr00001-0000-0000-0000-000000000001")

	found, err := repo.SoftDelete(context.Background(), "usr00001-0000-0000-0000-000000000001", "acc-test-001")
	require.NoError(t, err)
	assert.True(t, found)

	accounts, total, err := repo.List(context.Background(), "usr00001-0000-0000-0000-000000000001", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, accounts, 1)
	assert.Equal(t, "acc-test-002", accounts[0].ID)
}

func TestAccountSoftDelete_WrongUser(t *testing.T) {
	t.Parallel()
	database := setupAccountTestDB(t)
	repo := queries.NewAccountRepository(database)

	seedAccount(t, repo, "acc-test-001", "usr00001-0000-0000-0000-000000000001")

	found, err := repo.SoftDelete(context.Background(), "other-user", "acc-test-001")
	require.NoError(t, err)
	assert.False(t, found)
}
