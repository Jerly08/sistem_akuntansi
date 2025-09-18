package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"

	"github.com/gin-gonic/gin"
)

// UnifiedFinancialReportController handles all financial reporting endpoints
type UnifiedFinancialReportController struct {
	unifiedReportService *services.UnifiedFinancialReportService
}

// NewUnifiedFinancialReportController creates a new unified financial report controller
func NewUnifiedFinancialReportController(unifiedReportService *services.UnifiedFinancialReportService) *UnifiedFinancialReportController {
	return &UnifiedFinancialReportController{
		unifiedReportService: unifiedReportService,
	}
}

// Note: P&L Statement method removed - use Enhanced P&L Controller at /api/reports/enhanced/profit-loss instead

// ========================= BALANCE SHEET =========================

// GetBalanceSheet generates Balance Sheet
func (c *UnifiedFinancialReportController) GetBalanceSheet(ctx *gin.Context) {
	asOfDateStr := ctx.DefaultQuery("as_of_date", time.Now().Format("2006-01-02"))
	comparativeStr := ctx.DefaultQuery("comparative", "false")

	asOfDate, err := time.Parse("2006-01-02", asOfDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid as_of_date format. Use YYYY-MM-DD",
		})
		return
	}

	comparative := comparativeStr == "true"

	balanceSheet, err := c.unifiedReportService.GenerateComprehensiveBalanceSheet(ctx.Request.Context(), asOfDate, comparative)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate Balance Sheet",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Balance Sheet generated successfully",
		"data":    balanceSheet,
	})
}

// ========================= CASH FLOW STATEMENT =========================

// GetCashFlowStatement generates Cash Flow Statement
func (c *UnifiedFinancialReportController) GetCashFlowStatement(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	cashFlow, err := c.unifiedReportService.GenerateComprehensiveCashFlow(ctx.Request.Context(), startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate Cash Flow Statement",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Cash Flow Statement generated successfully",
		"data":    cashFlow,
	})
}

// ========================= TRIAL BALANCE =========================

// GetTrialBalance generates Trial Balance
func (c *UnifiedFinancialReportController) GetTrialBalance(ctx *gin.Context) {
	asOfDateStr := ctx.DefaultQuery("as_of_date", time.Now().Format("2006-01-02"))
	showZeroStr := ctx.DefaultQuery("show_zero", "false")

	asOfDate, err := time.Parse("2006-01-02", asOfDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid as_of_date format. Use YYYY-MM-DD",
		})
		return
	}

	showZero := showZeroStr == "true"

	trialBalance, err := c.unifiedReportService.GenerateComprehensiveTrialBalance(ctx.Request.Context(), asOfDate, showZero)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate Trial Balance",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Trial Balance generated successfully",
		"data":    trialBalance,
	})
}

// ========================= GENERAL LEDGER =========================

// GetGeneralLedger generates General Ledger for a specific account
func (c *UnifiedFinancialReportController) GetGeneralLedger(ctx *gin.Context) {
	accountIDStr := ctx.Param("account_id")
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	if accountIDStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "account_id is required",
		})
		return
	}

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid account_id",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	generalLedger, err := c.unifiedReportService.GenerateComprehensiveGeneralLedger(ctx.Request.Context(), uint(accountID), startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate General Ledger",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "General Ledger generated successfully",
		"data":    generalLedger,
	})
}

// ========================= SALES SUMMARY REPORT =========================

// GetSalesSummaryReport generates Sales Summary Report
func (c *UnifiedFinancialReportController) GetSalesSummaryReport(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	salesSummary, err := c.unifiedReportService.GenerateComprehensiveSalesSummary(ctx.Request.Context(), startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate Sales Summary Report",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Sales Summary Report generated successfully",
		"data":    salesSummary,
	})
}

// ========================= VENDOR ANALYSIS REPORT =========================

// GetVendorAnalysisReport generates Vendor Analysis Report
func (c *UnifiedFinancialReportController) GetVendorAnalysisReport(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	vendorAnalysis, err := c.unifiedReportService.GenerateComprehensiveVendorAnalysis(ctx.Request.Context(), startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate Vendor Analysis Report",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Vendor Analysis Report generated successfully",
		"data":    vendorAnalysis,
	})
}

