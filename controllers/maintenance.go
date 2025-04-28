package controllers

import (
	"RyanForce/config"
	"RyanForce/utils"
	"fmt"
)

// ClearDatabase deletes all records from the users and tickets tables.
// THIS ACTION IS IRREVERSIBLE AND SHOULD ONLY BE USED BY ADMINS.
// It's typically triggered from the CLI with a confirmation prompt.
// Useful for development resets.
func ClearDatabase() {
	utils.LogWarning("[Maintenance] Admin initiated full database wipe.")

	if err := config.DB.Exec("DELETE FROM users;").Error; err != nil {
		utils.LogError("[Maintenance] Failed to delete users", err)
		fmt.Println("[Error] Could not delete users.")
		return
	}

	if err := config.DB.Exec("DELETE FROM tickets;").Error; err != nil {
		utils.LogError("[Maintenance] Failed to delete tickets", err)
		fmt.Println("[Error] Could not delete tickets.")
		return
	}

	utils.LogInfo("[Maintenance] Users and tickets successfully deleted.")
	fmt.Println("Database cleared.")
}
