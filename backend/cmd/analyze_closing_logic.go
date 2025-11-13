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
	fmt.Println("ANALISIS LOGIKA CLOSING - MASALAH KONSEPTUAL & SOLUSI")
	fmt.Println("========================================================================")
	
	fmt.Println("\n📚 KONSEP CLOSING YANG BENAR:")
	fmt.Println("-----------------------------------------------------------------------")
	fmt.Println("1. Closing adalah proses memindahkan saldo TEMPORARY accounts (Revenue & Expense)")
	fmt.Println("   ke PERMANENT account (Retained Earnings) di akhir periode")
	fmt.Println()
	fmt.Println("2. SETELAH closing:")
	fmt.Println("   - Semua Revenue accounts = 0")
	fmt.Println("   - Semua Expense accounts = 0") 
	fmt.Println("   - Retained Earnings bertambah/berkurang sebesar Net Income")
	fmt.Println()
	fmt.Println("3. Journal Entry Closing:")
	fmt.Println("   a) Debit semua Revenue accounts (untuk zero-kan credit balance)")
	fmt.Println("   b) Credit Retained Earnings sebesar total Revenue")
	fmt.Println("   c) Credit semua Expense accounts (untuk zero-kan debit balance)")
	fmt.Println("   d) Debit Retained Earnings sebesar total Expense")
	
	// Analyze current situation
	fmt.Println("\n\n🔍 ANALISIS SITUASI SAAT INI:")
	fmt.Println("-----------------------------------------------------------------------")
	
	// Check calculation method
	fmt.Println("\n1. Metode Perhitungan Balance di Closing Service:")
	fmt.Println("   ❌ MASALAH: Menggunakan balance DALAM PERIODE (periodStartDate to endDate)")
	fmt.Println("   ✅ SEHARUSNYA: Menggunakan balance KUMULATIF dari awal fiscal year")
	
	// Query to check what balance method is being used
	var lastClosingID int
	err = db.QueryRow(`
		SELECT id FROM unified_journal_ledger 
		WHERE source_type = 'CLOSING' 
		ORDER BY entry_date DESC LIMIT 1
	`).Scan(&lastClosingID)
	
	if err == nil {
		fmt.Printf("\n   Last Closing Journal ID: %d\n", lastClosingID)
		
		// Check the amounts in closing entry
		rows, err := db.Query(`
			SELECT a.code, a.name, a.type, ujl.debit_amount, ujl.credit_amount
			FROM unified_journal_lines ujl
			JOIN accounts a ON a.id = ujl.account_id
			WHERE ujl.journal_id = $1
			ORDER BY ujl.line_number
		`, lastClosingID)
		
		if err == nil {
			defer rows.Close()
			fmt.Println("\n   Closing Entry Lines:")
			for rows.Next() {
				var code, name, accType string
				var debit, credit float64
				rows.Scan(&code, &name, &accType, &debit, &credit)
				fmt.Printf("   %s (%s): Debit=%.2f Credit=%.2f\n", code, accType, debit, credit)
			}
		}
	}
	
	// Check actual cumulative balances
	fmt.Println("\n2. Checking ACTUAL Cumulative Balances (What should be closed):")
	fmt.Println("   ---------------------------------------------------------------")
	
	// Revenue cumulative balance
	var revenueTotalDebit, revenueTotalCredit float64
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
	`).Scan(&revenueTotalDebit, &revenueTotalCredit)
	
	if err == nil {
		revenueBalance := revenueTotalCredit - revenueTotalDebit
		fmt.Printf("   REVENUE Cumulative: Debit=%.2f Credit=%.2f Balance=%.2f\n", 
			revenueTotalDebit, revenueTotalCredit, revenueBalance)
	}
	
	// Expense cumulative balance
	var expenseTotalDebit, expenseTotalCredit float64
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
	`).Scan(&expenseTotalDebit, &expenseTotalCredit)
	
	if err == nil {
		expenseBalance := expenseTotalDebit - expenseTotalCredit
		fmt.Printf("   EXPENSE Cumulative: Debit=%.2f Credit=%.2f Balance=%.2f\n", 
			expenseTotalDebit, expenseTotalCredit, expenseBalance)
	}
	
	// Analysis of problem
	fmt.Println("\n\n❌ ROOT CAUSE PROBLEMS IDENTIFIED:")
	fmt.Println("========================================================================")
	
	fmt.Println("\nPROBLEM #1: Period vs Cumulative Confusion")
	fmt.Println("------------------------------------------")
	fmt.Println("Kode closing menggunakan periodStartDate sampai endDate untuk menghitung balance.")
	fmt.Println("Ini SALAH jika ada multiple closing dalam setahun.")
	fmt.Println("SOLUSI: Selalu gunakan CUMULATIVE balance dari semua transaksi (exclude CLOSING entries)")
	
	fmt.Println("\nPROBLEM #2: Balance Update Method")
	fmt.Println("----------------------------------")
	fmt.Println("Kode lama mencoba update dengan delta (+/-) yang rawan error.")
	fmt.Println("SOLUSI: Recalculate exact balance dari unified_journal_lines")
	
	fmt.Println("\nPROBLEM #3: No Validation After Closing")
	fmt.Println("----------------------------------------")
	fmt.Println("Tidak ada validasi apakah Revenue/Expense sudah 0 setelah closing.")
	fmt.Println("SOLUSI: Add validation step untuk ensure semua temporary accounts = 0")
	
	// Proposed solution
	fmt.Println("\n\n✅ PROPOSED SOLUTION:")
	fmt.Println("========================================================================")
	
	fmt.Println(`
1. UBAH QUERY BALANCE (Line 98-137):
   Hapus filter periodStartDate, gunakan ALL transactions except CLOSING

2. UBAH BALANCE UPDATE (Line 270-313):
   Ganti dengan recalculation exact dari unified_journal_lines

3. TAMBAH VALIDATION:
   Setelah update balance, verify Revenue & Expense = 0

4. CORRECT CLOSING LOGIC:
   
   // Step 1: Calculate CUMULATIVE balances (all time, exclude closing)
   revenue_balance = SUM(credit) - SUM(debit) WHERE type='REVENUE' AND source!='CLOSING'
   expense_balance = SUM(debit) - SUM(credit) WHERE type='EXPENSE' AND source!='CLOSING'
   
   // Step 2: Create closing journal
   Debit: Revenue accounts    amount: revenue_balance
   Credit: Retained Earnings  amount: revenue_balance
   
   Debit: Retained Earnings   amount: expense_balance  
   Credit: Expense accounts   amount: expense_balance
   
   // Step 3: Update account balances to EXACT calculated values
   UPDATE accounts SET balance = 0 WHERE type IN ('REVENUE','EXPENSE')
   UPDATE accounts SET balance = calculated_from_journal WHERE code = '3201'
`)

	fmt.Println("\n========================================================================")
	fmt.Println("RECOMMENDATION: Implement perbaikan di unified_period_closing_service.go")
	fmt.Println("========================================================================")
}