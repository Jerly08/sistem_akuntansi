package services

import (
	"context"
	"fmt"
	"math"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/utils"
	"gorm.io/gorm"
)

type FinancialReportService interface {
	GenerateProfitLossStatement(ctx context.Context, req *models.FinancialReportRequest) (*models.ProfitLossStatement, error)
	GenerateBalanceSheet(ctx context.Context, req *models.FinancialReportRequest) (*models.BalanceSheet, error)
	GenerateCashFlowStatement(ctx context.Context, req *models.FinancialReportRequest) (*models.CashFlowStatement, error)
	GenerateTrialBalance(ctx context.Context, req *models.FinancialReportRequest) (*models.TrialBalance, error)
	GenerateGeneralLedger(ctx context.Context, accountID uint, startDate, endDate time.Time) (*models.GeneralLedger, error)
	GenerateFinancialDashboard(ctx context.Context) (*models.FinancialDashboard, error)
	CalculateFinancialRatios(ctx context.Context, startDate, endDate time.Time) (*models.FinancialRatios, error)
	GetRealTimeMetrics(ctx context.Context) (*models.RealTimeFinancialMetrics, error)
	CalculateFinancialHealthScore(ctx context.Context) (*models.FinancialHealthScore, error)
}

type FinancialReportServiceImpl struct {
	db              *gorm.DB
	accountRepo     repositories.AccountRepository
	journalRepo     repositories.JournalEntryRepository
}

func NewFinancialReportService(
	db *gorm.DB,
	accountRepo repositories.AccountRepository,
	journalRepo repositories.JournalEntryRepository,
) FinancialReportService {
	return &FinancialReportServiceImpl{
		db:          db,
		accountRepo: accountRepo,
		journalRepo: journalRepo,
	}
}

// GenerateProfitLossStatement generates a Profit & Loss statement from journal entries
func (s *FinancialReportServiceImpl) GenerateProfitLossStatement(ctx context.Context, req *models.FinancialReportRequest) (*models.ProfitLossStatement, error) {
	// Get all revenue accounts with balances from journal entries
	revenueAccounts, err := s.getAccountBalancesFromJournal(ctx, models.AccountTypeRevenue, req.StartDate, req.EndDate, req.ShowZero)
	if err != nil {
		return nil, err
	}

	// Get all expense accounts with balances from journal entries
	expenseAccounts, err := s.getAccountBalancesFromJournal(ctx, models.AccountTypeExpense, req.StartDate, req.EndDate, req.ShowZero)
	if err != nil {
		return nil, err
	}

	// Separate COGS from other expenses based on account categories
	var cogsAccounts, operatingExpenses []models.AccountLineItem
	for _, expense := range expenseAccounts {
		if expense.Category == models.CategoryCostOfGoodsSold || 
		   expense.Category == models.CategoryDirectMaterial ||
		   expense.Category == models.CategoryDirectLabor ||
		   expense.Category == models.CategoryManufacturingOverhead ||
		   expense.Category == models.CategoryFreightIn {
			cogsAccounts = append(cogsAccounts, expense)
		} else {
			operatingExpenses = append(operatingExpenses, expense)
		}
	}

	// Calculate totals - Revenue is normally credit balance, so we use absolute value
	totalRevenue := s.calculateTotalCreditBalance(revenueAccounts)
	totalCOGS := s.calculateTotalDebitBalance(cogsAccounts)
	totalExpenses := s.calculateTotalDebitBalance(operatingExpenses)
	
	grossProfit := totalRevenue - totalCOGS
	netIncome := grossProfit - totalExpenses

	pnl := &models.ProfitLossStatement{
		ReportHeader: models.ReportHeader{
			ReportType:    models.ReportTypeProfitLoss,
			CompanyName:   "PT. Sample Company", // Should be configurable
			ReportTitle:   "Profit & Loss Statement",
			StartDate:     req.StartDate,
			EndDate:       req.EndDate,
			GeneratedAt:   time.Now(),
			Currency:      "IDR",
			IsComparative: req.Comparative,
		},
		Revenue:       revenueAccounts,
		TotalRevenue:  totalRevenue,
		COGS:          cogsAccounts,
		TotalCOGS:     totalCOGS,
		GrossProfit:   grossProfit,
		Expenses:      operatingExpenses,
		TotalExpenses: totalExpenses,
		NetIncome:     netIncome,
	}

	// Add comparative data if requested
	if req.Comparative {
		prevStartDate, prevEndDate := s.calculatePreviousPeriod(req.StartDate, req.EndDate)
		prevPnL, err := s.GenerateProfitLossStatement(ctx, &models.FinancialReportRequest{
			ReportType:  req.ReportType,
			StartDate:   prevStartDate,
			EndDate:     prevEndDate,
			Comparative: false,
			ShowZero:    req.ShowZero,
		})
		if err == nil {
			pnl.Comparative = &models.ProfitLossComparative{
				PreviousPeriod: *prevPnL,
				Variance: models.ProfitLossVariance{
					RevenueVariance:     totalRevenue - prevPnL.TotalRevenue,
					COGSVariance:        totalCOGS - prevPnL.TotalCOGS,
					GrossProfitVariance: grossProfit - prevPnL.GrossProfit,
					ExpenseVariance:     totalExpenses - prevPnL.TotalExpenses,
					NetIncomeVariance:   netIncome - prevPnL.NetIncome,
				},
			}
		}
	}

	return pnl, nil
}

