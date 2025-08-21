package routes

import (
	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/middleware"
	"github.com/gin-gonic/gin"
)

// SetupReportRoutes sets up all report-related routes
func SetupReportRoutes(protected *gin.RouterGroup, reportController *controllers.ReportController) {
	// Reports routes with role-based access control
	reports := protected.Group("/reports")
	reports.Use(middleware.RoleRequired("admin", "director", "finance"))
	{
		// Get available reports list
		reports.GET("", reportController.GetReportsList)
		
		// Core Financial Statements
		reports.GET("/balance-sheet", reportController.GetBalanceSheet)
		reports.GET("/profit-loss", reportController.GetProfitLoss)
		reports.GET("/cash-flow", reportController.GetCashFlow)
		reports.GET("/trial-balance", reportController.GetTrialBalance)
		reports.GET("/general-ledger", reportController.GetGeneralLedger)
		
		// Receivables & Payables Reports
		reports.GET("/accounts-receivable", reportController.GetAccountsReceivable)
		reports.GET("/accounts-payable", reportController.GetAccountsPayable)
		
		// Operational Reports
		reports.GET("/sales-summary", reportController.GetSalesSummary)
		reports.GET("/purchase-summary", reportController.GetPurchaseSummary)
		reports.GET("/inventory-report", reportController.GetInventoryReport)
		
		// Analysis Reports
		reports.GET("/financial-ratios", reportController.GetFinancialRatios)
		
		// Report Templates Management (admin only)
		templates := reports.Group("/templates")
		templates.Use(middleware.RoleRequired("admin"))
		{
			templates.GET("", reportController.GetReportTemplates)
			templates.POST("", reportController.SaveReportTemplate)
		}
	}
}
