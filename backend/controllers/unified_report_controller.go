package controllers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UnifiedReportController struct {
	db                       *gorm.DB
	accountRepo              repositories.AccountRepository
	salesRepo                *repositories.SalesRepository
	purchaseRepo             *repositories.PurchaseRepository
	contactRepo              repositories.ContactRepository
	productRepo              *repositories.ProductRepository
	balanceSheetService      *services.StandardizedReportService
	profitLossService        *services.EnhancedProfitLossService
	cashFlowService          *services.StandardizedReportService
	trialBalanceService      *services.ReportService
	generalLedgerService     *services.ReportService
	salesSummaryService      *services.ReportService
	vendorAnalysisService    *services.ReportService
}

func NewUnifiedReportController(
	db *gorm.DB,
	accountRepo repositories.AccountRepository,
	salesRepo *repositories.SalesRepository,
	purchaseRepo *repositories.PurchaseRepository,
	contactRepo repositories.ContactRepository,
	productRepo *repositories.ProductRepository,
	reportService *services.ReportService,
	balanceSheetService *services.StandardizedReportService,
	profitLossService *services.EnhancedProfitLossService,
	cashFlowService *services.StandardizedReportService,
) *UnifiedReportController {
	return &UnifiedReportController{
		db:                    db,
		accountRepo:           accountRepo,
		salesRepo:             salesRepo,
		purchaseRepo:          purchaseRepo,
		contactRepo:           contactRepo,
		productRepo:           productRepo,
		balanceSheetService:   balanceSheetService,
		profitLossService:     profitLossService,
		cashFlowService:       cashFlowService,
		trialBalanceService:   reportService,
		generalLedgerService:  reportService,
		salesSummaryService:   reportService,
		vendorAnalysisService: reportService,
	}
}

// GenerateReport - Central endpoint for all report generation
func (ctrl *UnifiedReportController) GenerateReport(c *gin.Context) {
	startTime := time.Now()
	
	// Extract report type from URL path or param
	reportType := c.Param("type")
	if reportType == "" {
		// Extract from URL path for direct endpoints like /api/reports/balance-sheet
		path := c.Request.URL.Path
		if strings.Contains(path, "/balance-sheet") {
			reportType = "balance-sheet"
		} else if strings.Contains(path, "/profit-loss") {
			reportType = "profit-loss"
		} else if strings.Contains(path, "/cash-flow") {
			reportType = "cash-flow"
		} else if strings.Contains(path, "/trial-balance") {
			reportType = "trial-balance"
		} else if strings.Contains(path, "/general-ledger") {
			reportType = "general-ledger"
		} else if strings.Contains(path, "/sales-summary") {
			reportType = "sales-summary"
		} else if strings.Contains(path, "/vendor-analysis") {
			reportType = "vendor-analysis"
		}
	}

	// Validate report type
	if !ctrl.isValidReportType(reportType) {
		ctrl.sendErrorResponse(c, http.StatusBadRequest, "INVALID_REPORT_TYPE", 
			"Invalid report type specified", map[string]interface{}{
				"report_type": reportType,
				"valid_types": ctrl.getValidReportTypes(),
			})
		return
	}

	// Parse and validate parameters
	params, err := ctrl.parseReportParameters(c, reportType)
	if err != nil {
		ctrl.sendErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETERS", 
			err.Error(), map[string]interface{}{
				"report_type": reportType,
			})
		return
	}

	// Check format - if PDF/CSV/Excel, generate file; if JSON, return data
	format := strings.ToLower(params["format"].(string))
	
	// Generate report based on type and format
	switch reportType {
	case "balance-sheet":
		ctrl.generateBalanceSheet(c, params, format, startTime)
	case "profit-loss":
		ctrl.generateProfitLoss(c, params, format, startTime)
	case "cash-flow":
		ctrl.generateCashFlow(c, params, format, startTime)
	case "trial-balance":
		ctrl.generateTrialBalance(c, params, format, startTime)
	case "general-ledger":
		ctrl.generateGeneralLedger(c, params, format, startTime)
	case "sales-summary":
		ctrl.generateSalesSummary(c, params, format, startTime)
	case "vendor-analysis":
		ctrl.generateVendorAnalysis(c, params, format, startTime)
	default:
		ctrl.sendErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", 
			fmt.Sprintf("Report type '%s' is not yet implemented", reportType), nil)
	}
}

// PreviewReport - Generate preview data for reports
func (ctrl *UnifiedReportController) PreviewReport(c *gin.Context) {
	startTime := time.Now()
	reportType := c.Param("type")

	// Validate report type
	if !ctrl.isValidReportType(reportType) {
		ctrl.sendErrorResponse(c, http.StatusBadRequest, "INVALID_REPORT_TYPE", 
			"Invalid report type specified", map[string]interface{}{
				"report_type": reportType,
			})
		return
	}

	// Parse parameters - force JSON format for preview
	params, err := ctrl.parseReportParameters(c, reportType)
	if err != nil {
		ctrl.sendErrorResponse(c, http.StatusBadRequest, "INVALID_PARAMETERS", 
			err.Error(), nil)
		return
	}
	params["format"] = "json" // Force JSON for preview

	// Generate preview based on type
	switch reportType {
	case "balance-sheet":
		ctrl.generateBalanceSheet(c, params, "json", startTime)
	case "profit-loss":
		ctrl.generateProfitLoss(c, params, "json", startTime)
	case "cash-flow":
		ctrl.generateCashFlow(c, params, "json", startTime)
	case "trial-balance":
		ctrl.generateTrialBalance(c, params, "json", startTime)
	case "general-ledger":
		ctrl.generateGeneralLedger(c, params, "json", startTime)
	case "sales-summary":
		ctrl.generateSalesSummary(c, params, "json", startTime)
	case "vendor-analysis":
		ctrl.generateVendorAnalysis(c, params, "json", startTime)
	default:
		ctrl.sendErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", 
			fmt.Sprintf("Preview for report type '%s' is not yet implemented", reportType), nil)
	}
}

