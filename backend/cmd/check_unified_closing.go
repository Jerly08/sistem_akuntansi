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

	fmt.Println("=" + string(make([]byte, 70)))
	fmt.Println("ANALISIS CLOSING PERIOD - UNIFIED JOURNAL SYSTEM")
	fmt.Println("=" + string(make([]byte, 70)))

	// 1. Check unified_journal_ledger for CLOSING entries
	fmt.Println("\n1. CHECKING UNIFIED_JOURNAL_LEDGER FOR CLOSING ENTRIES:")
	fmt.Println("-" + string(make([]byte, 70)))
	
	query := `
		SELECT id, source_type, entry_date, description, total_debit, total_credit, status, created_at
		FROM unified_journal_ledger
		WHERE source_type = 'CLOSING'
		ORDER BY entry_date DESC
		LIMIT 10
	`
	
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Error querying unified_journal_ledger: %v", err)
	} else {
		defer rows.Close()
		
		count := 0
		for rows.Next() {
			var id int
			var sourceType, description, status string
			var entryDate time.Time
			var totalDebit, totalCredit float64
			var createdAt time.Time
			
			err := rows.Scan(&id, &sourceType, &entryDate, &description, &totalDebit, &totalCredit, &status, &createdAt)
			if err != nil {
				log.Printf("Error scanning row: %v", err)
				continue
			}
			
			fmt.Printf("\n✓ Closing Entry Found:\n")
			fmt.Printf("  ID: %d | Source: %s | Status: %s\n", id, sourceType, status)
			fmt.Printf("  Date: %s | Created: %s\n", entryDate.Format("2006-01-02"), createdAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Description: %s\n", description)
			fmt.Printf("  Total Debit: Rp %.2f | Total Credit: Rp %.2f\n", totalDebit, totalCredit)
			count++
		}
		
		if count == 0 {
			fmt.Println("\n⚠️  NO CLOSING ENTRIES FOUND in unified_journal_ledger!")
		} else {
			fmt.Printf("\n✓ Found %d closing entries\n", count)
		}
	}

	// 2. Check accounting_periods table
	fmt.Println("\n\n2. CHECKING ACCOUNTING_PERIODS TABLE:")
	fmt.Println("-" + string(make([]byte, 70)))
	
	query = `
		SELECT id, start_date, end_date, description, is_closed, is_locked, 
		       total_revenue, total_expense, net_income, closed_at
		FROM accounting_periods
		WHERE is_closed = true
		ORDER BY end_date DESC
		LIMIT 10
	`
	
	rows, err = db.Query(query)
	if err != nil {
		log.Printf("Error querying accounting_periods: %v", err)
	} else {
		defer rows.Close()
		
		count := 0
		for rows.Next() {
			var id int
			var description string
			var startDate, endDate time.Time
			var isClosed, isLocked bool
			var totalRevenue, totalExpense, netIncome float64
			var closedAt sql.NullTime
			
			err := rows.Scan(&id, &startDate, &endDate, &description, &isClosed, &isLocked, 
				&totalRevenue, &totalExpense, &netIncome, &closedAt)
			if err != nil {
				log.Printf("Error scanning: %v", err)
				continue
			}
			
			closedAtStr := "N/A"
			if closedAt.Valid {
				closedAtStr = closedAt.Time.Format("2006-01-02 15:04:05")
			}
			
			fmt.Printf("\n✓ Closed Period:\n")
			fmt.Printf("  ID: %d | Period: %s to %s\n", id, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
			fmt.Printf("  Description: %s\n", description)
			fmt.Printf("  Closed: %v | Locked: %v | Closed At: %s\n", isClosed, isLocked, closedAtStr)
			fmt.Printf("  Revenue: Rp %.2f | Expense: Rp %.2f | Net Income: Rp %.2f\n", 
				totalRevenue, totalExpense, netIncome)
			count++
		}
		
		if count == 0 {
			fmt.Println("\n⚠️  NO CLOSED PERIODS FOUND!")
		} else {
			fmt.Printf("\n✓ Found %d closed periods\n", count)
		}
	}

	// 3. Check REVENUE accounts with non-zero balance
	fmt.Println("\n\n3. REVENUE ACCOUNTS WITH NON-ZERO BALANCE:")
	fmt.Println("-" + string(make([]byte, 70)))
	
	query = `
		SELECT id, code, name, type, balance
		FROM accounts
		WHERE type = 'REVENUE' AND ABS(balance) > 0.01
		ORDER BY code
		LIMIT 20
	`
	
	rows, err = db.Query(query)
	if err != nil {
		log.Printf("Error querying revenue accounts: %v", err)
	} else {
		defer rows.Close()
		
		count := 0
		totalBalance := 0.0
		for rows.Next() {
			var id int
			var code, name, accountType string
			var balance float64
			
			err := rows.Scan(&id, &code, &name, &accountType, &balance)
			if err != nil {
				log.Printf("Error scanning: %v", err)
				continue
			}
			
			fmt.Printf("  %s - %s: Rp %.2f\n", code, name, balance)
			totalBalance += balance
			count++
		}
		
		if count == 0 {
			fmt.Println("  ✓ All revenue accounts have zero balance (CORRECT after closing)")
		} else {
			fmt.Printf("\n  ⚠️  %d revenue accounts still have balance!\n", count)
			fmt.Printf("  Total Revenue Balance: Rp %.2f (should be 0 after closing)\n", totalBalance)
		}
	}

	// 4. Check EXPENSE accounts with non-zero balance
	fmt.Println("\n\n4. EXPENSE ACCOUNTS WITH NON-ZERO BALANCE:")
	fmt.Println("-" + string(make([]byte, 70)))
	
	query = `
		SELECT id, code, name, type, balance
		FROM accounts
		WHERE type = 'EXPENSE' AND ABS(balance) > 0.01
		ORDER BY code
		LIMIT 20
	`
	
	rows, err = db.Query(query)
	if err != nil {
		log.Printf("Error querying expense accounts: %v", err)
	} else {
		defer rows.Close()
		
		count := 0
		totalBalance := 0.0
		for rows.Next() {
			var id int
			var code, name, accountType string
			var balance float64
			
			err := rows.Scan(&id, &code, &name, &accountType, &balance)
			if err != nil {
				log.Printf("Error scanning: %v", err)
				continue
			}
			
			fmt.Printf("  %s - %s: Rp %.2f\n", code, name, balance)
			totalBalance += balance
			count++
		}
		
		if count == 0 {
			fmt.Println("  ✓ All expense accounts have zero balance (CORRECT after closing)")
		} else {
			fmt.Printf("\n  ⚠️  %d expense accounts still have balance!\n", count)
			fmt.Printf("  Total Expense Balance: Rp %.2f (should be 0 after closing)\n", totalBalance)
		}
	}

	// 5. Check lines for the closing entry
	fmt.Println("\n\n5. CHECKING CLOSING JOURNAL LINES (if closing entry exists):")
	fmt.Println("-" + string(make([]byte, 70)))
	
	query = `
		SELECT ujl.id, ujl.journal_id, ujl.account_id, a.code, a.name, a.type,
		       ujl.debit_amount, ujl.credit_amount, ujl.description
		FROM unified_journal_lines ujl
		JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		JOIN accounts a ON a.id = ujl.account_id
		WHERE uje.source_type = 'CLOSING'
		ORDER BY uje.entry_date DESC, ujl.line_number
		LIMIT 50
	`
	
	rows, err = db.Query(query)
	if err != nil {
		log.Printf("Error querying closing lines: %v", err)
	} else {
		defer rows.Close()
		
		count := 0
		for rows.Next() {
			var lineID, journalID, accountID int
			var code, name, accountType, description string
			var debitAmount, creditAmount float64
			
			err := rows.Scan(&lineID, &journalID, &accountID, &code, &name, &accountType, 
				&debitAmount, &creditAmount, &description)
			if err != nil {
				log.Printf("Error scanning: %v", err)
				continue
			}
			
			fmt.Printf("\n  Line ID: %d | Journal ID: %d\n", lineID, journalID)
			fmt.Printf("  Account: %s - %s (%s)\n", code, name, accountType)
			fmt.Printf("  Debit: Rp %.2f | Credit: Rp %.2f\n", debitAmount, creditAmount)
			fmt.Printf("  Description: %s\n", description)
			count++
		}
		
		if count == 0 {
			fmt.Println("  ⚠️  NO CLOSING JOURNAL LINES FOUND!")
		} else {
			fmt.Printf("\n✓ Found %d closing journal lines\n", count)
		}
	}

	// 6. Summary and Analysis
	fmt.Println("\n\n" + string(make([]byte, 70)))
	fmt.Println("SUMMARY & DIAGNOSIS:")
	fmt.Println(string(make([]byte, 70)))
	
	// Check if closing was executed
	var hasClosingEntry bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM unified_journal_ledger WHERE source_type = 'CLOSING')").Scan(&hasClosingEntry)
	
	var hasClosedPeriod bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM accounting_periods WHERE is_closed = true)").Scan(&hasClosedPeriod)
	
	var revenueCount, expenseCount int
	db.QueryRow("SELECT COUNT(*) FROM accounts WHERE type = 'REVENUE' AND ABS(balance) > 0.01").Scan(&revenueCount)
	db.QueryRow("SELECT COUNT(*) FROM accounts WHERE type = 'EXPENSE' AND ABS(balance) > 0.01").Scan(&expenseCount)
	
	fmt.Println("\n✓ Status Check:")
	fmt.Printf("  - Closing Entry Created: %v\n", hasClosingEntry)
	fmt.Printf("  - Period Marked as Closed: %v\n", hasClosedPeriod)
	fmt.Printf("  - Revenue Accounts with Balance: %d (should be 0)\n", revenueCount)
	fmt.Printf("  - Expense Accounts with Balance: %d (should be 0)\n", expenseCount)
	
	fmt.Println("\n⚠️  DIAGNOSIS:")
	if hasClosingEntry && hasClosedPeriod {
		if revenueCount > 0 || expenseCount > 0 {
			fmt.Println("  ❌ PROBLEM DETECTED: Closing entry was created BUT account balances were NOT updated!")
			fmt.Println("  ")
			fmt.Println("  POSSIBLE CAUSES:")
			fmt.Println("  1. Account balance update logic in closing service failed")
			fmt.Println("  2. Transaction was rolled back after creating closing entry")
			fmt.Println("  3. Database triggers for balance sync are not working")
			fmt.Println("  ")
			fmt.Println("  SOLUTION:")
			fmt.Println("  - Check unified_period_closing_service.go line 275-310 (balance update logic)")
			fmt.Println("  - Verify database triggers for account balance synchronization")
			fmt.Println("  - Consider running manual balance recalculation")
		} else {
			fmt.Println("  ✓ All good! Closing entry created and balances are zeroed correctly.")
		}
	} else if !hasClosingEntry && !hasClosedPeriod {
		fmt.Println("  ⚠️  No closing has been performed yet.")
	} else if hasClosedPeriod && !hasClosingEntry {
		fmt.Println("  ❌ CRITICAL: Period marked as closed but NO closing journal entry exists!")
		fmt.Println("  This indicates incomplete closing process.")
	}
	
	fmt.Println("\n" + string(make([]byte, 70)))
	fmt.Println("Diagnostic complete!")
	fmt.Println(string(make([]byte, 70)))
}
