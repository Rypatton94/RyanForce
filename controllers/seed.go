package controllers

import (
	"RyanForce/config"
	"RyanForce/models"
	"RyanForce/utils"
	"encoding/json"
	"fmt"
)

// SeedDemoData creates realistic demo data for accounts, users, and tickets.
func SeedDemoData() {
	// Create Accounts
	accounts := []models.Account{
		{Name: "Acme Corp", Domain: "acme.com", Address: "123 Main St, Anytown, USA"},
		{Name: "Globex Inc", Domain: "globex.com", Address: "456 Elm St, Othertown, USA"},
	}

	for i := range accounts {
		if err := config.DB.Create(&accounts[i]).Error; err != nil {
			utils.LogError("[Seed] Failed to create account", err)
		} else {
			utils.LogInfo(fmt.Sprintf("[Seed] Account created: %s", accounts[i].Name))
		}
	}

	// Helper function to marshal skills
	marshalSkills := func(skills []string) string {
		bytes, _ := json.Marshal(skills)
		return string(bytes)
	}

	// Create Users
	users := []struct {
		Email    string
		Password string
		Role     string
		Name     string
		Skills   []string
		Account  string // Match to account Name
	}{
		// Admin
		{"admin@example.com", "Admin123!", "admin", "Admin User", nil, ""},

		// Techs
		{"alice.tech@example.com", "Tech123!", "tech", "Alice Anderson", []string{"Networking", "Security", "Windows"}, ""},
		{"bob.tech@example.com", "Tech123!", "tech", "Bob Brown", []string{"MacOS", "Hardware Repair", "Customer Support"}, ""},
		{"charlie.tech@example.com", "Tech123!", "tech", "Charlie Clark", []string{"Linux", "Cloud", "Scripting"}, ""},

		// Clients
		{"cindy.client@acme.com", "Client123!", "client", "Cindy Client", nil, "Acme Corp"},
		{"gary.client@globex.com", "Client123!", "client", "Gary Globex", nil, "Globex Inc"},
	}

	createdUsers := make(map[string]models.User)

	for _, u := range users {
		hash, err := utils.HashPassword(u.Password)
		if err != nil {
			utils.LogError(fmt.Sprintf("[Seed] Failed to hash password for %s", u.Email), err)
			continue
		}

		var accountID *uint
		if u.Account != "" {
			var account models.Account
			if err := config.DB.Where("name = ?", u.Account).First(&account).Error; err == nil {
				accountID = &account.ID
			}
		}

		user := models.User{
			Email:        u.Email,
			PasswordHash: hash,
			Role:         u.Role,
			Name:         u.Name,
			AccountID:    accountID,
		}

		// Marshal tech skills if user is a tech
		if len(u.Skills) > 0 {
			user.Skills = marshalSkills(u.Skills)
		}

		if err := config.DB.Create(&user).Error; err != nil {
			utils.LogError(fmt.Sprintf("[Seed] Failed to create user %s", u.Email), err)
		} else {
			createdUsers[u.Email] = user
			utils.LogInfo(fmt.Sprintf("[Seed] User created: %s (%s)", u.Email, u.Role))
		}
	}

	// Create Tickets
	tickets := []struct {
		Title             string
		Description       string
		Priority          string
		Status            string
		SkillsNeeded      []string
		ClientEmail       string
		AssignedTechEmail string
	}{
		{"Setup VPN Access", "Client needs secure VPN access configured.", "high", "open", []string{"Networking", "Security"}, "cindy.client@acme.com", "alice.tech@example.com"},
		{"Broken MacBook Pro", "Laptop not booting after update.", "medium", "open", []string{"MacOS", "Hardware Repair"}, "gary.client@globex.com", "bob.tech@example.com"},
		{"Cloud Backup Failure", "Scheduled backups to cloud are failing nightly.", "critical", "open", []string{"Cloud", "Linux"}, "cindy.client@acme.com", "charlie.tech@example.com"},
		{"Password Reset", "User forgot password and needs reset.", "low", "open", []string{"Customer Support"}, "gary.client@globex.com", "bob.tech@example.com"},
		{"New Laptop Setup", "Prepare a new laptop for onboarding.", "medium", "open", []string{"Windows"}, "cindy.client@acme.com", ""},
		{"Server Monitoring Scripts Broken", "Monitoring scripts aren't reporting server stats.", "high", "open", []string{"Scripting"}, "gary.client@globex.com", ""},
	}

	for _, t := range tickets {
		client, ok := createdUsers[t.ClientEmail]
		if !ok {
			utils.LogWarning(fmt.Sprintf("[Seed] Client not found: %s", t.ClientEmail))
			continue
		}

		var techID *uint
		if t.AssignedTechEmail != "" {
			tech, ok := createdUsers[t.AssignedTechEmail]
			if ok {
				techID = &tech.ID
			}
		}

		ticket := models.Ticket{
			Title:        t.Title,
			Description:  t.Description,
			Priority:     t.Priority,
			Status:       t.Status,
			ClientID:     client.ID,
			TechID:       techID,
			SkillsNeeded: marshalSkills(t.SkillsNeeded),
		}

		if err := config.DB.Create(&ticket).Error; err != nil {
			utils.LogError(fmt.Sprintf("[Seed] Failed to create ticket: %s", t.Title), err)
		} else {
			utils.LogInfo(fmt.Sprintf("[Seed] Ticket created: %s", t.Title))
		}
	}
}