// ========================= COMPREHENSIVE DASHBOARD =========================

// GetFinancialDashboard provides comprehensive financial dashboard
func (c *UnifiedFinancialReportController) GetFinancialDashboard(ctx *gin.Context) {
	// Default to current month
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endDate := now

	// Parse optional date parameters
	if startDateStr := ctx.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := ctx.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	// Generate all reports for dashboard
	pnl, err := c.unifiedReportService.GenerateComprehensiveProfitLoss(ctx.Request.Context(), startDate, endDate, false)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate P&L for dashboard",
			"error":   err.Error(),
		})
		return
	}

	balanceSheet, err := c.unifiedReportService.GenerateComprehensiveBalanceSheet(ctx.Request.Context(), endDate, false)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate Balance Sheet for dashboard",
			"error":   err.Error(),
		})
		return
	}

	cashFlow, err := c.unifiedReportService.GenerateComprehensiveCashFlow(ctx.Request.Context(), startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate Cash Flow for dashboard",
			"error":   err.Error(),
		})
		return
	}

	salesSummary, err := c.unifiedReportService.GenerateComprehensiveSalesSummary(ctx.Request.Context(), startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate Sales Summary for dashboard",
			"error":   err.Error(),
		})
		return
	}

	vendorAnalysis, err := c.unifiedReportService.GenerateComprehensiveVendorAnalysis(ctx.Request.Context(), startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate Vendor Analysis for dashboard",
			"error":   err.Error(),
		})
		return
	}

	// Calculate key financial ratios
	currentAssets := c.getCategoryTotal(balanceSheet.Assets, models.CategoryCurrentAsset)
	currentLiabilities := c.getCategoryTotal(balanceSheet.Liabilities, models.CategoryCurrentLiability)
	
	// Build comprehensive dashboard
	dashboard := gin.H{
		"period": gin.H{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		},
		"financial_overview": gin.H{
			"total_revenue":    pnl.TotalRevenue,
			"total_expenses":   pnl.TotalExpenses,
			"gross_profit":     pnl.GrossProfit,
			"net_income":       pnl.NetIncome,
			"total_assets":     balanceSheet.TotalAssets,
			"total_liabilities": balanceSheet.TotalLiabilities,
			"total_equity":     balanceSheet.TotalEquity,
			"cash_position":    cashFlow.EndingCash,
			"is_balanced":      balanceSheet.IsBalanced,
		},
		"profitability_metrics": gin.H{
			"gross_profit_margin": c.safeDiv(pnl.GrossProfit, pnl.TotalRevenue) * 100,
			"net_profit_margin":   c.safeDiv(pnl.NetIncome, pnl.TotalRevenue) * 100,
			"return_on_assets":    c.safeDiv(pnl.NetIncome, balanceSheet.TotalAssets) * 100,
			"return_on_equity":    c.safeDiv(pnl.NetIncome, balanceSheet.TotalEquity) * 100,
		},
		"liquidity_ratios": gin.H{
			"current_ratio":       c.safeDiv(currentAssets, currentLiabilities),
			"quick_ratio":         c.safeDiv(currentAssets-c.getInventoryBalance(balanceSheet.Assets), currentLiabilities),
			"cash_ratio":          c.safeDiv(cashFlow.EndingCash, currentLiabilities),
			"working_capital":     currentAssets - currentLiabilities,
		},
		"leverage_ratios": gin.H{
			"debt_to_assets":      c.safeDiv(balanceSheet.TotalLiabilities, balanceSheet.TotalAssets) * 100,
			"debt_to_equity":      c.safeDiv(balanceSheet.TotalLiabilities, balanceSheet.TotalEquity),
			"equity_multiplier":   c.safeDiv(balanceSheet.TotalAssets, balanceSheet.TotalEquity),
		},
		"cash_flow_summary": gin.H{
			"beginning_cash":       cashFlow.BeginningCash,
			"ending_cash":         cashFlow.EndingCash,
			"net_cash_flow":       cashFlow.NetCashFlow,
			"operating_cash_flow": cashFlow.OperatingActivities.Total,
			"investing_cash_flow": cashFlow.InvestingActivities.Total,
			"financing_cash_flow": cashFlow.FinancingActivities.Total,
		},
		"sales_performance": gin.H{
			"total_revenue":       salesSummary.TotalRevenue,
			"total_transactions":  salesSummary.TotalTransactions,
			"average_order_value": salesSummary.AverageOrderValue,
			"collection_rate":     c.safeDiv(salesSummary.TotalPaidAmount, salesSummary.TotalRevenue) * 100,
			"outstanding_amount":  salesSummary.TotalOutstanding,
			"growth_rate":         salesSummary.GrowthAnalysis.GrowthRate,
			"top_customers":       salesSummary.TopCustomers,
			"top_products":        salesSummary.TopProducts,
		},
		"vendor_performance": gin.H{
			"total_purchases":        vendorAnalysis.TotalPurchases,
			"total_vendors":          vendorAnalysis.TotalVendors,
			"average_purchase_value": vendorAnalysis.AveragePurchaseValue,
			"payment_performance":    vendorAnalysis.PaymentPerformance,
			"top_vendors":            vendorAnalysis.VendorPerformances[:c.minInt(len(vendorAnalysis.VendorPerformances), 5)],
		},
		"generated_at": time.Now(),
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Financial Dashboard generated successfully",
		"data":    dashboard,
	})
}

