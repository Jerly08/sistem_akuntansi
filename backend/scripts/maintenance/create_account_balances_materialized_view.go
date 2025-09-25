package main

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/database"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🔧 Creating Account Balances Materialized View (SSOT Compatible)")
	fmt.Println("================================================================")
	fmt.Println()

	db := database.ConnectDB()
	if db == nil {
		log.Fatal("❌ Gagal koneksi ke database")
	}

	fmt.Println("🔗 Berhasil terhubung ke database")

	// Step 1: Drop existing view/materialized view if exists
	fmt.Println("\n🗑️ Step 1: Menghapus account_balances yang sudah ada (jika ada)...")
	
	dropQueries := []string{
		"DROP MATERIALIZED VIEW IF EXISTS account_balances CASCADE",
		"DROP VIEW IF EXISTS account_balances CASCADE",
	}
	
	for _, query := range dropQueries {
		if err := db.Exec(query).Error; err != nil {
			fmt.Printf("   ⚠️ Warning: %v\n", err)
		}
	}
	fmt.Println("   ✅ Cleanup selesai")

	// Step 2: Check if SSOT tables exist
	fmt.Println("\n🔍 Step 2: Memeriksa tabel SSOT...")
	
	var ssoTExists bool
	err := db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_ledger')").Scan(&ssoTExists).Error
	if err != nil {
		fmt.Printf("❌ Error checking SSOT tables: %v\n", err)
		return
	}

	// Step 3: Create appropriate materialized view
	if ssoTExists {
		fmt.Println("   ✅ SSOT tables ditemukan - membuat materialized view SSOT")
		if err := createSSOTMaterializedView(db); err != nil {
			fmt.Printf("❌ Error creating SSOT materialized view: %v\n", err)
			return
		}
	} else {
		fmt.Println("   ⚠️ SSOT tables tidak ditemukan - membuat materialized view classic")
		if err := createClassicMaterializedView(db); err != nil {
			fmt.Printf("❌ Error creating classic materialized view: %v\n", err)
			return
		}
	}

	// Step 4: Create index
	fmt.Println("\n🔧 Step 4: Membuat index untuk performance...")
	indexQuery := "CREATE UNIQUE INDEX IF NOT EXISTS idx_account_balances_account_id ON account_balances(account_id)"
	if err := db.Exec(indexQuery).Error; err != nil {
		fmt.Printf("   ⚠️ Warning creating index: %v\n", err)
	} else {
		fmt.Println("   ✅ Index berhasil dibuat")
	}

	// Step 5: Initial refresh
	fmt.Println("\n🔄 Step 5: Initial refresh materialized view...")
	if err := db.Exec("REFRESH MATERIALIZED VIEW account_balances").Error; err != nil {
		fmt.Printf("❌ Error refreshing materialized view: %v\n", err)
		return
	}
	fmt.Println("   ✅ Materialized view berhasil di-refresh")

	// Step 6: Test the view
	fmt.Println("\n🧪 Step 6: Testing materialized view...")
	testMaterializedView(db)

	fmt.Println("\n🎉 MATERIALIZED VIEW ACCOUNT_BALANCES BERHASIL DIBUAT!")
	fmt.Println("✅ View sekarang kompatibel dengan SSOT Journal System")
	fmt.Println("✅ Dapat digunakan untuk financial reports")
	fmt.Println("✅ Script reset_transaction_data_gorm.go sekarang akan berfungsi")
}

func createSSOTMaterializedView(db *gorm.DB) error {
	fmt.Println("\n🏗️ Step 3a: Membuat SSOT Materialized View...")
	
	createQuery := `
	CREATE MATERIALIZED VIEW account_balances AS
	WITH journal_totals AS (
		SELECT 
			jl.account_id,
			SUM(jl.debit_amount) as total_debits,
			SUM(jl.credit_amount) as total_credits,
			COUNT(*) as transaction_count,
			MAX(jd.posted_at) as last_transaction_date
		FROM unified_journal_lines jl
		JOIN unified_journal_ledger jd ON jl.journal_id = jd.id
		WHERE jd.status = 'POSTED' 
		  AND jd.deleted_at IS NULL
		GROUP BY jl.account_id
	)
	SELECT 
		a.id as account_id,
		a.code as account_code,
		a.name as account_name,
		a.type as account_type,
		a.category as account_category,
		
		-- Get normal balance from account type
		CASE 
			WHEN a.type IN ('ASSET', 'EXPENSE') THEN 'DEBIT'
			WHEN a.type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN 'CREDIT'
			ELSE 'DEBIT'
		END as normal_balance,
		
		-- Journal totals
		COALESCE(jt.total_debits, 0) as total_debits,
		COALESCE(jt.total_credits, 0) as total_credits,
		COALESCE(jt.transaction_count, 0) as transaction_count,
		jt.last_transaction_date,
		
		-- Calculate current balance based on normal balance
		CASE 
			WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
				COALESCE(jt.total_debits, 0) - COALESCE(jt.total_credits, 0)
			WHEN a.type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN 
				COALESCE(jt.total_credits, 0) - COALESCE(jt.total_debits, 0)
			ELSE 0
		END as current_balance,
		
		-- Metadata
		NOW() as last_updated,
		a.is_active,
		a.is_header
	FROM accounts a
	LEFT JOIN journal_totals jt ON a.id = jt.account_id
	WHERE a.deleted_at IS NULL
	`

	err := db.Exec(createQuery).Error
	if err != nil {
		return fmt.Errorf("gagal membuat SSOT materialized view: %v", err)
	}

	fmt.Println("   ✅ SSOT Materialized View berhasil dibuat")
	return nil
}

