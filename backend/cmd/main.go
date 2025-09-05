package main

import (
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/routes"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set Gin mode based on configuration
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	db := database.ConnectDB()
	
	// Auto migrate models
	database.AutoMigrate(db)
	
	// Migrate permissions table
	if err := database.MigratePermissions(db); err != nil {
		log.Printf("Error migrating permissions: %v", err)
	}
	
	// Seed database with initial data
	database.SeedData(db)
	
	// Run startup tasks including fix account header status
	startupService := services.NewStartupService(db)
	startupService.RunStartupTasks()

	// Initialize Gin router without default middleware
	r := gin.New()

	// Add recovery middleware
	r.Use(gin.Recovery())

	// Add custom logger middleware only in development
	if cfg.Environment != "production" {
		r.Use(gin.Logger())
	}

	// Configure trusted proxies for security
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Setup routes
	routes.SetupRoutes(r, db, startupService)

	// Start server
	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
