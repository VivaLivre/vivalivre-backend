package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gabrieljose2004/vivalivre-backend/internal/auth"
	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/gabrieljose2004/vivalivre-backend/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize Database
	database.GetDB()
	if err := database.EnsureRatingsSchema(); err != nil {
		log.Fatalf("Failed to ensure ratings schema: %v", err)
	}
	defer database.CloseDB()

	// Setup Router
	r := gin.Default()

	// CORS Middleware (Basic)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Public Routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", handlers.Register)
		authGroup.POST("/login", handlers.Login)
	}

	// Admin Routes (Protegido por JWT + Validação de Role 'admin')
	admin := r.Group("/api/admin")
	admin.Use(auth.AuthMiddleware(), auth.RequireAdmin())
	{
		admin.GET("/bathrooms/pending", handlers.GetPendingBathrooms)
		admin.PATCH("/bathrooms/:id/status", handlers.UpdateBathroomStatus)
	}

	// Protected Routes
	api := r.Group("/api")
	api.Use(auth.AuthMiddleware())
	{
		api.GET("/users/me", handlers.GetMe)
		api.GET("/bathrooms/nearby", handlers.GetNearbyBathrooms)
		api.POST("/bathrooms/request", handlers.RequestBathroom)
		api.GET("/health/entries", handlers.GetHealthEntries)
		api.POST("/health/entries", handlers.CreateHealthEntry)
		api.DELETE("/health/entries/:id", handlers.DeleteHealthEntry)

		// Ratings routes
		api.POST("/bathrooms/:bathroom_id/reviews", handlers.CreateReview)
		api.GET("/bathrooms/:bathroom_id/reviews", handlers.ListReviews)
		api.GET("/bathrooms/:bathroom_id/rating-stats", handlers.GetRatingStats)
		api.GET("/reviews/:review_id", handlers.GetReview)
		api.PUT("/reviews/:review_id", handlers.UpdateReview)
		api.DELETE("/reviews/:review_id", handlers.DeleteReview)
		api.POST("/reviews/:review_id/helpful", handlers.VoteHelpful)
	}

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
