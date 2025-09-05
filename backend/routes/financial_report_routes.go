package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"github.com/gin-gonic/gin"
)

// SetupFinancialReportRoutes sets up all financial report related routes
func SetupFinancialReportRoutes(protected *gin.RouterGroup, controller *controllers.FinancialReportController) {
	// Financial Reports routes (enhanced version)
	reports := protected.Group("/reports/enhanced")
	reports.Use(middleware.RoleRequired("admin", "finance", "director")) // Only financial roles can access
	{
		// Core Financial Reports
		reports.POST("/profit-loss", controller.GenerateProfitLossStatement)
		reports.POST("/balance-sheet", controller.GenerateBalanceSheet)
		reports.POST("/cash-flow", controller.GenerateCashFlowStatement)
		reports.POST("/trial-balance", controller.GenerateTrialBalance)
		reports.GET("/general-ledger/:account_id", controller.GenerateGeneralLedger)

		// Advanced Reports and Analytics
		reports.GET("/dashboard", controller.GetFinancialDashboard)
		reports.GET("/real-time-metrics", controller.GetRealTimeMetrics)
		reports.GET("/financial-ratios", controller.CalculateFinancialRatios)
		reports.GET("/health-score", controller.GetFinancialHealthScore)

		// Report Metadata and Utilities
		reports.GET("/list", controller.GetReportsList)
		reports.POST("/validate", controller.ValidateReportRequest)
		reports.GET("/export-formats", controller.GetReportFormats)
		reports.GET("/summary", controller.GetReportSummary)
		
		// Quick Statistics (accessible to more roles for dashboard widgets)
		reports.GET("/quick-stats", middleware.RoleRequired("admin", "finance", "director", "inventory_manager", "employee"), controller.GetQuickStats)

		// Export functionality (placeholder for future implementation)
		reports.POST("/export", controller.ExportReport)
	}
}
