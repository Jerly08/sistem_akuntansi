package routes

import (
	"encoding/json"

	"app-sistem-akuntansi/controllers"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupUnifiedReportRoutes configures all financial report routes under the unified controller
func SetupUnifiedReportRoutes(router *gin.Engine, db *gorm.DB) {
	// Initialize repositories
	accountRepo := repositories.NewAccountRepository(db)
	journalEntryRepo := repositories.NewJournalEntryRepository(db)
	salesRepo := repositories.NewSalesRepository(db)
	purchaseRepo := repositories.NewPurchaseRepository(db)
	cashBankRepo := repositories.NewCashBankRepository(db)
	paymentRepo := repositories.NewPaymentRepository(db)
	contactRepo := repositories.NewContactRepository(db)
	productRepo := repositories.NewProductRepository(db)

	// Initialize unified service
	unifiedService := services.NewUnifiedFinancialReportService(
		db,
		accountRepo,
		journalEntryRepo,
		salesRepo,
		purchaseRepo,
		paymentRepo,
		cashBankRepo,
		contactRepo,
		productRepo,
	)

	// Initialize unified controller
	unifiedController := controllers.NewUnifiedFinancialReportController(unifiedService)

	// Create route group for unified financial reports
	unifiedReportsGroup := router.Group("/api/unified-reports")
	{
		// ============== FINANCIAL STATEMENTS ==============
		
		// Profit & Loss Statement
		unifiedReportsGroup.GET("/profit-loss", unifiedController.GetProfitLossStatement)
		
		// Balance Sheet
		unifiedReportsGroup.GET("/balance-sheet", unifiedController.GetBalanceSheet)
		
		// Cash Flow Statement
		unifiedReportsGroup.GET("/cash-flow", unifiedController.GetCashFlowStatement)
		
		// ============== ACCOUNTING REPORTS ==============
		
		// Trial Balance
		unifiedReportsGroup.GET("/trial-balance", unifiedController.GetTrialBalance)
		
		// General Ledger (specific account)
		unifiedReportsGroup.GET("/general-ledger/:account_id", unifiedController.GetGeneralLedger)
		
		// ============== OPERATIONAL REPORTS ==============
		
		// Sales Summary Report
		unifiedReportsGroup.GET("/sales-summary", unifiedController.GetSalesSummaryReport)
		
		// Vendor Analysis Report
		unifiedReportsGroup.GET("/vendor-analysis", unifiedController.GetVendorAnalysisReport)
		
		// ============== DASHBOARDS & ANALYTICS ==============
		
		// Financial Dashboard
		unifiedReportsGroup.GET("/dashboard", unifiedController.GetFinancialDashboard)
		
		// ============== UTILITIES & METADATA ==============
		
		// Available Reports Metadata
		unifiedReportsGroup.GET("/available", unifiedController.GetAvailableReports)
		
		// Generate All Reports (Batch)
		unifiedReportsGroup.GET("/all", unifiedController.GenerateAllReports)
		
		// Validate Report Parameters
		unifiedReportsGroup.GET("/validate", unifiedController.ValidateReportParameters)
	}

	// Legacy compatibility routes removed to avoid conflicts with existing /api/reports/ routes
	// The unified financial reports are available under /api/unified-reports/ instead
}

// SetupReportMiddleware adds common middleware for report routes
func SetupReportMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Add CORS headers for report endpoints
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
}


// Route documentation structure for API documentation generation
type UnifiedReportRoutes struct {
	BaseURL     string                    `json:"base_url"`
	Version     string                    `json:"version"`
	Description string                    `json:"description"`
	Routes      []UnifiedReportRouteInfo `json:"routes"`
}

