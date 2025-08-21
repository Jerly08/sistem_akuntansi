package controllers

import (
	"net/http"
	"strconv"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

type ReportController struct {
	reportService *services.ReportService
}

func NewReportController(reportService *services.ReportService) *ReportController {
	return &ReportController{
		reportService: reportService,
	}
}

// GetReportsList returns available reports list
func (rc *ReportController) GetReportsList(c *gin.Context) {
	reports := rc.reportService.GetAvailableReports()
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   reports,
	})
}

// GetBalanceSheet generates Balance Sheet report
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
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"message": "Invalid date format. Use YYYY-MM-DD",
			})
			return
		}
	}

	report, err := rc.reportService.GenerateBalanceSheet(date, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"message": "Failed to generate balance sheet",
			"error": err.Error(),
		})
		return
	}

	if format == "pdf" {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=balance_sheet.pdf")
		c.Data(http.StatusOK, "application/pdf", report.FileData)
		return
	} else if format == "excel" {
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=balance_sheet.xlsx")
		c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", report.FileData)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// GetProfitLoss generates Profit & Loss Statement
func (rc *ReportController) GetProfitLoss(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	format := c.DefaultQuery("format", "json")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	report, err := rc.reportService.GenerateProfitLoss(start, end, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"message": "Failed to generate profit & loss statement",
			"error": err.Error(),
		})
		return
	}

	if format == "pdf" {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=profit_loss.pdf")
		c.Data(http.StatusOK, "application/pdf", report.FileData)
		return
	} else if format == "excel" {
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=profit_loss.xlsx")
		c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", report.FileData)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// GetCashFlow generates Cash Flow Statement
func (rc *ReportController) GetCashFlow(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	format := c.DefaultQuery("format", "json")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	report, err := rc.reportService.GenerateCashFlow(start, end, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"message": "Failed to generate cash flow statement",
			"error": err.Error(),
		})
		return
	}

	if format == "pdf" {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=cash_flow.pdf")
		c.Data(http.StatusOK, "application/pdf", report.FileData)
		return
	} else if format == "excel" {
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=cash_flow.xlsx")
		c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", report.FileData)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// GetTrialBalance generates Trial Balance report
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
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"message": "Invalid date format. Use YYYY-MM-DD",
			})
			return
		}
	}

	report, err := rc.reportService.GenerateTrialBalance(date, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"message": "Failed to generate trial balance",
			"error": err.Error(),
		})
		return
	}

	if format == "pdf" {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=trial_balance.pdf")
		c.Data(http.StatusOK, "application/pdf", report.FileData)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// GetGeneralLedger generates General Ledger report
func (rc *ReportController) GetGeneralLedger(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	accountCode := c.Query("account_code")
	format := c.DefaultQuery("format", "json")

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	report, err := rc.reportService.GenerateGeneralLedger(start, end, accountCode, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"message": "Failed to generate general ledger",
			"error": err.Error(),
		})
		return
	}

	if format == "pdf" {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=general_ledger.pdf")
		c.Data(http.StatusOK, "application/pdf", report.FileData)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// GetAccountsReceivable generates Accounts Receivable report
func (rc *ReportController) GetAccountsReceivable(c *gin.Context) {
	asOfDate := c.Query("as_of_date")
	customerID := c.Query("customer_id")
	format := c.DefaultQuery("format", "json")

	var date time.Time
	var err error
	if asOfDate == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", asOfDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"message": "Invalid date format. Use YYYY-MM-DD",
			})
			return
		}
	}

	var custID *uint
	if customerID != "" {
		id, err := strconv.ParseUint(customerID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"message": "Invalid customer_id format",
			})
			return
		}
		custIDUint := uint(id)
		custID = &custIDUint
	}

	report, err := rc.reportService.GenerateAccountsReceivable(date, custID, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"message": "Failed to generate accounts receivable report",
			"error": err.Error(),
		})
		return
	}

	if format == "pdf" {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=accounts_receivable.pdf")
		c.Data(http.StatusOK, "application/pdf", report.FileData)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// GetAccountsPayable generates Accounts Payable report
func (rc *ReportController) GetAccountsPayable(c *gin.Context) {
	asOfDate := c.Query("as_of_date")
	vendorID := c.Query("vendor_id")
	format := c.DefaultQuery("format", "json")

	var date time.Time
	var err error
	if asOfDate == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", asOfDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"message": "Invalid date format. Use YYYY-MM-DD",
			})
			return
		}
	}

	var vendID *uint
	if vendorID != "" {
		id, err := strconv.ParseUint(vendorID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"message": "Invalid vendor_id format",
			})
			return
		}
		vendIDUint := uint(id)
		vendID = &vendIDUint
	}

	report, err := rc.reportService.GenerateAccountsPayable(date, vendID, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"message": "Failed to generate accounts payable report",
			"error": err.Error(),
		})
		return
	}

	if format == "pdf" {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=accounts_payable.pdf")
		c.Data(http.StatusOK, "application/pdf", report.FileData)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// GetSalesSummary generates Sales Summary report
func (rc *ReportController) GetSalesSummary(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	format := c.DefaultQuery("format", "json")
	groupBy := c.DefaultQuery("group_by", "month") // month, quarter, year

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	report, err := rc.reportService.GenerateSalesSummary(start, end, groupBy, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"message": "Failed to generate sales summary report",
			"error": err.Error(),
		})
		return
	}

	if format == "pdf" {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=sales_summary.pdf")
		c.Data(http.StatusOK, "application/pdf", report.FileData)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// GetPurchaseSummary generates Purchase Summary report
func (rc *ReportController) GetPurchaseSummary(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	format := c.DefaultQuery("format", "json")
	groupBy := c.DefaultQuery("group_by", "month") // month, quarter, year

	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	report, err := rc.reportService.GeneratePurchaseSummary(start, end, groupBy, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"message": "Failed to generate purchase summary report",
			"error": err.Error(),
		})
		return
	}

	if format == "pdf" {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=purchase_summary.pdf")
		c.Data(http.StatusOK, "application/pdf", report.FileData)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// GetInventoryReport generates Inventory report
func (rc *ReportController) GetInventoryReport(c *gin.Context) {
	asOfDate := c.Query("as_of_date")
	format := c.DefaultQuery("format", "json")
	includeValuation := c.DefaultQuery("include_valuation", "true") == "true"

	var date time.Time
	var err error
	if asOfDate == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", asOfDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"message": "Invalid date format. Use YYYY-MM-DD",
			})
			return
		}
	}

	report, err := rc.reportService.GenerateInventoryReport(date, includeValuation, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"message": "Failed to generate inventory report",
			"error": err.Error(),
		})
		return
	}

	if format == "pdf" {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=inventory_report.pdf")
		c.Data(http.StatusOK, "application/pdf", report.FileData)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// GetFinancialRatios generates Financial Ratios analysis
func (rc *ReportController) GetFinancialRatios(c *gin.Context) {
	asOfDate := c.Query("as_of_date")
	period := c.DefaultQuery("period", "current") // current, ytd, comparative
	format := c.DefaultQuery("format", "json")

	var date time.Time
	var err error
	if asOfDate == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", asOfDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"message": "Invalid date format. Use YYYY-MM-DD",
			})
			return
		}
	}

	report, err := rc.reportService.GenerateFinancialRatios(date, period, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"message": "Failed to generate financial ratios report",
			"error": err.Error(),
		})
		return
	}

	if format == "pdf" {
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", "attachment; filename=financial_ratios.pdf")
		c.Data(http.StatusOK, "application/pdf", report.FileData)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   report,
	})
}

// SaveReportTemplate saves a custom report template
func (rc *ReportController) SaveReportTemplate(c *gin.Context) {
	var request models.ReportTemplateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
			"message": "Invalid request data",
			"error": err.Error(),
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status": "error",
			"message": "User not authenticated",
		})
		return
	}

	template, err := rc.reportService.SaveReportTemplate(&request, userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"message": "Failed to save report template",
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status": "success",
		"data":   template,
	})
}

// GetReportTemplates returns available report templates
func (rc *ReportController) GetReportTemplates(c *gin.Context) {
	templates, err := rc.reportService.GetReportTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "error",
			"message": "Failed to fetch report templates",
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   templates,
	})
}
