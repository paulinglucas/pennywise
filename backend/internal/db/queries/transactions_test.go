package queries_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/models"
)

const testUserID = "usr00001-0000-0000-0000-000000000001"
const testUser2ID = "usr00002-0000-0000-0000-000000000002"
const testAccountID = "acc00001-0000-0000-0000-000000000001"
const testAccount2ID = "acc00002-0000-0000-0000-000000000002"

func setupTransactionTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database := setupTestDB(t)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		testUser2ID, "alex@example.com", "Alex", "$2a$10$hashedpassword",
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		testAccountID, testUserID, "Checking", "Chase", "checking", "USD", 1,
	)
	require.NoError(t, err)

	_, err = database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		testAccount2ID, testUser2ID, "Savings", "BofA", "savings", "USD", 1,
	)
	require.NoError(t, err)

	return database
}

func seedTransaction(t *testing.T, repo *queries.SQLiteTransactionRepository, id, userID, accountID string) {
	t.Helper()
	date, _ := time.Parse("2006-01-02", "2025-06-15")
	err := repo.Create(context.Background(), &models.Transaction{
		ID:        id,
		UserID:    userID,
		AccountID: accountID,
		Type:      "expense",
		Category:  "food",
		Amount:    25.50,
		Currency:  "USD",
		Date:      date,
	}, nil)
	require.NoError(t, err)
}

func TestTransactionCreate_And_GetByID(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	date, _ := time.Parse("2006-01-02", "2025-06-15")
	txn := &models.Transaction{
		ID:        "txn-test-001",
		UserID:    testUserID,
		AccountID: testAccountID,
		Type:      "expense",
		Category:  "food",
		Amount:    42.99,
		Currency:  "USD",
		Date:      date,
	}
	notes := "Lunch at deli"
	txn.Notes = &notes

	err := repo.Create(context.Background(), txn, []string{"lunch", "work"})
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), testUserID, "txn-test-001")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "txn-test-001", got.ID)
	assert.Equal(t, testUserID, got.UserID)
	assert.Equal(t, testAccountID, got.AccountID)
	assert.Equal(t, "expense", got.Type)
	assert.Equal(t, "food", got.Category)
	assert.InDelta(t, 42.99, got.Amount, 0.001)
	assert.Equal(t, "USD", got.Currency)
	assert.Equal(t, "Lunch at deli", *got.Notes)
	assert.ElementsMatch(t, []string{"lunch", "work"}, got.Tags)
}

