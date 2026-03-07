package queries

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jamespsullivan/pennywise/internal/models"
	"github.com/jamespsullivan/pennywise/internal/observability"
)

type SQLiteRecurringRepository struct {
	db *sql.DB
}

func NewRecurringRepository(db *sql.DB) *SQLiteRecurringRepository {
	return &SQLiteRecurringRepository{db: db}
}

func (r *SQLiteRecurringRepository) List(ctx context.Context, userID string, page, perPage int) ([]models.RecurringTransaction, int, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("list_recurring", time.Since(start)) }()

	var total int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM recurring_transactions WHERE user_id = ? AND deleted_at IS NULL",
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, account_id, type, category, amount, currency, frequency, next_occurrence, is_active, created_at, updated_at
		 FROM recurring_transactions WHERE user_id = ? AND deleted_at IS NULL
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var items []models.RecurringTransaction
	for rows.Next() {
		rt, err := scanRecurring(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, rt)
	}
	return items, total, rows.Err()
}

func (r *SQLiteRecurringRepository) Create(ctx context.Context, rec *models.RecurringTransaction) error {
	start := time.Now()
	defer func() { observability.RecordDBQuery("create_recurring", time.Since(start)) }()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO recurring_transactions (id, user_id, account_id, type, category, amount, currency, frequency, next_occurrence, is_active)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.ID, rec.UserID, rec.AccountID, rec.Type, rec.Category, rec.Amount, rec.Currency,
		rec.Frequency, rec.NextOccurrence.Format("2006-01-02"), rec.IsActive,
	)
	return err
}

func (r *SQLiteRecurringRepository) GetByID(ctx context.Context, userID, id string) (*models.RecurringTransaction, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("get_recurring", time.Since(start)) }()

	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, account_id, type, category, amount, currency, frequency, next_occurrence, is_active, created_at, updated_at
		 FROM recurring_transactions WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		id, userID,
	)

	rt, err := scanRecurringRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *SQLiteRecurringRepository) Update(ctx context.Context, rec *models.RecurringTransaction) (bool, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("update_recurring", time.Since(start)) }()

	result, err := r.db.ExecContext(ctx,
		`UPDATE recurring_transactions SET account_id=?, type=?, category=?, amount=?, currency=?, frequency=?, next_occurrence=?, is_active=?, updated_at=datetime('now')
		 WHERE id=? AND user_id=? AND deleted_at IS NULL`,
		rec.AccountID, rec.Type, rec.Category, rec.Amount, rec.Currency,
		rec.Frequency, rec.NextOccurrence.Format("2006-01-02"), rec.IsActive,
		rec.ID, rec.UserID,
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

func (r *SQLiteRecurringRepository) SoftDelete(ctx context.Context, userID, id string) (bool, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("delete_recurring", time.Since(start)) }()

	result, err := r.db.ExecContext(ctx,
		`UPDATE recurring_transactions SET deleted_at=datetime('now'), updated_at=datetime('now')
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

func scanRecurring(rows *sql.Rows) (models.RecurringTransaction, error) {
	var rt models.RecurringTransaction
	var nextOccStr string
	err := rows.Scan(&rt.ID, &rt.UserID, &rt.AccountID, &rt.Type, &rt.Category, &rt.Amount,
		&rt.Currency, &rt.Frequency, &nextOccStr, &rt.IsActive, &rt.CreatedAt, &rt.UpdatedAt)
	if err != nil {
		return rt, err
	}
	rt.NextOccurrence, _ = parseDateFallback(nextOccStr)
	return rt, nil
}

func scanRecurringRow(row *sql.Row) (models.RecurringTransaction, error) {
	var rt models.RecurringTransaction
	var nextOccStr string
	err := row.Scan(&rt.ID, &rt.UserID, &rt.AccountID, &rt.Type, &rt.Category, &rt.Amount,
		&rt.Currency, &rt.Frequency, &nextOccStr, &rt.IsActive, &rt.CreatedAt, &rt.UpdatedAt)
	if err != nil {
		return rt, err
	}
	rt.NextOccurrence, _ = parseDateFallback(nextOccStr)
	return rt, nil
}

func parseDateFallback(s string) (t time.Time, err error) {
	t, err = time.Parse("2006-01-02", s)
	if err != nil {
		t, err = time.Parse(time.RFC3339, s)
	}
	return
}
