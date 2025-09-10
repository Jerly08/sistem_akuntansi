package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"

	"gorm.io/gorm"
)

// EnhancedReportService provides comprehensive financial and operational reporting
// with proper accounting logic and business analytics
type EnhancedReportService struct {
	db              *gorm.DB
	accountRepo     repositories.AccountRepository
	salesRepo       *repositories.SalesRepository
	purchaseRepo    *repositories.PurchaseRepository
	productRepo     *repositories.ProductRepository
	contactRepo     repositories.ContactRepository
	paymentRepo     *repositories.PaymentRepository
	cashBankRepo    *repositories.CashBankRepository
	companyProfile  *models.CompanyProfile
}

// BalanceSheetData represents a comprehensive balance sheet structure
type BalanceSheetData struct {
	Company      CompanyInfo            `json:"company"`
	AsOfDate     time.Time             `json:"as_of_date"`
	Currency     string                `json:"currency"`
	Assets       BalanceSheetSection   `json:"assets"`
	Liabilities  BalanceSheetSection   `json:"liabilities"`
	Equity       BalanceSheetSection   `json:"equity"`
	TotalAssets  float64               `json:"total_assets"`
	TotalEquity  float64               `json:"total_equity"`
	IsBalanced   bool                  `json:"is_balanced"`
	Difference   float64               `json:"difference"`
	GeneratedAt  time.Time             `json:"generated_at"`
}

// ProfitLossData represents a comprehensive P&L statement structure
type ProfitLossData struct {
	Company            CompanyInfo         `json:"company"`
	StartDate          time.Time           `json:"start_date"`
	EndDate            time.Time           `json:"end_date"`
	Currency           string              `json:"currency"`
	Revenue            PLSection           `json:"revenue"`
	CostOfGoodsSold    PLSection           `json:"cost_of_goods_sold"`
	GrossProfit        float64             `json:"gross_profit"`
	GrossProfitMargin  float64             `json:"gross_profit_margin"`
	OperatingExpenses  PLSection           `json:"operating_expenses"`
	OperatingIncome    float64             `json:"operating_income"`
	OtherIncome        PLSection           `json:"other_income"`
	OtherExpenses      PLSection           `json:"other_expenses"`
	EBITDA             float64             `json:"ebitda"`
	EBIT               float64             `json:"ebit"`
	NetIncomeBeforeTax float64             `json:"net_income_before_tax"`
	TaxExpense         float64             `json:"tax_expense"`
	NetIncome          float64             `json:"net_income"`
	NetIncomeMargin    float64             `json:"net_income_margin"`
	EarningsPerShare   float64             `json:"earnings_per_share"`
	DilutedEPS         float64             `json:"diluted_eps"`
	SharesOutstanding  float64             `json:"shares_outstanding"`
	GeneratedAt        time.Time           `json:"generated_at"`
}

// ProfitLossComparative represents comparative P&L analysis
type ProfitLossComparative struct {
	CurrentPeriod    ProfitLossData    `json:"current_period"`
	PriorPeriod      ProfitLossData    `json:"prior_period"`
	Variances        PLVarianceData    `json:"variances"`
	TrendAnalysis    PLTrendAnalysis   `json:"trend_analysis"`
	GeneratedAt      time.Time         `json:"generated_at"`
}

// PLVarianceData represents variance analysis between periods
type PLVarianceData struct {
	Revenue           PLVariance `json:"revenue"`
	CostOfGoodsSold   PLVariance `json:"cost_of_goods_sold"`
	GrossProfit       PLVariance `json:"gross_profit"`
	OperatingExpenses PLVariance `json:"operating_expenses"`
	OperatingIncome   PLVariance `json:"operating_income"`
	OtherIncome       PLVariance `json:"other_income"`
	OtherExpenses     PLVariance `json:"other_expenses"`
	NetIncome         PLVariance `json:"net_income"`
	EBITDA            PLVariance `json:"ebitda"`
	EBIT              PLVariance `json:"ebit"`
}

// PLVariance represents individual line item variance
type PLVariance struct {
	Current        float64 `json:"current"`
	Prior          float64 `json:"prior"`
	AbsoluteChange float64 `json:"absolute_change"`
	PercentChange  float64 `json:"percent_change"`
	Trend          string  `json:"trend"` // INCREASING, DECREASING, STABLE
}

// PLTrendAnalysis represents trend analysis over multiple periods
type PLTrendAnalysis struct {
	RevenueGrowthRate     float64 `json:"revenue_growth_rate"`
	ProfitabilityTrend    string  `json:"profitability_trend"`
	CostManagementIndex   float64 `json:"cost_management_index"`
	OperationalEfficiency float64 `json:"operational_efficiency"`
	MarginStability       string  `json:"margin_stability"`
}

// CashFlowData represents a comprehensive cash flow statement structure
type CashFlowData struct {
	Company              CompanyInfo     `json:"company"`
	StartDate            time.Time       `json:"start_date"`
	EndDate              time.Time       `json:"end_date"`
	Currency             string          `json:"currency"`
	OperatingActivities  CashFlowSection `json:"operating_activities"`
	InvestingActivities  CashFlowSection `json:"investing_activities"`
	FinancingActivities  CashFlowSection `json:"financing_activities"`
	NetCashFlow          float64         `json:"net_cash_flow"`
	BeginningCash        float64         `json:"beginning_cash"`
	EndingCash           float64         `json:"ending_cash"`
	GeneratedAt          time.Time       `json:"generated_at"`
}

// SalesSummaryData represents comprehensive sales analytics
type SalesSummaryData struct {
	Company                CompanyInfo            `json:"company"`
	StartDate              time.Time              `json:"start_date"`
	EndDate                time.Time              `json:"end_date"`
	Currency               string                 `json:"currency"`
	TotalRevenue           float64                `json:"total_revenue"`
	TotalTransactions      int64                  `json:"total_transactions"`
	AverageOrderValue      float64                `json:"average_order_value"`
	TotalCustomers         int64                  `json:"total_customers"`
	NewCustomers           int64                  `json:"new_customers"`
	ReturningCustomers     int64                  `json:"returning_customers"`
	SalesByPeriod          []PeriodData           `json:"sales_by_period"`
	SalesByCustomer        []CustomerSalesData    `json:"sales_by_customer"`
	SalesByProduct         []ProductSalesData     `json:"sales_by_product"`
	SalesByStatus          []StatusData           `json:"sales_by_status"`
	TopPerformers          TopPerformersData      `json:"top_performers"`
	GrowthAnalysis         GrowthAnalysisData     `json:"growth_analysis"`
	GeneratedAt            time.Time              `json:"generated_at"`
}

// PurchaseSummaryData represents comprehensive purchase analytics
type PurchaseSummaryData struct {
	Company                CompanyInfo            `json:"company"`
	StartDate              time.Time              `json:"start_date"`
	EndDate                time.Time              `json:"end_date"`
	Currency               string                 `json:"currency"`
	TotalPurchases         float64                `json:"total_purchases"`
	TotalTransactions      int64                  `json:"total_transactions"`
	AveragePurchaseValue   float64                `json:"average_purchase_value"`
	TotalVendors           int64                  `json:"total_vendors"`
	NewVendors             int64                  `json:"new_vendors"`
	PurchasesByPeriod      []PeriodData           `json:"purchases_by_period"`
	PurchasesByVendor      []VendorPurchaseData   `json:"purchases_by_vendor"`
	PurchasesByCategory    []CategoryPurchaseData `json:"purchases_by_category"`
	PurchasesByStatus      []StatusData           `json:"purchases_by_status"`
	TopVendors             TopVendorsData         `json:"top_vendors"`
	CostAnalysis           CostAnalysisData       `json:"cost_analysis"`
	GeneratedAt            time.Time              `json:"generated_at"`
}

// Supporting data structures
type CompanyInfo struct {
	Name        string `json:"name"`
	Address     string `json:"address"`
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Website     string `json:"website"`
	TaxNumber   string `json:"tax_number"`
}

type BalanceSheetSection struct {
	Name       string                    `json:"name"`
	Items      []BalanceSheetItem        `json:"items"`
	Subtotals  []BalanceSheetSubtotal    `json:"subtotals"`
	Total      float64                   `json:"total"`
}

