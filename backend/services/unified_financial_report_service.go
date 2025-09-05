package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/utils"

	"gorm.io/gorm"
)

// UnifiedFinancialReportService provides comprehensive financial reporting
// integrated with the accounting system's business flow
type UnifiedFinancialReportService struct {
	db               *gorm.DB
	accountRepo      repositories.AccountRepository
	journalRepo      repositories.JournalEntryRepository
	salesRepo        *repositories.SalesRepository
	purchaseRepo     *repositories.PurchaseRepository
	paymentRepo      *repositories.PaymentRepository
	cashBankRepo     *repositories.CashBankRepository
	contactRepo      repositories.ContactRepository
	productRepo      *repositories.ProductRepository
	companyProfile   *models.CompanyProfile
}

// NewUnifiedFinancialReportService creates a new unified financial report service
func NewUnifiedFinancialReportService(
	db *gorm.DB,
	accountRepo repositories.AccountRepository,
	journalRepo repositories.JournalEntryRepository,
	salesRepo *repositories.SalesRepository,
	purchaseRepo *repositories.PurchaseRepository,
	paymentRepo *repositories.PaymentRepository,
	cashBankRepo *repositories.CashBankRepository,
	contactRepo repositories.ContactRepository,
	productRepo *repositories.ProductRepository,
) *UnifiedFinancialReportService {
	return &UnifiedFinancialReportService{
		db:             db,
		accountRepo:    accountRepo,
		journalRepo:    journalRepo,
		salesRepo:      salesRepo,
		purchaseRepo:   purchaseRepo,
		paymentRepo:    paymentRepo,
		cashBankRepo:   cashBankRepo,
		contactRepo:    contactRepo,
		productRepo:    productRepo,
	}
}

// ========================= PROFIT & LOSS STATEMENT =========================

