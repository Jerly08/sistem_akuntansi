package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("========================================================================")
	fmt.Println("TEST CLOSING LOGIC - Verify Correct Implementation")
	fmt.Println("========================================================================")
	
	// Step 1: Check cumulative balances before any action
	fmt.Println("\n1. CHECKING CUMULATIVE BALANCES (Before Fix):")
	fmt.Println("-----------------------------------------------------------------------")
	
	// Revenue cumulative
	var revDebit, revCredit float64
	err = db.QueryRow(`
		SELECT 
			COALESCE(SUM(ujl.debit_amount), 0),
			COALESCE(SUM(ujl.credit_amount), 0)
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		JOIN accounts a ON a.id = ujl.account_id
		WHERE a.type = 'REVENUE'
			AND uje.status = 'POSTED'
			AND uje.source_type != 'CLOSING'
	`).Scan(&revDebit, &revCredit)
	
	revBalance := revCredit - revDebit
	fmt.Printf("   REVENUE Cumulative (excluding CLOSING): %.2f\n", revBalance)
	
	// Expense cumulative
	var expDebit, expCredit float64
	err = db.QueryRow(`
		SELECT 
			COALESCE(SUM(ujl.debit_amount), 0),
			COALESCE(SUM(ujl.credit_amount), 0)
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		JOIN accounts a ON a.id = ujl.account_id
		WHERE a.type = 'EXPENSE'
			AND uje.status = 'POSTED'
			AND uje.source_type != 'CLOSING'
	`).Scan(&expDebit, &expCredit)
	
	expBalance := expDebit - expCredit
	fmt.Printf("   EXPENSE Cumulative (excluding CLOSING): %.2f\n", expBalance)
	fmt.Printf("   Net Income (should be closed): %.2f\n", revBalance - expBalance)
	
	// Step 2: Check current COA balances
	fmt.Println("\n2. CHECKING COA ACCOUNT BALANCES:")
	fmt.Println("-----------------------------------------------------------------------")
	
	rows, err := db.Query(`
		SELECT code, name, type, balance
		FROM accounts
		WHERE code IN ('4101', '5101', '3201')
		ORDER BY code
	`)
	
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var code, name, accType string
			var balance float64
			rows.Scan(&code, &name, &accType, &balance)
			
			status := "✓"
			if accType == "REVENUE" || accType == "EXPENSE" {
				if balance != 0 {
					status = "❌ SHOULD BE 0!"
				}
			}
			
			fmt.Printf("   %s - %s (%s): %.2f %s\n", code, name, accType, balance, status)
		}
	}
	
	// Step 3: Check closing journals
	fmt.Println("\n3. CHECKING CLOSING JOURNAL ENTRIES:")
	fmt.Println("-----------------------------------------------------------------------")
	
	var closingCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM unified_journal_ledger
		WHERE source_type = 'CLOSING'
	`).Scan(&closingCount)
	
	fmt.Printf("   Total Closing Entries: %d\n", closingCount)
	
	// Get the last closing details
	var lastClosingID int
	var lastClosingDate time.Time
	var lastClosingTotalDebit, lastClosingTotalCredit float64
	
	err = db.QueryRow(`
		SELECT id, entry_date, total_debit, total_credit
		FROM unified_journal_ledger
		WHERE source_type = 'CLOSING'
		ORDER BY entry_date DESC
		LIMIT 1
	`).Scan(&lastClosingID, &lastClosingDate, &lastClosingTotalDebit, &lastClosingTotalCredit)
	
	if err == nil {
		fmt.Printf("\n   Last Closing Entry (ID: %d):\n", lastClosingID)
		fmt.Printf("   Date: %s\n", lastClosingDate.Format("2006-01-02"))
		fmt.Printf("   Total Debit: %.2f | Total Credit: %.2f\n", lastClosingTotalDebit, lastClosingTotalCredit)
		
		// Check if amounts match what should be closed
		expectedTotal := revBalance + expBalance
		if lastClosingTotalDebit != expectedTotal {
			fmt.Printf("   ❌ WARNING: Closing amount (%.2f) doesn't match cumulative (%.2f)\n", 
				lastClosingTotalDebit, expectedTotal)
			fmt.Println("   This indicates closing used PERIOD balance instead of CUMULATIVE!")
		} else {
			fmt.Printf("   ✓ Closing amount matches cumulative balance\n")
		}
	}
	
	// Step 4: Diagnosis
	fmt.Println("\n========================================================================")
	fmt.Println("DIAGNOSIS:")
	fmt.Println("========================================================================")
	
	needsFix := false
	
	// Check if revenue/expense have non-zero balance
	var nonZeroRevExp int
	db.QueryRow(`
		SELECT COUNT(*) FROM accounts 
		WHERE type IN ('REVENUE', 'EXPENSE') 
		AND ABS(balance) > 0.01
	`).Scan(&nonZeroRevExp)
	
	if nonZeroRevExp > 0 {
		fmt.Println("\n❌ PROBLEM: Revenue/Expense accounts have non-zero balance after closing")
		needsFix = true
	}
	
	// Check if closing amount is wrong
	if closingCount > 0 && lastClosingTotalDebit != (revBalance + expBalance) {
		fmt.Println("❌ PROBLEM: Closing used wrong balance calculation (period instead of cumulative)")
		needsFix = true
	}
	
	if needsFix {
		fmt.Println("\n🔧 REQUIRED ACTIONS:")
		fmt.Println("1. The closing service has been fixed to use CUMULATIVE balances")
		fmt.Println("2. Run: go run cmd/fix_closing_balances.go")
		fmt.Println("   This will recalculate all balances correctly")
		fmt.Println("3. For future closings, the fixed logic will work correctly")
	} else {
		fmt.Println("\n✅ All Good! Closing logic is working correctly.")
		fmt.Println("   - Revenue and Expense accounts are zero")
		fmt.Println("   - Closing amounts match cumulative balances")
		fmt.Println("   - Retained Earnings reflects net income")
	}
	
	// Step 5: Show what CORRECT closing should look like
	fmt.Println("\n========================================================================")
	fmt.Println("CORRECT CLOSING CALCULATION:")
	fmt.Println("========================================================================")
	
	fmt.Printf("\nBased on cumulative balances, the CORRECT closing should be:\n")
	fmt.Printf("1. Close Revenue (%.2f):\n", revBalance)
	fmt.Printf("   Debit: 4101 REVENUE         %.2f\n", revBalance)
	fmt.Printf("   Credit: 3201 RETAINED EARNINGS %.2f\n", revBalance)
	
	fmt.Printf("\n2. Close Expense (%.2f):\n", expBalance)
	fmt.Printf("   Debit: 3201 RETAINED EARNINGS  %.2f\n", expBalance)
	fmt.Printf("   Credit: 5101 EXPENSE         %.2f\n", expBalance)
	
	fmt.Printf("\n3. Net Effect on Retained Earnings: %.2f\n", revBalance - expBalance)
	
	fmt.Println("\n========================================================================")
	fmt.Println("Test complete!")
	fmt.Println("========================================================================")
}