type BalanceSheetItem struct {
	AccountID   uint    `json:"account_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Balance     float64 `json:"balance"`
	Category    string  `json:"category"`
	Level       int     `json:"level"`
	IsHeader    bool    `json:"is_header"`
}

type BalanceSheetSubtotal struct {
	Name     string  `json:"name"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

type PLSection struct {
	Name      string     `json:"name"`
	Items     []PLItem   `json:"items"`
	Subtotal  float64    `json:"subtotal"`
}

type PLItem struct {
	AccountID   uint    `json:"account_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Percentage  float64 `json:"percentage"`
}

type CashFlowSection struct {
	Name  string          `json:"name"`
	Items []CashFlowItem  `json:"items"`
	Total float64         `json:"total"`
}

type CashFlowItem struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
}

type PeriodData struct {
	Period        string    `json:"period"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Amount        float64   `json:"amount"`
	Transactions  int64     `json:"transactions"`
	GrowthRate    float64   `json:"growth_rate"`
}

type CustomerSalesData struct {
	CustomerID       uint    `json:"customer_id"`
	CustomerName     string  `json:"customer_name"`
	TotalAmount      float64 `json:"total_amount"`
	TransactionCount int64   `json:"transaction_count"`
	AverageOrder     float64 `json:"average_order"`
	LastOrderDate    time.Time `json:"last_order_date"`
	FirstOrderDate   time.Time `json:"first_order_date"`
}

type ProductSalesData struct {
	ProductID        uint    `json:"product_id"`
	ProductName      string  `json:"product_name"`
	QuantitySold     int64   `json:"quantity_sold"`
	TotalAmount      float64 `json:"total_amount"`
	AveragePrice     float64 `json:"average_price"`
	TransactionCount int64   `json:"transaction_count"`
}

type VendorPurchaseData struct {
	VendorID         uint    `json:"vendor_id"`
	VendorName       string  `json:"vendor_name"`
	TotalAmount      float64 `json:"total_amount"`
	TransactionCount int64   `json:"transaction_count"`
	AverageOrder     float64 `json:"average_order"`
	LastOrderDate    time.Time `json:"last_order_date"`
	FirstOrderDate   time.Time `json:"first_order_date"`
}

type CategoryPurchaseData struct {
	CategoryID       uint    `json:"category_id"`
	CategoryName     string  `json:"category_name"`
	TotalAmount      float64 `json:"total_amount"`
	TransactionCount int64   `json:"transaction_count"`
	Percentage       float64 `json:"percentage"`
}

type StatusData struct {
	Status      string `json:"status"`
	Count       int64  `json:"count"`
	Amount      float64 `json:"amount"`
	Percentage  float64 `json:"percentage"`
}

type TopPerformersData struct {
	TopCustomers []CustomerSalesData `json:"top_customers"`
	TopProducts  []ProductSalesData  `json:"top_products"`
	TopSalespeople []SalespersonData  `json:"top_salespeople"`
}

type TopVendorsData struct {
	TopVendors     []VendorPurchaseData   `json:"top_vendors"`
	TopCategories  []CategoryPurchaseData `json:"top_categories"`
	TopProducts    []ProductPurchaseData  `json:"top_products"`
}

type SalespersonData struct {
	SalespersonID    uint    `json:"salesperson_id"`
	SalespersonName  string  `json:"salesperson_name"`
	TotalSales       float64 `json:"total_sales"`
	TransactionCount int64   `json:"transaction_count"`
	Commission       float64 `json:"commission"`
}

type ProductPurchaseData struct {
	ProductID        uint    `json:"product_id"`
	ProductName      string  `json:"product_name"`
	QuantityPurchased int64  `json:"quantity_purchased"`
	TotalAmount      float64 `json:"total_amount"`
	AveragePrice     float64 `json:"average_price"`
}

type GrowthAnalysisData struct {
	MonthOverMonth   float64 `json:"month_over_month"`
	QuarterOverQuarter float64 `json:"quarter_over_quarter"`
	YearOverYear     float64 `json:"year_over_year"`
	TrendDirection   string  `json:"trend_direction"`
	SeasonalityIndex float64 `json:"seasonality_index"`
}

type CostAnalysisData struct {
	TotalCostOfGoods float64 `json:"total_cost_of_goods"`
	AverageCostPerUnit float64 `json:"average_cost_per_unit"`
	CostVariance     float64 `json:"cost_variance"`
	InflationImpact  float64 `json:"inflation_impact"`
}

// NewEnhancedReportService creates a new enhanced report service
func NewEnhancedReportService(
	db *gorm.DB,
	accountRepo repositories.AccountRepository,
	salesRepo *repositories.SalesRepository,
	purchaseRepo *repositories.PurchaseRepository,
	productRepo *repositories.ProductRepository,
	contactRepo repositories.ContactRepository,
	paymentRepo *repositories.PaymentRepository,
	cashBankRepo *repositories.CashBankRepository,
) *EnhancedReportService {
	service := &EnhancedReportService{
		db:           db,
		accountRepo:  accountRepo,
		salesRepo:    salesRepo,
		purchaseRepo: purchaseRepo,
		productRepo:  productRepo,
		contactRepo:  contactRepo,
		paymentRepo:  paymentRepo,
		cashBankRepo: cashBankRepo,
	}
	
	// Load company profile
	service.loadCompanyProfile()
	
	return service
}

// loadCompanyProfile loads the company profile for report headers
func (ers *EnhancedReportService) loadCompanyProfile() {
	var profile models.CompanyProfile
	if err := ers.db.First(&profile).Error; err != nil {
		// Check for existing company data from user account or other sources
		// This is a placeholder - in production, you might want to:
		// 1. Load from environment variables
		// 2. Load from initial setup wizard data
		// 3. Load from user registration information
		// 4. Provide a setup interface for users to configure
		
		// Create default profile using environment variables or fallbacks
		profile = models.CompanyProfile{
			Name:            ers.getDefaultCompanyName(),
			Address:         ers.getDefaultCompanyAddress(),
			City:            ers.getDefaultCompanyCity(),
			State:           ers.getDefaultState(),
			Country:         ers.getDefaultCountry(),
			PostalCode:      ers.getDefaultPostalCode(),
			Phone:           ers.getDefaultCompanyPhone(),
			Email:           ers.getDefaultCompanyEmail(),
			Website:         ers.getDefaultCompanyWebsite(),
			Currency:        ers.getDefaultCurrency(),
			FiscalYearStart: "01-01", // January 1st (standard in Indonesia)
			TaxNumber:       ers.getDefaultTaxNumber(),
		}
		
		// Save the default profile to database
		ers.db.Create(&profile)
	}
	ers.companyProfile = &profile
}

// getCompanyInfo returns company information structure
func (ers *EnhancedReportService) getCompanyInfo() CompanyInfo {
	return CompanyInfo{
		Name:       ers.companyProfile.Name,
		Address:    ers.companyProfile.Address,
		City:       ers.companyProfile.City,
		State:      ers.companyProfile.State,
		PostalCode: ers.companyProfile.PostalCode,
		Phone:      ers.companyProfile.Phone,
		Email:      ers.companyProfile.Email,
		Website:    ers.companyProfile.Website,
		TaxNumber:  ers.companyProfile.TaxNumber,
	}
}

// GenerateBalanceSheet creates a comprehensive balance sheet with proper accounting logic
func (ers *EnhancedReportService) GenerateBalanceSheet(asOfDate time.Time) (*BalanceSheetData, error) {
	// Get all accounts with their balances
	ctx := context.Background()
	accounts, err := ers.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}

	// Initialize balance sheet structure
	balanceSheet := &BalanceSheetData{
		Company:     ers.getCompanyInfo(),
		AsOfDate:    asOfDate,
		Currency:    ers.companyProfile.Currency,
		GeneratedAt: time.Now(),
	}

	// Process each account and calculate balances
	var assets, liabilities, equity []BalanceSheetItem

	for _, account := range accounts {
		if !account.IsActive {
			continue
		}

		balance := ers.calculateAccountBalance(account.ID, asOfDate)
		
		// Skip accounts with zero balance (optional - can be configurable)
		if balance == 0 {
			continue
		}

		item := BalanceSheetItem{
			AccountID: account.ID,
			Code:      account.Code,
			Name:      account.Name,
			Balance:   balance,
			Category:  account.Category,
			Level:     account.Level,
			IsHeader:  account.IsHeader,
		}

		switch account.Type {
		case models.AccountTypeAsset:
			assets = append(assets, item)
			balanceSheet.TotalAssets += balance
		case models.AccountTypeLiability:
			liabilities = append(liabilities, item)
		case models.AccountTypeEquity:
			equity = append(equity, item)
			balanceSheet.TotalEquity += balance
		}
	}

	// Sort items by account code
	sort.Slice(assets, func(i, j int) bool { return assets[i].Code < assets[j].Code })
	sort.Slice(liabilities, func(i, j int) bool { return liabilities[i].Code < liabilities[j].Code })
	sort.Slice(equity, func(i, j int) bool { return equity[i].Code < equity[j].Code })

	// Build assets section with subtotals
	balanceSheet.Assets = ers.buildAssetsSection(assets)
	
	// Build liabilities section with subtotals
	balanceSheet.Liabilities = ers.buildLiabilitiesSection(liabilities)
	
	// Build equity section
	balanceSheet.Equity = ers.buildEquitySection(equity)

	// Calculate total liabilities and equity
	totalLiabilitiesAndEquity := balanceSheet.Liabilities.Total + balanceSheet.Equity.Total
	balanceSheet.TotalEquity = totalLiabilitiesAndEquity

	// Check if balance sheet is balanced
	balanceSheet.Difference = balanceSheet.TotalAssets - totalLiabilitiesAndEquity
	balanceSheet.IsBalanced = math.Abs(balanceSheet.Difference) < 0.01 // Allow for small rounding differences

	return balanceSheet, nil
}