// GenerateComprehensiveProfitLoss generates a comprehensive P&L statement
func (s *UnifiedFinancialReportService) GenerateComprehensiveProfitLoss(ctx context.Context, startDate, endDate time.Time, comparative bool) (*models.ProfitLossStatement, error) {
	// Get company profile
	companyInfo, err := s.getCompanyInfo(ctx)
	if err != nil {
		return nil, err
	}

	// Get revenue accounts and calculate balances for the period
	revenueAccounts, err := s.getAccountBalancesForPeriod(ctx, models.AccountTypeRevenue, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get expense accounts and separate COGS from operating expenses
	expenseAccounts, err := s.getAccountBalancesForPeriod(ctx, models.AccountTypeExpense, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Separate COGS and Operating Expenses
	var cogsAccounts, operatingExpenses []models.AccountLineItem
	for _, expense := range expenseAccounts {
		if s.isCOGSAccount(expense.Category) {
			cogsAccounts = append(cogsAccounts, expense)
		} else {
			operatingExpenses = append(operatingExpenses, expense)
		}
	}

	// Calculate totals
	totalRevenue := s.calculateTotalBalance(revenueAccounts)
	totalCOGS := s.calculateTotalBalance(cogsAccounts)
	totalOperatingExpenses := s.calculateTotalBalance(operatingExpenses)
	
	grossProfit := totalRevenue - totalCOGS
	operatingIncome := grossProfit - totalOperatingExpenses
	netIncome := operatingIncome // Simplified, can add other income/expenses


	// Build P&L statement
	pnl := &models.ProfitLossStatement{
		ReportHeader: models.ReportHeader{
			ReportType:    models.ReportTypeProfitLoss,
			CompanyName:   companyInfo.Name,
			ReportTitle:   "Profit & Loss Statement",
			StartDate:     startDate,
			EndDate:       endDate,
			GeneratedAt:   time.Now(),
			GeneratedBy:   "System",
			Currency:      companyInfo.Currency,
			IsComparative: comparative,
		},
		Revenue:       revenueAccounts,
		TotalRevenue:  totalRevenue,
		COGS:          cogsAccounts,
		TotalCOGS:     totalCOGS,
		GrossProfit:   grossProfit,
		Expenses:      operatingExpenses,
		TotalExpenses: totalOperatingExpenses,
		NetIncome:     netIncome,
	}

	// Add comparative analysis if requested
	if comparative {
		// Calculate previous period (same duration)
		duration := endDate.Sub(startDate)
		prevEndDate := startDate.Add(-time.Hour * 24)
		prevStartDate := prevEndDate.Add(-duration)

		prevPnL, err := s.GenerateComprehensiveProfitLoss(ctx, prevStartDate, prevEndDate, false)
		if err == nil {
			pnl.Comparative = &models.ProfitLossComparative{
				PreviousPeriod: *prevPnL,
				Variance: models.ProfitLossVariance{
					RevenueVariance:     totalRevenue - prevPnL.TotalRevenue,
					COGSVariance:        totalCOGS - prevPnL.TotalCOGS,
					GrossProfitVariance: grossProfit - prevPnL.GrossProfit,
					ExpenseVariance:     totalOperatingExpenses - prevPnL.TotalExpenses,
					NetIncomeVariance:   netIncome - prevPnL.NetIncome,
				},
			}
		}
	}

	return pnl, nil
}

// ========================= BALANCE SHEET =========================

// GenerateComprehensiveBalanceSheet generates a comprehensive balance sheet
func (s *UnifiedFinancialReportService) GenerateComprehensiveBalanceSheet(ctx context.Context, asOfDate time.Time, comparative bool) (*models.BalanceSheet, error) {
	companyInfo, err := s.getCompanyInfo(ctx)
	if err != nil {
		return nil, err
	}

	// Get account balances as of the specified date
	assetAccounts, err := s.getAccountBalancesAsOfDate(ctx, models.AccountTypeAsset, asOfDate)
	if err != nil {
		return nil, err
	}

	liabilityAccounts, err := s.getAccountBalancesAsOfDate(ctx, models.AccountTypeLiability, asOfDate)
	if err != nil {
		return nil, err
	}

	equityAccounts, err := s.getAccountBalancesAsOfDate(ctx, models.AccountTypeEquity, asOfDate)
	if err != nil {
		return nil, err
	}

	// Group accounts by categories
	assetSection := s.groupAccountsByCategory(assetAccounts)
	liabilitySection := s.groupAccountsByCategory(liabilityAccounts)
	equitySection := s.groupAccountsByCategory(equityAccounts)

	// Calculate totals
	totalAssets := s.calculateTotalBalance(assetAccounts)
	totalLiabilities := s.calculateTotalBalance(liabilityAccounts)
	totalEquity := s.calculateTotalBalance(equityAccounts)

	// Check if balanced (Assets = Liabilities + Equity)
	isBalanced := math.Abs(totalAssets-(totalLiabilities+totalEquity)) < 0.01

	balanceSheet := &models.BalanceSheet{
		ReportHeader: models.ReportHeader{
			ReportType:    models.ReportTypeBalanceSheet,
			CompanyName:   companyInfo.Name,
			ReportTitle:   "Balance Sheet",
			StartDate:     asOfDate,
			EndDate:       asOfDate,
			GeneratedAt:   time.Now(),
			GeneratedBy:   "System",
			Currency:      companyInfo.Currency,
			IsComparative: comparative,
		},
		Assets:           assetSection,
		Liabilities:      liabilitySection,
		Equity:           equitySection,
		TotalAssets:      totalAssets,
		TotalLiabilities: totalLiabilities,
		TotalEquity:      totalEquity,
		IsBalanced:       isBalanced,
	}

	// Add comparative analysis if requested
	if comparative {
		prevYearDate := asOfDate.AddDate(-1, 0, 0)
		prevBalanceSheet, err := s.GenerateComprehensiveBalanceSheet(ctx, prevYearDate, false)
		if err == nil {
			balanceSheet.Comparative = &models.BalanceSheetComparative{
				PreviousPeriod: *prevBalanceSheet,
				Variance: models.BalanceSheetVariance{
					AssetsVariance:      totalAssets - prevBalanceSheet.TotalAssets,
					LiabilitiesVariance: totalLiabilities - prevBalanceSheet.TotalLiabilities,
					EquityVariance:      totalEquity - prevBalanceSheet.TotalEquity,
				},
			}
		}
	}

	return balanceSheet, nil
}

// ========================= CASH FLOW STATEMENT =========================

// GenerateComprehensiveCashFlow generates a comprehensive cash flow statement
func (s *UnifiedFinancialReportService) GenerateComprehensiveCashFlow(ctx context.Context, startDate, endDate time.Time) (*models.CashFlowStatement, error) {
	companyInfo, err := s.getCompanyInfo(ctx)
	if err != nil {
		return nil, err
	}

	// Get beginning and ending cash balances
	beginningCash, err := s.getTotalCashBalanceAsOfDate(ctx, startDate.Add(-time.Hour*24))
	if err != nil {
		return nil, err
	}

	endingCash, err := s.getTotalCashBalanceAsOfDate(ctx, endDate)
	if err != nil {
		return nil, err
	}

	// Calculate cash flow activities
	operatingCF, err := s.calculateOperatingCashFlow(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	investingCF, err := s.calculateInvestingCashFlow(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	financingCF, err := s.calculateFinancingCashFlow(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	netCashFlow := operatingCF.Total + investingCF.Total + financingCF.Total

	return &models.CashFlowStatement{
		ReportHeader: models.ReportHeader{
			ReportType:    models.ReportTypeCashFlow,
			CompanyName:   companyInfo.Name,
			ReportTitle:   "Cash Flow Statement",
			StartDate:     startDate,
			EndDate:       endDate,
			GeneratedAt:   time.Now(),
			GeneratedBy:   "System",
			Currency:      companyInfo.Currency,
			IsComparative: false,
		},
		OperatingActivities: operatingCF,
		InvestingActivities: investingCF,
		FinancingActivities: financingCF,
		NetCashFlow:         netCashFlow,
		BeginningCash:       beginningCash,
		EndingCash:          endingCash,
	}, nil
}

// ========================= TRIAL BALANCE =========================

// GenerateComprehensiveTrialBalance generates a comprehensive trial balance
func (s *UnifiedFinancialReportService) GenerateComprehensiveTrialBalance(ctx context.Context, asOfDate time.Time, showZero bool) (*models.TrialBalance, error) {
	companyInfo, err := s.getCompanyInfo(ctx)
	if err != nil {
		return nil, err
	}

	var accounts []models.Account
	query := s.db.WithContext(ctx).Where("is_active = ?", true).Order("code ASC")
	
	if !showZero {
		query = query.Where("balance != ?", 0)
	}

	err = query.Find(&accounts).Error
	if err != nil {
		return nil, utils.NewInternalError("Failed to get accounts for trial balance", err)
	}

	var trialBalanceItems []models.TrialBalanceItem
	var totalDebits, totalCredits float64

	for _, account := range accounts {
		// Get actual balance as of date from journal entries
		balance, err := s.getAccountBalanceAsOfDate(ctx, account.ID, asOfDate)
		if err != nil {
			balance = account.Balance // Fallback to current balance
		}

		var debitBalance, creditBalance float64
		normalBalance := account.GetNormalBalance()

		// Show balance in appropriate column based on normal balance and actual balance
		if (balance >= 0 && normalBalance == models.NormalBalanceDebit) || 
		   (balance < 0 && normalBalance == models.NormalBalanceCredit) {
			debitBalance = math.Abs(balance)
			totalDebits += debitBalance
		} else {
			creditBalance = math.Abs(balance)
			totalCredits += creditBalance
		}

		trialBalanceItems = append(trialBalanceItems, models.TrialBalanceItem{
			AccountID:     account.ID,
			AccountCode:   account.Code,
			AccountName:   account.Name,
			AccountType:   account.Type,
			DebitBalance:  debitBalance,
			CreditBalance: creditBalance,
		})
	}

	return &models.TrialBalance{
		ReportHeader: models.ReportHeader{
			ReportType:  models.ReportTypeTrialBalance,
			CompanyName: companyInfo.Name,
			ReportTitle: "Trial Balance",
			StartDate:   asOfDate,
			EndDate:     asOfDate,
			GeneratedAt: time.Now(),
			GeneratedBy: "System",
			Currency:    companyInfo.Currency,
		},
		Accounts:     trialBalanceItems,
		TotalDebits:  totalDebits,
		TotalCredits: totalCredits,
		IsBalanced:   math.Abs(totalDebits-totalCredits) < 0.01,
	}, nil
}

// ========================= GENERAL LEDGER =========================

// GenerateComprehensiveGeneralLedger generates a comprehensive general ledger for an account
func (s *UnifiedFinancialReportService) GenerateComprehensiveGeneralLedger(ctx context.Context, accountID uint, startDate, endDate time.Time) (*models.GeneralLedger, error) {
	companyInfo, err := s.getCompanyInfo(ctx)
	if err != nil {
		return nil, err
	}

	// Get account details
	account, err := s.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Get beginning balance
	beginningBalance, err := s.getAccountBalanceAsOfDate(ctx, accountID, startDate.Add(-time.Hour*24))
	if err != nil {
		beginningBalance = 0
	}

	// Get all journal entries for this account in the period
	var journalLines []models.JournalLine
	err = s.db.WithContext(ctx).
		Joins("JOIN journal_entries je ON journal_lines.journal_entry_id = je.id").
		Where("journal_lines.account_id = ? AND je.entry_date BETWEEN ? AND ? AND je.status = ?", 
			accountID, startDate, endDate, models.JournalStatusPosted).
		Order("je.entry_date ASC, journal_lines.line_number ASC").
		Preload("JournalEntry").
		Find(&journalLines).Error

	if err != nil {
		return nil, utils.NewInternalError("Failed to get journal entries", err)
	}

	// Build transaction list with running balance
	var transactions []models.GeneralLedgerEntry
	var totalDebits, totalCredits float64
	runningBalance := beginningBalance

	for _, line := range journalLines {
		// Calculate running balance based on account's normal balance
		normalBalance := account.GetNormalBalance()
		if normalBalance == models.NormalBalanceDebit {
			runningBalance += line.DebitAmount - line.CreditAmount
		} else {
			runningBalance += line.CreditAmount - line.DebitAmount
		}

		totalDebits += line.DebitAmount
		totalCredits += line.CreditAmount

		transactions = append(transactions, models.GeneralLedgerEntry{
			Date:         line.JournalEntry.EntryDate,
			JournalCode:  line.JournalEntry.Code,
			Description:  line.Description,
			Reference:    line.JournalEntry.Reference,
			DebitAmount:  line.DebitAmount,
			CreditAmount: line.CreditAmount,
			Balance:      runningBalance,
		})
	}

	return &models.GeneralLedger{
		ReportHeader: models.ReportHeader{
			ReportType:  models.ReportTypeGeneralLedger,
			CompanyName: companyInfo.Name,
			ReportTitle: fmt.Sprintf("General Ledger - %s (%s)", account.Name, account.Code),
			StartDate:   startDate,
			EndDate:     endDate,
			GeneratedAt: time.Now(),
			GeneratedBy: "System",
			Currency:    companyInfo.Currency,
		},
		Account:          *account,
		Transactions:     transactions,
		BeginningBalance: beginningBalance,
		EndingBalance:    runningBalance,
		TotalDebits:      totalDebits,
		TotalCredits:     totalCredits,
	}, nil
}

// ========================= SALES SUMMARY REPORT =========================

// GenerateComprehensiveSalesSummary generates a comprehensive sales summary report
func (s *UnifiedFinancialReportService) GenerateComprehensiveSalesSummary(ctx context.Context, startDate, endDate time.Time) (*SalesSummaryReport, error) {
	companyInfo, err := s.getCompanyInfo(ctx)
	if err != nil {
		return nil, err
	}

	// Get all sales in the period
	filter := &models.SalesFilter{
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		Status:    "CONFIRMED,COMPLETED,INVOICED,PAID", // Only confirmed sales
		Limit:     10000,
	}

	// Use FindWithFilter as a temporary workaround
	filter_struct := models.SalesFilter{
		StartDate: filter.StartDate,
		EndDate:   filter.EndDate,
		Status:    filter.Status,
		Limit:     filter.Limit,
		Page:      1,
	}
	
	sales, _, err := s.salesRepo.FindWithFilter(filter_struct)
	if err != nil {
		return nil, err
	}

	// Calculate summary metrics
	var totalRevenue, totalPaidAmount, totalOutstanding float64
	var totalTransactions int64
	customerMap := make(map[uint]float64)
	productMap := make(map[uint]float64)
	statusMap := make(map[string]int64)

	for _, sale := range sales {
		totalRevenue += sale.TotalAmount
		totalPaidAmount += sale.PaidAmount
		totalOutstanding += sale.OutstandingAmount
		totalTransactions++

		// Track by customer
		customerMap[sale.CustomerID] += sale.TotalAmount

		// Track by status
		statusMap[sale.Status]++

		// Track by product
		for _, item := range sale.SaleItems {
			productMap[item.ProductID] += item.LineTotal
		}
	}

	avgOrderValue := s.safeDiv(totalRevenue, float64(totalTransactions))

	// Get top customers and products
	topCustomers := s.getTopCustomers(ctx, customerMap, 10)
	topProducts := s.getTopProducts(ctx, productMap, 10)

	// Calculate growth compared to previous period
	growthAnalysis, err := s.calculateSalesGrowth(ctx, startDate, endDate)
	if err != nil {
		growthAnalysis = &SalesGrowthAnalysis{} // Empty if can't calculate
	}

	return &SalesSummaryReport{
		ReportHeader: models.ReportHeader{
			ReportType:  "SALES_SUMMARY",
			CompanyName: companyInfo.Name,
			ReportTitle: "Sales Summary Report",
			StartDate:   startDate,
			EndDate:     endDate,
			GeneratedAt: time.Now(),
			GeneratedBy: "System",
			Currency:    companyInfo.Currency,
		},
		TotalRevenue:       totalRevenue,
		TotalTransactions:  totalTransactions,
		AverageOrderValue:  avgOrderValue,
		TotalPaidAmount:    totalPaidAmount,
		TotalOutstanding:   totalOutstanding,
		TopCustomers:       topCustomers,
		TopProducts:        topProducts,
		SalesByStatus:      s.buildStatusSummary(statusMap),
		GrowthAnalysis:     *growthAnalysis,
	}, nil
}

// ========================= VENDOR ANALYSIS REPORT =========================

// GenerateComprehensiveVendorAnalysis generates a comprehensive vendor analysis report
func (s *UnifiedFinancialReportService) GenerateComprehensiveVendorAnalysis(ctx context.Context, startDate, endDate time.Time) (*VendorAnalysisReport, error) {
	companyInfo, err := s.getCompanyInfo(ctx)
	if err != nil {
		return nil, err
	}

	// Get all purchases in the period
	purchases, err := s.purchaseRepo.FindAll()
	if err != nil {
		return nil, err
	}

	// Analyze vendor performance
	vendorMap := make(map[uint]*VendorPerformance)
	var totalPurchases, totalPaidAmount, totalOutstanding float64
	var totalTransactions int64

	for _, purchase := range purchases {
		if _, exists := vendorMap[purchase.VendorID]; !exists {
			vendorMap[purchase.VendorID] = &VendorPerformance{
				VendorID:   purchase.VendorID,
				VendorName: purchase.Vendor.Name,
			}
		}

		vendor := vendorMap[purchase.VendorID]
		vendor.TotalPurchases += purchase.TotalAmount
		vendor.TotalTransactions++
		vendor.TotalPaid += purchase.PaidAmount
		vendor.TotalOutstanding += purchase.OutstandingAmount

		// Calculate payment performance
		if purchase.OutstandingAmount == 0 {
			vendor.PaidOnTime++
		} else if purchase.DueDate.Before(time.Now()) {
			vendor.Overdue++
		}

		totalPurchases += purchase.TotalAmount
		totalPaidAmount += purchase.PaidAmount
		totalOutstanding += purchase.OutstandingAmount
		totalTransactions++
	}

	// Convert map to slice and sort by total purchases
	var vendorPerformances []VendorPerformance
	for _, vendor := range vendorMap {
		vendor.AveragePurchaseValue = s.safeDiv(vendor.TotalPurchases, float64(vendor.TotalTransactions))
		vendor.PaymentPerformanceScore = s.calculatePaymentScore(vendor.PaidOnTime, vendor.Overdue, vendor.TotalTransactions)
		vendorPerformances = append(vendorPerformances, *vendor)
	}

	// Sort by total purchases
	sort.Slice(vendorPerformances, func(i, j int) bool {
		return vendorPerformances[i].TotalPurchases > vendorPerformances[j].TotalPurchases
	})

	return &VendorAnalysisReport{
		ReportHeader: models.ReportHeader{
			ReportType:  "VENDOR_ANALYSIS",
			CompanyName: companyInfo.Name,
			ReportTitle: "Vendor Analysis Report",
			StartDate:   startDate,
			EndDate:     endDate,
			GeneratedAt: time.Now(),
			GeneratedBy: "System",
			Currency:    companyInfo.Currency,
		},
		TotalPurchases:        totalPurchases,
		TotalTransactions:     totalTransactions,
		TotalVendors:          int64(len(vendorPerformances)),
		AveragePurchaseValue:  s.safeDiv(totalPurchases, float64(totalTransactions)),
		VendorPerformances:    vendorPerformances,
		TotalPaidAmount:       totalPaidAmount,
		TotalOutstanding:      totalOutstanding,
		PaymentPerformance:    s.calculateOverallPaymentPerformance(vendorPerformances),
	}, nil
}

// ========================= HELPER METHODS =========================

// getCompanyInfo retrieves company information
func (s *UnifiedFinancialReportService) getCompanyInfo(ctx context.Context) (*UnifiedCompanyInfo, error) {
	if s.companyProfile != nil {
		return &UnifiedCompanyInfo{
			Name:       s.companyProfile.Name,
			Address:    s.companyProfile.Address,
			City:       s.companyProfile.City,
			State:      s.companyProfile.State,
			PostalCode: s.companyProfile.PostalCode,
			Phone:      s.companyProfile.Phone,
			Email:      s.companyProfile.Email,
			Website:    s.companyProfile.Website,
			TaxNumber:  s.companyProfile.TaxNumber,
			Currency:   "IDR", // Default currency
		}, nil
	}

	// Default company info if not configured
	return &UnifiedCompanyInfo{
		Name:     "PT. Sample Company",
		Currency: "IDR",
	}, nil
}

// getAccountBalancesForPeriod gets account balances for a specific period (P&L accounts)
func (s *UnifiedFinancialReportService) getAccountBalancesForPeriod(ctx context.Context, accountType string, startDate, endDate time.Time) ([]models.AccountLineItem, error) {
	var accounts []models.Account
	err := s.db.WithContext(ctx).
		Where("type = ? AND is_active = ? AND is_header = ?", accountType, true, false).
		Order("code ASC").
		Find(&accounts).Error

	if err != nil {
		return nil, utils.NewInternalError("Failed to get accounts", err)
	}

	var accountItems []models.AccountLineItem
	for _, account := range accounts {
		// Calculate balance for the period from journal entries
		balance := s.calculateAccountBalanceForPeriod(ctx, account.ID, startDate, endDate)
		
		// Only include accounts with non-zero balances for P&L
		if balance != 0 {
			accountItems = append(accountItems, models.AccountLineItem{
				AccountID:   account.ID,
				AccountCode: account.Code,
				AccountName: account.Name,
				AccountType: account.Type,
				Category:    account.Category,
				Balance:     math.Abs(balance), // Use absolute value for P&L
			})
		}
	}

	return accountItems, nil
}

// getAccountBalancesAsOfDate gets account balances as of a specific date (Balance Sheet accounts)
func (s *UnifiedFinancialReportService) getAccountBalancesAsOfDate(ctx context.Context, accountType string, asOfDate time.Time) ([]models.AccountLineItem, error) {
	var accounts []models.Account
	err := s.db.WithContext(ctx).
		Where("type = ? AND is_active = ? AND is_header = ?", accountType, true, false).
		Order("code ASC").
		Find(&accounts).Error

	if err != nil {
		return nil, utils.NewInternalError("Failed to get accounts", err)
	}

	var accountItems []models.AccountLineItem
	for _, account := range accounts {
		// Get balance as of specific date
		balance, err := s.getAccountBalanceAsOfDate(ctx, account.ID, asOfDate)
		if err != nil {
			balance = account.Balance // Fallback to current balance
		}

		// Include all accounts for balance sheet (even zero balances can be meaningful)
		accountItems = append(accountItems, models.AccountLineItem{
			AccountID:   account.ID,
			AccountCode: account.Code,
			AccountName: account.Name,
			AccountType: account.Type,
			Category:    account.Category,
			Balance:     balance,
		})
	}

	return accountItems, nil
}

// calculateAccountBalanceForPeriod calculates account balance for a specific period
func (s *UnifiedFinancialReportService) calculateAccountBalanceForPeriod(ctx context.Context, accountID uint, startDate, endDate time.Time) float64 {
	var total float64
	
	s.db.WithContext(ctx).
		Table("journal_lines jl").
		Joins("JOIN journal_entries je ON jl.journal_entry_id = je.id").
		Where("jl.account_id = ? AND je.entry_date BETWEEN ? AND ? AND je.status = ?", 
			accountID, startDate, endDate, models.JournalStatusPosted).
		Select("SUM(jl.debit_amount - jl.credit_amount)").
		Scan(&total)

	return total
}

// getAccountBalanceAsOfDate calculates account balance as of a specific date
func (s *UnifiedFinancialReportService) getAccountBalanceAsOfDate(ctx context.Context, accountID uint, asOfDate time.Time) (float64, error) {
	var total float64
	
	err := s.db.WithContext(ctx).
		Table("journal_lines jl").
		Joins("JOIN journal_entries je ON jl.journal_entry_id = je.id").
		Where("jl.account_id = ? AND je.entry_date <= ? AND je.status = ?", 
			accountID, asOfDate, models.JournalStatusPosted).
		Select("SUM(jl.debit_amount - jl.credit_amount)").
		Scan(&total).Error

	return total, err
}

// getTotalCashBalanceAsOfDate calculates total cash balance as of a specific date
func (s *UnifiedFinancialReportService) getTotalCashBalanceAsOfDate(ctx context.Context, asOfDate time.Time) (float64, error) {
	var total float64
	
	// Sum all cash and bank accounts
	err := s.db.WithContext(ctx).
		Table("accounts").
		Where("type = ? AND (category LIKE '%CASH%' OR code LIKE '110%' OR code LIKE '111%') AND is_active = ?", 
			models.AccountTypeAsset, true).
		Select("SUM(balance)").
		Scan(&total).Error

	return total, err
}

// calculateTotalBalance calculates total balance from account line items
func (s *UnifiedFinancialReportService) calculateTotalBalance(accounts []models.AccountLineItem) float64 {
	var total float64
	for _, account := range accounts {
		total += account.Balance
	}
	return total
}

// groupAccountsByCategory groups accounts by their categories
func (s *UnifiedFinancialReportService) groupAccountsByCategory(accounts []models.AccountLineItem) models.BalanceSheetSection {
	categoryMap := make(map[string][]models.AccountLineItem)
	
	for _, account := range accounts {
		category := account.Category
		if category == "" {
			category = "Other"
		}
		categoryMap[category] = append(categoryMap[category], account)
	}

	var categories []models.BalanceSheetCategory
	var totalBalance float64

	for categoryName, categoryAccounts := range categoryMap {
		categoryTotal := s.calculateTotalBalance(categoryAccounts)
		totalBalance += categoryTotal

		categories = append(categories, models.BalanceSheetCategory{
			Name:     categoryName,
			Accounts: categoryAccounts,
			Total:    categoryTotal,
		})
	}

	// Sort categories by total (descending)
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Total > categories[j].Total
	})

	return models.BalanceSheetSection{
		Categories: categories,
		Total:      totalBalance,
	}
}

// isCOGSAccount checks if an account is a Cost of Goods Sold account
func (s *UnifiedFinancialReportService) isCOGSAccount(category string) bool {
	cogsCategories := []string{
		models.CategoryCostOfGoodsSold,
		models.CategoryDirectMaterial,
		models.CategoryDirectLabor,
		models.CategoryManufacturingOverhead,
		models.CategoryFreightIn,
		models.CategoryPurchaseReturns,
	}
	
	for _, cogsCategory := range cogsCategories {
		if category == cogsCategory {
			return true
		}
	}
	return false
}

// safeDiv performs safe division avoiding division by zero
func (s *UnifiedFinancialReportService) safeDiv(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

// Additional data structures for enhanced reporting

type SalesSummaryReport struct {
	ReportHeader      models.ReportHeader    `json:"report_header"`
	TotalRevenue      float64               `json:"total_revenue"`
	TotalTransactions int64                 `json:"total_transactions"`
	AverageOrderValue float64               `json:"average_order_value"`
	TotalPaidAmount   float64               `json:"total_paid_amount"`
	TotalOutstanding  float64               `json:"total_outstanding"`
	TopCustomers      []TopCustomer         `json:"top_customers"`
	TopProducts       []TopProduct          `json:"top_products"`
	SalesByStatus     []StatusSummary       `json:"sales_by_status"`
	GrowthAnalysis    SalesGrowthAnalysis   `json:"growth_analysis"`
}

type VendorAnalysisReport struct {
	ReportHeader         models.ReportHeader    `json:"report_header"`
	TotalPurchases       float64               `json:"total_purchases"`
	TotalTransactions    int64                 `json:"total_transactions"`
	TotalVendors         int64                 `json:"total_vendors"`
	AveragePurchaseValue float64               `json:"average_purchase_value"`
	VendorPerformances   []VendorPerformance   `json:"vendor_performances"`
	TotalPaidAmount      float64               `json:"total_paid_amount"`
	TotalOutstanding     float64               `json:"total_outstanding"`
	PaymentPerformance   PaymentPerformance    `json:"payment_performance"`
}

type VendorPerformance struct {
	VendorID               uint    `json:"vendor_id"`
	VendorName             string  `json:"vendor_name"`
	TotalPurchases         float64 `json:"total_purchases"`
	TotalTransactions      int64   `json:"total_transactions"`
	AveragePurchaseValue   float64 `json:"average_purchase_value"`
	TotalPaid              float64 `json:"total_paid"`
	TotalOutstanding       float64 `json:"total_outstanding"`
	PaidOnTime             int64   `json:"paid_on_time"`
	Overdue                int64   `json:"overdue"`
	PaymentPerformanceScore float64 `json:"payment_performance_score"`
}

type TopCustomer struct {
	CustomerID   uint    `json:"customer_id"`
	CustomerName string  `json:"customer_name"`
	TotalSales   float64 `json:"total_sales"`
	Transactions int64   `json:"transactions"`
}

type TopProduct struct {
	ProductID   uint    `json:"product_id"`
	ProductName string  `json:"product_name"`
	TotalSales  float64 `json:"total_sales"`
	Quantity    int64   `json:"quantity"`
}

type StatusSummary struct {
	Status       string `json:"status"`
	Count        int64  `json:"count"`
	Percentage   float64 `json:"percentage"`
}

type SalesGrowthAnalysis struct {
	CurrentPeriodRevenue  float64 `json:"current_period_revenue"`
	PreviousPeriodRevenue float64 `json:"previous_period_revenue"`
	GrowthRate            float64 `json:"growth_rate"`
	GrowthAmount          float64 `json:"growth_amount"`
	Trend                 string  `json:"trend"`
}

type PaymentPerformance struct {
	OnTimePaymentRate float64 `json:"on_time_payment_rate"`
	OverdueRate       float64 `json:"overdue_rate"`
	AveragePaymentDays float64 `json:"average_payment_days"`
	TotalOverdueAmount float64 `json:"total_overdue_amount"`
}

type UnifiedCompanyInfo struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
	Website    string `json:"website"`
	TaxNumber  string `json:"tax_number"`
	Currency   string `json:"currency"`
}

// ========================= CASH FLOW CALCULATIONS =========================

// calculateOperatingCashFlow calculates cash flow from operating activities
func (s *UnifiedFinancialReportService) calculateOperatingCashFlow(ctx context.Context, startDate, endDate time.Time) (models.CashFlowSection, error) {
	// Get net income from P&L
	pnl, err := s.GenerateComprehensiveProfitLoss(ctx, startDate, endDate, false)
	if err != nil {
		return models.CashFlowSection{}, err
	}

	// Get non-cash items (depreciation, amortization)
	depreciationAmount := s.getDepreciationForPeriod(ctx, startDate, endDate)
	amortizationAmount := s.getAmortizationForPeriod(ctx, startDate, endDate)

	// Calculate working capital changes
	workingCapitalChange := s.calculateWorkingCapitalChange(ctx, startDate, endDate)

	// Build operating activities
	operatingItems := []models.CashFlowItem{
		{
			Description: "Net Income",
			Amount:      pnl.NetIncome,
			AccountCode: "",
			AccountName: "From Profit & Loss Statement",
		},
	}

	// Add depreciation if any
	if depreciationAmount != 0 {
		operatingItems = append(operatingItems, models.CashFlowItem{
			Description: "Depreciation Expense",
			Amount:      depreciationAmount,
			AccountCode: "",
			AccountName: "Non-cash expense",
		})
	}

	// Add amortization if any
	if amortizationAmount != 0 {
		operatingItems = append(operatingItems, models.CashFlowItem{
			Description: "Amortization Expense",
			Amount:      amortizationAmount,
			AccountCode: "",
			AccountName: "Non-cash expense",
		})
	}

	// Add working capital changes
	if workingCapitalChange.ReceivablesChange != 0 {
		operatingItems = append(operatingItems, models.CashFlowItem{
			Description: "Change in Accounts Receivable",
			Amount:      -workingCapitalChange.ReceivablesChange, // Negative because increase in AR decreases cash
			AccountCode: "1201",
			AccountName: "Accounts Receivable",
		})
	}

	if workingCapitalChange.InventoryChange != 0 {
		operatingItems = append(operatingItems, models.CashFlowItem{
			Description: "Change in Inventory",
			Amount:      -workingCapitalChange.InventoryChange, // Negative because increase in inventory decreases cash
			AccountCode: "1301",
			AccountName: "Inventory",
		})
	}

	if workingCapitalChange.PayablesChange != 0 {
		operatingItems = append(operatingItems, models.CashFlowItem{
			Description: "Change in Accounts Payable",
			Amount:      workingCapitalChange.PayablesChange, // Positive because increase in AP increases cash
			AccountCode: "2001",
			AccountName: "Accounts Payable",
		})
	}

	// Calculate total operating cash flow
	var totalOperating float64
	for _, item := range operatingItems {
		totalOperating += item.Amount
	}

	return models.CashFlowSection{
		Items: operatingItems,
		Total: totalOperating,
	}, nil
}

// calculateInvestingCashFlow calculates cash flow from investing activities
func (s *UnifiedFinancialReportService) calculateInvestingCashFlow(ctx context.Context, startDate, endDate time.Time) (models.CashFlowSection, error) {
	// Get changes in fixed assets
	fixedAssetPurchases := s.getFixedAssetPurchases(ctx, startDate, endDate)
	fixedAssetSales := s.getFixedAssetSales(ctx, startDate, endDate)

	investingItems := []models.CashFlowItem{}

	// Add fixed asset purchases (cash outflow)
	if fixedAssetPurchases != 0 {
		investingItems = append(investingItems, models.CashFlowItem{
			Description: "Purchase of Fixed Assets",
			Amount:      -fixedAssetPurchases, // Negative for cash outflow
			AccountCode: "120X",
			AccountName: "Fixed Assets",
		})
	}

	// Add fixed asset sales (cash inflow)
	if fixedAssetSales != 0 {
		investingItems = append(investingItems, models.CashFlowItem{
			Description: "Sale of Fixed Assets",
			Amount:      fixedAssetSales, // Positive for cash inflow
			AccountCode: "120X",
			AccountName: "Fixed Assets",
		})
	}

	// Calculate total investing cash flow
	var totalInvesting float64
	for _, item := range investingItems {
		totalInvesting += item.Amount
	}

	return models.CashFlowSection{
		Items: investingItems,
		Total: totalInvesting,
	}, nil
}

// calculateFinancingCashFlow calculates cash flow from financing activities
func (s *UnifiedFinancialReportService) calculateFinancingCashFlow(ctx context.Context, startDate, endDate time.Time) (models.CashFlowSection, error) {
	// Get changes in long-term debt and equity
	longTermDebtChange := s.getLongTermDebtChange(ctx, startDate, endDate)
	equityChange := s.getEquityChange(ctx, startDate, endDate)
	dividendPayments := s.getDividendPayments(ctx, startDate, endDate)

	financingItems := []models.CashFlowItem{}

	// Add new borrowings or debt repayments
	if longTermDebtChange > 0 {
		financingItems = append(financingItems, models.CashFlowItem{
			Description: "New Borrowings",
			Amount:      longTermDebtChange,
			AccountCode: "220X",
			AccountName: "Long-term Debt",
		})
	} else if longTermDebtChange < 0 {
		financingItems = append(financingItems, models.CashFlowItem{
			Description: "Debt Repayments",
			Amount:      longTermDebtChange, // Already negative
			AccountCode: "220X",
			AccountName: "Long-term Debt",
		})
	}

	// Add equity changes
	if equityChange != 0 {
		financingItems = append(financingItems, models.CashFlowItem{
			Description: "Equity Changes",
			Amount:      equityChange,
			AccountCode: "300X",
			AccountName: "Equity",
		})
	}

	// Add dividend payments
	if dividendPayments != 0 {
		financingItems = append(financingItems, models.CashFlowItem{
			Description: "Dividend Payments",
			Amount:      -dividendPayments, // Negative for cash outflow
			AccountCode: "330X",
			AccountName: "Dividends",
		})
	}

	// Calculate total financing cash flow
	var totalFinancing float64
	for _, item := range financingItems {
		totalFinancing += item.Amount
	}

	return models.CashFlowSection{
		Items: financingItems,
		Total: totalFinancing,
	}, nil
}

// ========================= CASH FLOW HELPER METHODS =========================

type WorkingCapitalChange struct {
	ReceivablesChange float64
	InventoryChange   float64
	PayablesChange    float64
}

// calculateWorkingCapitalChange calculates changes in working capital components
func (s *UnifiedFinancialReportService) calculateWorkingCapitalChange(ctx context.Context, startDate, endDate time.Time) WorkingCapitalChange {
	// Get beginning balances
	beginningDate := startDate.Add(-time.Hour * 24)
	
	// Accounts Receivable changes
	beginningAR := s.getAccountBalanceByCode(ctx, "1201", beginningDate)
	endingAR := s.getAccountBalanceByCode(ctx, "1201", endDate)
	
	// Inventory changes
	beginningInventory := s.getAccountBalanceByCode(ctx, "1301", beginningDate)
	endingInventory := s.getAccountBalanceByCode(ctx, "1301", endDate)
	
	// Accounts Payable changes
	beginningAP := s.getAccountBalanceByCode(ctx, "2001", beginningDate)
	endingAP := s.getAccountBalanceByCode(ctx, "2001", endDate)

	return WorkingCapitalChange{
		ReceivablesChange: endingAR - beginningAR,
		InventoryChange:   endingInventory - beginningInventory,
		PayablesChange:    endingAP - beginningAP,
	}
}

// getDepreciationForPeriod gets total depreciation expense for the period
func (s *UnifiedFinancialReportService) getDepreciationForPeriod(ctx context.Context, startDate, endDate time.Time) float64 {
	var total float64
	
	s.db.WithContext(ctx).
		Table("journal_lines jl").
		Joins("JOIN journal_entries je ON jl.journal_entry_id = je.id").
		Joins("JOIN accounts a ON jl.account_id = a.id").
		Where("a.category = ? AND je.entry_date BETWEEN ? AND ? AND je.status = ?", 
			models.CategoryDepreciationExp, startDate, endDate, models.JournalStatusPosted).
		Select("SUM(jl.debit_amount)").
		Scan(&total)

	return total
}

// getAmortizationForPeriod gets total amortization expense for the period
func (s *UnifiedFinancialReportService) getAmortizationForPeriod(ctx context.Context, startDate, endDate time.Time) float64 {
	var total float64
	
	s.db.WithContext(ctx).
		Table("journal_lines jl").
		Joins("JOIN journal_entries je ON jl.journal_entry_id = je.id").
		Joins("JOIN accounts a ON jl.account_id = a.id").
		Where("a.category = ? AND je.entry_date BETWEEN ? AND ? AND je.status = ?", 
			models.CategoryAmortizationExp, startDate, endDate, models.JournalStatusPosted).
		Select("SUM(jl.debit_amount)").
		Scan(&total)

	return total
}

// getFixedAssetPurchases gets total fixed asset purchases for the period
func (s *UnifiedFinancialReportService) getFixedAssetPurchases(ctx context.Context, startDate, endDate time.Time) float64 {
	var total float64
	
	s.db.WithContext(ctx).
		Table("journal_lines jl").
		Joins("JOIN journal_entries je ON jl.journal_entry_id = je.id").
		Joins("JOIN accounts a ON jl.account_id = a.id").
		Where("a.category = ? AND je.entry_date BETWEEN ? AND ? AND je.status = ? AND je.reference_type = ?", 
			models.CategoryFixedAsset, startDate, endDate, models.JournalStatusPosted, models.JournalRefAsset).
		Select("SUM(jl.debit_amount)").
		Scan(&total)

	return total
}

// getFixedAssetSales gets total fixed asset sales for the period
func (s *UnifiedFinancialReportService) getFixedAssetSales(ctx context.Context, startDate, endDate time.Time) float64 {
	// This would track asset disposal journal entries
	var total float64
	
	s.db.WithContext(ctx).
		Table("journal_lines jl").
		Joins("JOIN journal_entries je ON jl.journal_entry_id = je.id").
		Joins("JOIN accounts a ON jl.account_id = a.id").
		Where("a.category = ? AND je.entry_date BETWEEN ? AND ? AND je.status = ? AND je.reference_type = ?", 
			models.CategoryFixedAsset, startDate, endDate, models.JournalStatusPosted, "ASSET_DISPOSAL").
		Select("SUM(jl.credit_amount)").
		Scan(&total)

	return total
}

// getLongTermDebtChange gets change in long-term debt for the period
func (s *UnifiedFinancialReportService) getLongTermDebtChange(ctx context.Context, startDate, endDate time.Time) float64 {
	beginningDate := startDate.Add(-time.Hour * 24)
	
	beginningDebt := s.getAccountBalanceByCode(ctx, "2201", beginningDate) // Long-term debt account
	endingDebt := s.getAccountBalanceByCode(ctx, "2201", endDate)
	
	return endingDebt - beginningDebt
}

// getEquityChange gets change in equity for the period
func (s *UnifiedFinancialReportService) getEquityChange(ctx context.Context, startDate, endDate time.Time) float64 {
	// Get total equity change (excluding retained earnings which come from operations)
	var endingEquity float64
	
	s.db.WithContext(ctx).
		Table("accounts").
		Where("type = ? AND category IN (?, ?) AND is_active = ?", 
			models.AccountTypeEquity, models.CategoryShareCapital, models.CategoryEquity, true).
		Select("SUM(balance)").
		Scan(&endingEquity)
	
	// For beginning equity, we'd need to calculate based on journal entries
	// This is simplified - for now return 0 to avoid complex calculation
	return 0 // Simplified - would need complex calculation for actual changes
}

// getDividendPayments gets total dividend payments for the period
func (s *UnifiedFinancialReportService) getDividendPayments(ctx context.Context, startDate, endDate time.Time) float64 {
	var total float64
	
	// Look for dividend payment journal entries
	s.db.WithContext(ctx).
		Table("journal_lines jl").
		Joins("JOIN journal_entries je ON jl.journal_entry_id = je.id").
		Joins("JOIN accounts a ON jl.account_id = a.id").
		Where("(a.name LIKE '%dividend%' OR a.name LIKE '%dividen%') AND je.entry_date BETWEEN ? AND ? AND je.status = ?", 
			startDate, endDate, models.JournalStatusPosted).
		Select("SUM(jl.debit_amount)").
		Scan(&total)

	return total
}

// getAccountBalanceByCode gets account balance by account code at a specific date
func (s *UnifiedFinancialReportService) getAccountBalanceByCode(ctx context.Context, code string, asOfDate time.Time) float64 {
	var accountID uint
	err := s.db.WithContext(ctx).Select("id").Where("code = ? AND is_active = ?", code, true).First(&accountID).Error
	if err != nil {
		return 0
	}
	
	balance, _ := s.getAccountBalanceAsOfDate(ctx, accountID, asOfDate)
	return balance
}

// ========================= SALES & VENDOR ANALYSIS HELPERS =========================

// getTopCustomers gets top customers by sales amount
func (s *UnifiedFinancialReportService) getTopCustomers(ctx context.Context, customerMap map[uint]float64, limit int) []TopCustomer {
	type customerSales struct {
		CustomerID uint
		TotalSales float64
	}
	
	var customerSalesList []customerSales
	for customerID, totalSales := range customerMap {
		customerSalesList = append(customerSalesList, customerSales{
			CustomerID: customerID,
			TotalSales: totalSales,
		})
	}
	
	// Sort by total sales descending
	sort.Slice(customerSalesList, func(i, j int) bool {
		return customerSalesList[i].TotalSales > customerSalesList[j].TotalSales
	})
	
	// Get customer details and build result
	var topCustomers []TopCustomer
	for i, cs := range customerSalesList {
		if i >= limit {
			break
		}
		
		// Get customer name
		var customerName string
		s.db.WithContext(ctx).Table("contacts").Select("name").Where("id = ?", cs.CustomerID).Scan(&customerName)
		
		// Count transactions
		var transactionCount int64
		s.db.WithContext(ctx).Table("sales").Where("customer_id = ?", cs.CustomerID).Count(&transactionCount)
		
		topCustomers = append(topCustomers, TopCustomer{
			CustomerID:   cs.CustomerID,
			CustomerName: customerName,
			TotalSales:   cs.TotalSales,
			Transactions: transactionCount,
		})
	}
	
	return topCustomers
}

// getTopProducts gets top products by sales amount
func (s *UnifiedFinancialReportService) getTopProducts(ctx context.Context, productMap map[uint]float64, limit int) []TopProduct {
	type productSales struct {
		ProductID  uint
		TotalSales float64
	}
	
	var productSalesList []productSales
	for productID, totalSales := range productMap {
		productSalesList = append(productSalesList, productSales{
			ProductID:  productID,
			TotalSales: totalSales,
		})
	}
	
	// Sort by total sales descending
	sort.Slice(productSalesList, func(i, j int) bool {
		return productSalesList[i].TotalSales > productSalesList[j].TotalSales
	})
	
	// Get product details and build result
	var topProducts []TopProduct
	for i, ps := range productSalesList {
		if i >= limit {
			break
		}
		
		// Get product name
		var productName string
		s.db.WithContext(ctx).Table("products").Select("name").Where("id = ?", ps.ProductID).Scan(&productName)
		
		// Count quantity sold
		var quantitySold int64
		s.db.WithContext(ctx).Table("sale_items").Select("SUM(quantity)").Where("product_id = ?", ps.ProductID).Scan(&quantitySold)
		
		topProducts = append(topProducts, TopProduct{
			ProductID:   ps.ProductID,
			ProductName: productName,
			TotalSales:  ps.TotalSales,
			Quantity:    quantitySold,
		})
	}
	
	return topProducts
}

// buildStatusSummary builds status summary from status map
func (s *UnifiedFinancialReportService) buildStatusSummary(statusMap map[string]int64) []StatusSummary {
	var total int64
	for _, count := range statusMap {
		total += count
	}
	
	var statusSummaries []StatusSummary
	for status, count := range statusMap {
		percentage := s.safeDiv(float64(count), float64(total)) * 100
		statusSummaries = append(statusSummaries, StatusSummary{
			Status:     status,
			Count:      count,
			Percentage: percentage,
		})
	}
	
	// Sort by count descending
	sort.Slice(statusSummaries, func(i, j int) bool {
		return statusSummaries[i].Count > statusSummaries[j].Count
	})
	
	return statusSummaries
}

// calculateSalesGrowth calculates sales growth compared to previous period
func (s *UnifiedFinancialReportService) calculateSalesGrowth(ctx context.Context, startDate, endDate time.Time) (*SalesGrowthAnalysis, error) {
	// Calculate previous period
	duration := endDate.Sub(startDate)
	prevEndDate := startDate.Add(-time.Hour * 24)
	prevStartDate := prevEndDate.Add(-duration)
	
	// Get current period revenue
	currentRevenue := s.getSalesRevenueForPeriod(ctx, startDate, endDate)
	
	// Get previous period revenue
	prevRevenue := s.getSalesRevenueForPeriod(ctx, prevStartDate, prevEndDate)
	
	// Calculate growth
	growthAmount := currentRevenue - prevRevenue
	growthRate := s.safeDiv(growthAmount, prevRevenue) * 100
	
	var trend string
	if growthRate > 5 {
		trend = "GROWING"
	} else if growthRate < -5 {
		trend = "DECLINING"
	} else {
		trend = "STABLE"
	}
	
	return &SalesGrowthAnalysis{
		CurrentPeriodRevenue:  currentRevenue,
		PreviousPeriodRevenue: prevRevenue,
		GrowthRate:            growthRate,
		GrowthAmount:          growthAmount,
		Trend:                 trend,
	}, nil
}

// getSalesRevenueForPeriod gets total sales revenue for a period
func (s *UnifiedFinancialReportService) getSalesRevenueForPeriod(ctx context.Context, startDate, endDate time.Time) float64 {
	var total float64
	
	s.db.WithContext(ctx).
		Table("sales").
		Where("date BETWEEN ? AND ? AND status IN (?, ?, ?, ?)", 
			startDate, endDate, "CONFIRMED", "COMPLETED", "INVOICED", "PAID").
		Select("SUM(total_amount)").
		Scan(&total)
	
	return total
}

// calculatePaymentScore calculates payment performance score for a vendor
func (s *UnifiedFinancialReportService) calculatePaymentScore(paidOnTime, overdue, totalTransactions int64) float64 {
	if totalTransactions == 0 {
		return 0
	}
	
	onTimeRate := s.safeDiv(float64(paidOnTime), float64(totalTransactions))
	overdueRate := s.safeDiv(float64(overdue), float64(totalTransactions))
	
	// Score: 100 * on-time rate - 50 * overdue rate
	score := (onTimeRate * 100) - (overdueRate * 50)
	return math.Max(0, math.Min(100, score))
}

// calculateOverallPaymentPerformance calculates overall payment performance across all vendors
func (s *UnifiedFinancialReportService) calculateOverallPaymentPerformance(vendors []VendorPerformance) PaymentPerformance {
	var totalPaidOnTime, totalOverdue, totalTransactions int64
	var totalOverdueAmount float64
	
	for _, vendor := range vendors {
		totalPaidOnTime += vendor.PaidOnTime
		totalOverdue += vendor.Overdue
		totalTransactions += vendor.TotalTransactions
		totalOverdueAmount += vendor.TotalOutstanding
	}
	
	onTimeRate := s.safeDiv(float64(totalPaidOnTime), float64(totalTransactions)) * 100
	overdueRate := s.safeDiv(float64(totalOverdue), float64(totalTransactions)) * 100
	
	return PaymentPerformance{
		OnTimePaymentRate:  onTimeRate,
		OverdueRate:       overdueRate,
		AveragePaymentDays: 30, // Would need more complex calculation
		TotalOverdueAmount: totalOverdueAmount,
	}
}