// GenerateBalanceSheet generates a Balance Sheet from journal entries
func (s *FinancialReportServiceImpl) GenerateBalanceSheet(ctx context.Context, req *models.FinancialReportRequest) (*models.BalanceSheet, error) {
	// Get account balances by type as of end date from journal entries
	assetAccounts, err := s.getAccountBalancesFromJournalAsOfDate(ctx, models.AccountTypeAsset, req.EndDate, req.ShowZero)
	if err != nil {
		return nil, err
	}

	liabilityAccounts, err := s.getAccountBalancesFromJournalAsOfDate(ctx, models.AccountTypeLiability, req.EndDate, req.ShowZero)
	if err != nil {
		return nil, err
	}

	equityAccounts, err := s.getAccountBalancesFromJournalAsOfDate(ctx, models.AccountTypeEquity, req.EndDate, req.ShowZero)
	if err != nil {
		return nil, err
	}

	// Calculate net income for current period and add to retained earnings
	netIncome, err := s.calculateNetIncomeForPeriod(ctx, req.EndDate)
	if err == nil && netIncome != 0 {
		// Add net income to equity accounts (retained earnings)
		for i, equity := range equityAccounts {
			if equity.Category == models.CategoryRetainedEarnings {
				equityAccounts[i].Balance += netIncome
				break
			}
		}
	}

	// Group accounts by categories
	assetSection := s.groupAccountsByCategory(assetAccounts)
	liabilitySection := s.groupAccountsByCategory(liabilityAccounts)
	equitySection := s.groupAccountsByCategory(equityAccounts)

	totalAssets := s.calculateTotalBalance(assetAccounts)
	totalLiabilities := s.calculateTotalBalance(liabilityAccounts)
	totalEquity := s.calculateTotalBalance(equityAccounts)

	balanceSheet := &models.BalanceSheet{
		ReportHeader: models.ReportHeader{
			ReportType:    models.ReportTypeBalanceSheet,
			CompanyName:   "PT. Sample Company",
			ReportTitle:   "Balance Sheet",
			StartDate:     req.StartDate,
			EndDate:       req.EndDate,
			GeneratedAt:   time.Now(),
			Currency:      "IDR",
			IsComparative: req.Comparative,
		},
		Assets:           assetSection,
		Liabilities:      liabilitySection,
		Equity:           equitySection,
		TotalAssets:      totalAssets,
		TotalLiabilities: totalLiabilities,
		TotalEquity:      totalEquity,
		IsBalanced:       math.Abs(totalAssets-(totalLiabilities+totalEquity)) < 0.01, // Allow small rounding differences
	}

	// Add comparative data if requested
	if req.Comparative {
		prevEndDate := s.calculatePreviousYearEnd(req.EndDate)
		prevBalanceSheet, err := s.GenerateBalanceSheet(ctx, &models.FinancialReportRequest{
			ReportType:  req.ReportType,
			StartDate:   req.StartDate, // Keep same start date
			EndDate:     prevEndDate,
			Comparative: false,
			ShowZero:    req.ShowZero,
		})
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

// GenerateCashFlowStatement generates a Cash Flow statement (indirect method)
// Helper methods for accessing journal entries data

// getAccountBalancesFromJournal gets account balances from journal entries for a period
func (s *FinancialReportServiceImpl) getAccountBalancesFromJournal(ctx context.Context, accountType string, startDate, endDate time.Time, showZero bool) ([]models.AccountLineItem, error) {
	var accountItems []models.AccountLineItem
	
	// Query untuk mendapatkan balances dari journal entries yang sudah di-post
	query := `
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			a.category,
			COALESCE(SUM(jl.debit_amount), 0) as total_debit,
			COALESCE(SUM(jl.credit_amount), 0) as total_credit,
			COALESCE(SUM(jl.debit_amount) - SUM(jl.credit_amount), 0) as balance
		FROM accounts a
		LEFT JOIN journal_lines jl ON a.id = jl.account_id
		LEFT JOIN journal_entries je ON jl.journal_entry_id = je.id
		WHERE a.type = ? 
			AND a.is_active = true
			AND (je.id IS NULL OR (je.status = 'POSTED' AND je.entry_date BETWEEN ? AND ?))
			AND a.deleted_at IS NULL
		GROUP BY a.id, a.code, a.name, a.type, a.category
	`
	
	if !showZero {
		query += " HAVING ABS(COALESCE(SUM(jl.debit_amount) - SUM(jl.credit_amount), 0)) > 0.01"
	}
	
	query += " ORDER BY a.code"
	
	type QueryResult struct {
		AccountID     uint    `db:"account_id"`
		AccountCode   string  `db:"account_code"`
		AccountName   string  `db:"account_name"`
		AccountType   string  `db:"account_type"`
		Category      string  `db:"category"`
		TotalDebit    float64 `db:"total_debit"`
		TotalCredit   float64 `db:"total_credit"`
		Balance       float64 `db:"balance"`
	}
	
	var results []QueryResult
	err := s.db.Raw(query, accountType, startDate, endDate).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get account balances from journal: %v", err)
	}
	
	for _, result := range results {
		// Adjust balance based on account normal balance
		var adjustedBalance float64
		if accountType == models.AccountTypeRevenue || accountType == models.AccountTypeLiability || accountType == models.AccountTypeEquity {
			// Credit accounts - show as positive when credit > debit
			adjustedBalance = result.TotalCredit - result.TotalDebit
		} else {
			// Debit accounts (Assets, Expenses) - show as positive when debit > credit
			adjustedBalance = result.TotalDebit - result.TotalCredit
		}
		
		accountItems = append(accountItems, models.AccountLineItem{
			AccountID:   result.AccountID,
			AccountCode: result.AccountCode,
			AccountName: result.AccountName,
			AccountType: result.AccountType,
			Category:    result.Category,
			Balance:     adjustedBalance,
		})
	}
	
	return accountItems, nil
}

// getAccountBalancesFromJournalAsOfDate gets account balances as of a specific date
func (s *FinancialReportServiceImpl) getAccountBalancesFromJournalAsOfDate(ctx context.Context, accountType string, asOfDate time.Time, showZero bool) ([]models.AccountLineItem, error) {
	var accountItems []models.AccountLineItem
	
	// Query untuk mendapatkan balance sampai dengan tanggal tertentu
	query := `
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			a.category,
			COALESCE(SUM(jl.debit_amount), 0) as total_debit,
			COALESCE(SUM(jl.credit_amount), 0) as total_credit
		FROM accounts a
		LEFT JOIN journal_lines jl ON a.id = jl.account_id
		LEFT JOIN journal_entries je ON jl.journal_entry_id = je.id
		WHERE a.type = ? 
			AND a.is_active = true
			AND (je.id IS NULL OR (je.status = 'POSTED' AND je.entry_date <= ?))
			AND a.deleted_at IS NULL
		GROUP BY a.id, a.code, a.name, a.type, a.category
	`
	
	if !showZero {
		query += " HAVING ABS(COALESCE(SUM(jl.debit_amount) - SUM(jl.credit_amount), 0)) > 0.01"
	}
	
	query += " ORDER BY a.code"
	
	type QueryResult struct {
		AccountID     uint    `db:"account_id"`
		AccountCode   string  `db:"account_code"`
		AccountName   string  `db:"account_name"`
		AccountType   string  `db:"account_type"`
		Category      string  `db:"category"`
		TotalDebit    float64 `db:"total_debit"`
		TotalCredit   float64 `db:"total_credit"`
	}
	
	var results []QueryResult
	err := s.db.Raw(query, accountType, asOfDate).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get account balances as of date from journal: %v", err)
	}
	
	for _, result := range results {
		// Adjust balance based on account normal balance
		var adjustedBalance float64
		if accountType == models.AccountTypeRevenue || accountType == models.AccountTypeLiability || accountType == models.AccountTypeEquity {
			// Credit accounts
			adjustedBalance = result.TotalCredit - result.TotalDebit
		} else {
			// Debit accounts
			adjustedBalance = result.TotalDebit - result.TotalCredit
		}
		
		accountItems = append(accountItems, models.AccountLineItem{
			AccountID:   result.AccountID,
			AccountCode: result.AccountCode,
			AccountName: result.AccountName,
			AccountType: result.AccountType,
			Category:    result.Category,
			Balance:     adjustedBalance,
		})
	}
	
	return accountItems, nil
}

