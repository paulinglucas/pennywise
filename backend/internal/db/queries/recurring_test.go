package queries_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/models"
)

const (
	recurringTestUserID  = "usr00001-0000-0000-0000-000000000001"
	recurringTestUser2ID = "usr00002-0000-0000-0000-000000000002"
	recurringTestAcctID  = "acc00001-0000-0000-0000-000000000001"
)

func setupRecurringTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database := setupTestDB(t)
	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		recurringTestUser2ID, "alex@example.com", "Alex", "$2a$10$hashedpassword",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		recurringTestAcctID, recurringTestUserID, "Checking", "Bank", "checking", "USD",
	)
	require.NoError(t, err)

	return database
}

func TestRecurringCreate(t *testing.T) {
	t.Parallel()
	database := setupRecurringTestDB(t)
	repo := queries.NewRecurringRepository(database)

	rec := &models.RecurringTransaction{
		ID:             "rec00001-0000-0000-0000-000000000001",
		UserID:         recurringTestUserID,
		AccountID:      recurringTestAcctID,
		Type:           "expense",
		Category:       "housing",
		Amount:         1500,
		Currency:       "USD",
		Frequency:      "monthly",
		NextOccurrence: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		IsActive:       true,
	}
	err := repo.Create(context.Background(), rec)
	require.NoError(t, err)

	found, err := repo.GetByID(context.Background(), recurringTestUserID, rec.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "housing", found.Category)
	assert.Equal(t, 1500.0, found.Amount)
	assert.Equal(t, "monthly", found.Frequency)
	assert.True(t, found.IsActive)
}

