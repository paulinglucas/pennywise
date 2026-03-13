package queries

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/jamespsullivan/pennywise/internal/models"
	"github.com/jamespsullivan/pennywise/internal/observability"
)

type TransactionFilter struct {
	AccountID *string
	Category  *string
	Type      *string
	DateFrom  *time.Time
	DateTo    *time.Time
	AmountMin *float64
	AmountMax *float64
	Tags      []string
	Search    *string
	GroupID   *string
}

type SQLiteTransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *SQLiteTransactionRepository {
	return &SQLiteTransactionRepository{db: db}
}

func (r *SQLiteTransactionRepository) Create(ctx context.Context, txn *models.Transaction, tags []string) error {
	start := time.Now()
	defer func() { observability.RecordDBQuery("create_transaction", time.Since(start)) }()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date, notes, is_recurring, recurring_transaction_id, group_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		txn.ID, txn.UserID, txn.AccountID, txn.Type, txn.Category, txn.Amount, txn.Currency,
		txn.Date.Format("2006-01-02"), txn.Notes, txn.IsRecurring, txn.RecurringTransactionID, txn.GroupID,
	)
	if err != nil {
		return err
	}

	if err := insertTags(ctx, tx, txn.ID, tags); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SQLiteTransactionRepository) GetByID(ctx context.Context, userID, id string) (*models.Transaction, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("get_transaction", time.Since(start)) }()

	var txn models.Transaction
	var dateStr string
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, account_id, type, category, amount, currency, date, notes, is_recurring, recurring_transaction_id, group_id, created_at, updated_at
		 FROM transactions WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		id, userID,
	).Scan(&txn.ID, &txn.UserID, &txn.AccountID, &txn.Type, &txn.Category, &txn.Amount, &txn.Currency,
		&dateStr, &txn.Notes, &txn.IsRecurring, &txn.RecurringTransactionID, &txn.GroupID, &txn.CreatedAt, &txn.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	txn.Date, err = parseDateString(dateStr)
	if err != nil {
		return nil, err
	}

	tags, err := r.loadTags(ctx, txn.ID)
	if err != nil {
		return nil, err
	}
	txn.Tags = tags

	return &txn, nil
}

