package controllers

import (
	"RyanForce/config"
	"RyanForce/models"
	"RyanForce/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// CreateTicket adds a new ticket to the database via CLI.
// Logs success or failure to the console and log files.
func CreateTicket(title, desc, priority, status string, clientID uint) {
	ticket := models.Ticket{
		Title:       title,
		Description: desc,
		Priority:    priority,
		Status:      status,
		ClientID:    clientID,
	}

	if err := config.DB.Create(&ticket).Error; err != nil {
		utils.LogErrorIP("[TicketCLI] Failed to create ticket", err, "CLI-Local")
		fmt.Println("[Error] Failed to create ticket.")
		return
	}

	fmt.Println("Ticket created successfully.")
	utils.LogInfoIP(fmt.Sprintf("[TicketCLI] Ticket created by user %d — ID: %d", clientID, ticket.ID), "CLI-Local")
}

// ListTickets displays tickets to the CLI based on user role.
// Supports client, technician, and admin views.
func ListTickets(userID uint, role string) {
	var tickets []models.Ticket
	var err error

	switch role {
	case "client":
		err = config.DB.Where("client_id = ?", userID).Find(&tickets).Error
	case "tech":
		err = config.DB.Where("tech_id = ?", userID).Find(&tickets).Error
	case "admin":
		err = config.DB.Find(&tickets).Error
	default:
		utils.LogWarning(fmt.Sprintf("[TicketCLI] Unauthorized list attempt by user %d (role: %s)", userID, role))
		fmt.Println("[Error] Unauthorized access.")
		return
	}

	if err != nil {
		utils.LogError("[TicketCLI] Failed to retrieve tickets", err)
		fmt.Println("[Error] Could not retrieve tickets.")
		return
	}

	if len(tickets) == 0 {
		fmt.Println("No tickets found.")
		return
	}

	for _, t := range tickets {
		PrintTicketSummary(t)
	}

	utils.LogInfo(fmt.Sprintf("[TicketCLI] %d tickets listed for user %d (role: %s)", len(tickets), userID, role))
}

/*
// UpdateTicket updates a ticket's fields via CLI with role enforcement.
// Allows clients, techs, and admins to modify their assigned tickets.
func UpdateTicket(ticketID, userID uint, role, desc, priority, status string) {
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketID).Error; err != nil {
		utils.LogErrorIP("[TicketCLI] Ticket not found", err, "CLI-Local")
		fmt.Println("[Error] Ticket not found.")
		return
	}

	// Role enforcement
	switch role {
	case "client":
		if ticket.ClientID != userID {
			fmt.Println("[Error] You cannot update someone else's ticket.")
			return
		}
	case "tech":
		if ticket.TechID == nil || *ticket.TechID != userID {
			fmt.Println("[Error] This ticket is not assigned to you.")
			return
		}
	}

	// Update fields if provided
	if desc != "" {
		ticket.Description = desc
	}
	if priority != "" {
		ticket.Priority = priority
	}
	if status != "" {
		ticket.Status = status
	}

	// Handle closing tickets
	if status == "closed" && ticket.ClosedAt == nil {
		now := time.Now()
		ticket.ClosedAt = &now
	} else if status != "closed" {
		ticket.ClosedAt = nil
	}

	if err := config.DB.Save(&ticket).Error; err != nil {
		utils.LogErrorIP("[TicketCLI] Failed to update ticket", err, "CLI-Local")
		fmt.Println("[Error] Could not update ticket.")
		return
	}

	fmt.Println("Ticket updated successfully.")
	utils.LogInfoIP(fmt.Sprintf("[TicketCLI] Ticket %d updated by user %d (role: %s)", ticketID, userID, role), "CLI-Local")
}
*/

// AddCommentToTicket creates a new comment for a ticket
func AddCommentToTicket(ticketID uint, body string, authorID uint, authorEmail string, ip string) error {
	comment := models.Comment{
		TicketID:    ticketID,
		AuthorID:    authorID,
		AuthorEmail: authorEmail,
		Content:     body,
		CreatedAt:   time.Now(),
	}
	if err := config.DB.Create(&comment).Error; err != nil {
		utils.LogErrorIP("[Comment] Failed to add comment", err, ip)
		return err
	}

	utils.LogInfoIP(fmt.Sprintf("[Comment] User %d added Comment #%d to Ticket #%d", authorID, comment.ID, ticketID), ip)
	return nil
}

// EditComment updates the content of an existing comment
func EditComment(commentID uint, newContent string, ip string) error {
	var comment models.Comment
	if err := config.DB.First(&comment, commentID).Error; err != nil {
		utils.LogWarningIP(fmt.Sprintf("[Comment] Edit failed — comment %d not found", commentID), ip)
		return err
	}
	comment.Content = newContent

	if err := config.DB.Save(&comment).Error; err != nil {
		utils.LogErrorIP(fmt.Sprintf("[Comment] Failed to edit comment %d", commentID), err, ip)
		return err
	}

	utils.LogInfoIP(fmt.Sprintf("[Comment] Comment %d updated", commentID), ip)
	return nil
}

// DeleteComment removes a comment by its ID
func DeleteComment(commentID uint, ip string) error {
	if err := config.DB.Delete(&models.Comment{}, commentID).Error; err != nil {
		utils.LogErrorIP(fmt.Sprintf("[Comment] Failed to delete comment %d", commentID), err, ip)
		return err
	}

	utils.LogInfoIP(fmt.Sprintf("[Comment] Comment %d deleted", commentID), ip)
	return nil
}

// GetCommentsForTicket retrieves all comments associated with a ticket
func GetCommentsForTicket(ticketID uint) ([]models.Comment, error) {
	var comments []models.Comment
	if err := config.DB.Where("ticket_id = ?", ticketID).Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

// AssignTicket assigns a technician to a ticket via CLI (admin only).
// Logs the result of the assignment action.
func AssignTicket(ticketID, techID uint) {
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketID).Error; err != nil {
		utils.LogWarningIP(fmt.Sprintf("[TicketCLI] Assignment failed — ticket %d not found", ticketID), "CLI-Local")
		fmt.Println("[Error] Ticket not found.")
		return
	}

	ticket.TechID = &techID
	if err := config.DB.Save(&ticket).Error; err != nil {
		utils.LogErrorIP("[TicketCLI] Failed to assign technician", err, "CLI-Local")
		fmt.Println("[Error] Failed to assign technician.")
		return
	}

	fmt.Printf("Ticket %d assigned to technician %d successfully.\n", ticketID, techID)
	utils.LogInfoIP(fmt.Sprintf("[TicketCLI] Ticket %d assigned to technician %d", ticketID, techID), "CLI-Local")
}

// ViewTicket displays full ticket details via CLI with access control.
// Supports client, technician, and admin roles.
func ViewTicket(ticketID, userID uint, role string) {
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketID).Error; err != nil {
		utils.LogWarning(fmt.Sprintf("[TicketCLI] View denied — ticket %d not found", ticketID))
		fmt.Println("[Error] Ticket not found.")
		return
	}

	switch role {
	case "client":
		if ticket.ClientID != userID {
			fmt.Println("[Error] Unauthorized to view this ticket.")
			return
		}
	case "tech":
		if ticket.TechID == nil || *ticket.TechID != userID {
			fmt.Println("[Error] Unauthorized to view this ticket.")
			return
		}
	}

	fmt.Println("\n--- Ticket Details ---")
	fmt.Printf("ID:          %d\n", ticket.ID)
	fmt.Printf("Title:       %s\n", ticket.Title)
	fmt.Printf("Description: %s\n", ticket.Description)
	fmt.Printf("Priority:    %s\n", ticket.Priority)
	fmt.Printf("Status:      %s\n", ticket.Status)
	fmt.Printf("Client ID:   %d\n", ticket.ClientID)
	if ticket.TechID != nil {
		fmt.Printf("Assigned To: %d\n", *ticket.TechID)
	} else {
		fmt.Println("Assigned To: (unassigned)")
	}
	fmt.Println("------------------------")

	utils.LogInfo(fmt.Sprintf("[TicketCLI] Ticket %d viewed by user %d (role: %s)", ticket.ID, userID, role))
}

