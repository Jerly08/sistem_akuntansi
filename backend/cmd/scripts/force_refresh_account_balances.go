package main

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("FORCE REFRESH ACCOUNT BALANCES FROM SSOT JOURNAL")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()

	// Step 1: Current state
	fmt.Println("📊 STEP 1: Current Account Balances (Key Accounts)")
	fmt.Println(string(make([]byte, 80)))

	type AccountInfo struct {
		Code    string
		Name    string
		Balance float64
	}

	var before []AccountInfo
	db.Raw(`
		SELECT code, name, balance 
		FROM accounts 
		WHERE code IN ('5101', '1301', '4101')
		ORDER BY code
	`).Scan(&before)

	for _, acc := range before {
		fmt.Printf("%-8s %-40s %18.2f\n", acc.Code, acc.Name, acc.Balance)
	}
	fmt.Println()

	// Step 2: Calculate from ssot_journal_lines
	fmt.Println("📊 STEP 2: Calculate Balances from SSOT Journal Lines")
	fmt.Println(string(make([]byte, 80)))

	type CalculatedBalance struct {
		AccountID    uint
		Code         string
		Name         string
		TotalDebit   float64
		TotalCredit  float64
		NetBalance   float64
	}

	var calculated []CalculatedBalance
	db.Raw(`
		SELECT 
			a.id as account_id,
			a.code,
			a.name,
			COALESCE(SUM(jl.debit_amount), 0) as total_debit,
			COALESCE(SUM(jl.credit_amount), 0) as total_credit,
			COALESCE(SUM(jl.debit_amount), 0) - COALESCE(SUM(jl.credit_amount), 0) as net_balance
		FROM accounts a
		LEFT JOIN unified_journal_lines jl ON jl.account_id = a.id
		LEFT JOIN unified_journal_ledger ujl ON ujl.id = jl.journal_id
		WHERE a.code IN ('5101', '1301', '4101')
		  AND (ujl.status = 'POSTED' OR ujl.status IS NULL)
		  AND ujl.deleted_at IS NULL
		GROUP BY a.id, a.code, a.name
		ORDER BY a.code
	`).Scan(&calculated)

	for _, acc := range calculated {
		fmt.Printf("%-8s %-40s Dr: %12.2f | Cr: %12.2f | Net: %12.2f\n",
			acc.Code, acc.Name, acc.TotalDebit, acc.TotalCredit, acc.NetBalance)
	}
	fmt.Println()

	// Step 3: Update balances
	fmt.Println("🔄 STEP 3: Updating Account Balances")
	fmt.Println(string(make([]byte, 80)))

	updatedCount := 0
	for _, acc := range calculated {
		result := db.Exec(`
			UPDATE accounts 
			SET balance = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, acc.NetBalance, acc.AccountID)

		if result.Error != nil {
			fmt.Printf("❌ Failed to update %s: %v\n", acc.Code, result.Error)
		} else {
			fmt.Printf("✅ Updated %-8s %-40s New Balance: %12.2f\n",
				acc.Code, acc.Name, acc.NetBalance)
			updatedCount++
		}
	}
	fmt.Println()

	// Step 4: Update ALL accounts (comprehensive refresh)
	fmt.Println("🔄 STEP 4: Comprehensive Update for ALL Accounts")
	fmt.Println(string(make([]byte, 80)))

	// Update all non-header accounts
	result := db.Exec(`
		UPDATE accounts a
		SET balance = COALESCE((
			SELECT SUM(jl.debit_amount) - SUM(jl.credit_amount)
			FROM unified_journal_lines jl
			JOIN unified_journal_ledger ujl ON ujl.id = jl.journal_id
			WHERE jl.account_id = a.id
			  AND ujl.status = 'POSTED'
			  AND ujl.deleted_at IS NULL
		), 0),
		updated_at = CURRENT_TIMESTAMP
		WHERE a.is_header = false
	`)

	if result.Error != nil {
		fmt.Printf("❌ Failed comprehensive update: %v\n", result.Error)
	} else {
		fmt.Printf("✅ Updated %d accounts comprehensively\n", result.RowsAffected)
	}
	fmt.Println()

	// Step 5: Update parent balances
	fmt.Println("🔄 STEP 5: Updating Parent (Header) Account Balances")
	fmt.Println(string(make([]byte, 80)))

	// Recursive update for parent accounts (up to 5 levels)
	for level := 5; level >= 1; level-- {
		result := db.Exec(`
			UPDATE accounts parent
			SET balance = COALESCE((
				SELECT SUM(child.balance)
				FROM accounts child
				WHERE child.parent_id = parent.id
				  AND child.deleted_at IS NULL
			), 0),
			updated_at = CURRENT_TIMESTAMP
			WHERE parent.is_header = true
			  AND parent.level = ?
		`, level)

		if result.Error == nil && result.RowsAffected > 0 {
			fmt.Printf("✅ Updated level %d parent accounts (%d rows)\n", level, result.RowsAffected)
		}
	}
	fmt.Println()

	// Step 6: Verification
	fmt.Println("📊 STEP 6: Verification - Updated Balances")
	fmt.Println(string(make([]byte, 80)))

	var after []AccountInfo
	db.Raw(`
		SELECT code, name, balance 
		FROM accounts 
		WHERE code IN ('5101', '1301', '4101', '5000', '1000', '4000')
		ORDER BY code
	`).Scan(&after)

	for _, acc := range after {
		fmt.Printf("%-8s %-40s %18.2f\n", acc.Code, acc.Name, acc.Balance)
	}
	fmt.Println()

	// Step 7: Detailed COGS verification
	fmt.Println("📊 STEP 7: Detailed COGS Verification")
	fmt.Println(string(make([]byte, 80)))

	type COGSDetail struct {
		JournalID   uint
		EntryNumber string
		EntryDate   string
		Reference   string
		Description string
		Amount      float64
	}

	var cogsDetails []COGSDetail
	db.Raw(`
		SELECT 
			ujl.id as journal_id,
			ujl.entry_number,
			ujl.entry_date::text,
			ujl.reference,
			ujl.description,
			jl.debit_amount as amount
		FROM unified_journal_ledger ujl
		JOIN unified_journal_lines jl ON jl.journal_id = ujl.id
		JOIN accounts a ON a.id = jl.account_id
		WHERE a.code = '5101'
		  AND ujl.status = 'POSTED'
		  AND ujl.deleted_at IS NULL
		  AND jl.debit_amount > 0
		ORDER BY ujl.created_at
	`).Scan(&cogsDetails)

	fmt.Printf("Found %d COGS entries:\n\n", len(cogsDetails))

	totalCOGS := 0.0
	for _, cogs := range cogsDetails {
		fmt.Printf("Journal #%d | %s | %s\n", cogs.JournalID, cogs.EntryNumber, cogs.EntryDate)
		fmt.Printf("  Ref: %s\n", cogs.Reference)
		fmt.Printf("  Desc: %s\n", cogs.Description)
		fmt.Printf("  Amount: Rp %.2f\n\n", cogs.Amount)
		totalCOGS += cogs.Amount
	}

	fmt.Println(string(make([]byte, 80)))
	fmt.Printf("Total COGS (should match account balance): Rp %.2f\n", totalCOGS)
	fmt.Println()

	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println("✅ REFRESH COMPLETE")
	fmt.Println("=" + string(make([]byte, 80)))
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Hard refresh Chart of Accounts page (Ctrl+Shift+R or Ctrl+F5)")
	fmt.Println("  2. Verify all account balances are correct")
	fmt.Println("  3. Generate P&L Report again to confirm COGS appears correctly")
	fmt.Println()
}

