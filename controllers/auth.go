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
		utils.LogErrorIP("[Register] Failed to hash password", err, "CLI-Local")
		fmt.Println("[Error] Failed to hash password.")
		return
	}

	user := models.User{Email: email, PasswordHash: hash, Role: role}
	if err := config.DB.Create(&user).Error; err != nil {
		utils.LogErrorIP("[Register] Failed to create user", err, "CLI-Local")
		fmt.Println("[Error] Failed to create user.")
		return
	}

	fmt.Println("User registered.")
	utils.LogInfoIP(fmt.Sprintf("[Register] User created: %s (%s)", user.Email, user.Role), "CLI-Local")
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
		utils.LogWarningIP(fmt.Sprintf("[Auth] Login failed for unknown email: %s", email), "CLI-Local")
		return "", errors.New("user not found")
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		utils.LogWarningIP(fmt.Sprintf("[Auth] Invalid password for user: %s", email), "CLI-Local")
		return "", errors.New("invalid password")
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		utils.LogErrorIP("[Auth] Failed to generate token", err, "CLI-Local")
		return "", errors.New("failed to generate token")
	}

	utils.LogInfoIP(fmt.Sprintf("[Auth] User logged in: %s (%s)", user.Email, user.Role), "CLI-Local")
	return token, nil
}

// Login is used by the WebUI and API to validate credentials and return a JWT.
// It logs all login attempts for auditing purposes.
func Login(email, password, ip string) (string, error) {
	var user models.User
	cleanedEmail := strings.TrimSpace(email)

	if err := config.DB.Where("email = ?", cleanedEmail).First(&user).Error; err != nil {
		utils.LogWarningIP("[Login] Failed login: user not found — "+cleanedEmail, ip)
		return "", fmt.Errorf("invalid credentials")
	}

	if user.IsLocked {
		utils.LogWarningIP("[Login] Account locked: "+cleanedEmail, ip)
		return "", fmt.Errorf("account is locked due to repeated failed attempts")
	}

	if !utils.CheckPasswordHash(password, user.PasswordHash) {
		user.FailedAttempts++
		if user.FailedAttempts >= 5 {
			user.IsLocked = true
			utils.LogWarning("[Login] Account locked due to too many failed attempts — " + cleanedEmail)
		}
		config.DB.Save(&user)
		utils.LogWarningIP("[Login] Failed login: wrong password — "+cleanedEmail, ip)
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

	utils.LogInfoIP("[Login] Successful login — "+user.Email, ip)
	return token, nil
}

// ResetPassword allows a user to change their password if they provide the correct current password.
func ResetPassword(email, oldPassword, newPassword string) error {
	if !isValidPassword(newPassword) {
		return fmt.Errorf("password must be 8–32 characters and include a capital letter, number, and special character")
	}

	var user models.User
	if err := config.DB.First(&user, "email = ?", email).Error; err != nil {
		utils.LogWarningIP(fmt.Sprintf("[Reset] User not found: %s", email), "CLI-Local")
		return fmt.Errorf("user not found")
	}

	if !utils.CheckPasswordHash(oldPassword, user.PasswordHash) {
		utils.LogWarningIP(fmt.Sprintf("[Reset] Incorrect old password for %s", email), "CLI-Local")
		return fmt.Errorf("old password is incorrect")
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		utils.LogErrorIP("[Reset] Failed to hash new password", err, "CLI-Local")
		return fmt.Errorf("failed to hash new password")
	}

	user.PasswordHash = hash
	if err := config.DB.Save(&user).Error; err != nil {
		utils.LogErrorIP("[Reset] Failed to update password", err, "CLI-Local")
		return fmt.Errorf("failed to update password")
	}

	utils.LogInfoIP(fmt.Sprintf("[Reset] User changed password: %s", email), "CLI-Local")
	return nil
}

// AdminResetPassword allows an administrator to reset another user's password.
func AdminResetPassword(adminID uint, userEmail, newPassword string) error {
	if !isValidPassword(newPassword) {
		return fmt.Errorf("password must be 8–32 characters and include a capital letter, number, and special character")
	}

	var user models.User
	if err := config.DB.First(&user, "email = ?", userEmail).Error; err != nil {
		utils.LogWarningIP(fmt.Sprintf("[AdminReset] Target user not found: %s", userEmail), "CLI-Local")
		return fmt.Errorf("user not found")
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		utils.LogErrorIP("[AdminReset] Failed to hash new password", err, "CLI-Local")
		return fmt.Errorf("failed to hash new password")
	}

	user.PasswordHash = hash
	if err := config.DB.Save(&user).Error; err != nil {
		utils.LogErrorIP(fmt.Sprintf("[AdminReset] Failed to update password for %s", userEmail), err, "CLI-Local")
		return fmt.Errorf("failed to update password")
	}

	utils.LogInfoIP(fmt.Sprintf("[AdminReset] Admin %d reset password for %s", adminID, userEmail), "CLI-Local")
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
	utils.LogInfoIP("[API] Login attempt — "+req.Email, ip)

	token, err := Login(req.Email, req.Password, ip)
	if err != nil {
		utils.LogWarningIP("[API] Login failed — "+req.Email, ip)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	utils.LogInfoIP("[API] Login successful — "+req.Email, ip)
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