// GetAvailableReports - Return metadata about available reports
func (ctrl *UnifiedReportController) GetAvailableReports(c *gin.Context) {
	reports := []map[string]interface{}{
		{
			"id":          "balance-sheet",
			"name":        "Balance Sheet",
			"description": "Provides a company's assets, liabilities, and shareholders' equity at a specific point in time",
			"type":        "FINANCIAL",
			"parameters": map[string]interface{}{
				"as_of_date": map[string]string{
					"type":        "date",
					"required":    "false",
					"default":     "today",
					"description": "Date for balance sheet snapshot",
				},
				"format": map[string]string{
					"type":        "string",
					"required":    "false",
					"default":     "json",
					"options":     "json,pdf,excel",
					"description": "Output format",
				},
			},
			"endpoints": map[string]string{
				"generate": "/api/reports/balance-sheet",
				"preview":  "/api/reports/preview/balance-sheet",
			},
		},
		{
			"id":          "profit-loss",
			"name":        "Profit and Loss Statement",
			"description": "Comprehensive profit and loss statements provides a detailed view of financial transactions on a specific period",
			"type":        "FINANCIAL",
			"parameters": map[string]interface{}{
				"start_date": map[string]string{
					"type":        "date",
					"required":    "true",
					"description": "Start date for P&L period",
				},
				"end_date": map[string]string{
					"type":        "date",
					"required":    "true", 
					"description": "End date for P&L period",
				},
				"format": map[string]string{
					"type":        "string",
					"required":    "false",
					"default":     "json",
					"options":     "json,pdf,excel",
					"description": "Output format",
				},
			},
			"endpoints": map[string]string{
				"generate": "/api/reports/profit-loss",
				"preview":  "/api/reports/preview/profit-loss",
			},
		},
		{
			"id":          "cash-flow",
			"name":        "Cash Flow Statement", 
			"description": "Measures how well a company generates cash to pay its debt obligations and fund its operating expenditures",
			"type":        "FINANCIAL",
			"parameters": map[string]interface{}{
				"start_date": map[string]string{
					"type":        "date",
					"required":    "true",
					"description": "Start date for cash flow period",
				},
				"end_date": map[string]string{
					"type":        "date", 
					"required":    "true",
					"description": "End date for cash flow period",
				},
				"format": map[string]string{
					"type":        "string",
					"required":    "false",
					"default":     "json",
					"options":     "json,pdf",
					"description": "Output format",
				},
			},
			"endpoints": map[string]string{
				"generate": "/api/reports/cash-flow",
				"preview":  "/api/reports/preview/cash-flow",
			},
		},
		{
			"id":          "trial-balance",
			"name":        "Trial Balance",
			"description": "Summary of all account balances to ensure debits equal credits and verify accounting equation",
			"type":        "FINANCIAL",
			"parameters": map[string]interface{}{
				"as_of_date": map[string]string{
					"type":        "date",
					"required":    "false",
					"default":     "today",
					"description": "Date for trial balance snapshot",
				},
				"format": map[string]string{
					"type":        "string",
					"required":    "false",
					"default":     "json",
					"options":     "json,pdf,excel",
					"description": "Output format",
				},
			},
			"endpoints": map[string]string{
				"generate": "/api/reports/trial-balance",
				"preview":  "/api/reports/preview/trial-balance",
			},
		},
		{
			"id":          "general-ledger",
			"name":        "General Ledger",
			"description": "Complete record of all financial transactions organized by account for detailed analysis",
			"type":        "FINANCIAL",
			"parameters": map[string]interface{}{
				"start_date": map[string]string{
					"type":        "date",
					"required":    "true",
					"description": "Start date for ledger period",
				},
				"end_date": map[string]string{
					"type":        "date",
					"required":    "true", 
					"description": "End date for ledger period",
				},
				"account_code": map[string]string{
					"type":        "string",
					"required":    "false",
					"description": "Specific account code to filter (optional)",
				},
				"format": map[string]string{
					"type":        "string",
					"required":    "false",
					"default":     "json",
					"options":     "json,pdf,excel",
					"description": "Output format",
				},
			},
			"endpoints": map[string]string{
				"generate": "/api/reports/general-ledger",
				"preview":  "/api/reports/preview/general-ledger",
			},
		},
		{
			"id":          "sales-summary",
			"name":        "Sales Summary Report",
			"description": "Provides a summary of sales transactions over a period",
			"type":        "OPERATIONAL",
			"parameters": map[string]interface{}{
				"start_date": map[string]string{
					"type":        "date",
					"required":    "true",
					"description": "Start date for sales period",
				},
				"end_date": map[string]string{
					"type":        "date",
					"required":    "true",
					"description": "End date for sales period",
				},
				"group_by": map[string]string{
					"type":        "string", 
					"required":    "false",
					"default":     "month",
					"options":     "day,week,month,quarter,year",
					"description": "Grouping period",
				},
				"format": map[string]string{
					"type":        "string",
					"required":    "false",
					"default":     "json",
					"options":     "json,pdf,excel",
					"description": "Output format",
				},
			},
			"endpoints": map[string]string{
				"generate": "/api/reports/sales-summary",
				"preview":  "/api/reports/preview/sales-summary",
			},
		},
		{
			"id":          "vendor-analysis",
			"name":        "Vendor Analysis Report",
			"description": "Comprehensive analysis of vendor transactions, payment history, and performance metrics",
			"type":        "OPERATIONAL",
			"parameters": map[string]interface{}{
				"start_date": map[string]string{
					"type":        "date",
					"required":    "true",
					"description": "Start date for analysis period",
				},
				"end_date": map[string]string{
					"type":        "date",
					"required":    "true",
					"description": "End date for analysis period",
				},
				"vendor_id": map[string]string{
					"type":        "string",
					"required":    "false",
					"description": "Specific vendor ID to filter (optional)",
				},
				"format": map[string]string{
					"type":        "string",
					"required":    "false",
					"default":     "json",
					"options":     "json,pdf,excel",
					"description": "Output format",
				},
			},
			"endpoints": map[string]string{
				"generate": "/api/reports/vendor-analysis",
				"preview":  "/api/reports/preview/vendor-analysis",
			},
		},
	}

	ctrl.sendSuccessResponse(c, reports, models.Metadata{
		ReportType:     "available_reports",
		GeneratedAt:    time.Now(),
		GeneratedBy:    ctrl.getCurrentUser(c),
		Parameters:     map[string]interface{}{},
		GenerationTime: time.Since(time.Now()).String(),
		RecordCount:    len(reports),
		Version:        "1.0.0",
		Format:         "json",
	})
}

