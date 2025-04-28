package web

import (
	"RyanForce/config"
	"RyanForce/models"
	"RyanForce/utils"
	"encoding/csv"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// ListClients handles GET /admin/clients
// Displays all clients to the admin.
func ListClients(c *gin.Context) {
	accountFilter := c.Query("account")

	var clients []models.User
	db := config.DB.Preload("Account").Where("role = ?", "client")

	var selectedAccountID uint
	if accountFilter != "" {
		if id, err := strconv.ParseUint(accountFilter, 10, 64); err == nil {
			selectedAccountID = uint(id)
			db = db.Where("account_id = ?", selectedAccountID)
		}
	}

	if err := db.Find(&clients).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to retrieve clients")
		return
	}

	var accounts []models.Account
	config.DB.Find(&accounts)

	c.HTML(http.StatusOK, "admin_clients_list.html", gin.H{
		"clients":           clients,
		"accounts":          accounts,
		"selectedAccountID": selectedAccountID,
	})
}

// NewClientForm handles GET /admin/clients/new
// Creates a form for adding a new client.
func NewClientForm(c *gin.Context) {
	var accounts []models.Account
	if err := config.DB.Find(&accounts).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "admin_clients_list.html", gin.H{
			"error": "Failed to load accounts",
		})
		return
	}

	c.HTML(http.StatusOK, "admin_client_new.html", gin.H{
		"accounts": accounts,
	})
}

// CreateClient handles POST /admin/clients
// Creates a new client in the system.
func CreateClient(c *gin.Context) {
	var user models.User

	user.Email = c.PostForm("Email")
	user.Name = c.PostForm("Name")
	user.Role = "client"

	if accountIDStr := c.PostForm("AccountID"); accountIDStr != "" {
		if accountIDUint, err := strconv.ParseUint(accountIDStr, 10, 64); err == nil {
			accountID := uint(accountIDUint)
			user.AccountID = &accountID
		}
	}

	password := c.PostForm("Password")
	if !utils.IsValidPassword(password) {
		c.String(http.StatusBadRequest, "Password does not meet security requirements")
		return
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		c.String(http.StatusInternalServerError, "Password hashing failed")
		return
	}
	user.PasswordHash = hash

	if err := config.DB.Create(&user).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error creating client")
		return
	}

	c.Redirect(http.StatusFound, "/admin/clients")
}

// ShowClient handles GET /admin/clients/:id
func ShowClient(c *gin.Context) {
	id := c.Param("id")
	var client models.User
	if err := config.DB.First(&client, id).Error; err != nil {
		c.String(http.StatusNotFound, "Client not found")
		return
	}
	c.HTML(http.StatusOK, "admin_client_show.html", gin.H{"client": client})
}

// UpdateClient handles POST form submission to update client info.
// Updates email, name, and account fields.
func UpdateClient(c *gin.Context) {
	clientID := c.Param("id")
	var client models.User

	if err := config.DB.First(&client, clientID).Error; err != nil {
		utils.LogWarning(fmt.Sprintf("[AdminClient] Update failed — client %s not found", clientID))
		c.HTML(http.StatusNotFound, "404.html", nil)
		return
	}

	// Update fields
	email := c.PostForm("Email")
	name := c.PostForm("Name")
	accountID := c.PostForm("AccountID")

	if email != "" {
		client.Email = email
	}
	if name != "" {
		client.Name = name
	}
	if accountID != "" {
		id, err := strconv.ParseUint(accountID, 10, 64)
		if err == nil {
			parsedID := uint(id)
			client.AccountID = &parsedID
		}
	} else {
		client.AccountID = nil
	}

	if err := config.DB.Save(&client).Error; err != nil {
		utils.LogError("[AdminClient] Failed to update client", err)
		c.SetCookie("flash", "Failed to update client", 3, "/", "", false, true)
		c.Redirect(http.StatusSeeOther, "/admin/clients")
		return
	}

	utils.LogInfo(fmt.Sprintf("[AdminClient] Client %d updated successfully", client.ID))
	c.SetCookie("flash", "Client updated successfully", 3, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/admin/clients")
}

// DeleteClient handles POST /admin/clients/:id/delete
// Deletes the specified client from the system.
func DeleteClient(c *gin.Context) {
	id := c.Param("id")
	var client models.User
	if err := config.DB.Delete(&client, id).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error deleting client")
		return
	}
	c.Redirect(http.StatusFound, "/admin/clients")
}

func ExportClientsCSV(c *gin.Context) {
	accountFilter := c.Query("account")

	var clients []models.User
	db := config.DB.Preload("Account").Where("role = ?", "client")

	if accountFilter != "" {
		if id, err := strconv.ParseUint(accountFilter, 10, 64); err == nil {
			db = db.Where("account_id = ?", uint(id))
		}
	}

	if err := db.Find(&clients).Error; err != nil {
		c.String(http.StatusInternalServerError, "Could not export clients")
		return
	}

	// Configure CSV headers
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment;filename=clients.csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	writer.Write([]string{"ID", "Email", "Account"})

	for _, u := range clients {
		accountName := ""
		if u.AccountID != nil && u.Account.Name != "" {
			accountName = u.Account.Name
		}
		writer.Write([]string{
			strconv.Itoa(int(u.ID)),
			u.Email,
			accountName,
		})
	}
}
