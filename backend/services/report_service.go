package services

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type ReportService struct {
	db              *gorm.DB
	accountRepo     repositories.AccountRepository
	salesRepo       *repositories.SalesRepository
	purchaseRepo    *repositories.PurchaseRepository
	productRepo     *repositories.ProductRepository
	contactRepo     repositories.ContactRepository
	paymentRepo     *repositories.PaymentRepository
	cashBankRepo    *repositories.CashBankRepository
}

// Report response structures
type ReportResponse struct {
	ID           string                 `json:"id"`
	Title        string                 `json:"title"`
	Type         string                 `json:"type"`
	Period       string                 `json:"period"`
	GeneratedAt  time.Time              `json:"generated_at"`
	Data         map[string]interface{} `json:"data"`
	FileData     []byte                 `json:"-"` // For PDF/Excel data
	Summary      map[string]float64     `json:"summary"`
	Parameters   map[string]interface{} `json:"parameters"`
}

type BalanceSheetData struct {
	Assets      []ReportAccountBalance   `json:"assets"`
	Liabilities []ReportAccountBalance   `json:"liabilities"`
	Equity      []ReportAccountBalance   `json:"equity"`
	TotalAssets float64           `json:"total_assets"`
	TotalLiabilitiesEquity float64 `json:"total_liabilities_equity"`
	IsBalanced  bool              `json:"is_balanced"`
}

type ProfitLossData struct {
	Revenue       []ReportAccountBalance   `json:"revenue"`
	Expenses      []ReportAccountBalance   `json:"expenses"`
	GrossProfit   float64           `json:"gross_profit"`
	OperatingIncome float64         `json:"operating_income"`
	NetIncome     float64           `json:"net_income"`
	TotalRevenue  float64           `json:"total_revenue"`
	TotalExpenses float64           `json:"total_expenses"`
	PeriodData    []PeriodData      `json:"period_data"`
}

type CashFlowData struct {
	OperatingActivities []CashFlowItem `json:"operating_activities"`
	InvestingActivities []CashFlowItem `json:"investing_activities"`
	FinancingActivities []CashFlowItem `json:"financing_activities"`
	NetCashFlow         float64        `json:"net_cash_flow"`
	BeginningBalance    float64        `json:"beginning_balance"`
	EndingBalance       float64        `json:"ending_balance"`
}

type ReportAccountBalance struct {
	AccountID    uint    `json:"account_id"`
	AccountCode  string  `json:"account_code"`
	AccountName  string  `json:"account_name"`
	AccountType  string  `json:"account_type"`
	Balance      float64 `json:"balance"`
	DebitTotal   float64 `json:"debit_total"`
	CreditTotal  float64 `json:"credit_total"`
	Level        int     `json:"level"`
	IsHeader     bool    `json:"is_header"`
	ParentID     *uint   `json:"parent_id"`
}

type CashFlowItem struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"` // INFLOW, OUTFLOW
}

type PeriodData struct {
	Period string  `json:"period"`
	Amount float64 `json:"amount"`
}

type ReceivableData struct {
	CustomerID      uint      `json:"customer_id"`
	CustomerName    string    `json:"customer_name"`
	InvoiceNumber   string    `json:"invoice_number"`
	SaleID          uint      `json:"sale_id"`
	Date            time.Time `json:"date"`
	DueDate         time.Time `json:"due_date"`
	Amount          float64   `json:"amount"`
	PaidAmount      float64   `json:"paid_amount"`
	Outstanding     float64   `json:"outstanding"`
	DaysOverdue     int       `json:"days_overdue"`
	Status          string    `json:"status"`
}

type PayableData struct {
	VendorID        uint      `json:"vendor_id"`
	VendorName      string    `json:"vendor_name"`
	PurchaseID      uint      `json:"purchase_id"`
	Date            time.Time `json:"date"`
	DueDate         time.Time `json:"due_date"`
	Amount          float64   `json:"amount"`
	PaidAmount      float64   `json:"paid_amount"`
	Outstanding     float64   `json:"outstanding"`
	DaysOverdue     int       `json:"days_overdue"`
	Status          string    `json:"status"`
}

type InventoryReportData struct {
	ProductID       uint    `json:"product_id"`
	ProductCode     string  `json:"product_code"`
	ProductName     string  `json:"product_name"`
	Category        string  `json:"category"`
	CurrentStock    int     `json:"current_stock"`
	UnitCost        float64 `json:"unit_cost"`
	TotalValue      float64 `json:"total_value"`
	MinStock        int     `json:"min_stock"`
	Status          string  `json:"status"` // OK, LOW_STOCK, OUT_OF_STOCK
}

type FinancialRatiosData struct {
	LiquidityRatios      map[string]float64 `json:"liquidity_ratios"`
	ProfitabilityRatios  map[string]float64 `json:"profitability_ratios"`
	LeverageRatios       map[string]float64 `json:"leverage_ratios"`
	EfficiencyRatios     map[string]float64 `json:"efficiency_ratios"`
	MarketRatios         map[string]float64 `json:"market_ratios"`
	Period               string             `json:"period"`
}

type SalesSummaryData struct {
	TotalSales       float64             `json:"total_sales"`
	TotalTransactions int64              `json:"total_transactions"`
	AverageOrderValue float64            `json:"average_order_value"`
	TopCustomers     []CustomerSummary   `json:"top_customers"`
	SalesByPeriod    []PeriodData        `json:"sales_by_period"`
	SalesByProduct   []ProductSummary    `json:"sales_by_product"`
	StatusBreakdown  map[string]int64    `json:"status_breakdown"`
}

type PurchaseSummaryData struct {
	TotalPurchases       float64             `json:"total_purchases"`
	TotalTransactions    int64               `json:"total_transactions"`
	AverageOrderValue    float64             `json:"average_order_value"`
	TopVendors          []VendorSummary     `json:"top_vendors"`
	PurchasesByPeriod   []PeriodData        `json:"purchases_by_period"`
	PurchasesByProduct  []ProductSummary    `json:"purchases_by_product"`
	StatusBreakdown     map[string]int64    `json:"status_breakdown"`
}

type CustomerSummary struct {
	CustomerID   uint    `json:"customer_id"`
	CustomerName string  `json:"customer_name"`
	TotalSales   float64 `json:"total_sales"`
	OrderCount   int64   `json:"order_count"`
}

type VendorSummary struct {
	VendorID       uint    `json:"vendor_id"`
	VendorName     string  `json:"vendor_name"`
	TotalPurchases float64 `json:"total_purchases"`
	OrderCount     int64   `json:"order_count"`
}

type ProductSummary struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int64   `json:"quantity"`
	Amount      float64 `json:"amount"`
}

func NewReportService(db *gorm.DB, accountRepo repositories.AccountRepository, salesRepo *repositories.SalesRepository, purchaseRepo *repositories.PurchaseRepository, productRepo *repositories.ProductRepository, contactRepo repositories.ContactRepository, paymentRepo *repositories.PaymentRepository, cashBankRepo *repositories.CashBankRepository) *ReportService {
	return &ReportService{
		db:           db,
		accountRepo:  accountRepo,
		salesRepo:    salesRepo,
		purchaseRepo: purchaseRepo,
		productRepo:  productRepo,
		contactRepo:  contactRepo,
		paymentRepo:  paymentRepo,
		cashBankRepo: cashBankRepo,
	}
}

