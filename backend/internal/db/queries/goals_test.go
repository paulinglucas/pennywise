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
	goalTestUserID  = "usr00001-0000-0000-0000-000000000001"
	goalTestUser2ID = "usr00002-0000-0000-0000-000000000002"
)

func setupGoalTestDB(t *testing.T) *sql.DB {
	t.Helper()
	database := setupTestDB(t)
	_, err := database.ExecContext(context.Background(),
		`INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)`,
		goalTestUser2ID, "alex@example.com", "Alex", "$2a$10$hashedpassword",
	)
	require.NoError(t, err)
	return database
}

func seedGoal(t *testing.T, repo *queries.SQLiteGoalRepository, id, userID, goalType string, target, current float64, rank int) {
	t.Helper()
	err := repo.Create(context.Background(), &models.Goal{
		ID:            id,
		UserID:        userID,
		Name:          "Goal " + id,
		GoalType:      goalType,
		TargetAmount:  target,
		CurrentAmount: current,
		PriorityRank:  rank,
	})
	require.NoError(t, err)
}

func TestGoalCreate_And_GetByID(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	repo := queries.NewGoalRepository(database)

	deadline := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	err := repo.Create(context.Background(), &models.Goal{
		ID:            "goal-test-001",
		UserID:        goalTestUserID,
		Name:          "Emergency Fund",
		GoalType:      "savings",
		TargetAmount:  10000,
		CurrentAmount: 3000,
		Deadline:      &deadline,
		PriorityRank:  1,
	})
	require.NoError(t, err)

	goal, err := repo.GetByID(context.Background(), goalTestUserID, "goal-test-001")
	require.NoError(t, err)
	require.NotNil(t, goal)
	assert.Equal(t, "goal-test-001", goal.ID)
	assert.Equal(t, "Emergency Fund", goal.Name)
	assert.Equal(t, "savings", goal.GoalType)
	assert.Equal(t, 10000.0, goal.TargetAmount)
	assert.Equal(t, 3000.0, goal.CurrentAmount)
	assert.NotNil(t, goal.Deadline)
	assert.Equal(t, 2027, goal.Deadline.Year())
	assert.Equal(t, 1, goal.PriorityRank)
}

func TestGoalCreate_DebtPayoff(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	repo := queries.NewGoalRepository(database)

	err := repo.Create(context.Background(), &models.Goal{
		ID:            "goal-test-001",
		UserID:        goalTestUserID,
		Name:          "Pay Off Credit Card",
		GoalType:      "debt_payoff",
		TargetAmount:  5000,
		CurrentAmount: 3500,
		PriorityRank:  1,
	})
	require.NoError(t, err)

	goal, err := repo.GetByID(context.Background(), goalTestUserID, "goal-test-001")
	require.NoError(t, err)
	require.NotNil(t, goal)
	assert.Equal(t, "debt_payoff", goal.GoalType)
	assert.Nil(t, goal.Deadline)
}

func TestGoalGetByID_NotFound(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	repo := queries.NewGoalRepository(database)

	goal, err := repo.GetByID(context.Background(), goalTestUserID, "nonexistent")
	require.NoError(t, err)
	assert.Nil(t, goal)
}

func TestGoalGetByID_WrongUser(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	repo := queries.NewGoalRepository(database)

	seedGoal(t, repo, "goal-test-001", goalTestUserID, "savings", 10000, 0, 1)

	goal, err := repo.GetByID(context.Background(), goalTestUser2ID, "goal-test-001")
	require.NoError(t, err)
	assert.Nil(t, goal)
}

