package routes

import (
	"strings"
	
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterEnhancedReportRoutes registers comprehensive financial and operational reporting routes
func RegisterEnhancedReportRoutes(router *gin.Engine, enhancedReportController *controllers.EnhancedReportController) {
	// Create report routes group with authentication and authorization
	reportsGroup := router.Group("/api/reports")
	reportsGroup.Use(middleware.AuthMiddleware())
	reportsGroup.Use(middleware.RoleRequired("finance", "admin", "director", "auditor"))

	// Comprehensive financial reports group
	comprehensiveGroup := reportsGroup.Group("/comprehensive")
	
	// FINANCIAL STATEMENTS
	// Balance Sheet - provides detailed assets, liabilities, and equity analysis
	comprehensiveGroup.GET("/balance-sheet", enhancedReportController.GetComprehensiveBalanceSheet)
	
	// Profit & Loss Statement - comprehensive income statement analysis
	comprehensiveGroup.GET("/profit-loss", enhancedReportController.GetComprehensiveProfitLoss)
	
	// Cash Flow Statement - operating, investing, and financing activities
	comprehensiveGroup.GET("/cash-flow", enhancedReportController.GetComprehensiveCashFlow)
	
	// OPERATIONAL REPORTS
	// Sales Summary - comprehensive sales analytics and business intelligence
	comprehensiveGroup.GET("/sales-summary", enhancedReportController.GetComprehensiveSalesSummary)
	
	// Purchase Summary - comprehensive purchase analytics and cost analysis
	comprehensiveGroup.GET("/purchase-summary", enhancedReportController.GetComprehensivePurchaseSummary)

	// DASHBOARD & METADATA
	// Financial Dashboard - comprehensive executive dashboard with key metrics
	reportsGroup.GET("/financial-dashboard", enhancedReportController.GetFinancialDashboard)
	
	// Available Reports - metadata about all available comprehensive reports
	reportsGroup.GET("/available", enhancedReportController.GetAvailableReports)
	
	// Preview endpoints for quick report previews
	reportsGroup.GET("/preview/:type", enhancedReportController.GetReportPreview)

	// LEGACY COMPATIBILITY ROUTES (Optional)
	// These routes maintain backward compatibility with existing frontend code
	// while providing the enhanced functionality under the hood
	
	legacyGroup := reportsGroup.Group("/legacy")
	
	// Map legacy routes to enhanced versions for backward compatibility
	legacyGroup.GET("/balance-sheet", enhancedReportController.GetComprehensiveBalanceSheet)
	legacyGroup.GET("/profit-loss", enhancedReportController.GetComprehensiveProfitLoss)
	legacyGroup.GET("/sales-summary", enhancedReportController.GetComprehensiveSalesSummary)
	legacyGroup.GET("/purchase-summary", enhancedReportController.GetComprehensivePurchaseSummary)

	// API Documentation routes (for development/testing)
	if gin.Mode() == gin.DebugMode {
		docsGroup := reportsGroup.Group("/docs")
		
		// Endpoint documentation
		docsGroup.GET("/endpoints", func(c *gin.Context) {
			endpoints := []gin.H{
				{
					"endpoint":    "GET /api/reports/comprehensive/balance-sheet",
					"description": "Generate comprehensive balance sheet",
					"parameters": gin.H{
						"as_of_date": "string (optional, format: YYYY-MM-DD, default: today)",
						"format":     "string (optional, values: json|pdf|excel, default: json)",
					},
					"response": "BalanceSheetData object or binary file",
				},
				{
					"endpoint":    "GET /api/reports/comprehensive/profit-loss",
					"description": "Generate comprehensive profit & loss statement",
					"parameters": gin.H{
						"start_date": "string (required, format: YYYY-MM-DD)",
						"end_date":   "string (required, format: YYYY-MM-DD)",
						"format":     "string (optional, values: json|pdf|excel, default: json)",
					},
					"response": "ProfitLossData object or binary file",
				},
				{
					"endpoint":    "GET /api/reports/comprehensive/cash-flow",
					"description": "Generate comprehensive cash flow statement",
					"parameters": gin.H{
						"start_date": "string (required, format: YYYY-MM-DD)",
						"end_date":   "string (required, format: YYYY-MM-DD)",
						"format":     "string (optional, values: json|pdf, default: json)",
					},
					"response": "CashFlowData object or binary file",
				},
				{
					"endpoint":    "GET /api/reports/comprehensive/sales-summary",
					"description": "Generate comprehensive sales summary with analytics",
					"parameters": gin.H{
						"start_date": "string (required, format: YYYY-MM-DD)",
						"end_date":   "string (required, format: YYYY-MM-DD)",
						"group_by":   "string (optional, values: day|week|month|quarter|year, default: month)",
						"format":     "string (optional, values: json|pdf|excel, default: json)",
					},
					"response": "SalesSummaryData object or binary file",
				},
				{
					"endpoint":    "GET /api/reports/comprehensive/purchase-summary",
					"description": "Generate comprehensive purchase summary with analytics",
					"parameters": gin.H{
						"start_date": "string (required, format: YYYY-MM-DD)",
						"end_date":   "string (required, format: YYYY-MM-DD)",
						"group_by":   "string (optional, values: day|week|month|quarter|year, default: month)",
						"format":     "string (optional, values: json|pdf, default: json)",
					},
					"response": "PurchaseSummaryData object or binary file",
				},
				{
					"endpoint":    "GET /api/reports/financial-dashboard",
					"description": "Get comprehensive financial dashboard with key metrics",
					"parameters": gin.H{
						"start_date": "string (optional, format: YYYY-MM-DD, default: first day of current month)",
						"end_date":   "string (optional, format: YYYY-MM-DD, default: today)",
					},
					"response": "Financial dashboard object with all key metrics and ratios",
				},
				{
					"endpoint":    "GET /api/reports/available",
					"description": "Get metadata about all available comprehensive reports",
					"parameters": gin.H{},
					"response":    "Array of report metadata objects",
				},
			}

			c.JSON(200, gin.H{
				"status":    "success",
				"message":   "Enhanced Report API Documentation",
				"version":   "1.0.0",
				"endpoints": endpoints,
			})
		})

		// Sample requests for testing
		docsGroup.GET("/examples", func(c *gin.Context) {
			examples := gin.H{
				"balance_sheet_examples": []gin.H{
					{
						"description": "Get balance sheet as of today (JSON)",
						"url":         "/api/reports/comprehensive/balance-sheet",
					},
					{
						"description": "Get balance sheet as of specific date (PDF)",
						"url":         "/api/reports/comprehensive/balance-sheet?as_of_date=2024-12-31&format=pdf",
					},
					{
						"description": "Get balance sheet as Excel file",
						"url":         "/api/reports/comprehensive/balance-sheet?format=excel",
					},
				},
				"profit_loss_examples": []gin.H{
					{
						"description": "Get P&L for current month (JSON)",
						"url":         "/api/reports/comprehensive/profit-loss?start_date=2024-01-01&end_date=2024-01-31",
					},
					{
						"description": "Get quarterly P&L as PDF",
						"url":         "/api/reports/comprehensive/profit-loss?start_date=2024-01-01&end_date=2024-03-31&format=pdf",
					},
				},
				"sales_summary_examples": []gin.H{
					{
						"description": "Get monthly sales summary",
						"url":         "/api/reports/comprehensive/sales-summary?start_date=2024-01-01&end_date=2024-12-31&group_by=month",
					},
					{
						"description": "Get quarterly sales summary as PDF",
						"url":         "/api/reports/comprehensive/sales-summary?start_date=2024-01-01&end_date=2024-12-31&group_by=quarter&format=pdf",
					},
				},
				"dashboard_examples": []gin.H{
					{
						"description": "Get financial dashboard for current month",
						"url":         "/api/reports/financial-dashboard",
					},
					{
						"description": "Get financial dashboard for specific period",
						"url":         "/api/reports/financial-dashboard?start_date=2024-01-01&end_date=2024-12-31",
					},
				},
			}

			c.JSON(200, gin.H{
				"status":   "success",
				"message":  "Enhanced Report API Examples",
				"examples": examples,
			})
		})
	}
}

// RegisterEnhancedReportRoutesWithPrefix registers enhanced report routes with a custom prefix
func RegisterEnhancedReportRoutesWithPrefix(router *gin.Engine, prefix string, enhancedReportController *controllers.EnhancedReportController) {
	// Create report routes group with custom prefix
	reportsGroup := router.Group(prefix)
	reportsGroup.Use(middleware.AuthMiddleware())
	reportsGroup.Use(middleware.RoleRequired("finance", "admin", "director", "auditor"))

	// Financial statements
	reportsGroup.GET("/balance-sheet", enhancedReportController.GetComprehensiveBalanceSheet)
	reportsGroup.GET("/profit-loss", enhancedReportController.GetComprehensiveProfitLoss)
	reportsGroup.GET("/cash-flow", enhancedReportController.GetComprehensiveCashFlow)
	
	// Operational reports
	reportsGroup.GET("/sales-summary", enhancedReportController.GetComprehensiveSalesSummary)
	reportsGroup.GET("/purchase-summary", enhancedReportController.GetComprehensivePurchaseSummary)
	
	// Dashboard and metadata
	reportsGroup.GET("/dashboard", enhancedReportController.GetFinancialDashboard)
	reportsGroup.GET("/available", enhancedReportController.GetAvailableReports)
}

// RegisterEnhancedReportMiddleware registers middleware specific to enhanced reporting
func RegisterEnhancedReportMiddleware(router *gin.Engine) {
	// Add report-specific middleware here if needed
	// For example: rate limiting, caching, request logging, etc.
	
	router.Use(func(c *gin.Context) {
		// Add custom headers for enhanced reports
		if strings.Contains(c.Request.URL.Path, "/api/reports/comprehensive") {
			c.Header("X-Report-Version", "2.0")
			c.Header("X-Report-Features", "enhanced,comprehensive,analytics")
		}
		c.Next()
	})
}
