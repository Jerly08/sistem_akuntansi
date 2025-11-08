package main

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Account struct {
	ID      uint    `gorm:"primaryKey"`
	Code    string  `gorm:"size:50;not null;unique"`
	Name    string  `gorm:"size:255;not null"`
	Type    string  `gorm:"size:50;not null"`
	Balance float64 `gorm:"type:decimal(15,2);default:0"`
}

type SSOTJournalEntry struct {
	ID          uint      `gorm:"primaryKey"`
	SourceType  string    `gorm:"size:50;not null"`
	EntryDate   time.Time `gorm:"not null;index"`
	Description string    `gorm:"type:text"`
	TotalDebit  float64   `gorm:"type:decimal(15,2);not null"`
	TotalCredit float64   `gorm:"type:decimal(15,2);not null"`
	Status      string    `gorm:"size:20;not null;default:'POSTED'"`
	IsBalanced  bool      `gorm:"default:true"`
	CreatedAt   time.Time
}

type SSOTJournalLine struct {
	ID              uint    `gorm:"primaryKey"`
	JournalEntryID  uint    `gorm:"not null;index"`
	AccountID       uint64  `gorm:"not null;index"`
	LineNumber      int     `gorm:"not null"`
	Description     string  `gorm:"type:text"`
	DebitAmount     float64 `gorm:"type:decimal(15,2);not null"`
	CreditAmount    float64 `gorm:"type:decimal(15,2);not null"`
	Account         Account `gorm:"foreignKey:AccountID"`
}

type AccountingPeriod struct {
	ID           uint      `gorm:"primaryKey"`
	StartDate    time.Time `gorm:"not null;index"`
	EndDate      time.Time `gorm:"not null;index"`
	Description  string    `gorm:"type:text"`
	IsClosed     bool      `gorm:"default:false"`
	TotalRevenue float64   `gorm:"type:decimal(15,2)"`
	TotalExpense float64   `gorm:"type:decimal(15,2)"`
	NetIncome    float64   `gorm:"type:decimal(15,2)"`
	ClosedAt     *time.Time
	CreatedAt    time.Time
}

func (SSOTJournalEntry) TableName() string {
	return "ssot_journal_entries"
}

func (SSOTJournalLine) TableName() string {
	return "ssot_journal_lines"
}

func (Account) TableName() string {
	return "accounts"
}