// GenerateProfitLoss creates a comprehensive P&L statement with proper accounting logic
func (ers *EnhancedReportService) GenerateProfitLoss(startDate, endDate time.Time) (*ProfitLossData, error) {
	// Get all accounts with their period balances
	ctx := context.Background()
	accounts, err := ers.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}

	// Initialize P&L structure
	profitLoss := &ProfitLossData{
		Company:     ers.getCompanyInfo(),
		StartDate:   startDate,
		EndDate:     endDate,
		Currency:    ers.companyProfile.Currency,
		GeneratedAt: time.Now(),
	}

	// Initialize categorized items
	var operatingRevenueItems, nonOperatingRevenueItems []PLItem
	var cogsItems, operatingExpenseItems, nonOperatingExpenseItems, taxExpenseItems []PLItem
	var totalOperatingRevenue, totalNonOperatingRevenue, totalCOGS float64
	var totalOperatingExpenses, totalNonOperatingExpenses, totalTaxExpenses float64

	for _, account := range accounts {
		if !account.IsActive {
			continue
		}

		// Calculate activity for the period
		balance := ers.calculateAccountBalanceForPeriod(account.ID, startDate, endDate)
		
		// Skip accounts with no activity
		if balance == 0 {
			continue
		}

		item := PLItem{
			AccountID: account.ID,
			Code:      account.Code,
			Name:      account.Name,
			Amount:    balance,
			Category:  account.Category,
		}

		switch account.Type {
		case models.AccountTypeRevenue:
			if ers.isOperatingRevenue(account.Category) {
				operatingRevenueItems = append(operatingRevenueItems, item)
				totalOperatingRevenue += balance
			} else {
				nonOperatingRevenueItems = append(nonOperatingRevenueItems, item)
				totalNonOperatingRevenue += balance
			}
		case models.AccountTypeExpense:
			if ers.isCOGS(account.Category) {
				cogsItems = append(cogsItems, item)
				totalCOGS += balance
			} else if ers.isTaxExpense(account.Category) {
				taxExpenseItems = append(taxExpenseItems, item)
				totalTaxExpenses += balance
			} else if ers.isOperatingExpense(account.Category) {
				operatingExpenseItems = append(operatingExpenseItems, item)
				totalOperatingExpenses += balance
			} else {
				nonOperatingExpenseItems = append(nonOperatingExpenseItems, item)
				totalNonOperatingExpenses += balance
			}
		}
	}

	// Calculate total revenue
	totalRevenue := totalOperatingRevenue + totalNonOperatingRevenue

	// Calculate percentages for all items
	ers.calculateItemPercentages(operatingRevenueItems, totalRevenue)
	ers.calculateItemPercentages(nonOperatingRevenueItems, totalRevenue)
	ers.calculateItemPercentages(cogsItems, totalRevenue)
	ers.calculateItemPercentages(operatingExpenseItems, totalRevenue)
	ers.calculateItemPercentages(nonOperatingExpenseItems, totalRevenue)
	ers.calculateItemPercentages(taxExpenseItems, totalRevenue)

	// Build comprehensive P&L sections
	profitLoss.Revenue = PLSection{
		Name:     "Revenue",
		Items:    append(operatingRevenueItems, nonOperatingRevenueItems...),
		Subtotal: totalRevenue,
	}

	profitLoss.CostOfGoodsSold = PLSection{
		Name:     "Cost of Goods Sold",
		Items:    cogsItems,
		Subtotal: totalCOGS,
	}

	profitLoss.OperatingExpenses = PLSection{
		Name:     "Operating Expenses",
		Items:    operatingExpenseItems,
		Subtotal: totalOperatingExpenses,
	}

	profitLoss.OtherIncome = PLSection{
		Name:     "Other Income",
		Items:    nonOperatingRevenueItems,
		Subtotal: totalNonOperatingRevenue,
	}

	profitLoss.OtherExpenses = PLSection{
		Name:     "Other Expenses",
		Items:    nonOperatingExpenseItems,
		Subtotal: totalNonOperatingExpenses,
	}

	// Calculate comprehensive financial metrics
	profitLoss.GrossProfit = totalOperatingRevenue - totalCOGS
	if totalOperatingRevenue != 0 {
		profitLoss.GrossProfitMargin = (profitLoss.GrossProfit / totalOperatingRevenue) * 100
	}

	// Calculate EBITDA (approximation without depreciation detail)
	profitLoss.EBITDA = profitLoss.GrossProfit - totalOperatingExpenses + totalNonOperatingRevenue - totalNonOperatingExpenses
	
	// Calculate EBIT
	profitLoss.EBIT = profitLoss.EBITDA // Simplified - would subtract depreciation and amortization

	// Calculate Operating Income
	profitLoss.OperatingIncome = profitLoss.GrossProfit - totalOperatingExpenses

	// Calculate Net Income Before Tax
	profitLoss.NetIncomeBeforeTax = profitLoss.OperatingIncome + totalNonOperatingRevenue - totalNonOperatingExpenses

	// Tax Expenses
	profitLoss.TaxExpense = totalTaxExpenses

	// Calculate Net Income
	profitLoss.NetIncome = profitLoss.NetIncomeBeforeTax - totalTaxExpenses

	if totalRevenue != 0 {
		profitLoss.NetIncomeMargin = (profitLoss.NetIncome / totalRevenue) * 100
	}

	// Calculate EPS with proper shares outstanding
	profitLoss.SharesOutstanding = ers.getSharesOutstanding()
	if profitLoss.SharesOutstanding > 0 {
		profitLoss.EarningsPerShare = profitLoss.NetIncome / profitLoss.SharesOutstanding
		profitLoss.DilutedEPS = profitLoss.EarningsPerShare // Simplified - would account for dilutive securities
	} else {
		profitLoss.EarningsPerShare = profitLoss.NetIncome
		profitLoss.DilutedEPS = profitLoss.NetIncome
	}

	return profitLoss, nil
}

// GenerateCashFlow creates a comprehensive cash flow statement
func (ers *EnhancedReportService) GenerateCashFlow(startDate, endDate time.Time) (*CashFlowData, error) {
	// Initialize cash flow structure
	cashFlow := &CashFlowData{
		Company:     ers.getCompanyInfo(),
		StartDate:   startDate,
		EndDate:     endDate,
		Currency:    ers.companyProfile.Currency,
		GeneratedAt: time.Now(),
	}

	// Get cash and cash equivalent accounts
	cashAccounts := ers.getCashAccounts()
	
	// Calculate beginning and ending cash balances
	prevDay := startDate.AddDate(0, 0, -1)
	cashFlow.BeginningCash = ers.calculateTotalCashBalance(cashAccounts, prevDay)
	cashFlow.EndingCash = ers.calculateTotalCashBalance(cashAccounts, endDate)

	// Build operating activities section
	operatingItems := ers.calculateOperatingCashFlow(startDate, endDate)
	cashFlow.OperatingActivities = CashFlowSection{
		Name:  "Operating Activities",
		Items: operatingItems,
		Total: ers.sumCashFlowItems(operatingItems),
	}

	// Build investing activities section  
	investingItems := ers.calculateInvestingCashFlow(startDate, endDate)
	cashFlow.InvestingActivities = CashFlowSection{
		Name:  "Investing Activities",
		Items: investingItems,
		Total: ers.sumCashFlowItems(investingItems),
	}

	// Build financing activities section
	financingItems := ers.calculateFinancingCashFlow(startDate, endDate)
	cashFlow.FinancingActivities = CashFlowSection{
		Name:  "Financing Activities", 
		Items: financingItems,
		Total: ers.sumCashFlowItems(financingItems),
	}

	// Calculate net cash flow
	cashFlow.NetCashFlow = cashFlow.OperatingActivities.Total + 
						   cashFlow.InvestingActivities.Total + 
						   cashFlow.FinancingActivities.Total

	return cashFlow, nil
}

// GenerateSalesSummary creates comprehensive sales analytics
func (ers *EnhancedReportService) GenerateSalesSummary(startDate, endDate time.Time, groupBy string) (*SalesSummaryData, error) {
	// Query sales data
	var sales []models.Sale
	if err := ers.db.Preload("Customer").
		Preload("SaleItems").
		Preload("SaleItems.Product").
		Preload("SalesPerson").
		Where("date BETWEEN ? AND ?", startDate, endDate).
		Find(&sales).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch sales data: %v", err)
	}

	// Initialize sales summary
	summary := &SalesSummaryData{
		Company:           ers.getCompanyInfo(),
		StartDate:         startDate,
		EndDate:           endDate,
		Currency:          ers.companyProfile.Currency,
		TotalTransactions: int64(len(sales)),
		GeneratedAt:       time.Now(),
	}

	// Calculate basic metrics
	customerSet := make(map[uint]bool)
	productMap := make(map[uint]*ProductSalesData)
	customerMap := make(map[uint]*CustomerSalesData)
	periodMap := make(map[string]*PeriodData)
	statusMap := make(map[string]*StatusData)

	for _, sale := range sales {
		summary.TotalRevenue += sale.TotalAmount
		customerSet[sale.CustomerID] = true

		// Process customer data
		if customerData, exists := customerMap[sale.CustomerID]; exists {
			customerData.TotalAmount += sale.TotalAmount
			customerData.TransactionCount++
			if sale.Date.After(customerData.LastOrderDate) {
				customerData.LastOrderDate = sale.Date
			}
			if sale.Date.Before(customerData.FirstOrderDate) {
				customerData.FirstOrderDate = sale.Date
			}
		} else {
			customerMap[sale.CustomerID] = &CustomerSalesData{
				CustomerID:       sale.CustomerID,
				CustomerName:     sale.Customer.Name,
				TotalAmount:      sale.TotalAmount,
				TransactionCount: 1,
				LastOrderDate:    sale.Date,
				FirstOrderDate:   sale.Date,
			}
		}

		// Process period data
		period := ers.formatPeriod(sale.Date, groupBy)
		if periodData, exists := periodMap[period]; exists {
			periodData.Amount += sale.TotalAmount
			periodData.Transactions++
		} else {
			periodMap[period] = &PeriodData{
				Period:       period,
				Amount:       sale.TotalAmount,
				Transactions: 1,
				StartDate:    ers.getPeriodStart(sale.Date, groupBy),
				EndDate:      ers.getPeriodEnd(sale.Date, groupBy),
			}
		}

		// Process status data
		if statusData, exists := statusMap[sale.Status]; exists {
			statusData.Count++
			statusData.Amount += sale.TotalAmount
		} else {
			statusMap[sale.Status] = &StatusData{
				Status: sale.Status,
				Count:  1,
				Amount: sale.TotalAmount,
			}
		}

		// Process product data
		for _, item := range sale.SaleItems {
			if productData, exists := productMap[item.ProductID]; exists {
				productData.QuantitySold += int64(item.Quantity)
				productData.TotalAmount += item.LineTotal
				productData.TransactionCount++
			} else {
				productMap[item.ProductID] = &ProductSalesData{
					ProductID:        item.ProductID,
					ProductName:      item.Product.Name,
					QuantitySold:     int64(item.Quantity),
					TotalAmount:      item.LineTotal,
					TransactionCount: 1,
				}
			}
		}
	}

	// Calculate derived metrics
	summary.TotalCustomers = int64(len(customerSet))
	if summary.TotalTransactions > 0 {
		summary.AverageOrderValue = summary.TotalRevenue / float64(summary.TotalTransactions)
	}

	// Calculate average prices and order values
	for _, customerData := range customerMap {
		if customerData.TransactionCount > 0 {
			customerData.AverageOrder = customerData.TotalAmount / float64(customerData.TransactionCount)
		}
	}

	for _, productData := range productMap {
		if productData.QuantitySold > 0 {
			productData.AveragePrice = productData.TotalAmount / float64(productData.QuantitySold)
		}
	}

	// Calculate percentages for status data
	for _, statusData := range statusMap {
		if summary.TotalRevenue > 0 {
			statusData.Percentage = (statusData.Amount / summary.TotalRevenue) * 100
		}
	}

	// Convert maps to slices and sort
	summary.SalesByCustomer = ers.convertAndSortCustomerData(customerMap)
	summary.SalesByProduct = ers.convertAndSortProductData(productMap)
	summary.SalesByPeriod = ers.convertAndSortPeriodData(periodMap)
	summary.SalesByStatus = ers.convertStatusData(statusMap)

	// Build top performers
	summary.TopPerformers = TopPerformersData{
		TopCustomers: ers.getTopCustomers(summary.SalesByCustomer, 10),
		TopProducts:  ers.getTopProducts(summary.SalesByProduct, 10),
		TopSalespeople: ers.getTopSalespeople(sales),
	}

	// Calculate growth analysis
	summary.GrowthAnalysis = ers.calculateSalesGrowth(summary.SalesByPeriod, groupBy)

	return summary, nil
}