// Individual report generators
func (ctrl *UnifiedReportController) generateBalanceSheet(c *gin.Context, params map[string]interface{}, format string, startTime time.Time) {
	asOfDate := params["as_of_date"].(time.Time)
	
	if format == "json" {
		// Generate JSON data
		report, err := ctrl.balanceSheetService.GenerateStandardBalanceSheet(asOfDate, "json", false)
		if err != nil {
			ctrl.sendErrorResponse(c, http.StatusInternalServerError, "GENERATION_FAILED", 
				fmt.Sprintf("Failed to generate balance sheet: %v", err), nil)
			return
		}

		ctrl.sendSuccessResponse(c, report.Statement, models.Metadata{
			ReportType:     "balance-sheet",
			GeneratedAt:    time.Now(),
			GeneratedBy:    ctrl.getCurrentUser(c),
			Parameters:     params,
			GenerationTime: time.Since(startTime).String(),
			RecordCount:    len(report.Statement.Sections),
			Version:        "1.0.0",
			Format:         format,
		})
	} else {
		// Generate file (PDF/Excel)
		report, err := ctrl.balanceSheetService.GenerateStandardBalanceSheet(asOfDate, format, false)
		if err != nil {
			ctrl.sendErrorResponse(c, http.StatusInternalServerError, "GENERATION_FAILED", 
				fmt.Sprintf("Failed to generate balance sheet: %v", err), nil)
			return
		}

		ctrl.sendFileResponse(c, report.FileData, fmt.Sprintf("balance-sheet-%s.%s", 
			asOfDate.Format("2006-01-02"), format), format)
	}
}

func (ctrl *UnifiedReportController) generateProfitLoss(c *gin.Context, params map[string]interface{}, format string, startTime time.Time) {
	startDate := params["start_date"].(time.Time)
	endDate := params["end_date"].(time.Time)

	if format == "json" {
		// Generate JSON data  
		report, err := ctrl.profitLossService.GenerateEnhancedProfitLoss(startDate, endDate)
		if err != nil {
			ctrl.sendErrorResponse(c, http.StatusInternalServerError, "GENERATION_FAILED", 
				fmt.Sprintf("Failed to generate profit & loss: %v", err), nil)
			return
		}

		ctrl.sendSuccessResponse(c, report, models.Metadata{
			ReportType:     "profit-loss",
			GeneratedAt:    time.Now(),
			GeneratedBy:    ctrl.getCurrentUser(c),
			Parameters:     params,
			GenerationTime: time.Since(startTime).String(),
			RecordCount:    1,
			Version:        "1.0.0",
			Format:         format,
		})
	} else {
		// Generate file (PDF)
		report, err := ctrl.balanceSheetService.GenerateStandardProfitLoss(startDate, endDate, format, false)
		if err != nil {
			ctrl.sendErrorResponse(c, http.StatusInternalServerError, "GENERATION_FAILED", 
				fmt.Sprintf("Failed to generate profit & loss: %v", err), nil)
			return
		}

		ctrl.sendFileResponse(c, report.FileData, fmt.Sprintf("profit-loss-%s-to-%s.%s", 
			startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), format), format)
	}
}

func (ctrl *UnifiedReportController) generateCashFlow(c *gin.Context, params map[string]interface{}, format string, startTime time.Time) {
	startDate := params["start_date"].(time.Time)
	endDate := params["end_date"].(time.Time)

	if format == "json" {
		// Get real cash flow data from database
		operatingActivities := ctrl.getOperatingActivities(startDate, endDate)
		investingActivities := ctrl.getInvestingActivities(startDate, endDate)
		financingActivities := ctrl.getFinancingActivities(startDate, endDate)
		
		// Calculate totals
		operatingTotal := ctrl.calculateSectionTotal(operatingActivities.Items)
		investingTotal := ctrl.calculateSectionTotal(investingActivities.Items)
		financingTotal := ctrl.calculateSectionTotal(financingActivities.Items)
		
		netCashFlow := operatingTotal + investingTotal + financingTotal
		
		// Get beginning and ending cash balances from database
		beginningBalance := ctrl.getCashBalanceAsOf(startDate.AddDate(0, 0, -1))
		endingBalance := beginningBalance + netCashFlow

		cashFlowData := models.CashFlowData{
			FinancialReportData: models.FinancialReportData{
				Company:     ctrl.getCompanyInfo(),
				ReportTitle: "Cash Flow Statement",
				Period:      fmt.Sprintf("%s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")),
				StartDate:   &startDate,
				EndDate:     &endDate,
				Currency:    "IDR",
				Sections:    []models.ReportSection{operatingActivities, investingActivities, financingActivities},
				Totals:      map[string]float64{
					"operating_total":   operatingTotal,
					"investing_total":   investingTotal,
					"financing_total":   financingTotal,
					"net_cash_flow":     netCashFlow,
				},
			},
			OperatingActivities:    operatingActivities,
			InvestingActivities:    investingActivities,
			FinancingActivities:    financingActivities,
			NetCashFlow:            netCashFlow,
			BeginningCashBalance:   beginningBalance,
			EndingCashBalance:      endingBalance,
		}

		ctrl.sendSuccessResponse(c, cashFlowData, models.Metadata{
			ReportType:     "cash-flow",
			GeneratedAt:    time.Now(),
			GeneratedBy:    ctrl.getCurrentUser(c),
			Parameters:     params,
			GenerationTime: time.Since(startTime).String(),
			RecordCount:    3, // Operating, Investing, Financing
			Version:        "1.0.0",
			Format:         format,
		})
	} else {
		ctrl.sendErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", 
			fmt.Sprintf("Cash flow report in %s format is not yet implemented", format), nil)
	}
}

