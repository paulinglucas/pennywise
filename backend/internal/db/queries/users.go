package queries

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jamespsullivan/pennywise/internal/models"
	"github.com/jamespsullivan/pennywise/internal/observability"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailTaken = errors.New("email already taken")

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

func (r *SQLiteUserRepository) CountUsers(ctx context.Context) (int, error) {
	start := time.Now()
	defer func() { observability.RecordDBQuery("count_users", time.Since(start)) }()

	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *SQLiteUserRepository) CreateUser(ctx context.Context, user *models.User) error {
	start := time.Now()
	defer func() { observability.RecordDBQuery("create_user", time.Since(start)) }()

	_, err := r.db.ExecContext(ctx,
		"INSERT INTO users (id, email, name, password_hash) VALUES (?, ?, ?, ?)",
		user.ID, user.Email, user.Name, user.PasswordHash,
	)
	if err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return ErrEmailTaken
	}
	return err
}
