package handlers

import (
	"bufio"
	"fmt"
	"github.com/manifoldco/promptui"
	"os"
	"strconv"
	"strings"
	"time"

	"RyanForce/config"
	"RyanForce/controllers"
	"RyanForce/models"
	"RyanForce/utils"
	"golang.org/x/term"
)

// This file handles all CLI commands entered by the user. Each command is routed to a handler function
// that performs the appropriate logic based on the user's role and input. It also supports session-based login.

// HandleCommand maps user input to application functionality.
// Supports aliases for most commands and restricts access to certain features based on user roles.
func HandleCommand(cmd string) {
	switch cmd {
	case "register", "r", "signup":
		handleRegister() // New user registration
	case "create-account":
		handleCreateAccount()
	case "assign-account":
		handleAssignUserToAccount()
	case "list-accounts":
		handleListAccounts()
	case "delete-account":
		handleDeleteAccount()
	case "login", "l", "auth":
		handleLogin() // Log in to an existing account
	case "logout", "lo", "exit":
		handleLogout() // End session and remove token
	case "whoami", "me", "status":
		handleWhoami() // Show current user info
	case "reset-password", "rp", "reset":
		handleResetPassword() // User-initiated password reset
	case "admin-reset-password", "arp", "admin-reset":
		handleAdminResetPassword() // Admin resets another user's password
	case "create-ticket", "ct", "new":
		handleCreateTicket() // Client creates a new support ticket
	case "update-ticket", "ut", "edit":
		handleUpdateTicket() // Tech or admin updates an existing ticket
	case "comment-ticket", "ctc":
		handleCommentTicket() // Adds comment to ticket
	case "list-tickets", "lt", "list":
		handleListTickets() // Lists tickets relevant to the current user
	case "assign-ticket", "at", "assign":
		handleAssignTicket() // Admin assigns a tech to a ticket
	case "view-ticket", "vt", "show":
		handleViewTicket() // Shows ticket by ID, respecting role assignment
	case "filter-tickets", "ft", "search":
		handleFilterTickets()
	case "delete-ticket", "dt", "remove":
		handleDeleteTicket() // Delete ticket, admins only
	case "view-logs", "logs", "tail":
		handleViewLogs() // Shows recent logs
	case "list-users", "lu":
		controllers.HandleListUsers() // Allows admin's to view list of all users
	case "delete-user", "du":
		controllers.HandleDeleteUser() // Allows admin's to delete users
	case "report-status":
		controllers.ReportStatus() // Shows breakdown of ticket counts by status
	case "report-priority":
		controllers.ReportPriority() // Shows tickets at each priority level
	case "report-unassigned":
		controllers.ReportUnassigned() // Shows unassigned tickets
	case "report-resolve-time":
		controllers.ReportResolutionMetrics() // Shows tickets response times based on priority
	case "report-overdue":
		controllers.ReportOverdueTickets() // Shows tickets that have missed their response time
	case "report-all":
		controllers.ReportAll() // Shows all reports
	case "export-tickets":
		controllers.ExportTicketsCSV() // Exports tickets to CSV file
	case "help", "h", "?":
		handleHelp() // Display help/command list
	case "seed-demo":
		controllers.SeedDemoData() // Manually trigger seeding
	case "clear-db":
		handleClearDB() // DANGER - admin-only wipe of all users and tickets
	default:
		fmt.Println("[Error] Unknown command. Try 'help' or 'whoami'")
	}
}

// PromptLogin asks the user for email and password, attempts to authenticate them,
// and returns a JWT token on success. Allows up to 3 attempts.
func PromptLogin() (string, *utils.Claims, error) {
	reader := bufio.NewReader(os.Stdin)

	for attempts := 1; attempts <= 3; attempts++ {
		fmt.Print("Email: ")
		email, _ := reader.ReadString('\n')
		email = strings.TrimSpace(email)

		fmt.Print("Password: ")
		bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			fmt.Println("[Error] Failed to read password.")
			continue
		}
		password := string(bytePassword)

		token, err := controllers.Authenticate(email, password)
		if err == nil {
			claims, err := utils.ParseJWT(token)
			if err != nil {
				fmt.Println("[Error] Unable to parse session token.")
				utils.LogError("[Login] Failed to parse session token", err)
				return "", nil, err
			}
			utils.LogInfo(fmt.Sprintf("[Login] Successful login for %s", email))
			return token, claims, nil
		}

		fmt.Println("[Login Failed]", err)
		utils.LogWarning(fmt.Sprintf("[Login] Attempt %d failed for %s", attempts, email))
		if attempts < 3 {
			fmt.Printf("Try again (%d/3)\n\n", attempts)
		}
	}

	utils.LogWarning("[Login] Too many failed login attempts")
	return "", nil, fmt.Errorf("too many failed login attempts")
}

