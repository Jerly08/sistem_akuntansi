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

func main() {
	// Get database connection from environment
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:@tcp(localhost:3306)/accounting_db?charset=utf8mb4&parseTime=True&loc=Local"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("=== Checking Closing History ===\n")

	// Query all journal entries with reference_type = 'CLOSING'
	var closingEntries []JournalEntry
	result := db.Table("journal_entries").
		Where("reference_type = ?", "CLOSING").
		Order("entry_date DESC").
		Find(&closingEntries)

	if result.Error != nil {
		log.Fatal("Query error:", result.Error)
	}

	fmt.Printf("Found %d closing entries:\n\n", len(closingEntries))

	if len(closingEntries) == 0 {
		fmt.Println("❌ No closing entries found!")
		fmt.Println("\nChecking all journal entries to see what reference_types exist...")
		
		var allReferenceTypes []string
		db.Table("journal_entries").
			Select("DISTINCT reference_type").
			Pluck("reference_type", &allReferenceTypes)
		
		fmt.Println("\nExisting reference_types in database:")
		for _, rt := range allReferenceTypes {
			var count int64
			db.Table("journal_entries").Where("reference_type = ?", rt).Count(&count)
			fmt.Printf("  - %s: %d entries\n", rt, count)
		}
	} else {
		for i, entry := range closingEntries {
			fmt.Printf("%d. ID: %d\n", i+1, entry.ID)
			fmt.Printf("   Code: %s\n", entry.Code)
			fmt.Printf("   Description: %s\n", entry.Description)
			fmt.Printf("   Reference: %s\n", entry.Reference)
			fmt.Printf("   Reference Type: %s\n", entry.ReferenceType)
			fmt.Printf("   Entry Date: %s\n", entry.EntryDate)
			fmt.Printf("   Total Debit: %.2f\n", entry.TotalDebit)
			fmt.Printf("   Status: %s\n\n", entry.Status)
		}
	}
}
