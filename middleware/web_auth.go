package middleware

import (
	"RyanForce/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

// WebAuthMiddleware authenticates WebUI users by validating the token cookie.
// Injects the user's claims into the context for easy access by handlers.
func WebAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("token")
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		claims, err := utils.ParseJWT(token)
		if err != nil {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Set user claims into context
		c.Set("user", claims)
		c.Next()
	}
}
