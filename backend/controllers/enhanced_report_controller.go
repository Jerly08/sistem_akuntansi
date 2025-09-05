package controllers

import (
	"net/http"
	"time"

	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

// EnhancedReportController handles comprehensive financial and operational reporting endpoints
type EnhancedReportController struct {
	enhancedReportService *services.EnhancedReportService
	professionalService   *services.ProfessionalReportService
	standardizedService   *services.StandardizedReportService
}

// NewEnhancedReportController creates a new enhanced report controller
func NewEnhancedReportController(
	enhancedReportService *services.EnhancedReportService,
	professionalService *services.ProfessionalReportService,
	standardizedService *services.StandardizedReportService,
) *EnhancedReportController {
	return &EnhancedReportController{
		enhancedReportService: enhancedReportService,
		professionalService:   professionalService,
		standardizedService:   standardizedService,
	}
}

// GetComprehensiveBalanceSheet generates a comprehensive balance sheet with proper accounting logic
func (erc *EnhancedReportController) GetComprehensiveBalanceSheet(c *gin.Context) {
	asOfDate := c.Query("as_of_date")
	format := c.DefaultQuery("format", "json")

	var date time.Time
	var err error
	if asOfDate == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", asOfDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid date format. Use YYYY-MM-DD",
				"error":   err.Error(),
			})
			return
		}
	}

	// Generate balance sheet data
	balanceSheetData, err := erc.enhancedReportService.GenerateBalanceSheet(date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate balance sheet",
			"error":   err.Error(),
		})
		return
	}

	// Handle different output formats
	switch format {
	case "json":
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   balanceSheetData,
		})
	case "pdf":
		// Use professional service for PDF generation
		pdfData, err := erc.professionalService.GenerateBalanceSheetPDF(date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate PDF report",
				"error":   err.Error(),
			})
			return
		}
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=comprehensive_balance_sheet.pdf")
		c.Data(http.StatusOK, "application/pdf", pdfData)
	case "excel":
		// Use professional service for Excel generation
		excelData, err := erc.professionalService.GenerateBalanceSheetExcel(date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate Excel report",
				"error":   err.Error(),
			})
			return
		}
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=comprehensive_balance_sheet.xlsx")
		c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelData)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Unsupported format. Use json, pdf, or excel",
		})
	}
}

// GetComprehensiveProfitLoss generates a comprehensive P&L statement with proper accounting logic
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

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	// Generate P&L data
	profitLossData, err := erc.enhancedReportService.GenerateProfitLoss(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate profit & loss statement",
			"error":   err.Error(),
		})
		return
	}

	// Handle different output formats
	switch format {
	case "json":
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   profitLossData,
		})
	case "pdf":
		// Use professional service for PDF generation
		pdfData, err := erc.professionalService.GenerateProfitLossPDF(start, end)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate PDF report",
				"error":   err.Error(),
			})
			return
		}
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=comprehensive_profit_loss.pdf")
		c.Data(http.StatusOK, "application/pdf", pdfData)
	case "excel":
		// Use professional service for Excel generation
		excelData, err := erc.professionalService.GenerateProfitLossExcel(start, end)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate Excel report",
				"error":   err.Error(),
			})
			return
		}
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=comprehensive_profit_loss.xlsx")
		c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelData)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Unsupported format. Use json, pdf, or excel",
		})
	}
}

// GetComprehensiveCashFlow generates a comprehensive cash flow statement
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

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	// Generate cash flow data
	cashFlowData, err := erc.enhancedReportService.GenerateCashFlow(start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate cash flow statement",
			"error":   err.Error(),
		})
		return
	}

	// Handle different output formats
	switch format {
	case "json":
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   cashFlowData,
		})
	case "pdf":
		// Use professional service for PDF generation
		pdfData, err := erc.professionalService.GenerateCashFlowStatementPDF(start, end)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate PDF report",
				"error":   err.Error(),
			})
			return
		}
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=comprehensive_cash_flow.pdf")
		c.Data(http.StatusOK, "application/pdf", pdfData)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Unsupported format for cash flow. Use json or pdf",
		})
	}
}

