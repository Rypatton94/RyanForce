package models

import "time"

type Comment struct {
	ID          uint      `gorm:"primaryKey"`
	TicketID    uint      `gorm:"not null"`           // Foreign key to the related ticket
	AuthorID    uint      `gorm:"not null"`           // ID of the user who authored the comment
	AuthorEmail string    `gorm:"not null"`           // Email of the user who authored the comment
	Content     string    `gorm:"type:text;not null"` // Body of the comment
	CreatedAt   time.Time // Timestamp of when the comment was posted
}
