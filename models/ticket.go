package models

import (
	"time"
)

// Ticket represents a support ticket in the system.
type Ticket struct {
	ID           uint `gorm:"primaryKey"`
	Title        string
	Description  string
	Priority     string
	Status       string
	ClientID     uint
	Client       User `gorm:"foreignKey:ClientID"`
	TechID       *uint
	AssignedTech *User `gorm:"foreignKey:TechID"`
	ClosedAt     *time.Time
	Comments     []Comment `gorm:"foreignKey:TicketID"`
	SkillsNeeded string    `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
