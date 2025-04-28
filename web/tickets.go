package web

import (
	"RyanForce/config"
	"RyanForce/models"
	"RyanForce/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// DisplayComment defines a simplified comment for rendering.
type DisplayComment struct {
	Email     string
	Content   string
	Timestamp string
}

// ShowCreateTicketForm renders the ticket submission form.
func ShowCreateTicketForm(c *gin.Context) {
	c.HTML(http.StatusOK, "create_ticket.html", nil)
}

// ListClientTickets shows tickets submitted by the currently logged-in client.
func ListClientTickets(c *gin.Context) {
	claims := c.MustGet("user").(*utils.Claims)

	var tickets []models.Ticket
	if err := config.DB.Where("client_id = ?", claims.UserID).Find(&tickets).Error; err != nil {
		utils.LogError("[WebUI] Ticket listing failed", err)
		c.String(http.StatusInternalServerError, "Could not retrieve tickets")
		return
	}

	c.HTML(http.StatusOK, "client_tickets.html", gin.H{"tickets": tickets})
}

// ListTechTickets shows tickets assigned to the currently logged-in technician.
func ListTechTickets(c *gin.Context) {
	claims := c.MustGet("user").(*utils.Claims)

	var tickets []models.Ticket
	if err := config.DB.Where("tech_id = ?", claims.UserID).Find(&tickets).Error; err != nil {
		c.String(http.StatusInternalServerError, "Could not fetch assigned tickets")
		return
	}

	c.HTML(http.StatusOK, "tech_tickets.html", gin.H{"tickets": tickets})
}

// ViewTicketPage displays a full ticket detail view for the WebUI.
func ViewTicketPage(c *gin.Context) {
	id := c.Param("id")
	var ticket models.Ticket
	var comments []models.Comment

	if err := config.DB.First(&ticket, id).Error; err != nil {
		utils.LogWarning(fmt.Sprintf("[TicketWebUI] Ticket %s not found", id))
		c.HTML(http.StatusNotFound, "404.html", nil)
		return
	}

	config.DB.Where("ticket_id = ?", id).Order("created_at asc").Find(&comments)

	claims := c.MustGet("user").(*utils.Claims)

	if claims.Role == "client" && ticket.ClientID != claims.UserID {
		c.HTML(http.StatusForbidden, "403.html", nil)
		return
	}
	if claims.Role == "tech" && (ticket.TechID == nil || *ticket.TechID != claims.UserID) {
		c.HTML(http.StatusForbidden, "403.html", nil)
		return
	}

	// NEW: Parse SkillsNeeded JSON into []string
	var skills []string
	if ticket.SkillsNeeded != "" {
		if err := json.Unmarshal([]byte(ticket.SkillsNeeded), &skills); err != nil {
			utils.LogWarning(fmt.Sprintf("[TicketWebUI] Failed to parse skillsNeeded for ticket %d", ticket.ID))
			skills = []string{} // fallback safely
		}
	}

	type DisplayCommentFull struct {
		models.Comment
		CanEdit  bool
		TicketID uint
	}

	var displayComments []DisplayCommentFull
	for _, com := range comments {
		canEdit := com.AuthorID == claims.UserID || claims.Role == "admin"
		displayComments = append(displayComments, DisplayCommentFull{
			Comment:  com,
			CanEdit:  canEdit,
			TicketID: ticket.ID,
		})
	}

	flashMsg, _ := c.Cookie("flash")
	c.SetCookie("flash", "", -1, "/", "", false, true)

	// UPDATED: Pass parsed skills slice into template
	c.HTML(http.StatusOK, "ticket_view.html", gin.H{
		"Ticket":   ticket,
		"Skills":   skills, // <-- HERE
		"Comments": displayComments,
		"Flash":    flashMsg,
		"UserID":   claims.UserID,
		"UserRole": claims.Role,
	})
}

// Comment Management (Add, Edit, Delete)

// AddComment creates a comment from a form submission.
func AddComment(c *gin.Context) {
	idStr := c.Param("id")
	ticketID64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/dashboard")
		return
	}
	ticketID := uint(ticketID64)
	content := c.PostForm("content")
	if content == "" {
		c.Redirect(http.StatusSeeOther, "/tickets/"+idStr)
		return
	}
	claims := c.MustGet("user").(*utils.Claims)

	comment := models.Comment{
		TicketID:    ticketID,
		AuthorID:    claims.UserID,
		AuthorEmail: claims.Email,
		Content:     content,
		CreatedAt:   time.Now(),
	}

	_ = config.DB.Create(&comment)
	c.Redirect(http.StatusSeeOther, "/tickets/"+idStr)
}

// UpdateComment edits an existing comment.
func UpdateComment(c *gin.Context) {
	commentID := c.Param("commentID")
	newContent := c.PostForm("content")
	var comment models.Comment

	if err := config.DB.First(&comment, commentID).Error; err != nil {
		c.HTML(http.StatusNotFound, "404.html", nil)
		return
	}
	if newContent == "" {
		c.Redirect(http.StatusSeeOther, "/tickets/"+strconv.Itoa(int(comment.TicketID)))
		return
	}
	claims := c.MustGet("user").(*utils.Claims)

	if comment.AuthorID != claims.UserID && claims.Role != "admin" {
		c.Redirect(http.StatusSeeOther, "/tickets/"+strconv.Itoa(int(comment.TicketID)))
		return
	}

	comment.Content = newContent
	_ = config.DB.Save(&comment)
	c.Redirect(http.StatusSeeOther, "/tickets/"+strconv.Itoa(int(comment.TicketID)))
}

