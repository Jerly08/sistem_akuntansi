package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SSOTBalanceSheetController handles Balance Sheet reports from SSOT Journal System
type SSOTBalanceSheetController struct {
	ssotBalanceSheetService *services.SSOTBalanceSheetService
	pdfService              services.PDFServiceInterface
}

// NewSSOTBalanceSheetController creates a new SSOT Balance Sheet controller
func NewSSOTBalanceSheetController(db *gorm.DB) *SSOTBalanceSheetController {
	return &SSOTBalanceSheetController{
		ssotBalanceSheetService: services.NewSSOTBalanceSheetService(db),
		pdfService:              services.NewPDFService(db),
	}
}

// GenerateSSOTBalanceSheet generates a comprehensive Balance Sheet from SSOT journal system
// @Summary Generate SSOT Balance Sheet
// @Description Generate a comprehensive Balance Sheet using Single Source of Truth (SSOT) journal system with real-time data integration and automatic balance validation
// @Tags SSOT Reports
// @Accept json
// @Produce json
// @Param as_of_date query string false "As of date (YYYY-MM-DD)" example(2025-12-31)
// @Param format query string false "Output format" Enums(json,summary,pdf,csv) default(json)
// @Success 200 {object} map[string]interface{} "Balance Sheet generated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /reports/ssot/balance-sheet [get]
func (ctrl *SSOTBalanceSheetController) GenerateSSOTBalanceSheet(c *gin.Context) {
	// Get parameters
	asOfDate := c.DefaultQuery("as_of_date", time.Now().Format("2006-01-02"))
	format := c.DefaultQuery("format", "json")

	// Validate date format
	if _, err := time.Parse("2006-01-02", asOfDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid as_of_date format",
			"message": "Date must be in YYYY-MM-DD format",
			"example": "2024-12-31",
		})
		return
	}

	// Generate Balance Sheet
	balanceSheetData, err := ctrl.ssotBalanceSheetService.GenerateSSOTBalanceSheet(asOfDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to generate SSOT Balance Sheet",
			"message": err.Error(),
		})
		return
	}

	// Handle different output formats
	switch format {
	case "json":
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "SSOT Balance Sheet generated successfully",
			"data":    balanceSheetData,
		})
	case "summary":
		// Return a simplified summary view
		summary := createBalanceSheetSummary(balanceSheetData)
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "SSOT Balance Sheet summary generated successfully",
			"data":    summary,
		})
	case "pdf":
pdfBytes, err := ctrl.pdfService.GenerateBalanceSheetPDF(balanceSheetData, asOfDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Failed to generate Balance Sheet PDF",
				"error":   err.Error(),
			})
			return
		}
		filename := fmt.Sprintf("SSOT_BalanceSheet_%s.pdf", asOfDate)
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Header("Content-Length", strconv.Itoa(len(pdfBytes)))
		c.Data(http.StatusOK, "application/pdf", pdfBytes)
	case "csv":
		// Follow SSOT P&L style: return JSON with export metadata
		response := gin.H{
			"as_of_date":     asOfDate,
			"data":           balanceSheetData,
			"export_format":  "csv",
			"export_ready":   true,
			"csv_headers":    []string{"Section", "Account Code", "Account Name", "Balance"},
			"data_source":    "SSOT Journal System",
			"generated_at":   time.Now().Format(time.RFC3339),
			"report_title":   "SSOT Balance Sheet",
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "data": response, "format": "csv"})
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid format",
			"message": "Supported formats: json, summary, pdf, csv",
		})
	}
}

// GetSSOTBalanceSheetAccountDetails provides detailed account information for drilldown
// @Summary Get SSOT Balance Sheet Account Details
// @Description Get detailed account information for Balance Sheet drilldown analysis
// @Tags SSOT Reports
// @Accept json
// @Produce json
// @Param as_of_date query string false "As of date (YYYY-MM-DD)" example(2025-12-31)
// @Param account_type query string false "Filter by account type" Enums(ASSET,LIABILITY,EQUITY)
// @Success 200 {object} map[string]interface{} "Account details retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security BearerAuth
// @Router /reports/ssot/balance-sheet/account-details [get]
func (ctrl *SSOTBalanceSheetController) GetSSOTBalanceSheetAccountDetails(c *gin.Context) {
	asOfDate := c.DefaultQuery("as_of_date", time.Now().Format("2006-01-02"))
	accountType := c.Query("account_type") // ASSET, LIABILITY, or EQUITY
	
	// Validate parameters
	if _, err := time.Parse("2006-01-02", asOfDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid as_of_date format",
		})
		return
	}

	if accountType != "" && accountType != "ASSET" && accountType != "LIABILITY" && accountType != "EQUITY" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid account_type. Must be ASSET, LIABILITY, or EQUITY",
		})
		return
	}

	// Generate full balance sheet to get account details
	balanceSheetData, err := ctrl.ssotBalanceSheetService.GenerateSSOTBalanceSheet(asOfDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get Balance Sheet account details",
			"message": err.Error(),
		})
		return
	}

	// Filter account details by type if specified
	accountDetails := balanceSheetData.AccountDetails
	if accountType != "" {
		filteredDetails := []services.SSOTAccountBalance{}
		for _, detail := range accountDetails {
			if detail.AccountType == accountType {
				filteredDetails = append(filteredDetails, detail)
			}
		}
		accountDetails = filteredDetails
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Balance Sheet account details retrieved successfully",
		"data": map[string]interface{}{
			"as_of_date":      asOfDate,
			"account_type":    accountType,
			"account_details": accountDetails,
			"total_accounts":  len(accountDetails),
		},
	})
}