// GetAvailableReports returns list of available reports
func (rs *ReportService) GetAvailableReports() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"id":          "balance-sheet",
			"name":        "Balance Sheet",
			"description": "Statement of financial position showing assets, liabilities, and equity",
			"type":        "Financial",
			"category":    "Core Financial Statements",
			"parameters":  []string{"as_of_date", "format"},
		},
		{
			"id":          "profit-loss",
			"name":        "Profit & Loss Statement",
			"description": "Income statement showing revenues and expenses for a period",
			"type":        "Financial",
			"category":    "Core Financial Statements",
			"parameters":  []string{"start_date", "end_date", "format"},
		},
		{
			"id":          "cash-flow",
			"name":        "Cash Flow Statement",
			"description": "Statement of cash flows from operating, investing, and financing activities",
			"type":        "Financial",
			"category":    "Core Financial Statements",
			"parameters":  []string{"start_date", "end_date", "format"},
		},
		{
			"id":          "trial-balance",
			"name":        "Trial Balance",
			"description": "List of all accounts with their debit and credit balances",
			"type":        "Financial",
			"category":    "Supporting Reports",
			"parameters":  []string{"as_of_date", "format"},
		},
		{
			"id":          "general-ledger",
			"name":        "General Ledger",
			"description": "Detailed transaction history for all or specific accounts",
			"type":        "Financial",
			"category":    "Supporting Reports",
			"parameters":  []string{"start_date", "end_date", "account_code", "format"},
		},
		{
			"id":          "accounts-receivable",
			"name":        "Accounts Receivable",
			"description": "Outstanding amounts owed by customers",
			"type":        "Operational",
			"category":    "Receivables & Payables",
			"parameters":  []string{"as_of_date", "customer_id", "format"},
		},
		{
			"id":          "accounts-payable",
			"name":        "Accounts Payable",
			"description": "Outstanding amounts owed to vendors",
			"type":        "Operational",
			"category":    "Receivables & Payables",
			"parameters":  []string{"as_of_date", "vendor_id", "format"},
		},
		{
			"id":          "sales-summary",
			"name":        "Sales Summary",
			"description": "Summary of sales transactions and performance",
			"type":        "Operational",
			"category":    "Sales & Purchase Reports",
			"parameters":  []string{"start_date", "end_date", "group_by", "format"},
		},
		{
			"id":          "purchase-summary",
			"name":        "Purchase Summary",
			"description": "Summary of purchase transactions and vendor performance",
			"type":        "Operational",
			"category":    "Sales & Purchase Reports",
			"parameters":  []string{"start_date", "end_date", "group_by", "format"},
		},
		{
			"id":          "inventory-report",
			"name":        "Inventory Report",
			"description": "Current inventory levels, stock valuation, and status",
			"type":        "Operational",
			"category":    "Inventory Reports",
			"parameters":  []string{"as_of_date", "include_valuation", "format"},
		},
		{
			"id":          "financial-ratios",
			"name":        "Financial Ratios",
			"description": "Key financial ratios for performance analysis",
			"type":        "Analytical",
			"category":    "Analysis Reports",
			"parameters":  []string{"as_of_date", "period", "format"},
		},
	}
}

// GenerateBalanceSheet creates a balance sheet report
func (rs *ReportService) GenerateBalanceSheet(asOfDate time.Time, format string) (*ReportResponse, error) {
	ctx := context.Background()
	
	// Get all accounts with their balances
	accounts, err := rs.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}

	// Calculate account balances as of the specified date
	balanceSheetData := &BalanceSheetData{
		Assets:      []ReportAccountBalance{},
		Liabilities: []ReportAccountBalance{},
		Equity:      []ReportAccountBalance{},
	}

	for _, account := range accounts {
		// Calculate balance as of date (this is simplified - in production you'd sum from journal entries)
		balance := rs.calculateAccountBalance(account.ID, asOfDate)
		
		accountBalance := ReportAccountBalance{
			AccountID:   account.ID,
			AccountCode: account.Code,
			AccountName: account.Name,
			AccountType: account.Type,
			Balance:     balance,
			Level:       account.Level,
			IsHeader:    account.IsHeader,
			ParentID:    account.ParentID,
		}

		switch account.Type {
		case models.AccountTypeAsset:
			balanceSheetData.Assets = append(balanceSheetData.Assets, accountBalance)
			balanceSheetData.TotalAssets += balance
		case models.AccountTypeLiability:
			balanceSheetData.Liabilities = append(balanceSheetData.Liabilities, accountBalance)
			balanceSheetData.TotalLiabilitiesEquity += balance
		case models.AccountTypeEquity:
			balanceSheetData.Equity = append(balanceSheetData.Equity, accountBalance)
			balanceSheetData.TotalLiabilitiesEquity += balance
		}
	}

	balanceSheetData.IsBalanced = (balanceSheetData.TotalAssets == balanceSheetData.TotalLiabilitiesEquity)

	report := &ReportResponse{
		ID:          fmt.Sprintf("balance-sheet-%d", time.Now().Unix()),
		Title:       "Balance Sheet",
		Type:        models.ReportTypeBalanceSheet,
		Period:      asOfDate.Format("2006-01-02"),
		GeneratedAt: time.Now(),
		Data:        map[string]interface{}{"balance_sheet": balanceSheetData},
		Summary: map[string]float64{
			"total_assets":             balanceSheetData.TotalAssets,
			"total_liabilities_equity": balanceSheetData.TotalLiabilitiesEquity,
		},
		Parameters: map[string]interface{}{
			"as_of_date": asOfDate.Format("2006-01-02"),
			"format":     format,
		},
	}

	// Generate PDF or Excel if requested
	if format == "pdf" {
		pdfData, err := rs.generateBalanceSheetPDF(balanceSheetData, asOfDate)
		if err != nil {
			return nil, fmt.Errorf("failed to generate PDF: %v", err)
		}
		report.FileData = pdfData
	} else if format == "excel" {
		excelData, err := rs.generateBalanceSheetExcel(balanceSheetData, asOfDate)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Excel: %v", err)
		}
		report.FileData = excelData
	}

	return report, nil
}

// GenerateProfitLoss creates a profit & loss statement
func (rs *ReportService) GenerateProfitLoss(startDate, endDate time.Time, format string) (*ReportResponse, error) {
	ctx := context.Background()
	
	// Get revenue and expense accounts
	accounts, err := rs.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}

	plData := &ProfitLossData{
		Revenue:     []ReportAccountBalance{},
		Expenses:    []ReportAccountBalance{},
		PeriodData:  []PeriodData{},
	}

	for _, account := range accounts {
		balance := rs.calculateAccountBalanceForPeriod(account.ID, startDate, endDate)
		
		if balance == 0 {
			continue // Skip accounts with no activity
		}

		accountBalance := ReportAccountBalance{
			AccountID:   account.ID,
			AccountCode: account.Code,
			AccountName: account.Name,
			AccountType: account.Type,
			Balance:     balance,
			Level:       account.Level,
			IsHeader:    account.IsHeader,
			ParentID:    account.ParentID,
		}

		switch account.Type {
		case models.AccountTypeRevenue:
			plData.Revenue = append(plData.Revenue, accountBalance)
			plData.TotalRevenue += balance
		case models.AccountTypeExpense:
			plData.Expenses = append(plData.Expenses, accountBalance)
			plData.TotalExpenses += balance
		}
	}

	plData.GrossProfit = plData.TotalRevenue - plData.TotalExpenses
	plData.OperatingIncome = plData.GrossProfit // Simplified
	plData.NetIncome = plData.OperatingIncome   // Simplified

	// Generate period data for charts
	plData.PeriodData = rs.generatePeriodData(startDate, endDate, "month")

	report := &ReportResponse{
		ID:          fmt.Sprintf("profit-loss-%d", time.Now().Unix()),
		Title:       "Profit & Loss Statement",
		Type:        models.ReportTypeIncomeStatement,
		Period:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		GeneratedAt: time.Now(),
		Data:        map[string]interface{}{"profit_loss": plData},
		Summary: map[string]float64{
			"total_revenue":  plData.TotalRevenue,
			"total_expenses": plData.TotalExpenses,
			"gross_profit":   plData.GrossProfit,
			"net_income":     plData.NetIncome,
		},
		Parameters: map[string]interface{}{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
			"format":     format,
		},
	}

	// Generate PDF or Excel if requested
	if format == "pdf" {
		pdfData, err := rs.generateProfitLossPDF(plData, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("failed to generate PDF: %v", err)
		}
		report.FileData = pdfData
	} else if format == "excel" {
		excelData, err := rs.generateProfitLossExcel(plData, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Excel: %v", err)
		}
		report.FileData = excelData
	}

	return report, nil
}

