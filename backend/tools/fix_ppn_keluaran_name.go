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

	fmt.Println("=== FIXING PPN KELUARAN ACCOUNT NAME ===\n")

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

	// 1. Check current 2103 accounts
	fmt.Println("1. CHECKING CURRENT 2103 ACCOUNTS:")
	rows, err := tx.Query("SELECT id, code, name, type, balance FROM accounts WHERE code = '2103' ORDER BY id")
	if err != nil {
		fmt.Printf("Error checking accounts: %v\n", err)
		return
	}
	defer rows.Close()

	var accountsToUpdate []int
	for rows.Next() {
		var id int
		var code, name, accType string
		var balance float64
		err := rows.Scan(&id, &code, &name, &accType, &balance)
		if err != nil {
			continue
		}
		fmt.Printf("ID:%d - %s %s (%s): %.0f\n", id, code, name, accType, balance)
		
		if name != "PPN Keluaran" {
			accountsToUpdate = append(accountsToUpdate, id)
		}
	}

	// 2. Update account names to PPN Keluaran
	fmt.Println("\n2. UPDATING ACCOUNT NAMES:")
	for _, accountID := range accountsToUpdate {
		_, err = tx.Exec(`
			UPDATE accounts 
			SET name = 'PPN Keluaran',
				description = 'Output VAT - Sales Tax Payable',
				type = 'LIABILITY',
				category = 'CURRENT_LIABILITY',
				updated_at = $1
			WHERE id = $2
		`, time.Now(), accountID)
		if err != nil {
			fmt.Printf("Error updating account ID %d: %v\n", accountID, err)
			return
		}
		fmt.Printf("✅ Updated account ID:%d to 'PPN Keluaran'\n", accountID)
	}

	// 3. Remove duplicate 2103 accounts (keep only one)
	fmt.Println("\n3. REMOVING DUPLICATE ACCOUNTS:")
	
	// Get the account that has journal entries
	var activeAccountID int
	err = tx.QueryRow(`
		SELECT DISTINCT ujl.account_id
		FROM unified_journal_lines ujl
		JOIN accounts a ON ujl.account_id = a.id
		WHERE a.code = '2103'
		LIMIT 1
	`).Scan(&activeAccountID)
	if err != nil && err != sql.ErrNoRows {
		fmt.Printf("Error finding active account: %v\n", err)
		return
	}

	if activeAccountID > 0 {
		// Delete other 2103 accounts
		result, err := tx.Exec(`
			DELETE FROM accounts 
			WHERE code = '2103' AND id != $1
		`, activeAccountID)
		if err != nil {
			fmt.Printf("Error deleting duplicate accounts: %v\n", err)
			return
		}

		rowsDeleted, _ := result.RowsAffected()
		fmt.Printf("✅ Deleted %d duplicate 2103 accounts, kept active ID:%d\n", rowsDeleted, activeAccountID)
	} else {
		fmt.Println("ℹ️  No active account found, keeping first 2103 account")
		// Keep first account and delete others
		var firstAccountID int
		err = tx.QueryRow("SELECT id FROM accounts WHERE code = '2103' ORDER BY id LIMIT 1").Scan(&firstAccountID)
		if err != nil {
			fmt.Printf("Error getting first account: %v\n", err)
			return
		}

		result, err := tx.Exec("DELETE FROM accounts WHERE code = '2103' AND id != $1", firstAccountID)
		if err != nil {
			fmt.Printf("Error deleting accounts: %v\n", err)
			return
		}

		rowsDeleted, _ := result.RowsAffected()
		fmt.Printf("✅ Deleted %d accounts, kept first ID:%d\n", rowsDeleted, firstAccountID)
	}

	// 4. Update balance for the remaining account
	fmt.Println("\n4. UPDATING ACCOUNT BALANCE:")
	
	var ppnKeluaranBalance float64
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(ujl.credit_amount - ujl.debit_amount), 0)
		FROM unified_journal_lines ujl
		JOIN accounts a ON ujl.account_id = a.id
		WHERE a.code = '2103'
	`).Scan(&ppnKeluaranBalance)
	if err != nil {
		fmt.Printf("Error calculating balance: %v\n", err)
		return
	}

	_, err = tx.Exec("UPDATE accounts SET balance = $1 WHERE code = '2103'", ppnKeluaranBalance)
	if err != nil {
		fmt.Printf("Error updating balance: %v\n", err)
		return
	}
	fmt.Printf("✅ Updated PPN Keluaran balance: %.0f\n", ppnKeluaranBalance)

	// 5. Verification
	fmt.Println("\n5. VERIFICATION:")
	rows, err = tx.Query("SELECT code, name, type, balance FROM accounts WHERE code IN ('2102', '2103') ORDER BY code")
	if err != nil {
		fmt.Printf("Error verifying: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("Final PPN Accounts:")
	for rows.Next() {
		var code, name, accType string
		var balance float64
		err := rows.Scan(&code, &name, &accType, &balance)
		if err != nil {
			continue
		}
		fmt.Printf("  %s - %s (%s): %.0f\n", code, name, accType, balance)
	}

	fmt.Println("\n=== PPN KELUARAN NAME FIX COMPLETED ===")
}