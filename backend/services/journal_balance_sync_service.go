package services

import (
	"context"
	"fmt"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

// JournalBalanceSyncService ensures all account balances are synchronized with journal entries
type JournalBalanceSyncService struct {
	db          *gorm.DB
	accountRepo repositories.AccountRepository
}

func NewJournalBalanceSyncService(db *gorm.DB, accountRepo repositories.AccountRepository) *JournalBalanceSyncService {
	return &JournalBalanceSyncService{
		db:          db,
		accountRepo: accountRepo,
	}
}

// CalculateAccountBalanceFromJournals calculates account balance from journal entries
func (jbs *JournalBalanceSyncService) CalculateAccountBalanceFromJournals(accountID uint) (float64, error) {
	var account models.Account
	if err := jbs.db.First(&account, accountID).Error; err != nil {
		return 0, fmt.Errorf("account not found: %v", err)
	}

	// Get all journal lines for this account from posted journal entries
	var totalDebits, totalCredits float64
	err := jbs.db.Table("journal_lines").
		Joins("JOIN journal_entries ON journal_lines.journal_entry_id = journal_entries.id").
		Where("journal_lines.account_id = ? AND journal_entries.status = ?", 
			accountID, models.JournalStatusPosted).
		Select("COALESCE(SUM(journal_lines.debit_amount), 0) as total_debits, COALESCE(SUM(journal_lines.credit_amount), 0) as total_credits").
		Row().Scan(&totalDebits, &totalCredits)
	
	if err != nil {
		return 0, fmt.Errorf("failed to calculate journal totals: %v", err)
	}

	// Calculate balance based on account type (normal balance)
	switch account.Type {
	case models.AccountTypeAsset, models.AccountTypeExpense:
		// Debit normal balance: Debits increase, Credits decrease
		return totalDebits - totalCredits, nil
	case models.AccountTypeLiability, models.AccountTypeEquity, models.AccountTypeRevenue:
		// Credit normal balance: Credits increase, Debits decrease
		return totalCredits - totalDebits, nil
	default:
		return 0, fmt.Errorf("unknown account type: %s", account.Type)
	}
}

// SyncAllAccountBalances synchronizes all account balances with journal entries
func (jbs *JournalBalanceSyncService) SyncAllAccountBalances() error {
	ctx := context.Background()
	accounts, err := jbs.accountRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get accounts: %v", err)
	}

	var updated, errors int
	for _, account := range accounts {
		if !account.IsActive {
			continue
		}

		// Calculate balance from journals
		journalBalance, err := jbs.CalculateAccountBalanceFromJournals(account.ID)
		if err != nil {
			fmt.Printf("❌ Error calculating balance for %s: %v\n", account.Code, err)
			errors++
			continue
		}

		// Update if different
		if account.Balance != journalBalance {
			oldBalance := account.Balance
			err = jbs.db.Model(&account).Update("balance", journalBalance).Error
			if err != nil {
				fmt.Printf("❌ Error updating balance for %s: %v\n", account.Code, err)
				errors++
				continue
			}
			fmt.Printf("✅ Updated %s (%s): %.2f → %.2f\n", 
				account.Code, account.Name, oldBalance, journalBalance)
			updated++
		}
	}

	fmt.Printf("\n📊 Sync Summary: %d updated, %d errors\n", updated, errors)
	return nil
}

// ValidateAccountBalance checks if account balance matches journal entries
func (jbs *JournalBalanceSyncService) ValidateAccountBalance(accountID uint) (*BalanceValidationResult, error) {
	var account models.Account
	if err := jbs.db.First(&account, accountID).Error; err != nil {
		return nil, fmt.Errorf("account not found: %v", err)
	}

	journalBalance, err := jbs.CalculateAccountBalanceFromJournals(accountID)
	if err != nil {
		return nil, err
	}

	result := &BalanceValidationResult{
		AccountID:       accountID,
		AccountCode:     account.Code,
		AccountName:     account.Name,
		CurrentBalance:  account.Balance,
		JournalBalance:  journalBalance,
		Difference:      account.Balance - journalBalance,
		IsConsistent:    account.Balance == journalBalance,
		LastUpdated:     time.Now(),
	}

	return result, nil
}