// GenerateCashFlow creates a cash flow statement
func (rs *ReportService) GenerateCashFlow(startDate, endDate time.Time, format string) (*ReportResponse, error) {
	// Get cash and bank account balances
	beginningBalance := rs.calculateCashBalance(startDate.AddDate(0, 0, -1))
	endingBalance := rs.calculateCashBalance(endDate)
	
	// Get operating cash flows
	operatingActivities := rs.calculateOperatingCashFlows(startDate, endDate)
	
	// Get investing cash flows  
	investingActivities := rs.calculateInvestingCashFlows(startDate, endDate)
	
	// Get financing cash flows
	financingActivities := rs.calculateFinancingCashFlows(startDate, endDate)
	
	cashFlowData := &CashFlowData{
		OperatingActivities: operatingActivities,
		InvestingActivities: investingActivities,
		FinancingActivities: financingActivities,
		BeginningBalance: beginningBalance,
		EndingBalance: endingBalance,
	}

	// Calculate net cash flow
	netOperating := rs.calculateNetCashFlow(cashFlowData.OperatingActivities)
	netInvesting := rs.calculateNetCashFlow(cashFlowData.InvestingActivities)
	netFinancing := rs.calculateNetCashFlow(cashFlowData.FinancingActivities)
	
	cashFlowData.NetCashFlow = netOperating + netInvesting + netFinancing

	report := &ReportResponse{
		ID:          fmt.Sprintf("cash-flow-%d", time.Now().Unix()),
		Title:       "Cash Flow Statement",
		Type:        models.ReportTypeCashFlow,
		Period:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		GeneratedAt: time.Now(),
		Data:        map[string]interface{}{"cash_flow": cashFlowData},
		Summary: map[string]float64{
			"net_operating":     netOperating,
			"net_investing":     netInvesting,
			"net_financing":     netFinancing,
			"net_cash_flow":     cashFlowData.NetCashFlow,
			"ending_balance":    cashFlowData.EndingBalance,
		},
		Parameters: map[string]interface{}{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
			"format":     format,
		},
	}

	// Generate PDF or Excel if requested
	if format == "pdf" {
		pdfData, err := rs.generateCashFlowPDF(cashFlowData, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("failed to generate PDF: %v", err)
		}
		report.FileData = pdfData
	} else if format == "excel" {
		excelData, err := rs.generateCashFlowExcel(cashFlowData, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("failed to generate Excel: %v", err)
		}
		report.FileData = excelData
	}

	return report, nil
}

// GenerateTrialBalance creates a trial balance report
func (rs *ReportService) GenerateTrialBalance(asOfDate time.Time, format string) (*ReportResponse, error) {
	ctx := context.Background()
	
	accounts, err := rs.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}

	var trialBalanceData []ReportAccountBalance
	var totalDebits, totalCredits float64

	for _, account := range accounts {
		if account.IsHeader {
			continue // Skip header accounts
		}

		debitTotal, creditTotal := rs.calculateAccountTotals(account.ID, asOfDate)
		balance := rs.calculateAccountBalance(account.ID, asOfDate)

		if debitTotal == 0 && creditTotal == 0 {
			continue // Skip accounts with no activity
		}

		accountBalance := ReportAccountBalance{
			AccountID:    account.ID,
			AccountCode:  account.Code,
			AccountName:  account.Name,
			AccountType:  account.Type,
			Balance:      balance,
			DebitTotal:   debitTotal,
			CreditTotal:  creditTotal,
			Level:        account.Level,
			IsHeader:     account.IsHeader,
			ParentID:     account.ParentID,
		}

		trialBalanceData = append(trialBalanceData, accountBalance)
		totalDebits += debitTotal
		totalCredits += creditTotal
	}

	report := &ReportResponse{
		ID:          fmt.Sprintf("trial-balance-%d", time.Now().Unix()),
		Title:       "Trial Balance",
		Type:        models.ReportTypeTrialBalance,
		Period:      asOfDate.Format("2006-01-02"),
		GeneratedAt: time.Now(),
		Data:        map[string]interface{}{"trial_balance": trialBalanceData},
		Summary: map[string]float64{
			"total_debits":  totalDebits,
			"total_credits": totalCredits,
			"is_balanced":   func() float64 { if totalDebits == totalCredits { return 1 } else { return 0 } }(),
		},
		Parameters: map[string]interface{}{
			"as_of_date": asOfDate.Format("2006-01-02"),
			"format":     format,
		},
	}

	// Generate PDF if requested
	if format == "pdf" {
		pdfData, err := rs.generateTrialBalancePDF(trialBalanceData, asOfDate)
		if err != nil {
			return nil, fmt.Errorf("failed to generate PDF: %v", err)
		}
		report.FileData = pdfData
	}

	return report, nil
}

// GenerateGeneralLedger creates a general ledger report
func (rs *ReportService) GenerateGeneralLedger(startDate, endDate time.Time, accountCode string, format string) (*ReportResponse, error) {
	// This would implement detailed transaction history
	// For now, return a placeholder
	report := &ReportResponse{
		ID:          fmt.Sprintf("general-ledger-%d", time.Now().Unix()),
		Title:       "General Ledger",
		Type:        models.ReportTypeGeneralLedger,
		Period:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		GeneratedAt: time.Now(),
		Data:        map[string]interface{}{"message": "General Ledger report implementation pending"},
		Summary:     map[string]float64{},
		Parameters: map[string]interface{}{
			"start_date":   startDate.Format("2006-01-02"),
			"end_date":     endDate.Format("2006-01-02"),
			"account_code": accountCode,
			"format":       format,
		},
	}

	return report, nil
}

// GenerateAccountsReceivable creates accounts receivable report
func (rs *ReportService) GenerateAccountsReceivable(asOfDate time.Time, customerID *uint, format string) (*ReportResponse, error) {
	var receivables []ReceivableData
	var totalOutstanding float64

	// Query unpaid sales
	query := rs.db.Joins("Customer").Where("sales.outstanding_amount > 0 AND sales.date <= ?", asOfDate)
	if customerID != nil {
		query = query.Where("sales.customer_id = ?", *customerID)
	}

	var sales []models.Sale
	if err := query.Find(&sales).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch receivables: %v", err)
	}

	for _, sale := range sales {
		daysOverdue := 0
		if sale.DueDate.Before(asOfDate) {
			daysOverdue = int(asOfDate.Sub(sale.DueDate).Hours() / 24)
		}

		receivable := ReceivableData{
			CustomerID:      sale.CustomerID,
			CustomerName:    sale.Customer.Name,
			InvoiceNumber:   sale.InvoiceNumber,
			SaleID:          sale.ID,
			Date:            sale.Date,
			DueDate:         sale.DueDate,
			Amount:          sale.TotalAmount,
			PaidAmount:      sale.PaidAmount,
			Outstanding:     sale.OutstandingAmount,
			DaysOverdue:     daysOverdue,
			Status:          sale.Status,
		}

		receivables = append(receivables, receivable)
		totalOutstanding += sale.OutstandingAmount
	}

	report := &ReportResponse{
		ID:          fmt.Sprintf("accounts-receivable-%d", time.Now().Unix()),
		Title:       "Accounts Receivable",
		Type:        models.ReportTypeAccountsReceivable,
		Period:      asOfDate.Format("2006-01-02"),
		GeneratedAt: time.Now(),
		Data:        map[string]interface{}{"receivables": receivables},
		Summary: map[string]float64{
			"total_outstanding": totalOutstanding,
			"count":             float64(len(receivables)),
		},
		Parameters: map[string]interface{}{
			"as_of_date":  asOfDate.Format("2006-01-02"),
			"customer_id": customerID,
			"format":      format,
		},
	}

	// Generate PDF if requested
	if format == "pdf" {
		pdfData, err := rs.generateReceivablesPDF(receivables, asOfDate)
		if err != nil {
			return nil, fmt.Errorf("failed to generate PDF: %v", err)
		}
		report.FileData = pdfData
	}

	return report, nil
}

