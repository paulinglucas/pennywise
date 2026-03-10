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

func seedGoalForContribution(t *testing.T, repo *queries.SQLiteGoalRepository, goalID, userID string) {
	t.Helper()
	err := repo.Create(context.Background(), &models.Goal{
		ID:            goalID,
		UserID:        userID,
		Name:          "Test Goal",
		GoalType:      "savings",
		TargetAmount:  10000,
		CurrentAmount: 0,
		PriorityRank:  1,
	})
	require.NoError(t, err)
}

func TestContributionCreate_And_List(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	goalRepo := queries.NewGoalRepository(database)
	contribRepo := queries.NewGoalContributionRepository(database)

	seedGoalForContribution(t, goalRepo, "goal-c-001", goalTestUserID)

	notes := "First contribution"
	contribAt := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	contrib := &models.GoalContribution{
		ID:            "contrib-001",
		GoalID:        "goal-c-001",
		UserID:        goalTestUserID,
		Amount:        500,
		Notes:         &notes,
		ContributedAt: contribAt,
	}
	err := contribRepo.Create(context.Background(), contrib)
	require.NoError(t, err)

	contribs, total, err := contribRepo.ListByGoal(context.Background(), goalTestUserID, "goal-c-001", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, contribs, 1)
	assert.Equal(t, "contrib-001", contribs[0].ID)
	assert.Equal(t, 500.0, contribs[0].Amount)
	assert.Equal(t, "First contribution", *contribs[0].Notes)
}

func TestContributionList_OrderedByDate(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	goalRepo := queries.NewGoalRepository(database)
	contribRepo := queries.NewGoalContributionRepository(database)

	seedGoalForContribution(t, goalRepo, "goal-c-001", goalTestUserID)

	earlier := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	later := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)

	require.NoError(t, contribRepo.Create(context.Background(), &models.GoalContribution{
		ID: "contrib-old", GoalID: "goal-c-001", UserID: goalTestUserID, Amount: 100, ContributedAt: earlier,
	}))
	require.NoError(t, contribRepo.Create(context.Background(), &models.GoalContribution{
		ID: "contrib-new", GoalID: "goal-c-001", UserID: goalTestUserID, Amount: 200, ContributedAt: later,
	}))

	contribs, _, err := contribRepo.ListByGoal(context.Background(), goalTestUserID, "goal-c-001", 1, 20)
	require.NoError(t, err)
	require.Len(t, contribs, 2)
	assert.Equal(t, "contrib-new", contribs[0].ID)
	assert.Equal(t, "contrib-old", contribs[1].ID)
}

func TestContributionList_Pagination(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	goalRepo := queries.NewGoalRepository(database)
	contribRepo := queries.NewGoalContributionRepository(database)

	seedGoalForContribution(t, goalRepo, "goal-c-001", goalTestUserID)

	for i := 0; i < 5; i++ {
		require.NoError(t, contribRepo.Create(context.Background(), &models.GoalContribution{
			ID: "contrib-" + string(rune('a'+i)), GoalID: "goal-c-001", UserID: goalTestUserID,
			Amount: float64((i + 1) * 100), ContributedAt: time.Date(2026, 1, i+1, 0, 0, 0, 0, time.UTC),
		}))
	}

	contribs, total, err := contribRepo.ListByGoal(context.Background(), goalTestUserID, "goal-c-001", 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, contribs, 2)
}

func TestContributionList_UserScoping(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	goalRepo := queries.NewGoalRepository(database)
	contribRepo := queries.NewGoalContributionRepository(database)

	seedGoalForContribution(t, goalRepo, "goal-c-001", goalTestUserID)

	require.NoError(t, contribRepo.Create(context.Background(), &models.GoalContribution{
		ID: "contrib-001", GoalID: "goal-c-001", UserID: goalTestUserID, Amount: 500,
		ContributedAt: time.Now(),
	}))

	contribs, total, err := contribRepo.ListByGoal(context.Background(), goalTestUser2ID, "goal-c-001", 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, contribs)
}

func TestContributionGetByID(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	goalRepo := queries.NewGoalRepository(database)
	contribRepo := queries.NewGoalContributionRepository(database)

	seedGoalForContribution(t, goalRepo, "goal-c-001", goalTestUserID)

	require.NoError(t, contribRepo.Create(context.Background(), &models.GoalContribution{
		ID: "contrib-001", GoalID: "goal-c-001", UserID: goalTestUserID, Amount: 750,
		ContributedAt: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
	}))

	contrib, err := contribRepo.GetByID(context.Background(), goalTestUserID, "contrib-001")
	require.NoError(t, err)
	require.NotNil(t, contrib)
	assert.Equal(t, 750.0, contrib.Amount)
	assert.Equal(t, "goal-c-001", contrib.GoalID)
}

func TestContributionGetByID_NotFound(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	contribRepo := queries.NewGoalContributionRepository(database)

	contrib, err := contribRepo.GetByID(context.Background(), goalTestUserID, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, contrib)
}

