package main

import (
	"fmt"
	"log"
	"time"

	"github.com/accounting_proj/database"
	"github.com/accounting_proj/models"
)

func main() {
	// Initialize database connection
	db, err := database.Connect()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	fmt.Println("=== Sales and Journal Entry Analysis ===")
	fmt.Println()

	// Check recent sales transactions
	fmt.Println("1. Recent Sales Transactions:")
	var sales []models.Sale
	err = db.Where("created_at > ?", time.Now().Add(-24*time.Hour)).Order("created_at desc").Limit(10).Find(&sales).Error
	if err != nil {
		log.Printf("Error fetching sales: %v", err)
	} else {
		for _, sale := range sales {
			fmt.Printf("  Sale ID: %d, Invoice: %s, Customer: %s, Total: %.2f, Status: %s, Date: %s\n", 
				sale.ID, sale.InvoiceNumber, sale.CustomerName, sale.Total, sale.Status, sale.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}
	fmt.Println()

	// Check recent journal headers
	fmt.Println("2. Recent Journal Headers:")
	var journals []models.Journal
	err = db.Where("created_at > ?", time.Now().Add(-24*time.Hour)).Order("created_at desc").Limit(10).Find(&journals).Error
	if err != nil {
		log.Printf("Error fetching journals: %v", err)
	} else {
		for _, journal := range journals {
			fmt.Printf("  Journal ID: %d, Source: %s, Ref: %s, Amount: %.2f, Date: %s\n", 
				journal.ID, journal.SourceType, journal.ReferenceID, journal.Amount, journal.CreatedAt.Format("2006-01-02 15:04:05"))
		}
	}
	fmt.Println()

	// Check recent journal lines
	fmt.Println("3. Recent Journal Lines:")
	var journalLines []models.JournalLine
	err = db.Joins("JOIN journals ON journals.id = journal_lines.journal_id").
		Where("journals.created_at > ?", time.Now().Add(-24*time.Hour)).
		Order("journal_lines.created_at desc").
		Limit(20).
		Find(&journalLines).Error
	if err != nil {
		log.Printf("Error fetching journal lines: %v", err)
	} else {
		for _, line := range journalLines {
			fmt.Printf("  Journal Line ID: %d, Journal ID: %d, Account: %s, Debit: %.2f, Credit: %.2f, Desc: %s\n", 
				line.ID, line.JournalID, line.AccountCode, line.Debit, line.Credit, line.Description)
		}
	}
	fmt.Println()

	// Check account balances for key accounts
	fmt.Println("4. Key Account Balances:")
	var accounts []models.Account
	keyCodes := []string{"1201", "4101", "2103", "2102"} // AR, Revenue, PPN Payable
	err = db.Where("code IN ?", keyCodes).Find(&accounts).Error
	if err != nil {
		log.Printf("Error fetching accounts: %v", err)
	} else {
		for _, account := range accounts {
			fmt.Printf("  Account %s (%s): Balance = %.2f\n", account.Code, account.Name, account.Balance)
		}
	}
	fmt.Println()

	// Check for sales invoice INV-2025-1724
	fmt.Println("5. Checking specific invoice INV-2025-1724:")
	var specificSale models.Sale
	err = db.Where("invoice_number = ?", "INV/2025/09/0002").First(&specificSale).Error
	if err != nil {
		log.Printf("Error finding specific sale: %v", err)
	} else {
		fmt.Printf("  Found sale: ID=%d, Total=%.2f, Status=%s\n", specificSale.ID, specificSale.Total, specificSale.Status)
		
		// Check journals for this sale
		var relatedJournals []models.Journal
		err = db.Where("reference_id = ? AND source_type = ?", fmt.Sprintf("%d", specificSale.ID), "sale").Find(&relatedJournals).Error
		if err != nil {
			log.Printf("  Error finding related journals: %v", err)
		} else {
			fmt.Printf("  Found %d related journals:\n", len(relatedJournals))
			for _, journal := range relatedJournals {
				fmt.Printf("    Journal ID: %d, Amount: %.2f, Auto-Posted: %t\n", journal.ID, journal.Amount, journal.AutoPost)
				
				// Get journal lines for this journal
				var lines []models.JournalLine
				err = db.Where("journal_id = ?", journal.ID).Find(&lines).Error
				if err != nil {
					log.Printf("    Error getting journal lines: %v", err)
				} else {
					for _, line := range lines {
						fmt.Printf("      Line: Account %s, Debit: %.2f, Credit: %.2f, Desc: %s\n", 
							line.AccountCode, line.Debit, line.Credit, line.Description)
					}
				}
			}
		}
	}

	fmt.Println("\n=== Analysis Complete ===")
}