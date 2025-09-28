package database

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// EnsureAccountBalancesMaterializedView ensures the account_balances materialized view exists and is up to date
func EnsureAccountBalancesMaterializedView(db *gorm.DB) error {
	log.Println("🔄 Ensuring account_balances materialized view exists...")

	// Check if materialized view exists
	var mvCount int64
	err := db.Raw("SELECT COUNT(*) FROM pg_matviews WHERE matviewname = 'account_balances'").Scan(&mvCount).Error
	if err != nil {
		log.Printf("Warning: Could not check materialized view existence: %v", err)
		// Fallback to general table check
		err = db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'account_balances'").Scan(&mvCount).Error
		if err != nil {
			return fmt.Errorf("could not check table existence: %v", err)
		}
	}

	if mvCount == 0 {
		log.Println("❌ account_balances materialized view does not exist - creating it...")
		if err := createAccountBalancesMaterializedView(db); err != nil {
			return fmt.Errorf("failed to create materialized view: %v", err)
		}
	} else {
		log.Println("✅ account_balances materialized view found - refreshing...")
		if err := refreshAccountBalancesMaterializedView(db); err != nil {
			return fmt.Errorf("failed to refresh materialized view: %v", err)
		}
	}

	return nil
}

// createAccountBalancesMaterializedView creates the materialized view
func createAccountBalancesMaterializedView(db *gorm.DB) error {
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
    
    -- Transaction count and metadata
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
    
    -- Total debits and credits for transparency
    COALESCE((
        SELECT SUM(ujl.debit_amount)
        FROM unified_journal_lines ujl
        JOIN unified_journal_ledger ujd ON ujl.journal_id = ujd.id
        WHERE ujl.account_id = a.id 
          AND ujd.status = 'POSTED'
          AND ujd.deleted_at IS NULL
    ), 0) as total_debits,
    
    COALESCE((
        SELECT SUM(ujl.credit_amount)
        FROM unified_journal_lines ujl
        JOIN unified_journal_ledger ujd ON ujl.journal_id = ujd.id
        WHERE ujl.account_id = a.id 
          AND ujd.status = 'POSTED'
          AND ujd.deleted_at IS NULL
    ), 0) as total_credits,
    
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
		return err
	}

	log.Println("✅ Materialized view created successfully!")

	// Create indexes
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_account_balances_account_id ON account_balances(account_id);",
		"CREATE INDEX IF NOT EXISTS idx_account_balances_account_type ON account_balances(account_type);",
		"CREATE INDEX IF NOT EXISTS idx_account_balances_current_balance ON account_balances(current_balance) WHERE current_balance != 0;",
	}

	for _, idx := range indexes {
		err := db.Exec(idx).Error
		if err != nil {
			log.Printf("⚠️ Warning: Failed to create index: %v", err)
		}
	}

	log.Println("✅ Indexes created successfully!")

	// Initial refresh
	return refreshAccountBalancesMaterializedView(db)
}

// refreshAccountBalancesMaterializedView refreshes the materialized view
func refreshAccountBalancesMaterializedView(db *gorm.DB) error {
	err := db.Exec("REFRESH MATERIALIZED VIEW account_balances").Error
	if err != nil {
		return err
	}

	// Get statistics
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
		log.Printf("✅ Materialized view refreshed! Accounts: %d, With Balance: %d", 
			stats.TotalAccounts, stats.NonZeroBalances)
	}

	return nil
}

// RefreshAccountBalancesPublic provides public access to refresh functionality
func RefreshAccountBalancesPublic(db *gorm.DB) error {
	return refreshAccountBalancesMaterializedView(db)
}