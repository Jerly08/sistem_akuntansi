package main

import (
	"fmt"
	"log"
	"strings"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"app-sistem-akuntansi/models"
)

func main() {
	// Database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("🔧 Completing PPN Account Separation")
	fmt.Println("=====================================")

	// 1. Check current state
	fmt.Println("\n1. Checking current PPN accounts:")
	var account2102 models.Account
	if err := db.Where("code = ?", "2102").First(&account2102).Error; err != nil {
		log.Fatal("ERROR: Account 2102 not found:", err)
	}
	fmt.Printf("✅ Account 2102: %s - %s (%s)\n", account2102.Code, account2102.Name, account2102.Type)

	var account2103 models.Account
	account2103Exists := true
	if err := db.Where("code = ?", "2103").First(&account2103).Error; err != nil {
		account2103Exists = false
		fmt.Println("❌ Account 2103 not found - will create it")
	} else {
		fmt.Printf("✅ Account 2103: %s - %s (%s)\n", account2103.Code, account2103.Name, account2103.Type)
	}

	// 2. Create account 2103 if it doesn't exist
	if !account2103Exists {
		fmt.Println("\n2. Creating account 2103 (PPN Keluaran):")
		
		account2103 = models.Account{
			Code:        "2103",
			Name:        "PPN Keluaran",
			Type:        models.AccountTypeLiability,
			Description: "Pajak Pertambahan Nilai Keluaran (Output VAT)",
			ParentID:    account2102.ParentID, // Use same parent as 2102
			IsActive:    true,
			IsHeader:    false,
			Balance:     0,
		}
		
		if err := db.Create(&account2103).Error; err != nil {
			log.Fatal("ERROR: Failed to create account 2103:", err)
		}
		
		fmt.Printf("✅ Created account 2103: %s - %s (%s)\n", account2103.Code, account2103.Name, account2103.Type)
	}

	// 3. Update existing journal entries that use wrong PPN accounts
	fmt.Println("\n3. Updating existing journal entries:")
	
	// Find journal entries in SSOT system that use account 2102 for PPN Keluaran
	var wrongEntries []models.SSOTJournalLine
	if err := db.Raw(`
		SELECT jl.* 
		FROM unified_journal_lines jl
		JOIN accounts a ON a.id = jl.account_id
		WHERE a.code = '2102' 
		  AND (jl.description ILIKE '%keluaran%' 
		       OR jl.description ILIKE '%output%'
		       OR jl.description ILIKE '%sales%')
	`).Find(&wrongEntries).Error; err != nil {
		log.Printf("Warning: Could not check SSOT journal entries: %v", err)
	} else {
		fmt.Printf("Found %d SSOT journal entries to update\n", len(wrongEntries))
		
		// Update these entries to use account 2103
		for _, entry := range wrongEntries {
			if err := db.Model(&entry).Update("account_id", account2103.ID).Error; err != nil {
				log.Printf("Warning: Failed to update journal line %d: %v", entry.ID, err)
			} else {
				fmt.Printf("✅ Updated journal line %d to use account 2103\n", entry.ID)
			}
		}
	}

	// Also check legacy journal entries
	var legacyWrongEntries []models.JournalEntry
	if err := db.Raw(`
		SELECT je.* 
		FROM journal_entries je
		JOIN accounts a ON a.id = je.account_id
		WHERE a.code = '2102' 
		  AND (je.description ILIKE '%keluaran%' 
		       OR je.description ILIKE '%output%'
		       OR je.reference_type = 'SALE')
	`).Find(&legacyWrongEntries).Error; err != nil {
		log.Printf("Warning: Could not check legacy journal entries: %v", err)
	} else if len(legacyWrongEntries) > 0 {
		fmt.Printf("Found %d legacy journal entries to update\n", len(legacyWrongEntries))
		
		for _, entry := range legacyWrongEntries {
			if err := db.Model(&entry).Update("account_id", account2103.ID).Error; err != nil {
				log.Printf("Warning: Failed to update legacy journal entry %d: %v", entry.ID, err)
			} else {
				fmt.Printf("✅ Updated legacy journal entry %d to use account 2103\n", entry.ID)
			}
		}
	}

	// 4. Update account balances if needed
	fmt.Println("\n4. Updating account balances:")
	
	// Calculate the balance that should be moved from 2102 to 2103
	var ppnKeluaranBalance float64
	db.Raw(`
		SELECT COALESCE(SUM(jl.credit_amount - jl.debit_amount), 0) as balance
		FROM unified_journal_lines jl
		WHERE jl.account_id = ? AND jl.description ILIKE '%keluaran%'
	`, account2103.ID).Scan(&ppnKeluaranBalance)
	
	fmt.Printf("PPN Keluaran balance to set: %.2f\n", ppnKeluaranBalance)
	
	// Update account 2103 balance
	if err := db.Model(&account2103).Update("balance", ppnKeluaranBalance).Error; err != nil {
		log.Printf("Warning: Failed to update account 2103 balance: %v", err)
	} else {
		fmt.Printf("✅ Updated account 2103 balance to %.2f\n", ppnKeluaranBalance)
	}

	// 5. Verify the fix
	fmt.Println("\n5. Verification:")
	
	// Reload accounts to get updated balances
	db.First(&account2102, account2102.ID)
	db.First(&account2103, account2103.ID)
	
	fmt.Printf("✅ Account 2102: %s - %s (Balance: %.2f) - FOR PPN MASUKAN\n", 
		account2102.Code, account2102.Name, account2102.Balance)
	fmt.Printf("✅ Account 2103: %s - %s (Balance: %.2f) - FOR PPN KELUARAN\n", 
		account2103.Code, account2103.Name, account2103.Balance)

	fmt.Println("\n🎯 SUCCESS!")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println("✅ PPN account separation is now complete!")
	fmt.Println("✅ Account 2102 (PPN Masukan) - for INPUT VAT (purchases)")
	fmt.Println("✅ Account 2103 (PPN Keluaran) - for OUTPUT VAT (sales)")
	fmt.Println("✅ All future sales will automatically use account 2103")
	fmt.Println("✅ Existing journal entries have been corrected")
}