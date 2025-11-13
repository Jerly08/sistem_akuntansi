package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/utils"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// UnifiedPeriodClosingService handles period closing using unified journal system (SSOT)
type UnifiedPeriodClosingService struct {
	db                    *gorm.DB
	unifiedJournalService *UnifiedJournalService
	logger                *utils.JournalLogger
}

// NewUnifiedPeriodClosingService creates a new unified period closing service
func NewUnifiedPeriodClosingService(db *gorm.DB) *UnifiedPeriodClosingService {
	return &UnifiedPeriodClosingService{
		db:                    db,
		unifiedJournalService: NewUnifiedJournalService(db),
		logger:                utils.NewJournalLogger(db),
	}
}

// ExecutePeriodClosing performs period closing using unified journal system
func (s *UnifiedPeriodClosingService) ExecutePeriodClosing(ctx context.Context, startDate, endDate time.Time, description string, userID uint64) error {
	log.Printf("[UNIFIED CLOSING] Starting period closing: %s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Get retained earnings account
		var retainedEarnings models.Account
		if err := tx.Where("code = ? AND type = ?", "3201", "EQUITY").First(&retainedEarnings).Error; err != nil {
			return fmt.Errorf("retained earnings account (3201) not found: %v", err)
		}

		// 2. Get the last closed period to determine starting point
		// We should only close transactions AFTER the last closed period
		var lastClosedPeriod models.AccountingPeriod
		err := tx.Where("is_closed = ?", true).Order("end_date DESC").First(&lastClosedPeriod).Error
		var periodStartDate time.Time
		if err != nil {
			// No previous closed periods, use the start_date provided
			periodStartDate = startDate
			log.Printf("[UNIFIED CLOSING] No previous closed periods found, using provided start date: %s", startDate.Format("2006-01-02"))
		} else {
			// Start from the day after last closed period
			periodStartDate = lastClosedPeriod.EndDate.AddDate(0, 0, 1)
			log.Printf("[UNIFIED CLOSING] Last closed period end: %s, starting from: %s", 
				lastClosedPeriod.EndDate.Format("2006-01-02"), periodStartDate.Format("2006-01-02"))
			
			// Validate that startDate matches periodStartDate
			if !startDate.Equal(periodStartDate) {
				log.Printf("[UNIFIED CLOSING] ⚠️ WARNING: Provided startDate (%s) doesn't match expected (%s)", 
					startDate.Format("2006-01-02"), periodStartDate.Format("2006-01-02"))
			}
		}

		// 3. Get all revenue accounts (not just non-zero balances)
		var revenueAccounts []models.Account
		if err := tx.Where("type = ? AND is_header = false", "REVENUE").
			Find(&revenueAccounts).Error; err != nil {
			return fmt.Errorf("failed to get revenue accounts: %v", err)
		}

		// 4. Get all expense accounts (not just non-zero balances)
		var expenseAccounts []models.Account
		if err := tx.Where("type = ? AND is_header = false", "EXPENSE").
			Find(&expenseAccounts).Error; err != nil {
			return fmt.Errorf("failed to get expense accounts: %v", err)
		}

		// 5. Calculate CUMULATIVE balances from fiscal year start to closing date
		// Query all journal lines for Revenue/Expense accounts from fiscal year start to end date
		// This ensures we capture ALL transactions, not just current period
		type AccountBalance struct {
			AccountID uint64
			TotalDebit  float64
			TotalCredit float64
		}

		// Get cumulative balances for revenue accounts
		revenueAccountIDs := make([]uint64, 0, len(revenueAccounts))
		for _, acc := range revenueAccounts {
			revenueAccountIDs = append(revenueAccountIDs, uint64(acc.ID))
		}

		expenseAccountIDs := make([]uint64, 0, len(expenseAccounts))
		for _, acc := range expenseAccounts {
			expenseAccountIDs = append(expenseAccountIDs, uint64(acc.ID))
		}

		// IMPORTANT: Calculate CUMULATIVE balances (ALL TIME, exclude closing entries)
		// This ensures we close the TOTAL accumulated balance, not just current period
		var revenueBalances []AccountBalance
		if len(revenueAccountIDs) > 0 {
			if err := tx.Raw(`
				SELECT 
					ujl.account_id,
					COALESCE(SUM(ujl.debit_amount), 0) as total_debit,
					COALESCE(SUM(ujl.credit_amount), 0) as total_credit
				FROM unified_journal_lines ujl
				INNER JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
				WHERE ujl.account_id IN ?
					AND uje.entry_date <= ?
					AND uje.status = 'POSTED'
					AND uje.source_type != 'CLOSING'
				GROUP BY ujl.account_id
			`, revenueAccountIDs, endDate).Scan(&revenueBalances).Error; err != nil {
				return fmt.Errorf("failed to calculate cumulative revenue balances: %v", err)
			}
		}

	var expenseBalances []AccountBalance
		if len(expenseAccountIDs) > 0 {
			if err := tx.Raw(`
				SELECT 
					ujl.account_id,
					COALESCE(SUM(ujl.debit_amount), 0) as total_debit,
					COALESCE(SUM(ujl.credit_amount), 0) as total_credit
				FROM unified_journal_lines ujl
				INNER JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
				WHERE ujl.account_id IN ?
					AND uje.entry_date <= ?
					AND uje.status = 'POSTED'
					AND uje.source_type != 'CLOSING'
				GROUP BY ujl.account_id
			`, expenseAccountIDs, endDate).Scan(&expenseBalances).Error; err != nil {
				return fmt.Errorf("failed to calculate cumulative expense balances: %v", err)
			}
		}

		// 6. Build map of account balances
		revenueBalanceMap := make(map[uint64]decimal.Decimal)
		for _, bal := range revenueBalances {
			// Revenue balance calculation: Debit - Credit (will be negative for credit balance)
			// For closing, we need the ABSOLUTE value to debit revenue
			netBalance := decimal.NewFromFloat(bal.TotalDebit).Sub(decimal.NewFromFloat(bal.TotalCredit))
			// Take absolute value - if revenue has credit balance (negative), we need positive amount to close
			absBalance := netBalance.Abs()
			if absBalance.GreaterThan(decimal.NewFromFloat(0.01)) {
				revenueBalanceMap[bal.AccountID] = absBalance
			}
		}

		expenseBalanceMap := make(map[uint64]decimal.Decimal)
		for _, bal := range expenseBalances {
			// Expense: debit increases, credit decreases
			// Net balance = Debit - Credit (should be positive for expense)
			netBalance := decimal.NewFromFloat(bal.TotalDebit).Sub(decimal.NewFromFloat(bal.TotalCredit))
			if netBalance.GreaterThan(decimal.NewFromFloat(0.01)) {
				expenseBalanceMap[bal.AccountID] = netBalance
			}
		}

		if len(revenueBalanceMap) == 0 && len(expenseBalanceMap) == 0 {
			log.Println("[UNIFIED CLOSING] No revenue or expense balances to close")
			return nil
		}

		// 7. Calculate totals
		var totalRevenue, totalExpense decimal.Decimal
		for _, balance := range revenueBalanceMap {
			totalRevenue = totalRevenue.Add(balance)
		}
		for _, balance := range expenseBalanceMap {
			totalExpense = totalExpense.Add(balance)
		}

		netIncome := totalRevenue.Sub(totalExpense)

		log.Printf("[UNIFIED CLOSING] Period Start Date: %s", periodStartDate.Format("2006-01-02"))
		log.Printf("[UNIFIED CLOSING] Period End Date (Closing): %s", endDate.Format("2006-01-02"))
		log.Printf("[UNIFIED CLOSING] Period Revenue: %.2f, Period Expense: %.2f, Net Income: %.2f",
			totalRevenue.InexactFloat64(), totalExpense.InexactFloat64(), netIncome.InexactFloat64())

		// 8. Create unified journal entry for closing
		var journalLines []models.SSOTJournalLine
		lineNum := 1

		// Close Revenue accounts (Debit Revenue, Credit Retained Earnings)
		// Revenue accounts have CREDIT balances (credit > debit)
		// To close them: Debit Revenue to zero out the cumulative credit balance
		for _, acc := range revenueAccounts {
			balance, exists := revenueBalanceMap[uint64(acc.ID)]
			if exists && balance.GreaterThan(decimal.NewFromFloat(0.01)) {
				journalLines = append(journalLines, models.SSOTJournalLine{
					AccountID:    uint64(acc.ID),
					LineNumber:   lineNum,
					Description:  fmt.Sprintf("Close period revenue: %s", acc.Name),
					DebitAmount:  balance,
					CreditAmount: decimal.Zero,
				})
				lineNum++
				log.Printf("[UNIFIED CLOSING] Closing Revenue: %s (ID: %d) Balance: %.2f", acc.Name, acc.ID, balance.InexactFloat64())
			}
		}

		// Credit Retained Earnings with total revenue
		if totalRevenue.GreaterThan(decimal.Zero) {
			journalLines = append(journalLines, models.SSOTJournalLine{
				AccountID:    uint64(retainedEarnings.ID),
				LineNumber:   lineNum,
				Description:  "Transfer revenue to retained earnings",
				DebitAmount:  decimal.Zero,
				CreditAmount: totalRevenue,
			})
			lineNum++
		}

		// Debit Retained Earnings with total expense
		if totalExpense.GreaterThan(decimal.Zero) {
			journalLines = append(journalLines, models.SSOTJournalLine{
				AccountID:    uint64(retainedEarnings.ID),
				LineNumber:   lineNum,
				Description:  "Transfer expense from retained earnings",
				DebitAmount:  totalExpense,
				CreditAmount: decimal.Zero,
			})
			lineNum++
		}

		// Close Expense accounts (Credit Expense to zero out the cumulative debit balance)
		// Expense accounts have DEBIT balances (debit > credit)
		// To close them: Credit Expense to zero out the cumulative debit balance
		for _, acc := range expenseAccounts {
			balance, exists := expenseBalanceMap[uint64(acc.ID)]
			if exists && balance.GreaterThan(decimal.NewFromFloat(0.01)) {
				journalLines = append(journalLines, models.SSOTJournalLine{
					AccountID:    uint64(acc.ID),
					LineNumber:   lineNum,
					Description:  fmt.Sprintf("Close period expense: %s", acc.Name),
					DebitAmount:  decimal.Zero,
					CreditAmount: balance,
				})
				lineNum++
				log.Printf("[UNIFIED CLOSING] Closing Expense: %s (ID: %d) Balance: %.2f", acc.Name, acc.ID, balance.InexactFloat64())
			}
		}

		// 9. Create unified journal entry
		closingEntry := &models.SSOTJournalEntry{
			SourceType:      "CLOSING",
			EntryDate:       endDate,
			Description:     description,
			TotalDebit:      totalRevenue.Add(totalExpense),
			TotalCredit:     totalRevenue.Add(totalExpense),
			Status:          "POSTED",
			IsBalanced:      true,
			IsAutoGenerated: true,
			CreatedBy:       userID,
			Lines:           journalLines,
		}

		now := time.Now()
		closingEntry.PostedAt = &now
		closingEntry.PostedBy = &userID

		// Create journal entry in unified system (this will automatically update balances)
		if err := tx.Create(closingEntry).Error; err != nil {
			return fmt.Errorf("failed to create unified closing journal: %v", err)
		}

		log.Printf("[UNIFIED CLOSING] Created unified journal entry ID: %d with %d lines", closingEntry.ID, len(journalLines))

		// 7. Robust balance update: recalculate exact balances for affected accounts from SSOT lines
		// Rationale: DB triggers are disabled to avoid double-posting; manual per-line +/- is error-prone due to sign conventions.
		// We recalc the true balance for each affected account and set it explicitly.
		affectedIDsMap := make(map[uint64]struct{})
		for _, line := range journalLines {
			affectedIDsMap[line.AccountID] = struct{}{}
		}
		var affectedIDs []uint64
		for id := range affectedIDsMap {
			affectedIDs = append(affectedIDs, id)
		}

		if len(affectedIDs) > 0 {
			// Query correct balances from unified journals (POSTED only)
			type BalanceRow struct {
				AccountID uint64
				AccType   string
				Correct   float64
			}
		var rows []BalanceRow
		if err := tx.Raw(`
			SELECT a.id as account_id, a.type as acc_type,
				-- For ALL account types: Debit - Credit
				-- ASSET/EXPENSE will be positive (debit balance)
				-- LIABILITY/EQUITY/REVENUE will be negative (credit balance)
				COALESCE(SUM(ujl.debit_amount),0) - COALESCE(SUM(ujl.credit_amount),0) AS correct
			FROM accounts a
			LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
			LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
			WHERE a.id IN ? AND a.deleted_at IS NULL
				AND (uje.id IS NULL OR (uje.status = 'POSTED' AND uje.entry_date <= ?))
			GROUP BY a.id, a.type
		`, affectedIDs, endDate).Scan(&rows).Error; err != nil {
			return fmt.Errorf("failed to recalc balances for affected accounts: %v", err)
		}

			for _, r := range rows {
				if err := tx.Model(&models.Account{}).
					Where("id = ?", r.AccountID).
					Update("balance", r.Correct).Error; err != nil {
					return fmt.Errorf("failed to set recalculated balance for account %d: %v", r.AccountID, err)
				}
				log.Printf("[UNIFIED CLOSING] Recalculated balance set: account_id=%d type=%s new_balance=%.2f", r.AccountID, r.AccType, r.Correct)
			}
			
			// VALIDATION: Verify that all Revenue and Expense accounts are now ZERO
			var nonZeroCount int64
			if err := tx.Model(&models.Account{}).
				Where("type IN (?) AND ABS(balance) > 0.01", []string{"REVENUE", "EXPENSE"}).
				Count(&nonZeroCount).Error; err != nil {
				return fmt.Errorf("failed to validate zero balances: %v", err)
			}
			
			if nonZeroCount > 0 {
				// Log which accounts are not zero for debugging
				var problemAccounts []models.Account
				tx.Where("type IN (?) AND ABS(balance) > 0.01", []string{"REVENUE", "EXPENSE"}).
					Find(&problemAccounts)
				for _, acc := range problemAccounts {
					log.Printf("[UNIFIED CLOSING] ⚠️ WARNING: Account %s (%s) still has balance: %.2f", 
						acc.Code, acc.Name, acc.Balance)
				}
				return fmt.Errorf("closing validation failed: %d revenue/expense accounts still have non-zero balance", nonZeroCount)
			}
			
			log.Printf("[UNIFIED CLOSING] ✅ Validation passed: All Revenue and Expense accounts are ZERO")
		}

		// 8. Create accounting period record
		userIDUint := uint(userID)
		accountingPeriod := models.AccountingPeriod{
			StartDate:    startDate,
			EndDate:      endDate,
			Description:  description,
			IsClosed:     true,
			IsLocked:     true,
			ClosedBy:     &userIDUint,
			ClosedAt:     &now,
			TotalRevenue: totalRevenue.InexactFloat64(),
			TotalExpense: totalExpense.InexactFloat64(),
			NetIncome:    netIncome.InexactFloat64(),
		}

		if err := tx.Create(&accountingPeriod).Error; err != nil {
			return fmt.Errorf("failed to create accounting period: %v", err)
		}

		log.Printf("[UNIFIED CLOSING] ✅ Period closing completed successfully")
		log.Printf("[UNIFIED CLOSING] Net Income: %.2f transferred to Retained Earnings", netIncome.InexactFloat64())

		return nil
	})
}