// GetComprehensiveSalesSummary generates comprehensive sales analytics
func (erc *EnhancedReportController) GetComprehensiveSalesSummary(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "month")
	format := c.DefaultQuery("format", "json")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	// Validate groupBy parameter
	validGroupBy := map[string]bool{
		"day":     true,
		"week":    true,
		"month":   true,
		"quarter": true,
		"year":    true,
	}
	if !validGroupBy[groupBy] {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid group_by parameter. Use day, week, month, quarter, or year",
		})
		return
	}

	// Generate sales summary data
	salesSummary, err := erc.enhancedReportService.GenerateSalesSummary(start, end, groupBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate sales summary",
			"error":   err.Error(),
		})
		return
	}

	// Handle different output formats
	switch format {
	case "json":
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   salesSummary,
		})
	case "pdf":
		// Use professional service for PDF generation
		pdfData, err := erc.professionalService.GenerateSalesSummaryPDF(start, end, groupBy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate PDF report",
				"error":   err.Error(),
			})
			return
		}
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=comprehensive_sales_summary.pdf")
		c.Data(http.StatusOK, "application/pdf", pdfData)
	case "excel":
		// Use professional service for Excel generation
		excelData, err := erc.professionalService.GenerateSalesSummaryExcel(start, end, groupBy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate Excel report",
				"error":   err.Error(),
			})
			return
		}
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=comprehensive_sales_summary.xlsx")
		c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", excelData)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Unsupported format. Use json, pdf, or excel",
		})
	}
}

// GetComprehensivePurchaseSummary generates comprehensive purchase analytics
func (erc *EnhancedReportController) GetComprehensivePurchaseSummary(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "month")
	format := c.DefaultQuery("format", "json")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	// Validate groupBy parameter
	validGroupBy := map[string]bool{
		"day":     true,
		"week":    true,
		"month":   true,
		"quarter": true,
		"year":    true,
	}
	if !validGroupBy[groupBy] {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Invalid group_by parameter. Use day, week, month, quarter, or year",
		})
		return
	}

	// Generate purchase summary data
	purchaseSummary, err := erc.enhancedReportService.GeneratePurchaseSummary(start, end, groupBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate purchase summary",
			"error":   err.Error(),
		})
		return
	}

	// Handle different output formats
	switch format {
	case "json":
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   purchaseSummary,
		})
	case "pdf":
		// Use professional service for PDF generation
		pdfData, err := erc.professionalService.GeneratePurchaseSummaryPDF(start, end, groupBy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate PDF report",
				"error":   err.Error(),
			})
			return
		}
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=comprehensive_purchase_summary.pdf")
		c.Data(http.StatusOK, "application/pdf", pdfData)
	case "excel":
		// Use professional service for Excel generation (would need to implement this method)
		c.JSON(http.StatusNotImplemented, gin.H{
			"status":  "error",
			"message": "Excel format for purchase summary not yet implemented",
		})
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Unsupported format. Use json or pdf",
		})
	}
}

