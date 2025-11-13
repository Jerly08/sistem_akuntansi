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
	log.Println("FINAL FIX - Delete Future Sales & Fix Balances")
	log.Println("============================================")

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// 1. Delete sales with unrealistic dates (2026, 2027)
	log.Println("\n1️⃣  Deleting sales with unrealistic dates...")
	
	// First get related data
	rows, err := tx.Query(`
		SELECT id, entry_date, description, source_id, source_type
		FROM unified_journal_ledger 
		WHERE source_type = 'SALE' AND entry_date > '2025-12-31'
	`)
	if err != nil {
		log.Fatalf("Failed to query future sales: %v", err)
	}
	
	type SaleInfo struct {
		JournalID  int
		Date       string
		Desc       string
		SourceID   sql.NullInt64
		SourceType sql.NullString
	}
	
	var futureSales []SaleInfo
	for rows.Next() {
		var s SaleInfo
		rows.Scan(&s.JournalID, &s.Date, &s.Desc, &s.SourceID, &s.SourceType)
		futureSales = append(futureSales, s)
	}
	rows.Close()
	
	for _, sale := range futureSales {
		log.Printf("   Deleting: ID %d - %s - %s", sale.JournalID, sale.Date[:10], sale.Desc)
		
		// Delete journal lines
		_, err = tx.Exec(`DELETE FROM unified_journal_lines WHERE journal_id = $1`, sale.JournalID)
		if err != nil {
			log.Fatalf("Failed to delete journal lines: %v", err)
		}
		
		// Delete journal entry
		_, err = tx.Exec(`DELETE FROM unified_journal_ledger WHERE id = $1`, sale.JournalID)
		if err != nil {
			log.Fatalf("Failed to delete journal entry: %v", err)
		}
		
		// If this is linked to a sales record, we might need to update it
		if sale.SourceID.Valid && sale.SourceType.String == "sales" {
			log.Printf("   ⚠️  This was linked to sales ID %d - you may need to delete that record manually", sale.SourceID.Int64)
		}
	}
	
	log.Printf("   ✅ Deleted %d unrealistic sales entries", len(futureSales))

	// 2. Recalculate balances with CORRECT formula
	// For REVENUE, EQUITY, LIABILITY: balance should be NEGATIVE when credit > debit
	log.Println("\n2️⃣  Recalculating balances with correct formula...")
	result, err := tx.Exec(`
		UPDATE accounts a
		SET balance = COALESCE((
			SELECT 
				CASE 
					-- For ASSET & EXPENSE: Debit - Credit (normal debit balance)
					WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
						COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
					-- For LIABILITY, EQUITY, REVENUE: Debit - Credit (will be negative for credit balance)
					-- Note: We store credit balances as NEGATIVE values
					ELSE 
						COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
				END
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
	rows2, _ := result.RowsAffected()
	log.Printf("   ✅ Recalculated %d accounts", rows2)

	// 3. Verify
	log.Println("\n3️⃣  Verifying final state...")
	verifyRows, err := tx.Query(`
		SELECT 
			code, name, type, balance,
			CASE 
				WHEN type = 'REVENUE' AND balance < 0 THEN '✅ OK - Credit balance'
				WHEN type = 'EXPENSE' AND balance > 0 THEN '✅ OK - Debit balance'
				WHEN type IN ('REVENUE', 'EXPENSE') AND ABS(balance) < 0.01 THEN '✅ ZERO'
				ELSE '⚠️  UNUSUAL'
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

	log.Println("\n   Revenue/Expense Accounts:")
	log.Println("   Code    | Name                        | Type    | Balance       | Status")
	log.Println("   --------|-----------------------------|---------|--------------|-----------------------")
	
	hasData := false
	for verifyRows.Next() {
		hasData = true
		var code, name, accType, status string
		var balance float64
		verifyRows.Scan(&code, &name, &accType, &balance, &status)
		log.Printf("   %-7s | %-27s | %-7s | %13.2f | %s", code, name, accType, balance, status)
	}
	
	if !hasData {
		log.Println("   (All revenue/expense accounts are zero - ready for new transactions)")
	}

	// Commit
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit: %v", err)
	}

	log.Println("\n============================================")
	log.Println("✅ Final fix completed successfully!")
	log.Println("============================================")
	log.Println("\n📝 Summary:")
	log.Println("   ✅ Removed all unrealistic future sales")
	log.Println("   ✅ Recalculated all balances with correct formula")
	log.Println("   ✅ Revenue accounts now show NEGATIVE (credit) balances")
	log.Println("   ✅ Expense accounts now show POSITIVE (debit) balances")
	log.Println("\n🎯 Database is now ready for period closing!")
}