func createClassicMaterializedView(db *gorm.DB) error {
	fmt.Println("\n🏗️ Step 3b: Membuat Classic Materialized View...")
	
	createQuery := `
	CREATE MATERIALIZED VIEW account_balances AS
	WITH journal_totals AS (
		SELECT 
			jl.account_id,
			SUM(jl.debit_amount) as total_debits,
			SUM(jl.credit_amount) as total_credits,
			COUNT(*) as transaction_count,
			MAX(je.created_at) as last_transaction_date
		FROM journal_lines jl
		JOIN journal_entries je ON jl.journal_entry_id = je.id
		WHERE je.status = 'POSTED'
		  AND je.deleted_at IS NULL
		  AND jl.deleted_at IS NULL
		GROUP BY jl.account_id
	)
	SELECT 
		a.id as account_id,
		a.code as account_code,
		a.name as account_name,
		a.type as account_type,
		a.category as account_category,
		
		-- Get normal balance from account type
		CASE 
			WHEN a.type IN ('ASSET', 'EXPENSE') THEN 'DEBIT'
			WHEN a.type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN 'CREDIT'
			ELSE 'DEBIT'
		END as normal_balance,
		
		-- Journal totals
		COALESCE(jt.total_debits, 0) as total_debits,
		COALESCE(jt.total_credits, 0) as total_credits,
		COALESCE(jt.transaction_count, 0) as transaction_count,
		jt.last_transaction_date,
		
		-- Calculate current balance based on normal balance
		CASE 
			WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
				COALESCE(jt.total_debits, 0) - COALESCE(jt.total_credits, 0)
			WHEN a.type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN 
				COALESCE(jt.total_credits, 0) - COALESCE(jt.total_debits, 0)
			ELSE 0
		END as current_balance,
		
		-- Metadata
		NOW() as last_updated,
		a.is_active,
		a.is_header
	FROM accounts a
	LEFT JOIN journal_totals jt ON a.id = jt.account_id
	WHERE a.deleted_at IS NULL
	`

	err := db.Exec(createQuery).Error
	if err != nil {
		return fmt.Errorf("gagal membuat classic materialized view: %v", err)
	}

	fmt.Println("   ✅ Classic Materialized View berhasil dibuat")
	return nil
}

func testMaterializedView(db *gorm.DB) {
	// Test 1: Count total accounts
	var totalAccounts int64
	err := db.Raw("SELECT COUNT(*) FROM account_balances").Scan(&totalAccounts).Error
	if err != nil {
		fmt.Printf("   ❌ Error testing view: %v\n", err)
		return
	}
	fmt.Printf("   📊 Total accounts in view: %d\n", totalAccounts)

	// Test 2: Count accounts with activity
	var activeAccounts int64
	err = db.Raw("SELECT COUNT(*) FROM account_balances WHERE transaction_count > 0").Scan(&activeAccounts).Error
	if err != nil {
		fmt.Printf("   ❌ Error counting active accounts: %v\n", err)
		return
	}
	fmt.Printf("   💰 Accounts with transactions: %d\n", activeAccounts)

	// Test 3: Check balance totals
	type BalanceSummary struct {
		TotalAssets      float64 `gorm:"column:total_assets"`
		TotalLiabilities float64 `gorm:"column:total_liabilities"`
		TotalEquity      float64 `gorm:"column:total_equity"`
		TotalRevenue     float64 `gorm:"column:total_revenue"`
		TotalExpenses    float64 `gorm:"column:total_expenses"`
	}

	var summary BalanceSummary
	err = db.Raw(`
		SELECT 
			COALESCE(SUM(CASE WHEN account_type = 'ASSET' THEN current_balance ELSE 0 END), 0) as total_assets,
			COALESCE(SUM(CASE WHEN account_type = 'LIABILITY' THEN current_balance ELSE 0 END), 0) as total_liabilities,
			COALESCE(SUM(CASE WHEN account_type = 'EQUITY' THEN current_balance ELSE 0 END), 0) as total_equity,
			COALESCE(SUM(CASE WHEN account_type = 'REVENUE' THEN current_balance ELSE 0 END), 0) as total_revenue,
			COALESCE(SUM(CASE WHEN account_type = 'EXPENSE' THEN current_balance ELSE 0 END), 0) as total_expenses
		FROM account_balances
	`).Scan(&summary).Error

	if err != nil {
		fmt.Printf("   ❌ Error getting balance summary: %v\n", err)
		return
	}

	fmt.Printf("   💼 Balance Summary:\n")
	fmt.Printf("      Assets: %.2f\n", summary.TotalAssets)
	fmt.Printf("      Liabilities: %.2f\n", summary.TotalLiabilities)
	fmt.Printf("      Equity: %.2f\n", summary.TotalEquity)
	fmt.Printf("      Revenue: %.2f\n", summary.TotalRevenue)
	fmt.Printf("      Expenses: %.2f\n", summary.TotalExpenses)

	// Test 4: Check if balanced
	balanceDiff := summary.TotalAssets - (summary.TotalLiabilities + summary.TotalEquity + summary.TotalRevenue - summary.TotalExpenses)
	if balanceDiff < 0.01 && balanceDiff > -0.01 { // Allow for small floating point differences
		fmt.Printf("   ✅ Balance sheet is balanced (diff: %.2f)\n", balanceDiff)
	} else {
		fmt.Printf("   ⚠️  Balance sheet difference: %.2f\n", balanceDiff)
	}

	fmt.Println("   ✅ Testing selesai")
}