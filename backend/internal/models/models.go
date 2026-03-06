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
