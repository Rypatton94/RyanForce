package controllers

import (
	"RyanForce/config"
	"RyanForce/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GetUsers returns a list of all users in the system.
func GetUsers(c *gin.Context) {
	var users []models.User
	config.DB.Find(&users)
	c.JSON(http.StatusOK, users)
}

// DeleteUserAPI allows an admin to delete a user by ID.
func DeleteUserAPI(c *gin.Context) {
	id := c.Param("id")
	uid, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := config.DB.Delete(&models.User{}, uid).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}
