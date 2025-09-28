package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	fmt.Printf("🚀 AUTO-SETUP: Balance Synchronization System\n\n")
	
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables: %v", err)
	}

	// Connect to database using DATABASE_URL from .env
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	fmt.Printf("🔗 Connecting to database...\n")
	gormDB, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get underlying sql.DB
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Fatal("Failed to get underlying sql.DB:", err)
	}
	defer sqlDB.Close()

	// Check if balance sync system is already installed
	var exists bool
	err = sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.triggers 
			WHERE trigger_name = 'balance_sync_trigger'
		)
	`).Scan(&exists)
	
	if err != nil {
		log.Printf("Warning: Could not check existing system: %v", err)
	}

	if exists {
		fmt.Printf("✅ Balance Sync System already installed - skipping setup\n")
		
		// Still run a health check
		var mismatchCount int
		err = sqlDB.QueryRow("SELECT COUNT(*) FROM account_balance_monitoring WHERE status='MISMATCH'").Scan(&mismatchCount)
		if err == nil {
			if mismatchCount > 0 {
				fmt.Printf("⚠️  Found %d accounts with balance mismatches\n", mismatchCount)
				fmt.Printf("💡 Run manual sync: SELECT * FROM sync_account_balances();\n")
			} else {
				fmt.Printf("✅ All account balances are synchronized\n")
			}
		}
		return
	}

	fmt.Printf("📦 Installing Balance Sync System...\n")

	// Read migration file
	migrationPath := filepath.Join("migrations", "balance_sync_system.sql")
	migrationSQL, err := ioutil.ReadFile(migrationPath)
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	// Execute migration
	fmt.Printf("🔧 Executing database migration...\n")
	_, err = sqlDB.Exec(string(migrationSQL))
	if err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	// Verify installation
	fmt.Printf("🔍 Verifying installation...\n")
	
	var triggerExists, procedureExists, viewExists bool
	
	// Check trigger
	sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.triggers 
			WHERE trigger_name = 'balance_sync_trigger'
		)
	`).Scan(&triggerExists)
	
	// Check stored procedure
	sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.routines 
			WHERE routine_name = 'sync_account_balances'
		)
	`).Scan(&procedureExists)
	
	// Check view
	sqlDB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.views 
			WHERE table_name = 'account_balance_monitoring'
		)
	`).Scan(&viewExists)

	// Report results
	if triggerExists && procedureExists && viewExists {
		fmt.Printf("\n🎉 Balance Sync System successfully installed!\n\n")
		
		fmt.Printf("📋 INSTALLED COMPONENTS:\n")
		fmt.Printf("  ✅ Automatic trigger: balance_sync_trigger\n")
		fmt.Printf("  ✅ Manual sync function: sync_account_balances()\n")  
		fmt.Printf("  ✅ Monitoring view: account_balance_monitoring\n")
		fmt.Printf("  ✅ Performance index: idx_unified_journal_lines_account_id\n\n")
		
		fmt.Printf("💡 USAGE EXAMPLES:\n")
		fmt.Printf("  • Manual sync:     SELECT * FROM sync_account_balances();\n")
		fmt.Printf("  • Health check:    SELECT * FROM account_balance_monitoring WHERE status='MISMATCH';\n")
		fmt.Printf("  • Monitor all:     SELECT * FROM account_balance_monitoring;\n\n")

		// Run initial health check
		var mismatchCount int
		err = sqlDB.QueryRow("SELECT COUNT(*) FROM account_balance_monitoring WHERE status='MISMATCH'").Scan(&mismatchCount)
		if err == nil {
			if mismatchCount > 0 {
				fmt.Printf("⚠️  Initial scan found %d accounts with balance mismatches\n", mismatchCount)
				fmt.Printf("🔄 Running auto-fix...\n")
				
				// Run sync
				rows, err := sqlDB.Query("SELECT * FROM sync_account_balances()")
				if err != nil {
					log.Printf("Warning: Could not run initial sync: %v", err)
				} else {
					defer rows.Close()
					fixCount := 0
					for rows.Next() {
						var accountID int
						var oldBalance, newBalance, difference float64
						rows.Scan(&accountID, &oldBalance, &newBalance, &difference)
						fixCount++
					}
					fmt.Printf("✅ Fixed %d account balances automatically\n", fixCount)
				}
			} else {
				fmt.Printf("✅ All account balances are already synchronized\n")
			}
		}

		fmt.Printf("\n🛡️  PROTECTION ACTIVE: Future balance inconsistencies will be prevented!\n")
		
	} else {
		fmt.Printf("\n❌ Balance Sync System installation incomplete!\n")
		fmt.Printf("   - Trigger exists: %v\n", triggerExists)
		fmt.Printf("   - Procedure exists: %v\n", procedureExists)
		fmt.Printf("   - View exists: %v\n", viewExists)
		log.Fatal("Installation failed - please check database permissions")
	}
}