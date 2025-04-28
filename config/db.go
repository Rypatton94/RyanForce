package config

import (
	"RyanForce/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
	"path/filepath"
)

// DB is the global GORM database connection used throughout the app.
// It’s initialized once in Connect() and shared across all packages.
var DB *gorm.DB

// Connect sets up the SQLite database connection, ensures the directory exists,
// and runs auto-migration to apply model schemas (User, Ticket).
func Connect() {
	// Ensure the 'database/' directory exists
	err := os.MkdirAll(filepath.Join(".", "database"), os.ModePerm)
	if err != nil {
		log.Fatalf("failed to create database directory: %v", err)
	}

	// Open a connection to the SQLite DB and assign it to the global DB variable
	var dbErr error
	DB, dbErr = gorm.Open(sqlite.Open("database/ryanforce.db"), &gorm.Config{})
	if dbErr != nil {
		log.Fatalf("failed to connect to database: %v", dbErr)
	}

	// Automatically create or update database tables to match model structs.
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Ticket{},
		&models.Comment{}, // <-- Add this!
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

}
