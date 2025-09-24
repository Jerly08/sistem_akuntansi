package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"app-sistem-akuntansi/utils"
)

// OptimizedFinancialReportsController - Unified controller using materialized view
type OptimizedFinancialReportsController struct {
	db *gorm.DB
}

// NewOptimizedFinancialReportsController creates optimized controller
func NewOptimizedFinancialReportsController(db *gorm.DB) *OptimizedFinancialReportsController {
	return &OptimizedFinancialReportsController{db: db}
}

// AccountBalanceData represents materialized view data
type AccountBalanceData struct {
	AccountID         uint64          `json:"account_id" gorm:"column:account_id"`
	AccountCode       string          `json:"account_code" gorm:"column:account_code"`
	AccountName       string          `json:"account_name" gorm:"column:account_name"`
	AccountType       string          `json:"account_type" gorm:"column:account_type"`
	AccountCategory   string          `json:"account_category" gorm:"column:account_category"`
	NormalBalance     string          `json:"normal_balance" gorm:"column:normal_balance"`
	TotalDebits       decimal.Decimal `json:"total_debits" gorm:"column:total_debits"`
	TotalCredits      decimal.Decimal `json:"total_credits" gorm:"column:total_credits"`
	TransactionCount  int64           `json:"transaction_count" gorm:"column:transaction_count"`
	LastTransactionDate *time.Time    `json:"last_transaction_date" gorm:"column:last_transaction_date"`
	CurrentBalance    decimal.Decimal `json:"current_balance" gorm:"column:current_balance"`
	LastUpdated       time.Time       `json:"last_updated" gorm:"column:last_updated"`
	IsActive          bool            `json:"is_active" gorm:"column:is_active"`
	IsHeader          bool            `json:"is_header" gorm:"column:is_header"`
}

