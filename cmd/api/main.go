package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gabrieljose2004/vivalivre-backend/internal/auth"
	"github.com/gabrieljose2004/vivalivre-backend/internal/database"
	"github.com/gabrieljose2004/vivalivre-backend/internal/handlers"
	adminHandlers "github.com/gabrieljose2004/vivalivre-backend/internal/handlers/admin"
	"github.com/gabrieljose2004/vivalivre-backend/internal/repositories"
	"github.com/gabrieljose2004/vivalivre-backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize Database
	db := database.GetDB()
	if err := database.EnsureRatingsSchema(); err != nil {
		log.Fatalf("Failed to ensure ratings schema: %v", err)
	}
	defer database.CloseDB()

	// Setup Dependencies
	adminRepo := repositories.NewAdminRepository(db)
	adminService := services.NewAdminService(adminRepo)
	adminHandler := adminHandlers.NewAdminHandler(adminService)

	adminBathroomRepo := repositories.NewAdminBathroomRepository(db)
	adminBathroomService := services.NewAdminBathroomService(adminBathroomRepo)
	adminBathroomHandler := adminHandlers.NewAdminBathroomHandler(adminBathroomService)

	adminCrowdsourceRepo := repositories.NewAdminCrowdsourceRepository(db)
	adminCrowdsourceService := services.NewAdminCrowdsourceService(adminCrowdsourceRepo)
	adminCrowdsourceHandler := adminHandlers.NewAdminCrowdsourceHandler(adminCrowdsourceService)

	adminUserRepo := repositories.NewAdminUserRepository(db)
	adminUserService := services.NewAdminUserService(adminUserRepo)
	adminUserHandler := adminHandlers.NewAdminUserHandler(adminUserService)

	// Setup Router with explicit middlewares for production control
	r := gin.New()
	r.Use(gin.Recovery()) // Recover from panics
	r.Use(gin.Logger())   // Request logging

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

	r.POST("/api/auth/google", handlers.GoogleLogin)

	// Admin Routes (Protegido por JWT + Validação de Role 'admin')
	admin := r.Group("/api/admin")
	admin.Use(auth.AuthMiddleware(), auth.RequireAdmin())
	{
		admin.GET("/dashboard/overview", adminHandler.GetDashboardOverview)
		admin.GET("/bathrooms/pending", adminHandler.GetPendingBathrooms)
		admin.PATCH("/bathrooms/:id/status", adminHandler.UpdateBathroomStatus)

		// Full CRUD for Bathrooms
		admin.GET("/bathrooms", adminBathroomHandler.GetAdminBathrooms)
		admin.POST("/bathrooms", adminBathroomHandler.CreateAdminBathroom)
		admin.PATCH("/bathrooms/:id", adminBathroomHandler.UpdateAdminBathroom)
		admin.DELETE("/bathrooms/:id", adminBathroomHandler.DeleteAdminBathroom)

		// Crowdsource Admin Routes
		admin.GET("/reports", adminCrowdsourceHandler.GetAllReports)
		admin.PATCH("/reports/:id/status", adminCrowdsourceHandler.UpdateReportStatus)
		admin.GET("/suggestions", adminCrowdsourceHandler.GetAllSuggestions)
		admin.PATCH("/suggestions/:id/status", adminCrowdsourceHandler.UpdateSuggestionStatus)

		// Users Admin Routes
		admin.GET("/users", adminUserHandler.GetAdminUsers)
		admin.PATCH("/users/:id/status", adminUserHandler.UpdateAdminUserStatus)
	}

	// Protected Routes
	api := r.Group("/api")
	api.Use(auth.AuthMiddleware())
	{
		api.GET("/users/me", handlers.GetMe)
		api.PUT("/users/profile", handlers.UpdateProfile)
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

		// Crowdsourcing routes
		api.POST("/bathrooms/:bathroom_id/report", handlers.CreateBathroomReport)
		api.POST("/bathrooms/:bathroom_id/suggest", handlers.CreateBathroomSuggestion)
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
