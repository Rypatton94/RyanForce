package controllers

import (
	"RyanForce/config"
	"RyanForce/models"
	"RyanForce/utils"
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ListUsers prints all users in the system.
func ListUsers() {
	var users []models.User
	if err := config.DB.Preload("Account").Find(&users).Error; err != nil {
		fmt.Println("[Error] Could not retrieve users.")
		utils.LogError("[Admin] Failed to retrieve users", err)
		return
	}

	fmt.Println("\nRegistered Users")
	fmt.Println("------------------")
	for _, u := range users {
		accountName := "None"
		if u.AccountID != nil && u.Account.Name != "" {
			accountName = u.Account.Name
		}
		fmt.Printf("ID: %d | Email: %s | Role: %s | Account: %s\n", u.ID, u.Email, u.Role, accountName)
	}
}

// DeleteUser removes a user by their ID.
func DeleteUser(userID uint) {
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		fmt.Println("[Error] User not found.")
		utils.LogWarning(fmt.Sprintf("[Admin] Attempted to delete non-existent user ID %d", userID))
		return
	}

	if err := config.DB.Delete(&user).Error; err != nil {
		fmt.Println("[Error] Could not delete user.")
		return
	}

	fmt.Printf("User ID %d deleted.\n", userID)
	utils.LogInfo(fmt.Sprintf("[Admin] Deleted user ID %d (%s)", user.ID, user.Email))
}

// HandleListUsers displays a list of all users (admin only).
func HandleListUsers() {
	claims, err := utils.LoadClaims()
	if err != nil || claims == nil {
		return
	}
	if claims.Role != "admin" {
		fmt.Println("[Error] Only admins can view user list.")
		return
	}

	ListUsers()
}

// HandleDeleteUser prompts for a user ID and deletes that user (admin only).
func HandleDeleteUser() {
	claims, err := utils.LoadClaims()
	if err != nil || claims == nil {
		return
	}
	if claims.Role != "admin" {
		fmt.Println("[Error] Only admins can delete users.")
		return
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("User ID to delete: ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	id64, err := strconv.ParseUint(input, 10, 64)
	if err != nil {
		fmt.Println("[Error] Invalid user ID.")
		return
	}

	utils.LogInfo(fmt.Sprintf("[Audit] Admin %d attempting to delete user ID %d", claims.UserID, id64))
	DeleteUser(uint(id64))
}

// CreateAccount adds a new account to the database.
func CreateAccount(name, domain, address, notes string) error {
	account := models.Account{
		Name:    name,
		Domain:  domain,
		Address: address,
		Notes:   notes,
	}
	if err := config.DB.Create(&account).Error; err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	utils.LogInfo(fmt.Sprintf("[Admin] Created account: %s", name))
	return nil
}

// AssignUserToAccount links a user to an account.
func AssignUserToAccount(userID, accountID uint) error {
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	user.AccountID = &accountID
	if err := config.DB.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to assign account: %w", err)
	}
	utils.LogInfo(fmt.Sprintf("[Admin] User %d assigned to account %d", userID, accountID))
	return nil
}
