package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

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
	fmt.Println("TRACING CLOSING EXECUTION - Why Manual Balance Update Didn't Work")
	fmt.Println("========================================================================")
	
	// Check if closing journal was created
	var closingJournalID int
	var closingDate string
	err = db.QueryRow(`
		SELECT id, entry_date
		FROM unified_journal_ledger
		WHERE source_type = 'CLOSING'
		ORDER BY entry_date DESC
		LIMIT 1
	`).Scan(&closingJournalID, &closingDate)
	
	if err != nil {
		fmt.Println("\n❌ No closing journal found!")
		return
	}
	
	fmt.Printf("\n✓ Last Closing Journal ID: %d (Date: %s)\n", closingJournalID, closingDate)
	
	// Get closing journal lines
	fmt.Println("\n1. Checking closing journal lines...")
	fmt.Println("-----------------------------------------------------------------------")
	
	rows, err := db.Query(`
		SELECT ujl.id, ujl.account_id, a.code, a.name, a.type,
		       ujl.debit_amount, ujl.credit_amount
		FROM unified_journal_lines ujl
		JOIN accounts a ON a.id = ujl.account_id
		WHERE ujl.journal_id = $1
		ORDER BY ujl.line_number
	`, closingJournalID)
	
	if err != nil {
		log.Fatalf("Failed to query lines: %v", err)
	}
	defer rows.Close()
	
	type Line struct {
		LineID        int
		AccountID     int
		Code          string
		Name          string
		Type          string
		DebitAmount   float64
		CreditAmount  float64
	}
	
	var lines []Line
	for rows.Next() {
		var l Line
		if err := rows.Scan(&l.LineID, &l.AccountID, &l.Code, &l.Name, &l.Type, 
			&l.DebitAmount, &l.CreditAmount); err != nil {
			log.Printf("Error scanning: %v", err)
			continue
		}
		lines = append(lines, l)
	}
	
	fmt.Printf("Found %d journal lines\n\n", len(lines))
	
	// Simulate manual balance update logic
	fmt.Println("2. Simulating manual balance update logic (line 275-310)...")
	fmt.Println("-----------------------------------------------------------------------")
	
	for i, line := range lines {
		fmt.Printf("\n[Line %d/%d] Processing account %s - %s (%s)\n", 
			i+1, len(lines), line.Code, line.Name, line.Type)
		
		// Get current balance from database
		var currentBalance float64
		err = db.QueryRow("SELECT balance FROM accounts WHERE id = $1", line.AccountID).Scan(&currentBalance)
		if err != nil {
			fmt.Printf("  ❌ ERROR: Failed to get current balance: %v\n", err)
			continue
		}
		
		fmt.Printf("  Current Balance: %.2f\n", currentBalance)
		fmt.Printf("  Debit: %.2f | Credit: %.2f\n", line.DebitAmount, line.CreditAmount)
		
		// Calculate balance change based on account type (same logic as closing service)
		var balanceChange float64
		
		if line.Type == "REVENUE" || line.Type == "EQUITY" {
			// Credit normal accounts: credit increases, debit decreases
			balanceChange = line.CreditAmount - line.DebitAmount
			fmt.Printf("  Balance Change Calculation: %.2f - %.2f = %.2f\n", 
				line.CreditAmount, line.DebitAmount, balanceChange)
		} else if line.Type == "EXPENSE" || line.Type == "ASSET" {
			// Debit normal accounts: debit increases, credit decreases
			balanceChange = line.DebitAmount - line.CreditAmount
			fmt.Printf("  Balance Change Calculation: %.2f - %.2f = %.2f\n", 
				line.DebitAmount, line.CreditAmount, balanceChange)
		} else {
			// LIABILITY (credit normal, like EQUITY)
			balanceChange = line.CreditAmount - line.DebitAmount
			fmt.Printf("  Balance Change Calculation: %.2f - %.2f = %.2f\n", 
				line.CreditAmount, line.DebitAmount, balanceChange)
		}
		
		expectedNewBalance := currentBalance + balanceChange
		fmt.Printf("  Expected New Balance: %.2f + %.2f = %.2f\n", 
			currentBalance, balanceChange, expectedNewBalance)
		
		// For revenue and expense, should be zero after closing
		if line.Type == "REVENUE" || line.Type == "EXPENSE" {
			if expectedNewBalance > -0.01 && expectedNewBalance < 0.01 {
				fmt.Printf("  ✓ Expected to be ZERO after closing (correct!)\n")
			} else {
				fmt.Printf("  ⚠️  Expected to be ZERO but calculated as %.2f\n", expectedNewBalance)
			}
		}
	}
	
	// Check current balances
	fmt.Println("\n\n3. Checking ACTUAL current account balances...")
	fmt.Println("-----------------------------------------------------------------------")
	
	for _, line := range lines {
		var currentBalance float64
		err = db.QueryRow("SELECT balance FROM accounts WHERE id = $1", line.AccountID).Scan(&currentBalance)
		if err != nil {
			fmt.Printf("  ❌ Account %s: ERROR - %v\n", line.Code, err)
			continue
		}
		
		fmt.Printf("  Account %s - %s: %.2f", line.Code, line.Name, currentBalance)
		
		if line.Type == "REVENUE" || line.Type == "EXPENSE" {
			if currentBalance > -0.01 && currentBalance < 0.01 {
				fmt.Printf(" ✓ (correct - should be 0)\n")
			} else {
				fmt.Printf(" ❌ (WRONG - should be 0!)\n")
			}
		} else {
			fmt.Printf("\n")
		}
	}
	
	// Diagnosis
	fmt.Println("\n\n========================================================================")
	fmt.Println("DIAGNOSIS:")
	fmt.Println("========================================================================")
	
	// Check if manual update would work
	hasNonZeroRevenue := false
	hasNonZeroExpense := false
	
	for _, line := range lines {
		var currentBalance float64
		db.QueryRow("SELECT balance FROM accounts WHERE id = $1", line.AccountID).Scan(&currentBalance)
		
		if line.Type == "REVENUE" && (currentBalance > 0.01 || currentBalance < -0.01) {
			hasNonZeroRevenue = true
		}
		if line.Type == "EXPENSE" && (currentBalance > 0.01 || currentBalance < -0.01) {
			hasNonZeroRevenue = true
		}
	}
	
	fmt.Println("\nProblem Analysis:")
	fmt.Println("1. Database triggers for auto-sync: DISABLED (to prevent double posting)")
	fmt.Println("2. Migration recalculation: FAILED (using wrong table names)")
	fmt.Println("3. Manual update in closing service: ???")
	
	if hasNonZeroRevenue || hasNonZeroExpense {
		fmt.Println("\n❌ CONFIRMED: Manual balance update in closing service DID NOT EXECUTE!")
		fmt.Println("\nPossible reasons:")
		fmt.Println("  a) Loop was skipped due to error/return before reaching line 275")
		fmt.Println("  b) GORM Update statement failed silently")
		fmt.Println("  c) Transaction was rolled back after creating closing entry")
		fmt.Println("  d) Update affected 0 rows (ID mismatch or constraint)")
		fmt.Println("\nSolution:")
		fmt.Println("  - Run: go run cmd/fix_closing_balances.go")
		fmt.Println("  - This will recalculate and fix all balances from unified_journal_lines")
	} else {
		fmt.Println("\n✓ Balances are correct! Manual update must have worked.")
	}
	
	fmt.Println("\n========================================================================")
}