// DisplayDashboard prints a common welcome splash providing the User ID, session expiry, and role specific reports.
func DisplayDashboard(claims *utils.Claims) {
	fmt.Printf("\nWelcome, %s!\n", strings.Title(claims.Role))
	fmt.Printf("User ID       : %d\n", claims.UserID)
	fmt.Printf("Session Expires: %s\n", claims.ExpiresAt.Time.Format("2006-01-02 15:04:05"))

	switch claims.Role {
	case "admin":
		fmt.Println("\nAdmin Summary:")
		controllers.ReportStatus()
		controllers.ReportUnassigned()
	case "tech":
		fmt.Println("\nAssigned Tickets:")
		controllers.ListTickets(claims.UserID, claims.Role)
	case "client":
		fmt.Println("\nYour Active Tickets:")
		controllers.ListTickets(claims.UserID, claims.Role)
	}

	utils.LogInfo(fmt.Sprintf("[Dashboard] Displayed for user %d (%s)", claims.UserID, claims.Role))
}

// promptPasswordTwice prompts the user to enter and confirm a password.
// If both entries match, returns the password string. Otherwise, returns an error after 3 failed attempts.
func promptPasswordTwice(label string) (string, error) {
	for attempts := 0; attempts < 3; attempts++ {
		fmt.Print(label + ": ")
		pass1Bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return "", fmt.Errorf("failed to read password")
		}

		fmt.Print("Confirm " + label + ": ")
		pass2Bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return "", fmt.Errorf("failed to read confirmation")
		}

		pass1 := string(pass1Bytes)
		pass2 := string(pass2Bytes)

		if pass1 == pass2 {
			return pass1, nil
		}

		fmt.Println("[Error] Passwords do not match. Try again.")
		utils.LogWarning("[Password Entry] Mismatched password confirmation")
	}
	return "", fmt.Errorf("too many mismatches")
}

// handleRegister walks the user through creating a new account.
// Prompts for email, password (with confirmation), and a role (client, tech, or admin).
func handleRegister() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	password, err := promptPasswordTwice("Password")
	if err != nil {
		fmt.Println("[Error]", err)
		utils.LogWarning("[Register] Password mismatch or entry failure")
		return
	}
	fmt.Println()

	roleOptions := []string{"client", "tech", "admin"}
	role, err := utils.PromptSelect("Select Role", roleOptions, 0)
	if err != nil {
		fmt.Println("Registration cancelled.")
		utils.LogInfo("[Register] User cancelled role selection")
		return
	}

	if !utils.IsValidPassword(password) {
		fmt.Println("[Error] Password must be 8–32 characters with an uppercase letter, digit, and special character.")
		return
	}

	controllers.Register(email, password, role)
	utils.LogInfo(fmt.Sprintf("[Register] Attempted registration for %s as %s", email, role))
}

func handleCreateAccount() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Account Name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Domain: ")
	domain, _ := reader.ReadString('\n')
	domain = strings.TrimSpace(domain)

	fmt.Print("Address: ")
	address, _ := reader.ReadString('\n')
	address = strings.TrimSpace(address)

	fmt.Print("Notes: ")
	notes, _ := reader.ReadString('\n')
	notes = strings.TrimSpace(notes)

	if err := controllers.CreateAccount(name, domain, address, notes); err != nil {
		fmt.Println("[Error]", err)
	} else {
		fmt.Println("✅ Account created.")
	}
}

func handleAssignUserToAccount() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("User ID: ")
	userIDStr, _ := reader.ReadString('\n')
	userIDStr = strings.TrimSpace(userIDStr)
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		fmt.Println("[Error] Invalid User ID.")
		return
	}

	fmt.Print("Account ID: ")
	accountIDStr, _ := reader.ReadString('\n')
	accountIDStr = strings.TrimSpace(accountIDStr)
	accountID, err := strconv.ParseUint(accountIDStr, 10, 64)
	if err != nil {
		fmt.Println("[Error] Invalid Account ID.")
		return
	}

	if err := controllers.AssignUserToAccount(uint(userID), uint(accountID)); err != nil {
		fmt.Println("[Error]", err)
	} else {
		fmt.Println("✅ User assigned to account.")
	}
}

