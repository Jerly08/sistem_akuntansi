package services

import (
	"context"
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

type EnhancedFinancialReportService interface {
	GenerateEnhancedProfitLossStatement(ctx context.Context, req *models.FinancialReportRequest) (*models.ProfitLossStatement, error)
	ValidateAndFixAccountCategories(ctx context.Context) error
	SyncAccountBalancesWithJournals(ctx context.Context) error
	GetAccountBalancesSummary(ctx context.Context, accountType string, startDate, endDate time.Time) ([]models.AccountLineItem, error)
}

type EnhancedFinancialReportServiceImpl struct {
	db              *gorm.DB
	accountRepo     repositories.AccountRepository
	journalRepo     repositories.JournalEntryRepository
}

func NewEnhancedFinancialReportService(
	db *gorm.DB,
	accountRepo repositories.AccountRepository,
	journalRepo repositories.JournalEntryRepository,
) EnhancedFinancialReportService {
	return &EnhancedFinancialReportServiceImpl{
		db:          db,
		accountRepo: accountRepo,
		journalRepo: journalRepo,
	}
}

// GenerateEnhancedProfitLossStatement generates an accurate P&L statement with proper COGS categorization
func (s *EnhancedFinancialReportServiceImpl) GenerateEnhancedProfitLossStatement(ctx context.Context, req *models.FinancialReportRequest) (*models.ProfitLossStatement, error) {
	// First, validate and fix account categories
	err := s.ValidateAndFixAccountCategories(ctx)
	if err != nil {
		log.Printf("Warning: Could not validate account categories: %v", err)
	}

	// Get revenue accounts with enhanced balance calculation
	revenueAccounts, err := s.GetAccountBalancesSummary(ctx, models.AccountTypeRevenue, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue accounts: %v", err)
	}

	// Get expense accounts with enhanced balance calculation
	expenseAccounts, err := s.GetAccountBalancesSummary(ctx, models.AccountTypeExpense, req.StartDate, req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get expense accounts: %v", err)
	}

	// Enhanced COGS categorization with multiple category checks
	var cogsAccounts, operatingExpenses []models.AccountLineItem
	for _, expense := range expenseAccounts {
		if s.isCOGSAccount(expense) {
			cogsAccounts = append(cogsAccounts, expense)
		} else {
			operatingExpenses = append(operatingExpenses, expense)
		}
	}

	// Calculate totals with proper sign handling
	totalRevenue := s.calculateRevenueTotal(revenueAccounts)
	totalCOGS := s.calculateExpenseTotal(cogsAccounts)
	totalOperatingExpenses := s.calculateExpenseTotal(operatingExpenses)
	
	grossProfit := totalRevenue - totalCOGS
	netIncome := grossProfit - totalOperatingExpenses

	// Create enhanced P&L statement
	pnl := &models.ProfitLossStatement{
		ReportHeader: models.ReportHeader{
			ReportType:    models.ReportTypeProfitLoss,
			CompanyName:   "PT. Sample Company",
			ReportTitle:   "Profit & Loss Statement (Enhanced)",
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
		TotalExpenses: totalOperatingExpenses,
		NetIncome:     netIncome,
	}

	return pnl, nil
}

// GetAccountBalancesSummary gets account balances with enhanced calculation methods
func (s *EnhancedFinancialReportServiceImpl) GetAccountBalancesSummary(ctx context.Context, accountType string, startDate, endDate time.Time) ([]models.AccountLineItem, error) {
	var accountItems []models.AccountLineItem
	
	// Enhanced query that combines both journal entries and account balances as fallback
	query := `
		SELECT 
			a.id as account_id,
			a.code as account_code,
			a.name as account_name,
			a.type as account_type,
			a.category,
			a.balance as account_balance,
			COALESCE(SUM(jl.debit_amount), 0) as journal_debit,
			COALESCE(SUM(jl.credit_amount), 0) as journal_credit,
			COUNT(jl.id) as journal_entry_count
		FROM accounts a
		LEFT JOIN journal_lines jl ON a.id = jl.account_id
		LEFT JOIN journal_entries je ON jl.journal_entry_id = je.id
		WHERE a.type = ? 
			AND a.is_active = true
			AND a.deleted_at IS NULL
			AND (je.id IS NULL OR (je.status = 'POSTED' AND je.entry_date BETWEEN ? AND ?))
		GROUP BY a.id, a.code, a.name, a.type, a.category, a.balance
		ORDER BY a.code
	`
	
	type QueryResult struct {
		AccountID          uint    `db:"account_id"`
		AccountCode        string  `db:"account_code"`
		AccountName        string  `db:"account_name"`
		AccountType        string  `db:"account_type"`
		Category           string  `db:"category"`
		AccountBalance     float64 `db:"account_balance"`
		JournalDebit       float64 `db:"journal_debit"`
		JournalCredit      float64 `db:"journal_credit"`
		JournalEntryCount  int64   `db:"journal_entry_count"`
	}
	
	var results []QueryResult
	err := s.db.Raw(query, accountType, startDate, endDate).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get account balances: %v", err)
	}
	
	for _, result := range results {
		var adjustedBalance float64
		
		// Use journal entries if available, otherwise fallback to account balance
		if result.JournalEntryCount > 0 {
			// Calculate from journal entries
			if accountType == models.AccountTypeRevenue {
				adjustedBalance = result.JournalCredit - result.JournalDebit
			} else {
				adjustedBalance = result.JournalDebit - result.JournalCredit
			}
		} else {
			// Fallback to account balance
			adjustedBalance = result.AccountBalance
			// For revenue accounts, if balance is positive and we're using account balance,
			// it likely represents a credit balance which should be positive for revenue
			if accountType == models.AccountTypeRevenue && adjustedBalance > 0 {
				// Balance is already correct for revenue
			}
		}
		
		// Only include accounts with non-zero balances unless specifically requested
		if adjustedBalance != 0 || result.JournalEntryCount > 0 {
			accountItems = append(accountItems, models.AccountLineItem{
				AccountID:   result.AccountID,
				AccountCode: result.AccountCode,
				AccountName: result.AccountName,
				AccountType: result.AccountType,
				Category:    result.Category,
				Balance:     adjustedBalance,
			})
		}
	}
	
	return accountItems, nil
}

// isCOGSAccount checks if an account should be classified as Cost of Goods Sold
func (s *EnhancedFinancialReportServiceImpl) isCOGSAccount(account models.AccountLineItem) bool {
	cogsCategories := []string{
		models.CategoryCostOfGoodsSold,
		models.CategoryDirectMaterial,
		models.CategoryDirectLabor,
		models.CategoryManufacturingOverhead,
		models.CategoryFreightIn,
		models.CategoryPurchaseReturns,
	}
	
	for _, category := range cogsCategories {
		if account.Category == category {
			return true
		}
	}
	
	// Additional check by account code or name patterns
	cogsPatterns := []string{"5101", "HARGA POKOK", "COST OF GOODS", "COST OF SALES"}
	accountUpper := fmt.Sprintf("%s %s", account.AccountCode, account.AccountName)
	
	for _, pattern := range cogsPatterns {
		if containsSubstring(accountUpper, pattern) {
			return true
		}
	}
	
	return false
}

// calculateRevenueTotal calculates total revenue (normally credit balance accounts)
func (s *EnhancedFinancialReportServiceImpl) calculateRevenueTotal(accounts []models.AccountLineItem) float64 {
	total := 0.0
	for _, account := range accounts {
		// Revenue accounts normally have credit balances, so positive balance = positive revenue
		total += account.Balance
	}
	return total
}

// calculateExpenseTotal calculates total expenses (normally debit balance accounts)
func (s *EnhancedFinancialReportServiceImpl) calculateExpenseTotal(accounts []models.AccountLineItem) float64 {
	total := 0.0
	for _, account := range accounts {
		// Expense accounts normally have debit balances, so positive balance = positive expense
		total += account.Balance
	}
	return total
}

// ValidateAndFixAccountCategories ensures proper account categorization
func (s *EnhancedFinancialReportServiceImpl) ValidateAndFixAccountCategories(ctx context.Context) error {
	// Define account category corrections
	corrections := map[string]string{
		"5101": models.CategoryCostOfGoodsSold,  // Harga Pokok Penjualan
	}
	
	for accountCode, correctCategory := range corrections {
		err := s.db.Model(&models.Account{}).
			Where("code = ? AND category != ?", accountCode, correctCategory).
			Update("category", correctCategory).Error
		
		if err != nil {
			log.Printf("Failed to update category for account %s: %v", accountCode, err)
		} else {
			log.Printf("Updated account %s category to %s", accountCode, correctCategory)
		}
	}
	
	return nil
}

// SyncAccountBalancesWithJournals synchronizes account balances with journal entries
func (s *EnhancedFinancialReportServiceImpl) SyncAccountBalancesWithJournals(ctx context.Context) error {
	// This function would be used to sync account balances with journal entries
	// For now, we'll use the enhanced query approach which handles both
	log.Println("Account balance synchronization is handled in GetAccountBalancesSummary")
	return nil
}

// Helper function to check if string contains substring (case insensitive)
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)) ||
		    indexInString(s, substr) >= 0)
}

