package controllers

import (
	"net/http"
	"strconv"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

type FinancialReportController struct {
	financialReportService services.FinancialReportService
}

func NewFinancialReportController(financialReportService services.FinancialReportService) *FinancialReportController {
	return &FinancialReportController{
		financialReportService: financialReportService,
	}
}

// GenerateProfitLossStatement generates a Profit & Loss statement
func (c *FinancialReportController) GenerateProfitLossStatement(ctx *gin.Context) {
	var req models.FinancialReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate dates
	if req.EndDate.Before(req.StartDate) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "End date must be after start date"})
		return
	}

	pnl, err := c.financialReportService.GenerateProfitLossStatement(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate P&L statement", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "P&L statement generated successfully", "data": pnl})
}

// GenerateBalanceSheet generates a Balance Sheet
func (c *FinancialReportController) GenerateBalanceSheet(ctx *gin.Context) {
	var req models.FinancialReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate dates
	if req.EndDate.Before(req.StartDate) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "End date must be after start date"})
		return
	}

	balanceSheet, err := c.financialReportService.GenerateBalanceSheet(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate balance sheet", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Balance sheet generated successfully", "data": balanceSheet})
}

// GenerateCashFlowStatement generates a Cash Flow Statement
func (c *FinancialReportController) GenerateCashFlowStatement(ctx *gin.Context) {
	var req models.FinancialReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate dates
	if req.EndDate.Before(req.StartDate) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "End date must be after start date"})
		return
	}

	cashFlow, err := c.financialReportService.GenerateCashFlowStatement(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate cash flow statement", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Cash flow statement generated successfully", "data": cashFlow})
}

// GenerateTrialBalance generates a Trial Balance
func (c *FinancialReportController) GenerateTrialBalance(ctx *gin.Context) {
	var req models.FinancialReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate dates
	if req.EndDate.Before(req.StartDate) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "End date must be after start date"})
		return
	}

	trialBalance, err := c.financialReportService.GenerateTrialBalance(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate trial balance", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Trial balance generated successfully", "data": trialBalance})
}

// GenerateGeneralLedger generates a General Ledger for a specific account
func (c *FinancialReportController) GenerateGeneralLedger(ctx *gin.Context) {
	// Get account ID from path
	accountIDStr := ctx.Param("account_id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid account ID", "details": err.Error()})
		return
	}

	// Get date parameters
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD", "details": err.Error()})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD", "details": err.Error()})
		return
	}

	// Validate dates
	if endDate.Before(startDate) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "End date must be after start date"})
		return
	}

	generalLedger, err := c.financialReportService.GenerateGeneralLedger(ctx, uint(accountID), startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate general ledger", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "General ledger generated successfully", "data": generalLedger})
}

// GetFinancialDashboard gets the financial dashboard
func (c *FinancialReportController) GetFinancialDashboard(ctx *gin.Context) {
	dashboard, err := c.financialReportService.GenerateFinancialDashboard(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate financial dashboard", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Financial dashboard retrieved successfully", "data": dashboard})
}

// GetRealTimeMetrics gets real-time financial metrics
func (c *FinancialReportController) GetRealTimeMetrics(ctx *gin.Context) {
	metrics, err := c.financialReportService.GetRealTimeMetrics(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get real-time metrics", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Real-time metrics retrieved successfully", "data": metrics})
}

// CalculateFinancialRatios calculates financial ratios for a given period
func (c *FinancialReportController) CalculateFinancialRatios(ctx *gin.Context) {
	// Get date parameters
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	if startDateStr == "" || endDateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "start_date and end_date are required"})
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD", "details": err.Error()})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD", "details": err.Error()})
		return
	}

	// Validate dates
	if endDate.Before(startDate) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "End date must be after start date"})
		return
	}

	ratios, err := c.financialReportService.CalculateFinancialRatios(ctx, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate financial ratios", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Financial ratios calculated successfully", "data": ratios})
}

// GetFinancialHealthScore gets the overall financial health score
func (c *FinancialReportController) GetFinancialHealthScore(ctx *gin.Context) {
	healthScore, err := c.financialReportService.CalculateFinancialHealthScore(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate financial health score", "details": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Financial health score calculated successfully", "data": healthScore})
}