// DeleteTicket permanently removes a ticket via CLI (admin only).
// Logs the deletion action and provides CLI feedback.
func DeleteTicket(ticketID uint) {
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketID).Error; err != nil {
		utils.LogWarningIP(fmt.Sprintf("[TicketCLI] Deletion failed — ticket %d not found", ticketID), "CLI-Local")
		fmt.Println("[Error] Ticket not found.")
		return
	}

	if err := config.DB.Delete(&ticket).Error; err != nil {
		utils.LogErrorIP("[TicketCLI] Failed to delete ticket", err, "CLI-Local")
		fmt.Println("[Error] Failed to delete ticket.")
		return
	}

	fmt.Printf("Ticket %d deleted successfully.\n", ticketID)
	utils.LogInfoIP(fmt.Sprintf("[TicketCLI] Ticket %d deleted", ticketID), "CLI-Local")
}

// FilterTickets lists tickets based on priority/status filters via CLI.
// Applies role-based access control to filter results.
func FilterTickets(userID uint, role, priority, status string) {
	var tickets []models.Ticket
	query := config.DB.Model(&models.Ticket{})

	switch role {
	case "client":
		query = query.Where("client_id = ?", userID)
	case "tech":
		query = query.Where("tech_id = ?", userID)
	case "admin":
		// Admin can see all tickets
	default:
		fmt.Println("[Error] Unauthorized access.")
		utils.LogWarning(fmt.Sprintf("[TicketCLI] Unauthorized filter attempt by user %d (role: %s)", userID, role))
		return
	}

	if priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&tickets).Error; err != nil {
		utils.LogError("[TicketCLI] Failed to filter tickets", err)
		fmt.Println("[Error] Could not retrieve tickets.")
		return
	}

	if len(tickets) == 0 {
		fmt.Println("No tickets match the filter criteria.")
		return
	}

	for _, t := range tickets {
		PrintTicketSummary(t)
	}

	utils.LogInfo(fmt.Sprintf("[TicketCLI] %d tickets listed for user %d (role: %s)", len(tickets), userID, role))
}

