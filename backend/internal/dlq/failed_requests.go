package dlq

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/jamespsullivan/pennywise/internal/models"
)

type SQLiteFailedRequestWriter struct {
	db *sql.DB
}

func NewFailedRequestWriter(db *sql.DB) *SQLiteFailedRequestWriter {
	return &SQLiteFailedRequestWriter{db: db}
}

func (w *SQLiteFailedRequestWriter) Write(ctx context.Context, entry *models.FailedRequest) error {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	_, err := w.db.ExecContext(ctx,
		`INSERT INTO failed_requests (id, request_id, user_id, method, path, status_code, request_body, request_headers, error_code, error_message)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.ID, entry.RequestID, entry.UserID, entry.Method, entry.Path, entry.StatusCode,
		entry.RequestBody, entry.RequestHeaders, entry.ErrorCode, entry.ErrorMessage,
	)
	return err
}