func (ctrl *UnifiedReportController) generateTrialBalance(c *gin.Context, params map[string]interface{}, format string, startTime time.Time) {
	asOfDate := params["as_of_date"].(time.Time)

	if format == "json" {
		// Get all accounts with balances
		ctx := context.Background()
		accounts, err := ctrl.accountRepo.FindAll(ctx)
		if err != nil {
			ctrl.sendErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", 
				fmt.Sprintf("Failed to fetch accounts: %v", err), nil)
			return
		}

		var trialBalanceAccounts []models.TrialBalanceAccount
		var totalDebits, totalCredits float64

		for _, account := range accounts {
			if account.IsHeader {
				continue // Skip header accounts
			}

			// Calculate balance up to asOfDate
			balance := ctrl.calculateAccountBalance(account.ID, asOfDate)
			
			var debitBalance, creditBalance float64
			normalBalance := account.GetNormalBalance()
			
			if normalBalance == models.NormalBalanceDebit {
				if balance >= 0 {
					debitBalance = balance
				} else {
					creditBalance = -balance
				}
			} else {
				if balance >= 0 {
					creditBalance = balance
				} else {
					debitBalance = -balance
				}
			}

			if debitBalance != 0 || creditBalance != 0 {
				trialBalanceAccounts = append(trialBalanceAccounts, models.TrialBalanceAccount{
					AccountID:     account.ID,
					AccountCode:   account.Code,
					AccountName:   account.Name,
					AccountType:   account.Type,
					DebitBalance:  debitBalance,
					CreditBalance: creditBalance,
					Level:         account.Level,
					IsHeader:      account.IsHeader,
				})

				totalDebits += debitBalance
				totalCredits += creditBalance
			}
		}

		isBalanced := totalDebits == totalCredits
		balanceDifference := totalDebits - totalCredits

		trialBalanceData := models.TrialBalanceData{
			FinancialReportData: models.FinancialReportData{
				Company:     ctrl.getCompanyInfo(),
				ReportTitle: "Trial Balance",
				Period:      fmt.Sprintf("As of %s", asOfDate.Format("January 2, 2006")),
				AsOfDate:    &asOfDate,
				Currency:    "IDR",
				Totals:      map[string]float64{
					"total_debits":  totalDebits,
					"total_credits": totalCredits,
				},
			},
			Accounts:          trialBalanceAccounts,
			TotalDebits:       totalDebits,
			TotalCredits:      totalCredits,
			IsBalanced:        isBalanced,
			BalanceDifference: balanceDifference,
		}

		ctrl.sendSuccessResponse(c, trialBalanceData, models.Metadata{
			ReportType:     "trial-balance",
			GeneratedAt:    time.Now(),
			GeneratedBy:    ctrl.getCurrentUser(c),
			Parameters:     params,
			GenerationTime: time.Since(startTime).String(),
			RecordCount:    len(trialBalanceAccounts),
			Version:        "1.0.0",
			Format:         format,
		})
	} else {
		ctrl.sendErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", 
			fmt.Sprintf("Trial balance in %s format is not yet implemented", format), nil)
	}
}

