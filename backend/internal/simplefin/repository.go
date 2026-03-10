package simplefin

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/jamespsullivan/pennywise/internal/models"
)

type LinkedAccount struct {
	AccountID   string
	SimplefinID string
	AccountName string
	Institution string
}

type SQLiteSimplefinRepository struct {
	db *sql.DB
}

func NewSimplefinRepository(db *sql.DB) *SQLiteSimplefinRepository {
	return &SQLiteSimplefinRepository{db: db}
}

func (r *SQLiteSimplefinRepository) SaveConnection(ctx context.Context, userID, encryptedAccessURL string) error {
	_, err := r.db.ExecContext(ctx, //nolint:gosec // parameterized query
		`INSERT INTO simplefin_connections (id, user_id, access_url)
		 VALUES (?, ?, ?)
		 ON CONFLICT(user_id) DO UPDATE SET access_url=excluded.access_url, sync_error=NULL, updated_at=datetime('now')`,
		uuid.New().String(), userID, encryptedAccessURL,
	)
	return err
}

func (r *SQLiteSimplefinRepository) GetConnection(ctx context.Context, userID string) (*models.SimplefinConnection, error) {
	var conn models.SimplefinConnection
	err := r.db.QueryRowContext(ctx, //nolint:gosec // parameterized query
		`SELECT id, user_id, access_url, last_sync_at, sync_error, created_at, updated_at
		 FROM simplefin_connections WHERE user_id = ?`,
		userID,
	).Scan(&conn.ID, &conn.UserID, &conn.AccessURL, &conn.LastSyncAt, &conn.SyncError, &conn.CreatedAt, &conn.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (r *SQLiteSimplefinRepository) GetAllConnections(ctx context.Context) ([]models.SimplefinConnection, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, access_url, last_sync_at, sync_error, created_at, updated_at
		 FROM simplefin_connections`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var conns []models.SimplefinConnection
	for rows.Next() {
		var conn models.SimplefinConnection
		if err := rows.Scan(&conn.ID, &conn.UserID, &conn.AccessURL, &conn.LastSyncAt, &conn.SyncError, &conn.CreatedAt, &conn.UpdatedAt); err != nil {
			return nil, err
		}
		conns = append(conns, conn)
	}
	return conns, rows.Err()
}

func (r *SQLiteSimplefinRepository) DeleteConnection(ctx context.Context, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query, not injectable
		`UPDATE accounts SET simplefin_id = NULL, updated_at = datetime('now')
		 WHERE user_id = ? AND simplefin_id IS NOT NULL`,
		userID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query, not injectable
		`DELETE FROM simplefin_connections WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SQLiteSimplefinRepository) LinkAccount(ctx context.Context, userID, accountID, simplefinID string) error {
	result, err := r.db.ExecContext(ctx, //nolint:gosec // parameterized query, not injectable
		`UPDATE accounts SET simplefin_id = ?, updated_at = datetime('now')
		 WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		simplefinID, accountID, userID,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *SQLiteSimplefinRepository) UnlinkAccount(ctx context.Context, userID, accountID string) error {
	_, err := r.db.ExecContext(ctx, //nolint:gosec // parameterized query
		`UPDATE accounts SET simplefin_id = NULL, updated_at = datetime('now')
		 WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		accountID, userID,
	)
	return err
}

func (r *SQLiteSimplefinRepository) GetLinkedAccounts(ctx context.Context, userID string) ([]LinkedAccount, error) {
	rows, err := r.db.QueryContext(ctx, //nolint:gosec // parameterized query
		`SELECT id, simplefin_id, name, institution
		 FROM accounts
		 WHERE user_id = ? AND simplefin_id IS NOT NULL AND deleted_at IS NULL
		 ORDER BY name`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var linked []LinkedAccount
	for rows.Next() {
		var la LinkedAccount
		if err := rows.Scan(&la.AccountID, &la.SimplefinID, &la.AccountName, &la.Institution); err != nil {
			return nil, err
		}
		linked = append(linked, la)
	}
	return linked, rows.Err()
}

func (r *SQLiteSimplefinRepository) UpdateSyncSuccess(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, //nolint:gosec // parameterized query
		`UPDATE simplefin_connections SET last_sync_at = datetime('now'), sync_error = NULL, updated_at = datetime('now')
		 WHERE user_id = ?`,
		userID,
	)
	return err
}

func (r *SQLiteSimplefinRepository) UpdateSyncError(ctx context.Context, userID, syncError string) error {
	_, err := r.db.ExecContext(ctx, //nolint:gosec // parameterized query
		`UPDATE simplefin_connections SET sync_error = ?, updated_at = datetime('now')
		 WHERE user_id = ?`,
		syncError, userID,
	)
	return err
}

func (r *SQLiteSimplefinRepository) GetAssetForAccount(ctx context.Context, userID, accountID string) (*models.Asset, error) {
	var asset models.Asset
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, account_id, name, asset_type, current_value, currency, metadata, created_at, updated_at
		 FROM assets
		 WHERE user_id = ? AND account_id = ? AND deleted_at IS NULL`,
		userID, accountID,
	).Scan(&asset.ID, &asset.UserID, &asset.AccountID, &asset.Name, &asset.AssetType, &asset.CurrentValue, &asset.Currency, &asset.Metadata, &asset.CreatedAt, &asset.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *SQLiteSimplefinRepository) UpdateAssetValue(ctx context.Context, assetID string, newValue float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query
		`UPDATE assets SET current_value = ?, updated_at = datetime('now') WHERE id = ?`,
		newValue, assetID,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, //nolint:gosec // parameterized query
		`INSERT INTO asset_history (id, asset_id, value, recorded_at)
		 VALUES (?, ?, ?, datetime('now'))`,
		uuid.New().String(), assetID, newValue,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
