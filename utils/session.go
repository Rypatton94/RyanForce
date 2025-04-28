package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// secretKey is used for AES encryption/decryption of session tokens.
// Must be exactly 32 bytes to support AES-256 encryption.
var secretKey = []byte("32-byte-supersecret-key!!!!!!!!!")

// sessionFile is the local file where the encrypted session token is saved.
// Stored in the system's temporary directory.
var sessionFile = filepath.Join(os.TempDir(), ".ryanforce_session")

// ErrSessionExpired is returned when a session file is missing, corrupted, or expired.
var ErrSessionExpired = errors.New("session expired")

// SaveSession encrypts and saves the JWT token to disk.
// Used to persist user login sessions across CLI restarts.
func SaveSession(token string) error {
	encrypted, err := encrypt(token)
	if err != nil {
		LogError("[Session] Failed to encrypt token", err)
		return err
	}
	if err := os.WriteFile(sessionFile, []byte(encrypted), 0600); err != nil {
		LogError("[Session] Failed to write session file", err)
		return err
	}
	LogInfo("[Session] Token saved successfully.")
	return nil
}

// LoadSession reads and decrypts the session file and validates the token.
// Returns the decrypted token if valid, otherwise clears session and returns an error.
func LoadSession() (string, error) {
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		LogWarning("[Session] No session file found.")
		return "", ErrSessionExpired
	}

	decrypted, err := decrypt(string(data))
	if err != nil {
		LogWarning("[Session] Failed to decrypt token, clearing session.")
		err = ClearSession()
		return "", ErrSessionExpired
	}

	_, err = ParseJWT(decrypted)
	if err != nil {
		LogWarning("[Session] Invalid or expired JWT, clearing session.")
		ClearSession()
		return "", ErrSessionExpired
	}

	LogInfo("[Session] Token loaded and validated.")
	return decrypted, nil
}

// ClearSession deletes the stored session file from disk.
// Used to log out the current user.
func ClearSession() error {
	err := os.Remove(sessionFile)
	if err != nil && !os.IsNotExist(err) {
		LogError("[Session] Failed to delete session file", err)
		return err
	}
	LogInfo("[Session] Session cleared.")
	return nil
}

// LoadClaims loads and validates the session token and returns JWT claims.
// If session is missing or expired, it prints a friendly message and returns nil.
func LoadClaims() (*Claims, error) {
	tokenStr, err := LoadSession()
	if err != nil {
		fmt.Println("[Error] Session expired or missing. Please log in again.")
		return nil, err
	}

	claims, err := ParseJWT(tokenStr)
	if err != nil {
		fmt.Println("[Error] Invalid session. Please log in again.")
		return nil, err
	}

	return claims, nil
}

// encrypt takes a plaintext string and returns an AES-encrypted base64-encoded string.
func encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt takes a base64-encoded AES-encrypted string and returns the original plaintext.
func decrypt(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", errors.New("ciphertext too short")
	}
	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