func (r *SQLiteTransactionRepository) List(ctx context.Context, userID string, filter TransactionFilter, page, perPage int) ([]models.Transaction, int, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("list_transactions", time.Since(start)) }()

	where, args := buildFilterWhere(userID, filter)

	var total int
	if len(filter.Tags) > 0 {
		err := r.db.QueryRowContext(ctx,
			"SELECT COUNT(DISTINCT transactions.id) FROM transactions INNER JOIN transaction_tags ON transactions.id = transaction_tags.transaction_id "+where,
			args...,
		).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := r.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM transactions "+where,
			args...,
		).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	offset := (page - 1) * perPage
	dataArgs := make([]interface{}, len(args))
	copy(dataArgs, args)
	dataArgs = append(dataArgs, perPage, offset)

	selectCols := "transactions.id, transactions.user_id, transactions.account_id, transactions.type, transactions.category, transactions.amount, transactions.currency, transactions.date, transactions.notes, transactions.is_recurring, transactions.recurring_transaction_id, transactions.group_id, transactions.created_at, transactions.updated_at"

	var dataQuery string
	if len(filter.Tags) > 0 {
		dataQuery = "SELECT " + selectCols + " FROM transactions INNER JOIN transaction_tags ON transactions.id = transaction_tags.transaction_id " + where + " GROUP BY transactions.id ORDER BY transactions.date DESC, transactions.created_at DESC LIMIT ? OFFSET ?"
	} else {
		dataQuery = "SELECT " + selectCols + " FROM transactions " + where + " ORDER BY transactions.date DESC, transactions.created_at DESC LIMIT ? OFFSET ?"
	}

	rows, err := r.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var txns []models.Transaction
	var txnIDs []string
	for rows.Next() {
		var txn models.Transaction
		var dateStr string
		if err := rows.Scan(&txn.ID, &txn.UserID, &txn.AccountID, &txn.Type, &txn.Category, &txn.Amount, &txn.Currency,
			&dateStr, &txn.Notes, &txn.IsRecurring, &txn.RecurringTransactionID, &txn.GroupID, &txn.CreatedAt, &txn.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		txn.Date, err = parseDateString(dateStr)
		if err != nil {
			return nil, 0, err
		}
		txns = append(txns, txn)
		txnIDs = append(txnIDs, txn.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	if len(txnIDs) > 0 {
		tagMap, err := r.loadTagsBatch(ctx, txnIDs)
		if err != nil {
			return nil, 0, err
		}
		for i := range txns {
			txns[i].Tags = tagMap[txns[i].ID]
		}
	}

	return txns, total, nil
}

func (r *SQLiteTransactionRepository) Update(ctx context.Context, txn *models.Transaction, tags []string) (bool, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("update_transaction", time.Since(start)) }()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx,
		`UPDATE transactions SET account_id=?, type=?, category=?, amount=?, currency=?, date=?, notes=?, is_recurring=?, recurring_transaction_id=?, group_id=?, updated_at=datetime('now')
		 WHERE id=? AND user_id=? AND deleted_at IS NULL`,
		txn.AccountID, txn.Type, txn.Category, txn.Amount, txn.Currency,
		txn.Date.Format("2006-01-02"), txn.Notes, txn.IsRecurring, txn.RecurringTransactionID, txn.GroupID,
		txn.ID, txn.UserID,
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

	_, err = tx.ExecContext(ctx, "DELETE FROM transaction_tags WHERE transaction_id = ?", txn.ID)
	if err != nil {
		return false, err
	}

	if err := insertTags(ctx, tx, txn.ID, tags); err != nil {
		return false, err
	}

	return true, tx.Commit()
}

func (r *SQLiteTransactionRepository) SoftDelete(ctx context.Context, userID, id string) (bool, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("delete_transaction", time.Since(start)) }()

	result, err := r.db.ExecContext(ctx,
		`UPDATE transactions SET deleted_at=datetime('now'), updated_at=datetime('now')
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

func (r *SQLiteTransactionRepository) ListCategories(ctx context.Context, userID string) ([]string, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("list_categories", time.Since(start)) }()

	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT category FROM transactions WHERE user_id = ? AND deleted_at IS NULL
		 UNION
		 SELECT DISTINCT category FROM recurring_transactions WHERE user_id = ? AND deleted_at IS NULL
		 ORDER BY category`,
		userID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var categories []string
	for rows.Next() {
		var cat string
		if err := rows.Scan(&cat); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}
	return categories, rows.Err()
}

type CategoryUpdate struct {
	TransactionID string
	Category      string
}

func (r *SQLiteTransactionRepository) BulkCategorize(ctx context.Context, userID string, updates []CategoryUpdate) (int, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("bulk_categorize_transactions", time.Since(start)) }()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	updated := 0
	for _, u := range updates {
		result, err := tx.ExecContext(ctx,
			`UPDATE transactions SET category = ?, updated_at = datetime('now')
			 WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
			u.Category, u.TransactionID, userID,
		)
		if err != nil {
			return updated, err
		}
		n, _ := result.RowsAffected()
		updated += int(n)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return updated, nil
}

type BulkCreateError struct {
	Row     int
	Message string
}

func (r *SQLiteTransactionRepository) BulkCreate(ctx context.Context, txns []models.Transaction) (int, []BulkCreateError) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("bulk_create_transactions", time.Since(start)) }()

	var imported int
	var errs []BulkCreateError

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, []BulkCreateError{{Row: 0, Message: err.Error()}}
	}
	defer func() { _ = tx.Rollback() }()

	for i, txn := range txns {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO transactions (id, user_id, account_id, type, category, amount, currency, date, notes, is_recurring)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			txn.ID, txn.UserID, txn.AccountID, txn.Type, txn.Category, txn.Amount, txn.Currency,
			txn.Date.Format("2006-01-02"), txn.Notes, txn.IsRecurring,
		)
		if err != nil {
			errs = append(errs, BulkCreateError{Row: i + 1, Message: err.Error()})
			continue
		}
		imported++
	}

	if err := tx.Commit(); err != nil {
		return 0, []BulkCreateError{{Row: 0, Message: err.Error()}}
	}

	return imported, errs
}

func (r *SQLiteTransactionRepository) loadTags(ctx context.Context, transactionID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT tag FROM transaction_tags WHERE transaction_id = ?", transactionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *SQLiteTransactionRepository) loadTagsBatch(ctx context.Context, ids []string) (map[string][]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := "SELECT transaction_id, tag FROM transaction_tags WHERE transaction_id IN (" + strings.Join(placeholders, ",") + ")" //nolint:gosec // placeholders are parameterized ?
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	tagMap := make(map[string][]string)
	for rows.Next() {
		var txnID, tag string
		if err := rows.Scan(&txnID, &tag); err != nil {
			return nil, err
		}
		tagMap[txnID] = append(tagMap[txnID], tag)
	}
	return tagMap, rows.Err()
}

func insertTags(ctx context.Context, tx *sql.Tx, transactionID string, tags []string) error {
	for _, tag := range tags {
		_, err := tx.ExecContext(ctx,
			"INSERT INTO transaction_tags (id, transaction_id, tag) VALUES (?, ?, ?)",
			uuid.New().String(), transactionID, tag,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseDateString(s string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", s)
	if err == nil {
		return parsed, nil
	}
	return time.Parse(time.RFC3339, s)
}

func buildFilterWhere(userID string, filter TransactionFilter) (string, []interface{}) {
	conditions := []string{"transactions.user_id = ?", "transactions.deleted_at IS NULL"}
	args := []interface{}{userID}

	if filter.AccountID != nil {
		conditions = append(conditions, "transactions.account_id = ?")
		args = append(args, *filter.AccountID)
	}
	if filter.Category != nil {
		conditions = append(conditions, "transactions.category = ?")
		args = append(args, *filter.Category)
	}
	if filter.Type != nil {
		conditions = append(conditions, "transactions.type = ?")
		args = append(args, *filter.Type)
	}
	if filter.DateFrom != nil {
		conditions = append(conditions, "transactions.date >= ?")
		args = append(args, filter.DateFrom.Format("2006-01-02"))
	}
	if filter.DateTo != nil {
		conditions = append(conditions, "transactions.date <= ?")
		args = append(args, filter.DateTo.Format("2006-01-02"))
	}
	if filter.AmountMin != nil {
		conditions = append(conditions, "transactions.amount >= ?")
		args = append(args, *filter.AmountMin)
	}
	if filter.AmountMax != nil {
		conditions = append(conditions, "transactions.amount <= ?")
		args = append(args, *filter.AmountMax)
	}
	if len(filter.Tags) > 0 {
		placeholders := make([]string, len(filter.Tags))
		for i, tag := range filter.Tags {
			placeholders[i] = "?"
			args = append(args, tag)
		}
		conditions = append(conditions, "transaction_tags.tag IN ("+strings.Join(placeholders, ",")+")")
	}
	if filter.Search != nil {
		conditions = append(conditions, "(transactions.notes LIKE ? OR transactions.category LIKE ?)")
		searchTerm := "%" + *filter.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}
	if filter.GroupID != nil {
		conditions = append(conditions, "transactions.group_id = ?")
		args = append(args, *filter.GroupID)
	}

	return "WHERE " + strings.Join(conditions, " AND "), args
}