func (ctrl *UnifiedReportController) generateGeneralLedger(c *gin.Context, params map[string]interface{}, format string, startTime time.Time) {
	startDate := params["start_date"].(time.Time)
	endDate := params["end_date"].(time.Time)
	accountCode := ""
	if params["account_code"] != nil {
		accountCode = params["account_code"].(string)
	}

	if format == "json" {
		// Get accounts to process
		ctx := context.Background()
		var accounts []models.Account
		var err error

		if accountCode != "" {
			// Get specific account
			var account *models.Account
			account, err = ctrl.accountRepo.FindByCode(ctx, accountCode)
			if err != nil {
				ctrl.sendErrorResponse(c, http.StatusNotFound, "ACCOUNT_NOT_FOUND", 
					fmt.Sprintf("Account with code %s not found", accountCode), nil)
				return
			}
			accounts = []models.Account{*account}
		} else {
			// Get all non-header accounts
			var allAccounts []models.Account
			allAccounts, err = ctrl.accountRepo.FindAll(ctx)
			if err != nil {
				ctrl.sendErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", 
					fmt.Sprintf("Failed to fetch accounts: %v", err), nil)
				return
			}
			for _, acc := range allAccounts {
				if !acc.IsHeader {
					accounts = append(accounts, acc)
				}
			}
		}

		var generalLedgerAccounts []models.GeneralLedgerAccount

		for _, account := range accounts {
			// Get opening balance
			openingBalance := ctrl.calculateAccountBalance(account.ID, startDate.AddDate(0, 0, -1))
			
			// Get transactions for the period
			var transactions []models.GeneralLedgerTransaction
			var totalDebits, totalCredits float64
			runningBalance := openingBalance

			// Query journal entries for this account and period
			rows, err := ctrl.db.Raw(`
				SELECT j.date, j.code as reference, je.description, 
				       je.debit_amount, je.credit_amount, j.id as journal_id
				FROM journal_entries je
				JOIN journals j ON je.journal_id = j.id
				WHERE je.account_id = ? AND j.date BETWEEN ? AND ?
				  AND j.status = 'POSTED'
				ORDER BY j.date, j.id
			`, account.ID, startDate, endDate).Rows()
			
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var date time.Time
					var reference, description string
					var debitAmount, creditAmount float64
					var journalID uint
					
					rows.Scan(&date, &reference, &description, &debitAmount, &creditAmount, &journalID)
					
					// Update running balance based on account type
					normalBalance := account.GetNormalBalance()
					if normalBalance == models.NormalBalanceDebit {
						runningBalance += debitAmount - creditAmount
					} else {
						runningBalance += creditAmount - debitAmount
					}

					transactions = append(transactions, models.GeneralLedgerTransaction{
						Date:         date,
						Reference:    reference,
						Description:  description,
						DebitAmount:  debitAmount,
						CreditAmount: creditAmount,
						Balance:      runningBalance,
						JournalID:    journalID,
					})

					totalDebits += debitAmount
					totalCredits += creditAmount
				}
			}

			// Only include accounts with activity
			if len(transactions) > 0 || openingBalance != 0 {
				generalLedgerAccounts = append(generalLedgerAccounts, models.GeneralLedgerAccount{
					AccountID:      account.ID,
					AccountCode:    account.Code,
					AccountName:    account.Name,
					AccountType:    account.Type,
					OpeningBalance: openingBalance,
					ClosingBalance: runningBalance,
					TotalDebits:    totalDebits,
					TotalCredits:   totalCredits,
					Transactions:   transactions,
				})
			}
		}

		generalLedgerData := models.GeneralLedgerData{
			FinancialReportData: models.FinancialReportData{
				Company:     ctrl.getCompanyInfo(),
				ReportTitle: "General Ledger",
				Period:      fmt.Sprintf("%s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")),
				StartDate:   &startDate,
				EndDate:     &endDate,
				Currency:    "IDR",
			},
			Accounts: generalLedgerAccounts,
		}

		ctrl.sendSuccessResponse(c, generalLedgerData, models.Metadata{
			ReportType:     "general-ledger",
			GeneratedAt:    time.Now(),
			GeneratedBy:    ctrl.getCurrentUser(c),
			Parameters:     params,
			GenerationTime: time.Since(startTime).String(),
			RecordCount:    len(generalLedgerAccounts),
			Version:        "1.0.0",
			Format:         format,
		})
	} else {
		ctrl.sendErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", 
			fmt.Sprintf("General ledger in %s format is not yet implemented", format), nil)
	}
}

func (ctrl *UnifiedReportController) generateSalesSummary(c *gin.Context, params map[string]interface{}, format string, startTime time.Time) {
	startDate := params["start_date"].(time.Time)
	endDate := params["end_date"].(time.Time)
	groupBy := params["group_by"].(string)

	if format == "json" {
		// Get sales data
		ctx := context.Background()
		
		// Get sales summary by period
		salesByPeriod, err := ctrl.getSalesByPeriod(ctx, startDate, endDate, groupBy)
		if err != nil {
			ctrl.sendErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", 
				fmt.Sprintf("Failed to fetch sales data: %v", err), nil)
			return
		}

		// Calculate totals
		var totalRevenue float64
		var totalTransactions int
		for _, period := range salesByPeriod {
			totalRevenue += period.Amount
			totalTransactions += period.Count
		}

		averageOrderValue := float64(0)
		if totalTransactions > 0 {
			averageOrderValue = totalRevenue / float64(totalTransactions)
		}

		salesSummaryData := models.SalesSummaryData{
			FinancialReportData: models.FinancialReportData{
				Company:     ctrl.getCompanyInfo(),
				ReportTitle: "Sales Summary Report",
				Period:      fmt.Sprintf("%s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")),
				StartDate:   &startDate,
				EndDate:     &endDate,
				Currency:    "IDR",
				Totals:      map[string]float64{
					"total_revenue": totalRevenue,
					"total_transactions": float64(totalTransactions),
					"average_order_value": averageOrderValue,
				},
			},
			SalesByPeriod:     salesByPeriod,
			TotalRevenue:      totalRevenue,
			TotalTransactions: totalTransactions,
			AverageOrderValue: averageOrderValue,
		}

		ctrl.sendSuccessResponse(c, salesSummaryData, models.Metadata{
			ReportType:     "sales-summary",
			GeneratedAt:    time.Now(),
			GeneratedBy:    ctrl.getCurrentUser(c),
			Parameters:     params,
			GenerationTime: time.Since(startTime).String(),
			RecordCount:    len(salesByPeriod),
			Version:        "1.0.0",
			Format:         format,
		})
	} else {
		ctrl.sendErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", 
			fmt.Sprintf("Sales summary in %s format is not yet implemented", format), nil)
	}
}