// GeneratePurchaseSummary creates comprehensive purchase analytics
func (ers *EnhancedReportService) GeneratePurchaseSummary(startDate, endDate time.Time, groupBy string) (*PurchaseSummaryData, error) {
	// Query purchase data
	var purchases []models.Purchase
	if err := ers.db.Preload("Vendor").
		Preload("PurchaseItems").
		Preload("PurchaseItems.Product").
		Where("date BETWEEN ? AND ?", startDate, endDate).
		Find(&purchases).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch purchase data: %v", err)
	}

	// Initialize purchase summary
	summary := &PurchaseSummaryData{
		Company:           ers.getCompanyInfo(),
		StartDate:         startDate,
		EndDate:           endDate,
		Currency:          ers.companyProfile.Currency,
		TotalTransactions: int64(len(purchases)),
		GeneratedAt:       time.Now(),
	}

	// Calculate metrics similar to sales summary
	vendorSet := make(map[uint]bool)
	vendorMap := make(map[uint]*VendorPurchaseData)
	periodMap := make(map[string]*PeriodData)
	statusMap := make(map[string]*StatusData)
	categoryMap := make(map[uint]*CategoryPurchaseData)

	for _, purchase := range purchases {
		summary.TotalPurchases += purchase.TotalAmount
		vendorSet[purchase.VendorID] = true

		// Process vendor data
		if vendorData, exists := vendorMap[purchase.VendorID]; exists {
			vendorData.TotalAmount += purchase.TotalAmount
			vendorData.TransactionCount++
			if purchase.Date.After(vendorData.LastOrderDate) {
				vendorData.LastOrderDate = purchase.Date
			}
			if purchase.Date.Before(vendorData.FirstOrderDate) {
				vendorData.FirstOrderDate = purchase.Date
			}
		} else {
			vendorMap[purchase.VendorID] = &VendorPurchaseData{
				VendorID:         purchase.VendorID,
				VendorName:       purchase.Vendor.Name,
				TotalAmount:      purchase.TotalAmount,
				TransactionCount: 1,
				LastOrderDate:    purchase.Date,
				FirstOrderDate:   purchase.Date,
			}
		}

		// Process period data
		period := ers.formatPeriod(purchase.Date, groupBy)
		if periodData, exists := periodMap[period]; exists {
			periodData.Amount += purchase.TotalAmount
			periodData.Transactions++
		} else {
			periodMap[period] = &PeriodData{
				Period:       period,
				Amount:       purchase.TotalAmount,
				Transactions: 1,
				StartDate:    ers.getPeriodStart(purchase.Date, groupBy),
				EndDate:      ers.getPeriodEnd(purchase.Date, groupBy),
			}
		}

		// Process status data
		if statusData, exists := statusMap[purchase.Status]; exists {
			statusData.Count++
			statusData.Amount += purchase.TotalAmount
		} else {
			statusMap[purchase.Status] = &StatusData{
				Status: purchase.Status,
				Count:  1,
				Amount: purchase.TotalAmount,
			}
		}
	}

	// Calculate derived metrics
	summary.TotalVendors = int64(len(vendorSet))
	if summary.TotalTransactions > 0 {
		summary.AveragePurchaseValue = summary.TotalPurchases / float64(summary.TotalTransactions)
	}

	// Calculate average order values
	for _, vendorData := range vendorMap {
		if vendorData.TransactionCount > 0 {
			vendorData.AverageOrder = vendorData.TotalAmount / float64(vendorData.TransactionCount)
		}
	}

	// Convert maps to slices and sort
	summary.PurchasesByVendor = ers.convertAndSortVendorData(vendorMap)
	summary.PurchasesByPeriod = ers.convertAndSortPeriodData(periodMap)
	summary.PurchasesByStatus = ers.convertStatusData(statusMap)

	// Build top vendors
	summary.TopVendors = TopVendorsData{
		TopVendors:    ers.getTopVendors(summary.PurchasesByVendor, 10),
		TopCategories: ers.convertCategoryData(categoryMap),
	}

	// Calculate cost analysis
	summary.CostAnalysis = ers.calculateCostAnalysis(purchases)

	return summary, nil
}

// Helper methods for balance sheet construction
func (ers *EnhancedReportService) buildAssetsSection(assets []BalanceSheetItem) BalanceSheetSection {
	section := BalanceSheetSection{Name: "Assets"}
	
	var currentAssets, fixedAssets []BalanceSheetItem
	var currentTotal, fixedTotal float64

	for _, asset := range assets {
		if asset.Category == models.CategoryCurrentAsset {
			currentAssets = append(currentAssets, asset)
			currentTotal += asset.Balance
		} else if asset.Category == models.CategoryFixedAsset {
			fixedAssets = append(fixedAssets, asset)
			fixedTotal += asset.Balance
		}
	}

	section.Items = append(currentAssets, fixedAssets...)
	section.Subtotals = []BalanceSheetSubtotal{
		{Name: "Total Current Assets", Amount: currentTotal, Category: models.CategoryCurrentAsset},
		{Name: "Total Fixed Assets", Amount: fixedTotal, Category: models.CategoryFixedAsset},
	}
	section.Total = currentTotal + fixedTotal

	return section
}

func (ers *EnhancedReportService) buildLiabilitiesSection(liabilities []BalanceSheetItem) BalanceSheetSection {
	section := BalanceSheetSection{Name: "Liabilities"}
	
	var currentLiabilities, longTermLiabilities []BalanceSheetItem
	var currentTotal, longTermTotal float64

	for _, liability := range liabilities {
		if liability.Category == models.CategoryCurrentLiability {
			currentLiabilities = append(currentLiabilities, liability)
			currentTotal += liability.Balance
		} else if liability.Category == models.CategoryLongTermLiability {
			longTermLiabilities = append(longTermLiabilities, liability)
			longTermTotal += liability.Balance
		}
	}

	section.Items = append(currentLiabilities, longTermLiabilities...)
	section.Subtotals = []BalanceSheetSubtotal{
		{Name: "Total Current Liabilities", Amount: currentTotal, Category: models.CategoryCurrentLiability},
		{Name: "Total Long-term Liabilities", Amount: longTermTotal, Category: models.CategoryLongTermLiability},
	}
	section.Total = currentTotal + longTermTotal

	return section
}

func (ers *EnhancedReportService) buildEquitySection(equity []BalanceSheetItem) BalanceSheetSection {
	section := BalanceSheetSection{Name: "Equity"}
	
	var total float64
	for _, eq := range equity {
		total += eq.Balance
	}

	section.Items = equity
	section.Total = total

	return section
}

// CalculateAccountBalance calculates account balance as of a specific date (exported for use by report service)
func (ers *EnhancedReportService) CalculateAccountBalance(accountID uint, asOfDate time.Time) float64 {
	return ers.calculateAccountBalance(accountID, asOfDate)
}

// Helper method to calculate account balance as of a specific date
func (ers *EnhancedReportService) calculateAccountBalance(accountID uint, asOfDate time.Time) float64 {
	// Get account to determine normal balance type
	var account models.Account
	if err := ers.db.First(&account, accountID).Error; err != nil {
		return 0
	}

	// Calculate balance from journal entries up to asOfDate
	var totalDebits, totalCredits float64
	ers.db.Table("journal_entries").
		Joins("JOIN journals ON journal_entries.journal_id = journals.id").
		Where("journal_entries.account_id = ? AND journals.date <= ? AND journals.status = ?", 
			accountID, asOfDate, models.JournalStatusPosted).
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

// Helper method to calculate account balance for a specific period
func (ers *EnhancedReportService) calculateAccountBalanceForPeriod(accountID uint, startDate, endDate time.Time) float64 {
	// Get account to determine normal balance type
	var account models.Account
	if err := ers.db.First(&account, accountID).Error; err != nil {
		return 0
	}

	var totalDebits, totalCredits float64
	ers.db.Table("journal_entries").
		Joins("JOIN journals ON journal_entries.journal_id = journals.id").
		Where("journal_entries.account_id = ? AND journals.date BETWEEN ? AND ? AND journals.status = ?", 
			accountID, startDate, endDate, models.JournalStatusPosted).
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
		return ers.calculateAccountBalance(accountID, endDate)
	}
}

// Additional helper methods for formatting, sorting, and data conversion would be implemented here...
// These methods handle the detailed business logic for organizing and presenting the data

// formatPeriod formats a date according to the groupBy parameter
func (ers *EnhancedReportService) formatPeriod(date time.Time, groupBy string) string {
	switch groupBy {
	case "month":
		return date.Format("2006-01")
	case "quarter":
		quarter := (date.Month()-1)/3 + 1
		return fmt.Sprintf("%d-Q%d", date.Year(), quarter)
	case "year":
		return date.Format("2006")
	default:
		return date.Format("2006-01-02")
	}
}

// getPeriodStart gets the start date of a period
func (ers *EnhancedReportService) getPeriodStart(date time.Time, groupBy string) time.Time {
	switch groupBy {
	case "month":
		return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	case "quarter":
		quarter := (date.Month()-1)/3
		return time.Date(date.Year(), time.Month(quarter*3+1), 1, 0, 0, 0, 0, date.Location())
	case "year":
		return time.Date(date.Year(), 1, 1, 0, 0, 0, 0, date.Location())
	default:
		return date
	}
}

// getPeriodEnd gets the end date of a period
func (ers *EnhancedReportService) getPeriodEnd(date time.Time, groupBy string) time.Time {
	switch groupBy {
	case "month":
		return time.Date(date.Year(), date.Month()+1, 0, 23, 59, 59, 0, date.Location())
	case "quarter":
		quarter := (date.Month()-1)/3
		endMonth := quarter*3 + 3
		return time.Date(date.Year(), time.Month(endMonth)+1, 0, 23, 59, 59, 0, date.Location())
	case "year":
		return time.Date(date.Year(), 12, 31, 23, 59, 59, 0, date.Location())
	default:
		return date
	}
}

