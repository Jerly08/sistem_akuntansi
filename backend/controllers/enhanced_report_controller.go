package controllers

import (
	"fmt"
	"net/http"
	"time"

	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"

	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
)

// EnhancedReportController handles comprehensive financial and operational reporting endpoints
// Updated to support SSOT P&L integration
type EnhancedReportController struct {
	db *gorm.DB
}

// NewEnhancedReportController creates a new enhanced report controller
func NewEnhancedReportController(db *gorm.DB) *EnhancedReportController {
	return &EnhancedReportController{
		db: db,
	}
}

// GetComprehensiveBalanceSheet returns comprehensive balance sheet data from SSOT journal system
func (erc *EnhancedReportController) GetComprehensiveBalanceSheet(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	asOfDate := c.DefaultQuery("as_of_date", time.Now().Format("2006-01-02"))
	
	// Create SSOT Balance Sheet controller and delegate to it
	// This integrates the SSOT journal system with the enhanced report controller
	ssotController := NewSSOTBalanceSheetController(erc.db)
	
	// Set the format in the query parameters for the SSOT controller
	c.Request.URL.RawQuery = "as_of_date=" + asOfDate + "&format=" + format
	ssotController.GenerateSSOTBalanceSheet(c)
}

// GetComprehensiveProfitLoss generates P&L report using SSOT journal system
func (erc *EnhancedReportController) GetComprehensiveProfitLoss(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	format := c.DefaultQuery("format", "json")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	// Create SSOT P&L controller and delegate to it
	// This integrates the SSOT journal system with the enhanced report controller
	ssotController := NewSSOTProfitLossController(erc.db)
	
	// Set the format in the query parameters for the SSOT controller
	c.Request.URL.RawQuery = c.Request.URL.RawQuery + "&format=" + format
	ssotController.GetSSOTProfitLoss(c)
}

// GetComprehensiveCashFlow returns cash flow data with format support
func (erc *EnhancedReportController) GetComprehensiveCashFlow(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	format := c.DefaultQuery("format", "json")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	// For now, return error for non-JSON formats with user-friendly message
	if format == "pdf" || format == "csv" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": format + " export for Cash Flow is not yet implemented. Please use the View Report option and export from there.",
			"error_code": "FORMAT_NOT_SUPPORTED",
			"supported_formats": []string{"json"},
			"alternative": "Use the 'View Report' button to access the SSOT Cash Flow with export options",
		})
		return
	}

	emptyCashFlowData := gin.H{
		"start_date":             startDate,
		"end_date":               endDate,
		"company_name":           "Sistema Akuntansi",
		"beginning_cash":         0,
		"ending_cash":           0,
		"net_cash_flow":         0,
		"operating_activities":  0,
		"investing_activities":  0,
		"financing_activities":  0,
		"message":               "Report module is in safe mode - use SSOT Cash Flow for real data",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   emptyCashFlowData,
	})
}

