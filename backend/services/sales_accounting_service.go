package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"app-sistem-akuntansi/models"
)

// Enhanced Accounting Methods for Sales Module

// validateCustomerCreditLimit checks if customer has sufficient credit limit
func (s *SalesService) validateCustomerCreditLimit(customer *models.Contact, request models.SaleCreateRequest) error {
	// Calculate total amount from items
	totalAmount := 0.0
	for _, item := range request.Items {
		totalAmount += float64(item.Quantity) * item.UnitPrice
	}
	
	// Apply discounts
	if request.DiscountPercent > 0 {
		totalAmount = totalAmount - (totalAmount * request.DiscountPercent / 100)
	}
	
	// Add shipping cost
	totalAmount += request.ShippingCost
	
	// Check current outstanding amount
	outstandingAmount, err := s.salesRepo.GetCustomerOutstandingAmount(customer.ID)
	if err != nil {
		return err
	}
	
	// Check credit limit
	if customer.CreditLimit > 0 && (outstandingAmount + totalAmount) > customer.CreditLimit {
		return fmt.Errorf("credit limit exceeded. Available credit: %.2f", customer.CreditLimit - outstandingAmount)
	}
	
	return nil
}

// calculateSaleItemsWithCOGS calculates sale items with Cost of Goods Sold
func (s *SalesService) calculateSaleItemsWithCOGS(sale *models.Sale, items []models.SaleItemRequest) error {
	subtotal := 0.0
	totalCOGS := 0.0
	totalTax := 0.0
	
	sale.SaleItems = []models.SaleItem{}
	
	for _, itemReq := range items {
		// Get product details
		product, err := s.productRepo.FindByID(itemReq.ProductID)
		if err != nil {
			return fmt.Errorf("product %d not found", itemReq.ProductID)
		}
		
		// Check stock availability
		if product.Stock < itemReq.Quantity {
			return fmt.Errorf("insufficient stock for product %s. Available: %d", product.Name, product.Stock)
		}
		
		// Create sale item
		item := models.SaleItem{
			ProductID:        itemReq.ProductID,
			Quantity:         itemReq.Quantity,
			UnitPrice:        itemReq.UnitPrice,
			Discount:         itemReq.Discount,
			Tax:              itemReq.Tax,
			RevenueAccountID: itemReq.RevenueAccountID,
		}
		
		// Calculate line total
		lineSubtotal := float64(item.Quantity) * item.UnitPrice
		lineTotal := lineSubtotal - item.Discount
		
		// Calculate tax if applicable
		if sale.PPNPercent > 0 {
			item.Tax = lineTotal * sale.PPNPercent / 100
			totalTax += item.Tax
		}
		
		item.TotalPrice = lineTotal + item.Tax
		subtotal += lineTotal
		
		// Calculate COGS (Cost of Goods Sold)
		itemCOGS := float64(item.Quantity) * product.CostPrice
		totalCOGS += itemCOGS
		
		sale.SaleItems = append(sale.SaleItems, item)
	}
	
	// Apply order-level discount
	discountAmount := 0.0
	if sale.DiscountPercent > 0 {
		discountAmount = subtotal * sale.DiscountPercent / 100
	}
	
	// Calculate final totals
	sale.Tax = totalTax
	sale.TotalAmount = subtotal - discountAmount + totalTax + sale.ShippingCost
	sale.OutstandingAmount = sale.TotalAmount
	
	return nil
}