// Placeholder implementations for complex helper methods
// These would be fully implemented based on specific business requirements

func (ers *EnhancedReportService) categorizePLItems(items []PLItem, category string) []PLItem {
	// Implementation would categorize P&L items by type
	return items
}

func (ers *EnhancedReportService) getCashAccounts() []models.Account {
	// Implementation would get cash and cash equivalent accounts
	var accounts []models.Account
	ers.db.Where("type = ? AND (category LIKE ? OR name LIKE ?)", 
		models.AccountTypeAsset, "%CASH%", "%cash%").Find(&accounts)
	return accounts
}

func (ers *EnhancedReportService) calculateTotalCashBalance(accounts []models.Account, date time.Time) float64 {
	var total float64
	for _, account := range accounts {
		total += ers.calculateAccountBalance(account.ID, date)
	}
	return total
}

func (ers *EnhancedReportService) calculateOperatingCashFlow(startDate, endDate time.Time) []CashFlowItem {
	var items []CashFlowItem
	
	// Get net income for the period
	profitLoss, err := ers.GenerateProfitLoss(startDate, endDate)
	if err == nil {
		items = append(items, CashFlowItem{
			Description: "Net Income",
			Amount:      profitLoss.NetIncome,
			Category:    "NET_INCOME",
		})
	}
	
	// Add back non-cash expenses (depreciation, amortization)
	depreciationAmount := ers.calculateDepreciationForPeriod(startDate, endDate)
	if depreciationAmount > 0 {
		items = append(items, CashFlowItem{
			Description: "Depreciation and Amortization",
			Amount:      depreciationAmount,
			Category:    "NON_CASH_EXPENSES",
		})
	}
	
	// Calculate changes in working capital
	workingCapitalChanges := ers.calculateWorkingCapitalChanges(startDate, endDate)
	for _, change := range workingCapitalChanges {
		items = append(items, change)
	}
	
	// Add other operating cash flow items
	operatingItems := ers.getOtherOperatingCashFlowItems(startDate, endDate)
	items = append(items, operatingItems...)
	
	return items
}

func (ers *EnhancedReportService) calculateInvestingCashFlow(startDate, endDate time.Time) []CashFlowItem {
	var items []CashFlowItem
	
	// Capital expenditures (purchases of fixed assets)
	capitalExpenditures := ers.calculateCapitalExpenditures(startDate, endDate)
	if capitalExpenditures != 0 {
		items = append(items, CashFlowItem{
			Description: "Capital Expenditures",
			Amount:      -capitalExpenditures, // Negative because it's cash outflow
			Category:    "CAPITAL_EXPENDITURES",
		})
	}
	
	// Asset disposals
	assetDisposals := ers.calculateAssetDisposals(startDate, endDate)
	if assetDisposals > 0 {
		items = append(items, CashFlowItem{
			Description: "Proceeds from Asset Sales",
			Amount:      assetDisposals,
			Category:    "ASSET_DISPOSALS",
		})
	}
	
	// Investment activities
	investmentActivities := ers.calculateInvestmentActivities(startDate, endDate)
	for _, activity := range investmentActivities {
		items = append(items, activity)
	}
	
	return items
}

func (ers *EnhancedReportService) calculateFinancingCashFlow(startDate, endDate time.Time) []CashFlowItem {
	var items []CashFlowItem
	
	// Debt activities
	debtActivities := ers.calculateDebtActivities(startDate, endDate)
	for _, activity := range debtActivities {
		items = append(items, activity)
	}
	
	// Equity activities
	equityActivities := ers.calculateEquityActivities(startDate, endDate)
	for _, activity := range equityActivities {
		items = append(items, activity)
	}
	
	// Dividend payments
	dividendPayments := ers.calculateDividendPayments(startDate, endDate)
	if dividendPayments > 0 {
		items = append(items, CashFlowItem{
			Description: "Dividend Payments",
			Amount:      -dividendPayments, // Negative because it's cash outflow
			Category:    "DIVIDENDS",
		})
	}
	
	// Interest payments on debt
	interestPayments := ers.calculateInterestPayments(startDate, endDate)
	if interestPayments > 0 {
		items = append(items, CashFlowItem{
			Description: "Interest Paid",
			Amount:      -interestPayments, // Negative because it's cash outflow
			Category:    "INTEREST_PAID",
		})
	}
	
	return items
}

func (ers *EnhancedReportService) sumCashFlowItems(items []CashFlowItem) float64 {
	var total float64
	for _, item := range items {
		total += item.Amount
	}
	return total
}

// Data conversion and sorting helper methods
func (ers *EnhancedReportService) convertAndSortCustomerData(customerMap map[uint]*CustomerSalesData) []CustomerSalesData {
	var customers []CustomerSalesData
	for _, data := range customerMap {
		customers = append(customers, *data)
	}
	sort.Slice(customers, func(i, j int) bool {
		return customers[i].TotalAmount > customers[j].TotalAmount
	})
	return customers
}

func (ers *EnhancedReportService) convertAndSortProductData(productMap map[uint]*ProductSalesData) []ProductSalesData {
	var products []ProductSalesData
	for _, data := range productMap {
		products = append(products, *data)
	}
	sort.Slice(products, func(i, j int) bool {
		return products[i].TotalAmount > products[j].TotalAmount
	})
	return products
}

func (ers *EnhancedReportService) convertAndSortPeriodData(periodMap map[string]*PeriodData) []PeriodData {
	var periods []PeriodData
	for _, data := range periodMap {
		periods = append(periods, *data)
	}
	sort.Slice(periods, func(i, j int) bool {
		return periods[i].Period < periods[j].Period
	})
	return periods
}

func (ers *EnhancedReportService) convertStatusData(statusMap map[string]*StatusData) []StatusData {
	var statuses []StatusData
	for _, data := range statusMap {
		statuses = append(statuses, *data)
	}
	return statuses
}

func (ers *EnhancedReportService) convertAndSortVendorData(vendorMap map[uint]*VendorPurchaseData) []VendorPurchaseData {
	var vendors []VendorPurchaseData
	for _, data := range vendorMap {
		vendors = append(vendors, *data)
	}
	sort.Slice(vendors, func(i, j int) bool {
		return vendors[i].TotalAmount > vendors[j].TotalAmount
	})
	return vendors
}

func (ers *EnhancedReportService) convertCategoryData(categoryMap map[uint]*CategoryPurchaseData) []CategoryPurchaseData {
	var categories []CategoryPurchaseData
	for _, data := range categoryMap {
		categories = append(categories, *data)
	}
	return categories
}

func (ers *EnhancedReportService) getTopCustomers(customers []CustomerSalesData, limit int) []CustomerSalesData {
	if len(customers) <= limit {
		return customers
	}
	return customers[:limit]
}

func (ers *EnhancedReportService) getTopProducts(products []ProductSalesData, limit int) []ProductSalesData {
	if len(products) <= limit {
		return products
	}
	return products[:limit]
}

func (ers *EnhancedReportService) getTopSalespeople(sales []models.Sale) []SalespersonData {
	// Implementation would calculate top salespeople performance
	return []SalespersonData{}
}

func (ers *EnhancedReportService) getTopVendors(vendors []VendorPurchaseData, limit int) []VendorPurchaseData {
	if len(vendors) <= limit {
		return vendors
	}
	return vendors[:limit]
}

func (ers *EnhancedReportService) calculateSalesGrowth(periods []PeriodData, groupBy string) GrowthAnalysisData {
	growthData := GrowthAnalysisData{
		TrendDirection: "STABLE", // Default
	}
	
	if len(periods) < 2 {
		return growthData
	}
	
	// Calculate different growth metrics based on available periods
	switch groupBy {
	case "month":
		// Calculate month-over-month growth
		if len(periods) >= 2 {
			current := periods[len(periods)-1].Amount
			previous := periods[len(periods)-2].Amount
			if previous != 0 {
				growthData.MonthOverMonth = ((current - previous) / previous) * 100
			}
		}
		
		// Calculate quarter-over-quarter if we have enough data
		if len(periods) >= 3 {
			currentQuarter := ers.calculateQuarterlyAverage(periods[len(periods)-3:])
			if len(periods) >= 6 {
				prevQuarter := ers.calculateQuarterlyAverage(periods[len(periods)-6:len(periods)-3])
				if prevQuarter != 0 {
					growthData.QuarterOverQuarter = ((currentQuarter - prevQuarter) / prevQuarter) * 100
				}
			}
		}
		
		// Calculate year-over-year if we have 12+ months
		if len(periods) >= 12 {
			current := periods[len(periods)-1].Amount
			yearAgo := periods[len(periods)-12].Amount
			if yearAgo != 0 {
				growthData.YearOverYear = ((current - yearAgo) / yearAgo) * 100
			}
		}
		
	case "quarter":
		// Calculate quarter-over-quarter growth
		if len(periods) >= 2 {
			current := periods[len(periods)-1].Amount
			previous := periods[len(periods)-2].Amount
			if previous != 0 {
				growthData.QuarterOverQuarter = ((current - previous) / previous) * 100
			}
		}
		
		// Calculate year-over-year if we have 4+ quarters
		if len(periods) >= 4 {
			current := periods[len(periods)-1].Amount
			yearAgo := periods[len(periods)-4].Amount
			if yearAgo != 0 {
				growthData.YearOverYear = ((current - yearAgo) / yearAgo) * 100
			}
		}
		
	case "year":
		// Calculate year-over-year growth
		if len(periods) >= 2 {
			current := periods[len(periods)-1].Amount
			previous := periods[len(periods)-2].Amount
			if previous != 0 {
				growthData.YearOverYear = ((current - previous) / previous) * 100
			}
		}
	}
	
	// Determine trend direction based on most recent growth
	mostRecentGrowth := float64(0)
	if growthData.MonthOverMonth != 0 {
		mostRecentGrowth = growthData.MonthOverMonth
	} else if growthData.QuarterOverQuarter != 0 {
		mostRecentGrowth = growthData.QuarterOverQuarter
	} else if growthData.YearOverYear != 0 {
		mostRecentGrowth = growthData.YearOverYear
	}
	
	if mostRecentGrowth > 5 {
		growthData.TrendDirection = "UP"
	} else if mostRecentGrowth < -5 {
		growthData.TrendDirection = "DOWN"
	} else {
		growthData.TrendDirection = "STABLE"
	}
	
	// Calculate seasonality index (simplified)
	if len(periods) >= 12 {
		growthData.SeasonalityIndex = ers.calculateSeasonalityIndex(periods)
	}
	
	return growthData
}