// GenerateAccountsPayable creates accounts payable report
func (rs *ReportService) GenerateAccountsPayable(asOfDate time.Time, vendorID *uint, format string) (*ReportResponse, error) {
	var payables []PayableData
	var totalOutstanding float64

	// Query unpaid purchases
	query := rs.db.Joins("Vendor").Where("purchases.outstanding_amount > 0 AND purchases.date <= ?", asOfDate)
	if vendorID != nil {
		query = query.Where("purchases.vendor_id = ?", *vendorID)
	}

	var purchases []models.Purchase
	if err := query.Find(&purchases).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch payables: %v", err)
	}

	for _, purchase := range purchases {
		daysOverdue := 0
		if purchase.DueDate.Before(asOfDate) {
			daysOverdue = int(asOfDate.Sub(purchase.DueDate).Hours() / 24)
		}

		payable := PayableData{
			VendorID:      purchase.VendorID,
			VendorName:    purchase.Vendor.Name,
			PurchaseID:    purchase.ID,
			Date:          purchase.Date,
			DueDate:       purchase.DueDate,
			Amount:        purchase.TotalAmount,
			PaidAmount:    purchase.PaidAmount,
			Outstanding:   purchase.OutstandingAmount,
			DaysOverdue:   daysOverdue,
			Status:        purchase.Status,
		}

		payables = append(payables, payable)
		totalOutstanding += purchase.OutstandingAmount
	}

	report := &ReportResponse{
		ID:          fmt.Sprintf("accounts-payable-%d", time.Now().Unix()),
		Title:       "Accounts Payable",
		Type:        models.ReportTypeAccountsPayable,
		Period:      asOfDate.Format("2006-01-02"),
		GeneratedAt: time.Now(),
		Data:        map[string]interface{}{"payables": payables},
		Summary: map[string]float64{
			"total_outstanding": totalOutstanding,
			"count":             float64(len(payables)),
		},
		Parameters: map[string]interface{}{
			"as_of_date": asOfDate.Format("2006-01-02"),
			"vendor_id":  vendorID,
			"format":     format,
		},
	}

	// Generate PDF if requested
	if format == "pdf" {
		pdfData, err := rs.generatePayablesPDF(payables, asOfDate)
		if err != nil {
			return nil, fmt.Errorf("failed to generate PDF: %v", err)
		}
		report.FileData = pdfData
	}

	return report, nil
}

// GenerateSalesSummary creates sales summary report
func (rs *ReportService) GenerateSalesSummary(startDate, endDate time.Time, groupBy, format string) (*ReportResponse, error) {
	// Query sales data
	var sales []models.Sale
	if err := rs.db.Joins("Customer").Where("sales.date BETWEEN ? AND ?", startDate, endDate).Find(&sales).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch sales data: %v", err)
	}

	salesData := &SalesSummaryData{
		SalesByPeriod:   []PeriodData{},
		TopCustomers:    []CustomerSummary{},
		SalesByProduct:  []ProductSummary{},
		StatusBreakdown: make(map[string]int64),
	}

	// Calculate totals
	for _, sale := range sales {
		salesData.TotalSales += sale.TotalAmount
		salesData.TotalTransactions++
		salesData.StatusBreakdown[sale.Status]++
	}

	if salesData.TotalTransactions > 0 {
		salesData.AverageOrderValue = salesData.TotalSales / float64(salesData.TotalTransactions)
	}

	// Generate period data
	salesData.SalesByPeriod = rs.generateSalesPeriodData(sales, startDate, endDate, groupBy)

	// Generate top customers
	salesData.TopCustomers = rs.generateTopCustomers(sales)

	report := &ReportResponse{
		ID:          fmt.Sprintf("sales-summary-%d", time.Now().Unix()),
		Title:       "Sales Summary Report",
		Type:        "SALES_SUMMARY",
		Period:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		GeneratedAt: time.Now(),
		Data:        map[string]interface{}{"sales_summary": salesData},
		Summary: map[string]float64{
			"total_sales":         salesData.TotalSales,
			"total_transactions":  float64(salesData.TotalTransactions),
			"average_order_value": salesData.AverageOrderValue,
		},
		Parameters: map[string]interface{}{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
			"group_by":   groupBy,
			"format":     format,
		},
	}

	return report, nil
}

// GeneratePurchaseSummary creates purchase summary report  
func (rs *ReportService) GeneratePurchaseSummary(startDate, endDate time.Time, groupBy, format string) (*ReportResponse, error) {
	// Query purchase data
	var purchases []models.Purchase
	if err := rs.db.Joins("Vendor").Where("purchases.date BETWEEN ? AND ?", startDate, endDate).Find(&purchases).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch purchase data: %v", err)
	}

	purchaseData := &PurchaseSummaryData{
		PurchasesByPeriod:   []PeriodData{},
		TopVendors:          []VendorSummary{},
		PurchasesByProduct:  []ProductSummary{},
		StatusBreakdown:     make(map[string]int64),
	}

	// Calculate totals
	for _, purchase := range purchases {
		purchaseData.TotalPurchases += purchase.TotalAmount
		purchaseData.TotalTransactions++
		purchaseData.StatusBreakdown[purchase.Status]++
	}

	if purchaseData.TotalTransactions > 0 {
		purchaseData.AverageOrderValue = purchaseData.TotalPurchases / float64(purchaseData.TotalTransactions)
	}

	report := &ReportResponse{
		ID:          fmt.Sprintf("purchase-summary-%d", time.Now().Unix()),
		Title:       "Purchase Summary Report",
		Type:        "PURCHASE_SUMMARY",
		Period:      fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		GeneratedAt: time.Now(),
		Data:        map[string]interface{}{"purchase_summary": purchaseData},
		Summary: map[string]float64{
			"total_purchases":     purchaseData.TotalPurchases,
			"total_transactions":  float64(purchaseData.TotalTransactions),
			"average_order_value": purchaseData.AverageOrderValue,
		},
		Parameters: map[string]interface{}{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
			"group_by":   groupBy,
			"format":     format,
		},
	}

	return report, nil
}

// GenerateInventoryReport creates inventory report
func (rs *ReportService) GenerateInventoryReport(asOfDate time.Time, includeValuation bool, format string) (*ReportResponse, error) {
	var products []models.Product
	if err := rs.db.Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch products: %v", err)
	}

	var inventoryData []InventoryReportData
	var totalValue float64

	for _, product := range products {
		status := "OK"
		if product.Stock <= 0 {
			status = "OUT_OF_STOCK"
		} else if product.Stock <= product.MinStock {
			status = "LOW_STOCK"
		}

		itemValue := float64(product.Stock) * product.PurchasePrice
		
		category := ""
		if product.Category != nil {
			category = product.Category.Name
		}
		
		inventoryItem := InventoryReportData{
			ProductID:    product.ID,
			ProductCode:  product.Code,
			ProductName:  product.Name,
			Category:     category,
			CurrentStock: product.Stock,
			UnitCost:     product.PurchasePrice,
			TotalValue:   itemValue,
			MinStock:     product.MinStock,
			Status:       status,
		}

		inventoryData = append(inventoryData, inventoryItem)
		if includeValuation {
			totalValue += itemValue
		}
	}

	report := &ReportResponse{
		ID:          fmt.Sprintf("inventory-report-%d", time.Now().Unix()),
		Title:       "Inventory Report",
		Type:        models.ReportTypeInventory,
		Period:      asOfDate.Format("2006-01-02"),
		GeneratedAt: time.Now(),
		Data:        map[string]interface{}{"inventory": inventoryData},
		Summary: map[string]float64{
			"total_items":  float64(len(inventoryData)),
			"total_value":  totalValue,
		},
		Parameters: map[string]interface{}{
			"as_of_date":        asOfDate.Format("2006-01-02"),
			"include_valuation": includeValuation,
			"format":            format,
		},
	}

	return report, nil
}

// GenerateFinancialRatios creates financial ratios analysis
func (rs *ReportService) GenerateFinancialRatios(asOfDate time.Time, period, format string) (*ReportResponse, error) {
	// This would implement comprehensive ratio analysis
	// For now, return placeholder ratios
	ratiosData := &FinancialRatiosData{
		LiquidityRatios: map[string]float64{
			"current_ratio":  1.5,
			"quick_ratio":    1.2,
			"cash_ratio":     0.8,
		},
		ProfitabilityRatios: map[string]float64{
			"gross_profit_margin":  0.25,
			"operating_margin":     0.15,
			"net_profit_margin":    0.12,
			"return_on_assets":     0.10,
			"return_on_equity":     0.18,
		},
		LeverageRatios: map[string]float64{
			"debt_to_equity":      0.60,
			"debt_to_assets":      0.40,
			"interest_coverage":   5.5,
		},
		EfficiencyRatios: map[string]float64{
			"inventory_turnover":    6.0,
			"receivables_turnover":  8.0,
			"payables_turnover":     10.0,
			"asset_turnover":        1.2,
		},
		MarketRatios: map[string]float64{
			"price_to_earnings":     15.0,
			"price_to_book":         2.5,
		},
		Period: period,
	}

	report := &ReportResponse{
		ID:          fmt.Sprintf("financial-ratios-%d", time.Now().Unix()),
		Title:       "Financial Ratios Analysis",
		Type:        "FINANCIAL_RATIOS",
		Period:      asOfDate.Format("2006-01-02"),
		GeneratedAt: time.Now(),
		Data:        map[string]interface{}{"financial_ratios": ratiosData},
		Summary:     map[string]float64{},
		Parameters: map[string]interface{}{
			"as_of_date": asOfDate.Format("2006-01-02"),
			"period":     period,
			"format":     format,
		},
	}

	return report, nil
}