// GetReportsList gets available reports with metadata
func (c *FinancialReportController) GetReportsList(ctx *gin.Context) {
	reports := []models.ReportMetadata{
		{
			ReportType:          models.ReportTypeProfitLoss,
			Name:                "Profit & Loss Statement",
			Description:         "Comprehensive income statement showing revenues, expenses, and net income for a specific period",
			SupportsComparative: true,
			RequiredParams:      []string{"start_date", "end_date"},
			OptionalParams:      []string{"comparative", "show_zero"},
		},
		{
			ReportType:          models.ReportTypeBalanceSheet,
			Name:                "Balance Sheet",
			Description:         "Statement of financial position showing assets, liabilities, and equity at a specific point in time",
			SupportsComparative: true,
			RequiredParams:      []string{"end_date"},
			OptionalParams:      []string{"start_date", "comparative", "show_zero"},
		},
		{
			ReportType:          models.ReportTypeCashFlow,
			Name:                "Cash Flow Statement",
			Description:         "Statement showing cash receipts and payments during a specific period using indirect method",
			SupportsComparative: false,
			RequiredParams:      []string{"start_date", "end_date"},
			OptionalParams:      []string{"show_zero"},
		},
		{
			ReportType:          models.ReportTypeTrialBalance,
			Name:                "Trial Balance",
			Description:         "Summary of all account balances to ensure debits equal credits",
			SupportsComparative: false,
			RequiredParams:      []string{"end_date"},
			OptionalParams:      []string{"show_zero"},
		},
		{
			ReportType:          models.ReportTypeGeneralLedger,
			Name:                "General Ledger",
			Description:         "Detailed record of all financial transactions for a specific account",
			SupportsComparative: false,
			RequiredParams:      []string{"account_id", "start_date", "end_date"},
			OptionalParams:      []string{},
		},
		{
			ReportType:          "DASHBOARD",
			Name:                "Financial Dashboard",
			Description:         "Comprehensive financial overview with key metrics, ratios, and alerts",
			SupportsComparative: false,
			RequiredParams:      []string{},
			OptionalParams:      []string{},
		},
		{
			ReportType:          "FINANCIAL_RATIOS",
			Name:                "Financial Ratios",
			Description:         "Detailed financial ratio analysis including liquidity, profitability, and efficiency ratios",
			SupportsComparative: false,
			RequiredParams:      []string{"start_date", "end_date"},
			OptionalParams:      []string{},
		},
		{
			ReportType:          "HEALTH_SCORE",
			Name:                "Financial Health Score",
			Description:         "Overall financial health assessment with scoring and recommendations",
			SupportsComparative: false,
			RequiredParams:      []string{},
			OptionalParams:      []string{},
		},
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Available reports retrieved successfully", "data": reports})
}

// GetReportFormats gets available export formats for reports
func (c *FinancialReportController) GetReportFormats(ctx *gin.Context) {
	formats := []string{"JSON", "PDF", "EXCEL", "CSV"}
	ctx.JSON(http.StatusOK, gin.H{"message": "Export formats retrieved successfully", "data": formats})
}

// GetQuickStats gets quick financial statistics for widgets
func (c *FinancialReportController) GetQuickStats(ctx *gin.Context) {
	// Get real-time metrics for quick stats
	metrics, err := c.financialReportService.GetRealTimeMetrics(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get quick stats", "details": err.Error()})
		return
	}

	// Convert to quick stats format
	quickStats := models.QuickFinancialStats{
		CashBalance:         metrics.CashPosition,
		TodayRevenue:        metrics.DailyRevenue,
		TodayExpenses:       metrics.DailyExpenses,
		MonthToDateRevenue:  metrics.MonthlyRevenue,
		MonthToDateExpenses: metrics.MonthlyExpenses,
		YearToDateRevenue:   metrics.YearlyRevenue,
		YearToDateExpenses:  metrics.YearlyExpenses,
		PendingReceivables:  metrics.PendingReceivables,
		PendingPayables:     metrics.PendingPayables,
		InventoryValue:      metrics.InventoryValue,
		LastUpdated:         metrics.LastUpdated,
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Quick financial statistics retrieved successfully", "data": quickStats})
}

// ValidateReportRequest validates common report request parameters
func (c *FinancialReportController) ValidateReportRequest(ctx *gin.Context) {
	var req models.FinancialReportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	validationResult := map[string]interface{}{
		"valid": true,
		"errors": []string{},
		"warnings": []string{},
	}

	errors := []string{}
	warnings := []string{}

	// Validate dates
	if !req.EndDate.IsZero() && !req.StartDate.IsZero() && req.EndDate.Before(req.StartDate) {
		errors = append(errors, "End date must be after start date")
	}

	// Check for future dates
	now := time.Now()
	if req.EndDate.After(now) {
		warnings = append(warnings, "End date is in the future")
	}

	// Check for very long periods
	if !req.StartDate.IsZero() && !req.EndDate.IsZero() {
		duration := req.EndDate.Sub(req.StartDate)
		if duration > 366*24*time.Hour { // More than a year
			warnings = append(warnings, "Report period is longer than one year, which may affect performance")
		}
	}

	// Validate report type
	validReportTypes := []string{
		models.ReportTypeProfitLoss,
		models.ReportTypeBalanceSheet,
		models.ReportTypeCashFlow,
		models.ReportTypeTrialBalance,
		models.ReportTypeGeneralLedger,
	}

	isValidType := false
	for _, validType := range validReportTypes {
		if req.ReportType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType && req.ReportType != "" {
		errors = append(errors, "Invalid report type")
	}

	validationResult["errors"] = errors
	validationResult["warnings"] = warnings
	validationResult["valid"] = len(errors) == 0

	// Always return the validation result, but with appropriate status code
	if len(errors) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "data": validationResult})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Request validation completed", "data": validationResult})
}

// GetReportSummary gets a summary of recent report generation activity
func (c *FinancialReportController) GetReportSummary(ctx *gin.Context) {
	limitStr := ctx.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// This is a simplified implementation
	// In a real system, you'd track report generation history
	summary := models.ReportGenerationSummary{
		TotalReportsGenerated: 0,
		RecentReports:        []models.ReportGenerationLog{},
		PopularReports: []models.ReportPopularity{
			{ReportType: models.ReportTypeProfitLoss, GenerationCount: 0},
			{ReportType: models.ReportTypeBalanceSheet, GenerationCount: 0},
			{ReportType: models.ReportTypeCashFlow, GenerationCount: 0},
		},
		LastGeneratedAt: time.Now(),
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Report summary retrieved successfully", "data": summary})
}

// ExportReport exports a report in the specified format (placeholder for future implementation)
func (c *FinancialReportController) ExportReport(ctx *gin.Context) {
	var req models.ReportExportRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// For now, return a placeholder response
	// In a real implementation, you would generate the file in the requested format
	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "Export functionality not yet implemented"})
}
