package queries

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jamespsullivan/pennywise/internal/models"
	"github.com/jamespsullivan/pennywise/internal/observability"
)

var ErrUserNotFound = errors.New("user not found")

type SQLiteUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *SQLiteUserRepository {
	return &SQLiteUserRepository{db: db}
}

func (r *SQLiteUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("get_user_by_email", time.Since(start)) }()

	var user models.User
	err := r.db.QueryRowContext(ctx,
		"SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SQLiteUserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("get_user_by_id", time.Since(start)) }()

	var user models.User
	err := r.db.QueryRowContext(ctx,
		"SELECT id, email, name, password_hash, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
