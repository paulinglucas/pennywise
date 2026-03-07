package queries

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/jamespsullivan/pennywise/internal/models"
)

type SQLiteAssetRepository struct {
	db *sql.DB
}

func NewAssetRepository(db *sql.DB) *SQLiteAssetRepository {
	return &SQLiteAssetRepository{db: db}
}

func (r *SQLiteAssetRepository) List(ctx context.Context, userID string, page, perPage int) ([]models.Asset, int, error) {
	var total int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM assets WHERE user_id = ? AND deleted_at IS NULL",
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, account_id, name, asset_type, current_value, currency, metadata, created_at, updated_at
		 FROM assets WHERE user_id = ? AND deleted_at IS NULL
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var assets []models.Asset
	for rows.Next() {
		var a models.Asset
		if err := rows.Scan(&a.ID, &a.UserID, &a.AccountID, &a.Name, &a.AssetType, &a.CurrentValue, &a.Currency, &a.Metadata, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, 0, err
		}
		assets = append(assets, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

func (r *SQLiteAssetRepository) Create(ctx context.Context, asset *models.Asset) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO assets (id, user_id, account_id, name, asset_type, current_value, currency, metadata)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		asset.ID, asset.UserID, asset.AccountID, asset.Name, asset.AssetType, asset.CurrentValue, asset.Currency, asset.Metadata,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, datetime('now'))`,
		uuid.New().String(), asset.ID, asset.CurrentValue,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SQLiteAssetRepository) GetByID(ctx context.Context, userID, id string) (*models.Asset, error) {
	var a models.Asset
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, account_id, name, asset_type, current_value, currency, metadata, created_at, updated_at
		 FROM assets WHERE id = ? AND user_id = ? AND deleted_at IS NULL`,
		id, userID,
	).Scan(&a.ID, &a.UserID, &a.AccountID, &a.Name, &a.AssetType, &a.CurrentValue, &a.Currency, &a.Metadata, &a.CreatedAt, &a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *SQLiteAssetRepository) Update(ctx context.Context, asset *models.Asset, prevValue float64) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	result, err := tx.ExecContext(ctx,
		`UPDATE assets SET name=?, asset_type=?, current_value=?, currency=?, account_id=?, metadata=?, updated_at=datetime('now')
		 WHERE id=? AND user_id=? AND deleted_at IS NULL`,
		asset.Name, asset.AssetType, asset.CurrentValue, asset.Currency, asset.AccountID, asset.Metadata,
		asset.ID, asset.UserID,
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

	if asset.CurrentValue != prevValue {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO asset_history (id, asset_id, value, recorded_at) VALUES (?, ?, ?, datetime('now'))`,
			uuid.New().String(), asset.ID, asset.CurrentValue,
		)
		if err != nil {
			return false, err
		}
	}

	return true, tx.Commit()
}

func (r *SQLiteAssetRepository) SoftDelete(ctx context.Context, userID, id string) (bool, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE assets SET deleted_at=datetime('now'), updated_at=datetime('now')
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

func (r *SQLiteAssetRepository) GetHistory(ctx context.Context, userID, assetID string, since *time.Time) ([]models.AssetHistory, error) {
	exists, err := r.assetExists(ctx, userID, assetID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	query := `SELECT id, asset_id, value, recorded_at FROM asset_history WHERE asset_id = ?`
	args := []interface{}{assetID}

	if since != nil {
		query += " AND recorded_at >= ?"
		args = append(args, since.Format(time.RFC3339))
	}

	query += " ORDER BY recorded_at ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var entries []models.AssetHistory
	for rows.Next() {
		var h models.AssetHistory
		if err := rows.Scan(&h.ID, &h.AssetID, &h.Value, &h.RecordedAt); err != nil {
			return nil, err
		}
		entries = append(entries, h)
	}
	return entries, rows.Err()
}

type AllocationRow struct {
	AssetType  string
	TotalValue float64
}

func (r *SQLiteAssetRepository) GetAllocation(ctx context.Context, userID string) ([]AllocationRow, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT asset_type, SUM(current_value) as total_value
		 FROM assets WHERE user_id = ? AND deleted_at IS NULL
		 GROUP BY asset_type ORDER BY total_value DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []AllocationRow
	for rows.Next() {
		var row AllocationRow
		if err := rows.Scan(&row.AssetType, &row.TotalValue); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

type AllocationSnapshot struct {
	Date        string
	Allocations []AllocationRow
}

func (r *SQLiteAssetRepository) GetAllocationOverTime(ctx context.Context, userID string, since *time.Time) ([]AllocationSnapshot, error) {
	query := `SELECT date(ah.recorded_at) as snap_date, a.asset_type, SUM(ah.value) as total_value
		 FROM asset_history ah
		 INNER JOIN assets a ON ah.asset_id = a.id
		 WHERE a.user_id = ? AND a.deleted_at IS NULL`
	args := []interface{}{userID}

	if since != nil {
		query += " AND ah.recorded_at >= ?"
		args = append(args, since.Format(time.RFC3339))
	}

	query += " GROUP BY snap_date, a.asset_type ORDER BY snap_date ASC, a.asset_type ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	snapshotMap := make(map[string][]AllocationRow)
	var dates []string

	for rows.Next() {
		var date, assetType string
		var totalValue float64
		if err := rows.Scan(&date, &assetType, &totalValue); err != nil {
			return nil, err
		}
		if _, exists := snapshotMap[date]; !exists {
			dates = append(dates, date)
		}
		snapshotMap[date] = append(snapshotMap[date], AllocationRow{
			AssetType:  assetType,
			TotalValue: totalValue,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	snapshots := make([]AllocationSnapshot, len(dates))
	for i, date := range dates {
		snapshots[i] = AllocationSnapshot{
			Date:        date,
			Allocations: snapshotMap[date],
		}
	}
	return snapshots, nil
}

func (r *SQLiteAssetRepository) assetExists(ctx context.Context, userID, assetID string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM assets WHERE id = ? AND user_id = ? AND deleted_at IS NULL",
		assetID, userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
