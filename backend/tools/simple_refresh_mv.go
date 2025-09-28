package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Connect directly to database
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	fmt.Println("🔄 Checking and Creating Account Balances Materialized View...")

	// Check if account_balances exists as materialized view
	var mvCount int64
	err = db.Raw("SELECT COUNT(*) FROM pg_matviews WHERE matviewname = 'account_balances'").Scan(&mvCount).Error
	if err != nil {
		log.Printf("Warning: Could not check materialized view existence: %v", err)
		// Fallback to general table check
		err = db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'account_balances'").Scan(&mvCount).Error
		if err != nil {
			log.Printf("Warning: Could not check table existence: %v", err)
			return
		}
	}

	if mvCount == 0 {
		fmt.Println("❌ account_balances materialized view does not exist - creating it...")
		createAccountBalancesView(db)
	} else {
		fmt.Println("✅ account_balances materialized view found")
	}

	// Refresh the materialized view
	err = db.Exec("REFRESH MATERIALIZED VIEW account_balances").Error
	if err != nil {
		log.Fatalf("❌ Failed to refresh materialized view: %v", err)
	}

	fmt.Println("✅ Materialized view refreshed successfully!")

	// Check the results
	var stats struct {
		TotalAccounts   int64 `gorm:"column:total_accounts"`
		NonZeroBalances int64 `gorm:"column:non_zero_balances"`
	}

	err = db.Raw(`
		SELECT 
			COUNT(*) as total_accounts,
			COUNT(CASE WHEN current_balance != 0 THEN 1 END) as non_zero_balances
		FROM account_balances
	`).Scan(&stats).Error

	if err != nil {
		log.Printf("Warning: Could not get statistics: %v", err)
	} else {
		fmt.Printf("📊 Statistics:\n")
		fmt.Printf("   Total Accounts: %d\n", stats.TotalAccounts)
		fmt.Printf("   Accounts with Balance: %d\n", stats.NonZeroBalances)
	}

	fmt.Println()
	fmt.Println("✅ REFRESH COMPLETED!")
	fmt.Println("Now go to frontend and hard refresh (Ctrl+F5) to see updated balances.")
}

func createAccountBalancesView(db *gorm.DB) {
	fmt.Println("🚀 Creating account_balances materialized view...")
	
	createViewSQL := `
CREATE MATERIALIZED VIEW account_balances AS
SELECT 
    a.id as account_id,
    a.code as account_code,
    a.name as account_name,
    a.type as account_type,
    a.category as account_category,
    
    -- Current balance calculation from SSOT journal system
    CASE 
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_lines') THEN
            COALESCE((
                SELECT 
                    CASE 
                        WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                            SUM(ujl.debit_amount) - SUM(ujl.credit_amount)
                        ELSE 
                            SUM(ujl.credit_amount) - SUM(ujl.debit_amount)
                    END
                FROM unified_journal_lines ujl
                JOIN unified_journal_ledger ujd ON ujl.journal_id = ujd.id
                WHERE ujl.account_id = a.id 
                  AND ujd.status = 'POSTED'
                  AND ujd.deleted_at IS NULL
            ), 0)
        ELSE 
            a.balance  -- Fallback to account table balance
    END as current_balance,
    
    -- Transaction count
    COALESCE((
        SELECT COUNT(*)
        FROM unified_journal_lines ujl
        JOIN unified_journal_ledger ujd ON ujl.journal_id = ujd.id
        WHERE ujl.account_id = a.id 
          AND ujd.status = 'POSTED'
          AND ujd.deleted_at IS NULL
    ), 0) as transaction_count,
    
    -- Last transaction date
    (
        SELECT MAX(ujd.entry_date)
        FROM unified_journal_lines ujl
        JOIN unified_journal_ledger ujd ON ujl.journal_id = ujd.id
        WHERE ujl.account_id = a.id 
          AND ujd.status = 'POSTED'
          AND ujd.deleted_at IS NULL
    ) as last_transaction_date,
    
    -- Metadata
    CASE 
        WHEN a.type IN ('ASSET', 'EXPENSE') THEN 'DEBIT'
        ELSE 'CREDIT'
    END as normal_balance,
    a.is_active,
    a.parent_id IS NULL as is_header,
    NOW() as last_updated

FROM accounts a
WHERE a.deleted_at IS NULL;
`

	err := db.Exec(createViewSQL).Error
	if err != nil {
		log.Fatalf("❌ Failed to create materialized view: %v", err)
	}
	
	fmt.Println("✅ Materialized view created successfully!")
	
	// Create indexes
	fmt.Println("🚀 Creating indexes...")
	
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_account_balances_account_id ON account_balances(account_id);",
		"CREATE INDEX IF NOT EXISTS idx_account_balances_account_type ON account_balances(account_type);",
	}
	
	for _, idx := range indexes {
		err := db.Exec(idx).Error
		if err != nil {
			log.Printf("⚠️ Warning: Failed to create index: %v", err)
		}
	}
	
	fmt.Println("✅ Indexes created successfully!")
	
	// Initial refresh
	err = db.Exec("REFRESH MATERIALIZED VIEW account_balances").Error
	if err != nil {
		log.Printf("⚠️ Warning: Initial refresh failed: %v", err)
	} else {
		fmt.Println("✅ Initial materialized view refresh completed!")
	}
}