func (ctrl *UnifiedReportController) generateVendorAnalysis(c *gin.Context, params map[string]interface{}, format string, startTime time.Time) {
	startDate := params["start_date"].(time.Time)
	endDate := params["end_date"].(time.Time)

	if format == "json" {
		// Get purchases data
		ctx := context.Background()
		
		// Get purchases summary by period
		purchasesByPeriod, err := ctrl.getPurchasesByPeriod(ctx, startDate, endDate, "month")
		if err != nil {
			ctrl.sendErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", 
				fmt.Sprintf("Failed to fetch purchases data: %v", err), nil)
			return
		}

		// Calculate totals
		var totalPurchases float64
		var totalTransactions int
		for _, period := range purchasesByPeriod {
			totalPurchases += period.Amount
			totalTransactions += period.Count
		}

		averageOrderValue := float64(0)
		if totalTransactions > 0 {
			averageOrderValue = totalPurchases / float64(totalTransactions)
		}

		vendorAnalysisData := models.VendorAnalysisData{
			FinancialReportData: models.FinancialReportData{
				Company:     ctrl.getCompanyInfo(),
				ReportTitle: "Vendor Analysis Report",
				Period:      fmt.Sprintf("%s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")),
				StartDate:   &startDate,
				EndDate:     &endDate,
				Currency:    "IDR",
				Totals:      map[string]float64{
					"total_purchases": totalPurchases,
					"total_transactions": float64(totalTransactions),
					"average_order_value": averageOrderValue,
				},
			},
			PurchasesByPeriod: purchasesByPeriod,
			TotalPurchases:    totalPurchases,
			TotalTransactions: totalTransactions,
			AverageOrderValue: averageOrderValue,
		}

		ctrl.sendSuccessResponse(c, vendorAnalysisData, models.Metadata{
			ReportType:     "vendor-analysis",
			GeneratedAt:    time.Now(),
			GeneratedBy:    ctrl.getCurrentUser(c),
			Parameters:     params,
			GenerationTime: time.Since(startTime).String(),
			RecordCount:    len(purchasesByPeriod),
			Version:        "1.0.0",
			Format:         format,
		})
	} else {
		ctrl.sendErrorResponse(c, http.StatusNotImplemented, "NOT_IMPLEMENTED", 
			fmt.Sprintf("Vendor analysis in %s format is not yet implemented", format), nil)
	}
}

// Helper methods
func (ctrl *UnifiedReportController) isValidReportType(reportType string) bool {
	validTypes := ctrl.getValidReportTypes()
	for _, validType := range validTypes {
		if validType == reportType {
			return true
		}
	}
	return false
}

func (ctrl *UnifiedReportController) getValidReportTypes() []string {
	return []string{
		"balance-sheet",
		"profit-loss", 
		"cash-flow",
		"trial-balance",
		"general-ledger",
		"sales-summary",
		"vendor-analysis",
	}
}

func (ctrl *UnifiedReportController) parseReportParameters(c *gin.Context, reportType string) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	
	// Default format
	format := c.Query("format")
	if format == "" {
		format = "json"
	}
	params["format"] = format

	// Parse parameters based on report type
	switch reportType {
	case "balance-sheet", "trial-balance":
		asOfDateStr := c.Query("as_of_date")
		if asOfDateStr == "" {
			params["as_of_date"] = time.Now()
		} else {
			asOfDate, err := time.Parse("2006-01-02", asOfDateStr)
			if err != nil {
				return nil, fmt.Errorf("invalid as_of_date format. Use YYYY-MM-DD")
			}
			params["as_of_date"] = asOfDate
		}

	case "profit-loss", "cash-flow", "sales-summary", "vendor-analysis", "general-ledger":
		startDateStr := c.Query("start_date")
		endDateStr := c.Query("end_date")

		if startDateStr == "" || endDateStr == "" {
			return nil, fmt.Errorf("start_date and end_date are required for %s report", reportType)
		}

		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format. Use YYYY-MM-DD")
		}

		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format. Use YYYY-MM-DD")
		}

		if startDate.After(endDate) {
			return nil, fmt.Errorf("start_date cannot be after end_date")
		}

		params["start_date"] = startDate
		params["end_date"] = endDate

		// Additional parameters for specific reports
		if reportType == "sales-summary" || reportType == "vendor-analysis" {
			groupBy := c.Query("group_by")
			if groupBy == "" {
				groupBy = "month"
			}
			validGroupBy := []string{"day", "week", "month", "quarter", "year"}
			isValidGroupBy := false
			for _, valid := range validGroupBy {
				if groupBy == valid {
					isValidGroupBy = true
					break
				}
			}
			if !isValidGroupBy {
				return nil, fmt.Errorf("invalid group_by value. Valid options: %v", validGroupBy)
			}
			params["group_by"] = groupBy
		}

		if reportType == "general-ledger" {
			accountCode := c.Query("account_code")
			if accountCode != "" {
				params["account_code"] = accountCode
			}
		}

		if reportType == "vendor-analysis" {
			vendorID := c.Query("vendor_id")
			if vendorID != "" {
				vendorIDInt, err := strconv.Atoi(vendorID)
				if err != nil {
					return nil, fmt.Errorf("invalid vendor_id format")
				}
				params["vendor_id"] = vendorIDInt
			}
		}
	}

	return params, nil
}

// Response helpers
func (ctrl *UnifiedReportController) sendSuccessResponse(c *gin.Context, data interface{}, metadata models.Metadata) {
	c.JSON(http.StatusOK, models.StandardReportResponse{
		Success:   true,
		Data:      data,
		Metadata:  metadata,
		Timestamp: time.Now(),
	})
}

