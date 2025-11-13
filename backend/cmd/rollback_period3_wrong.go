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

	log.Println("🔄 Rolling back incorrect period 3 closing (2027-01-01 to 2027-12-31)...")

	// Start transaction
	err = db.Transaction(func(tx *gorm.DB) error {
		// 1. Find the incorrect closing entry (2027-12-31)
		var closingEntry UnifiedJournalEntry
		if err := tx.Where("source_type = ? AND entry_date = ?", "CLOSING", 
			time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC)).First(&closingEntry).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Println("⚠️ No closing entry found for 2027-12-31")
				return nil
			}
			return fmt.Errorf("failed to find closing entry: %v", err)
		}

		log.Printf("Found incorrect closing entry: ID=%d, Date=%s, Debit=%.2f, Credit=%.2f", 
			closingEntry.ID, closingEntry.EntryDate.Format("2006-01-02"), 
			closingEntry.TotalDebit, closingEntry.TotalCredit)

		// 2. Get journal lines for this closing entry
		var lines []UnifiedJournalLine
		if err := tx.Where("journal_id = ?", closingEntry.ID).Find(&lines).Error; err != nil {
			return fmt.Errorf("failed to get journal lines: %v", err)
		}

		log.Printf("Found %d journal lines", len(lines))

		// 3. Delete journal lines
		result := tx.Where("journal_id = ?", closingEntry.ID).Delete(&UnifiedJournalLine{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete journal lines: %v", result.Error)
		}
		log.Printf("✅ Deleted %d journal lines", result.RowsAffected)

		// 4. Delete journal entry
		if err := tx.Delete(&closingEntry).Error; err != nil {
			return fmt.Errorf("failed to delete closing entry: %v", err)
		}
		log.Printf("✅ Deleted closing journal entry")

		// 5. Delete accounting period record (2027-01-01 to 2027-12-31)
		result = tx.Where("end_date = ? AND is_closed = ?", 
			time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC), true).Delete(&AccountingPeriod{})
		if result.Error != nil {
			return fmt.Errorf("failed to delete accounting period: %v", result.Error)
		}
		log.Printf("✅ Deleted %d accounting period record", result.RowsAffected)

		// 6. Recalculate account balances from journal lines (excluding CLOSING entries)
		log.Println("🔄 Recalculating account balances...")

		// Reset revenue and expense balances to 0
		if err := tx.Model(&Account{}).Where("type IN ?", []string{"REVENUE", "EXPENSE"}).
			Update("balance", 0).Error; err != nil {
			return fmt.Errorf("failed to reset revenue/expense balances: %v", err)
		}

		// Recalculate from journal lines
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

		log.Printf("Recalculating %d accounts", len(balances))

		// Update account balances
		for _, bal := range balances {
			var account Account
			if err := tx.First(&account, bal.AccountID).Error; err != nil {
				return fmt.Errorf("failed to find account %d: %v", bal.AccountID, err)
			}

			var newBalance float64
			switch account.Type {
			case "ASSET", "EXPENSE":
				newBalance = bal.TotalDebit - bal.TotalCredit
			case "LIABILITY", "EQUITY", "REVENUE":
				newBalance = bal.TotalCredit - bal.TotalDebit
			default:
				return fmt.Errorf("unknown account type: %s", account.Type)
			}

			if err := tx.Model(&Account{}).Where("id = ?", account.ID).
				Update("balance", newBalance).Error; err != nil {
				return fmt.Errorf("failed to update account %s balance: %v", account.Code, err)
			}

			if account.Code == "3201" || account.Code == "4101" || account.Code == "5101" {
				log.Printf("  - %s (%s) Type=%s: Balance=%.2f", 
					account.Code, account.Name, account.Type, newBalance)
			}
		}

		log.Println("✅ Recalculated all account balances")

		return nil
	})

	if err != nil {
		log.Fatalf("❌ Rollback failed: %v", err)
	}

	log.Println("\n✅ Rollback completed successfully!")

	// Show final state
	var accounts []Account
	db.Where("code IN ?", []string{"3201", "4101", "5101"}).Order("code").Find(&accounts)
	
	log.Println("\n📊 Current Account Balances:")
	for _, acc := range accounts {
		log.Printf("  - %s (%s): %.2f", acc.Code, acc.Name, acc.Balance)
	}

	var periods []AccountingPeriod
	db.Where("is_closed = ?", true).Order("end_date").Find(&periods)
	
	log.Printf("\nClosed Periods: %d", len(periods))
	for _, p := range periods {
		log.Printf("  - %s to %s", 
			p.StartDate.Format("2006-01-02"), 
			p.EndDate.Format("2006-01-02"))
	}

	log.Println("\n🎯 Next steps:")
	log.Println("1. Restart backend server with fixed logic")
	log.Println("2. Close periode 3: 2027-01-01 to 2027-02-02 (expected: 7M revenue, 3.5M expense)")
	log.Println("3. Close periode 4: 2027-02-03 to 2027-12-31 (expected: 7M revenue, 3.5M expense)")
	log.Println("4. Verify Balance Sheet is balanced for all dates")
}