// ValidateSSOTBalanceSheet validates if the balance sheet balances correctly
func (ctrl *SSOTBalanceSheetController) ValidateSSOTBalanceSheet(c *gin.Context) {
	asOfDate := c.DefaultQuery("as_of_date", time.Now().Format("2006-01-02"))
	
	// Validate date format
	if _, err := time.Parse("2006-01-02", asOfDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid as_of_date format",
		})
		return
	}

	// Generate Balance Sheet
	balanceSheetData, err := ctrl.ssotBalanceSheetService.GenerateSSOTBalanceSheet(asOfDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to validate Balance Sheet",
			"message": err.Error(),
		})
		return
	}

	// Create validation result
	validationResult := map[string]interface{}{
		"as_of_date":                  asOfDate,
		"is_balanced":                balanceSheetData.IsBalanced,
		"total_assets":               balanceSheetData.Assets.TotalAssets,
		"total_liabilities_and_equity": balanceSheetData.TotalLiabilitiesAndEquity,
		"balance_difference":         balanceSheetData.BalanceDifference,
		"tolerance":                  0.01,
		"validation_status":          "PASS",
		"generated_at":               balanceSheetData.GeneratedAt,
	}

	if !balanceSheetData.IsBalanced {
		validationResult["validation_status"] = "FAIL"
		validationResult["issue"] = "Assets do not equal Liabilities + Equity"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Balance Sheet validation completed",
		"data":    validationResult,
	})
}

// GetSSOTBalanceSheetComparison compares balance sheets between two dates
func (ctrl *SSOTBalanceSheetController) GetSSOTBalanceSheetComparison(c *gin.Context) {
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")
	
	// Default to comparing current date with 1 year ago
	if fromDate == "" {
		fromDate = time.Now().AddDate(-1, 0, 0).Format("2006-01-02")
	}
	if toDate == "" {
		toDate = time.Now().Format("2006-01-02")
	}
	
	// Validate date formats
	if _, err := time.Parse("2006-01-02", fromDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid from_date format",
		})
		return
	}
	if _, err := time.Parse("2006-01-02", toDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid to_date format", 
		})
		return
	}

	// Generate both balance sheets
	fromBS, err := ctrl.ssotBalanceSheetService.GenerateSSOTBalanceSheet(fromDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate from_date Balance Sheet",
			"message": err.Error(),
		})
		return
	}

	toBS, err := ctrl.ssotBalanceSheetService.GenerateSSOTBalanceSheet(toDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate to_date Balance Sheet",
			"message": err.Error(),
		})
		return
	}

	// Create comparison
	comparison := map[string]interface{}{
		"from_date": fromDate,
		"to_date":   toDate,
		"comparison": map[string]interface{}{
			"total_assets": map[string]interface{}{
				"from":   fromBS.Assets.TotalAssets,
				"to":     toBS.Assets.TotalAssets,
				"change": toBS.Assets.TotalAssets - fromBS.Assets.TotalAssets,
				"change_percent": calculatePercentChange(fromBS.Assets.TotalAssets, toBS.Assets.TotalAssets),
			},
			"total_liabilities": map[string]interface{}{
				"from":   fromBS.Liabilities.TotalLiabilities,
				"to":     toBS.Liabilities.TotalLiabilities,
				"change": toBS.Liabilities.TotalLiabilities - fromBS.Liabilities.TotalLiabilities,
				"change_percent": calculatePercentChange(fromBS.Liabilities.TotalLiabilities, toBS.Liabilities.TotalLiabilities),
			},
			"total_equity": map[string]interface{}{
				"from":   fromBS.Equity.TotalEquity,
				"to":     toBS.Equity.TotalEquity,
				"change": toBS.Equity.TotalEquity - fromBS.Equity.TotalEquity,
				"change_percent": calculatePercentChange(fromBS.Equity.TotalEquity, toBS.Equity.TotalEquity),
			},
		},
		"balance_sheet_from": fromBS,
		"balance_sheet_to":   toBS,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Balance Sheet comparison completed",
		"data":    comparison,
	})
}

// Helper function to create a summary view of the balance sheet
func createBalanceSheetSummary(bs *services.SSOTBalanceSheetData) map[string]interface{} {
	return map[string]interface{}{
		"company":     bs.Company,
		"as_of_date":  bs.AsOfDate.Format("2006-01-02"),
		"currency":    bs.Currency,
		"assets": map[string]interface{}{
			"current_assets":     bs.Assets.CurrentAssets.TotalCurrentAssets,
			"non_current_assets": bs.Assets.NonCurrentAssets.TotalNonCurrentAssets,
			"total_assets":       bs.Assets.TotalAssets,
		},
		"liabilities": map[string]interface{}{
			"current_liabilities":     bs.Liabilities.CurrentLiabilities.TotalCurrentLiabilities,
			"non_current_liabilities": bs.Liabilities.NonCurrentLiabilities.TotalNonCurrentLiabilities,
			"total_liabilities":       bs.Liabilities.TotalLiabilities,
		},
		"equity": map[string]interface{}{
			"total_equity": bs.Equity.TotalEquity,
		},
		"balance_check": map[string]interface{}{
			"is_balanced":       bs.IsBalanced,
			"balance_difference": bs.BalanceDifference,
		},
		"generated_at": bs.GeneratedAt,
		"enhanced":     bs.Enhanced,
	}
}

// Helper function to calculate percentage change
func calculatePercentChange(from, to float64) float64 {
	if from == 0 {
		if to == 0 {
			return 0
		}
		return 100 // Arbitrary large percentage for new items
	}
	return ((to - from) / from) * 100
}