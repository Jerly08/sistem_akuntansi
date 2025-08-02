package main

import (
	"log"
	"app-sistem-akuntansi/config"
	"app-sistem-akuntansi/database"
	"app-sistem-akuntansi/routes"
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db := database.ConnectDB()
	
	// Auto migrate models
	database.AutoMigrate(db)
	
	// Seed database with initial data
	database.SeedData(db)

	// Initialize Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Setup routes
	routes.SetupRoutes(r, db)

	// Start server
	port := cfg.ServerPort
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
