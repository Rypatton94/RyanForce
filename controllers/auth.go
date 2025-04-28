package controllers

import (
	"RyanForce/config"
	"RyanForce/models"
	"RyanForce/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Register creates a new user in the system with a hashed password and a specified role.
// This is typically used by an admin to add a new user manually.
func Register(email, password, role string) {
	hash, err := utils.HashPassword(password)
	if err != nil {
		utils.LogError("[Register] Failed to hash password", err)
		fmt.Println("[Error] Failed to hash password.")
		return
	}

	user := models.User{Email: email, PasswordHash: hash, Role: role}
	if err := config.DB.Create(&user).Error; err != nil {
		utils.LogError("[Register] Failed to create user", err)
		fmt.Println("[Error] Failed to create user.")
		return
	}

	fmt.Println("User registered.")
	utils.LogInfo(fmt.Sprintf("[Register] User created: %s (%s)", user.Email, user.Role))
}

// Authenticate attempts to log in a user by checking their email and password.
// If successful, it returns a JWT token that can be used for future authenticated actions.
func Authenticate(email, password string) (string, error) {
	var user models.User

	if config.DB == nil {
		return "", errors.New("database not connected")
	}

	result := config.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		utils.LogWarning(fmt.Sprintf("[Auth] Login failed for unknown email: %s", email))
		return "", errors.New("user not found")
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		utils.LogWarning(fmt.Sprintf("[Auth] Invalid password for user: %s", email))
		return "", errors.New("invalid password")
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		utils.LogError("[Auth] Failed to generate token", err)
		return "", errors.New("failed to generate token")
	}

	utils.LogInfo(fmt.Sprintf("[Auth] User logged in: %s (%s)", user.Email, user.Role))
	return token, nil
}

// Login is used by the WebUI and API to validate credentials and return a JWT.
// It logs all login attempts for auditing purposes.
func Login(email, password string) (string, error) {
	var user models.User
	cleanedEmail := strings.TrimSpace(email)

	if err := config.DB.Where("email = ?", cleanedEmail).First(&user).Error; err != nil {
		utils.LogWarning("[Login] Failed login: user not found — " + cleanedEmail)
		return "", fmt.Errorf("invalid credentials")
	}

	if user.IsLocked {
		utils.LogWarning("[Login] Account locked: " + cleanedEmail)
		return "", fmt.Errorf("account is locked due to repeated failed attempts")
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		user.FailedAttempts++
		if user.FailedAttempts >= 5 {
			user.IsLocked = true
			utils.LogWarning("[Login] Account locked due to too many failed attempts — " + cleanedEmail)
		}
		config.DB.Save(&user)
		utils.LogWarning("[Login] Failed login: wrong password — " + cleanedEmail)
		return "", fmt.Errorf("invalid credentials")
	}

	// Reset failed attempts on success
	user.FailedAttempts = 0
	user.IsLocked = false
	now := time.Now()
	user.LastLogin = &now
	config.DB.Save(&user)

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		utils.LogError("[Login] Failed to generate JWT", err)
		return "", fmt.Errorf("token generation failed")
	}

	utils.LogInfo("[Login] Successful login — " + user.Email)
	return token, nil
}

// ResetPassword allows a user to change their password if they provide the correct current password.
func ResetPassword(email, oldPassword, newPassword string) error {
	if !isValidPassword(newPassword) {
		return fmt.Errorf("password must be 8–32 characters and include a capital letter, number, and special character")
	}

	var user models.User
	if err := config.DB.First(&user, "email = ?", email).Error; err != nil {
		utils.LogWarning(fmt.Sprintf("[Reset] User not found: %s", email))
		return fmt.Errorf("user not found")
	}

	if !utils.CheckPasswordHash(oldPassword, user.PasswordHash) {
		utils.LogWarning(fmt.Sprintf("[Reset] Incorrect old password for %s", email))
		return fmt.Errorf("old password is incorrect")
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		utils.LogError("[Reset] Failed to hash new password", err)
		return fmt.Errorf("failed to hash new password")
	}

	user.PasswordHash = hash
	if err := config.DB.Save(&user).Error; err != nil {
		utils.LogError("[Reset] Failed to update password", err)
		return fmt.Errorf("failed to update password")
	}

	utils.LogInfo(fmt.Sprintf("[Reset] User changed password: %s", email))
	return nil
}

// AdminResetPassword allows an administrator to reset another user's password.
func AdminResetPassword(adminID uint, userEmail, newPassword string) error {
	if !isValidPassword(newPassword) {
		return fmt.Errorf("password must be 8–32 characters and include a capital letter, number, and special character")
	}

	var user models.User
	if err := config.DB.First(&user, "email = ?", userEmail).Error; err != nil {
		utils.LogWarning(fmt.Sprintf("[AdminReset] Target user not found: %s", userEmail))
		return fmt.Errorf("user not found")
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		utils.LogError("[AdminReset] Failed to hash new password", err)
		return fmt.Errorf("failed to hash new password")
	}

	user.PasswordHash = hash
	if err := config.DB.Save(&user).Error; err != nil {
		utils.LogError("[AdminReset] Failed to update password for %s", err)
		return fmt.Errorf("failed to update password")
	}

	utils.LogInfo(fmt.Sprintf("[AdminReset] Admin %d reset password for %s", adminID, userEmail))
	return nil
}

// LoginAPI handles JSON-based login and returns a JWT on success.
// This is used by CLI clients or API.
func LoginAPI(c *gin.Context) {
	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogWarning("[API] Login attempt with malformed JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	ip := c.ClientIP()
	utils.LogInfo("[API] Login attempt from IP: " + ip + " — " + req.Email)

	token, err := Login(req.Email, req.Password)
	if err != nil {
		utils.LogWarning("[API] Login failed for " + req.Email + " from IP: " + ip)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	utils.LogInfo("[API] Login successful for " + req.Email + " from IP: " + ip)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// RegisterAPI creates a new user entry with hashed password.
func RegisterAPI(c *gin.Context) {
	type RegisterRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.User{
		Email:        req.Email,
		PasswordHash: hash,
		Role:         req.Role,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

// IsValidPassword checks for minimum password complexity
func isValidPassword(pw string) bool {
	if len(pw) < 8 || len(pw) > 32 {
		return false
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(pw)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(pw)
	hasSpecial := regexp.MustCompile(`[!@#%^&*()_+\-=\[\]{};':"\\|,.<>/?]`).MatchString(pw)

	return hasUpper && hasDigit && hasSpecial
}