func (ers *EnhancedReportService) calculateCostAnalysis(purchases []models.Purchase) CostAnalysisData {
	// Implementation would calculate cost analysis metrics
	var totalCost float64
	for _, purchase := range purchases {
		totalCost += purchase.TotalAmount
	}
	
	return CostAnalysisData{
		TotalCostOfGoods: totalCost,
	}
}

// ========== ENHANCED HELPER METHODS FOR PRODUCTION-READY P&L ==========

// isOperatingRevenue determines if an account category represents operating revenue
func (ers *EnhancedReportService) isOperatingRevenue(category string) bool {
	operatingRevenueCategories := []string{
		models.CategoryOperatingRevenue,
		models.CategoryServiceRevenue,
		models.CategorySalesRevenue,
	}
	
	for _, cat := range operatingRevenueCategories {
		if category == cat {
			return true
		}
	}
	return false
}

// isCOGS determines if an account category represents Cost of Goods Sold
func (ers *EnhancedReportService) isCOGS(category string) bool {
	cogsCategories := []string{
		models.CategoryCostOfGoodsSold,
		models.CategoryDirectMaterial,
		models.CategoryDirectLabor,
		models.CategoryManufacturingOverhead,
		models.CategoryFreightIn,
	}
	
	for _, cat := range cogsCategories {
		if category == cat {
			return true
		}
	}
	
	// Fallback to name-based matching for legacy accounts
	return strings.Contains(strings.ToLower(category), "cogs") ||
		   strings.Contains(strings.ToLower(category), "cost of goods")
}

// isOperatingExpense determines if an account category represents operating expenses
func (ers *EnhancedReportService) isOperatingExpense(category string) bool {
	operatingExpenseCategories := []string{
		models.CategoryOperatingExpense,
		models.CategoryAdministrativeExp,
		models.CategorySellingExpense,
		models.CategoryMarketingExpense,
		models.CategoryGeneralExpense,
		models.CategoryDepreciationExp,
		models.CategoryAmortizationExp,
		models.CategoryBadDebtExpense,
	}
	
	for _, cat := range operatingExpenseCategories {
		if category == cat {
			return true
		}
	}
	return false
}

// isTaxExpense determines if an account category represents tax expenses
func (ers *EnhancedReportService) isTaxExpense(category string) bool {
	return category == models.CategoryTaxExpense ||
		   strings.Contains(strings.ToLower(category), "tax")
}

// calculateItemPercentages calculates percentage of revenue for P&L items
func (ers *EnhancedReportService) calculateItemPercentages(items []PLItem, totalRevenue float64) {
	for i := range items {
		if totalRevenue != 0 {
			items[i].Percentage = (items[i].Amount / totalRevenue) * 100
		} else {
			items[i].Percentage = 0
		}
	}
}

// getSharesOutstanding retrieves shares outstanding from equity accounts
func (ers *EnhancedReportService) getSharesOutstanding() float64 {
	// First try to get shares outstanding from company profile
	if ers.companyProfile != nil && ers.companyProfile.SharesOutstanding > 0 {
		return ers.companyProfile.SharesOutstanding
	}
	
	// Try to find share capital account by multiple criteria
	var shareCapitalAccount models.Account
	err := ers.db.Where("type = ? AND (category LIKE ? OR category LIKE ? OR name LIKE ? OR name LIKE ?) AND is_active = ?", 
		models.AccountTypeEquity, "%SHARE_CAPITAL%", "%MODAL_SAHAM%", "%share%", "%saham%", true).First(&shareCapitalAccount).Error
	
	// Handle record not found gracefully - this is normal if no share capital accounts exist
	// Only log unexpected database errors, not "record not found" which is normal
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Unexpected error querying share capital account: %v", err)
	}
	
	if err == nil && shareCapitalAccount.Balance > 0 {
		// Try to get par value from company profile, otherwise use reasonable default
		parValue := float64(1000) // Default IDR 1000 per share
		if ers.companyProfile != nil && ers.companyProfile.ParValuePerShare > 0 {
			parValue = ers.companyProfile.ParValuePerShare
		}
		return shareCapitalAccount.Balance / parValue
	}
	
	// Try to calculate from total paid-in capital or total equity
	var totalPaidInCapital float64
	ers.db.Model(&models.Account{}).Where("type = ? AND (category LIKE ? OR category LIKE ?) AND is_active = ?", 
		models.AccountTypeEquity, "%PAID_IN%", "%MODAL_DISETOR%", true).Select("COALESCE(SUM(balance), 0)").Scan(&totalPaidInCapital)
	
	if totalPaidInCapital > 0 {
		parValue := float64(1000) // Default par value
		if ers.companyProfile != nil && ers.companyProfile.ParValuePerShare > 0 {
			parValue = ers.companyProfile.ParValuePerShare
		}
		return totalPaidInCapital / parValue
	}
	
	// If no share capital data available, return 0 (EPS cannot be calculated)
	return 0
}

// GenerateComparativeProfitLoss creates comparative P&L analysis
func (ers *EnhancedReportService) GenerateComparativeProfitLoss(currentStart, currentEnd, priorStart, priorEnd time.Time) (*ProfitLossComparative, error) {
	// Generate current period P&L
	currentPL, err := ers.GenerateProfitLoss(currentStart, currentEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to generate current period P&L: %v", err)
	}
	
	// Generate prior period P&L
	priorPL, err := ers.GenerateProfitLoss(priorStart, priorEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prior period P&L: %v", err)
	}
	
	// Calculate variances
	variances := ers.calculatePLVariances(currentPL, priorPL)
	
	// Calculate trend analysis
	trendAnalysis := ers.calculatePLTrendAnalysis(currentPL, priorPL)
	
	return &ProfitLossComparative{
		CurrentPeriod: *currentPL,
		PriorPeriod:   *priorPL,
		Variances:     variances,
		TrendAnalysis: trendAnalysis,
		GeneratedAt:   time.Now(),
	}, nil
}

// calculatePLVariances calculates variances between current and prior periods
func (ers *EnhancedReportService) calculatePLVariances(current, prior *ProfitLossData) PLVarianceData {
	return PLVarianceData{
		Revenue: ers.calculateVariance(current.Revenue.Subtotal, prior.Revenue.Subtotal),
		CostOfGoodsSold: ers.calculateVariance(current.CostOfGoodsSold.Subtotal, prior.CostOfGoodsSold.Subtotal),
		GrossProfit: ers.calculateVariance(current.GrossProfit, prior.GrossProfit),
		OperatingExpenses: ers.calculateVariance(current.OperatingExpenses.Subtotal, prior.OperatingExpenses.Subtotal),
		OperatingIncome: ers.calculateVariance(current.OperatingIncome, prior.OperatingIncome),
		OtherIncome: ers.calculateVariance(current.OtherIncome.Subtotal, prior.OtherIncome.Subtotal),
		OtherExpenses: ers.calculateVariance(current.OtherExpenses.Subtotal, prior.OtherExpenses.Subtotal),
		NetIncome: ers.calculateVariance(current.NetIncome, prior.NetIncome),
		EBITDA: ers.calculateVariance(current.EBITDA, prior.EBITDA),
		EBIT: ers.calculateVariance(current.EBIT, prior.EBIT),
	}
}

// calculateVariance calculates variance between two values
func (ers *EnhancedReportService) calculateVariance(current, prior float64) PLVariance {
	absoluteChange := current - prior
	var percentChange float64
	var trend string
	
	if prior != 0 {
		percentChange = (absoluteChange / math.Abs(prior)) * 100
	} else if current != 0 {
		percentChange = 100 // 100% change from zero
	}
	
	// Determine trend
	if math.Abs(percentChange) < 5 {
		trend = "STABLE"
	} else if absoluteChange > 0 {
		trend = "INCREASING"
	} else {
		trend = "DECREASING"
	}
	
	return PLVariance{
		Current:        current,
		Prior:          prior,
		AbsoluteChange: absoluteChange,
		PercentChange:  percentChange,
		Trend:          trend,
	}
}

// calculatePLTrendAnalysis calculates comprehensive trend analysis
func (ers *EnhancedReportService) calculatePLTrendAnalysis(current, prior *ProfitLossData) PLTrendAnalysis {
	// Revenue growth rate
	revenueGrowthRate := float64(0)
	if prior.Revenue.Subtotal != 0 {
		revenueGrowthRate = ((current.Revenue.Subtotal - prior.Revenue.Subtotal) / prior.Revenue.Subtotal) * 100
	}
	
	// Profitability trend
	profitabilityTrend := "STABLE"
	if current.NetIncomeMargin > prior.NetIncomeMargin {
		profitabilityTrend = "IMPROVING"
	} else if current.NetIncomeMargin < prior.NetIncomeMargin {
		profitabilityTrend = "DECLINING"
	}
	
	// Cost management index (lower is better)
	costManagementIndex := float64(0)
	if current.Revenue.Subtotal != 0 {
		costManagementIndex = (current.CostOfGoodsSold.Subtotal + current.OperatingExpenses.Subtotal) / current.Revenue.Subtotal
	}
	
	// Operational efficiency (higher is better)
	operationalEfficiency := float64(0)
	if current.Revenue.Subtotal != 0 {
		operationalEfficiency = current.OperatingIncome / current.Revenue.Subtotal * 100
	}
	
	// Margin stability
	marginStability := "STABLE"
	marginDiff := math.Abs(current.NetIncomeMargin - prior.NetIncomeMargin)
	if marginDiff > 10 {
		marginStability = "VOLATILE"
	} else if marginDiff < 2 {
		marginStability = "STABLE"
	} else {
		marginStability = "MODERATE"
	}
	
	return PLTrendAnalysis{
		RevenueGrowthRate:     revenueGrowthRate,
		ProfitabilityTrend:    profitabilityTrend,
		CostManagementIndex:   costManagementIndex,
		OperationalEfficiency: operationalEfficiency,
		MarginStability:       marginStability,
	}
}

