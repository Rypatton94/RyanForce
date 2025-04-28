package controllers

import (
	"RyanForce/config"
	"RyanForce/models"
	"RyanForce/utils"
	"fmt"
)

// SeedDemoUsers creates a small set of demo users in the database.
// These are pre-defined users with roles: admin, tech, and client.
func SeedDemoUsers() {
	users := []struct {
		Email    string
		Password string
		Role     string
	}{
		{"admin@example.com", "Admin123!", "admin"},
		{"tech@example.com", "Tech123!", "tech"},
		{"client@example.com", "Client123!", "client"},
	}

	for _, u := range users {
		hash, err := utils.HashPassword(u.Password)
		if err != nil {
			utils.LogError(fmt.Sprintf("[Seed] Failed to hash password for %s", u.Email), err)
			continue
		}

		user := models.User{
			Email:        u.Email,
			PasswordHash: hash,
			Role:         u.Role,
		}

		if err := config.DB.Create(&user).Error; err != nil {
			utils.LogError(fmt.Sprintf("[Seed] Failed to create user %s", u.Email), err)
		} else {
			fmt.Printf("Created user %s [%s]\n", u.Email, u.Role)
			utils.LogInfo(fmt.Sprintf("[Seed] User created: %s (%s)", u.Email, u.Role))
		}
	}
}

// MaybeSeedDemoUsers checks if the users table is empty.
// If no users are found, it runs SeedDemoUsers() to populate the DB with test accounts.
func MaybeSeedDemoUsers() {
	var count int64
	config.DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		fmt.Println("No users found — seeding demo users...")
		utils.LogWarning("[Seed] No users found. Seeding demo users.")
		SeedDemoUsers()
	}
}