// GetComprehensiveSalesSummary returns sales summary data with format support
func (erc *EnhancedReportController) GetComprehensiveSalesSummary(c *gin.Context) {
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "month")
	format := c.DefaultQuery("format", "json")

	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	// Parse dates
	start, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}
	end, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}

	// Handle export formats (CSV/PDF)
	if format == "csv" || format == "pdf" {
		// Build service dependencies inline (no DI available in this slim controller)
		accountRepo := repositories.NewAccountRepository(erc.db)
		salesRepo := repositories.NewSalesRepository(erc.db)
		purchaseRepo := repositories.NewPurchaseRepository(erc.db)
		productRepo := repositories.NewProductRepository(erc.db)
		contactRepo := repositories.NewContactRepository(erc.db)
		paymentRepo := repositories.NewPaymentRepository(erc.db)
		cashBankRepo := repositories.NewCashBankRepository(erc.db)
		cacheService := services.NewReportCacheService()
		enhancedReportService := services.NewEnhancedReportService(erc.db, accountRepo, salesRepo, purchaseRepo, productRepo, contactRepo, paymentRepo, cashBankRepo, cacheService)

		// Generate data
		summary, genErr := enhancedReportService.GenerateSalesSummary(start, end, groupBy)
		if genErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate sales summary",
				"error":   genErr.Error(),
			})
			return
		}

		// Export
		exporter := services.NewSalesSummaryExportService()
		if format == "csv" {
			bytes, err := exporter.ExportToCSV(summary)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to generate CSV", "error": err.Error()})
				return
			}
			filename := exporter.GetCSVFilename(summary)
			c.Header("Content-Type", "text/csv")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
			c.Header("Content-Length", fmt.Sprintf("%d", len(bytes)))
			c.Data(http.StatusOK, "text/csv", bytes)
			return
		}

		// PDF
		bytes, err := exporter.ExportToPDF(summary)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to generate PDF", "error": err.Error()})
			return
		}
		filename := exporter.GetPDFFilename(summary)
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		c.Header("Content-Length", fmt.Sprintf("%d", len(bytes)))
		c.Data(http.StatusOK, "application/pdf", bytes)
		return
	}

	// Default JSON (safe mode placeholder)
	emptySalesSummaryData := gin.H{
		"start_date":         startDateStr,
		"end_date":           endDateStr,
		"company_name":       "Sistema Akuntansi",
		"total_revenue":      0,
		"total_transactions": 0,
		"average_order_value": 0,
		"total_customers":    0,
		"sales_by_period":    []gin.H{},
		"message":            "Report module is in safe mode - no data integration",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   emptySalesSummaryData,
	})
}

// GetComprehensivePurchaseSummary returns safe empty purchase summary data
func (erc *EnhancedReportController) GetComprehensivePurchaseSummary(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	emptyPurchaseSummaryData := gin.H{
		"start_date":           startDate,
		"end_date":             endDate,
		"company_name":         "Sistema Akuntansi",
		"total_purchases":      0,
		"total_transactions":   0,
		"average_purchase_value": 0,
		"total_vendors":        0,
		"purchases_by_period":  []gin.H{},
		"message":              "Report module is in safe mode - no data integration",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   emptyPurchaseSummaryData,
	})
}

// GetVendorAnalysis returns safe empty vendor analysis data
func (erc *EnhancedReportController) GetVendorAnalysis(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	emptyVendorData := gin.H{
		"start_date":     startDate,
		"end_date":       endDate,
		"company_name":   "Sistema Akuntansi",
		"total_vendors":  0,
		"vendor_list":    []gin.H{},
		"message":        "Report module is in safe mode - no data integration",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   emptyVendorData,
	})
}

// GetTrialBalance returns trial balance data with format support
func (erc *EnhancedReportController) GetTrialBalance(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	asOfDate := c.DefaultQuery("as_of_date", time.Now().Format("2006-01-02"))
	
	// For now, return error for non-JSON formats with user-friendly message
	if format == "pdf" || format == "csv" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": format + " export for Trial Balance is not yet implemented. Please use the View Report option and export from there.",
			"error_code": "FORMAT_NOT_SUPPORTED",
			"supported_formats": []string{"json"},
			"alternative": "Use the 'View Report' button to access the SSOT Trial Balance with export options",
		})
		return
	}
	
	emptyTrialBalanceData := gin.H{
		"report_date":  asOfDate,
		"company_name": "Sistema Akuntansi",
		"accounts":     []gin.H{},
		"total_debits": 0,
		"total_credits": 0,
		"is_balanced":  true,
		"message":      "Report module is in safe mode - use SSOT Trial Balance for real data",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   emptyTrialBalanceData,
	})
}

// GetGeneralLedger returns general ledger data with format support
func (erc *EnhancedReportController) GetGeneralLedger(c *gin.Context) {
	accountIDStr := c.Query("account_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	format := c.DefaultQuery("format", "json")

	if accountIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "account_id is required. Use specific account ID or 'all' for all accounts",
		})
		return
	}

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	// For now, return error for non-JSON formats with user-friendly message
	if format == "pdf" || format == "csv" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": format + " export for General Ledger is not yet implemented. Please use the View Report option and export from there.",
			"error_code": "FORMAT_NOT_SUPPORTED",
			"supported_formats": []string{"json"},
			"alternative": "Use the 'View Report' button to access the SSOT General Ledger with export options",
		})
		return
	}

	emptyGeneralLedgerData := gin.H{
		"account_id":   accountIDStr,
		"start_date":   startDate,
		"end_date":     endDate,
		"company_name": "Sistema Akuntansi",
		"transactions": []gin.H{},
		"beginning_balance": 0,
		"ending_balance":    0,
		"message":          "Report module is in safe mode - use SSOT General Ledger for real data",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   emptyGeneralLedgerData,
	})
}