// calculateTotalCreditBalance calculates total for credit normal balance accounts
func (s *FinancialReportServiceImpl) calculateTotalCreditBalance(accounts []models.AccountLineItem) float64 {
	total := 0.0
	for _, account := range accounts {
		total += account.Balance
	}
	return total
}

// calculateTotalDebitBalance calculates total for debit normal balance accounts
func (s *FinancialReportServiceImpl) calculateTotalDebitBalance(accounts []models.AccountLineItem) float64 {
	total := 0.0
	for _, account := range accounts {
		total += account.Balance
	}
	return total
}

// calculateNetIncomeForPeriod calculates net income for a period to include in balance sheet
func (s *FinancialReportServiceImpl) calculateNetIncomeForPeriod(ctx context.Context, endDate time.Time) (float64, error) {
	// Calculate from beginning of year to end date
	startOfYear := time.Date(endDate.Year(), 1, 1, 0, 0, 0, 0, endDate.Location())
	
	// Get total revenue
	revenueQuery := `
		SELECT COALESCE(SUM(jl.credit_amount - jl.debit_amount), 0) as total_revenue
		FROM journal_lines jl
		JOIN journal_entries je ON jl.journal_entry_id = je.id
		JOIN accounts a ON jl.account_id = a.id
		WHERE a.type = 'REVENUE' 
			AND je.status = 'POSTED'
			AND je.entry_date BETWEEN ? AND ?
	`
	
	var totalRevenue float64
	err := s.db.Raw(revenueQuery, startOfYear, endDate).Scan(&totalRevenue).Error
	if err != nil {
		return 0, err
	}
	
	// Get total expenses
	expenseQuery := `
		SELECT COALESCE(SUM(jl.debit_amount - jl.credit_amount), 0) as total_expenses
		FROM journal_lines jl
		JOIN journal_entries je ON jl.journal_entry_id = je.id
		JOIN accounts a ON jl.account_id = a.id
		WHERE a.type = 'EXPENSE'
			AND je.status = 'POSTED'
			AND je.entry_date BETWEEN ? AND ?
	`
	
	var totalExpenses float64
	err = s.db.Raw(expenseQuery, startOfYear, endDate).Scan(&totalExpenses).Error
	if err != nil {
		return 0, err
	}
	
	return totalRevenue - totalExpenses, nil
}

// groupAccountsByCategory groups accounts by their categories for balance sheet
func (s *FinancialReportServiceImpl) groupAccountsByCategory(accounts []models.AccountLineItem) models.BalanceSheetSection {
	categoryMap := make(map[string][]models.AccountLineItem)
	totalBalance := 0.0
	
	for _, account := range accounts {
		category := account.Category
		if category == "" {
			category = "Other"
		}
		categoryMap[category] = append(categoryMap[category], account)
		totalBalance += account.Balance
	}
	
	var categories []models.BalanceSheetCategory
	for categoryName, categoryAccounts := range categoryMap {
		categoryTotal := 0.0
		for _, account := range categoryAccounts {
			categoryTotal += account.Balance
		}
		
		categories = append(categories, models.BalanceSheetCategory{
			Name:     categoryName,
			Accounts: categoryAccounts,
			Total:    categoryTotal,
		})
	}
	
	return models.BalanceSheetSection{
		Categories: categories,
		Total:      totalBalance,
	}
}

// calculatePreviousPeriod calculates previous period dates for comparison
func (s *FinancialReportServiceImpl) calculatePreviousPeriod(startDate, endDate time.Time) (time.Time, time.Time) {
	duration := endDate.Sub(startDate)
	prevEndDate := startDate.Add(-24 * time.Hour)
	prevStartDate := prevEndDate.Add(-duration)
	return prevStartDate, prevEndDate
}

// calculatePreviousYearEnd calculates previous year end date
func (s *FinancialReportServiceImpl) calculatePreviousYearEnd(endDate time.Time) time.Time {
	return time.Date(endDate.Year()-1, endDate.Month(), endDate.Day(), endDate.Hour(), endDate.Minute(), endDate.Second(), endDate.Nanosecond(), endDate.Location())
}