func handleListAccounts() {
	var accounts []models.Account
	err := config.DB.Preload("Users").Find(&accounts).Error
	if err != nil {
		fmt.Println("[Error] Failed to load accounts:", err)
		utils.LogError("[Admin] Failed to load account list", err)
		return
	}

	fmt.Println("\nAccounts and Assigned Users")
	fmt.Println("------------------------------")
	for _, acct := range accounts {
		fmt.Printf("Account ID %d: %s (Domain: %s)\n", acct.ID, acct.Name, acct.Domain)
		if len(acct.Users) == 0 {
			fmt.Println("  No users assigned.")
		} else {
			for _, u := range acct.Users {
				fmt.Printf("  - [%d] %s (%s)\n", u.ID, u.Email, u.Role)
			}
		}
		fmt.Println()
	}
}

func handleDeleteAccount() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Account ID to delete: ")
	idStr, _ := reader.ReadString('\n')
	idStr = strings.TrimSpace(idStr)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		fmt.Println("[Error] Invalid Account ID.")
		return
	}
	accountID := uint(id)

	// Check for users assigned to this account
	var count int64
	err = config.DB.Model(&models.User{}).Where("account_id = ?", accountID).Count(&count).Error
	if err != nil {
		fmt.Println("[Error] Failed to check account associations:", err)
		utils.LogError("[Admin] Failed account user check before deletion", err)
		return
	}

	if count > 0 {
		fmt.Printf("[Error] Cannot delete account ID %d — %d user(s) are still assigned.\n", accountID, count)
		fmt.Println("Reassign or remove users before deleting this account.")
		return
	}

	fmt.Printf("Are you sure you want to delete account %d? Type 'yes' to confirm: ", accountID)
	confirm, _ := reader.ReadString('\n')
	if strings.ToLower(strings.TrimSpace(confirm)) != "yes" {
		fmt.Println("Cancelled.")
		return
	}

	err = config.DB.Delete(&models.Account{}, accountID).Error
	if err != nil {
		fmt.Println("[Error] Failed to delete account:", err)
		utils.LogError("[Admin] Failed to delete account", err)
	} else {
		fmt.Printf("✅ Account ID %d deleted.\n", accountID)
		utils.LogInfo(fmt.Sprintf("[Admin] Deleted account ID %d", accountID))
	}
}

// handleLogin authenticates a user and stores their session token.
// If login is successful, a confirmation is printed.
func handleLogin() {
	token, claims, err := PromptLogin()
	if err != nil {
		fmt.Println("[Error]", err)
		return
	}

	if err := utils.SaveSession(token); err != nil {
		fmt.Println("[Error] Failed to save session:", err)
		utils.LogError("[Login] Failed to save session", err)
		return
	}

	fmt.Printf("Login successful. Welcome %s.\n", strings.Title(claims.Role))
	fmt.Printf("Session expires at: %s\n", claims.ExpiresAt.Time.Format("2006-01-02 15:04:05"))
	utils.LogInfo(fmt.Sprintf("[Login] Session saved for user %d (%s)", claims.UserID, claims.Role))

	DisplayDashboard(claims)
}

// handleLogout clears the user's session file to end their current login.
func handleLogout() {
	err := utils.ClearSession()
	if err != nil {
		fmt.Println("[Error] Couldn't log out:", err)
		utils.LogError("[Logout] Failed to clear session", err)
	} else {
		fmt.Println("Logged out.")
		utils.LogInfo("[Logout] Session cleared")
	}
}

// handleWhoami prints out information about the current session.
// Displays user ID, role, and token expiration if a valid session exists.
func handleWhoami() {
	claims, err := utils.LoadClaims()
	if err != nil || claims == nil {
		return
	}

	fmt.Println("Logged in as:")
	fmt.Println("User ID   :", claims.UserID)
	fmt.Println("Role      :", claims.Role)
	fmt.Println("Expires At:", claims.ExpiresAt.Time.Format("2006-01-02 15:04:05"))
	utils.LogInfo(fmt.Sprintf("[Whoami] Session viewed by user %d (%s)", claims.UserID, claims.Role))
}