type UnifiedReportRouteInfo struct {
	Path        string                 `json:"path"`
	Method      string                 `json:"method"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Example     string                 `json:"example"`
}

// GetUnifiedReportRouteDocumentation returns route documentation
func GetUnifiedReportRouteDocumentation() UnifiedReportRoutes {
	return UnifiedReportRoutes{
		BaseURL:     "/api/unified-reports",
		Version:     "1.0",
		Description: "Unified Financial Reporting API endpoints",
		Routes: []UnifiedReportRouteInfo{
			{
				Path:        "/profit-loss",
				Method:      "GET",
				Description: "Generate Profit & Loss Statement",
				Parameters: map[string]interface{}{
					"start_date":   "string (required, YYYY-MM-DD)",
					"end_date":     "string (required, YYYY-MM-DD)",
					"comparative":  "boolean (optional, default: false)",
				},
				Example: "/api/unified-reports/profit-loss?start_date=2024-01-01&end_date=2024-12-31&comparative=true",
			},
			{
				Path:        "/balance-sheet",
				Method:      "GET",
				Description: "Generate Balance Sheet",
				Parameters: map[string]interface{}{
					"as_of_date":   "string (optional, YYYY-MM-DD, default: today)",
					"comparative":  "boolean (optional, default: false)",
				},
				Example: "/api/unified-reports/balance-sheet?as_of_date=2024-12-31&comparative=false",
			},
			{
				Path:        "/cash-flow",
				Method:      "GET",
				Description: "Generate Cash Flow Statement",
				Parameters: map[string]interface{}{
					"start_date": "string (required, YYYY-MM-DD)",
					"end_date":   "string (required, YYYY-MM-DD)",
				},
				Example: "/api/unified-reports/cash-flow?start_date=2024-01-01&end_date=2024-12-31",
			},
			{
				Path:        "/trial-balance",
				Method:      "GET",
				Description: "Generate Trial Balance",
				Parameters: map[string]interface{}{
					"as_of_date": "string (optional, YYYY-MM-DD, default: today)",
					"show_zero":  "boolean (optional, default: false)",
				},
				Example: "/api/unified-reports/trial-balance?as_of_date=2024-12-31&show_zero=true",
			},
			{
				Path:        "/general-ledger/:account_id",
				Method:      "GET",
				Description: "Generate General Ledger for specific account",
				Parameters: map[string]interface{}{
					"account_id": "uint (required, path parameter)",
					"start_date": "string (required, YYYY-MM-DD)",
					"end_date":   "string (required, YYYY-MM-DD)",
				},
				Example: "/api/unified-reports/general-ledger/1?start_date=2024-01-01&end_date=2024-12-31",
			},
			{
				Path:        "/sales-summary",
				Method:      "GET",
				Description: "Generate Sales Summary Report",
				Parameters: map[string]interface{}{
					"start_date": "string (required, YYYY-MM-DD)",
					"end_date":   "string (required, YYYY-MM-DD)",
				},
				Example: "/api/unified-reports/sales-summary?start_date=2024-01-01&end_date=2024-12-31",
			},
			{
				Path:        "/vendor-analysis",
				Method:      "GET",
				Description: "Generate Vendor Analysis Report",
				Parameters: map[string]interface{}{
					"start_date": "string (required, YYYY-MM-DD)",
					"end_date":   "string (required, YYYY-MM-DD)",
				},
				Example: "/api/unified-reports/vendor-analysis?start_date=2024-01-01&end_date=2024-12-31",
			},
			{
				Path:        "/dashboard",
				Method:      "GET",
				Description: "Generate Financial Dashboard",
				Parameters: map[string]interface{}{
					"start_date": "string (optional, YYYY-MM-DD, default: first day of current month)",
					"end_date":   "string (optional, YYYY-MM-DD, default: today)",
				},
				Example: "/api/unified-reports/dashboard?start_date=2024-01-01&end_date=2024-12-31",
			},
			{
				Path:        "/available",
				Method:      "GET",
				Description: "Get metadata about all available reports",
				Parameters:  map[string]interface{}{},
				Example:    "/api/unified-reports/available",
			},
			{
				Path:        "/all",
				Method:      "GET",
				Description: "Generate all financial reports in batch",
				Parameters: map[string]interface{}{
					"start_date": "string (required, YYYY-MM-DD)",
					"end_date":   "string (required, YYYY-MM-DD)",
					"as_of_date": "string (optional, YYYY-MM-DD, default: today)",
				},
				Example: "/api/unified-reports/all?start_date=2024-01-01&end_date=2024-12-31",
			},
			{
				Path:        "/validate",
				Method:      "GET",
				Description: "Validate report parameters before generation",
				Parameters: map[string]interface{}{
					"start_date": "string (optional, YYYY-MM-DD)",
					"end_date":   "string (optional, YYYY-MM-DD)",
					"as_of_date": "string (optional, YYYY-MM-DD)",
				},
				Example: "/api/unified-reports/validate?start_date=2024-01-01&end_date=2024-12-31",
			},
		},
	}
}

// Helper function to get route documentation as JSON
func GetRouteDocumentationJSON() ([]byte, error) {
	doc := GetUnifiedReportRouteDocumentation()
	return json.Marshal(doc)
}
