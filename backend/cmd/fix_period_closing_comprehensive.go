package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	_ = godotenv.Load()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		host := getEnvOrDefault("DB_HOST", "localhost")
		port := getEnvOrDefault("DB_PORT", "5432")
		user := getEnvOrDefault("DB_USER", "postgres")
		password := getEnvOrDefault("DB_PASSWORD", "postgres")
		dbname := getEnvOrDefault("DB_NAME", "sistem_akuntansi")
		sslmode := getEnvOrDefault("DB_SSLMODE", "disable")
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			user, password, host, port, dbname, sslmode)
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	log.Println("============================================")
	log.Println("🔧 COMPREHENSIVE PERIOD CLOSING FIX")
	log.Println("============================================")

	// Step 1: Analyze current state
	log.Println("\n📊 STEP 1: Analyzing current state...")
	
	// Check for PAYMENT journal entries affecting revenue
	var paymentCount int
	err = db.QueryRow(`
		SELECT COUNT(DISTINCT uje.id)
		FROM unified_journal_lines ujl
		INNER JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		INNER JOIN accounts a ON a.id = ujl.account_id
		WHERE a.type = 'REVENUE'
		  AND uje.source_type = 'PAYMENT'
		  AND uje.status = 'POSTED'
	`).Scan(&paymentCount)
	
	if err != nil {
		log.Printf("Error checking payment entries: %v", err)
		paymentCount = 0
	}
	
	if paymentCount > 0 {
		log.Printf("   ⚠️  Found %d PAYMENT entries affecting REVENUE accounts", paymentCount)
		log.Println("   These are likely incorrect and need to be fixed")
	}

	// Step 2: Clean up bad entries
	log.Println("\n🗑️  STEP 2: Cleaning up incorrect entries...")
	
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Delete ALL closing entries
	result, err := tx.Exec(`
		DELETE FROM unified_journal_lines 
		WHERE journal_id IN (
			SELECT id FROM unified_journal_ledger WHERE source_type = 'CLOSING'
		)
	`)
	if err != nil {
		log.Fatalf("Failed to delete closing lines: %v", err)
	}
	rows1, _ := result.RowsAffected()
	
	result, err = tx.Exec(`DELETE FROM unified_journal_ledger WHERE source_type = 'CLOSING'`)
	if err != nil {
		log.Fatalf("Failed to delete closing entries: %v", err)
	}
	rows2, _ := result.RowsAffected()
	
	result, err = tx.Exec(`DELETE FROM accounting_periods`)
	if err != nil {
		log.Fatalf("Failed to delete periods: %v", err)
	}
	rows3, _ := result.RowsAffected()
	
	log.Printf("   ✅ Deleted %d closing lines, %d closing entries, %d periods", rows1, rows2, rows3)

	// Check for problematic PAYMENT entries
	log.Println("\n   Checking for problematic PAYMENT entries...")
	rows, err := tx.Query(`
		SELECT uje.id, uje.entry_date, ujl.debit_amount, ujl.credit_amount, a.code, a.name
		FROM unified_journal_lines ujl
		INNER JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
		INNER JOIN accounts a ON a.id = ujl.account_id
		WHERE a.type = 'REVENUE'
		  AND uje.source_type = 'PAYMENT'
		  AND ujl.debit_amount > 0
		ORDER BY uje.id
	`)
	if err != nil {
		log.Printf("   Error checking payment entries: %v", err)
	} else {
		defer rows.Close()
		problemCount := 0
		for rows.Next() {
			var id int
			var date string
			var debit, credit float64
			var code, name string
			rows.Scan(&id, &date, &debit, &credit, &code, &name)
			if debit > 0 {
				log.Printf("   ⚠️  Journal ID %d: DEBIT %.2f to REVENUE account %s (%s)", id, debit, code, name)
				problemCount++
			}
		}
		if problemCount > 0 {
			log.Printf("   Found %d problematic PAYMENT entries debiting revenue accounts", problemCount)
			log.Println("   These should be reviewed - payments shouldn't debit revenue directly")
		}
	}

	// Step 3: Recalculate ALL balances correctly
	log.Println("\n🔄 STEP 3: Recalculating ALL account balances...")
	
	result, err = tx.Exec(`
		UPDATE accounts a
		SET balance = COALESCE((
			SELECT 
				-- For ALL account types: Debit - Credit
				-- This gives positive for debit balances, negative for credit balances
				COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
			FROM unified_journal_lines ujl
			INNER JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
			WHERE ujl.account_id = a.id 
			  AND uje.status = 'POSTED'
		), 0)
		WHERE a.deleted_at IS NULL
	`)
	if err != nil {
		log.Fatalf("Failed to recalculate: %v", err)
	}
	rowsAffected, _ := result.RowsAffected()
	log.Printf("   ✅ Recalculated %d account balances", rowsAffected)

	// Step 4: Verify balances
	log.Println("\n✅ STEP 4: Verifying final balances...")
	
	verifyRows, err := tx.Query(`
		SELECT code, name, type, balance
		FROM accounts
		WHERE type IN ('REVENUE', 'EXPENSE')
		  AND is_header = false
		  AND ABS(balance) > 0.01
		  AND deleted_at IS NULL
		ORDER BY type, code
	`)
	if err != nil {
		log.Fatalf("Failed to verify: %v", err)
	}
	defer verifyRows.Close()

	log.Println("\n   Code    | Name                        | Type    | Balance       | Status")
	log.Println("   --------|-----------------------------|---------|--------------|---------")
	
	errorCount := 0
	for verifyRows.Next() {
		var code, name, accType string
		var balance float64
		verifyRows.Scan(&code, &name, &accType, &balance)
		
		status := "✅ OK"
		if accType == "REVENUE" && balance > 0 {
			status = "❌ WRONG SIGN"
			errorCount++
		} else if accType == "EXPENSE" && balance < 0 {
			status = "❌ WRONG SIGN"
			errorCount++
		}
		
		log.Printf("   %-7s | %-27s | %-7s | %13.2f | %s", code, name, accType, balance, status)
	}

	if errorCount > 0 {
		log.Printf("\n   ❌ Found %d accounts with wrong sign", errorCount)
		tx.Rollback()
		log.Println("\n❌ Fix aborted - balances still have wrong signs")
		log.Println("This indicates the backend is using OLD CODE")
		log.Println("\n📝 REQUIRED ACTIONS:")
		log.Println("   1. STOP the backend server (Ctrl+C)")
		log.Println("   2. Pull latest code: git pull origin main")
		log.Println("   3. Check file backend/services/unified_period_closing_service.go")
		log.Println("   4. Make sure line ~143 uses: decimal.NewFromFloat(bal.TotalDebit).Sub(decimal.NewFromFloat(bal.TotalCredit))")
		log.Println("   5. Restart backend: go run main.go")
		log.Println("   6. Run this script again")
		return
	}

	// Commit if everything is OK
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit: %v", err)
	}

	log.Println("\n============================================")
	log.Println("✅ Comprehensive fix completed successfully!")
	log.Println("============================================")

	// Final check
	log.Println("\n🔍 Final validation...")
	row := db.QueryRow(`
		SELECT COUNT(*) 
		FROM accounts 
		WHERE type = 'REVENUE' 
		  AND balance > 0.01 
		  AND is_header = false
		  AND deleted_at IS NULL
	`)
	
	var wrongRevCount int
	row.Scan(&wrongRevCount)
	
	if wrongRevCount > 0 {
		log.Printf("⚠️  WARNING: %d revenue accounts still have positive balance", wrongRevCount)
		log.Println("Please ensure backend is using latest code!")
	} else {
		log.Println("✅ All revenue accounts have correct (negative) balances")
		log.Println("\n🎯 Ready for period closing!")
	}
}