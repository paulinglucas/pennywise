package queries_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/db/queries"
	"github.com/jamespsullivan/pennywise/internal/models"
)

func setupGroupTestDB(t *testing.T) (*queries.SQLiteTransactionGroupRepository, *queries.SQLiteTransactionRepository) {
	t.Helper()
	database := setupTransactionTestDB(t)
	groupRepo := queries.NewTransactionGroupRepository(database)
	txnRepo := queries.NewTransactionRepository(database)
	return groupRepo, txnRepo
}

func TestGroupCreate_And_GetByID(t *testing.T) {
	repo, _ := setupGroupTestDB(t)

	group := &models.TransactionGroup{
		ID:     "grp-test-001",
		UserID: testUserID,
		Name:   "March Paycheck",
	}

	err := repo.Create(context.Background(), group)
	require.NoError(t, err)

	fetched, err := repo.GetByID(context.Background(), testUserID, group.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)

	assert.Equal(t, group.ID, fetched.ID)
	assert.Equal(t, group.UserID, fetched.UserID)
	assert.Equal(t, "March Paycheck", fetched.Name)
	assert.False(t, fetched.CreatedAt.IsZero())
	assert.False(t, fetched.UpdatedAt.IsZero())
}

func TestGroupGetByID_NotFound(t *testing.T) {
	repo, _ := setupGroupTestDB(t)

	fetched, err := repo.GetByID(context.Background(), testUserID, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestGroupGetByID_WrongUser(t *testing.T) {
	repo, _ := setupGroupTestDB(t)

	group := &models.TransactionGroup{
		ID:     "grp-test-002",
		UserID: testUserID,
		Name:   "My Group",
	}
	require.NoError(t, repo.Create(context.Background(), group))

	fetched, err := repo.GetByID(context.Background(), testUser2ID, group.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestGroupUpdate(t *testing.T) {
	repo, _ := setupGroupTestDB(t)

	group := &models.TransactionGroup{
		ID:     "grp-test-003",
		UserID: testUserID,
		Name:   "Old Name",
	}
	require.NoError(t, repo.Create(context.Background(), group))

	group.Name = "New Name"
	found, err := repo.Update(context.Background(), group)
	require.NoError(t, err)
	assert.True(t, found)

	fetched, err := repo.GetByID(context.Background(), testUserID, group.ID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", fetched.Name)
}

func TestGroupSoftDelete(t *testing.T) {
	repo, _ := setupGroupTestDB(t)

	group := &models.TransactionGroup{
		ID:     "grp-test-004",
		UserID: testUserID,
		Name:   "To Delete",
	}
	require.NoError(t, repo.Create(context.Background(), group))

	found, err := repo.SoftDelete(context.Background(), testUserID, group.ID)
	require.NoError(t, err)
	assert.True(t, found)

	fetched, err := repo.GetByID(context.Background(), testUserID, group.ID)
	require.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestGroupSoftDelete_WrongUser(t *testing.T) {
	repo, _ := setupGroupTestDB(t)

	group := &models.TransactionGroup{
		ID:     "grp-test-005",
		UserID: testUserID,
		Name:   "Not Yours",
	}
	require.NoError(t, repo.Create(context.Background(), group))

	found, err := repo.SoftDelete(context.Background(), testUser2ID, group.ID)
	require.NoError(t, err)
	assert.False(t, found)
}

func TestGroupList(t *testing.T) {
	repo, _ := setupGroupTestDB(t)

	for i := range 3 {
		require.NoError(t, repo.Create(context.Background(), &models.TransactionGroup{
			ID:     "grp-list-" + string(rune('a'+i)),
			UserID: testUserID,
			Name:   "Group " + string(rune('A'+i)),
		}))
	}

	groups, total, err := repo.List(context.Background(), testUserID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, groups, 3)
}

func TestGroupList_Pagination(t *testing.T) {
	repo, _ := setupGroupTestDB(t)

	for i := range 5 {
		require.NoError(t, repo.Create(context.Background(), &models.TransactionGroup{
			ID:     "grp-page-" + string(rune('a'+i)),
			UserID: testUserID,
			Name:   "Group " + string(rune('A'+i)),
		}))
	}

	groups, total, err := repo.List(context.Background(), testUserID, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, groups, 2)

	groups2, _, err := repo.List(context.Background(), testUserID, 2, 2)
	require.NoError(t, err)
	assert.Len(t, groups2, 2)
}

func TestGroupList_UserScoped(t *testing.T) {
	repo, _ := setupGroupTestDB(t)

	require.NoError(t, repo.Create(context.Background(), &models.TransactionGroup{
		ID:     "grp-scope-1",
		UserID: testUserID,
		Name:   "User1 Group",
	}))
	require.NoError(t, repo.Create(context.Background(), &models.TransactionGroup{
		ID:     "grp-scope-2",
		UserID: testUser2ID,
		Name:   "User2 Group",
	}))

	groups, total, err := repo.List(context.Background(), testUserID, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, groups, 1)
	assert.Equal(t, "User1 Group", groups[0].Name)
}

func TestGroupMembers(t *testing.T) {
	groupRepo, txnRepo := setupGroupTestDB(t)

	group := &models.TransactionGroup{
		ID:     "grp-members-001",
		UserID: testUserID,
		Name:   "Paycheck Split",
	}
	require.NoError(t, groupRepo.Create(context.Background(), group))

	date, _ := time.Parse("2006-01-02", "2025-06-15")
	groupID := group.ID

	txn1 := &models.Transaction{
		ID:        "txn-member-001",
		UserID:    testUserID,
		AccountID: testAccountID,
		Type:      "deposit",
		Category:  "salary",
		Amount:    4000,
		Currency:  "USD",
		Date:      date,
		GroupID:   &groupID,
	}
	require.NoError(t, txnRepo.Create(context.Background(), txn1, nil))

	txn2 := &models.Transaction{
		ID:        "txn-member-002",
		UserID:    testUserID,
		AccountID: testAccountID,
		Type:      "deposit",
		Category:  "401k",
		Amount:    500,
		Currency:  "USD",
		Date:      date,
		GroupID:   &groupID,
	}
	require.NoError(t, txnRepo.Create(context.Background(), txn2, nil))

	members, err := groupRepo.ListMembers(context.Background(), testUserID, group.ID)
	require.NoError(t, err)
	assert.Len(t, members, 2)

	var totalAmount float64
	for _, m := range members {
		totalAmount += m.Amount
	}
	assert.InDelta(t, 4500.0, totalAmount, 0.01)
}

func TestGroupSoftDelete_CascadesToMembers(t *testing.T) {
	groupRepo, txnRepo := setupGroupTestDB(t)

	group := &models.TransactionGroup{
		ID:     "grp-cascade-001",
		UserID: testUserID,
		Name:   "To Cascade Delete",
	}
	require.NoError(t, groupRepo.Create(context.Background(), group))

	date, _ := time.Parse("2006-01-02", "2025-06-15")
	groupID := group.ID

	require.NoError(t, txnRepo.Create(context.Background(), &models.Transaction{
		ID:        "txn-cascade-001",
		UserID:    testUserID,
		AccountID: testAccountID,
		Type:      "expense",
		Category:  "test",
		Amount:    100,
		Currency:  "USD",
		Date:      date,
		GroupID:   &groupID,
	}, nil))

	found, err := groupRepo.SoftDelete(context.Background(), testUserID, group.ID)
	require.NoError(t, err)
	assert.True(t, found)

	txn, err := txnRepo.GetByID(context.Background(), testUserID, "txn-cascade-001")
	require.NoError(t, err)
	assert.Nil(t, txn)
}