// ========================= REPORT METADATA =========================

// GetAvailableReports returns metadata about all available reports
func (c *UnifiedFinancialReportController) GetAvailableReports(ctx *gin.Context) {
	reports := []gin.H{
		{
			"id":          "profit_loss_statement",
			"name":        "Profit & Loss Statement",
			"type":        "FINANCIAL",
			"description": "Comprehensive income statement showing revenues, cost of goods sold, expenses, and net income",
			"category":    "FINANCIAL_STATEMENTS",
			"endpoint":    "/api/unified-reports/profit-loss",
			"method":      "GET",
			"parameters": gin.H{
				"start_date":   "string (required, format: YYYY-MM-DD)",
				"end_date":     "string (required, format: YYYY-MM-DD)",
				"comparative":  "boolean (optional, default: false)",
			},
			"supports_comparative": true,
			"supports_export":      true,
		},
		{
			"id":          "balance_sheet",
			"name":        "Balance Sheet",
			"type":        "FINANCIAL",
			"description": "Statement of financial position showing assets, liabilities, and equity",
			"category":    "FINANCIAL_STATEMENTS",
			"endpoint":    "/api/unified-reports/balance-sheet",
			"method":      "GET",
			"parameters": gin.H{
				"as_of_date":   "string (optional, format: YYYY-MM-DD, default: today)",
				"comparative":  "boolean (optional, default: false)",
			},
			"supports_comparative": true,
			"supports_export":      true,
		},
		{
			"id":          "cash_flow_statement",
			"name":        "Cash Flow Statement",
			"type":        "FINANCIAL",
			"description": "Statement showing cash flows from operating, investing, and financing activities",
			"category":    "FINANCIAL_STATEMENTS",
			"endpoint":    "/api/unified-reports/cash-flow",
			"method":      "GET",
			"parameters": gin.H{
				"start_date": "string (required, format: YYYY-MM-DD)",
				"end_date":   "string (required, format: YYYY-MM-DD)",
			},
			"supports_comparative": false,
			"supports_export":      true,
		},
		{
			"id":          "trial_balance",
			"name":        "Trial Balance",
			"type":        "FINANCIAL",
			"description": "Summary of all account balances to verify that debits equal credits",
			"category":    "ACCOUNTING_REPORTS",
			"endpoint":    "/api/unified-reports/trial-balance",
			"method":      "GET",
			"parameters": gin.H{
				"as_of_date": "string (optional, format: YYYY-MM-DD, default: today)",
				"show_zero":  "boolean (optional, default: false)",
			},
			"supports_comparative": false,
			"supports_export":      true,
		},
		{
			"id":          "general_ledger",
			"name":        "General Ledger",
			"type":        "ACCOUNTING",
			"description": "Detailed record of all transactions for a specific account",
			"category":    "ACCOUNTING_REPORTS",
			"endpoint":    "/api/unified-reports/general-ledger/{account_id}",
			"method":      "GET",
			"parameters": gin.H{
				"account_id": "uint (required, path parameter)",
				"start_date": "string (required, format: YYYY-MM-DD)",
				"end_date":   "string (required, format: YYYY-MM-DD)",
			},
			"supports_comparative": false,
			"supports_export":      true,
		},
		{
			"id":          "sales_summary_report",
			"name":        "Sales Summary Report",
			"type":        "OPERATIONAL",
			"description": "Comprehensive sales analytics with customer and product performance",
			"category":    "SALES_REPORTS",
			"endpoint":    "/api/unified-reports/sales-summary",
			"method":      "GET",
			"parameters": gin.H{
				"start_date": "string (required, format: YYYY-MM-DD)",
				"end_date":   "string (required, format: YYYY-MM-DD)",
			},
			"supports_comparative": false,
			"supports_export":      true,
		},
		{
			"id":          "vendor_analysis_report",
			"name":        "Vendor Analysis Report",
			"type":        "OPERATIONAL",
			"description": "Comprehensive vendor performance analysis with payment tracking",
			"category":    "PURCHASE_REPORTS",
			"endpoint":    "/api/unified-reports/vendor-analysis",
			"method":      "GET",
			"parameters": gin.H{
				"start_date": "string (required, format: YYYY-MM-DD)",
				"end_date":   "string (required, format: YYYY-MM-DD)",
			},
			"supports_comparative": false,
			"supports_export":      true,
		},
		{
			"id":          "financial_dashboard",
			"name":        "Financial Dashboard",
			"type":        "DASHBOARD",
			"description": "Comprehensive financial overview with key metrics, ratios, and performance indicators",
			"category":    "DASHBOARDS",
			"endpoint":    "/api/unified-reports/dashboard",
			"method":      "GET",
			"parameters": gin.H{
				"start_date": "string (optional, format: YYYY-MM-DD, default: first day of current month)",
				"end_date":   "string (optional, format: YYYY-MM-DD, default: today)",
			},
			"supports_comparative": false,
			"supports_export":      false,
		},
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Available reports retrieved successfully",
		"data":    reports,
	})
}