// DeleteComment removes a comment.
func DeleteComment(c *gin.Context) {
	commentID := c.Param("commentID")
	var comment models.Comment

	if err := config.DB.First(&comment, commentID).Error; err != nil {
		c.String(http.StatusNotFound, "Comment not found")
		return
	}
	claims := c.MustGet("user").(*utils.Claims)

	if comment.AuthorID != claims.UserID && claims.Role != "admin" {
		c.String(http.StatusForbidden, "Forbidden")
		return
	}

	_ = config.DB.Delete(&comment)
	c.Redirect(http.StatusSeeOther, "/tickets/"+strconv.Itoa(int(comment.TicketID)))
}

// ShowEditCommentForm renders the form for editing an existing comment via WebUI.
func ShowEditCommentForm(c *gin.Context) {
	commentID := c.Param("commentID")
	var comment models.Comment

	if err := config.DB.First(&comment, commentID).Error; err != nil {
		utils.LogWarning(fmt.Sprintf("[CommentWebUI] Edit form load failed — comment %s not found", commentID))
		c.HTML(http.StatusNotFound, "404.html", nil)
		return
	}

	claims := c.MustGet("user").(*utils.Claims)

	if comment.AuthorID != claims.UserID && claims.Role != "admin" {
		c.Redirect(http.StatusSeeOther, "/tickets/"+strconv.Itoa(int(comment.TicketID)))
		return
	}

	c.HTML(http.StatusOK, "edit_comment.html", gin.H{
		"TicketID": comment.TicketID,
		"Comment":  comment,
	})
}

// ShowUpdateTicketForm renders a form for technicians to update a ticket.
func ShowUpdateTicketForm(c *gin.Context) {
	id := c.Param("id")
	var ticket models.Ticket
	if err := config.DB.First(&ticket, id).Error; err != nil {
		c.String(http.StatusNotFound, "Ticket not found")
		return
	}

	var rawComments []models.Comment
	config.DB.Where("ticket_id = ?", ticket.ID).Order("created_at asc").Find(&rawComments)

	var displayComments []DisplayComment
	for _, com := range rawComments {
		email := com.AuthorEmail
		if email == "" {
			email = "Unknown"
		}
		displayComments = append(displayComments, DisplayComment{
			Email:     email,
			Content:   com.Content,
			Timestamp: com.CreatedAt.Format("2006-01-02 15:04"),
		})
	}

	c.HTML(http.StatusOK, "update_ticket.html", gin.H{
		"ticket":   ticket,
		"comments": displayComments,
	})
}

// HandleUpdateTicket processes ticket update submissions.
func HandleUpdateTicket(c *gin.Context) {
	ticketID := c.Param("id")
	var ticket models.Ticket
	if err := config.DB.First(&ticket, ticketID).Error; err != nil {
		c.String(http.StatusNotFound, "Ticket not found")
		return
	}

	newStatus := c.PostForm("status")
	if newStatus != "" && newStatus != ticket.Status {
		ticket.Status = newStatus
	}

	rawSkills := c.PostFormArray("SkillsNeeded")
	if len(rawSkills) > 0 {
		skills := []string{}
		for _, skill := range rawSkills {
			trimmed := strings.TrimSpace(skill)
			if trimmed != "" {
				skills = append(skills, trimmed)
			}
		}
		if len(skills) > 0 {
			skillsJSON, err := json.Marshal(skills)
			if err != nil {
				c.String(http.StatusInternalServerError, "Invalid skills input")
				return
			}
			ticket.SkillsNeeded = string(skillsJSON)
		}
	}

	if err := config.DB.Save(&ticket).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to update ticket")
		return
	}

	commentText := strings.TrimSpace(c.PostForm("comment"))
	if commentText != "" {
		claims := c.MustGet("user").(*utils.Claims)

		comment := models.Comment{
			TicketID:    ticket.ID,
			AuthorID:    claims.UserID,
			AuthorEmail: claims.Email,
			Content:     commentText,
			CreatedAt:   time.Now(),
		}

		_ = config.DB.Create(&comment)
	}

	utils.LogInfo(fmt.Sprintf("[UpdateTicket] Ticket %s updated", ticketID))
	c.Redirect(http.StatusFound, "/tickets/"+ticketID)
}

// HandleCreateTicket creates a new ticket from a form submission.
func HandleCreateTicket(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	priority := c.PostForm("priority")
	status := "initially reported"

	claims := c.MustGet("user").(*utils.Claims)

	rawSkills := c.PostFormArray("skillsNeeded")
	skills := []string{}
	for _, skill := range rawSkills {
		trimmed := strings.TrimSpace(skill)
		if trimmed != "" {
			skills = append(skills, trimmed)
		}
	}

	var skillsJSON string
	if len(skills) > 0 {
		skillsBytes, err := json.Marshal(skills)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid skills input")
			return
		}
		skillsJSON = string(skillsBytes)
	}

	ticket := models.Ticket{
		Title:        title,
		Description:  description,
		Priority:     priority,
		Status:       status,
		ClientID:     claims.UserID,
		SkillsNeeded: skillsJSON,
	}

	if err := config.DB.Create(&ticket).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to create ticket")
		return
	}

	utils.LogInfo(fmt.Sprintf("[CreateTicket] Ticket #%d created by user %d", ticket.ID, claims.UserID))
	c.Redirect(http.StatusFound, "/tickets/mine")
}
