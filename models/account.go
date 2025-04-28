package models

import "gorm.io/gorm"

type Account struct {
	gorm.Model

	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"uniqueIndex"`
	Domain  string
	Address string
	Notes   string

	Users []User `gorm:"foreignKey:AccountID"` // One-to-many
}
