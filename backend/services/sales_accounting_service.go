package services

import (
	"errors"
	"fmt"
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
	
	// Create journal entry
	journal := &models.Journal{
		Code:          s.generateJournalCode(),
		Date:          sale.Date,
		Description:   fmt.Sprintf("Sales Invoice %s - %s", sale.InvoiceNumber, sale.Customer.Name),
		ReferenceType: models.JournalRefTypeSale,
		ReferenceID:   &sale.ID,
		UserID:        userID,
		Status:        models.JournalStatusPending,
		Period:        sale.Date.Format("2006-01"),
	}
	
	entries := []models.JournalEntry{}
	
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
	entries = append(entries, models.JournalEntry{
		AccountID:    accountsReceivable.ID,
		Description:  fmt.Sprintf("AR - Invoice %s", sale.InvoiceNumber),
		DebitAmount:  sale.TotalAmount,
		CreditAmount: 0,
	})
	
	// 2. Credit: Sales Revenue (Subtotal)
	entries = append(entries, models.JournalEntry{
		AccountID:    salesRevenue.ID,
		Description:  fmt.Sprintf("Sales Revenue - Invoice %s", sale.InvoiceNumber),
		DebitAmount:  0,
		CreditAmount: subtotal,
	})
	
	// 3. Credit: PPN Payable (if applicable)
	if sale.PPNPercent > 0 && ppnAccount != nil {
		ppnAmount := subtotal * sale.PPNPercent / 100
		entries = append(entries, models.JournalEntry{
			AccountID:    ppnAccount.ID,
			Description:  fmt.Sprintf("PPN %.0f%% - Invoice %s", sale.PPNPercent, sale.InvoiceNumber),
			DebitAmount:  0,
			CreditAmount: ppnAmount,
		})
	}
	
	// 4. Debit: PPh Receivable (if applicable)
	if sale.PPhPercent > 0 && pphAccount != nil {
		pphAmount := subtotal * sale.PPhPercent / 100
		entries = append(entries, models.JournalEntry{
			AccountID:    pphAccount.ID,
			Description:  fmt.Sprintf("PPh %s %.2f%% - Invoice %s", sale.PPhType, sale.PPhPercent, sale.InvoiceNumber),
			DebitAmount:  pphAmount,
			CreditAmount: 0,
		})
	}
	
	// 5. Debit: COGS
	if totalCOGS > 0 {
		entries = append(entries, models.JournalEntry{
			AccountID:    cogsAccount.ID,
			Description:  fmt.Sprintf("COGS - Invoice %s", sale.InvoiceNumber),
			DebitAmount:  totalCOGS,
			CreditAmount: 0,
		})
		
		// 6. Credit: Inventory
		entries = append(entries, models.JournalEntry{
			AccountID:    inventoryAccount.ID,
			Description:  fmt.Sprintf("Inventory Reduction - Invoice %s", sale.InvoiceNumber),
			DebitAmount:  0,
			CreditAmount: totalCOGS,
		})
	}
	
	// Calculate totals for validation
	totalDebit := 0.0
	totalCredit := 0.0
	for _, entry := range entries {
		totalDebit += entry.DebitAmount
		totalCredit += entry.CreditAmount
	}
	
	// Validate balanced entry
	if totalDebit != totalCredit {
		return fmt.Errorf("unbalanced journal entry: debit %.2f != credit %.2f", totalDebit, totalCredit)
	}
	
	journal.TotalDebit = totalDebit
	journal.TotalCredit = totalCredit
	journal.JournalEntries = entries
	
	// Save journal through repository
	return s.salesRepo.CreateJournal(journal)
}

// generateOrderNumber generates order number
func (s *SalesService) generateOrderNumber() string {
	year := time.Now().Year()
	month := time.Now().Month()
	count, _ := s.salesRepo.CountOrdersByMonth(year, int(month))
	return fmt.Sprintf("ORD/%04d/%02d/%04d", year, month, count+1)
}

// generateJournalCode generates unique journal code
func (s *SalesService) generateJournalCode() string {
	year := time.Now().Year()
	month := time.Now().Month()
	count, _ := s.salesRepo.CountJournalsByMonth(year, int(month))
	return fmt.Sprintf("JV/%04d/%02d/%04d", year, month, count+1)
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
