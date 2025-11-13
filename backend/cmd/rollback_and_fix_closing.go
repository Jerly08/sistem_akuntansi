package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Account struct {
	ID      uint    `gorm:"primaryKey"`
	Code    string  `gorm:"size:20;uniqueIndex;not null"`
	Name    string  `gorm:"size:200;not null"`
	Type    string  `gorm:"size:20;not null"`
	Balance float64 `gorm:"type:decimal(20,2);default:0"`
}

type UnifiedJournalEntry struct {
	ID          uint      `gorm:"primaryKey"`
	SourceType  string    `gorm:"size:50;not null"`
	EntryDate   time.Time `gorm:"not null;index"`
	Description string    `gorm:"type:text"`
	TotalDebit  float64   `gorm:"type:decimal(20,2);not null"`
	TotalCredit float64   `gorm:"type:decimal(20,2);not null"`
}

type UnifiedJournalLine struct {
	ID           uint    `gorm:"primaryKey"`
	JournalID    uint    `gorm:"not null;index"`
	AccountID    uint64  `gorm:"not null;index"`
	LineNumber   int     `gorm:"not null"`
	Description  string  `gorm:"type:text"`
	DebitAmount  float64 `gorm:"type:decimal(20,2);default:0"`
	CreditAmount float64 `gorm:"type:decimal(20,2);default:0"`
}

type AccountingPeriod struct {
	ID          uint       `gorm:"primaryKey"`
	StartDate   time.Time  `gorm:"not null"`
	EndDate     time.Time  `gorm:"not null;index"`
	Description string     `gorm:"type:text"`
	IsClosed    bool       `gorm:"default:false;index"`
	IsLocked    bool       `gorm:"default:false"`
	ClosedBy    *uint      `gorm:"index"`
	ClosedAt    *time.Time
}

func (UnifiedJournalEntry) TableName() string {
	return "unified_journal_ledger"
}

func (UnifiedJournalLine) TableName() string {
	return "unified_journal_lines"
}

