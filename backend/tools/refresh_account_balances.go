package main

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/database"
)

type RefreshResult struct {
	Status               string    `json:"status"`
	TotalAccounts        int64     `json:"total_accounts"`
	AccountsWithBalance  int64     `json:"accounts_with_balance"`
	TotalDebitBalances   float64   `json:"total_debit_balances"`
	TotalCreditBalances  float64   `json:"total_credit_balances"`
	RefreshedAt          time.Time `json:"refreshed_at"`
}

type AccountBalance struct {
	AccountCode      string    `json:"account_code"`
	AccountName      string    `json:"account_name"`
	AccountType      string    `json:"account_type"`
	CurrentBalance   float64   `json:"current_balance"`
	TransactionCount int64     `json:"transaction_count"`
	LastUpdated      time.Time `json:"last_updated"`
}

func main() {
	fmt.Println("=== REFRESHING ACCOUNT BALANCES MATERIALIZED VIEW ===")
	fmt.Println()

	// Connect to database
	db := database.ConnectDB()
	if db == nil {
		log.Fatal("Failed to connect to database")
	}

	// Step 1: Refresh the materialized view
	fmt.Println("1. REFRESHING MATERIALIZED VIEW...")
	fmt.Println("=" + repeat("=", 50))
	
	start := time.Now()
	err := db.Exec("REFRESH MATERIALIZED VIEW account_balances").Error
	if err != nil {
		log.Fatalf("Failed to refresh materialized view: %v", err)
	}
	
	elapsed := time.Since(start)
	fmt.Printf("✅ Materialized view refreshed successfully in %v\n", elapsed)
	fmt.Println()

	// Step 2: Get refresh statistics
	fmt.Println("2. REFRESH STATISTICS")
	fmt.Println("=" + repeat("=", 50))
	
	var result RefreshResult
	err = db.Raw(`
		SELECT 
			'REFRESH COMPLETED' as status,
			COUNT(*) as total_accounts,
			COUNT(CASE WHEN current_balance != 0 THEN 1 END) as accounts_with_balance,
			COALESCE(SUM(CASE WHEN current_balance > 0 THEN current_balance ELSE 0 END), 0) as total_debit_balances,
			COALESCE(SUM(CASE WHEN current_balance < 0 THEN ABS(current_balance) ELSE 0 END), 0) as total_credit_balances,
			NOW() as refreshed_at
		FROM account_balances
	`).Scan(&result).Error
	
	if err != nil {
		log.Printf("Warning: Could not get refresh statistics: %v", err)
	} else {
		fmt.Printf("Status: %s\n", result.Status)
		fmt.Printf("Total Accounts: %d\n", result.TotalAccounts)
		fmt.Printf("Accounts with Balance: %d\n", result.AccountsWithBalance)
		fmt.Printf("Total Debit Balances: %.2f\n", result.TotalDebitBalances)
		fmt.Printf("Total Credit Balances: %.2f\n", result.TotalCreditBalances)
		fmt.Printf("Refreshed At: %v\n", result.RefreshedAt.Format("2006-01-02 15:04:05"))
	}
	fmt.Println()

	// Step 3: Show sample of updated balances
	fmt.Println("3. SAMPLE UPDATED BALANCES (Top 20)")
	fmt.Println("=" + repeat("=", 50))
	
	var balances []AccountBalance
	err = db.Raw(`
		SELECT 
			account_code,
			account_name,
			account_type,
			current_balance,
			transaction_count,
			last_updated
		FROM account_balances 
		WHERE current_balance != 0
		ORDER BY ABS(current_balance) DESC
		LIMIT 20
	`).Scan(&balances).Error
	
	if err != nil {
		log.Printf("Warning: Could not get sample balances: %v", err)
	} else {
		fmt.Printf("%-10s %-30s %-12s %15s %10s\n", "Code", "Name", "Type", "Balance", "TxCount")
		fmt.Printf("%s\n", repeat("-", 85))
		
		for _, balance := range balances {
			fmt.Printf("%-10s %-30s %-12s %15.2f %10d\n", 
				balance.AccountCode, 
				truncate(balance.AccountName, 30), 
				balance.AccountType, 
				balance.CurrentBalance, 
				balance.TransactionCount)
		}
	}
	fmt.Println()

	// Step 4: Check if materialized view is up to date
	fmt.Println("4. VERIFYING DATA FRESHNESS")
	fmt.Println("=" + repeat("=", 50))
	
	var checkResult struct {
		MaterializedViewRecords int64 `json:"materialized_view_records"`
		JournalLedgerRecords    int64 `json:"journal_ledger_records"`
		JournalLinesRecords     int64 `json:"journal_lines_records"`
	}
	
	err = db.Raw(`
		SELECT 
			(SELECT COUNT(*) FROM account_balances) as materialized_view_records,
			(SELECT COUNT(*) FROM unified_journal_ledger WHERE status = 'POSTED' AND deleted_at IS NULL) as journal_ledger_records,
			(SELECT COUNT(*) FROM unified_journal_lines) as journal_lines_records
	`).Scan(&checkResult).Error
	
	if err != nil {
		log.Printf("Warning: Could not verify data freshness: %v", err)
	} else {
		fmt.Printf("Materialized View Records: %d\n", checkResult.MaterializedViewRecords)
		fmt.Printf("Posted Journal Entries: %d\n", checkResult.JournalLedgerRecords)
		fmt.Printf("Journal Lines: %d\n", checkResult.JournalLinesRecords)
		
		if checkResult.MaterializedViewRecords > 0 {
			fmt.Printf("✅ Materialized view contains data and is ready\n")
		} else {
			fmt.Printf("⚠️ Materialized view is empty - check journal data\n")
		}
	}
	
	fmt.Println()
	fmt.Println("=== REFRESH COMPLETED ===")
	fmt.Println("The account balances should now be updated in the frontend.")
	fmt.Println("Recommend doing a hard refresh (Ctrl+F5) in the browser.")
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}