// PreviewPeriodClosing generates preview of period closing
func (s *UnifiedPeriodClosingService) PreviewPeriodClosing(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error) {
	// Get retained earnings account
	var retainedEarnings models.Account
	if err := s.db.Where("code = ? AND type = ?", "3201", "EQUITY").First(&retainedEarnings).Error; err != nil {
		return nil, fmt.Errorf("retained earnings account (3201) not found: %v", err)
	}

	// Get revenue and expense accounts
	var revenueAccounts []models.Account
	s.db.Where("type = ? AND ABS(balance) > 0.01 AND is_header = false", "REVENUE").Find(&revenueAccounts)

	var expenseAccounts []models.Account
	s.db.Where("type = ? AND ABS(balance) > 0.01 AND is_header = false", "EXPENSE").Find(&expenseAccounts)

	// Calculate totals
	// Use absolute values to ensure correct positive amounts
	var totalRevenue, totalExpense float64
	for _, acc := range revenueAccounts {
		// Revenue balance is stored as positive, use absolute value
		if acc.Balance < 0 {
			totalRevenue += -acc.Balance
		} else {
			totalRevenue += acc.Balance
		}
	}
	for _, acc := range expenseAccounts {
		// Expense balance - use absolute value
		if acc.Balance < 0 {
			totalExpense += -acc.Balance
		} else {
			totalExpense += acc.Balance
		}
	}

	netIncome := totalRevenue - totalExpense

	return map[string]interface{}{
		"start_date":        startDate.Format("2006-01-02"),
		"end_date":          endDate.Format("2006-01-02"),
		"total_revenue":     totalRevenue,
		"total_expense":     totalExpense,
		"net_income":        netIncome,
		"revenue_accounts":  len(revenueAccounts),
		"expense_accounts":  len(expenseAccounts),
		"can_close":         len(revenueAccounts) > 0 || len(expenseAccounts) > 0,
		"retained_earnings": retainedEarnings.Name,
	}, nil
}