// handleResetPassword allows a user to update their own password.
// Prompts for email, old password, and new password (with confirmation).
func handleResetPassword() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Print("Old Password: ")
	oldPassBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		fmt.Println("[Error] Failed to read old password.")
		return
	}
	oldPassword := string(oldPassBytes)

	newPassword, err := promptPasswordTwice("New Password")
	if err != nil {
		fmt.Println("[Error]", err)
		return
	}
	fmt.Println()

	if !utils.IsValidPassword(newPassword) {
		fmt.Println("[Error] Password does not meet strength requirements.")
		return
	}
	err = controllers.ResetPassword(email, oldPassword, newPassword)
	if err != nil {
		fmt.Println("[Error]", err)
	} else {
		fmt.Println("Password updated successfully.")
	}
}

// handleAdminResetPassword enables an admin to reset another user's password.
// Requires a valid session with admin role, and prompts for the target user's email and new password.
func handleAdminResetPassword() {
	claims, err := utils.LoadClaims()
	if err != nil || claims == nil {
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Target User Email: ")
	targetEmail, _ := reader.ReadString('\n')
	targetEmail = strings.TrimSpace(targetEmail)

	newPassword, err := promptPasswordTwice("New Password")
	if err != nil {
		fmt.Println("[Error]", err)
		return
	}
	fmt.Println()

	if !utils.IsValidPassword(newPassword) {
		fmt.Println("[Error] Password does not meet strength requirements.")
		return
	}
	err = controllers.AdminResetPassword(claims.UserID, targetEmail, newPassword)
	if err != nil {
		fmt.Println("[Error]", err)
	} else {
		fmt.Printf("Password for %s has been reset successfully.\n", targetEmail)
	}
}

// handleCreateTicket prompts a logged-in client to submit a new support request.
// It collects the ticket title, description, priority, and initial status.
func handleCreateTicket() {
	claims, err := utils.LoadClaims()
	if err != nil || claims == nil {
		return
	}

	if claims.Role != "client" {
		fmt.Println("[Error] Only clients can create tickets.")
		utils.LogWarning(fmt.Sprintf("[CreateTicket] Unauthorized attempt by user %d (%s)", claims.UserID, claims.Role))
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Title: ")
	title, _ := reader.ReadString('\n')
	title = strings.TrimSpace(title)

	fmt.Print("Description: ")
	desc, _ := reader.ReadString('\n')
	desc = strings.TrimSpace(desc)

	priorityOptions := []string{"low", "medium", "high", "critical"}
	priority, err := utils.PromptSelect("Select Priority", priorityOptions, 1)
	if err != nil {
		fmt.Println("Priority selection cancelled.")
		utils.LogInfo("[CreateTicket] Priority selection cancelled")
		return
	}

	status := "initially reported"
	controllers.CreateTicket(title, desc, priority, status, claims.UserID)
	utils.LogInfo(fmt.Sprintf("[CreateTicket] New ticket created by user %d", claims.UserID))
}

// handleUpdateTicket allows a technician or admin the ability to modify an existing ticket.
func handleUpdateTicket() {
	claims, err := utils.LoadClaims()
	if err != nil || claims == nil {
		return
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Ticket ID: ")
	idStr, _ := reader.ReadString('\n')
	idStr = strings.TrimSpace(idStr)
	idUint64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		fmt.Println("[Error] Invalid ticket ID.")
		return
	}
	ticketID := uint(idUint64)

	var ticket models.Ticket
	result := config.DB.First(&ticket, ticketID)
	if result.Error != nil {
		fmt.Println("[Error] Ticket not found.")
		utils.LogWarning(fmt.Sprintf("[UpdateTicket] Ticket %d not found", ticketID))
		return
	}

	for {
		prompt := promptui.Select{
			Label: "Select an action for this ticket",
			Items: []string{"Update Description", "Update Priority", "Update Status", "Manage Comments", "Return"},
		}

		_, action, err := prompt.Run()
		if err != nil {
			utils.LogError("[CLI] Failed during update menu prompt", err)
			fmt.Println("[Error] Failed to choose action.")
			return
		}

		switch action {
		case "Update Description":
			fmt.Printf("New Description [%s]: ", ticket.Description)
			newDesc, _ := reader.ReadString('\n')
			newDesc = strings.TrimSpace(newDesc)
			if newDesc == "" {
				newDesc = ticket.Description
			}
			ticket.Description = newDesc

		case "Update Priority":
			priorityOptions := []string{"low", "medium", "high", "critical"}
			currentPriorityIndex := utils.IndexOf(ticket.Priority, priorityOptions)
			newPriority, err := utils.PromptSelect("Select Priority", priorityOptions, currentPriorityIndex)
			if err == nil {
				ticket.Priority = newPriority
			}

		case "Update Status":
			statusOptions := []string{
				"initially reported",
				"customer to follow up",
				"support to follow up",
				"working",
				"closed",
			}
			currentStatusIndex := utils.IndexOf(ticket.Status, statusOptions)
			newStatus, err := utils.PromptSelect("Select Status", statusOptions, currentStatusIndex)
			if err == nil {
				ticket.Status = newStatus
			}

		case "Manage Comments":
			handleManageComments(ticketID, claims.UserID, claims.Email)

		case "Return":
			// Save ticket updates before exiting
			err := config.DB.Save(&ticket).Error
			if err != nil {
				fmt.Println("[Error] Failed to save ticket updates.")
				utils.LogError(fmt.Sprintf("[UpdateTicket] Failed to save ticket %d", ticketID), err)
			} else {
				fmt.Println("[Success] Ticket updated successfully.")
				utils.LogInfo(fmt.Sprintf("[UpdateTicket] User %d updated ticket %d", claims.UserID, ticketID))
			}
			return
		}
	}
}

// handleManageComments launches a subprocess to add, edit, or delete comments for a given ticket
func handleManageComments(ticketID uint, userID uint, userEmail string) {
	for {
		fmt.Println("\nComment Management for Ticket:", ticketID)

		prompt := promptui.Select{
			Label: "Choose an action",
			Items: []string{"Add Comment", "Edit Comment", "Delete Comment", "Return"},
		}

		_, action, err := prompt.Run()
		if err != nil {
			utils.LogError("[CLI] Failed during comment action prompt", err)
			fmt.Println("[Error] Failed to choose action.")
			return
		}

		switch action {
		case "Add Comment":
			handleAddComment(ticketID, userID, userEmail)
		case "Edit Comment":
			handleEditComment(ticketID)
		case "Delete Comment":
			handleDeleteComment(ticketID)
		case "Return":
			return
		}
	}
}

func handleAddComment(ticketID uint, userID uint, userEmail string) {
	prompt := promptui.Prompt{
		Label: "Enter your comment",
	}

	commentText, err := prompt.Run()
	if err != nil || strings.TrimSpace(commentText) == "" {
		fmt.Println("[Info] No comment entered. Aborting add.")
		return
	}

	err = controllers.AddCommentToTicket(ticketID, commentText, userID, userEmail, "CLI-Local")
	if err != nil {
		fmt.Println("[Error] Failed to add comment:", err)
	} else {
		fmt.Println("[Success] Comment added successfully.")
	}
}

func handleEditComment(ticketID uint) {
	comments, err := controllers.GetCommentsForTicket(ticketID)
	if err != nil || len(comments) == 0 {
		fmt.Println("[Info] No comments to edit.")
		return
	}

	commentItems := []string{}
	for _, c := range comments {
		commentItems = append(commentItems, fmt.Sprintf("ID %d: %s", c.ID, truncate(c.Content, 30)))
	}

	prompt := promptui.Select{
		Label: "Select a comment to edit",
		Items: commentItems,
	}

	index, _, err := prompt.Run()
	if err != nil {
		fmt.Println("[Info] Edit canceled.")
		return
	}

	selectedComment := comments[index]

	newPrompt := promptui.Prompt{
		Label:   "Enter new text",
		Default: selectedComment.Content,
	}

	newText, err := newPrompt.Run()
	if err != nil || strings.TrimSpace(newText) == "" {
		fmt.Println("[Info] Edit canceled.")
		return
	}

	err = controllers.EditComment(selectedComment.ID, newText, "CLI-Local")
	if err != nil {
		fmt.Println("[Error] Failed to edit comment.")
	} else {
		fmt.Println("[Success] Comment updated.")
	}
}

func handleDeleteComment(ticketID uint) {
	comments, err := controllers.GetCommentsForTicket(ticketID)
	if err != nil || len(comments) == 0 {
		fmt.Println("[Info] No comments to delete.")
		return
	}

	commentItems := []string{}
	for _, c := range comments {
		commentItems = append(commentItems, fmt.Sprintf("ID %d: %s", c.ID, truncate(c.Content, 30)))
	}

	prompt := promptui.Select{
		Label: "Select a comment to delete",
		Items: commentItems,
	}

	index, _, err := prompt.Run()
	if err != nil {
		fmt.Println("[Info] Delete canceled.")
		return
	}

	selectedComment := comments[index]

	confirm := promptui.Prompt{
		Label: "Are you sure you want to delete this comment? (yes/no)",
		Validate: func(input string) error {
			if input != "yes" && input != "no" {
				return fmt.Errorf("please type 'yes' or 'no'")
			}
			return nil
		},
	}

	confirmation, err := confirm.Run()
	if err != nil || confirmation != "yes" {
		fmt.Println("[Info] Delete canceled.")
		return
	}

	err = controllers.DeleteComment(selectedComment.ID, "CLI-Local")
	if err != nil {
		fmt.Println("[Error] Failed to delete comment.")
	} else {
		fmt.Println("[Success] Comment deleted.")
	}
}

// truncate returns a shortened version of a string with ellipsis if too long
func truncate(text string, limit int) string {
	if len(text) > limit {
		return text[:limit] + "..."
	}
	return text
}

// handleListTickets displays tickets relevant to the current user.
// Clients see only their own, techs see assigned, and admins see all.
func handleListTickets() {
	claims, err := utils.LoadClaims()
	if err != nil || claims == nil {
		return
	}

	controllers.ListTickets(claims.UserID, claims.Role)
	utils.LogInfo(fmt.Sprintf("[ListTickets] Tickets listed for user %d (%s)", claims.UserID, claims.Role))
}

// handleAssignTicket allows an admin to manually assign a ticket to a specific tech by user ID.
// Requires ticket ID and tech ID as input.
func handleAssignTicket() {
	claims, err := utils.LoadClaims()
	if err != nil || claims == nil {
		return
	}
	if claims.Role != "admin" {
		fmt.Println("[Error] Only admins can assign tickets.")
		utils.LogWarning(fmt.Sprintf("[AssignTicket] Unauthorized attempt by user %d (%s)", claims.UserID, claims.Role))
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Ticket ID to assign: ")
	tidStr, _ := reader.ReadString('\n')
	tidStr = strings.TrimSpace(tidStr)
	tid64, err := strconv.ParseUint(tidStr, 10, 64)
	if err != nil {
		fmt.Println("[Error] Invalid ticket ID.")
		return
	}
	ticketID := uint(tid64)

	fmt.Print("Tech User ID: ")
	techStr, _ := reader.ReadString('\n')
	techStr = strings.TrimSpace(techStr)
	tech64, err := strconv.ParseUint(techStr, 10, 64)
	if err != nil {
		fmt.Println("[Error] Invalid tech ID.")
		return
	}
	techID := uint(tech64)

	controllers.AssignTicket(ticketID, techID)
	utils.LogInfo(fmt.Sprintf("[AssignTicket] Admin %d assigned ticket %d to tech %d", claims.UserID, ticketID, techID))
}

// handleViewTicket lets a user inspect a ticket by ID, respecting their role visibility.
func handleViewTicket() {
	claims, err := utils.LoadClaims()
	if err != nil || claims == nil {
		fmt.Println("[Error] Session expired or invalid. Please log in again.")
		utils.LogWarning("[ViewTicket] Invalid or missing session")
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Ticket ID to view: ")
	idStr, _ := reader.ReadString('\n')
	idStr = strings.TrimSpace(idStr)

	idUint64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		fmt.Println("[Error] Invalid ticket ID.")
		return
	}
	ticketID := uint(idUint64)

	controllers.ViewTicket(ticketID, claims.UserID, claims.Role)
	utils.LogInfo(fmt.Sprintf("[ViewTicket] User %d (%s) viewed ticket %d", claims.UserID, claims.Role, ticketID))

	// Fetch and display comments inline
	var comments []models.Comment
	config.DB.Where("ticket_id = ?", ticketID).Order("created_at asc").Find(&comments)

	if len(comments) > 0 {
		fmt.Println("\nComments:")
		fmt.Println("-------------------------------------------------")
		for _, c := range comments {
			fmt.Printf("%s @ %s\n  %s\n\n", c.AuthorEmail, c.CreatedAt.Format("2006-01-02 15:04"), c.Content)
		}
	}
}

// handleFilterTickets prompts the user to select optional filters and shows matching tickets.
func handleFilterTickets() {
	claims, err := utils.LoadClaims()
	if err != nil || claims == nil {
		fmt.Println("[Error] Session expired or invalid. Please log in again.")
		utils.LogWarning("[FilterTickets] Invalid or missing session")
		return
	}

	priorityOptions := []string{"", "low", "medium", "high", "critical"}
	statusOptions := []string{
		"",
		"initially reported",
		"customer to follow up",
		"support to follow up",
		"working",
		"closed",
	}

	fmt.Println("Leave blank to skip a filter.")

	priority, err := utils.PromptSelect("Filter by Priority", priorityOptions, 0)
	if err != nil {
		fmt.Println("Priority selection cancelled.")
		return
	}

	status, err := utils.PromptSelect("Filter by Status", statusOptions, 0)
	if err != nil {
		fmt.Println("Status selection cancelled.")
		return
	}

	controllers.FilterTickets(claims.UserID, claims.Role, priority, status)
	utils.LogInfo(fmt.Sprintf("[FilterTickets] User %d (%s) filtered tickets", claims.UserID, claims.Role))
}

// handleCommentTicket adds a comment to a ticket
func handleCommentTicket() {
	claims, err := utils.LoadClaims()
	if err != nil || claims == nil {
		fmt.Println("[Error] Session expired or invalid. Please log in again.")
		utils.LogWarning("[CommentTicket] Invalid or missing session")
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Ticket ID to comment on: ")
	idStr, _ := reader.ReadString('\n')
	idStr = strings.TrimSpace(idStr)
	idUint64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		fmt.Println("[Error] Invalid ticket ID.")
		return
	}
	ticketID := uint(idUint64)

	fmt.Print("Enter your comment: ")
	commentText, _ := reader.ReadString('\n')
	commentText = strings.TrimSpace(commentText)
	if commentText == "" {
		fmt.Println("No comment entered.")
		return
	}

	comment := models.Comment{
		TicketID:    ticketID,
		AuthorID:    claims.UserID,
		AuthorEmail: claims.Email,
		Content:     commentText,
		CreatedAt:   time.Now(),
	}

	if err := config.DB.Create(&comment).Error; err != nil {
		fmt.Println("[Error] Failed to save comment.")
		utils.LogError("[CommentTicket] Failed to save comment", err)
		return
	}

	fmt.Println("Comment added.")
	utils.LogInfo(fmt.Sprintf("[CommentTicket] User %d commented on ticket %d", claims.UserID, ticketID))
}

// handleDeleteTicket allows an admin to delete a ticket by its ID.
// It prompts for confirmation before deletion to avoid accidental loss.
func handleDeleteTicket() {
	claims, err := utils.LoadClaims()
	if err != nil || claims == nil {
		return
	}
	if claims.Role != "admin" {
		fmt.Println("[Error] Only admins can delete tickets.")
		utils.LogWarning(fmt.Sprintf("[DeleteTicket] Unauthorized attempt by user %d (%s)", claims.UserID, claims.Role))
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Ticket ID to delete: ")
	idStr, _ := reader.ReadString('\n')
	idStr = strings.TrimSpace(idStr)
	idUint64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		fmt.Println("[Error] Invalid ticket ID.")
		return
	}
	ticketID := uint(idUint64)

	fmt.Printf("Are you sure you want to delete ticket %d? Type 'yes' to confirm: ", ticketID)
	confirm, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(confirm)) != "yes" {
		fmt.Println("Deletion cancelled.")
		utils.LogInfo(fmt.Sprintf("[DeleteTicket] Admin %d cancelled deletion of ticket %d", claims.UserID, ticketID))
		return
	}

	controllers.DeleteTicket(ticketID)
	utils.LogInfo(fmt.Sprintf("[DeleteTicket] Admin %d deleted ticket %d", claims.UserID, ticketID))
}

// handleViewLogs allows admins to view recent events from the log file.
// Logs are printed to the terminal from the logs/ryanforce.log file.
func handleViewLogs() {
	claims, err := utils.LoadClaims()
	if err != nil || claims == nil {
		return
	}
	if claims.Role != "admin" {
		fmt.Println("[Error] Only admins can view logs.")
		utils.LogWarning(fmt.Sprintf("[ViewLogs] Unauthorized access attempt by user %d (%s)", claims.UserID, claims.Role))
		return
	}

	data, err := os.ReadFile("logs/ryanforce.log")
	if err != nil {
		fmt.Println("[Error] Unable to read log file.")
		utils.LogError("[ViewLogs] Could not read ryanforce.log", err)
		return
	}

	fmt.Println("\nSystem Event Log")
	fmt.Println("------------------------")
	fmt.Println(string(data))
	utils.LogInfo(fmt.Sprintf("[ViewLogs] Admin %d viewed logs", claims.UserID))
}

// handleHelp shows available commands based on the current user's role.
// Admins, techs, and clients will only see commands they are allowed to use.
func handleHelp() {
	claims, err := utils.LoadClaims()
	if err != nil || claims == nil {
		fmt.Println("[Error] Session expired or invalid. Please log in again.")
		utils.LogWarning("[Help] Invalid or missing session")
		return
	}

	role := claims.Role

	fmt.Println("\nAvailable Commands:")
	fmt.Println("------------------------")
	fmt.Println("login           (l, auth)          Log in to the system")
	fmt.Println("logout          (lo, exit)         Log out of current session")
	fmt.Println("whoami          (me, status)       Show info about current user")
	fmt.Println("help            (h, ?)             Show this help message")

	if role == "client" {
		fmt.Println("create-ticket   (ct, new)          Create a new support ticket")
		fmt.Println("view-ticket     (vt, show)         View a specific ticket")
		fmt.Println("list-tickets    (lt, list)         List your submitted tickets")
		fmt.Println("filter-tickets  (ft, search)       Filter your tickets by status/priority")
	}
	if role == "tech" {
		fmt.Println("update-ticket   (ut, edit)         Update a ticket assigned to you")
		fmt.Println("view-ticket     (vt, show)         View a specific ticket")
		fmt.Println("list-tickets    (lt, list)         List your assigned tickets")
		fmt.Println("filter-tickets  (ft, search)       Filter assigned tickets")
	}
	if role == "admin" {
		fmt.Println("register        (r, signup)        Register a new user")
		fmt.Println("create-account     -             Create a new customer account")
		fmt.Println("assign-account     -             Assign user to an account by ID")
		fmt.Println("list-accounts      -             List all accounts and their users")
		fmt.Println("delete-account     -             Delete an account by ID")
		fmt.Println("admin-reset-password (arp, admin-reset) Reset a user's password")
		fmt.Println("assign-ticket   (at, assign)       Assign a ticket to a tech")
		fmt.Println("delete-ticket   (dt, remove)       Delete a ticket by ID")
		fmt.Println("clear-db        -                  Dangerously wipe all data")
		fmt.Println("view-logs       (logs, tail)       View system event log")
		fmt.Println("update-ticket   (ut, edit)         Update any ticket")
		fmt.Println("view-ticket     (vt, show)         View a specific ticket")
		fmt.Println("list-tickets    (lt, list)         List all tickets")
		fmt.Println("list-users      (lu)				Lists all users")
		fmt.Println("delete-user 	 (du) 				Delete a user by ID")
		fmt.Println("filter-tickets  (ft, search)       Filter all tickets")
		fmt.Println("report-status   -                  Show number of tickets by status")
		fmt.Println("report-priority  -                 Show number of tickets by priority")
		fmt.Println("report-unassigned -                List all tickets without an assigned tech")
		fmt.Println("report-overdue     -             Show open tickets that have passed SLA deadline")
		fmt.Println("report-all         -             Run full report summary (status, SLA, overdue)")
		fmt.Println("export-tickets     -             Export all tickets to CSV file")
	}
	utils.LogInfo(fmt.Sprintf("[Help] Help viewed by user %d (%s)", claims.UserID, claims.Role))
}

// handleClearDB wipes all users and tickets from the database.
// This is a DANGEROUS admin-only action that prompts for confirmation.
func handleClearDB() {
	claims, err := utils.LoadClaims()
	if err != nil || claims == nil {
		return
	}
	if claims.Role != "admin" {
		fmt.Println("[Error] Only admins can clear the database.")
		utils.LogWarning(fmt.Sprintf("[ClearDB] Unauthorized database wipe attempt by user %d (%s)", claims.UserID, claims.Role))
		return
	}

	fmt.Print("Are you sure you want to delete all users and tickets? Type 'yes' to confirm: ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(input)) != "yes" {
		fmt.Println("Cancelled.")
		utils.LogInfo(fmt.Sprintf("[ClearDB] Admin %d cancelled database wipe", claims.UserID))
		return
	}

	controllers.ClearDatabase(true)
	utils.LogWarning(fmt.Sprintf("[ClearDB] Admin %d wiped database", claims.UserID))
}
