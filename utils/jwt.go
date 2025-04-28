package utils

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"strings"
	"time"
)

// jwtKey is used to sign and verify JWT tokens.
var jwtKey = []byte("your-secret-key")

// Claims defines the structure stored inside the JWT.
// Includes the user's ID, email, role, and standard JWT expiration metadata.
type Claims struct {
	UserID uint
	Email  string
	Role   string
	jwt.RegisteredClaims
}

// GenerateJWT creates a signed JWT token using the user's ID, email, and role.
// Tokens are valid for 24 hours from the time of creation.
func GenerateJWT(userID uint, email, role string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// ParseJWT validates the given JWT token string and extracts the custom Claims.
// Returns an error if the token is invalid, expired, or improperly signed.
func ParseJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	tokenStr = strings.TrimSpace(tokenStr)

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		if err != nil {
			fmt.Println("ParseJWT error:", err)
			return nil, errors.New("invalid or expired token")
		}
	}

	return claims, nil
}