// createSaleAccountingEntries creates comprehensive journal entries for sale
func (s *SalesService) createSaleAccountingEntries(sale *models.Sale, userID uint) error {
	log.Printf("Creating accounting entries for sale %d", sale.ID)
	
	// Get required accounts with fallback
	accountsReceivable, err := s.accountRepo.GetAccountByCode("1201") // Piutang Usaha (Accounts Receivable)
	if err != nil {
		log.Printf("Warning: Piutang Usaha account (1201) not found, trying fallback accounts: %v", err)
		// Try alternative account codes
		accountsReceivable, err = s.accountRepo.GetAccountByCode("1200")
		if err != nil {
			log.Printf("Error: No accounts receivable account found (tried 1201, 1200): %v", err)
			return fmt.Errorf("accounts receivable account not found: %v", err)
		}
	}
	
	salesRevenue, err := s.accountRepo.GetAccountByCode("4101") // Pendapatan Penjualan (Sales Revenue)
	if err != nil {
		log.Printf("Warning: Sales revenue account (4101) not found, trying fallback accounts: %v", err)
		// Try alternative account codes
		salesRevenue, err = s.accountRepo.GetAccountByCode("4100")
		if err != nil {
			log.Printf("Error: No sales revenue account found (tried 4101, 4100): %v", err)
			return fmt.Errorf("sales revenue account not found: %v", err)
		}
	}
	
	cogsAccount, err := s.accountRepo.GetAccountByCode("5101") // Harga Pokok Penjualan (Cost of Goods Sold)
	if err != nil {
		log.Printf("Warning: COGS account (5101) not found, trying fallback accounts: %v", err)
		cogsAccount, err = s.accountRepo.GetAccountByCode("5100")
		if err != nil {
			log.Printf("Error: No COGS account found (tried 5101, 5100): %v", err)
			return fmt.Errorf("COGS account not found: %v", err)
		}
	}
	
	inventoryAccount, err := s.accountRepo.GetAccountByCode("1301") // Persediaan Barang Dagangan (Inventory)
	if err != nil {
		log.Printf("Warning: Inventory account (1301) not found, trying fallback accounts: %v", err)
		inventoryAccount, err = s.accountRepo.GetAccountByCode("1300")
		if err != nil {
			log.Printf("Error: No inventory account found (tried 1301, 1300): %v", err)
			return fmt.Errorf("inventory account not found: %v", err)
		}
	}
	
	// Tax accounts
	var ppnAccount, pphAccount *models.Account
	if sale.PPNPercent > 0 {
		ppnAccount, err = s.accountRepo.GetAccountByCode("2102") // Utang Pajak (Tax Payable) - use existing account
		if err != nil {
			return errors.New("PPN account not found")
		}
	}
	
	if sale.PPhPercent > 0 {
		pphAccount, err = s.accountRepo.GetAccountByCode("2102") // Utang Pajak (Tax Payable) - use existing account
		if err != nil {
			return errors.New("PPh account not found")
		}
	}
	
	// Create journal entry (let BeforeCreate hook generate the code)
	journalEntry := &models.JournalEntry{
		// Code will be auto-generated by BeforeCreate hook
		EntryDate:       sale.Date,
		Description:     fmt.Sprintf("Sales Invoice %s - %s", sale.InvoiceNumber, sale.Customer.Name),
		Reference:       sale.InvoiceNumber,
		ReferenceType:   models.JournalRefSale,
		ReferenceID:     &sale.ID,
		UserID:          userID,
		Status:          models.JournalStatusDraft,
		IsAutoGenerated: true,
	}
	
	journalLines := []models.JournalLine{}
	lineNumber := 1
	
	// Calculate amounts
	subtotal := sale.TotalAmount - sale.Tax - sale.ShippingCost
	totalCOGS := 0.0
	
	// Calculate COGS from items
	for _, item := range sale.SaleItems {
		product, _ := s.productRepo.FindByID(item.ProductID)
		if product != nil {
			totalCOGS += float64(item.Quantity) * product.CostPrice
		}
	}
	
	// 1. Debit: Accounts Receivable (Total Amount)
	journalLines = append(journalLines, models.JournalLine{
		AccountID:     accountsReceivable.ID,
		Description:   fmt.Sprintf("AR - Invoice %s", sale.InvoiceNumber),
		DebitAmount:   sale.TotalAmount,
		CreditAmount:  0,
		LineNumber:    lineNumber,
	})
	lineNumber++
	
	// 2. Credit: Sales Revenue (Subtotal)
	journalLines = append(journalLines, models.JournalLine{
		AccountID:     salesRevenue.ID,
		Description:   fmt.Sprintf("Sales Revenue - Invoice %s", sale.InvoiceNumber),
		DebitAmount:   0,
		CreditAmount:  subtotal,
		LineNumber:    lineNumber,
	})
	lineNumber++
	
	// 3. Credit: PPN Payable (if applicable)
	if sale.PPNPercent > 0 && ppnAccount != nil {
		ppnAmount := subtotal * sale.PPNPercent / 100
		journalLines = append(journalLines, models.JournalLine{
			AccountID:     ppnAccount.ID,
			Description:   fmt.Sprintf("PPN %.0f%% - Invoice %s", sale.PPNPercent, sale.InvoiceNumber),
			DebitAmount:   0,
			CreditAmount:  ppnAmount,
			LineNumber:    lineNumber,
		})
		lineNumber++
	}
	
	// 4. Debit: PPh Receivable (if applicable)
	if sale.PPhPercent > 0 && pphAccount != nil {
		pphAmount := subtotal * sale.PPhPercent / 100
		journalLines = append(journalLines, models.JournalLine{
			AccountID:     pphAccount.ID,
			Description:   fmt.Sprintf("PPh %s %.2f%% - Invoice %s", sale.PPhType, sale.PPhPercent, sale.InvoiceNumber),
			DebitAmount:   pphAmount,
			CreditAmount:  0,
			LineNumber:    lineNumber,
		})
		lineNumber++
	}
	
	// 5. Debit: COGS
	if totalCOGS > 0 {
		journalLines = append(journalLines, models.JournalLine{
			AccountID:     cogsAccount.ID,
			Description:   fmt.Sprintf("COGS - Invoice %s", sale.InvoiceNumber),
			DebitAmount:   totalCOGS,
			CreditAmount:  0,
			LineNumber:    lineNumber,
		})
		lineNumber++
		
		// 6. Credit: Inventory
		journalLines = append(journalLines, models.JournalLine{
			AccountID:     inventoryAccount.ID,
			Description:   fmt.Sprintf("Inventory Reduction - Invoice %s", sale.InvoiceNumber),
			DebitAmount:   0,
			CreditAmount:  totalCOGS,
			LineNumber:    lineNumber,
		})
		lineNumber++
	}
	
	// Calculate totals for validation
	totalDebit := 0.0
	totalCredit := 0.0
	for _, line := range journalLines {
		totalDebit += line.DebitAmount
		totalCredit += line.CreditAmount
	}
	
	// Validate balanced entry
	if totalDebit != totalCredit {
		return fmt.Errorf("unbalanced journal entry: debit %.2f != credit %.2f", totalDebit, totalCredit)
	}
	
	journalEntry.TotalDebit = totalDebit
	journalEntry.TotalCredit = totalCredit
	journalEntry.JournalLines = journalLines
	journalEntry.ValidateBalance()
	
	// Save journal entry with retry logic for duplicate key constraints
	var createErr error
	for retry := 0; retry < 3; retry++ {
		createErr = s.db.Create(journalEntry).Error
		if createErr == nil {
			break // Success, exit retry loop
		}
		
		// Check if it's a duplicate key error
		if strings.Contains(createErr.Error(), "duplicate key value violates unique constraint") {
			log.Printf("Duplicate journal entry code detected on attempt %d, retrying...", retry+1)
			// Clear the code to force regeneration
			journalEntry.Code = ""
			// Small delay to avoid immediate collision
			time.Sleep(time.Millisecond * time.Duration(retry*10+1))
			continue
		}
		
		// For other errors, don't retry
		break
	}
	
	if createErr != nil {
		return fmt.Errorf("failed to create journal entry after retries: %v", createErr)
	}
	
	// Update account balances for each journal line
	for _, line := range journalLines {
		err := s.accountRepo.UpdateBalance(context.Background(), line.AccountID, line.DebitAmount, line.CreditAmount)
		if err != nil {
			return fmt.Errorf("failed to update balance for account %d: %v", line.AccountID, err)
		}
	}
	
	return nil
}