// SaveReportTemplate saves a custom report template
func (rs *ReportService) SaveReportTemplate(request *models.ReportTemplateRequest, userID uint) (*models.ReportTemplate, error) {
	template := &models.ReportTemplate{
		Name:        request.Name,
		Type:        request.Type,
		Description: request.Description,
		Template:    request.Template,
		IsDefault:   request.IsDefault,
		IsActive:    true,
		UserID:      userID,
	}

	if err := rs.db.Create(template).Error; err != nil {
		return nil, fmt.Errorf("failed to save report template: %v", err)
	}

	return template, nil
}

// GetReportTemplates returns available report templates
func (rs *ReportService) GetReportTemplates() ([]models.ReportTemplate, error) {
	var templates []models.ReportTemplate
	if err := rs.db.Where("is_active = ?", true).Find(&templates).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch report templates: %v", err)
	}

	return templates, nil
}

// Helper methods

func (rs *ReportService) calculateAccountBalance(accountID uint, asOfDate time.Time) float64 {
	// Get account starting balance (opening balance)
	var account models.Account
	if err := rs.db.First(&account, accountID).Error; err != nil {
		return 0
	}

	// Calculate balance from journal entries up to asOfDate
	var totalDebits, totalCredits float64
	rs.db.Table("journal_entries").
		Joins("JOIN journals ON journal_entries.journal_id = journals.id").
		Where("journal_entries.account_id = ? AND journals.date <= ? AND journals.status = 'POSTED'", accountID, asOfDate).
		Select("COALESCE(SUM(journal_entries.debit_amount), 0) as total_debits, COALESCE(SUM(journal_entries.credit_amount), 0) as total_credits").
		Row().Scan(&totalDebits, &totalCredits)

	// Calculate balance based on account normal balance type
	switch account.Type {
	case models.AccountTypeAsset, models.AccountTypeExpense:
		// Assets and Expenses: Debit increases, Credit decreases
		return account.Balance + totalDebits - totalCredits
	case models.AccountTypeLiability, models.AccountTypeEquity, models.AccountTypeRevenue:
		// Liabilities, Equity, Revenue: Credit increases, Debit decreases
		return account.Balance + totalCredits - totalDebits
	default:
		return account.Balance
	}
}

func (rs *ReportService) calculateAccountBalanceForPeriod(accountID uint, startDate, endDate time.Time) float64 {
	// Calculate balance from journal entries for the specific period only
	var account models.Account
	if err := rs.db.First(&account, accountID).Error; err != nil {
		return 0
	}

	var totalDebits, totalCredits float64
	rs.db.Table("journal_entries").
		Joins("JOIN journals ON journal_entries.journal_id = journals.id").
		Where("journal_entries.account_id = ? AND journals.date BETWEEN ? AND ? AND journals.status = 'POSTED'", accountID, startDate, endDate).
		Select("COALESCE(SUM(journal_entries.debit_amount), 0) as total_debits, COALESCE(SUM(journal_entries.credit_amount), 0) as total_credits").
		Row().Scan(&totalDebits, &totalCredits)

	// For P&L accounts, we want the period activity
	switch account.Type {
	case models.AccountTypeRevenue:
		// Revenue: Credit is positive
		return totalCredits - totalDebits
	case models.AccountTypeExpense:
		// Expenses: Debit is positive
		return totalDebits - totalCredits
	default:
		// For balance sheet accounts, return cumulative balance
		return rs.calculateAccountBalance(accountID, endDate)
	}
}

func (rs *ReportService) calculateAccountTotals(accountID uint, asOfDate time.Time) (debitTotal, creditTotal float64) {
	// Calculate total debits and credits from journal entries
	rs.db.Table("journal_entries").
		Joins("JOIN journals ON journal_entries.journal_id = journals.id").
		Where("journal_entries.account_id = ? AND journals.date <= ? AND journals.status = 'POSTED'", accountID, asOfDate).
		Select("COALESCE(SUM(journal_entries.debit_amount), 0) as total_debits, COALESCE(SUM(journal_entries.credit_amount), 0) as total_credits").
		Row().Scan(&debitTotal, &creditTotal)

	return debitTotal, creditTotal
}

func (rs *ReportService) calculateNetCashFlow(items []CashFlowItem) float64 {
	var total float64
	for _, item := range items {
		total += item.Amount
	}
	return total
}

func (rs *ReportService) generatePeriodData(startDate, endDate time.Time, groupBy string) []PeriodData {
	// This would generate period-based data for charts
	// For now, return sample data
	return []PeriodData{
		{Period: startDate.Format("2006-01"), Amount: 100000},
		{Period: endDate.Format("2006-01"), Amount: 120000},
	}
}

func (rs *ReportService) generateSalesPeriodData(sales []models.Sale, startDate, endDate time.Time, groupBy string) []PeriodData {
	// Group sales by period
	periodMap := make(map[string]float64)
	
	for _, sale := range sales {
		var period string
		switch groupBy {
		case "month":
			period = sale.Date.Format("2006-01")
		case "quarter":
			period = fmt.Sprintf("%d-Q%d", sale.Date.Year(), (sale.Date.Month()-1)/3+1)
		case "year":
			period = sale.Date.Format("2006")
		default:
			period = sale.Date.Format("2006-01")
		}
		
		periodMap[period] += sale.TotalAmount
	}
	
	var result []PeriodData
	for period, amount := range periodMap {
		result = append(result, PeriodData{Period: period, Amount: amount})
	}
	
	return result
}

func (rs *ReportService) generateTopCustomers(sales []models.Sale) []CustomerSummary {
	customerMap := make(map[uint]CustomerSummary)
	
	for _, sale := range sales {
		summary, exists := customerMap[sale.CustomerID]
		if !exists {
			summary = CustomerSummary{
				CustomerID:   sale.CustomerID,
				CustomerName: sale.Customer.Name,
				TotalSales:   0,
				OrderCount:   0,
			}
		}
		
		summary.TotalSales += sale.TotalAmount
		summary.OrderCount++
		customerMap[sale.CustomerID] = summary
	}
	
	var result []CustomerSummary
	for _, summary := range customerMap {
		result = append(result, summary)
	}
	
	return result
}

// Cash flow calculation helper methods

// calculateCashBalance calculates total cash and bank balance at a specific date
func (rs *ReportService) calculateCashBalance(asOfDate time.Time) float64 {
	ctx := context.Background()
	
	// Get all cash and bank accounts
	accounts, err := rs.accountRepo.FindAll(ctx)
	if err != nil {
		return 0
	}

	var totalCash float64
	for _, account := range accounts {
		// Check if account is cash or bank account (by code prefix or type)
		if rs.isCashOrBankAccount(account) {
			balance := rs.calculateAccountBalance(account.ID, asOfDate)
			totalCash += balance
		}
	}
	
	return totalCash
}

// isCashOrBankAccount checks if an account is a cash or bank account
func (rs *ReportService) isCashOrBankAccount(account models.Account) bool {
	// Common cash/bank account codes start with 1-1 (assets) and contain cash/bank keywords
	cashKeywords := []string{"cash", "bank", "checking", "savings", "petty"}
	accountNameLower := strings.ToLower(account.Name)
	accountCodeLower := strings.ToLower(account.Code)
	
	// Check if account type is asset and contains cash/bank keywords
	if account.Type != models.AccountTypeAsset {
		return false
	}
	
	for _, keyword := range cashKeywords {
		if strings.Contains(accountNameLower, keyword) || strings.Contains(accountCodeLower, keyword) {
			return true
		}
	}
	
	// Also check if account code starts with typical cash/bank prefixes
	if strings.HasPrefix(account.Code, "1-1-001") || strings.HasPrefix(account.Code, "1-1-002") {
		return true
	}
	
	return false
}