// ========== CASH FLOW STATEMENT HELPER METHODS ==========

// calculateDepreciationForPeriod calculates depreciation and amortization for the period
func (ers *EnhancedReportService) calculateDepreciationForPeriod(startDate, endDate time.Time) float64 {
	var totalDepreciation float64
	
	// Find depreciation expense accounts
	var depreciationAccounts []models.Account
	ers.db.Where("category = ? AND is_active = ?", 
		models.CategoryDepreciationExp, true).Find(&depreciationAccounts)
	
	for _, account := range depreciationAccounts {
		amount := ers.calculateAccountBalanceForPeriod(account.ID, startDate, endDate)
		totalDepreciation += amount
	}
	
	return totalDepreciation
}

// calculateWorkingCapitalChanges calculates changes in working capital components
func (ers *EnhancedReportService) calculateWorkingCapitalChanges(startDate, endDate time.Time) []CashFlowItem {
	var items []CashFlowItem
	
	// Calculate changes in accounts receivable
	receivableChange := ers.calculateAccountsReceivableChange(startDate, endDate)
	if receivableChange != 0 {
		items = append(items, CashFlowItem{
			Description: "Change in Accounts Receivable",
			Amount:      -receivableChange, // Negative because increase in AR decreases cash
			Category:    "WORKING_CAPITAL",
		})
	}
	
	// Calculate changes in inventory
	inventoryChange := ers.calculateInventoryChange(startDate, endDate)
	if inventoryChange != 0 {
		items = append(items, CashFlowItem{
			Description: "Change in Inventory",
			Amount:      -inventoryChange, // Negative because increase in inventory decreases cash
			Category:    "WORKING_CAPITAL",
		})
	}
	
	// Calculate changes in accounts payable
	payableChange := ers.calculateAccountsPayableChange(startDate, endDate)
	if payableChange != 0 {
		items = append(items, CashFlowItem{
			Description: "Change in Accounts Payable",
			Amount:      payableChange, // Positive because increase in AP increases cash
			Category:    "WORKING_CAPITAL",
		})
	}
	
	// Calculate changes in prepaid expenses
	prepaidChange := ers.calculatePrepaidExpensesChange(startDate, endDate)
	if prepaidChange != 0 {
		items = append(items, CashFlowItem{
			Description: "Change in Prepaid Expenses",
			Amount:      -prepaidChange, // Negative because increase in prepaid decreases cash
			Category:    "WORKING_CAPITAL",
		})
	}
	
	return items
}

// calculateAccountsReceivableChange calculates the change in accounts receivable
func (ers *EnhancedReportService) calculateAccountsReceivableChange(startDate, endDate time.Time) float64 {
	var arAccounts []models.Account
	ers.db.Where("type = ? AND name ILIKE ? AND is_active = ?", 
		models.AccountTypeAsset, "%receivable%", true).Find(&arAccounts)
	
	var totalChange float64
	for _, account := range arAccounts {
		startBalance := ers.calculateAccountBalance(account.ID, startDate.AddDate(0, 0, -1))
		endBalance := ers.calculateAccountBalance(account.ID, endDate)
		totalChange += (endBalance - startBalance)
	}
	
	return totalChange
}

// calculateInventoryChange calculates the change in inventory
func (ers *EnhancedReportService) calculateInventoryChange(startDate, endDate time.Time) float64 {
	var inventoryAccounts []models.Account
	ers.db.Where("type = ? AND name ILIKE ? AND is_active = ?", 
		models.AccountTypeAsset, "%inventory%", true).Find(&inventoryAccounts)
	
	var totalChange float64
	for _, account := range inventoryAccounts {
		startBalance := ers.calculateAccountBalance(account.ID, startDate.AddDate(0, 0, -1))
		endBalance := ers.calculateAccountBalance(account.ID, endDate)
		totalChange += (endBalance - startBalance)
	}
	
	return totalChange
}

// calculateAccountsPayableChange calculates the change in accounts payable
func (ers *EnhancedReportService) calculateAccountsPayableChange(startDate, endDate time.Time) float64 {
	var apAccounts []models.Account
	ers.db.Where("type = ? AND name ILIKE ? AND is_active = ?", 
		models.AccountTypeLiability, "%payable%", true).Find(&apAccounts)
	
	var totalChange float64
	for _, account := range apAccounts {
		startBalance := ers.calculateAccountBalance(account.ID, startDate.AddDate(0, 0, -1))
		endBalance := ers.calculateAccountBalance(account.ID, endDate)
		totalChange += (endBalance - startBalance)
	}
	
	return totalChange
}

// calculatePrepaidExpensesChange calculates the change in prepaid expenses
func (ers *EnhancedReportService) calculatePrepaidExpensesChange(startDate, endDate time.Time) float64 {
	var prepaidAccounts []models.Account
	ers.db.Where("type = ? AND name ILIKE ? AND is_active = ?", 
		models.AccountTypeAsset, "%prepaid%", true).Find(&prepaidAccounts)
	
	var totalChange float64
	for _, account := range prepaidAccounts {
		startBalance := ers.calculateAccountBalance(account.ID, startDate.AddDate(0, 0, -1))
		endBalance := ers.calculateAccountBalance(account.ID, endDate)
		totalChange += (endBalance - startBalance)
	}
	
	return totalChange
}

// getOtherOperatingCashFlowItems gets other operating cash flow items
func (ers *EnhancedReportService) getOtherOperatingCashFlowItems(startDate, endDate time.Time) []CashFlowItem {
	var items []CashFlowItem
	
	// Add other operating activities like tax payments, interest received, etc.
	// This would be customized based on business needs
	
	return items
}

// calculateCapitalExpenditures calculates capital expenditures for the period
func (ers *EnhancedReportService) calculateCapitalExpenditures(startDate, endDate time.Time) float64 {
	var totalCapex float64
	
	// Look for purchases of fixed assets in journal entries
	var fixedAssetAccounts []models.Account
	ers.db.Where("type = ? AND category = ? AND is_active = ?", 
		models.AccountTypeAsset, models.CategoryFixedAsset, true).Find(&fixedAssetAccounts)
	
	for _, account := range fixedAssetAccounts {
		// Calculate net additions (debits) to fixed asset accounts during period
		var totalDebits float64
		ers.db.Table("journal_entries").
			Joins("JOIN journals ON journal_entries.journal_id = journals.id").
			Where("journal_entries.account_id = ? AND journals.date BETWEEN ? AND ? AND journals.status = ?", 
				account.ID, startDate, endDate, models.JournalStatusPosted).
			Select("COALESCE(SUM(journal_entries.debit_amount), 0)").
			Row().Scan(&totalDebits)
		
		totalCapex += totalDebits
	}
	
	return totalCapex
}

// calculateAssetDisposals calculates proceeds from asset disposals
func (ers *EnhancedReportService) calculateAssetDisposals(startDate, endDate time.Time) float64 {
	// This would typically come from gain/loss on asset disposal accounts
	var totalDisposals float64
	
	var disposalAccounts []models.Account
	ers.db.Where("category IN (?) AND is_active = ?", 
		[]string{models.CategoryGainOnSale, models.CategoryLossOnSale}, true).Find(&disposalAccounts)
	
	for _, account := range disposalAccounts {
		amount := ers.calculateAccountBalanceForPeriod(account.ID, startDate, endDate)
		totalDisposals += math.Abs(amount) // Take absolute value as this represents cash proceeds
	}
	
	return totalDisposals
}

// calculateInvestmentActivities calculates other investment activities
func (ers *EnhancedReportService) calculateInvestmentActivities(startDate, endDate time.Time) []CashFlowItem {
	var items []CashFlowItem
	
	// Look for investment accounts changes
	var investmentAccounts []models.Account
	ers.db.Where("type = ? AND category = ? AND is_active = ?", 
		models.AccountTypeAsset, models.CategoryInvestmentAsset, true).Find(&investmentAccounts)
	
	for _, account := range investmentAccounts {
		startBalance := ers.calculateAccountBalance(account.ID, startDate.AddDate(0, 0, -1))
		endBalance := ers.calculateAccountBalance(account.ID, endDate)
		change := endBalance - startBalance
		
		if change != 0 {
			description := fmt.Sprintf("Investment in %s", account.Name)
			if change < 0 {
				description = fmt.Sprintf("Proceeds from %s", account.Name)
			}
			
			items = append(items, CashFlowItem{
				Description: description,
				Amount:      -change, // Negative because investment outflow
				Category:    "INVESTMENTS",
			})
		}
	}
	
	return items
}

// calculateDebtActivities calculates debt-related financing activities
func (ers *EnhancedReportService) calculateDebtActivities(startDate, endDate time.Time) []CashFlowItem {
	var items []CashFlowItem
	
	// Look for changes in long-term debt
	var debtAccounts []models.Account
	ers.db.Where("type = ? AND category = ? AND is_active = ?", 
		models.AccountTypeLiability, models.CategoryLongTermLiability, true).Find(&debtAccounts)
	
	for _, account := range debtAccounts {
		if strings.Contains(strings.ToLower(account.Name), "loan") || 
		   strings.Contains(strings.ToLower(account.Name), "debt") ||
		   strings.Contains(strings.ToLower(account.Name), "note") {
			
			startBalance := ers.calculateAccountBalance(account.ID, startDate.AddDate(0, 0, -1))
			endBalance := ers.calculateAccountBalance(account.ID, endDate)
			change := endBalance - startBalance
			
			if change != 0 {
				description := "Debt Repayment"
				if change > 0 {
					description = "Proceeds from Debt"
				}
				
				items = append(items, CashFlowItem{
					Description: description,
					Amount:      change, // Positive for new debt, negative for repayment
					Category:    "DEBT",
				})
			}
		}
	}
	
	return items
}

