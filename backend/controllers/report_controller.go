package controllers

import (
	"net/http"
	"strconv"
	"time"

	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

// ReportController handles report generation endpoints
type ReportController struct {
	reportService      *services.ReportService
	professionalService *services.ProfessionalReportService
	standardizedService *services.StandardizedReportService
}

// NewReportController creates a new report controller
func NewReportController(
	reportService *services.ReportService,
	professionalService *services.ProfessionalReportService,
	standardizedService *services.StandardizedReportService,
) *ReportController {
	return &ReportController{
		reportService:      reportService,
		professionalService: professionalService,
		standardizedService: standardizedService,
	}
}

// GetProfessionalBalanceSheet generates a professional Balance Sheet report
func (rc *ReportController) GetProfessionalBalanceSheet(c *gin.Context) {
	asOfDate := c.Query("as_of_date")
	format := c.DefaultQuery("format", "pdf")

	var date time.Time
	var err error
	if asOfDate == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", asOfDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
	}

	if format == "pdf" {
		pdfData, err := rc.professionalService.GenerateBalanceSheetPDF(date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to generate PDF report", "error": err.Error()})
			return
		}
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=professional_balance_sheet.pdf")
		c.Data(http.StatusOK, "application/pdf", pdfData)
	} else if format == "csv" {
		csvData, err := rc.professionalService.GenerateBalanceSheetCSV(date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to generate CSV report", "error": err.Error()})
			return
		}
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=professional_balance_sheet.csv")
		c.Data(http.StatusOK, "text/csv", csvData)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Unsupported format"})
	}
}

// Note: GetProfessionalProfitLoss method removed - use Enhanced P&L Controller at /api/reports/enhanced/profit-loss instead

// GetProfessionalCashFlow generates a professional Cash Flow Statement report
func (rc *ReportController) GetProfessionalCashFlow(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	format := c.DefaultQuery("format", "pdf")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "start_date and end_date are required"})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}

	if format == "pdf" {
		pdfData, err := rc.professionalService.GenerateCashFlowStatementPDF(start, end)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to generate PDF report", "error": err.Error()})
			return
		}
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=professional_cash_flow.pdf")
		c.Data(http.StatusOK, "application/pdf", pdfData)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Unsupported format"})
	}
}

// GetProfessionalSalesSummary generates a professional Sales Summary report
func (rc *ReportController) GetProfessionalSalesSummary(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "month")
	format := c.DefaultQuery("format", "pdf")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "start_date and end_date are required"})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}

	if format == "pdf" {
		pdfData, err := rc.professionalService.GenerateSalesSummaryPDF(start, end, groupBy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to generate PDF report", "error": err.Error()})
			return
		}
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=professional_sales_summary.pdf")
		c.Data(http.StatusOK, "application/pdf", pdfData)
	} else if format == "csv" {
		csvData, err := rc.professionalService.GenerateSalesSummaryCSV(start, end, groupBy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to generate CSV report", "error": err.Error()})
			return
		}
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=professional_sales_summary.csv")
		c.Data(http.StatusOK, "text/csv", csvData)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Unsupported format"})
	}
}

// GetProfessionalPurchaseSummary generates a professional Purchase Summary report
func (rc *ReportController) GetProfessionalPurchaseSummary(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "month")
	format := c.DefaultQuery("format", "pdf")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "start_date and end_date are required"})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}

	if format == "pdf" {
		pdfData, err := rc.professionalService.GeneratePurchaseSummaryPDF(start, end, groupBy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to generate PDF report", "error": err.Error()})
			return
		}
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=professional_purchase_summary.pdf")
		c.Data(http.StatusOK, "application/pdf", pdfData)
	} else if format == "csv" {
		csvData, err := rc.professionalService.GeneratePurchaseSummaryCSV(start, end, groupBy)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to generate CSV report", "error": err.Error()})
			return
		}
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=professional_purchase_summary.csv")
		c.Data(http.StatusOK, "text/csv", csvData)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Unsupported format"})
	}
}

// GetReportsList returns the list of available reports
func (rc *ReportController) GetReportsList(c *gin.Context) {
	reports := rc.reportService.GetAvailableReports()
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": reports})
}

// GetBalanceSheet generates a standard balance sheet
func (rc *ReportController) GetBalanceSheet(c *gin.Context) {
	asOfDate := c.Query("as_of_date")
	format := c.DefaultQuery("format", "json")

	var date time.Time
	var err error
	if asOfDate == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", asOfDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
	}

	report, err := rc.reportService.GenerateBalanceSheet(date, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": report})
}

// Note: GetProfitLoss method removed - use Enhanced P&L Controller at /api/reports/enhanced/profit-loss instead

