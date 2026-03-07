package queries

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jamespsullivan/pennywise/internal/models"
	"github.com/jamespsullivan/pennywise/internal/observability"
)

type SQLiteGoalRepository struct {
	db *sql.DB
}

func NewGoalRepository(db *sql.DB) *SQLiteGoalRepository {
	return &SQLiteGoalRepository{db: db}
}

func (r *SQLiteGoalRepository) List(ctx context.Context, userID string, page, perPage int) ([]models.Goal, int, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("list_goals", time.Since(start)) }()

	var total int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM goals WHERE user_id = ? AND deleted_at IS NULL",
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, name, goal_type, target_amount, current_amount, deadline, linked_account_id, priority_rank, created_at, updated_at
		 FROM goals WHERE user_id = ? AND deleted_at IS NULL
		 ORDER BY priority_rank ASC, created_at ASC LIMIT ? OFFSET ?`,
		userID, perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var goals []models.Goal
	for rows.Next() {
		g, err := scanGoal(rows)
		if err != nil {
			return nil, 0, err
		}
		goals = append(goals, g)
	}
	return goals, total, rows.Err()
}

func (r *SQLiteGoalRepository) Create(ctx context.Context, goal *models.Goal) error {
	start := time.Now()
	defer func() { observability.RecordDBQuery("create_goal", time.Since(start)) }()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO goals (id, user_id, name, goal_type, target_amount, current_amount, deadline, linked_account_id, priority_rank)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		goal.ID, goal.UserID, goal.Name, goal.GoalType, goal.TargetAmount, goal.CurrentAmount,
		formatOptionalDate(goal.Deadline), goal.LinkedAccountID, goal.PriorityRank,
	)
	return err
}

func (r *SQLiteGoalRepository) GetByID(ctx context.Context, userID, id string) (*models.Goal, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("get_goal", time.Since(start)) }()

	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, goal_type, target_amount, current_amount, deadline, linked_account_id, priority_rank, created_at, updated_at
		 FROM goals WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		id, userID,
	)

	g, err := scanGoalRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *SQLiteGoalRepository) Update(ctx context.Context, goal *models.Goal) (bool, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("update_goal", time.Since(start)) }()

	result, err := r.db.ExecContext(ctx,
		`UPDATE goals SET name=?, goal_type=?, target_amount=?, current_amount=?, deadline=?, linked_account_id=?, priority_rank=?, updated_at=datetime('now')
		 WHERE id=? AND user_id=? AND deleted_at IS NULL`,
		goal.Name, goal.GoalType, goal.TargetAmount, goal.CurrentAmount,
		formatOptionalDate(goal.Deadline), goal.LinkedAccountID, goal.PriorityRank,
		goal.ID, goal.UserID,
	)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (r *SQLiteGoalRepository) SoftDelete(ctx context.Context, userID, id string) (bool, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("delete_goal", time.Since(start)) }()

	result, err := r.db.ExecContext(ctx,
		`UPDATE goals SET deleted_at=datetime('now'), updated_at=datetime('now')
		 WHERE id=? AND user_id=? AND deleted_at IS NULL`,
		id, userID,
	)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (r *SQLiteGoalRepository) Reorder(ctx context.Context, userID string, rankings []GoalRanking) error {
	start := time.Now()
	defer func() { observability.RecordDBQuery("reorder_goals", time.Since(start)) }()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for _, ranking := range rankings {
		result, err := tx.ExecContext(ctx,
			`UPDATE goals SET priority_rank=?, updated_at=datetime('now')
			 WHERE id=? AND user_id=? AND deleted_at IS NULL`,
			ranking.Rank, ranking.ID, userID,
		)
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return ErrGoalNotFound
		}
	}

	return tx.Commit()
}

func (r *SQLiteGoalRepository) NextPriorityRank(ctx context.Context, userID string) (int, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("next_priority_rank", time.Since(start)) }()

	var maxRank sql.NullInt64
	err := r.db.QueryRowContext(ctx,
		"SELECT MAX(priority_rank) FROM goals WHERE user_id = ? AND deleted_at IS NULL",
		userID,
	).Scan(&maxRank)
	if err != nil {
		return 1, err
	}
	if !maxRank.Valid {
		return 1, nil
	}
	return int(maxRank.Int64) + 1, nil
}

type GoalRanking struct {
	ID   string
	Rank int
}

var ErrGoalNotFound = errors.New("goal not found")

func scanGoal(rows *sql.Rows) (models.Goal, error) {
	var g models.Goal
	var deadlineStr *string
	err := rows.Scan(&g.ID, &g.UserID, &g.Name, &g.GoalType, &g.TargetAmount, &g.CurrentAmount,
		&deadlineStr, &g.LinkedAccountID, &g.PriorityRank, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return g, err
	}
	g.Deadline = parseOptionalDate(deadlineStr)
	return g, nil
}

func scanGoalRow(row *sql.Row) (models.Goal, error) {
	var g models.Goal
	var deadlineStr *string
	err := row.Scan(&g.ID, &g.UserID, &g.Name, &g.GoalType, &g.TargetAmount, &g.CurrentAmount,
		&deadlineStr, &g.LinkedAccountID, &g.PriorityRank, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return g, err
	}
	g.Deadline = parseOptionalDate(deadlineStr)
	return g, nil
}

func parseOptionalDate(s *string) *time.Time {
	if s == nil || *s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, *s)
		if err != nil {
			return nil
		}
	}
	return &t
}

func formatOptionalDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}