// generateOrderNumber generates order number
func (s *SalesService) generateOrderNumber() string {
	year := time.Now().Year()
	month := time.Now().Month()
	count, _ := s.salesRepo.CountOrdersByMonth(year, int(month))
	return fmt.Sprintf("ORD/%04d/%02d/%04d", year, month, count+1)
}

// generateJournalCode is deprecated - journal codes are now auto-generated by BeforeCreate hook
// This function is kept for backward compatibility but should not be used
func (s *SalesService) generateJournalCode() string {
	// This function is deprecated - codes are now auto-generated
	// Return empty string to force auto-generation
	return ""
}

// Enhanced business rule validations and accounting integrations for sales
// Models are defined in models/sale_extras.go to avoid duplication

// Additional helper functions for sales accounting integration
func (s *SalesService) validateCustomerBusinessRules(customer *models.Contact, totalAmount float64) error {
	// Enhanced customer validation
	if !customer.IsActive {
		return errors.New("cannot create sale for inactive customer")
	}
	
	// Check credit limit
	if customer.CreditLimit > 0 {
		outstandingAmount, err := s.salesRepo.GetCustomerOutstandingAmount(customer.ID)
		if err != nil {
			return fmt.Errorf("failed to check customer outstanding: %v", err)
		}
		
		if (outstandingAmount + totalAmount) > customer.CreditLimit {
			return fmt.Errorf("credit limit exceeded. Available: %.2f, Required: %.2f", 
				customer.CreditLimit - outstandingAmount, totalAmount)
		}
	}
	
	return nil
}