// GetCashFlow generates a cash flow statement
func (rc *ReportController) GetCashFlow(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	format := c.DefaultQuery("format", "json")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "start_date and end_date are required"})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}

	report, err := rc.reportService.GenerateCashFlow(start, end, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": report})
}

// GetTrialBalance generates a trial balance
func (rc *ReportController) GetTrialBalance(c *gin.Context) {
	asOfDate := c.Query("as_of_date")
	format := c.DefaultQuery("format", "json")

	var date time.Time
	var err error
	if asOfDate == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", asOfDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
	}

	report, err := rc.reportService.GenerateTrialBalance(date, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": report})
}

// GetGeneralLedger generates a general ledger report
func (rc *ReportController) GetGeneralLedger(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	accountCode := c.Query("account_code")
	format := c.DefaultQuery("format", "json")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "start_date and end_date are required"})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}

	report, err := rc.reportService.GenerateGeneralLedger(start, end, accountCode, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": report})
}

// GetAccountsReceivable generates accounts receivable report
func (rc *ReportController) GetAccountsReceivable(c *gin.Context) {
	asOfDate := c.Query("as_of_date")
	customerIDStr := c.Query("customer_id")
	format := c.DefaultQuery("format", "json")

	var date time.Time
	var err error
	if asOfDate == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", asOfDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
	}

	var customerID *uint
	if customerIDStr != "" {
		var id uint64
		id, err = strconv.ParseUint(customerIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid customer_id format"})
			return
		}
		cid := uint(id)
		customerID = &cid
	}

	report, err := rc.reportService.GenerateAccountsReceivable(date, customerID, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": report})
}

// GetAccountsPayable generates accounts payable report
func (rc *ReportController) GetAccountsPayable(c *gin.Context) {
	asOfDate := c.Query("as_of_date")
	vendorIDStr := c.Query("vendor_id")
	format := c.DefaultQuery("format", "json")

	var date time.Time
	var err error
	if asOfDate == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", asOfDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
	}

	var vendorID *uint
	if vendorIDStr != "" {
		var id uint64
		id, err = strconv.ParseUint(vendorIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid vendor_id format"})
			return
		}
		vid := uint(id)
		vendorID = &vid
	}

	report, err := rc.reportService.GenerateAccountsPayable(date, vendorID, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": report})
}

// GetSalesSummary generates sales summary report
func (rc *ReportController) GetSalesSummary(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "month")
	format := c.DefaultQuery("format", "json")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "start_date and end_date are required"})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}

	report, err := rc.reportService.GenerateSalesSummary(start, end, groupBy, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": report})
}

// GetPurchaseSummary generates purchase summary report
func (rc *ReportController) GetPurchaseSummary(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "month")
	format := c.DefaultQuery("format", "json")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "start_date and end_date are required"})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}

	report, err := rc.reportService.GeneratePurchaseSummary(start, end, groupBy, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": report})
}

// GetVendorAnalysis generates vendor analysis report
func (rc *ReportController) GetVendorAnalysis(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	groupBy := c.DefaultQuery("group_by", "month")
	format := c.DefaultQuery("format", "json")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "start_date and end_date are required"})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}

	report, err := rc.reportService.GenerateVendorAnalysis(start, end, groupBy, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": report})
}

// GetInventoryReport generates inventory report
func (rc *ReportController) GetInventoryReport(c *gin.Context) {
	asOfDate := c.Query("as_of_date")
	includeValuationStr := c.DefaultQuery("include_valuation", "false")
	format := c.DefaultQuery("format", "json")

	var date time.Time
	var err error
	if asOfDate == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", asOfDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
	}

	includeValuation := includeValuationStr == "true"

	report, err := rc.reportService.GenerateInventoryReport(date, includeValuation, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": report})
}

// GetFinancialRatios generates financial ratios report
func (rc *ReportController) GetFinancialRatios(c *gin.Context) {
	asOfDate := c.Query("as_of_date")
	period := c.DefaultQuery("period", "annual")
	format := c.DefaultQuery("format", "json")

	var date time.Time
	var err error
	if asOfDate == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", asOfDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid date format. Use YYYY-MM-DD"})
			return
		}
	}

	report, err := rc.reportService.GenerateFinancialRatios(date, period, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": report})
}

// GetReportTemplates returns report templates
func (rc *ReportController) GetReportTemplates(c *gin.Context) {
	templates, err := rc.reportService.GetReportTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "data": templates})
}

// SaveReportTemplate saves a report template
func (rc *ReportController) SaveReportTemplate(c *gin.Context) {
	// This method would need proper implementation based on models.ReportTemplateRequest
	c.JSON(http.StatusNotImplemented, gin.H{"status": "error", "message": "Report template saving not yet implemented"})
}