// calculateOperatingCashFlows calculates cash flows from operating activities
func (rs *ReportService) calculateOperatingCashFlows(startDate, endDate time.Time) []CashFlowItem {
	var items []CashFlowItem
	
	// Calculate cash receipts from customers (collections from accounts receivable)
	collections := rs.calculateCustomerCollections(startDate, endDate)
	if collections != 0 {
		items = append(items, CashFlowItem{
			Description: "Collections from customers",
			Amount:      collections,
			Type:        "INFLOW",
		})
	}
	
	// Calculate cash payments to suppliers
	supplierPayments := rs.calculateSupplierPayments(startDate, endDate)
	if supplierPayments != 0 {
		items = append(items, CashFlowItem{
			Description: "Payments to suppliers",
			Amount:      -supplierPayments, // Negative for outflow
			Type:        "OUTFLOW",
		})
	}
	
	// Calculate employee salary payments
	salaryPayments := rs.calculateSalaryPayments(startDate, endDate)
	if salaryPayments != 0 {
		items = append(items, CashFlowItem{
			Description: "Employee salaries and benefits",
			Amount:      -salaryPayments, // Negative for outflow
			Type:        "OUTFLOW",
		})
	}
	
	// Calculate other operating expense payments
	operatingExpenses := rs.calculateOperatingExpensePayments(startDate, endDate)
	if operatingExpenses != 0 {
		items = append(items, CashFlowItem{
			Description: "Operating expenses",
			Amount:      -operatingExpenses, // Negative for outflow
			Type:        "OUTFLOW",
		})
	}
	
	return items
}

// calculateInvestingCashFlows calculates cash flows from investing activities
func (rs *ReportService) calculateInvestingCashFlows(startDate, endDate time.Time) []CashFlowItem {
	var items []CashFlowItem
	
	// Calculate equipment and asset purchases
	assetPurchases := rs.calculateAssetPurchases(startDate, endDate)
	if assetPurchases != 0 {
		items = append(items, CashFlowItem{
			Description: "Equipment and asset purchases",
			Amount:      -assetPurchases, // Negative for outflow
			Type:        "OUTFLOW",
		})
	}
	
	// Calculate asset sales/disposals
	assetSales := rs.calculateAssetSales(startDate, endDate)
	if assetSales != 0 {
		items = append(items, CashFlowItem{
			Description: "Proceeds from asset sales",
			Amount:      assetSales,
			Type:        "INFLOW",
		})
	}
	
	return items
}

// calculateFinancingCashFlows calculates cash flows from financing activities
func (rs *ReportService) calculateFinancingCashFlows(startDate, endDate time.Time) []CashFlowItem {
	var items []CashFlowItem
	
	// Calculate loan proceeds
	loanProceeds := rs.calculateLoanProceeds(startDate, endDate)
	if loanProceeds != 0 {
		items = append(items, CashFlowItem{
			Description: "Loan proceeds",
			Amount:      loanProceeds,
			Type:        "INFLOW",
		})
	}
	
	// Calculate loan repayments
	loanRepayments := rs.calculateLoanRepayments(startDate, endDate)
	if loanRepayments != 0 {
		items = append(items, CashFlowItem{
			Description: "Loan repayments",
			Amount:      -loanRepayments, // Negative for outflow
			Type:        "OUTFLOW",
		})
	}
	
	// Calculate owner investments/contributions
	ownerInvestments := rs.calculateOwnerInvestments(startDate, endDate)
	if ownerInvestments != 0 {
		items = append(items, CashFlowItem{
			Description: "Owner investments",
			Amount:      ownerInvestments,
			Type:        "INFLOW",
		})
	}
	
	// Calculate dividends/owner withdrawals
	dividendPayments := rs.calculateDividendPayments(startDate, endDate)
	if dividendPayments != 0 {
		items = append(items, CashFlowItem{
			Description: "Dividends and withdrawals",
			Amount:      -dividendPayments, // Negative for outflow
			Type:        "OUTFLOW",
		})
	}
	
	return items
}

// Helper methods for cash flow calculations

func (rs *ReportService) calculateCustomerCollections(startDate, endDate time.Time) float64 {
	// Sum up payments received from customers (credits to accounts receivable)
	var total float64
	
	// Look for accounts receivable account and sum credits (collections)
	ctx := context.Background()
	accounts, _ := rs.accountRepo.FindAll(ctx)
	
	for _, account := range accounts {
		if rs.isAccountsReceivableAccount(account) {
			// Credits to A/R represent collections
			_, credits := rs.calculateAccountTotalsForPeriod(account.ID, startDate, endDate)
			total += credits
		}
	}
	
	return total
}

func (rs *ReportService) calculateSupplierPayments(startDate, endDate time.Time) float64 {
	// Sum up payments made to suppliers (debits to accounts payable)
	var total float64
	
	ctx := context.Background()
	accounts, _ := rs.accountRepo.FindAll(ctx)
	
	for _, account := range accounts {
		if rs.isAccountsPayableAccount(account) {
			// Debits to A/P represent payments
			debits, _ := rs.calculateAccountTotalsForPeriod(account.ID, startDate, endDate)
			total += debits
		}
	}
	
	return total
}

func (rs *ReportService) calculateSalaryPayments(startDate, endDate time.Time) float64 {
	// Sum up salary and payroll related payments
	var total float64
	
	ctx := context.Background()
	accounts, _ := rs.accountRepo.FindAll(ctx)
	
	for _, account := range accounts {
		if rs.isSalaryExpenseAccount(account) {
			// For expense accounts, debits represent the expense
			debits, _ := rs.calculateAccountTotalsForPeriod(account.ID, startDate, endDate)
			total += debits
		}
	}
	
	return total
}

func (rs *ReportService) calculateOperatingExpensePayments(startDate, endDate time.Time) float64 {
	// Sum up other operating expense payments (excluding salaries)
	var total float64
	
	ctx := context.Background()
	accounts, _ := rs.accountRepo.FindAll(ctx)
	
	for _, account := range accounts {
		if rs.isOperatingExpenseAccount(account) && !rs.isSalaryExpenseAccount(account) {
			debits, _ := rs.calculateAccountTotalsForPeriod(account.ID, startDate, endDate)
			total += debits
		}
	}
	
	return total
}

func (rs *ReportService) calculateAssetPurchases(startDate, endDate time.Time) float64 {
	// Sum up fixed asset purchases
	var total float64
	
	ctx := context.Background()
	accounts, _ := rs.accountRepo.FindAll(ctx)
	
	for _, account := range accounts {
		if rs.isFixedAssetAccount(account) {
			// Debits to fixed assets represent purchases
			debits, _ := rs.calculateAccountTotalsForPeriod(account.ID, startDate, endDate)
			total += debits
		}
	}
	
	return total
}

func (rs *ReportService) calculateAssetSales(startDate, endDate time.Time) float64 {
	// This would require analyzing asset disposal transactions
	// For now, return 0 as this requires more complex transaction analysis
	return 0
}

func (rs *ReportService) calculateLoanProceeds(startDate, endDate time.Time) float64 {
	// Sum up loan proceeds (credits to loan liability accounts)
	var total float64
	
	ctx := context.Background()
	accounts, _ := rs.accountRepo.FindAll(ctx)
	
	for _, account := range accounts {
		if rs.isLoanLiabilityAccount(account) {
			// Credits to loan accounts represent new borrowing
			_, credits := rs.calculateAccountTotalsForPeriod(account.ID, startDate, endDate)
			total += credits
		}
	}
	
	return total
}

func (rs *ReportService) calculateLoanRepayments(startDate, endDate time.Time) float64 {
	// Sum up loan repayments (debits to loan liability accounts)
	var total float64
	
	ctx := context.Background()
	accounts, _ := rs.accountRepo.FindAll(ctx)
	
	for _, account := range accounts {
		if rs.isLoanLiabilityAccount(account) {
			// Debits to loan accounts represent repayments
			debits, _ := rs.calculateAccountTotalsForPeriod(account.ID, startDate, endDate)
			total += debits
		}
	}
	
	return total
}

func (rs *ReportService) calculateOwnerInvestments(startDate, endDate time.Time) float64 {
	// Sum up owner capital contributions
	var total float64
	
	ctx := context.Background()
	accounts, _ := rs.accountRepo.FindAll(ctx)
	
	for _, account := range accounts {
		if rs.isOwnerEquityAccount(account) {
			// Credits to equity represent investments
			_, credits := rs.calculateAccountTotalsForPeriod(account.ID, startDate, endDate)
			total += credits
		}
	}
	
	return total
}

