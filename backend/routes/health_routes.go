package routes

import (
	"app-sistem-akuntansi/handlers"
	"app-sistem-akuntansi/middleware"
	"github.com/gin-gonic/gin"
)

// SetupHealthRoutes sets up health check routes
func SetupHealthRoutes(router *gin.Engine, healthHandler *handlers.HealthHandler, jwtManager *middleware.JWTManager) {
	// Health check routes (no authentication required)
	health := router.Group("/health")
	{
		// Comprehensive health check
		health.GET("", healthHandler.Health)
		
		// Kubernetes-style checks
		health.GET("/live", healthHandler.LivenessCheck)
		health.GET("/ready", healthHandler.ReadinessCheck)
	}
	
	// Admin-only health endpoints
	admin := router.Group("/api/health")
	admin.Use(jwtManager.AuthRequired())
	admin.Use(middleware.RoleRequired("admin"))
	{
		// Database statistics
		admin.GET("/database-stats", healthHandler.DatabaseStats)
	}
}
