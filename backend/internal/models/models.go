package models

import "time"

type User struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Account struct {
	ID          string
	UserID      string
	Name        string
	Institution string
	AccountType string
	Currency    string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type Transaction struct {
	ID                     string
	UserID                 string
	AccountID              string
	Type                   string
	Category               string
	Amount                 float64
	Currency               string
	Date                   time.Time
	Notes                  *string
	IsRecurring            bool
	RecurringTransactionID *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
	DeletedAt              *time.Time
	Tags                   []string
}

type TransactionTag struct {
	ID            string
	TransactionID string
	Tag           string
}

type AuditLog struct {
	ID           string
	UserID       string
	EntityType   string
	EntityID     string
	Action       string
	PreviousData *string
	NewData      *string
	CreatedAt    time.Time
}

type FailedRequest struct {
	ID             string
	RequestID      *string
	UserID         *string
	Method         string
	Path           string
	StatusCode     int
	RequestBody    *string
	RequestHeaders *string
	ErrorCode      *string
	ErrorMessage   *string
	StackTrace     *string
	CreatedAt      time.Time
}
