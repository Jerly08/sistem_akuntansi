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

	fmt.Println("=" + string(make([]byte, 70)))
	fmt.Println("FIX CLOSING BALANCES - MANUAL RECALCULATION")
	fmt.Println("=" + string(make([]byte, 70)))
	
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// 1. Get all REVENUE and EXPENSE accounts that should be zeroed
	fmt.Println("\n1. Checking accounts that need balance reset...")
	
	type Account struct {
		ID      int
		Code    string
		Name    string
		Type    string
		Balance float64
	}
	
	var accounts []Account
	
	rows, err := tx.Query(`
		SELECT id, code, name, type, balance
		FROM accounts
		WHERE type IN ('REVENUE', 'EXPENSE') AND ABS(balance) > 0.01
		ORDER BY type, code
	`)
	if err != nil {
		log.Fatalf("Failed to query accounts: %v", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var acc Account
		if err := rows.Scan(&acc.ID, &acc.Code, &acc.Name, &acc.Type, &acc.Balance); err != nil {
			log.Printf("Error scanning account: %v", err)
			continue
		}
		accounts = append(accounts, acc)
	}
	
	if len(accounts) == 0 {
		fmt.Println("✓ All revenue and expense accounts already have zero balance!")
		return
	}
	
	fmt.Printf("\nFound %d accounts with non-zero balance:\n", len(accounts))
	for _, acc := range accounts {
		fmt.Printf("  %s - %s (%s): Rp %.2f\n", acc.Code, acc.Name, acc.Type, acc.Balance)
	}
	
	// 2. Recalculate balances from unified_journal_lines
	fmt.Println("\n2. Recalculating balances from unified_journal_lines...")
	
	for _, acc := range accounts {
		// Calculate correct balance from ALL posted journal lines
		var totalDebit, totalCredit float64
		err := tx.QueryRow(`
			SELECT 
				COALESCE(SUM(ujl.debit_amount), 0) as total_debit,
				COALESCE(SUM(ujl.credit_amount), 0) as total_credit
			FROM unified_journal_lines ujl
			JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
			WHERE ujl.account_id = $1
			  AND uje.status = 'POSTED'
		`, acc.ID).Scan(&totalDebit, &totalCredit)
		
		if err != nil {
			log.Printf("Error calculating balance for account %s: %v", acc.Code, err)
			continue
		}
		
		// Calculate net balance based on account type
		var correctBalance float64
		if acc.Type == "REVENUE" || acc.Type == "EQUITY" || acc.Type == "LIABILITY" {
			// Credit normal accounts: credit increases, debit decreases
			correctBalance = totalCredit - totalDebit
		} else if acc.Type == "EXPENSE" || acc.Type == "ASSET" {
			// Debit normal accounts: debit increases, credit decreases
			correctBalance = totalDebit - totalCredit
		}
		
		fmt.Printf("\n  Account: %s - %s (%s)\n", acc.Code, acc.Name, acc.Type)
		fmt.Printf("    Current Balance: Rp %.2f\n", acc.Balance)
		fmt.Printf("    Total Debit: Rp %.2f | Total Credit: Rp %.2f\n", totalDebit, totalCredit)
		fmt.Printf("    Correct Balance: Rp %.2f\n", correctBalance)
		
		if acc.Type == "REVENUE" || acc.Type == "EXPENSE" {
			if correctBalance < 0.01 && correctBalance > -0.01 {
				fmt.Printf("    ✓ Should be ZERO after closing\n")
			} else {
				fmt.Printf("    ⚠️  WARNING: Balance should be zero after closing but calculated as %.2f\n", correctBalance)
			}
		}
		
		// Update the account balance
		result, err := tx.Exec(`
			UPDATE accounts
			SET balance = $1
			WHERE id = $2
		`, correctBalance, acc.ID)
		
		if err != nil {
			log.Printf("Error updating account %s: %v", acc.Code, err)
			continue
		}
		
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			fmt.Printf("    ✓ Balance updated: %.2f → %.2f\n", acc.Balance, correctBalance)
		}
	}
	
	// 3. Verify retained earnings balance
	fmt.Println("\n3. Verifying retained earnings balance...")
	
	var retainedEarningsID int
	var retainedEarningsBalance float64
	
	err = tx.QueryRow(`
		SELECT id, balance
		FROM accounts
		WHERE code = '3201' AND type = 'EQUITY'
	`).Scan(&retainedEarningsID, &retainedEarningsBalance)
	
	if err != nil {
		log.Printf("Warning: Could not find retained earnings account: %v", err)
	} else {
		// Recalculate retained earnings
		var totalDebit, totalCredit float64
		err := tx.QueryRow(`
			SELECT 
				COALESCE(SUM(ujl.debit_amount), 0) as total_debit,
				COALESCE(SUM(ujl.credit_amount), 0) as total_credit
			FROM unified_journal_lines ujl
			JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
			WHERE ujl.account_id = $1
			  AND uje.status = 'POSTED'
		`, retainedEarningsID).Scan(&totalDebit, &totalCredit)
		
		if err == nil {
			correctREBalance := totalCredit - totalDebit
			
			fmt.Printf("  Current Retained Earnings: Rp %.2f\n", retainedEarningsBalance)
			fmt.Printf("  Total Debit: Rp %.2f | Total Credit: Rp %.2f\n", totalDebit, totalCredit)
			fmt.Printf("  Correct Balance: Rp %.2f\n", correctREBalance)
			
			if retainedEarningsBalance != correctREBalance {
				_, err = tx.Exec(`
					UPDATE accounts
					SET balance = $1
					WHERE id = $2
				`, correctREBalance, retainedEarningsID)
				
				if err != nil {
					log.Printf("Error updating retained earnings: %v", err)
				} else {
					fmt.Printf("  ✓ Retained Earnings updated: %.2f → %.2f\n", retainedEarningsBalance, correctREBalance)
				}
			} else {
				fmt.Println("  ✓ Retained Earnings balance is correct")
			}
		}
	}
	
	// Ask for confirmation before committing (supports AUTO_COMMIT)
	fmt.Println("\n" + string(make([]byte, 70)))
	auto := os.Getenv("AUTO_COMMIT")
	if auto == "1" || auto == "true" || auto == "yes" || auto == "Y" || auto == "y" {
		if err := tx.Commit(); err != nil {
			log.Fatalf("Failed to commit transaction: %v", err)
		}
		fmt.Println("\n✓ Changes committed automatically (AUTO_COMMIT)")
	} else {
		fmt.Println("Do you want to commit these changes? (yes/no)")
		var response string
		fmt.Scanln(&response)
		if response == "yes" || response == "y" {
			if err := tx.Commit(); err != nil {
				log.Fatalf("Failed to commit transaction: %v", err)
			}
			fmt.Println("\n✓ Changes committed successfully!")
		} else {
			fmt.Println("\n⚠️  Changes rolled back (not committed)")
		}
	}
	
	fmt.Println(string(make([]byte, 70)))
	fmt.Println("Fix complete!")
}