// GetJournalEntryAnalysis returns journal entry analysis data with format support
func (erc *EnhancedReportController) GetJournalEntryAnalysis(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	format := c.DefaultQuery("format", "json")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	// For now, return error for non-JSON formats with user-friendly message
	if format == "pdf" || format == "csv" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": format + " export for Journal Entry Analysis is not yet implemented. Please use the View Report option and export from there.",
			"error_code": "FORMAT_NOT_SUPPORTED",
			"supported_formats": []string{"json"},
			"alternative": "Use the 'View Report' button to access the SSOT Journal Analysis with export options",
		})
		return
	}

	emptyAnalysisData := gin.H{
		"start_date":     startDate,
		"end_date":       endDate,
		"company_name":   "Sistema Akuntansi",
		"journal_entries": []gin.H{},
		"total_entries":  0,
		"message":        "Report module is in safe mode - use SSOT Journal Analysis for real data",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   emptyAnalysisData,
	})
}

// GetFinancialDashboard returns safe empty dashboard data
func (erc *EnhancedReportController) GetFinancialDashboard(c *gin.Context) {
	emptyDashboardData := gin.H{
		"period": gin.H{
			"start_date": time.Now().AddDate(0, -1, 0).Format("2006-01-02"),
			"end_date":   time.Now().Format("2006-01-02"),
		},
		"balance_sheet": gin.H{
			"total_assets":      0,
			"total_liabilities": 0,
			"total_equity":      0,
		},
		"profit_loss": gin.H{
			"total_revenue": 0,
			"net_income":    0,
		},
		"cash_flow": gin.H{
			"net_cash_flow": 0,
		},
		"key_ratios": gin.H{
			"current_ratio":   0,
			"debt_to_equity": 0,
		},
		"message": "Report module is in safe mode - no data integration",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   emptyDashboardData,
	})
}

// GetAvailableReports returns metadata about available reports
func (erc *EnhancedReportController) GetAvailableReports(c *gin.Context) {
	reports := []gin.H{
	{
		"id":          "comprehensive_balance_sheet",
		"name":        "Balance Sheet",
		"type":        "FINANCIAL",
		"description": "Enhanced Balance Sheet from SSOT Journal System",
		"endpoint":    "/api/reports/balance-sheet",
		"status":      "SSOT_INTEGRATED",
	},
		{
			"id":          "comprehensive_profit_loss",
			"name":        "Profit & Loss Statement",
			"type":        "FINANCIAL",
			"description": "Enhanced P&L statement from SSOT Journal System",
			"endpoint":    "/api/reports/profit-loss",
			"status":      "SSOT_INTEGRATED",
		},
		{
			"id":          "comprehensive_cash_flow",
			"name":        "Cash Flow Statement",
			"type":        "FINANCIAL",
			"description": "Cash flow statement (safe mode)",
			"endpoint":    "/api/reports/cash-flow",
			"status":      "SAFE_MODE",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    reports,
		"message": "Report module is in safe mode - no data integration",
	})
}

// GetReportPreview returns safe empty preview data
func (erc *EnhancedReportController) GetReportPreview(c *gin.Context) {
	reportType := c.Param("type")

	previewData := gin.H{
		"report_type":  reportType,
		"company_name": "Sistema Akuntansi",
		"preview_data": gin.H{},
		"message":      "Report module is in safe mode - no data integration",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   previewData,
		"meta":   gin.H{"is_preview": true, "safe_mode": true},
	})
}

// GetReportValidation returns safe validation data
func (erc *EnhancedReportController) GetReportValidation(c *gin.Context) {
	validationReport := gin.H{
		"validation_date": time.Now().Format("2006-01-02"),
		"company_name":    "Sistema Akuntansi",
		"status":          "SAFE_MODE",
		"checks":          []gin.H{},
		"message":         "Report module is in safe mode - validation disabled",
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   validationReport,
	})
}