// calculateEquityActivities calculates equity-related financing activities
func (ers *EnhancedReportService) calculateEquityActivities(startDate, endDate time.Time) []CashFlowItem {
	var items []CashFlowItem
	
	// Look for changes in share capital
	var equityAccounts []models.Account
	ers.db.Where("type = ? AND category IN (?) AND is_active = ?", 
		models.AccountTypeEquity, []string{models.CategoryShareCapital, models.CategoryEquity}, true).Find(&equityAccounts)
	
	for _, account := range equityAccounts {
		if strings.Contains(strings.ToLower(account.Name), "capital") || 
		   strings.Contains(strings.ToLower(account.Name), "stock") ||
		   strings.Contains(strings.ToLower(account.Name), "share") {
			
			startBalance := ers.calculateAccountBalance(account.ID, startDate.AddDate(0, 0, -1))
			endBalance := ers.calculateAccountBalance(account.ID, endDate)
			change := endBalance - startBalance
			
			if change != 0 {
				description := "Stock Repurchase"
				if change > 0 {
					description = "Proceeds from Stock Issuance"
				}
				
				items = append(items, CashFlowItem{
					Description: description,
					Amount:      change, // Positive for new equity, negative for repurchase
					Category:    "EQUITY",
				})
			}
		}
	}
	
	return items
}

// calculateDividendPayments calculates dividend payments for the period
func (ers *EnhancedReportService) calculateDividendPayments(startDate, endDate time.Time) float64 {
	var totalDividends float64
	
	// Look for dividend expense accounts
	var dividendAccounts []models.Account
	ers.db.Where("(name ILIKE ? OR name ILIKE ?) AND is_active = ?", 
		"%dividend%", "%distribution%", true).Find(&dividendAccounts)
	
	for _, account := range dividendAccounts {
		amount := ers.calculateAccountBalanceForPeriod(account.ID, startDate, endDate)
		totalDividends += amount
	}
	
	return totalDividends
}

// calculateInterestPayments calculates interest payments on debt
func (ers *EnhancedReportService) calculateInterestPayments(startDate, endDate time.Time) float64 {
	var totalInterest float64
	
	// Look for interest expense accounts
	var interestAccounts []models.Account
	ers.db.Where("category = ? AND is_active = ?", 
		models.CategoryInterestExpense, true).Find(&interestAccounts)
	
	for _, account := range interestAccounts {
		amount := ers.calculateAccountBalanceForPeriod(account.ID, startDate, endDate)
		totalInterest += amount
	}
	
	return totalInterest
}

// calculateQuarterlyAverage calculates the average of a set of periods (typically 3 months for a quarter)
func (ers *EnhancedReportService) calculateQuarterlyAverage(periods []PeriodData) float64 {
	if len(periods) == 0 {
		return 0
	}
	
	var total float64
	for _, period := range periods {
		total += period.Amount
	}
	
	return total / float64(len(periods))
}

// calculateSeasonalityIndex calculates a simple seasonality index based on monthly data
func (ers *EnhancedReportService) calculateSeasonalityIndex(periods []PeriodData) float64 {
	if len(periods) < 12 {
		return 0 // Not enough data for seasonality analysis
	}
	
	// Group by month (assuming periods are monthly)
	monthlyTotals := make(map[int]float64)
	monthlyCounts := make(map[int]int)
	
	for _, period := range periods {
		// Parse period to extract month
		if len(period.Period) >= 7 { // Format: YYYY-MM
			if month := period.StartDate.Month(); month > 0 {
				monthlyTotals[int(month)] += period.Amount
				monthlyCounts[int(month)]++
			}
		}
	}
	
	// Calculate monthly averages
	monthlyAverages := make([]float64, 12)
	var yearlyAverage float64
	validMonths := 0
	
	for month := 1; month <= 12; month++ {
		if count := monthlyCounts[month]; count > 0 {
			monthlyAverages[month-1] = monthlyTotals[month] / float64(count)
			yearlyAverage += monthlyAverages[month-1]
			validMonths++
		}
	}
	
	if validMonths == 0 {
		return 0
	}
	
	yearlyAverage /= float64(validMonths)
	
	// Calculate seasonality index as coefficient of variation
	var variance float64
	for _, avg := range monthlyAverages {
		if avg > 0 {
			diff := avg - yearlyAverage
			variance += diff * diff
		}
	}
	
	variance /= float64(validMonths)
	stdDev := math.Sqrt(variance)
	
	if yearlyAverage == 0 {
		return 0
	}
	
// Return coefficient of variation as seasonality index
	return (stdDev / yearlyAverage) * 100
}

// ========== HELPER METHODS FOR DYNAMIC DEFAULT VALUES ==========

// getDefaultCompanyName returns the default company name from environment or fallback
func (ers *EnhancedReportService) getDefaultCompanyName() string {
	// Try to get from environment variable
	if companyName := os.Getenv("COMPANY_NAME"); companyName != "" {
		return companyName
	}
	
	// Try alternative environment variables
	if appName := os.Getenv("APP_NAME"); appName != "" {
		return appName + " Company"
	}
	
	if businessName := os.Getenv("BUSINESS_NAME"); businessName != "" {
		return businessName
	}
	
	// Default fallback that prompts user to update
	return "[Please update company name]"
}

// getDefaultState returns the default state/province from environment or fallback
func (ers *EnhancedReportService) getDefaultState() string {
	// Try to get from environment variable
	if state := os.Getenv("COMPANY_STATE"); state != "" {
		return state
	}
	
	if province := os.Getenv("COMPANY_PROVINCE"); province != "" {
		return province
	}
	
	// For Indonesia, common provinces as fallback
	if region := os.Getenv("DEFAULT_REGION"); region != "" {
		return region
	}
	
	// Default fallback that prompts user to update
	return "[Please update state/province]"
}

// getDefaultCompanyEmail returns the default company email from environment or fallback
func (ers *EnhancedReportService) getDefaultCompanyEmail() string {
	// Try to get from environment variable
	if email := os.Getenv("COMPANY_EMAIL"); email != "" {
		return email
	}
	
	if adminEmail := os.Getenv("ADMIN_EMAIL"); adminEmail != "" {
		return adminEmail
	}
	
	if contactEmail := os.Getenv("CONTACT_EMAIL"); contactEmail != "" {
		return contactEmail
	}
	
	// Default fallback that prompts user to update
	return "[Please update email address]"
}

// getDefaultCompanyPhone returns the default company phone from environment or fallback
func (ers *EnhancedReportService) getDefaultCompanyPhone() string {
	// Try to get from environment variable
	if phone := os.Getenv("COMPANY_PHONE"); phone != "" {
		return phone
	}
	
	if contactPhone := os.Getenv("CONTACT_PHONE"); contactPhone != "" {
		return contactPhone
	}
	
	if businessPhone := os.Getenv("BUSINESS_PHONE"); businessPhone != "" {
		return businessPhone
	}
	
	// Default fallback that prompts user to update
	return "[Please update phone number]"
}

// getDefaultCompanyAddress returns the default company address from environment or fallback
func (ers *EnhancedReportService) getDefaultCompanyAddress() string {
	// Try to get from environment variable
	if address := os.Getenv("COMPANY_ADDRESS"); address != "" {
		return address
	}
	
	if businessAddress := os.Getenv("BUSINESS_ADDRESS"); businessAddress != "" {
		return businessAddress
	}
	
	// Default fallback that prompts user to update
	return "[Please update company address]"
}

// getDefaultCompanyCity returns the default company city from environment or fallback
func (ers *EnhancedReportService) getDefaultCompanyCity() string {
	// Try to get from environment variable
	if city := os.Getenv("COMPANY_CITY"); city != "" {
		return city
	}
	
	if businessCity := os.Getenv("BUSINESS_CITY"); businessCity != "" {
		return businessCity
	}
	
	// Default fallback that prompts user to update
	return "[Please update city]"
}

// getDefaultCompanyWebsite returns the default company website from environment or fallback
func (ers *EnhancedReportService) getDefaultCompanyWebsite() string {
	// Try to get from environment variable
	if website := os.Getenv("COMPANY_WEBSITE"); website != "" {
		return website
	}
	
	if businessWebsite := os.Getenv("BUSINESS_WEBSITE"); businessWebsite != "" {
		return businessWebsite
	}
	
	if appUrl := os.Getenv("APP_URL"); appUrl != "" {
		return appUrl
	}
	
	// Default fallback that prompts user to update
	return "[Please update website]"
}

// getDefaultTaxNumber returns the default tax number from environment or fallback
func (ers *EnhancedReportService) getDefaultTaxNumber() string {
	// Try to get from environment variable
	if taxNumber := os.Getenv("COMPANY_TAX_NUMBER"); taxNumber != "" {
		return taxNumber
	}
	
	if npwp := os.Getenv("COMPANY_NPWP"); npwp != "" {
		return npwp
	}
	
	if businessTaxId := os.Getenv("BUSINESS_TAX_ID"); businessTaxId != "" {
		return businessTaxId
	}
	
	// Default fallback that prompts user to update
	return "[Please update tax number]"
}

// getDefaultPostalCode returns the default postal code from environment or fallback
func (ers *EnhancedReportService) getDefaultPostalCode() string {
	// Try to get from environment variable
	if postalCode := os.Getenv("COMPANY_POSTAL_CODE"); postalCode != "" {
		return postalCode
	}
	
	if zipCode := os.Getenv("COMPANY_ZIP_CODE"); zipCode != "" {
		return zipCode
	}
	
	// Default fallback that prompts user to update
	return "[Please update postal code]"
}

// getDefaultCurrency returns the default currency from environment or fallback
func (ers *EnhancedReportService) getDefaultCurrency() string {
	// Try to get from environment variable
	if currency := os.Getenv("DEFAULT_CURRENCY"); currency != "" {
		return currency
	}
	
	if companyCurrency := os.Getenv("COMPANY_CURRENCY"); companyCurrency != "" {
		return companyCurrency
	}
	
	// Default to IDR for Indonesian companies
	return "IDR"
}

// getDefaultCountry returns the default country from environment or fallback
func (ers *EnhancedReportService) getDefaultCountry() string {
	// Try to get from environment variable
	if country := os.Getenv("COMPANY_COUNTRY"); country != "" {
		return country
	}
	
	if defaultCountry := os.Getenv("DEFAULT_COUNTRY"); defaultCountry != "" {
		return defaultCountry
	}
	
	// Default to Indonesia
	return "Indonesia"
}
