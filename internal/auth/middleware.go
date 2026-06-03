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
		userID, err := ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Store user ID in context for subsequent handlers
		c.Set("userID", userID)
		c.Next()
	}
}

// RequireAdmin verifies if the authenticated user has the 'admin' role in the database
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Usuário não autenticado"})
			c.Abort()
			return
		}

		var role string
		// Usando o pool de conexões do pacote database para fazer a query rápida
		err := database.GetDB().QueryRow(context.Background(), "SELECT role FROM users WHERE id = $1", userID).Scan(&role)
		if err != nil || role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Acesso negado. Apenas administradores."})
			c.Abort()
			return
		}

		c.Next()
	}
}