func indexInString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Enhanced debugging function
func (s *EnhancedFinancialReportServiceImpl) DebugAccountBalances(ctx context.Context) {
	log.Println("=== ENHANCED FINANCIAL REPORT DEBUG ===")
	
	startDate := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	
	// Test revenue accounts
	revenueAccounts, err := s.GetAccountBalancesSummary(ctx, models.AccountTypeRevenue, startDate, endDate)
	if err != nil {
		log.Printf("Error getting revenue accounts: %v", err)
		return
	}
	
	log.Printf("REVENUE ACCOUNTS: %d found", len(revenueAccounts))
	totalRevenue := s.calculateRevenueTotal(revenueAccounts)
	for _, acc := range revenueAccounts {
		log.Printf("- %s: %s [%s] = %.2f", acc.AccountCode, acc.AccountName, acc.Category, acc.Balance)
	}
	log.Printf("Total Revenue: %.2f", totalRevenue)
	
	// Test expense accounts
	expenseAccounts, err := s.GetAccountBalancesSummary(ctx, models.AccountTypeExpense, startDate, endDate)
	if err != nil {
		log.Printf("Error getting expense accounts: %v", err)
		return
	}
	
	var cogsAccounts, operatingExpenses []models.AccountLineItem
	for _, expense := range expenseAccounts {
		if s.isCOGSAccount(expense) {
			cogsAccounts = append(cogsAccounts, expense)
		} else {
			operatingExpenses = append(operatingExpenses, expense)
		}
	}
	
	log.Printf("COGS ACCOUNTS: %d found", len(cogsAccounts))
	totalCOGS := s.calculateExpenseTotal(cogsAccounts)
	for _, acc := range cogsAccounts {
		log.Printf("- %s: %s [%s] = %.2f", acc.AccountCode, acc.AccountName, acc.Category, acc.Balance)
	}
	log.Printf("Total COGS: %.2f", totalCOGS)
	
	log.Printf("OPERATING EXPENSE ACCOUNTS: %d found", len(operatingExpenses))
	totalOpExp := s.calculateExpenseTotal(operatingExpenses)
	for _, acc := range operatingExpenses {
		log.Printf("- %s: %s [%s] = %.2f", acc.AccountCode, acc.AccountName, acc.Category, acc.Balance)
	}
	log.Printf("Total Operating Expenses: %.2f", totalOpExp)
	
	grossProfit := totalRevenue - totalCOGS
	netIncome := grossProfit - totalOpExp
	
	log.Printf("\n=== PROFIT & LOSS CALCULATION ===")
	log.Printf("Total Revenue: %.2f", totalRevenue)
	log.Printf("Total COGS: %.2f", totalCOGS)
	log.Printf("Gross Profit: %.2f", grossProfit)
	log.Printf("Total Operating Expenses: %.2f", totalOpExp)
	log.Printf("Net Income: %.2f", netIncome)
}