func TestContributionGetByID_WrongUser(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	goalRepo := queries.NewGoalRepository(database)
	contribRepo := queries.NewGoalContributionRepository(database)

	seedGoalForContribution(t, goalRepo, "goal-c-001", goalTestUserID)

	require.NoError(t, contribRepo.Create(context.Background(), &models.GoalContribution{
		ID: "contrib-001", GoalID: "goal-c-001", UserID: goalTestUserID, Amount: 500,
		ContributedAt: time.Now(),
	}))

	contrib, err := contribRepo.GetByID(context.Background(), goalTestUser2ID, "contrib-001")
	require.NoError(t, err)
	assert.Nil(t, contrib)
}

func TestContributionDelete(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	goalRepo := queries.NewGoalRepository(database)
	contribRepo := queries.NewGoalContributionRepository(database)

	seedGoalForContribution(t, goalRepo, "goal-c-001", goalTestUserID)

	require.NoError(t, contribRepo.Create(context.Background(), &models.GoalContribution{
		ID: "contrib-001", GoalID: "goal-c-001", UserID: goalTestUserID, Amount: 500,
		ContributedAt: time.Now(),
	}))

	found, err := contribRepo.Delete(context.Background(), goalTestUserID, "contrib-001")
	require.NoError(t, err)
	assert.True(t, found)

	contrib, err := contribRepo.GetByID(context.Background(), goalTestUserID, "contrib-001")
	require.NoError(t, err)
	assert.Nil(t, contrib)
}

func TestContributionDelete_NotFound(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	contribRepo := queries.NewGoalContributionRepository(database)

	found, err := contribRepo.Delete(context.Background(), goalTestUserID, "nonexistent")
	require.NoError(t, err)
	assert.False(t, found)
}

func TestContributionDelete_WrongUser(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	goalRepo := queries.NewGoalRepository(database)
	contribRepo := queries.NewGoalContributionRepository(database)

	seedGoalForContribution(t, goalRepo, "goal-c-001", goalTestUserID)

	require.NoError(t, contribRepo.Create(context.Background(), &models.GoalContribution{
		ID: "contrib-001", GoalID: "goal-c-001", UserID: goalTestUserID, Amount: 500,
		ContributedAt: time.Now(),
	}))

	found, err := contribRepo.Delete(context.Background(), goalTestUser2ID, "contrib-001")
	require.NoError(t, err)
	assert.False(t, found)
}

func TestContributionCreate_NilNotes(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	goalRepo := queries.NewGoalRepository(database)
	contribRepo := queries.NewGoalContributionRepository(database)

	seedGoalForContribution(t, goalRepo, "goal-c-001", goalTestUserID)

	require.NoError(t, contribRepo.Create(context.Background(), &models.GoalContribution{
		ID: "contrib-001", GoalID: "goal-c-001", UserID: goalTestUserID, Amount: 250,
		ContributedAt: time.Now(),
	}))

	contrib, err := contribRepo.GetByID(context.Background(), goalTestUserID, "contrib-001")
	require.NoError(t, err)
	require.NotNil(t, contrib)
	assert.Nil(t, contrib.Notes)
	assert.Equal(t, 250.0, contrib.Amount)
}

func seedAccountAndTransaction(t *testing.T, database *sql.DB, txnID, userID string) {
	t.Helper()
	ctx := context.Background()
	_, err := database.ExecContext(ctx,
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"acct-contrib-01", userID, "Checking", "Chase", "checking", "USD", 1,
	)
	require.NoError(t, err)

	_, err = database.ExecContext(ctx,
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		txnID, userID, "acct-contrib-01", "expense", "savings", 500.0, "USD", "2026-03-01",
	)
	require.NoError(t, err)
}

func TestContributionCreate_WithTransactionID(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	goalRepo := queries.NewGoalRepository(database)
	contribRepo := queries.NewGoalContributionRepository(database)

	seedGoalForContribution(t, goalRepo, "goal-c-001", goalTestUserID)
	seedAccountAndTransaction(t, database, "txn-001", goalTestUserID)

	txnID := "txn-001"
	require.NoError(t, contribRepo.Create(context.Background(), &models.GoalContribution{
		ID: "contrib-001", GoalID: "goal-c-001", UserID: goalTestUserID, Amount: 500,
		TransactionID: &txnID,
		ContributedAt: time.Now(),
	}))

	contrib, err := contribRepo.GetByID(context.Background(), goalTestUserID, "contrib-001")
	require.NoError(t, err)
	require.NotNil(t, contrib)
	require.NotNil(t, contrib.TransactionID)
	assert.Equal(t, "txn-001", *contrib.TransactionID)
}

func TestContributionCreate_WithoutTransactionID(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	goalRepo := queries.NewGoalRepository(database)
	contribRepo := queries.NewGoalContributionRepository(database)

	seedGoalForContribution(t, goalRepo, "goal-c-001", goalTestUserID)

	require.NoError(t, contribRepo.Create(context.Background(), &models.GoalContribution{
		ID: "contrib-001", GoalID: "goal-c-001", UserID: goalTestUserID, Amount: 250,
		ContributedAt: time.Now(),
	}))

	contrib, err := contribRepo.GetByID(context.Background(), goalTestUserID, "contrib-001")
	require.NoError(t, err)
	require.NotNil(t, contrib)
	assert.Nil(t, contrib.TransactionID)
}