// ========================= BATCH REPORT GENERATION =========================

// GenerateAllReports generates all financial reports for a given period
func (c *UnifiedFinancialReportController) GenerateAllReports(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	asOfDateStr := ctx.DefaultQuery("as_of_date", time.Now().Format("2006-01-02"))

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	asOfDate, err := time.Parse("2006-01-02", asOfDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid as_of_date format. Use YYYY-MM-DD",
		})
		return
	}

	// Generate all reports concurrently
	type reportResult struct {
		name string
		data interface{}
		err  error
	}

	results := make(chan reportResult, 7)

	// Generate P&L
	go func() {
		pnl, err := c.unifiedReportService.GenerateComprehensiveProfitLoss(ctx.Request.Context(), startDate, endDate, false)
		results <- reportResult{"profit_loss", pnl, err}
	}()

	// Generate Balance Sheet
	go func() {
		bs, err := c.unifiedReportService.GenerateComprehensiveBalanceSheet(ctx.Request.Context(), asOfDate, false)
		results <- reportResult{"balance_sheet", bs, err}
	}()

	// Generate Cash Flow
	go func() {
		cf, err := c.unifiedReportService.GenerateComprehensiveCashFlow(ctx.Request.Context(), startDate, endDate)
		results <- reportResult{"cash_flow", cf, err}
	}()

	// Generate Trial Balance
	go func() {
		tb, err := c.unifiedReportService.GenerateComprehensiveTrialBalance(ctx.Request.Context(), asOfDate, false)
		results <- reportResult{"trial_balance", tb, err}
	}()

	// Generate Sales Summary
	go func() {
		ss, err := c.unifiedReportService.GenerateComprehensiveSalesSummary(ctx.Request.Context(), startDate, endDate)
		results <- reportResult{"sales_summary", ss, err}
	}()

	// Generate Vendor Analysis
	go func() {
		va, err := c.unifiedReportService.GenerateComprehensiveVendorAnalysis(ctx.Request.Context(), startDate, endDate)
		results <- reportResult{"vendor_analysis", va, err}
	}()

	// Collect results
	allReports := make(map[string]interface{})
	var errorList []string

	for i := 0; i < 6; i++ {
		result := <-results
		if result.err != nil {
			errorList = append(errorList, fmt.Sprintf("%s: %s", result.name, result.err.Error()))
		} else {
			allReports[result.name] = result.data
		}
	}

	if len(errorList) > 0 {
		ctx.JSON(http.StatusPartialContent, gin.H{
			"status":  "partial_success",
			"message": "Some reports failed to generate",
			"data":    allReports,
			"errors":  errorList,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "All financial reports generated successfully",
		"data":    allReports,
	})
}