func (ctrl *UnifiedReportController) sendErrorResponse(c *gin.Context, statusCode int, errorCode, message string, details map[string]interface{}) {
	c.JSON(statusCode, models.StandardReportResponse{
		Success: false,
		Error: &models.ErrorInfo{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	})
}

func (ctrl *UnifiedReportController) sendFileResponse(c *gin.Context, fileData []byte, fileName, format string) {
	var contentType string
	switch format {
	case "pdf":
		contentType = "application/pdf"
	case "excel":
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "csv":
		contentType = "text/csv"
	default:
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Length", fmt.Sprintf("%d", len(fileData)))
	c.Data(http.StatusOK, contentType, fileData)
}

// Data helper methods
func (ctrl *UnifiedReportController) calculateAccountBalance(accountID uint, asOfDate time.Time) float64 {
	var account models.Account
	if err := ctrl.db.First(&account, accountID).Error; err != nil {
		return 0
	}

	var totalDebits, totalCredits float64
	ctrl.db.Table("journal_entries").
		Joins("JOIN journals ON journal_entries.journal_id = journals.id").
		Where("journal_entries.account_id = ? AND journals.date <= ? AND journals.status = ?", 
			accountID, asOfDate, models.JournalStatusPosted).
		Select("COALESCE(SUM(journal_entries.debit_amount), 0) as total_debits, COALESCE(SUM(journal_entries.credit_amount), 0) as total_credits").
		Row().Scan(&totalDebits, &totalCredits)

	// Apply normal balance rules
	switch account.Type {
	case models.AccountTypeAsset, models.AccountTypeExpense:
		return account.Balance + totalDebits - totalCredits
	case models.AccountTypeLiability, models.AccountTypeEquity, models.AccountTypeRevenue:
		return account.Balance + totalCredits - totalDebits
	default:
		return account.Balance
	}
}

func (ctrl *UnifiedReportController) getSalesByPeriod(ctx context.Context, startDate, endDate time.Time, groupBy string) ([]models.PeriodData, error) {
	var results []models.PeriodData
	
	// Build date format based on groupBy parameter
	var dateFormat string
	var truncFunc string
	
	switch groupBy {
	case "day":
		dateFormat = "2006-01-02"
		truncFunc = "DATE(date)"
	case "week":
		dateFormat = "2006-W02"
		truncFunc = "DATE_TRUNC('week', date)"
	case "quarter":
		dateFormat = "2006-Q1"
		truncFunc = "DATE_TRUNC('quarter', date)"
	case "year":
		dateFormat = "2006"
		truncFunc = "DATE_TRUNC('year', date)"
	default: // month
		dateFormat = "2006-01"
		truncFunc = "DATE_TRUNC('month', date)"
	}
	
	// Query sales data grouped by period
	rows, err := ctrl.db.Raw(`
		SELECT 
			` + truncFunc + ` as period_date,
			COUNT(*) as count,
			COALESCE(SUM(total_amount), 0) as amount
		FROM sales 
		WHERE date BETWEEN ? AND ?
		  AND status IN ('INVOICED', 'PAID', 'COMPLETED')
		GROUP BY ` + truncFunc + `
		ORDER BY ` + truncFunc + `
	`, startDate, endDate).Rows()
	
	if err != nil {
		return nil, fmt.Errorf("failed to query sales data: %v", err)
	}
	defer rows.Close()
	
	// Process results
	for rows.Next() {
		var periodDate time.Time
		var count int
		var amount float64
		
		if err := rows.Scan(&periodDate, &count, &amount); err != nil {
			continue
		}
		
		results = append(results, models.PeriodData{
			Period: periodDate.Format(dateFormat),
			Amount: amount,
			Count:  count,
		})
	}
	
	return results, nil
}

func (ctrl *UnifiedReportController) getPurchasesByPeriod(ctx context.Context, startDate, endDate time.Time, groupBy string) ([]models.PeriodData, error) {
	var results []models.PeriodData
	
	// Build date format based on groupBy parameter
	var dateFormat string
	var truncFunc string
	
	switch groupBy {
	case "day":
		dateFormat = "2006-01-02"
		truncFunc = "DATE(date)"
	case "week":
		dateFormat = "2006-W02"
		truncFunc = "DATE_TRUNC('week', date)"
	case "quarter":
		dateFormat = "2006-Q1"
		truncFunc = "DATE_TRUNC('quarter', date)"
	case "year":
		dateFormat = "2006"
		truncFunc = "DATE_TRUNC('year', date)"
	default: // month
		dateFormat = "2006-01"
		truncFunc = "DATE_TRUNC('month', date)"
	}
	
	// Query purchase data grouped by period
	rows, err := ctrl.db.Raw(`
		SELECT 
			` + truncFunc + ` as period_date,
			COUNT(*) as count,
			COALESCE(SUM(total_amount), 0) as amount
		FROM purchases 
		WHERE date BETWEEN ? AND ?
		  AND status IN ('APPROVED', 'RECEIVED', 'COMPLETED')
		GROUP BY ` + truncFunc + `
		ORDER BY ` + truncFunc + `
	`, startDate, endDate).Rows()
	
	if err != nil {
		return nil, fmt.Errorf("failed to query purchase data: %v", err)
	}
	defer rows.Close()
	
	// Process results
	for rows.Next() {
		var periodDate time.Time
		var count int
		var amount float64
		
		if err := rows.Scan(&periodDate, &count, &amount); err != nil {
			continue
		}
		
		results = append(results, models.PeriodData{
			Period: periodDate.Format(dateFormat),
			Amount: amount,
			Count:  count,
		})
	}
	
	return results, nil
}

func (ctrl *UnifiedReportController) getCompanyInfo() models.CompanyInfo {
	var profile models.CompanyProfile
	if err := ctrl.db.First(&profile).Error; err != nil {
		// Default company info
		return models.CompanyInfo{
			Name:       "Your Company Name",
			Address:    "Company Address",
			City:       "City",
			State:      "State",
			PostalCode: "12345",
			Phone:      "+62-21-1234567",
			Email:      "contact@company.com",
			Website:    "www.company.com",
			TaxNumber:  "",
		}
	}

	return models.CompanyInfo{
		Name:       profile.Name,
		Address:    profile.Address,
		City:       profile.City,
		State:      profile.State,
		PostalCode: profile.PostalCode,
		Phone:      profile.Phone,
		Email:      profile.Email,
		Website:    profile.Website,
		TaxNumber:  profile.TaxNumber,
	}
}

func (ctrl *UnifiedReportController) getCurrentUser(c *gin.Context) string {
	if user, exists := c.Get("user"); exists {
		if userModel, ok := user.(models.User); ok {
			return userModel.Username
		}
	}
	return "system"
}

// Cash Flow Helper Methods
func (ctrl *UnifiedReportController) getOperatingActivities(startDate, endDate time.Time) models.ReportSection {
	// Get operating activities from journal entries
	// This includes revenue and expense accounts
	var items []models.ReportItem
	
	// Query for revenue accounts (operating income)
	rows, err := ctrl.db.Raw(`
		SELECT 
			a.code, a.name,
			COALESCE(SUM(je.credit_amount - je.debit_amount), 0) as amount
		FROM accounts a
		LEFT JOIN journal_entries je ON a.id = je.account_id
		LEFT JOIN journals j ON je.journal_id = j.id
		WHERE a.type = 'REVENUE'
		  AND j.date BETWEEN ? AND ?
		  AND j.status = 'POSTED'
		GROUP BY a.id, a.code, a.name
		HAVING COALESCE(SUM(je.credit_amount - je.debit_amount), 0) != 0
		ORDER BY a.code
	`, startDate, endDate).Rows()
	
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var code, name string
			var amount float64
			rows.Scan(&code, &name, &amount)
			items = append(items, models.ReportItem{
				Name:   fmt.Sprintf("%s - %s", code, name),
				Amount: amount,
			})
		}
	}
	
	// Query for expense accounts (operating expenses) - negative values for cash flow
	rows2, err := ctrl.db.Raw(`
		SELECT 
			a.code, a.name,
			-COALESCE(SUM(je.debit_amount - je.credit_amount), 0) as amount
		FROM accounts a
		LEFT JOIN journal_entries je ON a.id = je.account_id
		LEFT JOIN journals j ON je.journal_id = j.id
		WHERE a.type = 'EXPENSE'
		  AND j.date BETWEEN ? AND ?
		  AND j.status = 'POSTED'
		GROUP BY a.id, a.code, a.name
		HAVING COALESCE(SUM(je.debit_amount - je.credit_amount), 0) != 0
		ORDER BY a.code
	`, startDate, endDate).Rows()
	
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var code, name string
			var amount float64
			rows2.Scan(&code, &name, &amount)
			items = append(items, models.ReportItem{
				Name:   fmt.Sprintf("%s - %s", code, name),
				Amount: amount,
			})
		}
	}
	
	return models.ReportSection{
		Name:     "Operating Activities",
		Items:    items,
		Subtotal: ctrl.calculateSectionTotal(items),
	}
}

