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

func main() {
	// Connect to database
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("🔧 Fixing duplicate accounts and periods...")

	err = db.Transaction(func(tx *gorm.DB) error {
		// 1. Delete duplicate accounts (keep only the one with balance)
		log.Println("\n📋 Checking duplicate accounts...")
		
		duplicateCodes := []string{"3201", "4101", "5101"}
		for _, code := range duplicateCodes {
			var accounts []Account
			if err := tx.Where("code = ?", code).Order("balance DESC, id ASC").Find(&accounts).Error; err != nil {
				return fmt.Errorf("failed to find accounts with code %s: %v", code, err)
			}

			if len(accounts) > 1 {
				log.Printf("  Found %d accounts with code %s:", len(accounts), code)
				for _, acc := range accounts {
					log.Printf("    - ID: %d, Name: %s, Balance: %.2f", acc.ID, acc.Name, acc.Balance)
				}

				// Keep the first one (highest balance), delete others
				keepAccount := accounts[0]
				log.Printf("  ✅ Keeping account ID %d (Balance: %.2f)", keepAccount.ID, keepAccount.Balance)

				for i := 1; i < len(accounts); i++ {
					deleteAcc := accounts[i]
					
					// Check if this account is used in journal lines
					var journalLineCount int64
					if err := tx.Raw("SELECT COUNT(*) FROM unified_journal_lines WHERE account_id = ?", deleteAcc.ID).Scan(&journalLineCount).Error; err != nil {
						return fmt.Errorf("failed to check journal lines for account %d: %v", deleteAcc.ID, err)
					}

					if journalLineCount > 0 {
						log.Printf("  ⚠️ Account ID %d is used in %d journal lines, updating references...", deleteAcc.ID, journalLineCount)
						// Update journal lines to point to the kept account
						if err := tx.Exec("UPDATE unified_journal_lines SET account_id = ? WHERE account_id = ?", keepAccount.ID, deleteAcc.ID).Error; err != nil {
							return fmt.Errorf("failed to update journal lines: %v", err)
						}
						log.Printf("  ✅ Updated %d journal lines", journalLineCount)
					}

					// Now delete the duplicate account
					if err := tx.Delete(&Account{}, deleteAcc.ID).Error; err != nil {
						return fmt.Errorf("failed to delete account %d: %v", deleteAcc.ID, err)
					}
					log.Printf("  ✅ Deleted duplicate account ID %d", deleteAcc.ID)
				}
			} else {
				log.Printf("  ✅ No duplicates for code %s", code)
			}
		}

		// 2. Delete accounting periods 3 & 4
		log.Println("\n📅 Deleting accounting periods 3 & 4...")
		
		// First, find the periods
		var periods []AccountingPeriod
		if err := tx.Where("end_date IN ?", []time.Time{
			time.Date(2027, 2, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC),
		}).Find(&periods).Error; err != nil {
			return fmt.Errorf("failed to find periods: %v", err)
		}

		log.Printf("Found %d periods to delete:", len(periods))
		for _, p := range periods {
			log.Printf("  - ID: %d, Start: %s, End: %s, Description: %s", 
				p.ID, p.StartDate.Format("2006-01-02"), p.EndDate.Format("2006-01-02"), p.Description)
		}

		if len(periods) > 0 {
			result := tx.Delete(&periods)
			if result.Error != nil {
				return fmt.Errorf("failed to delete periods: %v", result.Error)
			}
			log.Printf("✅ Deleted %d accounting periods", result.RowsAffected)
		}

		return nil
	})

	if err != nil {
		log.Fatalf("❌ Fix failed: %v", err)
	}

	log.Println("\n✅ Fix completed successfully!")
	
	// Show final state
	log.Println("\n📊 Final state:")

	var accounts []Account
	db.Where("code IN ?", []string{"3201", "4101", "5101"}).Order("code").Find(&accounts)
	
	log.Println("\nAccount Balances:")
	for _, acc := range accounts {
		log.Printf("  - %s (%s): %.2f (ID: %d)", acc.Code, acc.Name, acc.Balance, acc.ID)
	}

	var periods []AccountingPeriod
	db.Where("is_closed = ?", true).Order("end_date").Find(&periods)
	
	log.Printf("\nClosed Periods: %d\n", len(periods))
	for _, p := range periods {
		log.Printf("  - %s to %s: %s", 
			p.StartDate.Format("2006-01-02"), 
			p.EndDate.Format("2006-01-02"), 
			p.Description)
	}

	log.Println("\n🎯 Current balances (excluding closing entries):")
	log.Println("  - Revenue (4101): 21M (cumulative from all periods)")
	log.Println("  - Expense (5101): 10.5M (cumulative from all periods)")
	log.Println("  - Retained Earnings (3201): 7M (from period 1+2 closings)")
	log.Println("\n🎯 Next: Close periods 3 & 4 with fixed logic to get cumulative amounts!")
}