func TestRecurringGetByID_NotFound(t *testing.T) {
	t.Parallel()
	database := setupRecurringTestDB(t)
	repo := queries.NewRecurringRepository(database)

	found, err := repo.GetByID(context.Background(), recurringTestUserID, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestRecurringGetByID_WrongUser(t *testing.T) {
	t.Parallel()
	database := setupRecurringTestDB(t)
	repo := queries.NewRecurringRepository(database)

	rec := &models.RecurringTransaction{
		ID:             "rec00001-0000-0000-0000-000000000001",
		UserID:         recurringTestUserID,
		AccountID:      recurringTestAcctID,
		Type:           "expense",
		Category:       "housing",
		Amount:         1500,
		Currency:       "USD",
		Frequency:      "monthly",
		NextOccurrence: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
		IsActive:       true,
	}
	require.NoError(t, repo.Create(context.Background(), rec))

	found, err := repo.GetByID(context.Background(), recurringTestUser2ID, rec.ID)
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestRecurringList(t *testing.T) {
	t.Parallel()
	database := setupRecurringTestDB(t)
	repo := queries.NewRecurringRepository(database)

	for i, cat := range []string{"housing", "utilities", "insurance"} {
		rec := &models.RecurringTransaction{
			ID:             "rec0000" + string(rune('1'+i)) + "-0000-0000-0000-000000000001",
			UserID:         recurringTestUserID,
			AccountID:      recurringTestAcctID,
			Type:           "expense",
			Category:       cat,
			Amount:         float64((i + 1) * 500),
			Currency:       "USD",
			Frequency:      "monthly",
			NextOccurrence: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			IsActive:       true,
		}
		require.NoError(t, repo.Create(context.Background(), rec))
	}

	items, total, err := repo.List(context.Background(), recurringTestUserID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, items, 3)
	assert.Equal(t, 3, total)
}

func TestRecurringList_Pagination(t *testing.T) {
	t.Parallel()
	database := setupRecurringTestDB(t)
	repo := queries.NewRecurringRepository(database)

	for i := 0; i < 3; i++ {
		rec := &models.RecurringTransaction{
			ID:             "rec0000" + string(rune('a'+i)) + "-0000-0000-0000-000000000001",
			UserID:         recurringTestUserID,
			AccountID:      recurringTestAcctID,
			Type:           "expense",
			Category:       "cat",
			Amount:         100,
			Currency:       "USD",
			Frequency:      "monthly",
			NextOccurrence: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			IsActive:       true,
		}
		require.NoError(t, repo.Create(context.Background(), rec))
	}

	items, total, err := repo.List(context.Background(), recurringTestUserID, 1, 2)
	require.NoError(t, err)
	assert.Len(t, items, 2)
	assert.Equal(t, 3, total)
}

func TestRecurringList_UserScoping(t *testing.T) {
	t.Parallel()
	database := setupRecurringTestDB(t)
	repo := queries.NewRecurringRepository(database)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency) VALUES (?, ?, ?, ?, ?, ?)`,
		"acc00002-0000-0000-0000-000000000002", recurringTestUser2ID, "Other Acct", "OtherBank", "checking", "USD",
	)
	require.NoError(t, err)

	rec1 := &models.RecurringTransaction{
		ID: "rec00001-0000-0000-0000-000000000001", UserID: recurringTestUserID, AccountID: recurringTestAcctID,
		Type: "expense", Category: "housing", Amount: 1500, Currency: "USD", Frequency: "monthly",
		NextOccurrence: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), IsActive: true,
	}
	require.NoError(t, repo.Create(context.Background(), rec1))

	rec2 := &models.RecurringTransaction{
		ID: "rec00002-0000-0000-0000-000000000002", UserID: recurringTestUser2ID, AccountID: "acc00002-0000-0000-0000-000000000002",
		Type: "expense", Category: "food", Amount: 200, Currency: "USD", Frequency: "weekly",
		NextOccurrence: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), IsActive: true,
	}
	require.NoError(t, repo.Create(context.Background(), rec2))

	items, total, err := repo.List(context.Background(), recurringTestUserID, 1, 10)
	require.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, 1, total)
}

func TestRecurringUpdate(t *testing.T) {
	t.Parallel()
	database := setupRecurringTestDB(t)
	repo := queries.NewRecurringRepository(database)

	rec := &models.RecurringTransaction{
		ID: "rec00001-0000-0000-0000-000000000001", UserID: recurringTestUserID, AccountID: recurringTestAcctID,
		Type: "expense", Category: "housing", Amount: 1500, Currency: "USD", Frequency: "monthly",
		NextOccurrence: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), IsActive: true,
	}
	require.NoError(t, repo.Create(context.Background(), rec))

	rec.Amount = 1600
	rec.Category = "rent"
	found, err := repo.Update(context.Background(), rec)
	require.NoError(t, err)
	assert.True(t, found)

	updated, err := repo.GetByID(context.Background(), recurringTestUserID, rec.ID)
	require.NoError(t, err)
	assert.Equal(t, 1600.0, updated.Amount)
	assert.Equal(t, "rent", updated.Category)
}

func TestRecurringUpdate_WrongUser(t *testing.T) {
	t.Parallel()
	database := setupRecurringTestDB(t)
	repo := queries.NewRecurringRepository(database)

	rec := &models.RecurringTransaction{
		ID: "rec00001-0000-0000-0000-000000000001", UserID: recurringTestUserID, AccountID: recurringTestAcctID,
		Type: "expense", Category: "housing", Amount: 1500, Currency: "USD", Frequency: "monthly",
		NextOccurrence: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), IsActive: true,
	}
	require.NoError(t, repo.Create(context.Background(), rec))

	rec.UserID = recurringTestUser2ID
	found, err := repo.Update(context.Background(), rec)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestRecurringSoftDelete(t *testing.T) {
	t.Parallel()
	database := setupRecurringTestDB(t)
	repo := queries.NewRecurringRepository(database)

	rec := &models.RecurringTransaction{
		ID: "rec00001-0000-0000-0000-000000000001", UserID: recurringTestUserID, AccountID: recurringTestAcctID,
		Type: "expense", Category: "housing", Amount: 1500, Currency: "USD", Frequency: "monthly",
		NextOccurrence: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), IsActive: true,
	}
	require.NoError(t, repo.Create(context.Background(), rec))

	found, err := repo.SoftDelete(context.Background(), recurringTestUserID, rec.ID)
	require.NoError(t, err)
	assert.True(t, found)

	got, err := repo.GetByID(context.Background(), recurringTestUserID, rec.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestRecurringSoftDelete_ExcludedFromList(t *testing.T) {
	t.Parallel()
	database := setupRecurringTestDB(t)
	repo := queries.NewRecurringRepository(database)

	rec := &models.RecurringTransaction{
		ID: "rec00001-0000-0000-0000-000000000001", UserID: recurringTestUserID, AccountID: recurringTestAcctID,
		Type: "expense", Category: "housing", Amount: 1500, Currency: "USD", Frequency: "monthly",
		NextOccurrence: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC), IsActive: true,
	}
	require.NoError(t, repo.Create(context.Background(), rec))

	_, err := repo.SoftDelete(context.Background(), recurringTestUserID, rec.ID)
	require.NoError(t, err)

	items, total, err := repo.List(context.Background(), recurringTestUserID, 1, 10)
	require.NoError(t, err)
	assert.Empty(t, items)
	assert.Equal(t, 0, total)
}
