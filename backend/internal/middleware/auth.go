package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/vikukumar/Pushpaka/internal/services"
)

const UserIDKey = "userID"

// JWT validates the Bearer token and sets userID in context.
// For WebSocket connections that cannot set headers, the token may
// alternatively be passed as the `token` query parameter.
func JWT(authSvc *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var raw string

		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format, expected: Bearer <token>"})
				return
			}
			raw = parts[1]
		} else if q := c.Query("token"); q != "" {
			// Fallback: ?token= for WebSocket upgrades (browsers cannot set custom headers)
			raw = q
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			return
		}

		userID, err := authSvc.ValidateToken(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(UserIDKey, userID)
		c.Next()
	}
}

// GetUserID retrieves the authenticated user ID from context
func GetUserID(c *gin.Context) string {
	userID, _ := c.Get(UserIDKey)
	return userID.(string)
}
