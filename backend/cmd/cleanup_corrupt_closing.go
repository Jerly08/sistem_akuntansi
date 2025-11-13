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
	// Load .env file if exists
	_ = godotenv.Load()

	// Connect to database
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
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("✅ Connected to database")
	log.Println("============================================")
	log.Println("Cleanup Script for Corrupt Period Closing")
	log.Println("============================================")

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// 1. Delete failed accounting periods
	log.Println("\n1️⃣  Deleting failed accounting periods...")
	result, err := tx.Exec(`
		DELETE FROM accounting_periods 
		WHERE is_closed = true 
		  AND closed_at >= '2025-11-13 00:00:00'
	`)
	if err != nil {
		log.Fatalf("Failed to delete accounting periods: %v", err)
	}
	rows, _ := result.RowsAffected()
	log.Printf("   ✅ Deleted %d failed accounting periods", rows)

	// 2. Delete corrupt closing journal entries
	log.Println("\n2️⃣  Deleting corrupt closing journal entries...")
	result, err = tx.Exec(`
		DELETE FROM unified_journal_lines 
		WHERE journal_id IN (
			SELECT id FROM unified_journal_ledger 
			WHERE source_type = 'CLOSING' 
			  AND created_at >= '2025-11-13 00:00:00'
		)
	`)
	if err != nil {
		log.Fatalf("Failed to delete journal lines: %v", err)
	}
	rows, _ = result.RowsAffected()
	log.Printf("   ✅ Deleted %d corrupt journal lines", rows)

	result, err = tx.Exec(`
		DELETE FROM unified_journal_ledger 
		WHERE source_type = 'CLOSING' 
		  AND created_at >= '2025-11-13 00:00:00'
	`)
	if err != nil {
		log.Fatalf("Failed to delete journal entries: %v", err)
	}
	rows, _ = result.RowsAffected()
	log.Printf("   ✅ Deleted %d corrupt journal entries", rows)

	// 3. Recalculate ALL account balances from POSTED journals
	log.Println("\n3️⃣  Recalculating ALL account balances from SSOT...")
	result, err = tx.Exec(`
		UPDATE accounts a
		SET balance = COALESCE((
			SELECT 
				CASE 
					WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
						COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
					ELSE 
						COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
				END
			FROM unified_journal_lines ujl
			INNER JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
			WHERE ujl.account_id = a.id 
			  AND uje.status = 'POSTED'
		), 0)
		WHERE a.deleted_at IS NULL
	`)
	if err != nil {
		log.Fatalf("Failed to recalculate balances: %v", err)
	}
	rows, _ = result.RowsAffected()
	log.Printf("   ✅ Recalculated %d account balances", rows)

	// 4. Verify: Check revenue and expense accounts
	log.Println("\n4️⃣  Verifying revenue and expense accounts...")
	verifyRows, err := tx.Query(`
		SELECT 
			code,
			name,
			type,
			balance,
			CASE 
				WHEN type = 'REVENUE' AND balance < 0 THEN 'OK - Credit balance'
				WHEN type = 'EXPENSE' AND balance > 0 THEN 'OK - Debit balance'
				WHEN type IN ('REVENUE', 'EXPENSE') AND ABS(balance) < 0.01 THEN 'ZERO - Ready for closing'
				ELSE '⚠️  WARNING - Unusual balance'
			END as status
		FROM accounts
		WHERE type IN ('REVENUE', 'EXPENSE')
		  AND is_header = false
		  AND deleted_at IS NULL
		ORDER BY type, code
	`)
	if err != nil {
		log.Fatalf("Failed to verify accounts: %v", err)
	}
	defer verifyRows.Close()

	log.Println("\n   Account Status:")
	log.Println("   " + "Code    | Name                        | Type    | Balance      | Status")
	log.Println("   " + "--------|-----------------------------|---------|--------------|--------------------------")
	
	revenueCount := 0
	expenseCount := 0
	for verifyRows.Next() {
		var code, name, accType, status string
		var balance float64
		if err := verifyRows.Scan(&code, &name, &accType, &balance, &status); err != nil {
			log.Printf("   Error scanning row: %v", err)
			continue
		}
		log.Printf("   %-7s | %-27s | %-7s | %12.2f | %s", code, name, accType, balance, status)
		
		if accType == "REVENUE" {
			revenueCount++
		} else if accType == "EXPENSE" {
			expenseCount++
		}
	}
	log.Printf("\n   📊 Total: %d Revenue accounts, %d Expense accounts", revenueCount, expenseCount)

	// 5. Verify: Show journal entries count by source type
	log.Println("\n5️⃣  Verifying journal entries...")
	journalRows, err := tx.Query(`
		SELECT 
			source_type,
			COUNT(*) as count,
			SUM(CASE WHEN status = 'POSTED' THEN 1 ELSE 0 END) as posted_count
		FROM unified_journal_ledger
		WHERE deleted_at IS NULL
		GROUP BY source_type
		ORDER BY source_type
	`)
	if err != nil {
		log.Fatalf("Failed to verify journals: %v", err)
	}
	defer journalRows.Close()

	log.Println("\n   Journal Entry Summary:")
	log.Println("   " + "Source Type     | Total | Posted")
	log.Println("   " + "----------------|-------|-------")
	for journalRows.Next() {
		var sourceType string
		var count, postedCount int
		if err := journalRows.Scan(&sourceType, &count, &postedCount); err != nil {
			log.Printf("   Error scanning row: %v", err)
			continue
		}
		log.Printf("   %-15s | %5d | %6d", sourceType, count, postedCount)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	log.Println("\n============================================")
	log.Println("✅ Cleanup completed successfully!")
	log.Println("============================================")
	log.Println("\n📝 Next steps:")
	log.Println("   1. Restart backend server")
	log.Println("   2. Try period closing again from UI")
	log.Println("   3. Verify the closing completes without errors")
}
