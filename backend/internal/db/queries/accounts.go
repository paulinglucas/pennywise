package queries

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jamespsullivan/pennywise/internal/models"
	"github.com/jamespsullivan/pennywise/internal/observability"
)

type SQLiteAccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *SQLiteAccountRepository {
	return &SQLiteAccountRepository{db: db}
}

func (r *SQLiteAccountRepository) List(ctx context.Context, userID string, page, perPage int) ([]models.Account, int, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("list_accounts", time.Since(start)) }()

	var total int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM accounts WHERE user_id = ? AND deleted_at IS NULL",
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, name, institution, account_type, currency, is_active, created_at, updated_at
		 FROM accounts WHERE user_id = ? AND deleted_at IS NULL
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var accounts []models.Account
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Institution, &a.AccountType, &a.Currency, &a.IsActive, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, err
		}
		accounts = append(accounts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return accounts, total, nil
}

func (r *SQLiteAccountRepository) Create(ctx context.Context, account *models.Account) error {
	start := time.Now()
	defer func() { observability.RecordDBQuery("create_account", time.Since(start)) }()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO accounts (id, user_id, name, institution, account_type, currency, is_active)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		account.ID, account.UserID, account.Name, account.Institution, account.AccountType, account.Currency, account.IsActive,
	)
	return err
}

func (r *SQLiteAccountRepository) GetByID(ctx context.Context, userID, id string) (*models.Account, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("get_account", time.Since(start)) }()

	var a models.Account
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, institution, account_type, currency, is_active, created_at, updated_at
		 FROM accounts WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		id, userID,
	).Scan(&a.ID, &a.UserID, &a.Name, &a.Institution, &a.AccountType, &a.Currency, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *SQLiteAccountRepository) Update(ctx context.Context, account *models.Account) (bool, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("update_account", time.Since(start)) }()

	result, err := r.db.ExecContext(ctx,
		`UPDATE accounts SET name=?, institution=?, account_type=?, currency=?, is_active=?, updated_at=datetime('now')
		 WHERE id=? AND user_id=? AND deleted_at IS NULL`,
		account.Name, account.Institution, account.AccountType, account.Currency, account.IsActive,
		account.ID, account.UserID,
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (r *SQLiteAccountRepository) SoftDelete(ctx context.Context, userID, id string) (bool, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("delete_account", time.Since(start)) }()

	result, err := r.db.ExecContext(ctx,
		`UPDATE accounts SET deleted_at=datetime('now'), updated_at=datetime('now')
		 WHERE id=? AND user_id=? AND deleted_at IS NULL`,
		id, userID,
	)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}