// ValidateAllBalances validates all account balances
func (jbs *JournalBalanceSyncService) ValidateAllBalances() ([]*BalanceValidationResult, error) {
	ctx := context.Background()
	accounts, err := jbs.accountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %v", err)
	}

	var results []*BalanceValidationResult
	for _, account := range accounts {
		if !account.IsActive {
			continue
		}

		result, err := jbs.ValidateAccountBalance(account.ID)
		if err != nil {
			// Create error result
			result = &BalanceValidationResult{
				AccountID:    account.ID,
				AccountCode:  account.Code,
				AccountName:  account.Name,
				IsConsistent: false,
				Error:        err.Error(),
			}
		}
		results = append(results, result)
	}

	return results, nil
}

// BalanceValidationResult represents balance validation result
type BalanceValidationResult struct {
	AccountID       uint      `json:"account_id"`
	AccountCode     string    `json:"account_code"`
	AccountName     string    `json:"account_name"`
	CurrentBalance  float64   `json:"current_balance"`
	JournalBalance  float64   `json:"journal_balance"`
	Difference      float64   `json:"difference"`
	IsConsistent    bool      `json:"is_consistent"`
	LastUpdated     time.Time `json:"last_updated"`
	Error           string    `json:"error,omitempty"`
}

// GetInconsistentAccounts returns accounts with balance inconsistencies
func (jbs *JournalBalanceSyncService) GetInconsistentAccounts() ([]*BalanceValidationResult, error) {
	allResults, err := jbs.ValidateAllBalances()
	if err != nil {
		return nil, err
	}

	var inconsistent []*BalanceValidationResult
	for _, result := range allResults {
		if !result.IsConsistent {
			inconsistent = append(inconsistent, result)
		}
	}

	return inconsistent, nil
}

// CreateMissingJournalEntries creates journal entries for accounts with balances but no journals
// WARNING: This should only be used for data migration/correction purposes
func (jbs *JournalBalanceSyncService) CreateMissingJournalEntries(description string) error {
	inconsistent, err := jbs.GetInconsistentAccounts()
	if err != nil {
		return err
	}

	// Create opening balance journal entry
	journalEntry := &models.JournalEntry{
		EntryDate:       time.Now().AddDate(0, 0, -1), // Yesterday
		Reference:       "OPENING-BALANCE-" + time.Now().Format("20060102"),
		Description:     description,
		Status:          models.JournalStatusPosted,
		UserID:          1, // System user
		TotalDebit:      0,
		TotalCredit:     0,
		IsAutoGenerated: true,
		ReferenceType:   models.JournalRefOpening,
	}

	if err := jbs.db.Create(journalEntry).Error; err != nil {
		return fmt.Errorf("failed to create opening balance journal entry: %v", err)
	}

	var totalDebit, totalCredit float64
	for _, result := range inconsistent {
		if result.CurrentBalance == 0 {
			continue
		}

		var account models.Account
		if err := jbs.db.First(&account, result.AccountID).Error; err != nil {
			continue
		}

		// Create journal line for this account
		line := &models.JournalLine{
			JournalEntryID: journalEntry.ID,
			AccountID:      account.ID,
			Description:    fmt.Sprintf("Opening balance for %s", account.Name),
			LineNumber:     1,
		}

		switch account.Type {
		case models.AccountTypeAsset, models.AccountTypeExpense:
			if result.CurrentBalance > 0 {
				line.DebitAmount = result.CurrentBalance
				totalDebit += result.CurrentBalance
			} else {
				line.CreditAmount = -result.CurrentBalance
				totalCredit += -result.CurrentBalance
			}
		case models.AccountTypeLiability, models.AccountTypeEquity, models.AccountTypeRevenue:
			if result.CurrentBalance > 0 {
				line.CreditAmount = result.CurrentBalance
				totalCredit += result.CurrentBalance
			} else {
				line.DebitAmount = -result.CurrentBalance
				totalDebit += -result.CurrentBalance
			}
		}

		if err := jbs.db.Create(line).Error; err != nil {
			fmt.Printf("❌ Failed to create line for %s: %v\n", account.Code, err)
		} else {
			fmt.Printf("✅ Created line for %s: %.2f\n", account.Code, result.CurrentBalance)
		}
	}

	// Update journal entry totals
	journalEntry.TotalDebit = totalDebit
	journalEntry.TotalCredit = totalCredit
	journalEntry.IsBalanced = (totalDebit == totalCredit)
	jbs.db.Save(journalEntry)

	fmt.Printf("\n📝 Created opening balance journal with total debit: %.2f, credit: %.2f\n", 
		totalDebit, totalCredit)

	return nil
}
