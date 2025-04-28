package controllers

import (
	"RyanForce/config"
	"RyanForce/utils"
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// ClearDatabase deletes all records from the users and tickets tables.
// THIS ACTION IS IRREVERSIBLE AND SHOULD ONLY BE USED BY ADMINS.
// It's typically triggered from the CLI with a confirmation prompt.
// Useful for development resets.
func ClearDatabase(confirm bool) {
	utils.LogWarning("[Maintenance] Admin initiated full database wipe.")

	if confirm {
		fmt.Print("Are you sure you want to delete all users and tickets? Type 'yes' to confirm: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(input)) != "yes" {
			fmt.Println("Cancelled.")
			utils.LogInfo("[Maintenance] Database wipe cancelled by admin.")
			return
		}
	}

	// Continue deleting
	if err := config.DB.Exec("DELETE FROM comments;").Error; err != nil {
		utils.LogError("[Maintenance] Failed to delete comments", err)
		fmt.Println("[Error] Could not delete comments.")
	}

	if err := config.DB.Exec("DELETE FROM tickets;").Error; err != nil {
		utils.LogError("[Maintenance] Failed to delete tickets", err)
		fmt.Println("[Error] Could not delete tickets.")
	}

	if err := config.DB.Exec("DELETE FROM users;").Error; err != nil {
		utils.LogError("[Maintenance] Failed to delete users", err)
		fmt.Println("[Error] Could not delete users.")
	}

	if err := config.DB.Exec("DELETE FROM accounts;").Error; err != nil {
		utils.LogError("[Maintenance] Failed to delete accounts", err)
		fmt.Println("[Error] Could not delete accounts.")
	}

	utils.LogInfo("[Maintenance] Database cleared successfully.")
	fmt.Println("Database cleared.")

	// Immediately reseed
	fmt.Println("Reseeding demo data...")
	utils.LogInfo("[Maintenance] Starting reseed of demo data.")
	SeedDemoData()
	utils.ClearSession()

	// Add pause and success message after reseeding
	time.Sleep(500 * time.Millisecond)
	fmt.Println("\033[32m[✔] Database reseed complete.\033[0m")
	utils.LogInfo("[Maintenance] Database reseed finished successfully.")

	// Clear session so old token doesn't break login
	utils.ClearSession()
	fmt.Println("Old session cleared. Please log in again.")
}
