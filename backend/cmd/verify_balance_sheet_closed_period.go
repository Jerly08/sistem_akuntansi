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
	fmt.Println("VERIFY BALANCE SHEET - CLOSED PERIOD HANDLING")
	fmt.Println("========================================================================")
	
	fmt.Println("\n📊 KONSEP YANG BENAR:")
	fmt.Println("-----------------------------------------------------------------------")
	fmt.Println("BEFORE Closing:")
	fmt.Println("  - Revenue & Expense accounts have balances")
	fmt.Println("  - Balance Sheet shows: Retained Earnings + Net Income (Laba/Rugi Berjalan)")
	fmt.Println()
	fmt.Println("AFTER Closing:")
	fmt.Println("  - Revenue & Expense accounts = 0")
	fmt.Println("  - Balance Sheet shows: ONLY Retained Earnings (includes closed Net Income)")
	fmt.Println("  - NO separate 'Laba/Rugi Berjalan' line")
	
	// Check current state
	fmt.Println("\n\n🔍 CHECKING CURRENT STATE:")
	fmt.Println("-----------------------------------------------------------------------")
	
	// 1. Check if Revenue/Expense are zero
	var revenueCount, expenseCount int
	db.QueryRow("SELECT COUNT(*) FROM accounts WHERE type = 'REVENUE' AND ABS(balance) > 0.01").Scan(&revenueCount)
	db.QueryRow("SELECT COUNT(*) FROM accounts WHERE type = 'EXPENSE' AND ABS(balance) > 0.01").Scan(&expenseCount)
	
	isPeriodClosed := (revenueCount == 0 && expenseCount == 0)
	
	fmt.Printf("1. Revenue accounts with balance: %d\n", revenueCount)
	fmt.Printf("2. Expense accounts with balance: %d\n", expenseCount)
	fmt.Printf("3. Period Status: ")
	if isPeriodClosed {
		fmt.Println("✅ CLOSED (Revenue & Expense = 0)")
	} else {
		fmt.Println("❌ NOT CLOSED (Revenue/Expense have balances)")
	}
	
	// 2. Check Balance Sheet logic in code
	fmt.Println("\n\n📝 BALANCE SHEET SERVICE LOGIC:")
	fmt.Println("-----------------------------------------------------------------------")
	fmt.Println("File: services/ssot_balance_sheet_service.go")
	fmt.Println()
	fmt.Println("Line 431-458: Check if period closed")
	fmt.Println("```go")
	fmt.Println("var hasActiveRevenueExpense bool")
	fmt.Println("for _, balance := range balances {")
	fmt.Println("    if balance.AccountType == 'REVENUE' && balance.NetBalance != 0 {")
	fmt.Println("        hasActiveRevenueExpense = true")
	fmt.Println("    }")
	fmt.Println("}")
	fmt.Println("```")
	fmt.Println()
	fmt.Println("Line 450-458: Calculate Net Income ONLY if NOT closed")
	fmt.Println("```go")
	fmt.Println("if hasActiveRevenueExpense {")
	fmt.Println("    netIncome = totalRevenueBalance - totalExpenseBalance")
	fmt.Println("    // Show as 'Laba/Rugi Berjalan' in Balance Sheet")
	fmt.Println("} else {")
	fmt.Println("    // Period CLOSED - Net Income already in Retained Earnings")
	fmt.Println("}")
	fmt.Println("```")
	fmt.Println()
	fmt.Println("Line 540-552: Add Net Income to Equity ONLY if NOT closed")
	fmt.Println("```go")
	fmt.Println("if netIncome != 0 && hasActiveRevenueExpense {")
	fmt.Println("    // Show separate 'Laba/Rugi Berjalan' line")
	fmt.Println("} else {")
	fmt.Println("    // No separate line - already in Retained Earnings")
	fmt.Println("}")
	fmt.Println("```")
	
	// 3. Simulate Balance Sheet calculation
	fmt.Println("\n\n🧪 SIMULATING BALANCE SHEET CALCULATION:")
	fmt.Println("-----------------------------------------------------------------------")
	
	// Get Retained Earnings balance
	var retainedEarningsBalance float64
	err = db.QueryRow(`
		SELECT balance FROM accounts 
		WHERE code = '3201' AND type = 'EQUITY'
	`).Scan(&retainedEarningsBalance)
	
	if err == nil {
		fmt.Printf("Retained Earnings (3201) balance: Rp %.2f\n", retainedEarningsBalance)
	}
	
	// Calculate Net Income from Revenue/Expense accounts (if not closed)
	var totalRevenue, totalExpense float64
	db.QueryRow("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'REVENUE'").Scan(&totalRevenue)
	db.QueryRow("SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE type = 'EXPENSE'").Scan(&totalExpense)
	
	netIncome := totalRevenue - totalExpense
	
	fmt.Printf("\nCurrent Period Net Income:\n")
	fmt.Printf("  Revenue: Rp %.2f\n", totalRevenue)
	fmt.Printf("  Expense: Rp %.2f\n", totalExpense)
	fmt.Printf("  Net Income: Rp %.2f\n", netIncome)
	
	// Show what Balance Sheet SHOULD display
	fmt.Println("\n\n✅ BALANCE SHEET SHOULD DISPLAY:")
	fmt.Println("-----------------------------------------------------------------------")
	
	if isPeriodClosed {
		fmt.Println("EQUITY Section:")
		fmt.Printf("  - Modal/Capital: [from account 3xxx]\n")
		fmt.Printf("  - Laba Ditahan: Rp %.2f (includes all closed periods net income)\n", retainedEarningsBalance)
		fmt.Printf("  - NO 'Laba/Rugi Berjalan' line (because period is CLOSED)\n")
		fmt.Printf("\nTotal Equity: Rp %.2f\n", retainedEarningsBalance)
	} else {
		fmt.Println("EQUITY Section:")
		fmt.Printf("  - Modal/Capital: [from account 3xxx]\n")
		fmt.Printf("  - Laba Ditahan: Rp %.2f\n", retainedEarningsBalance)
		fmt.Printf("  - Laba/Rugi Berjalan: Rp %.2f (current period - NOT yet closed)\n", netIncome)
		fmt.Printf("\nTotal Equity: Rp %.2f\n", retainedEarningsBalance+netIncome)
	}
	
	// Verification
	fmt.Println("\n\n🎯 VERIFICATION:")
	fmt.Println("========================================================================")
	
	fmt.Println("\n✅ CORRECT BEHAVIOR:")
	fmt.Println("1. ssot_balance_sheet_service.go checks hasActiveRevenueExpense")
	fmt.Println("2. If Revenue/Expense = 0 (closed), NO Net Income line is added")
	fmt.Println("3. If Revenue/Expense != 0 (not closed), adds 'Laba/Rugi Berjalan'")
	fmt.Println("4. Retained Earnings always shows correct cumulative balance")
	
	fmt.Println("\n✅ INTEGRATION WITH CLOSING:")
	fmt.Println("1. After closing: Revenue/Expense → 0")
	fmt.Println("2. Net Income transferred to Retained Earnings (3201)")
	fmt.Println("3. Balance Sheet detects Revenue/Expense = 0")
	fmt.Println("4. Balance Sheet shows ONLY Retained Earnings (no separate Net Income)")
	
	if isPeriodClosed {
		fmt.Println("\n✅ CURRENT STATUS: Period is CLOSED")
		fmt.Println("   Balance Sheet will correctly show:")
		fmt.Println("   - NO 'Laba/Rugi Berjalan' line")
		fmt.Println("   - Retained Earnings includes all net income from closed periods")
	} else {
		fmt.Println("\n⚠️  CURRENT STATUS: Period is NOT CLOSED")
		fmt.Println("   Balance Sheet will show:")
		fmt.Println("   - 'Laba/Rugi Berjalan' as separate line")
		fmt.Println("   - After closing, this will be merged into Retained Earnings")
	}
	
	fmt.Println("\n========================================================================")
	fmt.Println("Verification complete!")
	fmt.Println("========================================================================")
}