func TestGoalList_OrderedByPriority(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	repo := queries.NewGoalRepository(database)

	seedGoal(t, repo, "goal-rank-3", goalTestUserID, "savings", 10000, 0, 3)
	seedGoal(t, repo, "goal-rank-1", goalTestUserID, "debt_payoff", 5000, 0, 1)
	seedGoal(t, repo, "goal-rank-2", goalTestUserID, "savings", 20000, 0, 2)

	goals, total, err := repo.List(context.Background(), goalTestUserID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	assert.Len(t, goals, 3)
	assert.Equal(t, 1, goals[0].PriorityRank)
	assert.Equal(t, 2, goals[1].PriorityRank)
	assert.Equal(t, 3, goals[2].PriorityRank)
}

func TestGoalList_FiltersByUser(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	repo := queries.NewGoalRepository(database)

	seedGoal(t, repo, "goal-u1-001", goalTestUserID, "savings", 10000, 0, 1)
	seedGoal(t, repo, "goal-u2-001", goalTestUser2ID, "savings", 5000, 0, 1)

	goals, total, err := repo.List(context.Background(), goalTestUserID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, goals, 1)
}

func TestGoalList_Pagination(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	repo := queries.NewGoalRepository(database)

	for i := 1; i <= 5; i++ {
		seedGoal(t, repo, "goal-page-"+string(rune('0'+i)), goalTestUserID, "savings", float64(i*1000), 0, i)
	}

	goals, total, err := repo.List(context.Background(), goalTestUserID, 1, 2)
	require.NoError(t, err)
	assert.Equal(t, 5, total)
	assert.Len(t, goals, 2)
}

func TestGoalUpdate(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	repo := queries.NewGoalRepository(database)

	seedGoal(t, repo, "goal-test-001", goalTestUserID, "savings", 10000, 3000, 1)

	goal, err := repo.GetByID(context.Background(), goalTestUserID, "goal-test-001")
	require.NoError(t, err)

	goal.Name = "Updated Goal"
	goal.CurrentAmount = 5000
	found, err := repo.Update(context.Background(), goal)
	require.NoError(t, err)
	assert.True(t, found)

	updated, err := repo.GetByID(context.Background(), goalTestUserID, "goal-test-001")
	require.NoError(t, err)
	assert.Equal(t, "Updated Goal", updated.Name)
	assert.Equal(t, 5000.0, updated.CurrentAmount)
}

func TestGoalUpdate_WrongUser(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	repo := queries.NewGoalRepository(database)

	seedGoal(t, repo, "goal-test-001", goalTestUserID, "savings", 10000, 0, 1)

	found, err := repo.Update(context.Background(), &models.Goal{
		ID:           "goal-test-001",
		UserID:       goalTestUser2ID,
		Name:         "Hacked",
		GoalType:     "savings",
		TargetAmount: 99999,
		PriorityRank: 1,
	})
	require.NoError(t, err)
	assert.False(t, found)
}

func TestGoalSoftDelete(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	repo := queries.NewGoalRepository(database)

	seedGoal(t, repo, "goal-test-001", goalTestUserID, "savings", 10000, 0, 1)

	found, err := repo.SoftDelete(context.Background(), goalTestUserID, "goal-test-001")
	require.NoError(t, err)
	assert.True(t, found)

	goal, err := repo.GetByID(context.Background(), goalTestUserID, "goal-test-001")
	require.NoError(t, err)
	assert.Nil(t, goal)
}

func TestGoalSoftDelete_ExcludedFromList(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	repo := queries.NewGoalRepository(database)

	seedGoal(t, repo, "goal-test-001", goalTestUserID, "savings", 10000, 0, 1)
	seedGoal(t, repo, "goal-test-002", goalTestUserID, "savings", 20000, 0, 2)

	found, err := repo.SoftDelete(context.Background(), goalTestUserID, "goal-test-001")
	require.NoError(t, err)
	assert.True(t, found)

	goals, total, err := repo.List(context.Background(), goalTestUserID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, goals, 1)
	assert.Equal(t, "goal-test-002", goals[0].ID)
}

func TestGoalReorder(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	repo := queries.NewGoalRepository(database)

	seedGoal(t, repo, "goal-a", goalTestUserID, "savings", 10000, 0, 1)
	seedGoal(t, repo, "goal-b", goalTestUserID, "savings", 20000, 0, 2)
	seedGoal(t, repo, "goal-c", goalTestUserID, "savings", 30000, 0, 3)

	err := repo.Reorder(context.Background(), goalTestUserID, []queries.GoalRanking{
		{ID: "goal-a", Rank: 3},
		{ID: "goal-b", Rank: 1},
		{ID: "goal-c", Rank: 2},
	})
	require.NoError(t, err)

	goals, _, err := repo.List(context.Background(), goalTestUserID, 1, 20)
	require.NoError(t, err)
	assert.Equal(t, "goal-b", goals[0].ID)
	assert.Equal(t, 1, goals[0].PriorityRank)
	assert.Equal(t, "goal-c", goals[1].ID)
	assert.Equal(t, 2, goals[1].PriorityRank)
	assert.Equal(t, "goal-a", goals[2].ID)
	assert.Equal(t, 3, goals[2].PriorityRank)
}

func TestGoalReorder_NonexistentGoal_Fails(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	repo := queries.NewGoalRepository(database)

	seedGoal(t, repo, "goal-a", goalTestUserID, "savings", 10000, 0, 1)

	err := repo.Reorder(context.Background(), goalTestUserID, []queries.GoalRanking{
		{ID: "goal-a", Rank: 2},
		{ID: "nonexistent", Rank: 1},
	})
	assert.ErrorIs(t, err, queries.ErrGoalNotFound)

	goal, err := repo.GetByID(context.Background(), goalTestUserID, "goal-a")
	require.NoError(t, err)
	assert.Equal(t, 1, goal.PriorityRank)
}

func TestGoalReorder_WrongUser_Fails(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	repo := queries.NewGoalRepository(database)

	seedGoal(t, repo, "goal-a", goalTestUserID, "savings", 10000, 0, 1)

	err := repo.Reorder(context.Background(), goalTestUser2ID, []queries.GoalRanking{
		{ID: "goal-a", Rank: 2},
	})
	assert.ErrorIs(t, err, queries.ErrGoalNotFound)
}

func TestGoalNextPriorityRank(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	repo := queries.NewGoalRepository(database)

	rank, err := repo.NextPriorityRank(context.Background(), goalTestUserID)
	require.NoError(t, err)
	assert.Equal(t, 1, rank)

	seedGoal(t, repo, "goal-a", goalTestUserID, "savings", 10000, 0, 1)
	seedGoal(t, repo, "goal-b", goalTestUserID, "savings", 20000, 0, 2)

	rank, err = repo.NextPriorityRank(context.Background(), goalTestUserID)
	require.NoError(t, err)
	assert.Equal(t, 3, rank)
}

func TestGoalWithLinkedAccount(t *testing.T) {
	t.Parallel()
	database := setupGoalTestDB(t)
	goalRepo := queries.NewGoalRepository(database)
	accountRepo := queries.NewAccountRepository(database)

	err := accountRepo.Create(context.Background(), &models.Account{
		ID:          "acc-test-001",
		UserID:      goalTestUserID,
		Name:        "Savings Account",
		Institution: "Test Bank",
		AccountType: "savings",
		Currency:    "USD",
		IsActive:    true,
	})
	require.NoError(t, err)

	accountID := "acc-test-001"
	err = goalRepo.Create(context.Background(), &models.Goal{
		ID:              "goal-test-001",
		UserID:          goalTestUserID,
		Name:            "Ring Fund",
		GoalType:        "savings",
		TargetAmount:    8000,
		CurrentAmount:   2000,
		LinkedAccountID: &accountID,
		PriorityRank:    1,
	})
	require.NoError(t, err)

	goal, err := goalRepo.GetByID(context.Background(), goalTestUserID, "goal-test-001")
	require.NoError(t, err)
	require.NotNil(t, goal)
	require.NotNil(t, goal.LinkedAccountID)
	assert.Equal(t, "acc-test-001", *goal.LinkedAccountID)
}
