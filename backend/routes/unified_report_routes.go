package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterUnifiedReportRoutes registers the unified report routes with proper endpoint structure
func RegisterUnifiedReportRoutes(router *gin.Engine, controller *controllers.UnifiedReportController, jwtManager *middleware.JWTManager) {
	// Create the main reports group that matches frontend expectations - using v1 API path
	reportsGroup := router.Group("/api/v1/reports")
	reportsGroup.Use(jwtManager.AuthRequired())
	reportsGroup.Use(middleware.RoleRequired("finance", "admin", "director"))

	// Direct report endpoints (matching frontend service expectations)
	reportsGroup.GET("/balance-sheet", controller.GenerateReport)
	reportsGroup.GET("/profit-loss", controller.GenerateReport)
	reportsGroup.GET("/cash-flow", controller.GenerateReport) 
	reportsGroup.GET("/trial-balance", controller.GenerateReport)
	reportsGroup.GET("/general-ledger", controller.GenerateReport)
	reportsGroup.GET("/sales-summary", controller.GenerateReport)
	reportsGroup.GET("/vendor-analysis", controller.GenerateReport)
	
	// Also register unified-reports endpoints for frontend compatibility
	unifiedGroup := router.Group("/api/v1/unified-reports")
	unifiedGroup.Use(jwtManager.AuthRequired())
	unifiedGroup.Use(middleware.RoleRequired("finance", "admin", "director"))
	
	unifiedGroup.GET("/balance-sheet", controller.GenerateReport)
	unifiedGroup.GET("/profit-loss", controller.GenerateReport)
	unifiedGroup.GET("/cash-flow", controller.GenerateReport)
	unifiedGroup.GET("/trial-balance", controller.GenerateReport)
	unifiedGroup.GET("/general-ledger", controller.GenerateReport)
	unifiedGroup.GET("/sales-summary", controller.GenerateReport)
	unifiedGroup.GET("/vendor-analysis", controller.GenerateReport)

	// Preview endpoints for quick data preview
	reportsGroup.GET("/preview/:type", controller.PreviewReport)

	// Metadata and discovery endpoints
	reportsGroup.GET("/available", controller.GetAvailableReports)

	// RESTful endpoints with report type as parameter
	reportsGroup.GET("/:type", controller.GenerateReport)
}

// RegisterUnifiedReportMiddleware adds unified report middleware
func RegisterUnifiedReportMiddleware(router *gin.Engine) {
	router.Use(func(c *gin.Context) {
		// Add report-specific headers
		c.Header("X-Report-API-Version", "2.0")
		c.Header("X-Report-System", "unified")
		c.Next()
	})
}

// RegisterLegacyReportCompatibility maintains backward compatibility
func RegisterLegacyReportCompatibility(router *gin.Engine, controller *controllers.UnifiedReportController) {
	// Legacy comprehensive routes for backward compatibility
	legacyGroup := router.Group("/api/reports/comprehensive")
	legacyGroup.Use(middleware.AuthMiddleware())
	legacyGroup.Use(middleware.RoleRequired("finance", "admin", "director", "auditor"))

	legacyGroup.GET("/balance-sheet", func(c *gin.Context) {
		c.Param("type") // Override param
		c.Set("report_type", "balance-sheet")
		controller.GenerateReport(c)
	})
	
	legacyGroup.GET("/profit-loss", func(c *gin.Context) {
		c.Set("report_type", "profit-loss") 
		controller.GenerateReport(c)
	})
	
	legacyGroup.GET("/cash-flow", func(c *gin.Context) {
		c.Set("report_type", "cash-flow")
		controller.GenerateReport(c)
	})
	
	legacyGroup.GET("/sales-summary", func(c *gin.Context) {
		c.Set("report_type", "sales-summary")
		controller.GenerateReport(c)
	})
	
	legacyGroup.GET("/purchase-summary", func(c *gin.Context) {
		c.Set("report_type", "vendor-analysis")
		controller.GenerateReport(c)
	})
}
