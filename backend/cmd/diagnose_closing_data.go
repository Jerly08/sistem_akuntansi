package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type JournalEntry struct {
	ID            uint64  `gorm:"primaryKey"`
	Code          string
	Description   string
	Reference     string
	ReferenceType string
	EntryDate     string
	TotalDebit    float64
	TotalCredit   float64
	Status        string
}

type AccountingPeriod struct {
	ID               uint64  `gorm:"primaryKey"`
	StartDate        string
	EndDate          string
	Description      string
	IsClosed         bool
	IsLocked         bool
	TotalRevenue     float64
	TotalExpense     float64
	NetIncome        float64
	ClosingJournalID *uint64
	ClosedAt         *string
}

func main() {
	// Get database connection from environment
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:@tcp(localhost:3306)/accounting_db?charset=utf8mb4&parseTime=True&loc=Local"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║          DIAGNOSTIC: Period Closing Data Analysis             ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// ============================================================
	// 1. Check Fiscal Year Closing (journal_entries)
	// ============================================================
	fmt.Println("📊 1. FISCAL YEAR CLOSING (journal_entries table)")
	fmt.Println("   Source: journal_entries WHERE reference_type = 'CLOSING'")
	fmt.Println("   " + string([]rune{0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500}))
	
	var fiscalClosings []JournalEntry
	result := db.Table("journal_entries").
		Where("reference_type = ?", "CLOSING").
		Order("entry_date DESC").
		Find(&fiscalClosings)

	if result.Error != nil {
		fmt.Printf("   ❌ Query error: %v\n", result.Error)
	} else {
		fmt.Printf("   ✅ Found: %d fiscal year closing entries\n\n", len(fiscalClosings))
		
		if len(fiscalClosings) > 0 {
			fmt.Println("   Latest entries:")
			for i, entry := range fiscalClosings {
				if i >= 3 { // Show max 3
					break
				}
				fmt.Printf("   %d. Code: %s\n", i+1, entry.Code)
				fmt.Printf("      Date: %s\n", entry.EntryDate)
				fmt.Printf("      Description: %s\n", entry.Description)
				fmt.Printf("      Amount: Rp %.2f\n", entry.TotalDebit)
				fmt.Println()
			}
		} else {
			fmt.Println("   ⚠️  No fiscal year closing entries found")
			fmt.Println()
		}
	}

	// ============================================================
	// 2. Check Period Closing (accounting_periods)
	// ============================================================
	fmt.Println("📊 2. PERIOD CLOSING (accounting_periods table)")
	fmt.Println("   Source: accounting_periods WHERE is_closed = true")
	fmt.Println("   " + string([]rune{0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500, 0x2500}))
	
	// Check if table exists
	var tableExists bool
	db.Raw("SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'accounting_periods'").Scan(&tableExists)
	
	if !tableExists {
		fmt.Println("   ❌ Table 'accounting_periods' does NOT exist!")
		fmt.Println("   ⚠️  Period closing feature is not implemented in database")
		fmt.Println()
	} else {
		var periodClosings []AccountingPeriod
		result := db.Table("accounting_periods").
			Where("is_closed = ?", true).
			Order("end_date DESC").
			Find(&periodClosings)

		if result.Error != nil {
			fmt.Printf("   ❌ Query error: %v\n", result.Error)
		} else {
			fmt.Printf("   ✅ Found: %d period closing entries\n\n", len(periodClosings))
			
			if len(periodClosings) > 0 {
				fmt.Println("   Latest closed periods:")
				for i, period := range periodClosings {
					if i >= 3 { // Show max 3
						break
					}
					fmt.Printf("   %d. Period: %s to %s\n", i+1, period.StartDate, period.EndDate)
					fmt.Printf("      Description: %s\n", period.Description)
					fmt.Printf("      Net Income: Rp %.2f\n", period.NetIncome)
					fmt.Printf("      Closed: %v, Locked: %v\n", period.IsClosed, period.IsLocked)
					if period.ClosedAt != nil {
						fmt.Printf("      Closed At: %s\n", *period.ClosedAt)
					}
					fmt.Println()
				}
			} else {
				fmt.Println("   ⚠️  No closed periods found in accounting_periods table")
				fmt.Println()
			}
		}
	}

	// ============================================================
	// 3. Summary & Recommendations
	// ============================================================
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    SUMMARY & RECOMMENDATIONS                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	
	fiscalCount := len(fiscalClosings)
	periodCount := 0
	if tableExists {
		var periodClosings []AccountingPeriod
		db.Table("accounting_periods").Where("is_closed = ?", true).Find(&periodClosings)
		periodCount = len(periodClosings)
	}
	
	fmt.Println("📌 Current Status:")
	fmt.Printf("   - Fiscal Year Closings: %d entries\n", fiscalCount)
	fmt.Printf("   - Period Closings: %d entries\n", periodCount)
	fmt.Println()
	
	if fiscalCount > 0 && periodCount == 0 {
		fmt.Println("💡 Recommendation:")
		fmt.Println("   ✅ You have fiscal year closing data")
		fmt.Println("   ⚠️  Frontend currently shows fiscal year closings")
		fmt.Println("   📝 To show period closings, you need to:")
		fmt.Println("      1. Implement period closing feature")
		fmt.Println("      2. Create entries in accounting_periods table")
		fmt.Println("      3. Update frontend to fetch from accounting_periods")
		fmt.Println()
	} else if fiscalCount == 0 && periodCount == 0 {
		fmt.Println("⚠️  WARNING:")
		fmt.Println("   ❌ No closing data found in either table")
		fmt.Println("   📝 Action needed:")
		fmt.Println("      1. Perform fiscal year closing, OR")
		fmt.Println("      2. Implement and perform period closing")
		fmt.Println()
	} else if fiscalCount > 0 && periodCount > 0 {
		fmt.Println("✅ GOOD:")
		fmt.Println("   ✅ You have both fiscal year and period closing data")
		fmt.Println("   📝 Frontend can show both types")
		fmt.Println()
	}
	
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                         API ENDPOINTS                          ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Currently used endpoint:")
	fmt.Println("   GET /api/v1/fiscal-closing/history")
	fmt.Println("   → Returns: journal_entries WHERE reference_type = 'CLOSING'")
	fmt.Println()
	fmt.Println("Alternative (if period closing implemented):")
	fmt.Println("   GET /api/v1/period-closing/history")
	fmt.Println("   → Returns: accounting_periods WHERE is_closed = true")
	fmt.Println()
}