// GetFinancialDashboard provides a comprehensive financial dashboard with key metrics
func (erc *EnhancedReportController) GetFinancialDashboard(c *gin.Context) {
	// Get date parameters, default to current month
	endDate := time.Now()
	startDate := time.Date(endDate.Year(), endDate.Month(), 1, 0, 0, 0, 0, endDate.Location())

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	// Generate all reports for dashboard
	balanceSheet, err := erc.enhancedReportService.GenerateBalanceSheet(endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate balance sheet for dashboard",
			"error":   err.Error(),
		})
		return
	}

	profitLoss, err := erc.enhancedReportService.GenerateProfitLoss(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate P&L for dashboard",
			"error":   err.Error(),
		})
		return
	}

	cashFlow, err := erc.enhancedReportService.GenerateCashFlow(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate cash flow for dashboard",
			"error":   err.Error(),
		})
		return
	}

	salesSummary, err := erc.enhancedReportService.GenerateSalesSummary(startDate, endDate, "month")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate sales summary for dashboard",
			"error":   err.Error(),
		})
		return
	}

	purchaseSummary, err := erc.enhancedReportService.GeneratePurchaseSummary(startDate, endDate, "month")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "Failed to generate purchase summary for dashboard",
			"error":   err.Error(),
		})
		return
	}

	// Compile dashboard data
	dashboard := gin.H{
		"period": gin.H{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		},
		"balance_sheet": gin.H{
			"total_assets":            balanceSheet.TotalAssets,
			"total_liabilities":       balanceSheet.Liabilities.Total,
			"total_equity":            balanceSheet.Equity.Total,
			"is_balanced":             balanceSheet.IsBalanced,
			"current_assets":          erc.getSubtotalByCategory(balanceSheet.Assets.Subtotals, "CURRENT_ASSET"),
			"fixed_assets":            erc.getSubtotalByCategory(balanceSheet.Assets.Subtotals, "FIXED_ASSET"),
			"current_liabilities":     erc.getSubtotalByCategory(balanceSheet.Liabilities.Subtotals, "CURRENT_LIABILITY"),
			"long_term_liabilities":   erc.getSubtotalByCategory(balanceSheet.Liabilities.Subtotals, "LONG_TERM_LIABILITY"),
		},
		"profit_loss": gin.H{
			"total_revenue":       profitLoss.Revenue.Subtotal,
			"cost_of_goods_sold": profitLoss.CostOfGoodsSold.Subtotal,
			"gross_profit":       profitLoss.GrossProfit,
			"gross_profit_margin": profitLoss.GrossProfitMargin,
			"operating_expenses":  profitLoss.OperatingExpenses.Subtotal,
			"operating_income":    profitLoss.OperatingIncome,
			"net_income":         profitLoss.NetIncome,
			"net_income_margin":   profitLoss.NetIncomeMargin,
		},
		"cash_flow": gin.H{
			"beginning_cash":       cashFlow.BeginningCash,
			"ending_cash":         cashFlow.EndingCash,
			"net_cash_flow":       cashFlow.NetCashFlow,
			"operating_cash_flow": cashFlow.OperatingActivities.Total,
			"investing_cash_flow": cashFlow.InvestingActivities.Total,
			"financing_cash_flow": cashFlow.FinancingActivities.Total,
		},
		"sales_summary": gin.H{
			"total_revenue":      salesSummary.TotalRevenue,
			"total_transactions": salesSummary.TotalTransactions,
			"average_order_value": salesSummary.AverageOrderValue,
			"total_customers":     salesSummary.TotalCustomers,
			"growth_analysis":     salesSummary.GrowthAnalysis,
		},
		"purchase_summary": gin.H{
			"total_purchases":        purchaseSummary.TotalPurchases,
			"total_transactions":     purchaseSummary.TotalTransactions,
			"average_purchase_value": purchaseSummary.AveragePurchaseValue,
			"total_vendors":          purchaseSummary.TotalVendors,
			"cost_analysis":          purchaseSummary.CostAnalysis,
		},
		"key_ratios": gin.H{
			"current_ratio":     erc.calculateCurrentRatio(balanceSheet),
			"debt_to_equity":    erc.calculateDebtToEquityRatio(balanceSheet),
			"return_on_assets":  erc.calculateROA(profitLoss, balanceSheet),
			"return_on_equity":  erc.calculateROE(profitLoss, balanceSheet),
		},
		"generated_at": time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   dashboard,
	})
}