func (ctrl *UnifiedReportController) getInvestingActivities(startDate, endDate time.Time) models.ReportSection {
	// Get investing activities (asset purchases/sales, investments)
	var items []models.ReportItem
	
	// Query for asset-related transactions
	rows, err := ctrl.db.Raw(`
		SELECT 
			a.code, a.name,
			COALESCE(SUM(je.debit_amount - je.credit_amount), 0) as amount
		FROM accounts a
		LEFT JOIN journal_entries je ON a.id = je.account_id
		LEFT JOIN journals j ON je.journal_id = j.id
		WHERE a.type = 'ASSET'
		  AND a.code LIKE '15%' -- Fixed assets typically start with 15
		  AND j.date BETWEEN ? AND ?
		  AND j.status = 'POSTED'
		GROUP BY a.id, a.code, a.name
		HAVING COALESCE(SUM(je.debit_amount - je.credit_amount), 0) != 0
		ORDER BY a.code
	`, startDate, endDate).Rows()
	
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var code, name string
			var amount float64
			rows.Scan(&code, &name, &amount)
			// Negative for cash outflows (asset purchases)
			items = append(items, models.ReportItem{
				Name:   fmt.Sprintf("%s - %s", code, name),
				Amount: -amount, // Asset purchases are cash outflows
			})
		}
	}
	
	return models.ReportSection{
		Name:     "Investing Activities",
		Items:    items,
		Subtotal: ctrl.calculateSectionTotal(items),
	}
}

func (ctrl *UnifiedReportController) getFinancingActivities(startDate, endDate time.Time) models.ReportSection {
	// Get financing activities (loans, equity, dividends)
	var items []models.ReportItem
	
	// Query for liability and equity transactions
	rows, err := ctrl.db.Raw(`
		SELECT 
			a.code, a.name,
			COALESCE(SUM(je.credit_amount - je.debit_amount), 0) as amount
		FROM accounts a
		LEFT JOIN journal_entries je ON a.id = je.account_id
		LEFT JOIN journals j ON je.journal_id = j.id
		WHERE (a.type = 'LIABILITY' OR a.type = 'EQUITY')
		  AND a.code NOT LIKE '21%' -- Exclude accounts payable (operating)
		  AND j.date BETWEEN ? AND ?
		  AND j.status = 'POSTED'
		GROUP BY a.id, a.code, a.name
		HAVING COALESCE(SUM(je.credit_amount - je.debit_amount), 0) != 0
		ORDER BY a.code
	`, startDate, endDate).Rows()
	
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var code, name string
			var amount float64
			rows.Scan(&code, &name, &amount)
			items = append(items, models.ReportItem{
				Name:   fmt.Sprintf("%s - %s", code, name),
				Amount: amount,
			})
		}
	}
	
	return models.ReportSection{
		Name:     "Financing Activities",
		Items:    items,
		Subtotal: ctrl.calculateSectionTotal(items),
	}
}

func (ctrl *UnifiedReportController) calculateSectionTotal(items []models.ReportItem) float64 {
	var total float64
	for _, item := range items {
		total += item.Amount
	}
	return total
}

func (ctrl *UnifiedReportController) getCashBalanceAsOf(asOfDate time.Time) float64 {
	// Get cash balance from cash and bank accounts
	var balance float64
	
	ctrl.db.Raw(`
		SELECT COALESCE(SUM(
			a.balance + 
			COALESCE((SELECT SUM(je.debit_amount - je.credit_amount) 
					  FROM journal_entries je 
					  JOIN journals j ON je.journal_id = j.id 
					  WHERE je.account_id = a.id 
					    AND j.date <= ? 
					    AND j.status = 'POSTED'), 0)
		), 0) as total_balance
		FROM accounts a
		WHERE a.type = 'ASSET'
		  AND (a.code LIKE '11%' OR a.code LIKE '12%') -- Cash and bank accounts
		  AND a.is_header = false
	`, asOfDate).Row().Scan(&balance)
	
	return balance
}
