package queries

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/jamespsullivan/pennywise/internal/models"
	"github.com/jamespsullivan/pennywise/internal/observability"
)

type SQLiteAuditLogRepository struct {
	db *sql.DB
}

func NewAuditLogRepository(db *sql.DB) *SQLiteAuditLogRepository {
	return &SQLiteAuditLogRepository{db: db}
}

func (r *SQLiteAuditLogRepository) Record(ctx context.Context, entry *models.AuditLog) error {
	start := time.Now()
	defer func() { observability.RecordDBQuery("record_audit_log", time.Since(start)) }()

	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO audit_log (id, user_id, entity_type, entity_id, action, previous_data, new_data)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.UserID, entry.EntityType, entry.EntityID, entry.Action, entry.PreviousData, entry.NewData,
	)
	return err
}
