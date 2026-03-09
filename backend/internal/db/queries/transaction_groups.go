package queries

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jamespsullivan/pennywise/internal/models"
	"github.com/jamespsullivan/pennywise/internal/observability"
)

type SQLiteTransactionGroupRepository struct {
	db *sql.DB
}

func NewTransactionGroupRepository(db *sql.DB) *SQLiteTransactionGroupRepository {
	return &SQLiteTransactionGroupRepository{db: db}
}

func (r *SQLiteTransactionGroupRepository) Create(ctx context.Context, group *models.TransactionGroup) error {
	start := time.Now()
	defer func() { observability.RecordDBQuery("create_transaction_group", time.Since(start)) }()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO transaction_groups (id, user_id, name) VALUES (?, ?, ?)`,
		group.ID, group.UserID, group.Name,
	)
	return err
}

func (r *SQLiteTransactionGroupRepository) GetByID(ctx context.Context, userID, id string) (*models.TransactionGroup, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("get_transaction_group", time.Since(start)) }()

	var group models.TransactionGroup
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, created_at, updated_at
		 FROM transaction_groups WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		id, userID,
	).Scan(&group.ID, &group.UserID, &group.Name, &group.CreatedAt, &group.UpdatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *SQLiteTransactionGroupRepository) Update(ctx context.Context, group *models.TransactionGroup) (bool, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("update_transaction_group", time.Since(start)) }()

	result, err := r.db.ExecContext(ctx,
		`UPDATE transaction_groups SET name=?, updated_at=datetime('now')
		 WHERE id=? AND user_id=? AND deleted_at IS NULL`,
		group.Name, group.ID, group.UserID,
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

func (r *SQLiteTransactionGroupRepository) SoftDelete(ctx context.Context, userID, id string) (bool, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("delete_transaction_group", time.Since(start)) }()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx,
		`UPDATE transaction_groups SET deleted_at=datetime('now'), updated_at=datetime('now')
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
	if affected == 0 {
		return false, nil
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE transactions SET deleted_at=datetime('now'), updated_at=datetime('now')
		 WHERE group_id=? AND user_id=? AND deleted_at IS NULL`,
		id, userID,
	)
	if err != nil {
		return false, err
	}

	return true, tx.Commit()
}

func (r *SQLiteTransactionGroupRepository) List(ctx context.Context, userID string, page, perPage int) ([]models.TransactionGroup, int, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("list_transaction_groups", time.Since(start)) }()

	var total int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM transaction_groups WHERE user_id = ? AND deleted_at IS NULL`,
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, name, created_at, updated_at
		 FROM transaction_groups WHERE user_id = ? AND deleted_at IS NULL
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var groups []models.TransactionGroup
	for rows.Next() {
		var g models.TransactionGroup
		if err := rows.Scan(&g.ID, &g.UserID, &g.Name, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, 0, err
		}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

func (r *SQLiteTransactionGroupRepository) ListMembers(ctx context.Context, userID, groupID string) ([]models.Transaction, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("list_group_members", time.Since(start)) }()

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, account_id, type, category, amount, currency, date, notes, is_recurring, recurring_transaction_id, group_id, created_at, updated_at
		 FROM transactions WHERE group_id = ? AND user_id = ? AND deleted_at IS NULL
		 ORDER BY category ASC`,
		groupID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var txns []models.Transaction
	for rows.Next() {
		var txn models.Transaction
		var dateStr string
		if err := rows.Scan(&txn.ID, &txn.UserID, &txn.AccountID, &txn.Type, &txn.Category, &txn.Amount, &txn.Currency,
			&dateStr, &txn.Notes, &txn.IsRecurring, &txn.RecurringTransactionID, &txn.GroupID, &txn.CreatedAt, &txn.UpdatedAt,
		); err != nil {
			return nil, err
		}
		txn.Date, err = parseDateString(dateStr)
		if err != nil {
			return nil, err
		}
		txns = append(txns, txn)
	}
	return txns, rows.Err()
}
