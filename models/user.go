package models

import (
	"gorm.io/gorm"
	"time"
)

// User represents an authenticated system user.
// A user can be an administrator, technician, or client depending on their role.
type User struct {
	gorm.Model

	ID             uint   `gorm:"primaryKey"`
	Email          string `gorm:"uniqueIndex"`
	PasswordHash   string
	Name           string
	Skills         string
	Role           string
	FailedAttempts int
	IsLocked       bool
	LastLogin      *time.Time

	AccountID *uint   // Foreign key
	Account   Account `gorm:"foreignKey:AccountID"`
}
