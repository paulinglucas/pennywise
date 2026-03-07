package queries

import (
	"context"
	"database/sql"
	"time"

	"github.com/jamespsullivan/pennywise/internal/models"
	"github.com/jamespsullivan/pennywise/internal/observability"
)

type SQLiteAlertRepository struct {
	db *sql.DB
}

func NewAlertRepository(db *sql.DB) *SQLiteAlertRepository {
	return &SQLiteAlertRepository{db: db}
}

func (r *SQLiteAlertRepository) List(ctx context.Context, userID string, page, perPage int) ([]models.Alert, int, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("list_alerts", time.Since(start)) }()

	var total int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM alerts WHERE user_id = ? AND is_read = 0",
		userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, alert_type, message, severity, is_read, related_entity_type, related_entity_id, created_at
		 FROM alerts WHERE user_id = ? AND is_read = 0
		 ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		userID, perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = rows.Close() }()

	var items []models.Alert
	for rows.Next() {
		var a models.Alert
		err := rows.Scan(&a.ID, &a.UserID, &a.AlertType, &a.Message, &a.Severity,
			&a.IsRead, &a.RelatedEntityType, &a.RelatedEntityID, &a.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, a)
	}
	return items, total, rows.Err()
}

func (r *SQLiteAlertRepository) MarkRead(ctx context.Context, userID, id string) (bool, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("mark_alert_read", time.Since(start)) }()

	result, err := r.db.ExecContext(ctx,
		"UPDATE alerts SET is_read = 1 WHERE id = ? AND user_id = ? AND is_read = 0",
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