func main() {
	// Connect to database
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("🔄 Starting rollback process for periods 3 & 4...")

	// Start transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		// 1. Find incorrect closing entries
		var closingEntries []UnifiedJournalEntry
		if err := tx.Where("source_type = ? AND entry_date IN ?", "CLOSING", 
			[]string{"2027-02-02", "2027-12-31"}).Find(&closingEntries).Error; err != nil {
			return fmt.Errorf("failed to find closing entries: %v", err)
		}

		log.Printf("Found %d closing entries to rollback", len(closingEntries))
		for _, entry := range closingEntries {
			log.Printf("  - ID: %d, Date: %s, Debit: %.2f, Credit: %.2f", 
				entry.ID, entry.EntryDate.Format("2006-01-02"), 
				entry.TotalDebit, entry.TotalCredit)
		}

		if len(closingEntries) == 0 {
			log.Println("⚠️ No closing entries found. Nothing to rollback.")
			return nil
		}

		// 2. Get closing entry IDs
		closingIDs := make([]uint, len(closingEntries))
		for i, entry := range closingEntries {
			closingIDs[i] = entry.ID
		}

		// 3. Delete journal lines
		result := tx.Where("journal_id IN ?", closingIDs).Delete(&UnifiedJournalLine{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete journal lines: %v", result.Error)
		}
		log.Printf("✅ Deleted %d journal lines", result.RowsAffected)

		// 4. Delete journal entries
		result = tx.Where("id IN ?", closingIDs).Delete(&UnifiedJournalEntry{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete journal entries: %v", result.Error)
		}
		log.Printf("✅ Deleted %d journal entries", result.RowsAffected)

		// 5. Delete accounting period records
		result = tx.Where("end_date IN ? AND is_closed = ?", 
			[]string{"2027-02-02", "2027-12-31"}, true).Delete(&AccountingPeriod{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete accounting periods: %v", result.Error)
		}
		log.Printf("✅ Deleted %d accounting period records", result.RowsAffected)

		// 6. Recalculate account balances from journal lines (excluding CLOSING entries)
		log.Println("🔄 Recalculating account balances from journal lines...")

		// Reset all revenue and expense account balances to 0
		if err := tx.Model(&Account{}).Where("type IN ?", []string{"REVENUE", "EXPENSE"}).
			Update("balance", 0).Error; err != nil {
			return fmt.Errorf("failed to reset revenue/expense balances: %v", err)
		}
		log.Println("✅ Reset all REVENUE and EXPENSE account balances to 0")

		// Recalculate balances from journal lines (excluding CLOSING entries)
		type BalanceResult struct {
			AccountID    uint64
			TotalDebit   float64
			TotalCredit  float64
		}

		var balances []BalanceResult
		if err := tx.Raw(`
			SELECT 
				ujl.account_id,
				COALESCE(SUM(ujl.debit_amount), 0) as total_debit,
				COALESCE(SUM(ujl.credit_amount), 0) as total_credit
			FROM unified_journal_lines ujl
			INNER JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
			WHERE uje.status = 'POSTED'
				AND uje.source_type != 'CLOSING'
			GROUP BY ujl.account_id
		`).Scan(&balances).Error; err != nil {
			return fmt.Errorf("failed to calculate balances: %v", err)
		}

		log.Printf("Found balances for %d accounts", len(balances))

		// Update account balances based on account type
		for _, bal := range balances {
			var account Account
			if err := tx.First(&account, bal.AccountID).Error; err != nil {
				return fmt.Errorf("failed to find account %d: %v", bal.AccountID, err)
			}

			var newBalance float64
			switch account.Type {
			case "ASSET", "EXPENSE":
				// Debit normal: balance = debit - credit
				newBalance = bal.TotalDebit - bal.TotalCredit
			case "LIABILITY", "EQUITY", "REVENUE":
				// Credit normal: balance = credit - debit
				newBalance = bal.TotalCredit - bal.TotalDebit
			default:
				return fmt.Errorf("unknown account type: %s for account %s", account.Type, account.Code)
			}

			if err := tx.Model(&Account{}).Where("id = ?", account.ID).
				Update("balance", newBalance).Error; err != nil {
				return fmt.Errorf("failed to update account %s balance: %v", account.Code, err)
			}

			log.Printf("  - Account %s (%s) Type=%s: Balance=%.2f (Debit: %.2f, Credit: %.2f)",
				account.Code, account.Name, account.Type, newBalance, bal.TotalDebit, bal.TotalCredit)
		}

		log.Println("✅ Recalculated all account balances")

		return nil
	})

	if err != nil {
		log.Fatalf("❌ Rollback failed: %v", err)
	}

	log.Println("\n✅ Rollback completed successfully!")
	log.Println("\n📊 Current state:")

	// Show current account balances
	var accounts []Account
	db.Where("code IN ?", []string{"3201", "4101", "5101"}).Order("code").Find(&accounts)
	
	log.Println("\nAccount Balances:")
	for _, acc := range accounts {
		log.Printf("  - %s (%s): %.2f", acc.Code, acc.Name, acc.Balance)
	}

	// Show remaining closed periods
	var periods []AccountingPeriod
	db.Where("is_closed = ?", true).Order("end_date").Find(&periods)
	
	log.Printf("\nClosed Periods: %d\n", len(periods))
	for _, p := range periods {
		log.Printf("  - %s to %s: %s", 
			p.StartDate.Format("2006-01-02"), 
			p.EndDate.Format("2006-01-02"), 
			p.Description)
	}

	log.Println("\n🎯 Next steps:")
	log.Println("1. Restart backend server to apply the fixed period closing logic")
	log.Println("2. Use the frontend to close period 3 (2025-12-01 to 2027-02-02)")
	log.Println("3. Use the frontend to close period 4 (2027-02-03 to 2027-12-31)")
	log.Println("4. Verify Balance Sheet shows correct historical data for all dates")
}