// PrintTicketSummary prints a brief summary of a ticket in CLI format.
// Shows ticket ID, title, priority, status, and assigned technician.
func PrintTicketSummary(t models.Ticket) {
	assigned := "(unassigned)"
	if t.TechID != nil {
		assigned = fmt.Sprintf("%d", *t.TechID)
	}

	fmt.Printf("ID: %d | Title: %s | Priority: %s | Status: %s | Assigned To: %s\n",
		t.ID, t.Title, t.Priority, t.Status, assigned)
}

// SaveNewTicket saves a new ticket to the database.
// Returns an error if creation fails.
func SaveNewTicket(ticket *models.Ticket) error {
	return config.DB.Create(ticket).Error
}

// ModifyTicket updates an existing ticket in the database.
// Returns an error if update fails.
func ModifyTicket(ticket *models.Ticket) error {
	return config.DB.Save(ticket).Error
}

// RemoveTicket deletes a ticket by its string ID.
// Returns an error if deletion fails.
func RemoveTicket(id string) error {
	return config.DB.Delete(&models.Ticket{}, id).Error
}

// CreateTicketAPI handles creating a new ticket via POST (JSON input).
// Returns the created ticket or an error response.
func CreateTicketAPI(c *gin.Context) {
	var ticket models.Ticket
	if err := c.ShouldBindJSON(&ticket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := SaveNewTicket(&ticket); err != nil {
		utils.LogErrorIP("[TicketAPI] Failed to create ticket", err, c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save ticket"})
		return
	}

	utils.LogInfoIP(fmt.Sprintf("[TicketAPI] Ticket created successfully — ID: %d", ticket.ID), c.ClientIP())
	c.JSON(http.StatusCreated, ticket)
}

// UpdateTicketAPI updates an existing ticket via PUT/PATCH (JSON input).
// Returns the updated ticket or an error response.
func UpdateTicketAPI(c *gin.Context) {
	id := c.Param("id")
	var ticket models.Ticket

	if err := config.DB.First(&ticket, id).Error; err != nil {
		utils.LogWarningIP(fmt.Sprintf("[TicketAPI] Update failed — ticket %s not found", id), c.ClientIP())
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	if err := c.ShouldBindJSON(&ticket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := ModifyTicket(&ticket); err != nil {
		utils.LogErrorIP("[TicketAPI] Failed to update ticket", err, c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update ticket"})
		return
	}

	utils.LogInfoIP(fmt.Sprintf("[TicketAPI] Ticket %d updated successfully", ticket.ID), c.ClientIP())
	c.JSON(http.StatusOK, ticket)
}

// DeleteTicketAPI deletes a ticket by ID via REST API.
// Returns a success message or an error response.
func DeleteTicketAPI(c *gin.Context) {
	id := c.Param("id")

	if err := RemoveTicket(id); err != nil {
		utils.LogErrorIP(fmt.Sprintf("[TicketAPI] Failed to delete ticket %s", id), err, c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete ticket"})
		return
	}

	utils.LogInfoIP(fmt.Sprintf("[TicketAPI] Ticket %s deleted successfully", id), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Ticket deleted successfully"})
}

// ViewTicketAPI returns detailed ticket information via REST API.
// Enforces role-based access control before displaying ticket data.
func ViewTicketAPI(c *gin.Context) {
	user := c.MustGet("user").(*utils.Claims)
	ticketID := c.Param("id")

	var ticket models.Ticket
	if err := config.DB.
		Preload("AssignedTech").
		Preload("Client").
		First(&ticket, ticketID).Error; err != nil {
		utils.LogWarning(fmt.Sprintf("[TicketAPI] View failed — ticket %s not found", ticketID))
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	// Role-based access control
	switch user.Role {
	case "client":
		if ticket.ClientID != user.UserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to view this ticket"})
			return
		}
	case "tech":
		if ticket.TechID == nil || *ticket.TechID != user.UserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Ticket not assigned to you"})
			return
		}
	case "admin":
		// Admin allowed
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Unknown role"})
		return
	}

	// Prepare response payload
	assignedInfo := interface{}("(unassigned)")
	if ticket.AssignedTech != nil {
		assignedInfo = gin.H{
			"id":    ticket.AssignedTech.ID,
			"email": ticket.AssignedTech.Email,
		}
	}

	clientInfo := gin.H{
		"id":    ticket.Client.ID,
		"email": ticket.Client.Email,
	}

	utils.LogInfo(fmt.Sprintf("[TicketAPI] Ticket %d viewed by user %d (role: %s)", ticket.ID, user.UserID, user.Role))
	c.JSON(http.StatusOK, gin.H{
		"id":          ticket.ID,
		"title":       ticket.Title,
		"description": ticket.Description,
		"priority":    ticket.Priority,
		"status":      ticket.Status,
		"client":      clientInfo,
		"assigned_to": assignedInfo,
		"created_at":  ticket.CreatedAt,
		"updated_at":  ticket.UpdatedAt,
		"closed_at":   ticket.ClosedAt,
	})
}

// ListTicketsAPI lists tickets viewable by the authenticated user via REST API.
// Supports client, technician, and admin role-based listing.
func ListTicketsAPI(c *gin.Context) {
	user := c.MustGet("user").(*utils.Claims)
	var tickets []models.Ticket

	query := config.DB.Model(&models.Ticket{})

	switch user.Role {
	case "client":
		query = query.Where("client_id = ?", user.UserID)
	case "tech":
		query = query.Where("tech_id = ?", user.UserID)
	case "admin":
		// Admin sees all tickets
	default:
		utils.LogWarning(fmt.Sprintf("[TicketAPI] Unauthorized list attempt by user %d (role: %s)", user.UserID, user.Role))
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized role"})
		return
	}

	if err := query.Find(&tickets).Error; err != nil {
		utils.LogError("[TicketAPI] Failed to list tickets", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tickets"})
		return
	}

	utils.LogInfo(fmt.Sprintf("[TicketAPI] %d tickets listed for user %d (role: %s)", len(tickets), user.UserID, user.Role))
	c.JSON(http.StatusOK, tickets)
}

// FilterTicketsAPI lists tickets with optional priority and status filters via REST API.
// Applies role-based access control to results.
func FilterTicketsAPI(c *gin.Context) {
	user := c.MustGet("user").(*utils.Claims)
	var tickets []models.Ticket

	priority := c.Query("priority")
	status := c.Query("status")

	query := config.DB.Model(&models.Ticket{})

	switch user.Role {
	case "client":
		query = query.Where("client_id = ?", user.UserID)
	case "tech":
		query = query.Where("tech_id = ?", user.UserID)
	case "admin":
		// Admin sees all
	default:
		utils.LogWarning(fmt.Sprintf("[TicketAPI] Unauthorized filter attempt by user %d (role: %s)", user.UserID, user.Role))
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized role"})
		return
	}

	if priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&tickets).Error; err != nil {
		utils.LogError("[TicketAPI] Failed to filter tickets", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to filter tickets"})
		return
	}

	utils.LogInfo(fmt.Sprintf("[TicketAPI] %d tickets filtered for user %d (role: %s)", len(tickets), user.UserID, user.Role))
	c.JSON(http.StatusOK, tickets)
}

// AssignTicketAPI assigns a technician to a ticket via REST API (admin only).
// Requires techID field in JSON body.
func AssignTicketAPI(c *gin.Context) {
	user := c.MustGet("user").(*utils.Claims)
	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can assign technicians"})
		return
	}

	ticketID := c.Param("id")
	var body struct {
		TechID uint `json:"tech_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	ticket.TechID = &body.TechID
	if err := config.DB.Save(&ticket).Error; err != nil {
		utils.LogErrorIP("[TicketAPI] Failed to assign technician", err, c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign technician"})
		return
	}

	utils.LogInfoIP(fmt.Sprintf("[TicketAPI] Ticket %s assigned to tech %d by admin %d", ticketID, body.TechID, user.UserID), c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Technician assigned successfully"})
}

/*
// AddComment handles creating a comment from a form submission in the WebUI.
// Validates input and associates the comment with the ticket and user.
func AddComment(c *gin.Context) {
	idStr := c.Param("id")
	ticketID64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.SetCookie("flash", "Invalid ticket ID", 3, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/dashboard")
		return
	}
	ticketID := uint(ticketID64)

	content := c.PostForm("content")
	if content == "" {
		c.SetCookie("flash", "Comment content cannot be empty", 3, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/tickets/"+idStr)
		return
	}

	tokenStr, err := utils.LoadSession()
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	claims, err := utils.ParseJWT(tokenStr)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	comment := models.Comment{
		TicketID:    ticketID,
		AuthorID:    claims.UserID,
		AuthorEmail: claims.Email,
		Content:     content,
		CreatedAt:   time.Now(),
	}

	if err := config.DB.Create(&comment).Error; err != nil {
		utils.LogError("[Comment] Failed to save", err)
		c.SetCookie("flash", "Failed to save comment", 3, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/tickets/"+idStr)
		return
	}

	utils.LogInfo(fmt.Sprintf("[Comment] User %d added comment to ticket %d", claims.UserID, ticketID))
	c.SetCookie("flash", "Comment posted successfully", 3, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/tickets/"+idStr)
}
*/

/*
// ShowEditCommentForm renders the form for editing an existing comment via WebUI.
// Enforces user permissions before displaying the edit form.
func ShowEditCommentForm(c *gin.Context) {
	commentID := c.Param("commentID")
	var comment models.Comment

	if err := config.DB.First(&comment, commentID).Error; err != nil {
		utils.LogWarning(fmt.Sprintf("[CommentWebUI] Edit form load failed — comment %s not found", commentID))
		c.HTML(http.StatusNotFound, "404.html", nil)
		return
	}

	tokenStr, err := utils.LoadSession()
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	claims, err := utils.ParseJWT(tokenStr)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	if comment.AuthorID != claims.UserID && claims.Role != "admin" {
		c.SetCookie("flash", "You are not allowed to edit this comment", 3, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/tickets/"+strconv.Itoa(int(comment.TicketID)))
		return
	}

	c.HTML(http.StatusOK, "edit_comment.html", gin.H{
		"TicketID": comment.TicketID,
		"Comment":  comment,
	})
}
*/

/*
// UpdateComment updates the content of an existing comment via WebUI.
// Validates user permissions before applying the update.
func UpdateComment(c *gin.Context) {
	commentID := c.Param("commentID")
	newContent := c.PostForm("content")

	var comment models.Comment
	if err := config.DB.First(&comment, commentID).Error; err != nil {
		utils.LogWarning(fmt.Sprintf("[CommentWebUI] Update failed — comment %s not found", commentID))
		c.HTML(http.StatusNotFound, "404.html", nil)
		return
	}

	if newContent == "" {
		c.SetCookie("flash", "Comment cannot be empty", 3, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/tickets/"+strconv.Itoa(int(comment.TicketID)))
		return
	}

	tokenStr, err := utils.LoadSession()
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	claims, err := utils.ParseJWT(tokenStr)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	if comment.AuthorID != claims.UserID && claims.Role != "admin" {
		c.SetCookie("flash", "You are not allowed to update this comment", 3, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/tickets/"+strconv.Itoa(int(comment.TicketID)))
		return
	}

	comment.Content = newContent
	if err := config.DB.Save(&comment).Error; err != nil {
		utils.LogError("[CommentWebUI] Failed to update comment", err)
		c.SetCookie("flash", "Failed to update comment", 3, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/tickets/"+strconv.Itoa(int(comment.TicketID)))
		return
	}

	utils.LogInfo(fmt.Sprintf("[CommentWebUI] Comment %d updated by user %d", comment.ID, claims.UserID))
	c.SetCookie("flash", "Comment updated successfully", 3, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/tickets/"+strconv.Itoa(int(comment.TicketID)))
}
*/

/*
// DeleteComment removes a comment from the database
func DeleteComment(c *gin.Context) {
	commentID := c.Param("commentID")

	var comment models.Comment
	if err := config.DB.First(&comment, commentID).Error; err != nil {
		c.String(http.StatusNotFound, "Comment not found")
		return
	}

	tokenStr, err := utils.LoadSession()
	if err != nil {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return
	}
	claims, err := utils.ParseJWT(tokenStr)
	if err != nil {
		c.String(http.StatusUnauthorized, "Invalid session")
		return
	}

	if comment.AuthorID != claims.UserID && claims.Role != "admin" {
		c.String(http.StatusForbidden, "You are not allowed to delete this comment")
		return
	}

	if err := config.DB.Delete(&comment).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to delete comment")
		return
	}

	utils.LogInfo(fmt.Sprintf("[Comment] Comment %d deleted by user %d", comment.ID, claims.UserID))
	c.SetCookie("flash", "Comment deleted successfully", 3, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/tickets/"+strconv.Itoa(int(comment.TicketID)))
}
*/

// FindMatchingTechs displays a list of best-matched technicians for a ticket.
// Renders the admin assignment page with tech skill match scores.
func FindMatchingTechs(c *gin.Context) {
	ticketID := c.Param("id")
	var ticket models.Ticket

	if err := config.DB.First(&ticket, ticketID).Error; err != nil {
		utils.LogWarning(fmt.Sprintf("[TicketAdmin] Ticket %s not found for assignment", ticketID))
		c.HTML(http.StatusNotFound, "404.html", nil)
		return
	}

	neededSkills, err := utils.ParseSkills(ticket.SkillsNeeded)
	if err != nil {
		utils.LogError("[TicketAdmin] Failed to parse required skills", err)
		c.String(http.StatusInternalServerError, "Invalid skills needed format")
		return
	}

	var techs []models.User
	if err := config.DB.Where("role = ?", "tech").Find(&techs).Error; err != nil {
		utils.LogError("[TicketAdmin] Failed to fetch technicians", err)
		c.String(http.StatusInternalServerError, "Error fetching technicians")
		return
	}

	type ScoredTech struct {
		Tech  models.User
		Score int
	}
	var matches []ScoredTech

	for _, tech := range techs {
		techSkills, err := utils.ParseSkills(tech.Skills)
		if err != nil {
			continue // skip bad records
		}
		score := utils.MatchScore(techSkills, neededSkills)
		if score > 0 {
			matches = append(matches, ScoredTech{Tech: tech, Score: score})
		}
	}

	// Sort matches by descending score
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	utils.LogInfo(fmt.Sprintf("[TicketAdmin] Found %d matching technicians for ticket %s", len(matches), ticketID))
	c.HTML(http.StatusOK, "admin_assign.html", gin.H{
		"ticket":            ticket,
		"needed":            neededSkills,
		"matches":           matches,
		"totalSkillsNeeded": len(neededSkills),
	})
}

// AssignTechToTicket assigns a technician to a ticket based on admin selection.
// Handles POST requests from the assignment interface.
func AssignTechToTicket(c *gin.Context) {
	ticketID := c.Param("id")
	techID := c.Param("tech_id")
	var ticket models.Ticket

	if err := config.DB.First(&ticket, ticketID).Error; err != nil {
		utils.LogWarning(fmt.Sprintf("[TicketAdmin] Assign failed — ticket %s not found", ticketID))
		c.HTML(http.StatusNotFound, "404.html", nil)
		return
	}

	parsedTechID, err := strconv.ParseUint(techID, 10, 64)
	if err != nil {
		c.SetCookie("flash", "Invalid technician ID", 3, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/admin/unassigned-tickets")
		return
	}

	techIDUint := uint(parsedTechID)
	ticket.TechID = &techIDUint

	if err := config.DB.Save(&ticket).Error; err != nil {
		utils.LogError("[TicketAdmin] Failed to assign technician", err)
		c.SetCookie("flash", "Failed to assign technician", 3, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/admin/unassigned-tickets")
		return
	}

	utils.LogInfo(fmt.Sprintf("[TicketAdmin] Ticket %s assigned to tech %d", ticketID, techIDUint))
	c.SetCookie("flash", "Technician assigned successfully", 3, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/admin/unassigned-tickets")
}

// UnassignTechFromTicket handles POST /admin/tickets/:id/unassign
// Removes the assigned technician from a ticket.
func UnassignTechFromTicket(c *gin.Context) {
	ticketID := c.Param("id")

	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketID).Error; err != nil {
		c.String(http.StatusNotFound, "Ticket not found")
		return
	}

	ticket.TechID = nil

	if err := config.DB.Save(&ticket).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to unassign technician")
		return
	}

	c.Redirect(http.StatusFound, "/admin/assigned-tickets?success=Ticket+"+ticketID+"+unassigned+successfully")
}