// createSaleReversalJournalEntries creates reversal journal entries when sale is cancelled
func (s *SalesService) createSaleReversalJournalEntries(sale *models.Sale, userID uint, reason string) error {
	// Get required accounts
	accountsReceivable, err := s.accountRepo.GetAccountByCode("1200") // Accounts Receivable
	if err != nil {
		return errors.New("accounts receivable account not found")
	}
	
	salesRevenue, err := s.accountRepo.GetAccountByCode("4100") // Sales Revenue
	if err != nil {
		return errors.New("sales revenue account not found")
	}
	
	cogsAccount, err := s.accountRepo.GetAccountByCode("5100") // Cost of Goods Sold
	if err != nil {
		return errors.New("COGS account not found")
	}
	
	inventoryAccount, err := s.accountRepo.GetAccountByCode("1300") // Inventory
	if err != nil {
		return errors.New("inventory account not found")
	}
	
	// Tax accounts
	var ppnAccount, pphAccount *models.Account
	if sale.PPNPercent > 0 {
		ppnAccount, err = s.accountRepo.GetAccountByCode("2300") // PPN Payable
		if err != nil {
			return errors.New("PPN account not found")
		}
	}
	
	if sale.PPhPercent > 0 {
		pphAccount, err = s.accountRepo.GetAccountByCode("1250") // PPh Receivable
		if err != nil {
			return errors.New("PPh account not found")
		}
	}
	
	// Create reversal journal entry (let BeforeCreate hook generate the code)
	journalEntry := &models.JournalEntry{
		// Code will be auto-generated by BeforeCreate hook
		EntryDate:       time.Now(), // Reversal date is current date
		Description:     fmt.Sprintf("REVERSAL - Sales Invoice %s - %s. Reason: %s", sale.InvoiceNumber, sale.Customer.Name, reason),
		Reference:       sale.InvoiceNumber,
		ReferenceType:   models.JournalRefSale,
		ReferenceID:     &sale.ID,
		UserID:          userID,
		Status:          models.JournalStatusDraft,
		IsAutoGenerated: true,
	}
	
	journalLines := []models.JournalLine{}
	lineNumber := 1
	
	// Calculate amounts
	subtotal := sale.TotalAmount - sale.Tax - sale.ShippingCost
	totalCOGS := 0.0
	
	// Calculate COGS from items
	for _, item := range sale.SaleItems {
		product, _ := s.productRepo.FindByID(item.ProductID)
		if product != nil {
			totalCOGS += float64(item.Quantity) * product.CostPrice
		}
	}
	
	// REVERSAL ENTRIES - All amounts are reversed from original entries
	
	// 1. Credit: Accounts Receivable (Reverse the debit)
	journalLines = append(journalLines, models.JournalLine{
		AccountID:     accountsReceivable.ID,
		Description:   fmt.Sprintf("REVERSAL - AR - Invoice %s", sale.InvoiceNumber),
		DebitAmount:   0,
		CreditAmount:  sale.TotalAmount,
		LineNumber:    lineNumber,
	})
	lineNumber++
	
	// 2. Debit: Sales Revenue (Reverse the credit)
	journalLines = append(journalLines, models.JournalLine{
		AccountID:     salesRevenue.ID,
		Description:   fmt.Sprintf("REVERSAL - Sales Revenue - Invoice %s", sale.InvoiceNumber),
		DebitAmount:   subtotal,
		CreditAmount:  0,
		LineNumber:    lineNumber,
	})
	lineNumber++
	
	// 3. Debit: PPN Payable (Reverse the credit)
	if sale.PPNPercent > 0 && ppnAccount != nil {
		ppnAmount := subtotal * sale.PPNPercent / 100
		journalLines = append(journalLines, models.JournalLine{
			AccountID:     ppnAccount.ID,
			Description:   fmt.Sprintf("REVERSAL - PPN %.0f%% - Invoice %s", sale.PPNPercent, sale.InvoiceNumber),
			DebitAmount:   ppnAmount,
			CreditAmount:  0,
			LineNumber:    lineNumber,
		})
		lineNumber++
	}
	
	// 4. Credit: PPh Receivable (Reverse the debit)
	if sale.PPhPercent > 0 && pphAccount != nil {
		pphAmount := subtotal * sale.PPhPercent / 100
		journalLines = append(journalLines, models.JournalLine{
			AccountID:     pphAccount.ID,
			Description:   fmt.Sprintf("REVERSAL - PPh %s %.2f%% - Invoice %s", sale.PPhType, sale.PPhPercent, sale.InvoiceNumber),
			DebitAmount:   0,
			CreditAmount:  pphAmount,
			LineNumber:    lineNumber,
		})
		lineNumber++
	}
	
	// 5. Credit: COGS (Reverse the debit)
	if totalCOGS > 0 {
		journalLines = append(journalLines, models.JournalLine{
			AccountID:     cogsAccount.ID,
			Description:   fmt.Sprintf("REVERSAL - COGS - Invoice %s", sale.InvoiceNumber),
			DebitAmount:   0,
			CreditAmount:  totalCOGS,
			LineNumber:    lineNumber,
		})
		lineNumber++
		
		// 6. Debit: Inventory (Reverse the credit)
		journalLines = append(journalLines, models.JournalLine{
			AccountID:     inventoryAccount.ID,
			Description:   fmt.Sprintf("REVERSAL - Inventory Restoration - Invoice %s", sale.InvoiceNumber),
			DebitAmount:   totalCOGS,
			CreditAmount:  0,
			LineNumber:    lineNumber,
		})
		lineNumber++
	}
	
	// Calculate totals for validation
	totalDebit := 0.0
	totalCredit := 0.0
	for _, line := range journalLines {
		totalDebit += line.DebitAmount
		totalCredit += line.CreditAmount
	}
	
	// Validate balanced entry
	if totalDebit != totalCredit {
		return fmt.Errorf("unbalanced reversal journal entry: debit %.2f != credit %.2f", totalDebit, totalCredit)
	}
	
	journalEntry.TotalDebit = totalDebit
	journalEntry.TotalCredit = totalCredit
	journalEntry.JournalLines = journalLines
	journalEntry.ValidateBalance()
	
	// Save reversal journal entry with retry logic for duplicate key constraints
	var createErr error
	for retry := 0; retry < 3; retry++ {
		createErr = s.db.Create(journalEntry).Error
		if createErr == nil {
			break // Success, exit retry loop
		}
		
		// Check if it's a duplicate key error
		if strings.Contains(createErr.Error(), "duplicate key value violates unique constraint") {
			log.Printf("Duplicate reversal journal entry code detected on attempt %d, retrying...", retry+1)
			// Clear the code to force regeneration
			journalEntry.Code = ""
			// Small delay to avoid immediate collision
			time.Sleep(time.Millisecond * time.Duration(retry*10+1))
			continue
		}
		
		// For other errors, don't retry
		break
	}
	
	if createErr != nil {
		return fmt.Errorf("failed to create reversal journal entry after retries: %v", createErr)
	}
	
	// Update account balances for each journal line
	for _, line := range journalLines {
		err := s.accountRepo.UpdateBalance(context.Background(), line.AccountID, line.DebitAmount, line.CreditAmount)
		if err != nil {
			return fmt.Errorf("failed to update balance for account %d: %v", line.AccountID, err)
		}
	}
	
	return nil
}
