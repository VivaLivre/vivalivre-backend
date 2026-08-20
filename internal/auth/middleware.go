package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the JWT token in the Authorization header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be in format 'Bearer <token>'"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		userID, role, err := ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Store user ID and Role in context for subsequent handlers
		c.Set("userID", userID)
		c.Set("userRole", role)
		c.Next()
	}
}

// RequireAdmin verifies if the authenticated user has the 'admin' role in the JWT token
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado. Apenas administradores."})
			c.Abort()
			return
		}

		c.Next()
	}
}