// ========================= HELPER METHODS =========================

// getCategoryTotal gets total balance for a specific category in balance sheet section
func (c *UnifiedFinancialReportController) getCategoryTotal(section models.BalanceSheetSection, category string) float64 {
	for _, cat := range section.Categories {
		if cat.Name == category {
			return cat.Total
		}
	}
	return 0
}

// getInventoryBalance gets inventory balance from assets section
func (c *UnifiedFinancialReportController) getInventoryBalance(assets models.BalanceSheetSection) float64 {
	// Look for inventory accounts in current assets
	for _, category := range assets.Categories {
		if category.Name == models.CategoryCurrentAsset {
			for _, account := range category.Accounts {
				if account.AccountCode[:2] == "13" { // Inventory accounts typically start with 13
					return account.Balance
				}
			}
		}
	}
	return 0
}

// safeDiv performs safe division avoiding division by zero
func (c *UnifiedFinancialReportController) safeDiv(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

// minInt returns the minimum of two integers
func (c *UnifiedFinancialReportController) minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ========================= VALIDATION HELPERS =========================

// ValidateReportParameters validates common report parameters
func (c *UnifiedFinancialReportController) ValidateReportParameters(ctx *gin.Context) {
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	asOfDateStr := ctx.Query("as_of_date")

	validationResult := gin.H{
		"valid":    true,
		"errors":   []string{},
		"warnings": []string{},
	}

	errors := []string{}
	warnings := []string{}

	// Validate start and end dates if provided
	if startDateStr != "" && endDateStr != "" {
		startDate, err1 := time.Parse("2006-01-02", startDateStr)
		endDate, err2 := time.Parse("2006-01-02", endDateStr)

		if err1 != nil {
			errors = append(errors, "Invalid start_date format. Use YYYY-MM-DD")
		}
		if err2 != nil {
			errors = append(errors, "Invalid end_date format. Use YYYY-MM-DD")
		}

		if err1 == nil && err2 == nil {
			if endDate.Before(startDate) {
				errors = append(errors, "End date must be after start date")
			}

			// Check for very long periods
			duration := endDate.Sub(startDate)
			if duration > 366*24*time.Hour {
				warnings = append(warnings, "Report period longer than one year may affect performance")
			}

			// Check for future dates
			if endDate.After(time.Now()) {
				warnings = append(warnings, "End date is in the future")
			}
		}
	}

	// Validate as_of_date if provided
	if asOfDateStr != "" {
		_, err := time.Parse("2006-01-02", asOfDateStr)
		if err != nil {
			errors = append(errors, "Invalid as_of_date format. Use YYYY-MM-DD")
		}
	}

	validationResult["errors"] = errors
	validationResult["warnings"] = warnings
	validationResult["valid"] = len(errors) == 0

	if len(errors) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"data":   validationResult,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   validationResult,
	})
}