func (s *FinancialReportServiceImpl) GenerateCashFlowStatement(ctx context.Context, req *models.FinancialReportRequest) (*models.CashFlowStatement, error) {
	// Get cash account balances at beginning and end of period
	beginningCash, err := s.getTotalCashBalance(ctx, req.StartDate)
	if err != nil {
		return nil, err
	}

	endingCash, err := s.getTotalCashBalance(ctx, req.EndDate)
	if err != nil {
		return nil, err
	}

	// Get operating activities (simplified - based on P&L items and working capital changes)
	operatingActivities, err := s.calculateOperatingCashFlow(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	// Get investing activities (changes in fixed assets)
	investingActivities, err := s.calculateInvestingCashFlow(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	// Get financing activities (changes in debt and equity)
	financingActivities, err := s.calculateFinancingCashFlow(ctx, req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	netCashFlow := operatingActivities.Total + investingActivities.Total + financingActivities.Total

	cashFlow := &models.CashFlowStatement{
		ReportHeader: models.ReportHeader{
			ReportType:    models.ReportTypeCashFlow,
			CompanyName:   "PT. Sample Company",
			ReportTitle:   "Cash Flow Statement",
			StartDate:     req.StartDate,
			EndDate:       req.EndDate,
			GeneratedAt:   time.Now(),
			Currency:      "IDR",
			IsComparative: req.Comparative,
		},
		OperatingActivities: operatingActivities,
		InvestingActivities: investingActivities,
		FinancingActivities: financingActivities,
		NetCashFlow:         netCashFlow,
		BeginningCash:       beginningCash,
		EndingCash:          endingCash,
	}

	return cashFlow, nil
}

// GenerateTrialBalance generates a Trial Balance from journal entries
func (s *FinancialReportServiceImpl) GenerateTrialBalance(ctx context.Context, req *models.FinancialReportRequest) (*models.TrialBalance, error) {
	// Get all accounts with their balances from journal entries
	var trialBalanceItems []models.TrialBalanceItem
	
	query := `
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			COALESCE(SUM(jl.debit_amount), 0) as total_debit,
			COALESCE(SUM(jl.credit_amount), 0) as total_credit
		FROM accounts a
		LEFT JOIN journal_lines jl ON a.id = jl.account_id
		LEFT JOIN journal_entries je ON jl.journal_entry_id = je.id
		WHERE a.is_active = true
			AND (je.id IS NULL OR (je.status = 'POSTED' AND je.entry_date <= ?))
			AND a.deleted_at IS NULL
		GROUP BY a.id, a.code, a.name, a.type
	`
	
	if !req.ShowZero {
		query += " HAVING ABS(COALESCE(SUM(jl.debit_amount) - SUM(jl.credit_amount), 0)) > 0.01"
	}
	
	query += " ORDER BY a.code"
	
	type QueryResult struct {
		AccountID     uint    `db:"account_id"`
		AccountCode   string  `db:"account_code"`
		AccountName   string  `db:"account_name"`
		AccountType   string  `db:"account_type"`
		TotalDebit    float64 `db:"total_debit"`
		TotalCredit   float64 `db:"total_credit"`
	}
	
	var results []QueryResult
	err := s.db.Raw(query, req.EndDate).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to generate trial balance: %v", err)
	}
	
	var totalDebits, totalCredits float64
	
	for _, result := range results {
		// For trial balance, we show actual debit/credit balances
		var debitBalance, creditBalance float64
		
		// Determine which side the balance should appear on
		netBalance := result.TotalDebit - result.TotalCredit
		
		if netBalance > 0 {
			debitBalance = netBalance
			creditBalance = 0
		} else if netBalance < 0 {
			debitBalance = 0
			creditBalance = -netBalance
		} else {
			debitBalance = 0
			creditBalance = 0
		}
		
		totalDebits += debitBalance
		totalCredits += creditBalance
		
		trialBalanceItems = append(trialBalanceItems, models.TrialBalanceItem{
			AccountID:     result.AccountID,
			AccountCode:   result.AccountCode,
			AccountName:   result.AccountName,
			AccountType:   result.AccountType,
			DebitBalance:  debitBalance,
			CreditBalance: creditBalance,
		})
	}
	
	trialBalance := &models.TrialBalance{
		ReportHeader: models.ReportHeader{
			ReportType:    models.ReportTypeTrialBalance,
			CompanyName:   "PT. Sample Company",
			ReportTitle:   "Trial Balance",
			StartDate:     req.StartDate,
			EndDate:       req.EndDate,
			GeneratedAt:   time.Now(),
			Currency:      "IDR",
			IsComparative: false,
		},
		Accounts:     trialBalanceItems,
		TotalDebits:  totalDebits,
		TotalCredits: totalCredits,
		IsBalanced:   math.Abs(totalDebits-totalCredits) < 0.01,
	}
	
	return trialBalance, nil
}

// GenerateGeneralLedger generates a General Ledger for a specific account from journal entries
func (s *FinancialReportServiceImpl) GenerateGeneralLedger(ctx context.Context, accountID uint, startDate, endDate time.Time) (*models.GeneralLedger, error) {
	// Get account information
	var account models.Account
	err := s.db.Where("id = ? AND is_active = true", accountID).First(&account).Error
	if err != nil {
		return nil, fmt.Errorf("account not found: %v", err)
	}
	
	// Get beginning balance (before start date)
	beginningBalanceQuery := `
		SELECT COALESCE(SUM(jl.debit_amount - jl.credit_amount), 0) as beginning_balance
		FROM journal_lines jl
		JOIN journal_entries je ON jl.journal_entry_id = je.id
		WHERE jl.account_id = ?
			AND je.status = 'POSTED'
			AND je.entry_date < ?
	`
	
	var beginningBalance float64
	err = s.db.Raw(beginningBalanceQuery, accountID, startDate).Scan(&beginningBalance).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get beginning balance: %v", err)
	}
	
	// Adjust beginning balance based on account type
	if account.Type == models.AccountTypeRevenue || account.Type == models.AccountTypeLiability || account.Type == models.AccountTypeEquity {
		beginningBalance = -beginningBalance // Credit accounts
	}
	
	// Get all transactions for the period
	transactionQuery := `
		SELECT 
			je.entry_date,
			je.code as journal_code,
			je.description,
			je.reference,
			jl.debit_amount,
			jl.credit_amount
		FROM journal_lines jl
		JOIN journal_entries je ON jl.journal_entry_id = je.id
		WHERE jl.account_id = ?
			AND je.status = 'POSTED'
			AND je.entry_date BETWEEN ? AND ?
		ORDER BY je.entry_date, je.code
	`
	
	type TransactionResult struct {
		EntryDate    time.Time `db:"entry_date"`
		JournalCode  string    `db:"journal_code"`
		Description  string    `db:"description"`
		Reference    string    `db:"reference"`
		DebitAmount  float64   `db:"debit_amount"`
		CreditAmount float64   `db:"credit_amount"`
	}
	
	var transactions []TransactionResult
	err = s.db.Raw(transactionQuery, accountID, startDate, endDate).Scan(&transactions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %v", err)
	}
	
	// Build general ledger entries with running balance
	var ledgerEntries []models.GeneralLedgerEntry
	runningBalance := beginningBalance
	totalDebits := 0.0
	totalCredits := 0.0
	
	for _, txn := range transactions {
		// Calculate balance change
		var balanceChange float64
		if account.Type == models.AccountTypeRevenue || account.Type == models.AccountTypeLiability || account.Type == models.AccountTypeEquity {
			// Credit accounts
			balanceChange = txn.CreditAmount - txn.DebitAmount
		} else {
			// Debit accounts
			balanceChange = txn.DebitAmount - txn.CreditAmount
		}
		
		runningBalance += balanceChange
		totalDebits += txn.DebitAmount
		totalCredits += txn.CreditAmount
		
		ledgerEntries = append(ledgerEntries, models.GeneralLedgerEntry{
			Date:          txn.EntryDate,
			JournalCode:   txn.JournalCode,
			Description:   txn.Description,
			Reference:     txn.Reference,
			DebitAmount:   txn.DebitAmount,
			CreditAmount:  txn.CreditAmount,
			Balance:       runningBalance,
		})
	}
	
	generalLedger := &models.GeneralLedger{
		ReportHeader: models.ReportHeader{
			ReportType:    models.ReportTypeGeneralLedger,
			CompanyName:   "PT. Sample Company",
			ReportTitle:   fmt.Sprintf("General Ledger - %s (%s)", account.Name, account.Code),
			StartDate:     startDate,
			EndDate:       endDate,
			GeneratedAt:   time.Now(),
			Currency:      "IDR",
			IsComparative: false,
		},
		Account:          account,
		Transactions:     ledgerEntries,
		BeginningBalance: beginningBalance,
		EndingBalance:    runningBalance,
		TotalDebits:      totalDebits,
		TotalCredits:     totalCredits,
	}
	
	return generalLedger, nil
}

// GenerateFinancialDashboard generates a comprehensive financial dashboard
func (s *FinancialReportServiceImpl) GenerateFinancialDashboard(ctx context.Context) (*models.FinancialDashboard, error) {
	now := time.Now()
	
	// Get key metrics
	keyMetrics, err := s.getFinancialKeyMetrics(ctx, now)
	if err != nil {
		return nil, err
	}

	// Calculate financial ratios
	ratios, err := s.CalculateFinancialRatios(ctx, now.AddDate(-1, 0, 0), now)
	if err != nil {
		return nil, err
	}

	// Get cash position
	cashPosition, err := s.getCashPositionSummary(ctx)
	if err != nil {
		return nil, err
	}

	// Get account balance summary
	accountBalances, err := s.getAccountBalanceSummary(ctx)
	if err != nil {
		return nil, err
	}

	// Get recent activity
	recentActivity, err := s.getRecentActivity(ctx, 10)
	if err != nil {
		return nil, err
	}

	// Generate financial alerts
	alerts := s.generateFinancialAlerts(keyMetrics, ratios)

	return &models.FinancialDashboard{
		ReportDate:      now,
		KeyMetrics:      *keyMetrics,
		Ratios:          *ratios,
		CashPosition:    *cashPosition,
		AccountBalances: accountBalances,
		RecentActivity:  recentActivity,
		Alerts:          alerts,
	}, nil
}

// CalculateFinancialRatios calculates various financial ratios
func (s *FinancialReportServiceImpl) CalculateFinancialRatios(ctx context.Context, startDate, endDate time.Time) (*models.FinancialRatios, error) {
	// Get balance sheet data
	balanceSheetReq := &models.FinancialReportRequest{
		ReportType: models.ReportTypeBalanceSheet,
		StartDate:  startDate,
		EndDate:    endDate,
	}
	balanceSheet, err := s.GenerateBalanceSheet(ctx, balanceSheetReq)
	if err != nil {
		return nil, err
	}

	// Get P&L data
	pnlReq := &models.FinancialReportRequest{
		ReportType: models.ReportTypeProfitLoss,
		StartDate:  startDate,
		EndDate:    endDate,
	}
	pnl, err := s.GenerateProfitLossStatement(ctx, pnlReq)
	if err != nil {
		return nil, err
	}

	// Calculate specific balance components
	currentAssets := s.getBalanceByCategory(balanceSheet.Assets, models.CategoryCurrentAsset)
	currentLiabilities := s.getBalanceByCategory(balanceSheet.Liabilities, models.CategoryCurrentLiability)
	inventory, _ := s.getBalanceByAccountCode(balanceSheet.Assets, "1301") // Assuming 1301 is inventory
	cash, _ := s.getBalanceByAccountCode(balanceSheet.Assets, "1101") // Assuming 1101 is cash

	ratios := &models.FinancialRatios{
		// Liquidity Ratios
		CurrentRatio: s.safeDiv(currentAssets, currentLiabilities),
		QuickRatio:   s.safeDiv(currentAssets-inventory, currentLiabilities),
		CashRatio:    s.safeDiv(cash, currentLiabilities),

		// Profitability Ratios
		GrossProfitMargin: s.safeDiv(pnl.GrossProfit, pnl.TotalRevenue) * 100,
		NetProfitMargin:   s.safeDiv(pnl.NetIncome, pnl.TotalRevenue) * 100,
		ROA:               s.safeDiv(pnl.NetIncome, balanceSheet.TotalAssets) * 100,
		ROE:               s.safeDiv(pnl.NetIncome, balanceSheet.TotalEquity) * 100,

		// Leverage Ratios
		DebtRatio:         s.safeDiv(balanceSheet.TotalLiabilities, balanceSheet.TotalAssets) * 100,
		DebtToEquityRatio: s.safeDiv(balanceSheet.TotalLiabilities, balanceSheet.TotalEquity),

		// Efficiency Ratios
		AssetTurnover:     s.safeDiv(pnl.TotalRevenue, balanceSheet.TotalAssets),
		InventoryTurnover: s.safeDiv(pnl.TotalCOGS, inventory),

		CalculatedAt: time.Now(),
		PeriodStart:  startDate,
		PeriodEnd:    endDate,
	}

	return ratios, nil
}

// GetRealTimeMetrics gets real-time financial metrics for dashboards
func (s *FinancialReportServiceImpl) GetRealTimeMetrics(ctx context.Context) (*models.RealTimeFinancialMetrics, error) {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	// Get current cash position
	cashPosition, err := s.getTotalCashBalance(ctx, now)
	if err != nil {
		return nil, err
	}

	// Get daily metrics
	dailyRevenue, err := s.getTotalByAccountTypeAndPeriod(ctx, models.AccountTypeRevenue, today, now)
	if err != nil {
		return nil, err
	}

	dailyExpenses, err := s.getTotalByAccountTypeAndPeriod(ctx, models.AccountTypeExpense, today, now)
	if err != nil {
		return nil, err
	}

	// Get monthly metrics
	monthlyRevenue, err := s.getTotalByAccountTypeAndPeriod(ctx, models.AccountTypeRevenue, monthStart, now)
	if err != nil {
		return nil, err
	}

	monthlyExpenses, err := s.getTotalByAccountTypeAndPeriod(ctx, models.AccountTypeExpense, monthStart, now)
	if err != nil {
		return nil, err
	}

	// Get yearly metrics
	yearlyRevenue, err := s.getTotalByAccountTypeAndPeriod(ctx, models.AccountTypeRevenue, yearStart, now)
	if err != nil {
		return nil, err
	}

	yearlyExpenses, err := s.getTotalByAccountTypeAndPeriod(ctx, models.AccountTypeExpense, yearStart, now)
	if err != nil {
		return nil, err
	}

	// Get receivables and payables (use direct database lookup)
	receivables := s.getAccountBalanceByCode(ctx, "1201") // Assuming 1201 is accounts receivable
	payables := s.getAccountBalanceByCode(ctx, "2001")    // Assuming 2001 is accounts payable
	inventory := s.getAccountBalanceByCode(ctx, "1301")   // Assuming 1301 is inventory

	return &models.RealTimeFinancialMetrics{
		AsOfDate:           now,
		CashPosition:       cashPosition,
		DailyRevenue:       dailyRevenue,
		DailyExpenses:      dailyExpenses,
		DailyNetIncome:     dailyRevenue - dailyExpenses,
		MonthlyRevenue:     monthlyRevenue,
		MonthlyExpenses:    monthlyExpenses,
		MonthlyNetIncome:   monthlyRevenue - monthlyExpenses,
		YearlyRevenue:      yearlyRevenue,
		YearlyExpenses:     yearlyExpenses,
		YearlyNetIncome:    yearlyRevenue - yearlyExpenses,
		PendingReceivables: receivables,
		PendingPayables:    payables,
		InventoryValue:     inventory,
		LastUpdated:        now,
	}, nil
}

// CalculateFinancialHealthScore calculates an overall financial health score
func (s *FinancialReportServiceImpl) CalculateFinancialHealthScore(ctx context.Context) (*models.FinancialHealthScore, error) {
	now := time.Now()
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	// Get financial ratios
	ratios, err := s.CalculateFinancialRatios(ctx, yearStart, now)
	if err != nil {
		return nil, err
	}

	// Calculate component scores (0-100)
	components := models.FinancialHealthComponents{
		LiquidityScore:     s.calculateLiquidityScore(ratios),
		ProfitabilityScore: s.calculateProfitabilityScore(ratios),
		LeverageScore:      s.calculateLeverageScore(ratios),
		EfficiencyScore:    s.calculateEfficiencyScore(ratios),
		GrowthScore:        50, // Would need historical data for growth calculation
	}

	// Calculate overall score (weighted average)
	overallScore := (components.LiquidityScore*0.25 + 
		components.ProfitabilityScore*0.30 + 
		components.LeverageScore*0.20 + 
		components.EfficiencyScore*0.15 + 
		components.GrowthScore*0.10)

	// Determine grade
	grade := s.getScoreGrade(overallScore)

	// Generate recommendations
	recommendations := s.generateHealthRecommendations(components, ratios)

	return &models.FinancialHealthScore{
		OverallScore:    overallScore,
		ScoreGrade:      grade,
		Components:      components,
		Recommendations: recommendations,
		CalculatedAt:    time.Now(),
	}, nil
}

// Helper methods

func (s *FinancialReportServiceImpl) getAccountBalancesByType(ctx context.Context, accountType string, startDate, endDate time.Time, showZero bool) ([]models.AccountLineItem, error) {
	var accounts []models.Account
	query := s.db.WithContext(ctx).Where("type = ? AND is_active = ?", accountType, true)
	
	if !showZero {
		query = query.Where("balance != ?", 0)
	}

	err := query.Find(&accounts).Error
	if err != nil {
		return nil, utils.NewInternalError("Failed to get accounts by type", err)
	}

	var accountItems []models.AccountLineItem
	for _, account := range accounts {
		// For P&L accounts, calculate the balance change during the period
		balance := s.calculateAccountBalanceForPeriod(ctx, account.ID, startDate, endDate)
		
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

func (s *FinancialReportServiceImpl) getAccountBalancesAsOfDate(ctx context.Context, accountType string, asOfDate time.Time, showZero bool) ([]models.AccountLineItem, error) {
	var accounts []models.Account
	query := s.db.WithContext(ctx).Where("type = ? AND is_active = ?", accountType, true)
	
	if !showZero {
		query = query.Where("balance != ?", 0)
	}

	err := query.Find(&accounts).Error
	if err != nil {
		return nil, utils.NewInternalError("Failed to get accounts by type", err)
	}

	var accountItems []models.AccountLineItem
	for _, account := range accounts {
		// For balance sheet accounts, get balance as of specific date
		balance, err := s.getAccountBalanceAsOfDate(ctx, account.ID, asOfDate)
		if err != nil {
			balance = account.Balance // Fallback to current balance
		}
		
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

func (s *FinancialReportServiceImpl) calculateTotalBalance(accounts []models.AccountLineItem) float64 {
	var total float64
	for _, account := range accounts {
		total += account.Balance
	}
	return total
}

// Methods are already defined above, removing duplicates

func (s *FinancialReportServiceImpl) calculateAccountBalanceForPeriod(ctx context.Context, accountID uint, startDate, endDate time.Time) float64 {
	// This is a simplified implementation
	// In a real implementation, you would sum up all journal entries for the account in the period
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

func (s *FinancialReportServiceImpl) getAccountBalanceAsOfDate(ctx context.Context, accountID uint, asOfDate time.Time) (float64, error) {
	// This would calculate the account balance as of a specific date
	// by summing all journal entries up to that date
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

func (s *FinancialReportServiceImpl) getTotalCashBalance(ctx context.Context, asOfDate time.Time) (float64, error) {
	// Sum all cash and bank account balances
	var total float64
	
	err := s.db.WithContext(ctx).
		Table("accounts").
		Where("type = ? AND (category LIKE '%CASH%' OR category LIKE '%BANK%') AND is_active = ?", 
			models.AccountTypeAsset, true).
		Select("SUM(balance)").
		Scan(&total).Error

	return total, err
}

func (s *FinancialReportServiceImpl) safeDiv(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

func (s *FinancialReportServiceImpl) getBalanceByCategory(section models.BalanceSheetSection, category string) float64 {
	for _, cat := range section.Categories {
		if cat.Name == category {
			return cat.Total
		}
	}
	return 0
}

func (s *FinancialReportServiceImpl) getBalanceByAccountCode(section models.BalanceSheetSection, code string) (float64, error) {
	// This is a simplified implementation
	var balance float64
	err := s.db.Where("code = ? AND is_active = ?", code, true).
		Select("balance").
		First(&balance).Error
	return balance, err
}

// Helper method for direct account balance lookup by code
func (s *FinancialReportServiceImpl) getAccountBalanceByCode(ctx context.Context, code string) float64 {
	var balance float64
	s.db.WithContext(ctx).Where("code = ? AND is_active = ?", code, true).
		Select("balance").
		First(&balance)
	return balance
}

// Placeholder implementations for complex calculations

func (s *FinancialReportServiceImpl) calculateOperatingCashFlow(ctx context.Context, startDate, endDate time.Time) (models.CashFlowSection, error) {
	// Simplified implementation - would need more complex logic
	return models.CashFlowSection{
		Items: []models.CashFlowItem{
			{Description: "Net Income", Amount: 0},
			{Description: "Depreciation", Amount: 0},
			{Description: "Changes in Working Capital", Amount: 0},
		},
		Total: 0,
	}, nil
}

func (s *FinancialReportServiceImpl) calculateInvestingCashFlow(ctx context.Context, startDate, endDate time.Time) (models.CashFlowSection, error) {
	return models.CashFlowSection{
		Items: []models.CashFlowItem{
			{Description: "Purchase of Fixed Assets", Amount: 0},
			{Description: "Sale of Fixed Assets", Amount: 0},
		},
		Total: 0,
	}, nil
}

func (s *FinancialReportServiceImpl) calculateFinancingCashFlow(ctx context.Context, startDate, endDate time.Time) (models.CashFlowSection, error) {
	return models.CashFlowSection{
		Items: []models.CashFlowItem{
			{Description: "New Borrowings", Amount: 0},
			{Description: "Debt Repayments", Amount: 0},
			{Description: "Dividend Payments", Amount: 0},
		},
		Total: 0,
	}, nil
}

func (s *FinancialReportServiceImpl) getFinancialKeyMetrics(ctx context.Context, asOfDate time.Time) (*models.FinancialKeyMetrics, error) {
	// Simplified implementation
	return &models.FinancialKeyMetrics{
		TotalRevenue:       0,
		TotalExpenses:      0,
		NetIncome:          0,
		TotalAssets:        0,
		TotalLiabilities:   0,
		TotalEquity:        0,
		CashBalance:        0,
		AccountsReceivable: 0,
		AccountsPayable:    0,
		Inventory:          0,
	}, nil
}

func (s *FinancialReportServiceImpl) getCashPositionSummary(ctx context.Context) (*models.CashPositionSummary, error) {
	return &models.CashPositionSummary{
		TotalCash:     0,
		CashAccounts:  []models.CashAccount{},
		BankAccounts:  []models.BankAccount{},
		CashFlow30Day: 0,
	}, nil
}

func (s *FinancialReportServiceImpl) getAccountBalanceSummary(ctx context.Context) ([]models.AccountBalanceSummary, error) {
	return []models.AccountBalanceSummary{}, nil
}

func (s *FinancialReportServiceImpl) getRecentActivity(ctx context.Context, limit int) ([]models.RecentActivityItem, error) {
	return []models.RecentActivityItem{}, nil
}

func (s *FinancialReportServiceImpl) generateFinancialAlerts(keyMetrics *models.FinancialKeyMetrics, ratios *models.FinancialRatios) []models.FinancialAlert {
	var alerts []models.FinancialAlert
	
	// Example alerts based on thresholds
	if ratios.CurrentRatio < 1.0 {
		alerts = append(alerts, models.FinancialAlert{
			Type:        "LOW_LIQUIDITY",
			Severity:    "HIGH",
			Title:       "Low Current Ratio",
			Description: "Current ratio is below 1.0, indicating potential liquidity issues",
			Value:       ratios.CurrentRatio,
			Threshold:   1.0,
			CreatedAt:   time.Now(),
		})
	}

	return alerts
}

func (s *FinancialReportServiceImpl) getTotalByAccountTypeAndPeriod(ctx context.Context, accountType string, startDate, endDate time.Time) (float64, error) {
	var total float64
	
	err := s.db.WithContext(ctx).
		Table("journal_lines jl").
		Joins("JOIN journal_entries je ON jl.journal_entry_id = je.id").
		Joins("JOIN accounts a ON jl.account_id = a.id").
		Where("a.type = ? AND je.entry_date BETWEEN ? AND ? AND je.status = ?", 
			accountType, startDate, endDate, models.JournalStatusPosted).
		Select("SUM(jl.debit_amount - jl.credit_amount)").
		Scan(&total).Error

	return math.Abs(total), err // Return absolute value for expenses/revenue
}

func (s *FinancialReportServiceImpl) calculateLiquidityScore(ratios *models.FinancialRatios) float64 {
	score := 0.0
	
	// Current Ratio scoring (0-40 points)
	if ratios.CurrentRatio >= 2.0 {
		score += 40
	} else if ratios.CurrentRatio >= 1.5 {
		score += 30
	} else if ratios.CurrentRatio >= 1.0 {
		score += 20
	} else {
		score += 10
	}

	// Quick Ratio scoring (0-30 points)
	if ratios.QuickRatio >= 1.0 {
		score += 30
	} else if ratios.QuickRatio >= 0.75 {
		score += 20
	} else if ratios.QuickRatio >= 0.5 {
		score += 10
	}

	// Cash Ratio scoring (0-30 points)
	if ratios.CashRatio >= 0.2 {
		score += 30
	} else if ratios.CashRatio >= 0.1 {
		score += 20
	} else if ratios.CashRatio >= 0.05 {
		score += 10
	}

	return math.Min(score, 100)
}

func (s *FinancialReportServiceImpl) calculateProfitabilityScore(ratios *models.FinancialRatios) float64 {
	score := 0.0
	
	// Net Profit Margin (0-30 points)
	if ratios.NetProfitMargin >= 20 {
		score += 30
	} else if ratios.NetProfitMargin >= 10 {
		score += 20
	} else if ratios.NetProfitMargin >= 5 {
		score += 15
	} else if ratios.NetProfitMargin > 0 {
		score += 10
	}

	// ROA (0-35 points)
	if ratios.ROA >= 15 {
		score += 35
	} else if ratios.ROA >= 10 {
		score += 25
	} else if ratios.ROA >= 5 {
		score += 15
	} else if ratios.ROA > 0 {
		score += 10
	}

	// ROE (0-35 points)
	if ratios.ROE >= 20 {
		score += 35
	} else if ratios.ROE >= 15 {
		score += 25
	} else if ratios.ROE >= 10 {
		score += 15
	} else if ratios.ROE > 0 {
		score += 10
	}

	return math.Min(score, 100)
}

func (s *FinancialReportServiceImpl) calculateLeverageScore(ratios *models.FinancialRatios) float64 {
	score := 100.0 // Start with perfect score and deduct

	// Debt Ratio penalty
	if ratios.DebtRatio >= 80 {
		score -= 50
	} else if ratios.DebtRatio >= 60 {
		score -= 30
	} else if ratios.DebtRatio >= 40 {
		score -= 15
	}

	// Debt to Equity Ratio penalty
	if ratios.DebtToEquityRatio >= 2.0 {
		score -= 50
	} else if ratios.DebtToEquityRatio >= 1.5 {
		score -= 30
	} else if ratios.DebtToEquityRatio >= 1.0 {
		score -= 15
	}

	return math.Max(score, 0)
}

func (s *FinancialReportServiceImpl) calculateEfficiencyScore(ratios *models.FinancialRatios) float64 {
	score := 0.0

	// Asset Turnover (0-50 points)
	if ratios.AssetTurnover >= 2.0 {
		score += 50
	} else if ratios.AssetTurnover >= 1.5 {
		score += 40
	} else if ratios.AssetTurnover >= 1.0 {
		score += 30
	} else if ratios.AssetTurnover >= 0.5 {
		score += 20
	}

	// Inventory Turnover (0-50 points)
	if ratios.InventoryTurnover >= 12 {
		score += 50
	} else if ratios.InventoryTurnover >= 8 {
		score += 40
	} else if ratios.InventoryTurnover >= 6 {
		score += 30
	} else if ratios.InventoryTurnover >= 4 {
		score += 20
	}

	return math.Min(score, 100)
}

func (s *FinancialReportServiceImpl) getScoreGrade(score float64) string {
	if score >= 95 {
		return "A+"
	} else if score >= 90 {
		return "A"
	} else if score >= 85 {
		return "B+"
	} else if score >= 80 {
		return "B"
	} else if score >= 75 {
		return "C+"
	} else if score >= 70 {
		return "C"
	} else if score >= 60 {
		return "D"
	} else {
		return "F"
	}
}

func (s *FinancialReportServiceImpl) generateHealthRecommendations(components models.FinancialHealthComponents, ratios *models.FinancialRatios) []models.HealthRecommendation {
	var recommendations []models.HealthRecommendation

	// Liquidity recommendations
	if components.LiquidityScore < 70 {
		recommendations = append(recommendations, models.HealthRecommendation{
			Category:    "Liquidity",
			Priority:    "HIGH",
			Title:       "Improve Cash Flow Management",
			Description: "Current liquidity ratios indicate potential cash flow issues",
			Action:      "Consider improving accounts receivable collection, reducing inventory levels, or securing additional credit facilities",
		})
	}

	// Profitability recommendations
	if components.ProfitabilityScore < 70 {
		recommendations = append(recommendations, models.HealthRecommendation{
			Category:    "Profitability",
			Priority:    "HIGH",
			Title:       "Enhance Profit Margins",
			Description: "Profitability ratios are below optimal levels",
			Action:      "Review pricing strategies, reduce operating costs, or improve operational efficiency",
		})
	}

	return recommendations
}
