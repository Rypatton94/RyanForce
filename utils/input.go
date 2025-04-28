package utils

import (
	"github.com/manifoldco/promptui"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"unicode"
)

// PromptSelect displays a CLI selection menu with a label and options list.
// Returns the option selected by the user. Useful for roles, status, or priority selection.
func PromptSelect(label string, items []string, defaultIndex int) (string, error) {
	prompt := promptui.Select{
		Label:     label,
		Items:     items,
		CursorPos: defaultIndex,
		Size:      5,
	}
	_, result, err := prompt.Run()
	return result, err
}

// IndexOf returns the index of a target string in a slice (case-insensitive).
// If not found, it returns 0. Helps sync user input with option lists.
func IndexOf(target string, options []string) int {
	for i, opt := range options {
		if strings.EqualFold(opt, target) {
			return i
		}
	}
	return 0
}

// IsValidPassword checks if a password meets strength requirements:
// 8–32 characters, 1 uppercase, 1 digit, 1 special character
func IsValidPassword(pw string) bool {
	if len(pw) < 8 || len(pw) > 32 {
		return false
	}

	hasUpper := false
	hasDigit := false
	hasSpecial := false

	for _, r := range pw {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	return hasUpper && hasDigit && hasSpecial
}

// HashPassword generates a secure bcrypt hash of the provided plaintext password.
// This hash is what's stored in the database, never the raw password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a plaintext password against a bcrypt hash.
// Returns true if the password matches the hash, false otherwise.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