// GetAvailableReports returns metadata about all available reports
func (erc *EnhancedReportController) GetAvailableReports(c *gin.Context) {
	reports := []gin.H{
		{
			"id":          "comprehensive_balance_sheet",
			"name":        "Comprehensive Balance Sheet",
			"type":        "FINANCIAL",
			"description": "Detailed balance sheet with proper accounting logic showing assets, liabilities, and equity",
			"category":    "FINANCIAL_STATEMENTS",
			"required_params": []string{"as_of_date"},
			"optional_params": []string{"format"},
			"supported_formats": []string{"json", "pdf", "excel"},
			"endpoint":    "/api/reports/comprehensive/balance-sheet",
		},
		{
			"id":          "comprehensive_profit_loss",
			"name":        "Comprehensive Profit & Loss Statement",
			"type":        "FINANCIAL",
			"description": "Detailed P&L statement with revenue, expenses, gross profit, and net income analysis",
			"category":    "FINANCIAL_STATEMENTS",
			"required_params": []string{"start_date", "end_date"},
			"optional_params": []string{"format"},
			"supported_formats": []string{"json", "pdf", "excel"},
			"endpoint":    "/api/reports/comprehensive/profit-loss",
		},
		{
			"id":          "comprehensive_cash_flow",
			"name":        "Comprehensive Cash Flow Statement",
			"type":        "FINANCIAL",
			"description": "Cash flow statement with operating, investing, and financing activities",
			"category":    "FINANCIAL_STATEMENTS",
			"required_params": []string{"start_date", "end_date"},
			"optional_params": []string{"format"},
			"supported_formats": []string{"json", "pdf"},
			"endpoint":    "/api/reports/comprehensive/cash-flow",
		},
		{
			"id":          "comprehensive_sales_summary",
			"name":        "Comprehensive Sales Summary Report",
			"type":        "OPERATIONAL",
			"description": "Detailed sales analytics with customer, product, and period analysis",
			"category":    "SALES_ANALYTICS",
			"required_params": []string{"start_date", "end_date"},
			"optional_params": []string{"group_by", "format"},
			"supported_formats": []string{"json", "pdf", "excel"},
			"endpoint":    "/api/reports/comprehensive/sales-summary",
		},
		{
			"id":          "comprehensive_purchase_summary",
			"name":        "Comprehensive Purchase Summary Report",
			"type":        "OPERATIONAL",
			"description": "Detailed purchase analytics with vendor, category, and cost analysis",
			"category":    "PURCHASE_ANALYTICS",
			"required_params": []string{"start_date", "end_date"},
			"optional_params": []string{"group_by", "format"},
			"supported_formats": []string{"json", "pdf"},
			"endpoint":    "/api/reports/comprehensive/purchase-summary",
		},
		{
			"id":          "financial_dashboard",
			"name":        "Financial Dashboard",
			"type":        "DASHBOARD",
			"description": "Comprehensive financial dashboard with key metrics and ratios",
			"category":    "DASHBOARDS",
			"required_params": []string{},
			"optional_params": []string{"start_date", "end_date"},
			"supported_formats": []string{"json"},
			"endpoint":    "/api/reports/financial-dashboard",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   reports,
	})
}

// Helper methods for dashboard calculations

func (erc *EnhancedReportController) getSubtotalByCategory(subtotals []services.BalanceSheetSubtotal, category string) float64 {
	for _, subtotal := range subtotals {
		if subtotal.Category == category {
			return subtotal.Amount
		}
	}
	return 0
}

func (erc *EnhancedReportController) calculateCurrentRatio(balanceSheet *services.BalanceSheetData) float64 {
	currentAssets := erc.getSubtotalByCategory(balanceSheet.Assets.Subtotals, "CURRENT_ASSET")
	currentLiabilities := erc.getSubtotalByCategory(balanceSheet.Liabilities.Subtotals, "CURRENT_LIABILITY")
	
	if currentLiabilities == 0 {
		return 0
	}
	return currentAssets / currentLiabilities
}

func (erc *EnhancedReportController) calculateDebtToEquityRatio(balanceSheet *services.BalanceSheetData) float64 {
	totalLiabilities := balanceSheet.Liabilities.Total
	totalEquity := balanceSheet.Equity.Total
	
	if totalEquity == 0 {
		return 0
	}
	return totalLiabilities / totalEquity
}

