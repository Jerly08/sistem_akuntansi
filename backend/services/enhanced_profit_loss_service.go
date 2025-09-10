package services

import (
	"context"
	"fmt"
	"strings"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

// Enhanced Profit & Loss Data Structures
type EnhancedPLSection struct {
	Name     string            `json:"name"`
	Items    []EnhancedPLItem  `json:"items"`
	Subtotal float64           `json:"subtotal"`
}

type EnhancedPLItem struct {
	AccountID      uint    `json:"account_id"`
	Code           string  `json:"code"`
	Name           string  `json:"name"`
	Amount         float64 `json:"amount"`
	Percentage     float64 `json:"percentage"`
	Category       string  `json:"category"`
	IsSubtotal     bool    `json:"is_subtotal"`
}

type EnhancedProfitLossData struct {
	Company     CompanyInfo    `json:"company"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     time.Time      `json:"end_date"`
	Currency    string         `json:"currency"`
	
	// Revenue Section
	Revenue struct {
		SalesRevenue    EnhancedPLSection `json:"sales_revenue"`
		ServiceRevenue  EnhancedPLSection `json:"service_revenue"`
		OtherRevenue    EnhancedPLSection `json:"other_revenue"`
		TotalRevenue    float64           `json:"total_revenue"`
	} `json:"revenue"`
	
	// COGS Section
	CostOfGoodsSold struct {
		DirectMaterials     EnhancedPLSection `json:"direct_materials"`
		DirectLabor         EnhancedPLSection `json:"direct_labor"`
		ManufacturingOH     EnhancedPLSection `json:"manufacturing_overhead"`
		OtherCOGS          EnhancedPLSection `json:"other_cogs"`
		TotalCOGS          float64           `json:"total_cogs"`
	} `json:"cost_of_goods_sold"`
	
	// Profitability Metrics
	GrossProfit       float64 `json:"gross_profit"`
	GrossProfitMargin float64 `json:"gross_profit_margin"`
	
	// Operating Expenses
	OperatingExpenses struct {
		Administrative   EnhancedPLSection `json:"administrative"`
		SellingMarketing EnhancedPLSection `json:"selling_marketing"`
		General          EnhancedPLSection `json:"general"`
		Depreciation     EnhancedPLSection `json:"depreciation"`
		TotalOpex        float64           `json:"total_operating_expenses"`
	} `json:"operating_expenses"`
	
	// Operating Performance
	OperatingIncome   float64 `json:"operating_income"`
	OperatingMargin   float64 `json:"operating_margin"`
	EBITDA           float64 `json:"ebitda"`
	EBITDAMargin     float64 `json:"ebitda_margin"`
	
	// Non-Operating Items
	OtherIncomeExpense struct {
		InterestIncome   EnhancedPLSection `json:"interest_income"`
		InterestExpense  EnhancedPLSection `json:"interest_expense"`
		OtherIncome      EnhancedPLSection `json:"other_income"`
		OtherExpense     EnhancedPLSection `json:"other_expense"`
		NetOtherIncome   float64           `json:"net_other_income"`
	} `json:"other_income_expense"`
	
	// Tax and Final Result
	IncomeBeforeTax   float64 `json:"income_before_tax"`
	TaxExpense        float64 `json:"tax_expense"`
	TaxRate          float64 `json:"tax_rate"`
	NetIncome        float64 `json:"net_income"`
	NetIncomeMargin  float64 `json:"net_income_margin"`
	
	// Additional Financial Metrics
	EarningsPerShare  float64 `json:"earnings_per_share"`
	SharesOutstanding float64 `json:"shares_outstanding"`
	
	GeneratedAt      time.Time `json:"generated_at"`
}

// Enhanced P&L Categories
const (
	// Revenue Categories
	CategorySalesRevenue        = "SALES_REVENUE"
	CategoryServiceRevenue      = "SERVICE_REVENUE"
	CategoryOtherOperatingRev   = "OTHER_OPERATING_REVENUE"
	CategoryNonOperatingRevenue = "NON_OPERATING_REVENUE"
	CategoryInterestIncome      = "INTEREST_INCOME"
	
	// COGS Categories  
	CategoryDirectMaterials     = "DIRECT_MATERIALS"
	CategoryDirectLabor         = "DIRECT_LABOR"
	CategoryManufacturingOH     = "MANUFACTURING_OVERHEAD"
	CategoryFreightIn           = "FREIGHT_IN"
	CategoryCostOfGoodsSold     = "COST_OF_GOODS_SOLD" // Generic COGS
	
	// Operating Expense Categories
	CategoryAdminExpense        = "ADMINISTRATIVE_EXPENSE"
	CategorySellingExpense      = "SELLING_MARKETING_EXPENSE"
	CategoryGeneralExpense      = "GENERAL_EXPENSE"
	CategoryDepreciationExp     = "DEPRECIATION_AMORTIZATION"
	CategoryOperatingExpense    = "OPERATING_EXPENSE" // Generic operating expense
	
	// Non-Operating Categories
	CategoryInterestExpense     = "INTEREST_EXPENSE"
	CategoryOtherNonOpExp      = "OTHER_NON_OPERATING_EXPENSE"
	CategoryTaxExpense         = "INCOME_TAX_EXPENSE"
)

type EnhancedProfitLossService struct {
	db          *gorm.DB
	accountRepo repositories.AccountRepository
}

func NewEnhancedProfitLossService(db *gorm.DB, accountRepo repositories.AccountRepository) *EnhancedProfitLossService {
	return &EnhancedProfitLossService{
		db:          db,
		accountRepo: accountRepo,
	}
}

// GenerateEnhancedProfitLoss creates a comprehensive P&L statement with proper accounting logic
func (epls *EnhancedProfitLossService) GenerateEnhancedProfitLoss(startDate, endDate time.Time) (*EnhancedProfitLossData, error) {
	// Get all active accounts
	ctx := context.Background()
	accounts, err := epls.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch accounts: %v", err)
	}

	// Initialize P&L structure
	profitLoss := &EnhancedProfitLossData{
		Company:     epls.getCompanyInfo(),
		StartDate:   startDate,
		EndDate:     endDate,
		Currency:    "IDR",
		GeneratedAt: time.Now(),
	}

	// Initialize item collections
	var revenueItems []EnhancedPLItem
	var cogsItems []EnhancedPLItem
	var operatingExpenseItems []EnhancedPLItem
	var nonOperatingIncomeItems []EnhancedPLItem
	var nonOperatingExpenseItems []EnhancedPLItem
	var taxExpenseItems []EnhancedPLItem

	// Process each account
	for _, account := range accounts {
		if !account.IsActive {
			continue
		}

		// Calculate period activity
		balance := epls.calculateAccountBalanceForPeriod(account.ID, startDate, endDate)
		
		// Skip accounts with no activity
		if balance == 0 {
			continue
		}

		item := EnhancedPLItem{
			AccountID: account.ID,
			Code:      account.Code,
			Name:      account.Name,
			Amount:    balance,
			Category:  account.Category,
		}

		// Categorize items based on account type and category
		switch account.Type {
		case models.AccountTypeRevenue:
			revenueItems = append(revenueItems, item)
		case models.AccountTypeExpense:
			if epls.isCOGSAccount(account) {
				cogsItems = append(cogsItems, item)
			} else if epls.isTaxExpenseCategory(account.Category) {
				taxExpenseItems = append(taxExpenseItems, item)
			} else if epls.isOperatingExpenseCategory(account.Category) {
				operatingExpenseItems = append(operatingExpenseItems, item)
			} else {
				nonOperatingExpenseItems = append(nonOperatingExpenseItems, item)
			}
		}
	}

	// Build revenue sections
	epls.buildRevenueSection(profitLoss, revenueItems)
	
	// Build COGS sections
	epls.buildCOGSSection(profitLoss, cogsItems)
	
	// Calculate Gross Profit
	profitLoss.GrossProfit = profitLoss.Revenue.TotalRevenue - profitLoss.CostOfGoodsSold.TotalCOGS
	if profitLoss.Revenue.TotalRevenue != 0 {
		profitLoss.GrossProfitMargin = (profitLoss.GrossProfit / profitLoss.Revenue.TotalRevenue) * 100
	}

	// Build operating expenses sections
	epls.buildOperatingExpenseSection(profitLoss, operatingExpenseItems)
	
	// Calculate Operating Income (EBIT)
	profitLoss.OperatingIncome = profitLoss.GrossProfit - profitLoss.OperatingExpenses.TotalOpex
	if profitLoss.Revenue.TotalRevenue != 0 {
		profitLoss.OperatingMargin = (profitLoss.OperatingIncome / profitLoss.Revenue.TotalRevenue) * 100
	}

	// Calculate EBITDA (add back depreciation)
	depreciationAmount := epls.getDepreciationAmount(operatingExpenseItems)
	profitLoss.EBITDA = profitLoss.OperatingIncome + depreciationAmount
	if profitLoss.Revenue.TotalRevenue != 0 {
		profitLoss.EBITDAMargin = (profitLoss.EBITDA / profitLoss.Revenue.TotalRevenue) * 100
	}

	// Build other income/expense sections
	epls.buildOtherIncomeExpenseSection(profitLoss, nonOperatingIncomeItems, nonOperatingExpenseItems)
	
	// Calculate Income Before Tax
	profitLoss.IncomeBeforeTax = profitLoss.OperatingIncome + profitLoss.OtherIncomeExpense.NetOtherIncome

	// Calculate Tax
	profitLoss.TaxExpense = epls.sumItems(taxExpenseItems)
	if profitLoss.IncomeBeforeTax != 0 {
		profitLoss.TaxRate = (profitLoss.TaxExpense / profitLoss.IncomeBeforeTax) * 100
	}

	// Calculate Net Income
	profitLoss.NetIncome = profitLoss.IncomeBeforeTax - profitLoss.TaxExpense
	if profitLoss.Revenue.TotalRevenue != 0 {
		profitLoss.NetIncomeMargin = (profitLoss.NetIncome / profitLoss.Revenue.TotalRevenue) * 100
	}

	// Calculate EPS (if shares outstanding available)
	profitLoss.SharesOutstanding = epls.getSharesOutstanding()
	if profitLoss.SharesOutstanding > 0 {
		profitLoss.EarningsPerShare = profitLoss.NetIncome / profitLoss.SharesOutstanding
	}

	// Calculate percentages for all items
	epls.calculatePercentages(profitLoss)

	return profitLoss, nil
}

// Helper methods for categorization
func (epls *EnhancedProfitLossService) isCOGSCategory(category string) bool {
	cogsCategories := []string{
		CategoryDirectMaterials,
		CategoryDirectLabor,
		CategoryManufacturingOH,
		CategoryFreightIn,
		CategoryCostOfGoodsSold,
		"COST_OF_GOODS_SOLD", // Legacy support
	}
	return containsString(cogsCategories, category)
}

// Check if account should be treated as COGS based on name pattern
func (epls *EnhancedProfitLossService) isCOGSAccount(account models.Account) bool {
	// Check category first
	if epls.isCOGSCategory(account.Category) {
		return true
	}
	
	// Check by account code and name patterns for Indonesian COA
	if account.Code == "5101" || 
	   account.Name == "Harga Pokok Penjualan" ||
	   account.Name == "Cost of Goods Sold" {
		return true
	}
	
	// Additional Indonesian COGS patterns
	if strings.Contains(strings.ToLower(account.Name), "harga pokok") ||
	   strings.Contains(strings.ToLower(account.Name), "pokok penjualan") ||
	   strings.Contains(strings.ToLower(account.Name), "cost of goods") {
		return true
	}
	
	return false
}

func (epls *EnhancedProfitLossService) isOperatingExpenseCategory(category string) bool {
	opexCategories := []string{
		CategoryAdminExpense,
		CategorySellingExpense,
		CategoryGeneralExpense,
		CategoryDepreciationExp,
		CategoryOperatingExpense,
		"ADMINISTRATIVE_EXPENSE",  // Legacy support
		"SELLING_EXPENSE",
		"MARKETING_EXPENSE",
		"GENERAL_EXPENSE",
		"DEPRECIATION_EXPENSE",
		"OPERATING_EXPENSE",
	}
	return containsString(opexCategories, category)
}

func (epls *EnhancedProfitLossService) isTaxExpenseCategory(category string) bool {
	taxCategories := []string{
		CategoryTaxExpense,
		"TAX_EXPENSE",
		"INCOME_TAX_EXPENSE",
	}
	return containsString(taxCategories, category)
}

// Build revenue sections
func (epls *EnhancedProfitLossService) buildRevenueSection(pl *EnhancedProfitLossData, items []EnhancedPLItem) {
	var salesItems, serviceItems, otherItems []EnhancedPLItem

	for _, item := range items {
		switch item.Category {
		case CategorySalesRevenue, "OPERATING_REVENUE": // Handle existing category
			salesItems = append(salesItems, item)
		case CategoryServiceRevenue: // "SERVICE_REVENUE"
			serviceItems = append(serviceItems, item)
		case CategoryOtherOperatingRev, CategoryNonOperatingRevenue, "OTHER_INCOME": // Handle existing categories
			otherItems = append(otherItems, item)
		default:
			// Default operating revenue items go to sales
			if item.Code == "4101" || item.Name == "Pendapatan Penjualan" {
				salesItems = append(salesItems, item)
			} else {
				otherItems = append(otherItems, item)
			}
		}
	}

	pl.Revenue.SalesRevenue = EnhancedPLSection{
		Name:     "Sales Revenue",
		Items:    salesItems,
		Subtotal: epls.sumItems(salesItems),
	}

	pl.Revenue.ServiceRevenue = EnhancedPLSection{
		Name:     "Service Revenue",
		Items:    serviceItems,
		Subtotal: epls.sumItems(serviceItems),
	}

	pl.Revenue.OtherRevenue = EnhancedPLSection{
		Name:     "Other Revenue",
		Items:    otherItems,
		Subtotal: epls.sumItems(otherItems),
	}

	pl.Revenue.TotalRevenue = pl.Revenue.SalesRevenue.Subtotal + 
							  pl.Revenue.ServiceRevenue.Subtotal + 
							  pl.Revenue.OtherRevenue.Subtotal
}

// Build COGS sections
func (epls *EnhancedProfitLossService) buildCOGSSection(pl *EnhancedProfitLossData, items []EnhancedPLItem) {
	var directMaterialItems, directLaborItems, manufacturingOHItems, otherCOGSItems []EnhancedPLItem

	for _, item := range items {
		switch item.Category {
		case CategoryDirectMaterials: // "DIRECT_MATERIALS"
			directMaterialItems = append(directMaterialItems, item)
		case CategoryDirectLabor: // "DIRECT_LABOR"
			directLaborItems = append(directLaborItems, item)
		case CategoryManufacturingOH: // "MANUFACTURING_OVERHEAD"
			manufacturingOHItems = append(manufacturingOHItems, item)
		default:
			otherCOGSItems = append(otherCOGSItems, item)
		}
	}

	pl.CostOfGoodsSold.DirectMaterials = EnhancedPLSection{
		Name:     "Direct Materials",
		Items:    directMaterialItems,
		Subtotal: epls.sumItems(directMaterialItems),
	}

	pl.CostOfGoodsSold.DirectLabor = EnhancedPLSection{
		Name:     "Direct Labor",
		Items:    directLaborItems,
		Subtotal: epls.sumItems(directLaborItems),
	}

	pl.CostOfGoodsSold.ManufacturingOH = EnhancedPLSection{
		Name:     "Manufacturing Overhead",
		Items:    manufacturingOHItems,
		Subtotal: epls.sumItems(manufacturingOHItems),
	}

	pl.CostOfGoodsSold.OtherCOGS = EnhancedPLSection{
		Name:     "Other Cost of Goods Sold",
		Items:    otherCOGSItems,
		Subtotal: epls.sumItems(otherCOGSItems),
	}

	pl.CostOfGoodsSold.TotalCOGS = pl.CostOfGoodsSold.DirectMaterials.Subtotal +
								   pl.CostOfGoodsSold.DirectLabor.Subtotal +
								   pl.CostOfGoodsSold.ManufacturingOH.Subtotal +
								   pl.CostOfGoodsSold.OtherCOGS.Subtotal
}

// Build operating expense sections
func (epls *EnhancedProfitLossService) buildOperatingExpenseSection(pl *EnhancedProfitLossData, items []EnhancedPLItem) {
	var adminItems, sellingMarketingItems, generalItems, depreciationItems []EnhancedPLItem

	for _, item := range items {
		switch item.Category {
		case CategoryAdminExpense: // "ADMINISTRATIVE_EXPENSE"
			adminItems = append(adminItems, item)
		case CategorySellingExpense, "SELLING_EXPENSE", "MARKETING_EXPENSE": // Legacy support for different naming
			sellingMarketingItems = append(sellingMarketingItems, item)
		case CategoryDepreciationExp, "DEPRECIATION_EXPENSE", "AMORTIZATION_EXPENSE": // Legacy support for different naming
			depreciationItems = append(depreciationItems, item)
		default:
			generalItems = append(generalItems, item)
		}
	}

	pl.OperatingExpenses.Administrative = EnhancedPLSection{
		Name:     "Administrative Expenses",
		Items:    adminItems,
		Subtotal: epls.sumItems(adminItems),
	}

	pl.OperatingExpenses.SellingMarketing = EnhancedPLSection{
		Name:     "Selling & Marketing Expenses",
		Items:    sellingMarketingItems,
		Subtotal: epls.sumItems(sellingMarketingItems),
	}

	pl.OperatingExpenses.General = EnhancedPLSection{
		Name:     "General Expenses",
		Items:    generalItems,
		Subtotal: epls.sumItems(generalItems),
	}

	pl.OperatingExpenses.Depreciation = EnhancedPLSection{
		Name:     "Depreciation & Amortization",
		Items:    depreciationItems,
		Subtotal: epls.sumItems(depreciationItems),
	}

	pl.OperatingExpenses.TotalOpex = pl.OperatingExpenses.Administrative.Subtotal +
									 pl.OperatingExpenses.SellingMarketing.Subtotal +
									 pl.OperatingExpenses.General.Subtotal +
									 pl.OperatingExpenses.Depreciation.Subtotal
}

// Build other income/expense sections
func (epls *EnhancedProfitLossService) buildOtherIncomeExpenseSection(pl *EnhancedProfitLossData, incomeItems, expenseItems []EnhancedPLItem) {
	var interestIncomeItems, otherIncomeItems []EnhancedPLItem
	var interestExpenseItems, otherExpenseItems []EnhancedPLItem

	for _, item := range incomeItems {
		if item.Category == CategoryInterestIncome || item.Category == "INTEREST_INCOME" {
			interestIncomeItems = append(interestIncomeItems, item)
		} else {
			otherIncomeItems = append(otherIncomeItems, item)
		}
	}

	for _, item := range expenseItems {
		if item.Category == CategoryInterestExpense || item.Category == "INTEREST_EXPENSE" {
			interestExpenseItems = append(interestExpenseItems, item)
		} else {
			otherExpenseItems = append(otherExpenseItems, item)
		}
	}

	pl.OtherIncomeExpense.InterestIncome = EnhancedPLSection{
		Name:     "Interest Income",
		Items:    interestIncomeItems,
		Subtotal: epls.sumItems(interestIncomeItems),
	}

	pl.OtherIncomeExpense.InterestExpense = EnhancedPLSection{
		Name:     "Interest Expense",
		Items:    interestExpenseItems,
		Subtotal: epls.sumItems(interestExpenseItems),
	}

	pl.OtherIncomeExpense.OtherIncome = EnhancedPLSection{
		Name:     "Other Non-Operating Income",
		Items:    otherIncomeItems,
		Subtotal: epls.sumItems(otherIncomeItems),
	}

	pl.OtherIncomeExpense.OtherExpense = EnhancedPLSection{
		Name:     "Other Non-Operating Expenses",
		Items:    otherExpenseItems,
		Subtotal: epls.sumItems(otherExpenseItems),
	}

	pl.OtherIncomeExpense.NetOtherIncome = pl.OtherIncomeExpense.InterestIncome.Subtotal +
										   pl.OtherIncomeExpense.OtherIncome.Subtotal -
										   pl.OtherIncomeExpense.InterestExpense.Subtotal -
										   pl.OtherIncomeExpense.OtherExpense.Subtotal
}

// Helper methods
func (epls *EnhancedProfitLossService) calculateAccountBalanceForPeriod(accountID uint, startDate, endDate time.Time) float64 {
	// Get account to determine normal balance type
	var account models.Account
	if err := epls.db.First(&account, accountID).Error; err != nil {
		return 0
	}

	// First try to get period activity from journal lines
	var totalDebits, totalCredits float64
	epls.db.Table("journal_lines").
		Joins("JOIN journal_entries ON journal_lines.journal_entry_id = journal_entries.id").
		Where("journal_lines.account_id = ? AND journal_entries.entry_date BETWEEN ? AND ? AND journal_entries.status = ?", 
			accountID, startDate, endDate, models.JournalStatusPosted).
		Select("COALESCE(SUM(journal_lines.debit_amount), 0) as total_debits, COALESCE(SUM(journal_lines.credit_amount), 0) as total_credits").
		Row().Scan(&totalDebits, &totalCredits)

	// Calculate period activity from journal entries
	var periodActivity float64
	switch account.Type {
	case models.AccountTypeRevenue:
		// Revenue: Credit is positive
		periodActivity = totalCredits - totalDebits
	case models.AccountTypeExpense:
		// Expenses: Debit is positive  
		periodActivity = totalDebits - totalCredits
	default:
		return 0
	}

	// IMPORTANT: After synchronization, we should ONLY rely on journal entries
	// If there's no journal activity, check if we need to use cumulative balance
	// for period reporting (when period spans from beginning of time)
	if periodActivity == 0 && account.Balance != 0 {
		// Check if period starts from a very early date (indicating full period report)
		if startDate.Year() < 2020 || startDate.IsZero() {
			// For full period reports, use account balance
			switch account.Type {
			case models.AccountTypeRevenue:
				return account.Balance
			case models.AccountTypeExpense:
				return account.Balance
			}
		}
		// For specific period reports, no journal activity means no activity
		return 0
	}

	return periodActivity
}

func (epls *EnhancedProfitLossService) sumItems(items []EnhancedPLItem) float64 {
	var total float64
	for _, item := range items {
		total += item.Amount
	}
	return total
}

func (epls *EnhancedProfitLossService) getDepreciationAmount(items []EnhancedPLItem) float64 {
	var total float64
	for _, item := range items {
		if item.Category == CategoryDepreciationExp || 
		   item.Category == "DEPRECIATION_EXPENSE" || 
		   item.Category == "AMORTIZATION_EXPENSE" {
			total += item.Amount
		}
	}
	return total
}

func (epls *EnhancedProfitLossService) getSharesOutstanding() float64 {
	// Try to get shares outstanding from company profile first
	var profile models.CompanyProfile
	if err := epls.db.First(&profile).Error; err == nil && profile.SharesOutstanding > 0 {
		return profile.SharesOutstanding
	}
	
	// Try to calculate from share capital accounts - suppress error logging
	var shareCapitalAccount models.Account
	err := epls.db.Where("type = ? AND (category LIKE ? OR category LIKE ? OR name ILIKE ?) AND is_active = ?", 
		models.AccountTypeEquity, "%SHARE_CAPITAL%", "%MODAL_SAHAM%", "%modal%", true).First(&shareCapitalAccount).Error
	
	if err == nil && shareCapitalAccount.Balance > 0 {
		// Assuming par value of 1000 per share (can be configured in company profile)
		parValue := float64(1000)
		if profile.ParValuePerShare > 0 {
			parValue = profile.ParValuePerShare
		}
		return shareCapitalAccount.Balance / parValue
	}
	
	// If no share capital account found, try to find from other equity accounts
	var totalEquity float64
	epls.db.Model(&models.Account{}).Where("type = ? AND is_active = ?", models.AccountTypeEquity, true).Select("COALESCE(SUM(balance), 0)").Scan(&totalEquity)
	
	if totalEquity > 0 {
		// Use default par value calculation
		parValue := float64(1000) // Default 1000 per share
		return totalEquity / parValue
	}
	
	// Return default 1 share if no equity data available to avoid division by zero
	return 1
}

func (epls *EnhancedProfitLossService) calculatePercentages(pl *EnhancedProfitLossData) {
	totalRevenue := pl.Revenue.TotalRevenue
	if totalRevenue == 0 {
		return
	}

	// Calculate percentages for all sections
	epls.calculateSectionPercentages(&pl.Revenue.SalesRevenue, totalRevenue)
	epls.calculateSectionPercentages(&pl.Revenue.ServiceRevenue, totalRevenue)
	epls.calculateSectionPercentages(&pl.Revenue.OtherRevenue, totalRevenue)
	
	epls.calculateSectionPercentages(&pl.CostOfGoodsSold.DirectMaterials, totalRevenue)
	epls.calculateSectionPercentages(&pl.CostOfGoodsSold.DirectLabor, totalRevenue)
	epls.calculateSectionPercentages(&pl.CostOfGoodsSold.ManufacturingOH, totalRevenue)
	epls.calculateSectionPercentages(&pl.CostOfGoodsSold.OtherCOGS, totalRevenue)
	
	epls.calculateSectionPercentages(&pl.OperatingExpenses.Administrative, totalRevenue)
	epls.calculateSectionPercentages(&pl.OperatingExpenses.SellingMarketing, totalRevenue)
	epls.calculateSectionPercentages(&pl.OperatingExpenses.General, totalRevenue)
	epls.calculateSectionPercentages(&pl.OperatingExpenses.Depreciation, totalRevenue)
	
	epls.calculateSectionPercentages(&pl.OtherIncomeExpense.InterestIncome, totalRevenue)
	epls.calculateSectionPercentages(&pl.OtherIncomeExpense.InterestExpense, totalRevenue)
	epls.calculateSectionPercentages(&pl.OtherIncomeExpense.OtherIncome, totalRevenue)
	epls.calculateSectionPercentages(&pl.OtherIncomeExpense.OtherExpense, totalRevenue)
}

func (epls *EnhancedProfitLossService) calculateSectionPercentages(section *EnhancedPLSection, totalRevenue float64) {
	for i := range section.Items {
		if totalRevenue != 0 {
			section.Items[i].Percentage = (section.Items[i].Amount / totalRevenue) * 100
		}
	}
}

// getCompanyInfo returns basic company information
func (epls *EnhancedProfitLossService) getCompanyInfo() CompanyInfo {
	// Fetch company profile from database
	var profile models.CompanyProfile
	err := epls.db.First(&profile).Error
	if err != nil {
		// If no company profile exists, create a default one
		profile = models.CompanyProfile{
			Name:       "Your Company Name",
			Address:    "Company Address",
			City:       "City",
			State:      "State",
			Country:    "Indonesia",
			PostalCode: "12345",
			Phone:      "+62-21-1234567",
			Email:      "contact@company.com",
			Website:    "www.company.com",
			Currency:   "IDR",
			TaxNumber:  "",
			FiscalYearStart: "01-01",
		}
		// Save the default profile to database
		epls.db.Create(&profile)
	}
	
	return CompanyInfo{
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

// Utility function
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
