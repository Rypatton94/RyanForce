package web

import (
	"RyanForce/config"
	"RyanForce/controllers"
	"RyanForce/models"
	"RyanForce/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// ShowLoginPage renders the login HTML form.
func ShowLoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{"error": ""})
}

// HandleWebLogin processes the WebUI login form and sets the JWT token in a cookie.
func HandleWebLogin(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	ip := c.ClientIP()

	utils.LogInfo("[WebUI] Login attempt from IP: " + ip + " — " + email)

	token, err := controllers.Login(email, password)
	if err != nil {
		utils.LogWarning("[WebUI] Login failed for " + email + " from IP: " + ip)
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{"error": "Invalid credentials"})
		return
	}

	utils.LogInfo("[WebUI] Login successful for " + email + " from IP: " + ip)
	c.SetCookie("token", token, 3600, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/dashboard")
}

// ShowDashboard displays the role-specific dashboard with last login timestamp.
func ShowDashboard(c *gin.Context) {
	tokenStr, err := c.Cookie("token")
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	claims, err := utils.ParseJWT(tokenStr)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	var user models.User
	if err := config.DB.First(&user, claims.UserID).Error; err != nil {
		c.String(http.StatusInternalServerError, "Could not load user data")
		return
	}

	data := gin.H{
		"user":      user.Email,
		"lastLogin": user.LastLogin,
	}

	switch user.Role {
	case "admin":
		c.HTML(http.StatusOK, "admin_dashboard.html", data)
	case "tech":
		c.HTML(http.StatusOK, "tech_dashboard.html", data)
	case "client":
		c.HTML(http.StatusOK, "client_dashboard.html", data)
	default:
		c.String(http.StatusForbidden, "Invalid role")
	}
}

// HandleLogout clears the token cookie and logs the event.
func HandleLogout(c *gin.Context) {
	token, err := c.Cookie("token")
	if err == nil {
		if claims, err := utils.ParseJWT(token); err == nil {
			utils.LogInfo("[Logout] User logged out: " + claims.Email)
		}
	}

	c.SetCookie("token", "", -1, "/", "localhost", false, true)
	c.Redirect(http.StatusFound, "/login")
}

// ShowResetForm renders the password reset form.
func ShowResetForm(c *gin.Context) {
	c.HTML(http.StatusOK, "reset_password.html", gin.H{"error": ""})
}

// HandleResetPassword processes the reset form and updates the user's password.
func HandleResetPassword(c *gin.Context) {
	email := c.PostForm("email")
	oldPassword := c.PostForm("old_password")
	newPassword := c.PostForm("new_password")

	if !utils.IsValidPassword(newPassword) {
		utils.LogWarning("[Reset] Weak password submitted for " + email)
		c.HTML(http.StatusBadRequest, "reset_password.html", gin.H{
			"error": "Password must be 8–32 characters and include a capital letter, number, and special character.",
		})
		return
	}

	err := controllers.ResetPassword(email, oldPassword, newPassword)
	if err != nil {
		utils.LogWarning("[Reset] Password reset failed for " + email + ": " + err.Error())
		c.HTML(http.StatusBadRequest, "reset_password.html", gin.H{"error": err.Error()})
		return
	}

	utils.LogInfo("[Reset] Password successfully updated for " + email)
	c.HTML(http.StatusOK, "reset_password.html", gin.H{"success": "Password updated successfully!"})
}

// ShowAdminResetForm displays the admin password reset page
func ShowAdminResetForm(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_reset_password.html", gin.H{"error": ""})
}

// HandleAdminResetPassword allows an admin to reset a user password without the old password
func HandleAdminResetPassword(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	claims, err := utils.ParseJWT(token)
	if err != nil || claims.Role != "admin" {
		c.String(http.StatusForbidden, "Unauthorized")
		return
	}

	email := c.PostForm("email")
	newPassword := c.PostForm("new_password")

	if !utils.IsValidPassword(newPassword) {
		utils.LogWarning(fmt.Sprintf("[AdminReset] Weak password submitted for %s", email))
		c.HTML(http.StatusBadRequest, "admin_reset_password.html", gin.H{
			"error": "Password must be 8–32 characters and include a capital letter, number, and special character.",
		})
		return
	}

	err = controllers.AdminResetPassword(claims.UserID, email, newPassword)
	if err != nil {
		utils.LogWarning("[AdminReset] Failed password reset for " + email)
		c.HTML(http.StatusBadRequest, "admin_reset_password.html", gin.H{"error": err.Error()})
		return
	}

	utils.LogInfo(fmt.Sprintf("[AdminReset] Admin %d reset password for %s", claims.UserID, email))
	c.HTML(http.StatusOK, "admin_reset_password.html", gin.H{"success": "Password reset successful for " + email})
}

// ShowUnlockForm displays the unlock user form
func ShowUnlockForm(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_unlock_user.html", gin.H{"error": "", "success": ""})
}

// HandleUnlockUser allows admin to unlock a locked-out user account
func HandleUnlockUser(c *gin.Context) {
	token, err := c.Cookie("token")
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	claims, err := utils.ParseJWT(token)
	if err != nil || claims.Role != "admin" {
		c.String(http.StatusForbidden, "Unauthorized")
		return
	}

	email := c.PostForm("email")
	var user models.User
	if err := config.DB.First(&user, "email = ?", email).Error; err != nil {
		c.HTML(http.StatusBadRequest, "admin_unlock_user.html", gin.H{"error": "User not found"})
		return
	}

	if !user.IsLocked {
		c.HTML(http.StatusOK, "admin_unlock_user.html", gin.H{"success": "Account is already unlocked."})
		return
	}

	user.IsLocked = false
	user.FailedAttempts = 0
	if err := config.DB.Save(&user).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "admin_unlock_user.html", gin.H{"error": "Failed to update user."})
		return
	}

	ip := c.ClientIP()
	utils.LogInfo(fmt.Sprintf("[AdminUnlock] Admin %d unlocked user %s from IP %s", claims.UserID, email, ip))
	c.HTML(http.StatusOK, "admin_unlock_user.html", gin.H{"success": "Account successfully unlocked."})
}
