package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable TimeZone=Asia/Jakarta"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("REFRESH ACCOUNT BALANCES FROM SSOT JOURNAL")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()

	// Step 1: Get current balances before refresh
	fmt.Println("📊 STEP 1: Current Account Balances (Before Refresh)")
	fmt.Println(string(make([]byte, 80)))
	
	type AccountBalance struct {
		Code    string
		Name    string
		Type    string
		Balance float64
	}

	var beforeBalances []AccountBalance
	db.Raw(`
		SELECT code, name, account_type as type, balance 
		FROM accounts 
		WHERE code IN ('5101', '1301', '2101', '2103', '4101')
		ORDER BY code
	`).Scan(&beforeBalances)

	for _, acc := range beforeBalances {
		fmt.Printf("%-6s %-40s %15.2f\n", acc.Code, acc.Name, acc.Balance)
	}
	fmt.Println()

	// Step 2: Calculate correct balances from unified_journal_ledger
	fmt.Println("📊 STEP 2: Calculating Balances from SSOT Journal")
	fmt.Println(string(make([]byte, 80)))

	type JournalBalance struct {
		AccountID    uint
		Code         string
		Name         string
		TotalDebit   float64
		TotalCredit  float64
		NetBalance   float64
	}

	var journalBalances []JournalBalance
	db.Raw(`
		SELECT 
			a.id as account_id,
			a.code,
			a.name,
			COALESCE(SUM(jl.debit_amount), 0) as total_debit,
			COALESCE(SUM(jl.credit_amount), 0) as total_credit,
			COALESCE(SUM(jl.debit_amount), 0) - COALESCE(SUM(jl.credit_amount), 0) as net_balance
		FROM accounts a
		LEFT JOIN unified_journal_ledger jl ON jl.account_id = a.id
		WHERE a.code IN ('5101', '1301', '2101', '2103', '4101')
		GROUP BY a.id, a.code, a.name
		ORDER BY a.code
	`).Scan(&journalBalances)

	for _, acc := range journalBalances {
		fmt.Printf("%-6s %-40s Dr: %15.2f | Cr: %15.2f | Net: %15.2f\n", 
			acc.Code, acc.Name, acc.TotalDebit, acc.TotalCredit, acc.NetBalance)
	}
	fmt.Println()

	// Step 3: Update account balances
	fmt.Println("🔄 STEP 3: Updating Account Balances")
	fmt.Println(string(make([]byte, 80)))

	for _, acc := range journalBalances {
		result := db.Exec(`
			UPDATE accounts 
			SET balance = ? 
			WHERE id = ?
		`, acc.NetBalance, acc.AccountID)

		if result.Error != nil {
			fmt.Printf("❌ Failed to update %s: %v\n", acc.Code, result.Error)
		} else {
			fmt.Printf("✅ Updated %-6s %-40s Balance: %15.2f\n", 
				acc.Code, acc.Name, acc.NetBalance)
		}
	}
	fmt.Println()

	// Step 4: Refresh parent account balances (recursive)
	fmt.Println("🔄 STEP 4: Refreshing Parent Account Balances")
	fmt.Println(string(make([]byte, 80)))

	// This is a simplified version - should be done recursively in production
	refreshParentSQL := `
		UPDATE accounts parent
		SET balance = (
			SELECT COALESCE(SUM(child.balance), 0)
			FROM accounts child
			WHERE child.parent_id = parent.id
		)
		WHERE parent.is_header = true
	`
	
	result := db.Exec(refreshParentSQL)
	if result.Error != nil {
		fmt.Printf("❌ Failed to refresh parent balances: %v\n", result.Error)
	} else {
		fmt.Printf("✅ Refreshed parent account balances (affected: %d rows)\n", result.RowsAffected)
	}
	fmt.Println()

	// Step 5: Verify results
	fmt.Println("📊 STEP 5: Verification - Account Balances (After Refresh)")
	fmt.Println(string(make([]byte, 80)))

	var afterBalances []AccountBalance
	db.Raw(`
		SELECT code, name, account_type as type, balance 
		FROM accounts 
		WHERE code IN ('5101', '1301', '2101', '2103', '4101', '5000', '2000', '4000')
		ORDER BY code
	`).Scan(&afterBalances)

	for _, acc := range afterBalances {
		fmt.Printf("%-6s %-40s %15.2f\n", acc.Code, acc.Name, acc.Balance)
	}
	fmt.Println()

	// Step 6: Check Balance Sheet equation
	fmt.Println("📊 STEP 6: Balance Sheet Verification")
	fmt.Println(string(make([]byte, 80)))

	type BSCheck struct {
		TotalAssets      float64
		TotalLiabilities float64
		TotalEquity      float64
		Difference       float64
	}

	var bsCheck BSCheck
	db.Raw(`
		SELECT 
			COALESCE(SUM(CASE WHEN account_type = 'ASSET' THEN balance ELSE 0 END), 0) as total_assets,
			COALESCE(SUM(CASE WHEN account_type = 'LIABILITY' THEN balance ELSE 0 END), 0) as total_liabilities,
			COALESCE(SUM(CASE WHEN account_type = 'EQUITY' THEN balance ELSE 0 END), 0) as total_equity
		FROM accounts
		WHERE is_header = false
	`).Scan(&bsCheck)

	bsCheck.Difference = bsCheck.TotalAssets - (bsCheck.TotalLiabilities + bsCheck.TotalEquity)

	fmt.Printf("Assets:      %20.2f\n", bsCheck.TotalAssets)
	fmt.Printf("Liabilities: %20.2f\n", bsCheck.TotalLiabilities)
	fmt.Printf("Equity:      %20.2f\n", bsCheck.TotalEquity)
	fmt.Println(string(make([]byte, 80)))
	fmt.Printf("Difference:  %20.2f", bsCheck.Difference)
	
	if bsCheck.Difference == 0 {
		fmt.Println(" ✅ BALANCED!")
	} else {
		fmt.Printf(" ⚠️ NOT BALANCED! (off by Rp %.2f)\n", bsCheck.Difference)
	}
	fmt.Println()

	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("✅ REFRESH COMPLETE")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Refresh Chart of Accounts page in browser (Ctrl+F5)")
	fmt.Println("  2. Verify account balances are correct")
	fmt.Println("  3. Check that Balance Sheet balances")
	fmt.Println()
}

