package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Read DB URL from environment or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "host=localhost user=postgres password=postgres dbname=accounting_db port=5432 sslmode=disable"
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("📊 Applying migration 025: Increase payment code size...")

	// Execute migration
	sql := `ALTER TABLE payments ALTER COLUMN code TYPE VARCHAR(30);`
	
	if err := db.Exec(sql).Error; err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Println("✅ Migration 025 applied successfully!")
	
	// Verify the change
	var result string
	db.Raw(`
		SELECT data_type || '(' || character_maximum_length || ')' as type
		FROM information_schema.columns 
		WHERE table_name = 'payments' AND column_name = 'code'
	`).Scan(&result)
	
	fmt.Printf("✅ Verified: payments.code is now %s\n", result)
}
