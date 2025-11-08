package main

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Account struct {
	ID      uint    `gorm:"primaryKey"`
	Code    string  `gorm:"size:50;not null;unique"`
	Name    string  `gorm:"size:255;not null"`
	Type    string  `gorm:"size:50;not null"`
	Balance float64 `gorm:"type:decimal(15,2);default:0"`
}

func (Account) TableName() string {
	return "accounts"
}

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=sistem_akuntansi port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("========================================")
	fmt.Println("FIXING CLOSING DATA")
	fmt.Println("========================================\n")

	err = db.Transaction(func(tx *gorm.DB) error {
		// 1. Delete the incorrect closing period record
		fmt.Println("Step 1: Delete incorrect closing period record...")
		result := tx.Exec("DELETE FROM accounting_periods WHERE is_closed = true")
		if result.Error != nil {
			return fmt.Errorf("failed to delete closing period: %v", result.Error)
		}
		fmt.Printf("✓ Deleted %d closing period record(s)\n\n", result.RowsAffected)

		// 2. Reset Revenue account balance to 0 (should be 0 before proper closing)
		// Actually, we should keep the balance so it can be closed properly
		// So we DON'T reset it here

		// 3. Set Expense account balance correctly
		// Fix ALL expense accounts with negative balances
		fmt.Println("Step 2: Fix expense account balance signs...")
		
		var expenseAccounts []Account
		err := tx.Where("type = ? AND is_active = true", "EXPENSE").Find(&expenseAccounts).Error
		if err != nil {
			return fmt.Errorf("failed to find expense accounts: %v", err)
		}
		
		fixedCount := 0
		for _, acc := range expenseAccounts {
			if acc.Balance < 0 {
				fmt.Printf("Found expense account [%s] %s with NEGATIVE balance: %.2f\n", 
					acc.Code, acc.Name, acc.Balance)
				
				newBalance := -acc.Balance // Make it positive
				err = tx.Model(&Account{}).Where("id = ?", acc.ID).
					Update("balance", newBalance).Error
				if err != nil {
					return fmt.Errorf("failed to update expense balance: %v", err)
				}
				fmt.Printf("✓ Fixed to positive balance: %.2f\n", newBalance)
				fixedCount++
			}
		}
		
		if fixedCount == 0 {
			fmt.Println("✓ No negative expense balances found\n")
		} else {
			fmt.Printf("\n✓ Fixed %d expense account(s)\n\n", fixedCount)
		}

		// 4. Reset Retained Earnings to 0 (will be populated after proper closing)
		fmt.Println("Step 3: Reset Retained Earnings to 0...")
		err = tx.Model(&Account{}).Where("code = ?", "3201").
			Update("balance", 0).Error
		if err != nil {
			return fmt.Errorf("failed to reset retained earnings: %v", err)
		}
		fmt.Println("✓ Retained Earnings reset to 0\n")

		// 5. Verify balance sheet equation
		fmt.Println("Step 4: Verify balance sheet after fixes...")
		
		type BalanceResult struct {
			Type    string
			Balance float64
		}
		
		var results []BalanceResult
		tx.Raw(`
			SELECT type, SUM(balance) as balance
			FROM accounts
			WHERE is_active = true AND COALESCE(is_header, false) = false
			GROUP BY type
			ORDER BY type
		`).Scan(&results)
		
		var assets, liabilities, equity, revenue, expense float64
		for _, r := range results {
			fmt.Printf("%-15s: Rp %15.2f\n", r.Type, r.Balance)
			switch r.Type {
			case "ASSET":
				assets = r.Balance
			case "LIABILITY":
				liabilities = r.Balance
			case "EQUITY":
				equity = r.Balance
			case "REVENUE":
				revenue = r.Balance
			case "EXPENSE":
				expense = r.Balance
			}
		}
		
		// Accounting equation BEFORE closing should be:
		// Assets = Liabilities + Equity + (Revenue - Expense)
		// Or: Assets + Expense = Liabilities + Equity + Revenue
		
		leftSide := assets + expense
		rightSide := liabilities + equity + revenue
		diff := leftSide - rightSide
		
		fmt.Println("----------------------------------------")
		fmt.Printf("Left (A+E)     : Rp %15.2f\n", leftSide)
		fmt.Printf("Right (L+Eq+R) : Rp %15.2f\n", rightSide)
		fmt.Printf("DIFFERENCE     : Rp %15.2f ", diff)
		
		if diff > 0.01 || diff < -0.01 {
			fmt.Println("❌ NOT BALANCED")
			return fmt.Errorf("balance sheet still not balanced after fixes")
		} else {
			fmt.Println("✓ BALANCED")
		}
		
		fmt.Println("\n✅ All fixes applied successfully!")
		fmt.Println("Now you can run period closing again through the UI.")
		
		return nil
	})

	if err != nil {
		log.Fatal("Transaction failed:", err)
	}

	fmt.Println("\n========================================")
	fmt.Println("COMPLETED")
	fmt.Println("========================================")
}