func TestTransactionGetByID_NotFound(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	got, err := repo.GetByID(context.Background(), testUserID, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestTransactionGetByID_WrongUser(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	seedTransaction(t, repo, "txn-test-001", testUserID, testAccountID)

	got, err := repo.GetByID(context.Background(), testUser2ID, "txn-test-001")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestTransactionList_FiltersByUser(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	seedTransaction(t, repo, "txn-user1-001", testUserID, testAccountID)
	seedTransaction(t, repo, "txn-user1-002", testUserID, testAccountID)
	seedTransaction(t, repo, "txn-user2-001", testUser2ID, testAccount2ID)

	txns, total, err := repo.List(context.Background(), testUserID, queries.TransactionFilter{}, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, txns, 2)
}

func TestTransactionList_Pagination(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	for i := range 5 {
		seedTransaction(t, repo, fmt.Sprintf("txn-page-%03d", i+1), testUserID, testAccountID)
	}

	txns, total, err := repo.List(context.Background(), testUserID, queries.TransactionFilter{}, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, txns, 2)

	txns2, _, err := repo.List(context.Background(), testUserID, queries.TransactionFilter{}, 2, 2)
	require.NoError(t, err)
	assert.Len(t, txns2, 2)

	txns3, _, err := repo.List(context.Background(), testUserID, queries.TransactionFilter{}, 3, 2)
	require.NoError(t, err)
	assert.Len(t, txns3, 1)
}

func TestTransactionList_FilterByCategory(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	date, _ := time.Parse("2006-01-02", "2025-06-15")
	for _, cat := range []string{"food", "food", "transport"} {
		id := fmt.Sprintf("txn-%s-%d", cat, time.Now().UnixNano())
		err := repo.Create(context.Background(), &models.Transaction{
			ID: id, UserID: testUserID, AccountID: testAccountID,
			Type: "expense", Category: cat, Amount: 10, Currency: "USD", Date: date,
		}, nil)
		require.NoError(t, err)
	}

	category := "food"
	txns, total, err := repo.List(context.Background(), testUserID, queries.TransactionFilter{Category: &category}, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, txns, 2)
}

func TestTransactionList_FilterByAccount(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"acc-other-001", testUserID, "Savings", "Chase", "savings", "USD", 1,
	)
	require.NoError(t, err)

	seedTransaction(t, repo, "txn-acc1-001", testUserID, testAccountID)
	seedTransaction(t, repo, "txn-acc2-001", testUserID, "acc-other-001")

	accountID := testAccountID
	txns, total, err := repo.List(context.Background(), testUserID, queries.TransactionFilter{AccountID: &accountID}, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, txns, 1)
	assert.Equal(t, testAccountID, txns[0].AccountID)
}

func TestTransactionList_FilterByDateRange(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	dates := []string{"2025-05-01", "2025-06-15", "2025-07-20"}
	for i, d := range dates {
		date, _ := time.Parse("2006-01-02", d)
		err := repo.Create(context.Background(), &models.Transaction{
			ID: fmt.Sprintf("txn-date-%d", i), UserID: testUserID, AccountID: testAccountID,
			Type: "expense", Category: "food", Amount: 10, Currency: "USD", Date: date,
		}, nil)
		require.NoError(t, err)
	}

	from, _ := time.Parse("2006-01-02", "2025-06-01")
	to, _ := time.Parse("2006-01-02", "2025-06-30")
	txns, total, err := repo.List(context.Background(), testUserID, queries.TransactionFilter{
		DateFrom: &from,
		DateTo:   &to,
	}, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, txns, 1)
}

func TestTransactionList_FilterByAmountRange(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	date, _ := time.Parse("2006-01-02", "2025-06-15")
	for _, amt := range []float64{5.00, 25.50, 100.00} {
		err := repo.Create(context.Background(), &models.Transaction{
			ID: fmt.Sprintf("txn-amt-%.0f", amt), UserID: testUserID, AccountID: testAccountID,
			Type: "expense", Category: "food", Amount: amt, Currency: "USD", Date: date,
		}, nil)
		require.NoError(t, err)
	}

	minAmt := float64(10)
	maxAmt := float64(50)
	txns, total, err := repo.List(context.Background(), testUserID, queries.TransactionFilter{
		AmountMin: &minAmt,
		AmountMax: &maxAmt,
	}, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, txns, 1)
	assert.InDelta(t, 25.50, txns[0].Amount, 0.001)
}

func TestTransactionList_FilterByType(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	date, _ := time.Parse("2006-01-02", "2025-06-15")
	err := repo.Create(context.Background(), &models.Transaction{
		ID: "txn-exp", UserID: testUserID, AccountID: testAccountID,
		Type: "expense", Category: "food", Amount: 10, Currency: "USD", Date: date,
	}, nil)
	require.NoError(t, err)

	err = repo.Create(context.Background(), &models.Transaction{
		ID: "txn-dep", UserID: testUserID, AccountID: testAccountID,
		Type: "deposit", Category: "salary", Amount: 5000, Currency: "USD", Date: date,
	}, nil)
	require.NoError(t, err)

	txnType := "deposit"
	txns, total, err := repo.List(context.Background(), testUserID, queries.TransactionFilter{Type: &txnType}, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, txns, 1)
	assert.Equal(t, "deposit", txns[0].Type)
}

func TestTransactionList_FilterByTags(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	date, _ := time.Parse("2006-01-02", "2025-06-15")
	err := repo.Create(context.Background(), &models.Transaction{
		ID: "txn-tagged", UserID: testUserID, AccountID: testAccountID,
		Type: "expense", Category: "food", Amount: 10, Currency: "USD", Date: date,
	}, []string{"lunch", "work"})
	require.NoError(t, err)

	err = repo.Create(context.Background(), &models.Transaction{
		ID: "txn-untagged", UserID: testUserID, AccountID: testAccountID,
		Type: "expense", Category: "food", Amount: 20, Currency: "USD", Date: date,
	}, nil)
	require.NoError(t, err)

	tags := []string{"lunch"}
	txns, total, err := repo.List(context.Background(), testUserID, queries.TransactionFilter{Tags: tags}, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, txns, 1)
	assert.Equal(t, "txn-tagged", txns[0].ID)
}

func TestTransactionList_Search(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	date, _ := time.Parse("2006-01-02", "2025-06-15")
	notes1 := "Chipotle burrito"
	err := repo.Create(context.Background(), &models.Transaction{
		ID: "txn-search-1", UserID: testUserID, AccountID: testAccountID,
		Type: "expense", Category: "food", Amount: 10, Currency: "USD", Date: date, Notes: &notes1,
	}, nil)
	require.NoError(t, err)

	notes2 := "Gas station"
	err = repo.Create(context.Background(), &models.Transaction{
		ID: "txn-search-2", UserID: testUserID, AccountID: testAccountID,
		Type: "expense", Category: "transport", Amount: 40, Currency: "USD", Date: date, Notes: &notes2,
	}, nil)
	require.NoError(t, err)

	search := "chipotle"
	txns, total, err := repo.List(context.Background(), testUserID, queries.TransactionFilter{Search: &search}, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, txns, 1)
	assert.Equal(t, "txn-search-1", txns[0].ID)
}

func TestTransactionUpdate(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	seedTransaction(t, repo, "txn-test-001", testUserID, testAccountID)

	txn, err := repo.GetByID(context.Background(), testUserID, "txn-test-001")
	require.NoError(t, err)

	txn.Category = "dining"
	txn.Amount = 35.00
	found, err := repo.Update(context.Background(), txn, []string{"updated-tag"})
	require.NoError(t, err)
	assert.True(t, found)

	updated, err := repo.GetByID(context.Background(), testUserID, "txn-test-001")
	require.NoError(t, err)
	assert.Equal(t, "dining", updated.Category)
	assert.InDelta(t, 35.00, updated.Amount, 0.001)
	assert.ElementsMatch(t, []string{"updated-tag"}, updated.Tags)
}

func TestTransactionSoftDelete(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	seedTransaction(t, repo, "txn-test-001", testUserID, testAccountID)

	found, err := repo.SoftDelete(context.Background(), testUserID, "txn-test-001")
	require.NoError(t, err)
	assert.True(t, found)

	got, err := repo.GetByID(context.Background(), testUserID, "txn-test-001")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestTransactionSoftDelete_ExcludedFromList(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	seedTransaction(t, repo, "txn-test-001", testUserID, testAccountID)
	seedTransaction(t, repo, "txn-test-002", testUserID, testAccountID)

	found, err := repo.SoftDelete(context.Background(), testUserID, "txn-test-001")
	require.NoError(t, err)
	assert.True(t, found)

	txns, total, err := repo.List(context.Background(), testUserID, queries.TransactionFilter{}, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, txns, 1)
	assert.Equal(t, "txn-test-002", txns[0].ID)
}

func TestTransactionSoftDelete_WrongUser(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	seedTransaction(t, repo, "txn-test-001", testUserID, testAccountID)

	found, err := repo.SoftDelete(context.Background(), testUser2ID, "txn-test-001")
	require.NoError(t, err)
	assert.False(t, found)
}

func TestTransactionBulkCreate(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	date, _ := time.Parse("2006-01-02", "2025-06-15")
	txns := []models.Transaction{
		{ID: "txn-bulk-001", UserID: testUserID, AccountID: testAccountID, Type: "expense", Category: "food", Amount: 10, Currency: "USD", Date: date},
		{ID: "txn-bulk-002", UserID: testUserID, AccountID: testAccountID, Type: "expense", Category: "transport", Amount: 20, Currency: "USD", Date: date},
		{ID: "txn-bulk-003", UserID: testUserID, AccountID: testAccountID, Type: "deposit", Category: "salary", Amount: 5000, Currency: "USD", Date: date},
	}

	imported, errors := repo.BulkCreate(context.Background(), txns)
	assert.Equal(t, 3, imported)
	assert.Empty(t, errors)

	list, total, err := repo.List(context.Background(), testUserID, queries.TransactionFilter{}, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, list, 3)
}

func TestTransactionCreate_WithTags_TagsReturned(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	date, _ := time.Parse("2006-01-02", "2025-06-15")
	err := repo.Create(context.Background(), &models.Transaction{
		ID: "txn-tags", UserID: testUserID, AccountID: testAccountID,
		Type: "expense", Category: "food", Amount: 10, Currency: "USD", Date: date,
	}, []string{"work", "lunch"})
	require.NoError(t, err)

	got, err := repo.GetByID(context.Background(), testUserID, "txn-tags")
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"work", "lunch"}, got.Tags)
}

func TestTransactionList_TagsLoadedOnListItems(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	date, _ := time.Parse("2006-01-02", "2025-06-15")
	err := repo.Create(context.Background(), &models.Transaction{
		ID: "txn-list-tags", UserID: testUserID, AccountID: testAccountID,
		Type: "expense", Category: "food", Amount: 10, Currency: "USD", Date: date,
	}, []string{"tag-a", "tag-b"})
	require.NoError(t, err)

	txns, _, err := repo.List(context.Background(), testUserID, queries.TransactionFilter{}, 1, 20)
	require.NoError(t, err)
	require.Len(t, txns, 1)
	assert.ElementsMatch(t, []string{"tag-a", "tag-b"}, txns[0].Tags)
}

func TestTransactionList_PaginationBeyondRange(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	seedTransaction(t, repo, "txn-test-001", testUserID, testAccountID)

	txns, total, err := repo.List(context.Background(), testUserID, queries.TransactionFilter{}, 10, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Empty(t, txns)
}

func TestListCategories_Empty(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	categories, err := repo.ListCategories(context.Background(), testUserID)
	require.NoError(t, err)
	assert.Empty(t, categories)
}

func TestListCategories_FromTransactions(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	date, _ := time.Parse("2006-01-02", "2026-01-15")
	for i, cat := range []string{"food", "rent", "food"} {
		require.NoError(t, repo.Create(context.Background(), &models.Transaction{
			ID: fmt.Sprintf("txn-cat-%d", i), UserID: testUserID, AccountID: testAccountID,
			Type: "expense", Category: cat, Amount: 10, Currency: "USD", Date: date,
		}, nil))
	}

	categories, err := repo.ListCategories(context.Background(), testUserID)
	require.NoError(t, err)
	assert.Equal(t, []string{"food", "rent"}, categories)
}

func TestListCategories_IncludesRecurring(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	date, _ := time.Parse("2006-01-02", "2026-01-15")
	require.NoError(t, repo.Create(context.Background(), &models.Transaction{
		ID: "txn-cat-r1", UserID: testUserID, AccountID: testAccountID,
		Type: "expense", Category: "food", Amount: 10, Currency: "USD", Date: date,
	}, nil))

	_, err := database.ExecContext(context.Background(),
		`INSERT INTO recurring_transactions (id, user_id, account_id, type, category, amount, currency, frequency, next_occurrence)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"rec-cat-1", testUserID, testAccountID, "expense", "utilities", 100, "USD", "monthly", "2026-02-01",
	)
	require.NoError(t, err)

	categories, err := repo.ListCategories(context.Background(), testUserID)
	require.NoError(t, err)
	assert.Equal(t, []string{"food", "utilities"}, categories)
}

func TestListCategories_ExcludesDeleted(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	date, _ := time.Parse("2006-01-02", "2026-01-15")
	require.NoError(t, repo.Create(context.Background(), &models.Transaction{
		ID: "txn-cat-d1", UserID: testUserID, AccountID: testAccountID,
		Type: "expense", Category: "food", Amount: 10, Currency: "USD", Date: date,
	}, nil))
	require.NoError(t, repo.Create(context.Background(), &models.Transaction{
		ID: "txn-cat-d2", UserID: testUserID, AccountID: testAccountID,
		Type: "expense", Category: "deleted_cat", Amount: 10, Currency: "USD", Date: date,
	}, nil))

	_, err := repo.SoftDelete(context.Background(), testUserID, "txn-cat-d2")
	require.NoError(t, err)

	categories, err := repo.ListCategories(context.Background(), testUserID)
	require.NoError(t, err)
	assert.Equal(t, []string{"food"}, categories)
}

func TestListCategories_UserScoped(t *testing.T) {
	t.Parallel()
	database := setupTransactionTestDB(t)
	repo := queries.NewTransactionRepository(database)

	date, _ := time.Parse("2006-01-02", "2026-01-15")
	require.NoError(t, repo.Create(context.Background(), &models.Transaction{
		ID: "txn-cat-u1", UserID: testUserID, AccountID: testAccountID,
		Type: "expense", Category: "food", Amount: 10, Currency: "USD", Date: date,
	}, nil))
	require.NoError(t, repo.Create(context.Background(), &models.Transaction{
		ID: "txn-cat-u2", UserID: testUser2ID, AccountID: testAccount2ID,
		Type: "expense", Category: "other_user_cat", Amount: 10, Currency: "USD", Date: date,
	}, nil))

	categories, err := repo.ListCategories(context.Background(), testUserID)
	require.NoError(t, err)
	assert.Equal(t, []string{"food"}, categories)
}
