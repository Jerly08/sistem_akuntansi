package main

import (
	"database/sql"
	"fmt"
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
	log.Println("VERIFY & FIX - Period Closing Issue")
	log.Println("============================================")

	// Step 1: Check current account balances
	log.Println("\n1️⃣  Checking current account balances...")
	rows, err := db.Query(`
		SELECT code, name, type, balance
		FROM accounts
		WHERE type IN ('REVENUE', 'EXPENSE')
		  AND is_header = false
		  AND ABS(balance) > 0.01
		  AND deleted_at IS NULL
		ORDER BY type, code
	`)
	if err != nil {
		log.Fatalf("Failed to query accounts: %v", err)
	}

	log.Println("\n   Code    | Name                        | Type    | Balance      | Issue?")
	log.Println("   --------|-----------------------------|---------|--------------|--------------")
	
	var issueCount int
	for rows.Next() {
		var code, name, accType string
		var balance float64
		rows.Scan(&code, &name, &accType, &balance)
		
		issue := ""
		if accType == "REVENUE" && balance > 0 {
			issue = "⚠️  WRONG SIGN!"
			issueCount++
		} else if accType == "EXPENSE" && balance < 0 {
			issue = "⚠️  WRONG SIGN!"
			issueCount++
		}
		
		log.Printf("   %-7s | %-27s | %-7s | %12.2f | %s", code, name, accType, balance, issue)
	}
	rows.Close()

	if issueCount > 0 {
		log.Printf("\n   ❌ Found %d accounts with WRONG SIGN", issueCount)
		log.Println("   This means balance calculation is using OLD FORMULA")
		log.Println("\n   🔧 Applying fix...")
		
		// Step 2: Apply fix
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("Failed to begin transaction: %v", err)
		}
		defer tx.Rollback()

		// Delete all closing entries first
		log.Println("\n2️⃣  Removing corrupt closing entries...")
		result, err := tx.Exec(`
			DELETE FROM unified_journal_lines 
			WHERE journal_id IN (
				SELECT id FROM unified_journal_ledger WHERE source_type = 'CLOSING'
			)
		`)
		if err != nil {
			log.Fatalf("Failed to delete journal lines: %v", err)
		}
		rows1, _ := result.RowsAffected()
		
		result, err = tx.Exec(`DELETE FROM unified_journal_ledger WHERE source_type = 'CLOSING'`)
		if err != nil {
			log.Fatalf("Failed to delete journals: %v", err)
		}
		rows2, _ := result.RowsAffected()
		log.Printf("   ✅ Deleted %d lines and %d journals", rows1, rows2)

		result, err = tx.Exec(`DELETE FROM accounting_periods`)
		if err != nil {
			log.Fatalf("Failed to delete periods: %v", err)
		}
		rows3, _ := result.RowsAffected()
		log.Printf("   ✅ Deleted %d accounting periods", rows3)

		// Recalculate with CORRECT formula
		log.Println("\n3️⃣  Recalculating balances with CORRECT FORMULA...")
		result, err = tx.Exec(`
			UPDATE accounts a
			SET balance = COALESCE((
				SELECT 
					-- For ALL types: Debit - Credit
					-- ASSET/EXPENSE will be positive (debit balance)
					-- LIABILITY/EQUITY/REVENUE will be negative (credit balance)
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
		log.Printf("   ✅ Recalculated %d accounts", rowsAffected)

		// Commit
		if err := tx.Commit(); err != nil {
			log.Fatalf("Failed to commit: %v", err)
		}

		log.Println("\n✅ FIX APPLIED!")
	} else {
		log.Println("\n   ✅ All balances have CORRECT SIGN")
	}

	// Step 3: Verify final state
	log.Println("\n4️⃣  Verifying final state...")
	rows, err = db.Query(`
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

	log.Println("\n   Code    | Name                        | Type    | Balance      | Status")
	log.Println("   --------|-----------------------------|---------|--------------|---------")
	
	hasData := false
	for rows.Next() {
		hasData = true
		var code, name, accType string
		var balance float64
		rows.Scan(&code, &name, &accType, &balance)
		
		status := "✅ OK"
		if (accType == "REVENUE" && balance > 0) || (accType == "EXPENSE" && balance < 0) {
			status = "❌ STILL WRONG"
		}
		
		log.Printf("   %-7s | %-27s | %-7s | %12.2f | %s", code, name, accType, balance, status)
	}
	rows.Close()
	
	if !hasData {
		log.Println("   (All revenue/expense accounts are zero)")
	}

	log.Println("\n============================================")
	log.Println("✅ Verification and fix completed!")
	log.Println("============================================")
	log.Println("\n📝 Instructions for the other PC:")
	log.Println("   1. Pull latest code from GitHub:")
	log.Println("      git pull origin main")
	log.Println("   2. Run this script:")
	log.Println("      go run cmd/verify_and_fix_pc.go")
	log.Println("   3. Restart backend server")
	log.Println("   4. Try period closing again")
	fmt.Println("\n⚠️  IMPORTANT: Make sure to pull latest code FIRST!")
}