func (AccountingPeriod) TableName() string {
	return "accounting_periods"
}

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("========================================")
	fmt.Println("DIAGNOSA MASALAH CLOSING BALANCE SHEET")
	fmt.Println("========================================\n")

	// 1. Check Balance Sheet
	fmt.Println("1. BALANCE SHEET STATUS")
	fmt.Println("----------------------------------------")
	
	type BalanceResult struct {
		Type    string
		Balance float64
	}
	
	var results []BalanceResult
	db.Raw(`
		SELECT type, SUM(balance) as balance
		FROM accounts
		WHERE is_active = true AND COALESCE(is_header, false) = false
		GROUP BY type
		ORDER BY type
	`).Scan(&results)
	
	var assets, liabilities, equity float64
	for _, r := range results {
		fmt.Printf("%-15s: Rp %15.2f\n", r.Type, r.Balance)
		switch r.Type {
		case "ASSET":
			assets = r.Balance
		case "LIABILITY":
			liabilities = r.Balance
		case "EQUITY":
			equity = r.Balance
		}
	}
	
	diff := assets - (liabilities + equity)
	fmt.Println("----------------------------------------")
	fmt.Printf("Total Assets   : Rp %15.2f\n", assets)
	fmt.Printf("Total Liab+Eq  : Rp %15.2f\n", liabilities+equity)
	fmt.Printf("DIFFERENCE     : Rp %15.2f ", diff)
	if diff > 0.01 || diff < -0.01 {
		fmt.Println("❌ NOT BALANCED")
	} else {
		fmt.Println("✓ BALANCED")
	}
	fmt.Println()

	// 2. Check SSOT Journal Entries
	fmt.Println("2. SSOT JOURNAL ENTRIES (CLOSING)")
	fmt.Println("----------------------------------------")
	
	var closingJournals []SSOTJournalEntry
	db.Where("source_type = ?", "CLOSING").
		Order("entry_date DESC").
		Find(&closingJournals)
	
	fmt.Printf("Total Closing Entries: %d\n\n", len(closingJournals))
	
	for _, je := range closingJournals {
		fmt.Printf("ID: %d | Date: %s | Status: %s\n", 
			je.ID, je.EntryDate.Format("2006-01-02"), je.Status)
		fmt.Printf("Description: %s\n", je.Description)
		fmt.Printf("Total Debit : Rp %15.2f\n", je.TotalDebit)
		fmt.Printf("Total Credit: Rp %15.2f\n", je.TotalCredit)
		fmt.Printf("Is Balanced : %v\n", je.IsBalanced)
		
		// Get lines for this journal
		var lines []SSOTJournalLine
		db.Preload("Account").
			Where("journal_entry_id = ?", je.ID).
			Order("line_number").
			Find(&lines)
		
		fmt.Printf("\nJournal Lines (%d):\n", len(lines))
		for _, line := range lines {
			fmt.Printf("  %d. [%s] %s\n", 
				line.LineNumber, line.Account.Code, line.Account.Name)
			if line.DebitAmount > 0 {
				fmt.Printf("     Debit : Rp %15.2f\n", line.DebitAmount)
			}
			if line.CreditAmount > 0 {
				fmt.Printf("     Credit: Rp %15.2f\n", line.CreditAmount)
			}
		}
		fmt.Println("----------------------------------------")
	}
	fmt.Println()

	// 3. Check Accounting Periods
	fmt.Println("3. ACCOUNTING PERIODS")
	fmt.Println("----------------------------------------")
	
	var periods []AccountingPeriod
	db.Where("is_closed = true").
		Order("end_date DESC").
		Find(&periods)
	
	fmt.Printf("Total Closed Periods: %d\n\n", len(periods))
	
	for _, p := range periods {
		fmt.Printf("Period: %s to %s\n", 
			p.StartDate.Format("2006-01-02"), p.EndDate.Format("2006-01-02"))
		fmt.Printf("Total Revenue: Rp %15.2f\n", p.TotalRevenue)
		fmt.Printf("Total Expense: Rp %15.2f\n", p.TotalExpense)
		fmt.Printf("Net Income   : Rp %15.2f\n", p.NetIncome)
		if p.ClosedAt != nil {
			fmt.Printf("Closed At    : %s\n", p.ClosedAt.Format("2006-01-02 15:04:05"))
		}
		fmt.Println("----------------------------------------")
	}
	fmt.Println()

	// 4. Check Revenue & Expense Account Balances
	fmt.Println("4. REVENUE & EXPENSE ACCOUNT BALANCES")
	fmt.Println("----------------------------------------")
	
	var tempAccounts []Account
	db.Raw(`
		SELECT code, name, type, balance
		FROM accounts
		WHERE type IN ('REVENUE', 'EXPENSE')
		  AND is_active = true
		  AND COALESCE(is_header, false) = false
		ORDER BY type, code
	`).Scan(&tempAccounts)
	
	var totalRevBalance, totalExpBalance float64
	
	fmt.Println("REVENUE Accounts:")
	for _, acc := range tempAccounts {
		if acc.Type == "REVENUE" {
			fmt.Printf("  [%s] %-30s: Rp %15.2f\n", acc.Code, acc.Name, acc.Balance)
			totalRevBalance += acc.Balance
		}
	}
	fmt.Printf("  Total Revenue Balance: Rp %15.2f\n\n", totalRevBalance)
	
	fmt.Println("EXPENSE Accounts:")
	for _, acc := range tempAccounts {
		if acc.Type == "EXPENSE" {
			fmt.Printf("  [%s] %-30s: Rp %15.2f\n", acc.Code, acc.Name, acc.Balance)
			totalExpBalance += acc.Balance
		}
	}
	fmt.Printf("  Total Expense Balance: Rp %15.2f\n", totalExpBalance)
	
	if totalRevBalance == 0 && totalExpBalance == 0 {
		fmt.Println("\n✓ All temporary accounts are closed (balance = 0)")
	} else {
		fmt.Println("\n❌ WARNING: Temporary accounts still have balances!")
		fmt.Println("This should be 0 after period closing.")
	}
	fmt.Println("----------------------------------------")
	fmt.Println()

	// 5. Check Retained Earnings
	fmt.Println("5. RETAINED EARNINGS (3201)")
	fmt.Println("----------------------------------------")
	
	var retainedEarnings Account
	err = db.Where("code = ?", "3201").First(&retainedEarnings).Error
	if err != nil {
		fmt.Println("❌ Retained Earnings account not found!")
	} else {
		fmt.Printf("Account: [%s] %s\n", retainedEarnings.Code, retainedEarnings.Name)
		fmt.Printf("Balance: Rp %15.2f\n", retainedEarnings.Balance)
		
		// Calculate from journal lines
		var lines []SSOTJournalLine
		db.Raw(`
			SELECT sjl.*
			FROM ssot_journal_lines sjl
			JOIN ssot_journal_entries sje ON sjl.journal_entry_id = sje.id
			WHERE sjl.account_id = ?
			  AND sje.status = 'POSTED'
			ORDER BY sje.entry_date
		`, retainedEarnings.ID).Scan(&lines)
		
		fmt.Printf("\nJournal Lines: %d\n", len(lines))
		
		var totalDebit, totalCredit float64
		for _, line := range lines {
			totalDebit += line.DebitAmount
			totalCredit += line.CreditAmount
		}
		
		// Retained Earnings is EQUITY (credit normal)
		calculatedBalance := totalCredit - totalDebit
		
		fmt.Printf("Total Debit : Rp %15.2f\n", totalDebit)
		fmt.Printf("Total Credit: Rp %15.2f\n", totalCredit)
		fmt.Printf("Calculated  : Rp %15.2f\n", calculatedBalance)
		fmt.Printf("Difference  : Rp %15.2f ", retainedEarnings.Balance-calculatedBalance)
		
		if retainedEarnings.Balance == calculatedBalance {
			fmt.Println("✓ MATCH")
		} else {
			fmt.Println("❌ MISMATCH")
		}
	}
	fmt.Println("----------------------------------------")
	fmt.Println()

	// 6. DIAGNOSIS
	fmt.Println("6. DIAGNOSIS")
	fmt.Println("========================================")
	
	if diff > 0.01 || diff < -0.01 {
		fmt.Println("❌ PROBLEM DETECTED: Balance Sheet NOT Balanced")
		fmt.Printf("   Difference: Rp %.2f\n\n", diff)
		
		fmt.Println("POSSIBLE CAUSES:")
		fmt.Println("1. Closing entries created journal lines but didn't update account balances correctly")
		fmt.Println("2. Account balance update logic has bugs (wrong debit/credit direction)")
		fmt.Println("3. Revenue accounts treated incorrectly (should have negative balance)")
		fmt.Println("4. Double-posting issue in multiple closings")
		fmt.Println()
		
		fmt.Println("RECOMMENDED ACTIONS:")
		fmt.Println("1. Check unified_period_closing_service.go line 162-188 (balance update logic)")
		fmt.Println("2. Verify Revenue account balance calculation (line 67: should be -acc.Balance)")
		fmt.Println("3. Check if balanceChange calculation considers account type correctly")
		fmt.Println("4. Consider recalculating all account balances from SSOT journal lines")
	} else {
		fmt.Println("✓ Balance Sheet is BALANCED")
		fmt.Println("  No issues detected.")
	}
	
	fmt.Println("\n========================================")
	fmt.Println("END OF DIAGNOSIS")
	fmt.Println("========================================")
}