// GetLastClosingInfo returns information about the last closed period
func (s *UnifiedPeriodClosingService) GetLastClosingInfo(ctx context.Context) (map[string]interface{}, error) {
	var lastPeriod models.AccountingPeriod
	err := s.db.Where("is_closed = ?", true).
		Order("end_date DESC").
		First(&lastPeriod).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return map[string]interface{}{
				"has_previous_closing": false,
			}, nil
		}
		return nil, fmt.Errorf("failed to query last closing period: %v", err)
	}

	nextStart := lastPeriod.EndDate.AddDate(0, 0, 1)

	return map[string]interface{}{
		"has_previous_closing": true,
		"last_closing_date":    lastPeriod.EndDate.Format("2006-01-02"),
		"next_start_date":      nextStart.Format("2006-01-02"),
		"last_net_income":      lastPeriod.NetIncome,
	}, nil
}

// IsDateInClosedPeriod checks if a given date falls within a closed period
func (s *UnifiedPeriodClosingService) IsDateInClosedPeriod(ctx context.Context, date time.Time) (bool, error) {
	var count int64
	err := s.db.Model(&models.AccountingPeriod{}).
		Where("is_closed = ? AND ? BETWEEN start_date AND end_date", true, date).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check closed period: %v", err)
	}

	return count > 0, nil
}

// GetPeriodInfoForDate returns period information for a specific date
func (s *UnifiedPeriodClosingService) GetPeriodInfoForDate(ctx context.Context, date time.Time) map[string]interface{} {
	var period models.AccountingPeriod
	err := s.db.Preload("ClosedByUser").Where("? BETWEEN start_date AND end_date AND is_closed = ?", date, true).
		First(&period).Error
	
	if err != nil {
		return nil
	}
	
	info := map[string]interface{}{
		"start_date":  period.StartDate,
		"end_date":    period.EndDate,
		"description": period.Description,
		"is_locked":   period.IsLocked,
	}
	
	if period.ClosedBy != nil {
		info["closed_by"] = *period.ClosedBy
	}
	
	if period.ClosedByUser.ID != 0 {
		info["closed_by_name"] = period.ClosedByUser.GetDisplayName()
	}
	
	return info
}
