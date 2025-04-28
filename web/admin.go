package web

import (
	"RyanForce/config"
	"RyanForce/models"
	"RyanForce/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

func ShowUnassignedTickets(c *gin.Context) {
	var tickets []models.Ticket
	if err := config.DB.Where("tech_id IS NULL").Find(&tickets).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to load unassigned tickets")
		return
	}

	success := c.Query("success")
	c.HTML(http.StatusOK, "admin_unassigned.html", gin.H{
		"tickets": tickets,
		"success": success,
	})
}

func ShowEditClientForm(c *gin.Context) {
	clientID := c.Param("id")

	var client models.User
	if err := config.DB.Preload("Account").First(&client, clientID).Error; err != nil {
		c.HTML(http.StatusNotFound, "404.html", nil)
		return
	}

	var accounts []models.Account
	if err := config.DB.Find(&accounts).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "admin_clients_list.html", gin.H{
			"error": "Failed to load accounts",
		})
		return
	}

	accountID := uint(0)
	if client.AccountID != nil {
		accountID = *client.AccountID
	}

	c.HTML(http.StatusOK, "admin_client_edit.html", gin.H{
		"client":    client,
		"accounts":  accounts,
		"accountID": accountID,
	})
}

func CreateAccount(c *gin.Context) {
	account := models.Account{
		Name:    c.PostForm("Name"),
		Domain:  c.PostForm("Domain"),
		Address: c.PostForm("Address"),
		Notes:   c.PostForm("Notes"),
	}

	if err := config.DB.Create(&account).Error; err != nil {
		utils.LogError("[Admin] Failed to create account", err)
		c.String(http.StatusInternalServerError, "Could not create account")
		return
	}

	c.Redirect(http.StatusFound, "/admin/accounts")
}

func EditAccountForm(c *gin.Context) {
	id := c.Param("id")
	var account models.Account
	if err := config.DB.First(&account, id).Error; err != nil {
		utils.LogError("[Admin] Failed to load account", err)
		c.String(http.StatusNotFound, "Account not found")
		return
	}

	c.HTML(http.StatusOK, "admin_account_edit.html", gin.H{"account": account})
}

func UpdateAccount(c *gin.Context) {
	id := c.Param("id")
	var account models.Account
	if err := config.DB.First(&account, id).Error; err != nil {
		utils.LogError("[Admin] Account not found", err)
		c.String(http.StatusNotFound, "Account not found")
		return
	}

	account.Name = c.PostForm("Name")
	account.Domain = c.PostForm("Domain")
	account.Address = c.PostForm("Address")
	account.Notes = c.PostForm("Notes")

	if err := config.DB.Save(&account).Error; err != nil {
		utils.LogError("[Admin] Failed to update account", err)
		c.String(http.StatusInternalServerError, "Failed to update account")
		return
	}

	c.Redirect(http.StatusFound, "/admin/accounts")
}

func ListAccounts(c *gin.Context) {
	var accounts []models.Account
	err := config.DB.Preload("Users").Find(&accounts).Error
	if err != nil {
		utils.LogError("[Admin] Failed to load accounts", err)
		c.HTML(http.StatusInternalServerError, "admin_dashboard.html", gin.H{
			"error": "Failed to load accounts",
		})
		return
	}

	c.HTML(http.StatusOK, "admin_accounts.html", gin.H{
		"accounts": accounts,
	})
}

func DeleteAccount(c *gin.Context) {
	accountID := c.Param("id")

	// Check if any users are still assigned
	var count int64
	if err := config.DB.Model(&models.User{}).Where("account_id = ?", accountID).Count(&count).Error; err != nil {
		utils.LogError("[Admin] Failed account-user check", err)
		c.String(http.StatusInternalServerError, "Failed to check account")
		return
	}

	if count > 0 {
		c.String(http.StatusForbidden, "Cannot delete account: %d user(s) are still assigned.", count)
		return
	}

	if err := config.DB.Delete(&models.Account{}, accountID).Error; err != nil {
		utils.LogError("[Admin] Failed to delete account", err)
		c.String(http.StatusInternalServerError, "Failed to delete account")
		return
	}

	c.Redirect(http.StatusFound, "/admin/accounts")
}

// ShowAssignedTickets displays all tickets that have a technician assigned.
func ShowAssignedTickets(c *gin.Context) {
	var tickets []models.Ticket
	if err := config.DB.Preload("AssignedTech").Where("tech_id IS NOT NULL").Find(&tickets).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to load assigned tickets")
		return
	}

	success := c.Query("success")
	c.HTML(http.StatusOK, "admin_assigned.html", gin.H{
		"tickets": tickets,
		"success": success,
	})
}
