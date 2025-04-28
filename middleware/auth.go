package middleware

import (
	"RyanForce/utils"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware ensures the incoming request has a valid token.
// If not, it halts the request and returns a 401 error.
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			utils.LogWarning("[JWTAuth] Authorization header missing")
			redirectOrJSON(c, "Authorization header missing")
			return
		}

		tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
		utils.LogInfo(fmt.Sprintf("[JWTAuth] Token received: %s", tokenStr))

		claims, err := utils.ParseJWT(tokenStr)
		if err != nil {
			utils.LogError("[JWTAuth] Invalid or expired token", err)
			redirectOrJSON(c, "Invalid or expired token")
			return
		}

		c.Set("user", claims)
		c.Next()
	}
}

// redirectOrJSON decides how to handle errors based on client type
func redirectOrJSON(c *gin.Context, message string) {
	accept := c.GetHeader("Accept")
	if strings.Contains(accept, "text/html") {
		c.Redirect(http.StatusFound, "/login")
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": message})
	}
	c.Abort()
}
