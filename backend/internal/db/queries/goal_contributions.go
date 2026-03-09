package queries

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jamespsullivan/pennywise/internal/models"
	"github.com/jamespsullivan/pennywise/internal/observability"
)

type SQLiteGoalContributionRepository struct {
	db *sql.DB
}

func NewGoalContributionRepository(db *sql.DB) *SQLiteGoalContributionRepository {
	return &SQLiteGoalContributionRepository{db: db}
}

func (r *SQLiteGoalContributionRepository) Create(ctx context.Context, contrib *models.GoalContribution) error {
	start := time.Now()
	defer func() { observability.RecordDBQuery("create_goal_contribution", time.Since(start)) }()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO goal_contributions (id, goal_id, user_id, amount, notes, transaction_id, contributed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		contrib.ID, contrib.GoalID, contrib.UserID, contrib.Amount, contrib.Notes,
		contrib.TransactionID, contrib.ContributedAt.Format("2006-01-02"),
	)
	return err
}

func (r *SQLiteGoalContributionRepository) ListByGoal(ctx context.Context, userID, goalID string, page, perPage int) ([]models.GoalContribution, int, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("list_goal_contributions", time.Since(start)) }()

	var total int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM goal_contributions WHERE goal_id = ? AND user_id = ?",
		goalID, userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, goal_id, user_id, amount, notes, transaction_id, contributed_at, created_at
		 FROM goal_contributions WHERE goal_id = ? AND user_id = ?
		 ORDER BY contributed_at DESC, created_at DESC LIMIT ? OFFSET ?`,
		goalID, userID, perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var contribs []models.GoalContribution
	for rows.Next() {
		c, scanErr := scanContribution(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		contribs = append(contribs, c)
	}
	return contribs, total, rows.Err()
}

func (r *SQLiteGoalContributionRepository) GetByID(ctx context.Context, userID, id string) (*models.GoalContribution, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("get_goal_contribution", time.Since(start)) }()

	row := r.db.QueryRowContext(ctx,
		`SELECT id, goal_id, user_id, amount, notes, transaction_id, contributed_at, created_at
		 FROM goal_contributions WHERE id = ? AND user_id = ?`,
		id, userID,
	)

	c, err := scanContributionRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *SQLiteGoalContributionRepository) Delete(ctx context.Context, userID, id string) (bool, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("delete_goal_contribution", time.Since(start)) }()

	result, err := r.db.ExecContext(ctx,
		"DELETE FROM goal_contributions WHERE id = ? AND user_id = ?",
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

func scanContribution(rows *sql.Rows) (models.GoalContribution, error) {
	var c models.GoalContribution
	var contributedAtStr string
	err := rows.Scan(&c.ID, &c.GoalID, &c.UserID, &c.Amount, &c.Notes, &c.TransactionID, &contributedAtStr, &c.CreatedAt)
	if err != nil {
		return c, err
	}
	c.ContributedAt = parseContributionDate(contributedAtStr)
	return c, nil
}

func scanContributionRow(row *sql.Row) (models.GoalContribution, error) {
	var c models.GoalContribution
	var contributedAtStr string
	err := row.Scan(&c.ID, &c.GoalID, &c.UserID, &c.Amount, &c.Notes, &c.TransactionID, &contributedAtStr, &c.CreatedAt)
	if err != nil {
		return c, err
	}
	c.ContributedAt = parseContributionDate(contributedAtStr)
	return c, nil
}

func parseContributionDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		t, _ = time.Parse(time.RFC3339, s)
	}
	return t
}
