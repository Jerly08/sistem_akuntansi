package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	dsn := "postgres://postgres:postgres@localhost/sistem_akuntansi?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("=== IMPLEMENTING PPN ACCOUNTS SEPARATION FIX V2 ===\n")

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatal("Failed to begin transaction:", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			fmt.Printf("Transaction rolled back due to error: %v\n", err)
		} else {
			err = tx.Commit()
			if err != nil {
				fmt.Printf("Failed to commit transaction: %v\n", err)
			} else {
				fmt.Println("✅ Transaction committed successfully!")
			}
		}
	}()

	// 1. Clean up duplicate 2102 accounts
	fmt.Println("1. CLEANING UP DUPLICATE 2102 ACCOUNTS:")
	
	// Find the active PPN Masukan account (the one with balance)
	var activePPNMasukanID int
	err = tx.QueryRow(`
		SELECT id FROM accounts 
		WHERE code = '2102' AND name = 'PPN Masukan' AND balance > 0
		LIMIT 1
	`).Scan(&activePPNMasukanID)
	if err != nil {
		fmt.Printf("Error finding active PPN Masukan: %v\n", err)
		return
	}
	fmt.Printf("✅ Found active PPN Masukan account ID: %d\n", activePPNMasukanID)

	// Delete unused 2102 accounts
	result, err := tx.Exec(`
		DELETE FROM accounts 
		WHERE code = '2102' AND id != $1 AND balance = 0
	`, activePPNMasukanID)
	if err != nil {
		fmt.Printf("Error deleting duplicate accounts: %v\n", err)
		return
	}

	rowsDeleted, _ := result.RowsAffected()
	fmt.Printf("✅ Deleted %d duplicate 2102 accounts\n", rowsDeleted)

	// 2. Create PPN Keluaran account (Output VAT - Liability)
	fmt.Println("\n2. CREATING PPN KELUARAN ACCOUNT:")
	
	var existingCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM accounts WHERE code = '2103'").Scan(&existingCount)
	if err != nil {
		return
	}

	if existingCount == 0 {
		_, err = tx.Exec(`
			INSERT INTO accounts (code, name, description, type, category, parent_id, level, is_header, is_active, balance, created_at, updated_at)
			VALUES ('2103', 'PPN Keluaran', 'Output VAT - Sales Tax Payable', 'LIABILITY', 'CURRENT_LIABILITY', 
					(SELECT id FROM accounts WHERE code = '2100' LIMIT 1), 3, false, true, 0, $1, $1)
		`, time.Now())
		if err != nil {
			fmt.Printf("Error creating PPN Keluaran account: %v\n", err)
			return
		}
		fmt.Println("✅ Created account 2103 - PPN Keluaran (LIABILITY)")
	} else {
		fmt.Println("ℹ️  Account 2103 already exists, skipping creation")
	}

	// 3. Ensure PPN Masukan is properly set up as Asset
	fmt.Println("\n3. UPDATING PPN MASUKAN ACCOUNT:")
	
	_, err = tx.Exec(`
		UPDATE accounts 
		SET type = 'ASSET', 
			category = 'CURRENT_ASSET',
			name = 'PPN Masukan',
			description = 'Input VAT - Recoverable Tax Asset',
			updated_at = $1
		WHERE id = $2
	`, time.Now(), activePPNMasukanID)
	if err != nil {
		fmt.Printf("Error updating PPN Masukan: %v\n", err)
		return
	}
	fmt.Println("✅ Updated PPN Masukan account to ASSET type")

	// 4. Check current journal entries using wrong PPN account for sales
	fmt.Println("\n4. ANALYZING CURRENT JOURNAL ENTRIES:")
	
	rows, err := tx.Query(`
		SELECT l.id, l.source_code, l.source_type, ujl.credit_amount
		FROM unified_journal_ledger l
		JOIN unified_journal_lines ujl ON ujl.journal_id = l.id
		WHERE l.source_type = 'SALE' 
		  AND ujl.account_id = $1
		  AND ujl.credit_amount > 0
	`, activePPNMasukanID)
	if err != nil {
		fmt.Printf("Error analyzing journal entries: %v\n", err)
		return
	}
	defer rows.Close()

	var salesToFix []struct {
		JournalID   int
		SourceCode  string
		SourceType  string
		CreditAmount float64
	}

	for rows.Next() {
		var sale struct {
			JournalID   int
			SourceCode  string
			SourceType  string
			CreditAmount float64
		}
		err := rows.Scan(&sale.JournalID, &sale.SourceCode, &sale.SourceType, &sale.CreditAmount)
		if err != nil {
			continue
		}
		salesToFix = append(salesToFix, sale)
		fmt.Printf("Found sales journal ID:%d (%s) with PPN Keluaran %.0f using wrong account\n", 
			sale.JournalID, sale.SourceCode, sale.CreditAmount)
	}

	// 5. Fix journal entries to use correct PPN Keluaran account
	fmt.Println("\n5. FIXING JOURNAL ENTRIES:")
	
	if len(salesToFix) > 0 {
		// Get PPN Keluaran account ID
		var ppnKeluaranID int
		err = tx.QueryRow("SELECT id FROM accounts WHERE code = '2103'").Scan(&ppnKeluaranID)
		if err != nil {
			fmt.Printf("Error getting PPN Keluaran account ID: %v\n", err)
			return
		}

		for _, sale := range salesToFix {
			// Update journal lines for sales to use PPN Keluaran instead of PPN Masukan
			_, err = tx.Exec(`
				UPDATE unified_journal_lines 
				SET account_id = $1
				WHERE journal_id = $2 
				  AND account_id = $3
				  AND credit_amount > 0
			`, ppnKeluaranID, sale.JournalID, activePPNMasukanID)
			if err != nil {
				fmt.Printf("Error fixing journal entry %d: %v\n", sale.JournalID, err)
				return
			}
			fmt.Printf("✅ Fixed journal ID:%d (%s) to use PPN Keluaran account\n", 
				sale.JournalID, sale.SourceCode)
		}
	} else {
		fmt.Println("ℹ️  No sales journal entries need fixing")
	}

	// 6. Update account balances
	fmt.Println("\n6. UPDATING ACCOUNT BALANCES:")
	
	// Calculate PPN Masukan balance (should be debit/positive for asset)
	var ppnMasukanBalance float64
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(ujl.debit_amount - ujl.credit_amount), 0)
		FROM unified_journal_lines ujl
		WHERE ujl.account_id = $1
	`, activePPNMasukanID).Scan(&ppnMasukanBalance)
	if err != nil {
		fmt.Printf("Error calculating PPN Masukan balance: %v\n", err)
		return
	}

	// Calculate PPN Keluaran balance (should be credit/positive for liability)
	var ppnKeluaranBalance float64
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(ujl.credit_amount - ujl.debit_amount), 0)
		FROM unified_journal_lines ujl
		JOIN accounts a ON ujl.account_id = a.id
		WHERE a.code = '2103'
	`).Scan(&ppnKeluaranBalance)
	if err != nil {
		fmt.Printf("Error calculating PPN Keluaran balance: %v\n", err)
		return
	}

	// Update balances
	_, err = tx.Exec("UPDATE accounts SET balance = $1 WHERE id = $2", ppnMasukanBalance, activePPNMasukanID)
	if err != nil {
		fmt.Printf("Error updating PPN Masukan balance: %v\n", err)
		return
	}

	_, err = tx.Exec("UPDATE accounts SET balance = $1 WHERE code = '2103'", ppnKeluaranBalance)
	if err != nil {
		fmt.Printf("Error updating PPN Keluaran balance: %v\n", err)
		return
	}

	fmt.Printf("✅ Updated PPN Masukan balance: %.0f (Asset - Debit Balance)\n", ppnMasukanBalance)
	fmt.Printf("✅ Updated PPN Keluaran balance: %.0f (Liability - Credit Balance)\n", ppnKeluaranBalance)

	// 7. Verify the fix
	fmt.Println("\n7. VERIFICATION:")
	
	// Check PPN accounts
	rows, err = tx.Query(`
		SELECT code, name, type, balance
		FROM accounts 
		WHERE code IN ('2102', '2103')
		ORDER BY code
	`)
	if err != nil {
		fmt.Printf("Error verifying accounts: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("PPN Accounts after fix:")
	for rows.Next() {
		var code, name, accType string
		var balance float64
		err := rows.Scan(&code, &name, &accType, &balance)
		if err != nil {
			continue
		}
		fmt.Printf("  %s - %s (%s): %.0f\n", code, name, accType, balance)
	}

	// Check journal entries are now using correct accounts
	fmt.Println("\nJournal entries verification:")
	rows, err = tx.Query(`
		SELECT l.source_type, l.source_code, a.code, a.name, 
			   ujl.debit_amount, ujl.credit_amount
		FROM unified_journal_ledger l
		JOIN unified_journal_lines ujl ON ujl.journal_id = l.id
		JOIN accounts a ON ujl.account_id = a.id
		WHERE a.code IN ('2102', '2103')
		ORDER BY l.source_type, l.source_code, a.code
	`)
	if err != nil {
		fmt.Printf("Error verifying journal entries: %v\n", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var sourceType, sourceCode, accountCode, accountName string
		var debitAmount, creditAmount float64
		err := rows.Scan(&sourceType, &sourceCode, &accountCode, &accountName,
			&debitAmount, &creditAmount)
		if err != nil {
			continue
		}

		if debitAmount > 0 {
			fmt.Printf("  %s %s: Dr. %s (%s) %.0f\n", 
				sourceType, sourceCode, accountCode, accountName, debitAmount)
		}
		if creditAmount > 0 {
			fmt.Printf("  %s %s: Cr. %s (%s) %.0f\n", 
				sourceType, sourceCode, accountCode, accountName, creditAmount)
		}
	}

	fmt.Println("\n8. SUMMARY OF CHANGES:")
	fmt.Println("✅ Cleaned up duplicate 2102 accounts")
	fmt.Println("✅ Created account 2103 - PPN Keluaran (LIABILITY) for sales tax")
	fmt.Println("✅ Ensured account 2102 - PPN Masukan (ASSET) for purchase tax")
	fmt.Println("✅ Fixed existing journal entries to use correct PPN accounts")
	fmt.Println("✅ Updated account balances based on journal transactions")
	fmt.Println("✅ Purchase journals use PPN Masukan (2102) - ASSET")
	fmt.Println("✅ Sales journals use PPN Keluaran (2103) - LIABILITY")

	fmt.Println("\n=== PPN ACCOUNTS SEPARATION FIX V2 COMPLETED ===")
}