func (rs *ReportService) calculateDividendPayments(startDate, endDate time.Time) float64 {
	// Sum up dividend and withdrawal payments
	var total float64
	
	ctx := context.Background()
	accounts, _ := rs.accountRepo.FindAll(ctx)
	
	for _, account := range accounts {
		if rs.isDividendAccount(account) {
			// Debits to dividend accounts represent payments
			debits, _ := rs.calculateAccountTotalsForPeriod(account.ID, startDate, endDate)
			total += debits
		}
	}
	
	return total
}

// Account classification helper methods

func (rs *ReportService) isAccountsReceivableAccount(account models.Account) bool {
	return strings.Contains(strings.ToLower(account.Name), "receivable") ||
		strings.Contains(strings.ToLower(account.Code), "receivable") ||
		strings.HasPrefix(account.Code, "1-2") // Typical A/R account code
}

func (rs *ReportService) isAccountsPayableAccount(account models.Account) bool {
	return strings.Contains(strings.ToLower(account.Name), "payable") ||
		strings.Contains(strings.ToLower(account.Code), "payable") ||
		strings.HasPrefix(account.Code, "2-1") // Typical A/P account code
}

func (rs *ReportService) isSalaryExpenseAccount(account models.Account) bool {
	if account.Type != models.AccountTypeExpense {
		return false
	}
	salaryKeywords := []string{"salary", "wage", "payroll", "compensation", "benefit"}
	accountNameLower := strings.ToLower(account.Name)
	
	for _, keyword := range salaryKeywords {
		if strings.Contains(accountNameLower, keyword) {
			return true
		}
	}
	return false
}

func (rs *ReportService) isOperatingExpenseAccount(account models.Account) bool {
	if account.Type != models.AccountTypeExpense {
		return false
	}
	// Exclude non-operating expenses like interest, depreciation
	nonOperatingKeywords := []string{"interest", "depreciation", "amortization", "loss", "tax"}
	accountNameLower := strings.ToLower(account.Name)
	
	for _, keyword := range nonOperatingKeywords {
		if strings.Contains(accountNameLower, keyword) {
			return false
		}
	}
	return true
}

func (rs *ReportService) isFixedAssetAccount(account models.Account) bool {
	if account.Type != models.AccountTypeAsset {
		return false
	}
	fixedAssetKeywords := []string{"equipment", "building", "furniture", "vehicle", "machinery", "computer"}
	accountNameLower := strings.ToLower(account.Name)
	
	for _, keyword := range fixedAssetKeywords {
		if strings.Contains(accountNameLower, keyword) {
			return true
		}
	}
	// Also check for typical fixed asset account codes
	return strings.HasPrefix(account.Code, "1-5") || strings.HasPrefix(account.Code, "1-6")
}

func (rs *ReportService) isLoanLiabilityAccount(account models.Account) bool {
	if account.Type != models.AccountTypeLiability {
		return false
	}
	loanKeywords := []string{"loan", "note", "mortgage", "debt", "borrowing"}
	accountNameLower := strings.ToLower(account.Name)
	
	for _, keyword := range loanKeywords {
		if strings.Contains(accountNameLower, keyword) {
			return true
		}
	}
	return false
}

func (rs *ReportService) isOwnerEquityAccount(account models.Account) bool {
	if account.Type != models.AccountTypeEquity {
		return false
	}
	equityKeywords := []string{"capital", "investment", "contribution", "equity"}
	accountNameLower := strings.ToLower(account.Name)
	
	for _, keyword := range equityKeywords {
		if strings.Contains(accountNameLower, keyword) {
			return true
		}
	}
	return false
}

func (rs *ReportService) isDividendAccount(account models.Account) bool {
	// Dividends can be equity (contra) or expense accounts
	dividendKeywords := []string{"dividend", "withdrawal", "draw"}
	accountNameLower := strings.ToLower(account.Name)
	
	for _, keyword := range dividendKeywords {
		if strings.Contains(accountNameLower, keyword) {
			return true
		}
	}
	return false
}

// calculateAccountTotalsForPeriod calculates debit and credit totals for a specific period
func (rs *ReportService) calculateAccountTotalsForPeriod(accountID uint, startDate, endDate time.Time) (debitTotal, creditTotal float64) {
	rs.db.Table("journal_entries").
		Joins("JOIN journals ON journal_entries.journal_id = journals.id").
		Where("journal_entries.account_id = ? AND journals.date BETWEEN ? AND ? AND journals.status = 'POSTED'", accountID, startDate, endDate).
		Select("COALESCE(SUM(journal_entries.debit_amount), 0) as total_debits, COALESCE(SUM(journal_entries.credit_amount), 0) as total_credits").
		Row().Scan(&debitTotal, &creditTotal)
	
	return debitTotal, creditTotal
}

// PDF Generation methods (simplified implementations)

func (rs *ReportService) generateBalanceSheetPDF(data *BalanceSheetData, asOfDate time.Time) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "BALANCE SHEET")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(190, 8, fmt.Sprintf("As of %s", asOfDate.Format("January 2, 2006")))
	pdf.Ln(15)
	
	// Assets section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "ASSETS")
	pdf.Ln(8)
	
	pdf.SetFont("Arial", "", 10)
	for _, asset := range data.Assets {
		if !asset.IsHeader {
			pdf.Cell(120, 6, fmt.Sprintf("  %s", asset.AccountName))
			pdf.Cell(70, 6, fmt.Sprintf("%.2f", asset.Balance))
		}
	}
	pdf.Ln(5)
	
	// Total Assets
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(120, 6, "TOTAL ASSETS")
	pdf.Cell(70, 6, fmt.Sprintf("%.2f", data.TotalAssets))
	
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}

func (rs *ReportService) generateProfitLossPDF(data *ProfitLossData, startDate, endDate time.Time) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "PROFIT & LOSS STATEMENT")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(190, 8, fmt.Sprintf("From %s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))
	pdf.Ln(15)
	
	// Revenue section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "REVENUE")
	pdf.Ln(8)
	
	pdf.SetFont("Arial", "", 10)
	for _, revenue := range data.Revenue {
		pdf.Cell(120, 6, fmt.Sprintf("  %s", revenue.AccountName))
		pdf.Cell(70, 6, fmt.Sprintf("%.2f", revenue.Balance))
	}
	pdf.Ln(5)
	
	// Total Revenue
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(120, 6, "TOTAL REVENUE")
	pdf.Cell(70, 6, fmt.Sprintf("%.2f", data.TotalRevenue))
	pdf.Ln(5)
	
	// Net Income
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(120, 8, "NET INCOME")
	pdf.Cell(70, 8, fmt.Sprintf("%.2f", data.NetIncome))
	
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}

func (rs *ReportService) generateCashFlowPDF(data *CashFlowData, startDate, endDate time.Time) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "CASH FLOW STATEMENT")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(190, 8, fmt.Sprintf("From %s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))
	pdf.Ln(15)
	
	// Operating Activities
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 8, "OPERATING ACTIVITIES")
	pdf.Ln(8)
	
	pdf.SetFont("Arial", "", 10)
	for _, item := range data.OperatingActivities {
		pdf.Cell(120, 6, fmt.Sprintf("  %s", item.Description))
		pdf.Cell(70, 6, fmt.Sprintf("%.2f", item.Amount))
	}
	
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}

func (rs *ReportService) generateTrialBalancePDF(data []ReportAccountBalance, asOfDate time.Time) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "TRIAL BALANCE")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(190, 8, fmt.Sprintf("As of %s", asOfDate.Format("January 2, 2006")))
	pdf.Ln(15)
	
	// Headers
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(60, 8, "Account Name")
	pdf.Cell(40, 8, "Debit")
	pdf.Cell(40, 8, "Credit")
	
	// Data
	pdf.SetFont("Arial", "", 9)
	for _, account := range data {
		pdf.Cell(60, 6, account.AccountName)
		pdf.Cell(40, 6, fmt.Sprintf("%.2f", account.DebitTotal))
		pdf.Cell(40, 6, fmt.Sprintf("%.2f", account.CreditTotal))
	}
	
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}

// Excel Generation methods

