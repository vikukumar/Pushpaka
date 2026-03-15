package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/yourusername/pushpaka/internal/services"
)

const UserIDKey = "userID"

// JWT validates the Bearer token and sets userID in context
func JWT(authSvc *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format, expected: Bearer <token>"})
			return
		}

		userID, err := authSvc.ValidateToken(parts[1])
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
