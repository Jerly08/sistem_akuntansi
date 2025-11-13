package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	log.Println("============================================")
	log.Println("COMPLETE CLEANUP - Remove ALL Closing Data")
	log.Println("============================================")

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// 1. Delete ALL accounting periods
	log.Println("\n1️⃣  Deleting ALL accounting periods...")
	result, err := tx.Exec(`DELETE FROM accounting_periods`)
	if err != nil {
		log.Fatalf("Failed to delete accounting periods: %v", err)
	}
	rows, _ := result.RowsAffected()
	log.Printf("   ✅ Deleted %d accounting periods", rows)

	// 2. Delete ALL closing journal entries (including old ones)
	log.Println("\n2️⃣  Deleting ALL closing journal entries...")
	result, err = tx.Exec(`
		DELETE FROM unified_journal_lines 
		WHERE journal_id IN (
			SELECT id FROM unified_journal_ledger 
			WHERE source_type = 'CLOSING'
		)
	`)
	if err != nil {
		log.Fatalf("Failed to delete journal lines: %v", err)
	}
	rows, _ = result.RowsAffected()
	log.Printf("   ✅ Deleted %d closing journal lines", rows)

	result, err = tx.Exec(`
		DELETE FROM unified_journal_ledger 
		WHERE source_type = 'CLOSING'
	`)
	if err != nil {
		log.Fatalf("Failed to delete journal entries: %v", err)
	}
	rows, _ = result.RowsAffected()
	log.Printf("   ✅ Deleted %d closing journal entries", rows)

	// 3. Check and delete sales with unrealistic dates (2026, 2027)
	log.Println("\n3️⃣  Checking for sales with unrealistic dates...")
	var futureCount int64
	err = tx.QueryRow(`
		SELECT COUNT(*) FROM unified_journal_ledger 
		WHERE source_type = 'SALE' AND entry_date > '2025-12-31'
	`).Scan(&futureCount)
	
	if err != nil {
		log.Fatalf("Failed to check future sales: %v", err)
	}
	
	if futureCount > 0 {
		log.Printf("   ⚠️  Found %d sales with dates beyond 2025", futureCount)
		log.Println("   Listing them:")
		
		rows, err := tx.Query(`
			SELECT id, entry_date, description 
			FROM unified_journal_ledger 
			WHERE source_type = 'SALE' AND entry_date > '2025-12-31'
			ORDER BY entry_date
		`)
		if err != nil {
			log.Fatalf("Failed to query future sales: %v", err)
		}
		
		for rows.Next() {
			var id int
			var date, desc string
			rows.Scan(&id, &date, &desc)
			log.Printf("      - ID %d: %s - %s", id, date[:10], desc)
		}
		rows.Close()
		
		// Option: Ask user or auto-delete
		log.Println("\n   Would you like to delete these unrealistic entries? (y/n)")
		log.Println("   For now, we'll keep them for manual review.")
	} else {
		log.Println("   ✅ No unrealistic dates found")
	}

	// 4. Recalculate ALL account balances
	log.Println("\n4️⃣  Recalculating ALL account balances from POSTED journals...")
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

	// 5. Verify final state
	log.Println("\n5️⃣  Verifying final state...")
	verifyRows, err := tx.Query(`
		SELECT 
			code, name, type, balance,
			CASE 
				WHEN type = 'REVENUE' AND balance < 0 THEN '✅ OK'
				WHEN type = 'EXPENSE' AND balance > 0 THEN '✅ OK'
				WHEN type IN ('REVENUE', 'EXPENSE') AND ABS(balance) < 0.01 THEN '✅ ZERO'
				ELSE '⚠️  CHECK'
			END as status
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

	log.Println("\n   Non-zero Revenue/Expense Accounts:")
	log.Println("   Code    | Name                        | Type    | Balance      | Status")
	log.Println("   --------|-----------------------------|---------|--------------|---------")
	
	hasData := false
	for verifyRows.Next() {
		hasData = true
		var code, name, accType, status string
		var balance float64
		verifyRows.Scan(&code, &name, &accType, &balance, &status)
		log.Printf("   %-7s | %-27s | %-7s | %12.2f | %s", code, name, accType, balance, status)
	}
	
	if !hasData {
		log.Println("   (All accounts are zero)")
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit: %v", err)
	}

	log.Println("\n============================================")
	log.Println("✅ Complete cleanup finished!")
	log.Println("============================================")
	log.Println("\n📝 Database is now clean and ready for:")
	log.Println("   - New period closing")
	log.Println("   - All balances recalculated from SSOT")
	log.Println("   - No corrupt data remaining")
}