func (rs *ReportService) generateBalanceSheetExcel(data *BalanceSheetData, asOfDate time.Time) ([]byte, error) {
	f := excelize.NewFile()
	sheetName := "Balance Sheet"
	
	// Set active sheet name
	f.SetSheetName("Sheet1", sheetName)
	
	// Company header
	f.SetCellValue(sheetName, "A1", "BALANCE SHEET")
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("As of %s", asOfDate.Format("January 2, 2006")))
	
	// Style for title
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(sheetName, "A1", "B1", titleStyle)
	
	// Headers style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"E2E8F0"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	
	// Number format
	numberStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 4, // Number format with comma separator
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	
	row := 4
	
	// Assets section
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "ASSETS")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++
	
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Account Name")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Balance")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++
	
	for _, asset := range data.Assets {
		if !asset.IsHeader {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), asset.AccountName)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), asset.Balance)
			f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), numberStyle)
			row++
		}
	}
	
	// Total Assets
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "TOTAL ASSETS")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), data.TotalAssets)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row += 2
	
	// Liabilities section
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "LIABILITIES")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++
	
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Account Name")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Balance")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++
	
	liabilitiesTotal := 0.0
	for _, liability := range data.Liabilities {
		if !liability.IsHeader {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), liability.AccountName)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), liability.Balance)
			f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), numberStyle)
			liabilitiesTotal += liability.Balance
			row++
		}
	}
	
	// Equity section
	row++
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "EQUITY")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++
	
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Account Name")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Balance")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++
	
	for _, equity := range data.Equity {
		if !equity.IsHeader {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), equity.AccountName)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), equity.Balance)
			f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), numberStyle)
			row++
		}
	}
	
	// Total Liabilities & Equity
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "TOTAL LIABILITIES & EQUITY")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), data.TotalLiabilitiesEquity)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	
	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 40)
	f.SetColWidth(sheetName, "B", "B", 20)
	
	// Save to buffer
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

func (rs *ReportService) generateProfitLossExcel(data *ProfitLossData, startDate, endDate time.Time) ([]byte, error) {
	f := excelize.NewFile()
	sheetName := "Profit & Loss"
	
	// Set active sheet name
	f.SetSheetName("Sheet1", sheetName)
	
	// Company header
	f.SetCellValue(sheetName, "A1", "PROFIT & LOSS STATEMENT")
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("From %s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))
	
	// Style for title
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(sheetName, "A1", "B1", titleStyle)
	
	// Headers style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"E2E8F0"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	
	// Number format
	numberStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 4, // Number format with comma separator
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	
	row := 4
	
	// Revenue section
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "REVENUE")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++
	
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Account Name")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Amount")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++
	
	for _, revenue := range data.Revenue {
		if !revenue.IsHeader {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), revenue.AccountName)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), revenue.Balance)
			f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), numberStyle)
			row++
		}
	}
	
	// Total Revenue
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "TOTAL REVENUE")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), data.TotalRevenue)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row += 2
	
	// Expenses section
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "EXPENSES")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++
	
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Account Name")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), "Amount")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++
	
	for _, expense := range data.Expenses {
		if !expense.IsHeader {
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), expense.AccountName)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), expense.Balance)
			f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), numberStyle)
			row++
		}
	}
	
	// Total Expenses
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "TOTAL EXPENSES")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), data.TotalExpenses)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row += 2
	
	// Net Income
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "NET INCOME")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), data.NetIncome)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	
	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 40)
	f.SetColWidth(sheetName, "B", "B", 20)
	
	// Save to buffer
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

func (rs *ReportService) generateCashFlowExcel(data *CashFlowData, startDate, endDate time.Time) ([]byte, error) {
	f := excelize.NewFile()
	sheetName := "Cash Flow"
	
	// Set active sheet name
	f.SetSheetName("Sheet1", sheetName)
	
	// Company header
	f.SetCellValue(sheetName, "A1", "CASH FLOW STATEMENT")
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("From %s to %s", startDate.Format("January 2, 2006"), endDate.Format("January 2, 2006")))
	
	// Style for title
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellStyle(sheetName, "A1", "B1", titleStyle)
	
	// Headers style
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"E2E8F0"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	
	// Number format
	numberStyle, _ := f.NewStyle(&excelize.Style{
		NumFmt: 4, // Number format with comma separator
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	
	row := 4
	
	// Beginning Balance
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Beginning Cash Balance")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), data.BeginningBalance)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row += 2
	
	// Operating Activities section
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "OPERATING ACTIVITIES")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++
	
	for _, activity := range data.OperatingActivities {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), activity.Description)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), activity.Amount)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), numberStyle)
		row++
	}
	row++
	
	// Investing Activities section
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "INVESTING ACTIVITIES")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++
	
	for _, activity := range data.InvestingActivities {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), activity.Description)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), activity.Amount)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), numberStyle)
		row++
	}
	row++
	
	// Financing Activities section
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "FINANCING ACTIVITIES")
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++
	
	for _, activity := range data.FinancingActivities {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), activity.Description)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), activity.Amount)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), numberStyle)
		row++
	}
	row += 2
	
	// Net Cash Flow
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "NET CASH FLOW")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), data.NetCashFlow)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	row++
	
	// Ending Balance
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "Ending Cash Balance")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), data.EndingBalance)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("B%d", row), headerStyle)
	
	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 40)
	f.SetColWidth(sheetName, "B", "B", 20)
	
	// Save to buffer
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

func (rs *ReportService) generateReceivablesPDF(data []ReceivableData, asOfDate time.Time) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "") // Landscape for more columns
	pdf.AddPage()
	
	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(270, 10, "ACCOUNTS RECEIVABLE")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(270, 8, fmt.Sprintf("As of %s", asOfDate.Format("January 2, 2006")))
	pdf.Ln(15)
	
	// Headers
	pdf.SetFont("Arial", "B", 9)
	pdf.Cell(50, 8, "Customer")
	pdf.Cell(30, 8, "Invoice")
	pdf.Cell(25, 8, "Date")
	pdf.Cell(25, 8, "Due Date")
	pdf.Cell(30, 8, "Amount")
	pdf.Cell(30, 8, "Paid")
	pdf.Cell(30, 8, "Outstanding")
	pdf.Cell(25, 8, "Days Due")
	
	// Data
	pdf.SetFont("Arial", "", 8)
	for _, receivable := range data {
		pdf.Cell(50, 6, receivable.CustomerName)
		pdf.Cell(30, 6, receivable.InvoiceNumber)
		pdf.Cell(25, 6, receivable.Date.Format("2006-01-02"))
		pdf.Cell(25, 6, receivable.DueDate.Format("2006-01-02"))
		pdf.Cell(30, 6, fmt.Sprintf("%.2f", receivable.Amount))
		pdf.Cell(30, 6, fmt.Sprintf("%.2f", receivable.PaidAmount))
		pdf.Cell(30, 6, fmt.Sprintf("%.2f", receivable.Outstanding))
		pdf.Cell(25, 6, strconv.Itoa(receivable.DaysOverdue))
	}
	
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}

func (rs *ReportService) generatePayablesPDF(data []PayableData, asOfDate time.Time) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "") // Landscape for more columns
	pdf.AddPage()
	
	// Title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(270, 10, "ACCOUNTS PAYABLE")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(270, 8, fmt.Sprintf("As of %s", asOfDate.Format("January 2, 2006")))
	pdf.Ln(15)
	
	// Headers
	pdf.SetFont("Arial", "B", 9)
	pdf.Cell(50, 8, "Vendor")
	pdf.Cell(30, 8, "Purchase")
	pdf.Cell(25, 8, "Date")
	pdf.Cell(25, 8, "Due Date")
	pdf.Cell(30, 8, "Amount")
	pdf.Cell(30, 8, "Paid")
	pdf.Cell(30, 8, "Outstanding")
	pdf.Cell(25, 8, "Days Due")
	
	// Data
	pdf.SetFont("Arial", "", 8)
	for _, payable := range data {
		pdf.Cell(50, 6, payable.VendorName)
		pdf.Cell(30, 6, fmt.Sprintf("PO-%d", payable.PurchaseID))
		pdf.Cell(25, 6, payable.Date.Format("2006-01-02"))
		pdf.Cell(25, 6, payable.DueDate.Format("2006-01-02"))
		pdf.Cell(30, 6, fmt.Sprintf("%.2f", payable.Amount))
		pdf.Cell(30, 6, fmt.Sprintf("%.2f", payable.PaidAmount))
		pdf.Cell(30, 6, fmt.Sprintf("%.2f", payable.Outstanding))
		pdf.Cell(25, 6, strconv.Itoa(payable.DaysOverdue))
	}
	
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}
