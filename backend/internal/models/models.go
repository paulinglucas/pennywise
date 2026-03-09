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
	GroupID                *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
	DeletedAt              *time.Time
	Tags                   []string
}

type TransactionGroup struct {
	ID        string
	UserID    string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
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

type Asset struct {
	ID           string
	UserID       string
	AccountID    *string
	Name         string
	AssetType    string
	CurrentValue float64
	Currency     string
	Metadata     *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

type AssetHistory struct {
	ID         string
	AssetID    string
	Value      float64
	RecordedAt time.Time
}

type Goal struct {
	ID              string
	UserID          string
	Name            string
	GoalType        string
	TargetAmount    float64
	CurrentAmount   float64
	Deadline        *time.Time
	LinkedAccountID *string
	PriorityRank    int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type RecurringTransaction struct {
	ID             string
	UserID         string
	AccountID      string
	Type           string
	Category       string
	Amount         float64
	Currency       string
	Frequency      string
	NextOccurrence time.Time
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}

type Alert struct {
	ID                string
	UserID            string
	AlertType         string
	Message           string
	Severity          string
	IsRead            bool
	RelatedEntityType *string
	RelatedEntityID   *string
	CreatedAt         time.Time
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
