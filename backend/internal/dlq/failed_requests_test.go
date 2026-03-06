package dlq_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jamespsullivan/pennywise/internal/db"
	"github.com/jamespsullivan/pennywise/internal/dlq"
	"github.com/jamespsullivan/pennywise/internal/models"
)

func TestFailedRequestWriter_Write(t *testing.T) {
	database, err := db.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close() })

	err = db.Migrate(database)
	require.NoError(t, err)

	writer := dlq.NewFailedRequestWriter(database)

	reqID := "req-001"
	userID := "usr-001"
	body := `{"bad":"data"}`
	errCode := "VALIDATION_FAILED"
	errMsg := "Invalid field"
	err = writer.Write(context.Background(), &models.FailedRequest{
		RequestID:    &reqID,
		UserID:       &userID,
		Method:       "POST",
		Path:         "/api/v1/transactions",
		StatusCode:   400,
		RequestBody:  &body,
		ErrorCode:    &errCode,
		ErrorMessage: &errMsg,
	})
	require.NoError(t, err)

	var count int
	err = database.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM failed_requests WHERE method = 'POST' AND path = '/api/v1/transactions'",
	).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