func (erc *EnhancedReportController) calculateROA(profitLoss *services.ProfitLossData, balanceSheet *services.BalanceSheetData) float64 {
	netIncome := profitLoss.NetIncome
	totalAssets := balanceSheet.TotalAssets
	
	if totalAssets == 0 {
		return 0
	}
	return (netIncome / totalAssets) * 100
}

func (erc *EnhancedReportController) calculateROE(profitLoss *services.ProfitLossData, balanceSheet *services.BalanceSheetData) float64 {
	netIncome := profitLoss.NetIncome
	totalEquity := balanceSheet.Equity.Total
	
	if totalEquity == 0 {
		return 0
	}
	return (netIncome / totalEquity) * 100
}

// GetReportPreview generates a lightweight preview of reports for quick viewing
func (erc *EnhancedReportController) GetReportPreview(c *gin.Context) {
	reportType := c.Param("type")
	
	// Extract common parameters
	asOfDate := c.DefaultQuery("as_of_date", time.Now().Format("2006-01-02"))
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "month")
	
	switch reportType {
	case "balance-sheet":
		// Parse as of date
		date, err := time.Parse("2006-01-02", asOfDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid as_of_date format. Use YYYY-MM-DD",
			})
			return
		}
		
		// Generate limited balance sheet data for preview
		balanceSheetData, err := erc.enhancedReportService.GenerateBalanceSheet(date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate balance sheet preview",
				"error":   err.Error(),
			})
			return
		}
		
		// Return limited data for preview
		previewData := gin.H{
			"company":     balanceSheetData.Company,
			"as_of_date":  balanceSheetData.AsOfDate,
			"assets":      gin.H{
				"items": balanceSheetData.Assets.Items[:min(len(balanceSheetData.Assets.Items), 10)], // Limit to 10 items
				"total": balanceSheetData.Assets.Total,
			},
			"liabilities": gin.H{
				"items": balanceSheetData.Liabilities.Items[:min(len(balanceSheetData.Liabilities.Items), 10)],
				"total": balanceSheetData.Liabilities.Total,
			},
			"equity":      gin.H{
				"items": balanceSheetData.Equity.Items[:min(len(balanceSheetData.Equity.Items), 10)],
				"total": balanceSheetData.Equity.Total,
			},
			"is_balanced": balanceSheetData.IsBalanced,
		}
		
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   previewData,
			"meta":   gin.H{"is_preview": true},
		})
		
	case "profit-loss":
		if startDate == "" || endDate == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "start_date and end_date are required for profit-loss preview",
			})
			return
		}
		
		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid start_date format. Use YYYY-MM-DD",
			})
			return
		}
		
		end, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid end_date format. Use YYYY-MM-DD",
			})
			return
		}
		
		// Generate P&L data
		profitLossData, err := erc.enhancedReportService.GenerateProfitLoss(start, end)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate profit loss preview",
				"error":   err.Error(),
			})
			return
		}
		
		// Return limited data for preview
		previewData := gin.H{
			"company":     profitLossData.Company,
			"start_date":  profitLossData.StartDate,
			"end_date":    profitLossData.EndDate,
			"revenue":     gin.H{
				"items":    profitLossData.Revenue.Items[:min(len(profitLossData.Revenue.Items), 5)],
				"subtotal": profitLossData.Revenue.Subtotal,
			},
			"cost_of_goods_sold": gin.H{
				"items":    profitLossData.CostOfGoodsSold.Items[:min(len(profitLossData.CostOfGoodsSold.Items), 5)],
				"subtotal": profitLossData.CostOfGoodsSold.Subtotal,
			},
			"operating_expenses": gin.H{
				"items":    profitLossData.OperatingExpenses.Items[:min(len(profitLossData.OperatingExpenses.Items), 5)],
				"subtotal": profitLossData.OperatingExpenses.Subtotal,
			},
			"net_income":      profitLossData.NetIncome,
			"gross_profit":    profitLossData.GrossProfit,
			"operating_income": profitLossData.OperatingIncome,
		}
		
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   previewData,
			"meta":   gin.H{"is_preview": true},
		})
		
	case "cash-flow":
		if startDate == "" || endDate == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "start_date and end_date are required for cash-flow preview",
			})
			return
		}
		
		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid start_date format. Use YYYY-MM-DD",
			})
			return
		}
		
		end, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid end_date format. Use YYYY-MM-DD",
			})
			return
		}
		
		// Generate cash flow data
		cashFlowData, err := erc.enhancedReportService.GenerateCashFlow(start, end)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate cash flow preview",
				"error":   err.Error(),
			})
			return
		}
		
		// Return limited data for preview
		previewData := gin.H{
			"company":    cashFlowData.Company,
			"start_date": cashFlowData.StartDate,
			"end_date":   cashFlowData.EndDate,
			"operating_activities": gin.H{
				"items": cashFlowData.OperatingActivities.Items[:min(len(cashFlowData.OperatingActivities.Items), 5)],
				"total": cashFlowData.OperatingActivities.Total,
			},
			"investing_activities": gin.H{
				"items": cashFlowData.InvestingActivities.Items[:min(len(cashFlowData.InvestingActivities.Items), 5)],
				"total": cashFlowData.InvestingActivities.Total,
			},
			"financing_activities": gin.H{
				"items": cashFlowData.FinancingActivities.Items[:min(len(cashFlowData.FinancingActivities.Items), 5)],
				"total": cashFlowData.FinancingActivities.Total,
			},
			"net_cash_flow": cashFlowData.NetCashFlow,
		}
		
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   previewData,
			"meta":   gin.H{"is_preview": true},
		})
		
	case "sales-summary":
		if startDate == "" || endDate == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "start_date and end_date are required for sales-summary preview",
			})
			return
		}
		
		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid start_date format. Use YYYY-MM-DD",
			})
			return
		}
		
		end, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid end_date format. Use YYYY-MM-DD",
			})
			return
		}
		
		// Generate sales summary data
		salesSummaryData, err := erc.enhancedReportService.GenerateSalesSummary(start, end, groupBy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate sales summary preview",
				"error":   err.Error(),
			})
			return
		}
		
		// Return limited data for preview
		previewData := gin.H{
			"company":          salesSummaryData.Company,
			"start_date":       salesSummaryData.StartDate,
			"end_date":         salesSummaryData.EndDate,
			"total_revenue":    salesSummaryData.TotalRevenue,
			"total_transactions": salesSummaryData.TotalTransactions,
			"sales_by_period":  salesSummaryData.SalesByPeriod[:min(len(salesSummaryData.SalesByPeriod), 10)],
		}
		
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   previewData,
			"meta":   gin.H{"is_preview": true},
		})
		
	case "purchase-summary", "vendor-analysis":
		if startDate == "" || endDate == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "start_date and end_date are required for purchase summary preview",
			})
			return
		}
		
		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid start_date format. Use YYYY-MM-DD",
			})
			return
		}
		
		end, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  "error",
				"message": "Invalid end_date format. Use YYYY-MM-DD",
			})
			return
		}
		
		// Generate purchase summary data
		purchaseSummaryData, err := erc.enhancedReportService.GeneratePurchaseSummary(start, end, groupBy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate purchase summary preview",
				"error":   err.Error(),
			})
			return
		}
		
		// Return limited data for preview
		previewData := gin.H{
			"company":            purchaseSummaryData.Company,
			"start_date":         purchaseSummaryData.StartDate,
			"end_date":           purchaseSummaryData.EndDate,
			"total_purchases":    purchaseSummaryData.TotalPurchases,
			"total_transactions": purchaseSummaryData.TotalTransactions,
			"purchases_by_period": purchaseSummaryData.PurchasesByPeriod[:min(len(purchaseSummaryData.PurchasesByPeriod), 10)],
		}
		
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   previewData,
			"meta":   gin.H{"is_preview": true},
		})
		
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Unsupported report type for preview",
		})
	}
}

// Helper function to get minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