// StandardReportResponse - Unified response format
type StandardReportResponse struct {
	Status    string      `json:"status"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Metadata  ReportMetadata `json:"metadata"`
	Generated time.Time   `json:"generated_at"`
}

type ReportMetadata struct {
	Source        string        `json:"source"`
	GenerationMS  int64         `json:"generation_time_ms"`
	DataFreshness string        `json:"data_freshness"`
	RecordCount   int           `json:"record_count"`
	UsesMaterView bool          `json:"uses_materialized_view"`
}

// OPTIMIZED BALANCE SHEET - Uses materialized view directly
// @Summary Generate Optimized Balance Sheet
// @Description Generate balance sheet using materialized view for optimal performance
// @Tags Optimized Reports
// @Accept json
// @Produce json
// @Param as_of_date query string false "As of date (YYYY-MM-DD)"
// @Success 200 {object} StandardReportResponse
// @Router /api/v1/reports/optimized/balance-sheet [get]
func (ctrl *OptimizedFinancialReportsController) GetOptimizedBalanceSheet(c *gin.Context) {
	startTime := time.Now()
	
	// Refresh materialized view first for fresh data
	if err := ctrl.refreshMaterializedView(); err != nil {
		appErr := utils.NewInternalError("Failed to refresh materialized view", err)
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Query materialized view directly - SUPER FAST!
	var accounts []AccountBalanceData
	result := ctrl.db.Table("account_balances").
		Where("is_active = ? AND account_type IN (?)", true, []string{"ASSET", "LIABILITY", "EQUITY"}).
		Order("account_type, account_code").
		Find(&accounts)
		
	if result.Error != nil {
		appErr := utils.NewInternalError("Failed to query account balances", result.Error)
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Categorize accounts
	balanceSheet := ctrl.buildBalanceSheet(accounts)
	
	// Calculate generation time
	generationTime := time.Since(startTime).Milliseconds()

	response := StandardReportResponse{
		Status:  "success",
		Message: "Balance sheet generated successfully using materialized view",
		Data:    balanceSheet,
		Metadata: ReportMetadata{
			Source:        "materialized_view_account_balances",
			GenerationMS:  generationTime,
			DataFreshness: "real_time",
			RecordCount:   len(accounts),
			UsesMaterView: true,
		},
		Generated: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// OPTIMIZED TRIAL BALANCE - Lightning fast with materialized view
// @Summary Generate Optimized Trial Balance  
// @Description Generate trial balance using materialized view for optimal performance
// @Tags Optimized Reports
// @Accept json
// @Produce json
// @Param as_of_date query string false "As of date (YYYY-MM-DD)"
// @Success 200 {object} StandardReportResponse
// @Router /api/v1/reports/optimized/trial-balance [get]
func (ctrl *OptimizedFinancialReportsController) GetOptimizedTrialBalance(c *gin.Context) {
	startTime := time.Now()
	
	// Refresh materialized view
	if err := ctrl.refreshMaterializedView(); err != nil {
		appErr := utils.NewInternalError("Failed to refresh materialized view", err)
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Query ALL account balances from materialized view
	var accounts []AccountBalanceData
	result := ctrl.db.Table("account_balances").
		Where("is_active = ? AND (total_debits > 0 OR total_credits > 0)", true).
		Order("account_code").
		Find(&accounts)
		
	if result.Error != nil {
		appErr := utils.NewInternalError("Failed to query account balances", result.Error)
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Build trial balance
	trialBalance := ctrl.buildTrialBalance(accounts)
	
	generationTime := time.Since(startTime).Milliseconds()

	response := StandardReportResponse{
		Status:  "success", 
		Message: "Trial balance generated successfully using materialized view",
		Data:    trialBalance,
		Metadata: ReportMetadata{
			Source:        "materialized_view_account_balances",
			GenerationMS:  generationTime,
			DataFreshness: "real_time",
			RecordCount:   len(accounts),
			UsesMaterView: true,
		},
		Generated: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// OPTIMIZED PROFIT & LOSS - Fast calculation from materialized view
// @Summary Generate Optimized Profit & Loss
// @Description Generate P&L using materialized view for optimal performance
// @Tags Optimized Reports
// @Accept json  
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} StandardReportResponse
// @Router /api/v1/reports/optimized/profit-loss [get]
func (ctrl *OptimizedFinancialReportsController) GetOptimizedProfitLoss(c *gin.Context) {
	startTime := time.Now()
	
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	
	if startDate == "" || endDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "start_date and end_date are required",
		})
		return
	}

	// Refresh materialized view
	if err := ctrl.refreshMaterializedView(); err != nil {
		appErr := utils.NewInternalError("Failed to refresh materialized view", err)
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Query REVENUE and EXPENSE accounts from materialized view
	var accounts []AccountBalanceData
	result := ctrl.db.Table("account_balances").
		Where("is_active = ? AND account_type IN (?)", true, []string{"REVENUE", "EXPENSE"}).
		Order("account_type, account_code").
		Find(&accounts)
		
	if result.Error != nil {
		appErr := utils.NewInternalError("Failed to query account balances", result.Error)
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}

	// Build P&L statement
	profitLoss := ctrl.buildProfitLoss(accounts, startDate, endDate)
	
	generationTime := time.Since(startTime).Milliseconds()

	response := StandardReportResponse{
		Status:  "success",
		Message: "Profit & Loss generated successfully using materialized view",
		Data:    profitLoss,
		Metadata: ReportMetadata{
			Source:        "materialized_view_account_balances", 
			GenerationMS:  generationTime,
			DataFreshness: "real_time",
			RecordCount:   len(accounts),
			UsesMaterView: true,
		},
		Generated: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// REFRESH MATERIALIZED VIEW - Manual refresh endpoint
// @Summary Refresh Account Balances Materialized View
// @Description Manually refresh the materialized view for updated data
// @Tags Optimized Reports
// @Accept json
// @Produce json
// @Success 200 {object} StandardReportResponse
// @Router /api/v1/reports/optimized/refresh-balances [post]
func (ctrl *OptimizedFinancialReportsController) RefreshAccountBalances(c *gin.Context) {
	startTime := time.Now()
	
	err := ctrl.refreshMaterializedView()
	if err != nil {
		appErr := utils.NewInternalError("Failed to refresh materialized view", err)
		c.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		return
	}
	
	generationTime := time.Since(startTime).Milliseconds()

	response := StandardReportResponse{
		Status:  "success",
		Message: "Account balances materialized view refreshed successfully",
		Data: gin.H{
			"refreshed_at": time.Now(),
			"status":       "completed",
		},
		Metadata: ReportMetadata{
			Source:        "materialized_view_refresh",
			GenerationMS:  generationTime,
			DataFreshness: "just_updated",
			RecordCount:   0,
			UsesMaterView: true,
		},
		Generated: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}

// Helper Methods

func (ctrl *OptimizedFinancialReportsController) refreshMaterializedView() error {
	// Refresh materialized view for latest data
	result := ctrl.db.Exec("REFRESH MATERIALIZED VIEW account_balances")
	return result.Error
}

func (ctrl *OptimizedFinancialReportsController) buildBalanceSheet(accounts []AccountBalanceData) gin.H {
	assets := make([]AccountBalanceData, 0)
	liabilities := make([]AccountBalanceData, 0)
	equity := make([]AccountBalanceData, 0)
	
	var totalAssets, totalLiabilities, totalEquity decimal.Decimal

	for _, account := range accounts {
		switch account.AccountType {
		case "ASSET":
			assets = append(assets, account)
			totalAssets = totalAssets.Add(account.CurrentBalance)
		case "LIABILITY":
			liabilities = append(liabilities, account)
			totalLiabilities = totalLiabilities.Add(account.CurrentBalance)
		case "EQUITY":
			equity = append(equity, account)
			totalEquity = totalEquity.Add(account.CurrentBalance)
		}
	}

	return gin.H{
		"report_title":    "Balance Sheet",
		"generated_at":    time.Now(),
		"assets": gin.H{
			"accounts": assets,
			"total":    totalAssets,
		},
		"liabilities": gin.H{
			"accounts": liabilities,
			"total":    totalLiabilities,
		},
		"equity": gin.H{
			"accounts": equity,
			"total":    totalEquity,
		},
		"total_assets":             totalAssets,
		"total_liabilities_equity": totalLiabilities.Add(totalEquity),
		"is_balanced":              totalAssets.Equal(totalLiabilities.Add(totalEquity)),
	}
}

func (ctrl *OptimizedFinancialReportsController) buildTrialBalance(accounts []AccountBalanceData) gin.H {
	var totalDebits, totalCredits decimal.Decimal
	
	for _, account := range accounts {
		totalDebits = totalDebits.Add(account.TotalDebits)
		totalCredits = totalCredits.Add(account.TotalCredits)
	}

	return gin.H{
		"report_title":   "Trial Balance",
		"generated_at":   time.Now(),
		"accounts":       accounts,
		"total_debits":   totalDebits,
		"total_credits":  totalCredits,
		"is_balanced":    totalDebits.Equal(totalCredits),
		"total_accounts": len(accounts),
	}
}

func (ctrl *OptimizedFinancialReportsController) buildProfitLoss(accounts []AccountBalanceData, startDate, endDate string) gin.H {
	revenues := make([]AccountBalanceData, 0)
	expenses := make([]AccountBalanceData, 0)
	
	var totalRevenue, totalExpenses decimal.Decimal

	for _, account := range accounts {
		if account.AccountType == "REVENUE" {
			revenues = append(revenues, account)
			totalRevenue = totalRevenue.Add(account.CurrentBalance)
		} else if account.AccountType == "EXPENSE" {
			expenses = append(expenses, account)
			totalExpenses = totalExpenses.Add(account.CurrentBalance)
		}
	}

	netIncome := totalRevenue.Sub(totalExpenses)

	return gin.H{
		"report_title": "Profit & Loss Statement",
		"period": gin.H{
			"start_date": startDate,
			"end_date":   endDate,
		},
		"generated_at": time.Now(),
		"revenue": gin.H{
			"accounts": revenues,
			"total":    totalRevenue,
		},
		"expenses": gin.H{
			"accounts": expenses,
			"total":    totalExpenses,
		},
		"total_revenue":  totalRevenue,
		"total_expenses": totalExpenses,
		"net_income":     netIncome,
		"profit_margin":  ctrl.calculateProfitMargin(netIncome, totalRevenue),
	}
}

func (ctrl *OptimizedFinancialReportsController) calculateProfitMargin(netIncome, totalRevenue decimal.Decimal) decimal.Decimal {
	if totalRevenue.IsZero() {
		return decimal.Zero
	}
	return netIncome.Div(totalRevenue).Mul(decimal.NewFromInt(100))
}