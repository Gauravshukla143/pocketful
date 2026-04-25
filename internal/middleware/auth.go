package middleware

import (
	"net/http"
	"strings"

	"pocketful/internal/config"
	"pocketful/internal/utils"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyUserID = "userID"
	ContextKeyEmail  = "email"
	ContextKeyRole   = "role"
)

// AuthRequired is a Gin middleware that validates the JWT token in the Authorization header.
// It injects user claims (userID, email, role) into the Gin context for downstream handlers.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be: Bearer <token>"})
			return
		}

		claims, err := utils.ParseToken(parts[1], config.AppConfig.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

// AdminRequired is a middleware that ensures the authenticated user has the ADMIN role.
// Must be used after AuthRequired.
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		if role.(string) != "ADMIN" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}
		c.Next()
	}
}
