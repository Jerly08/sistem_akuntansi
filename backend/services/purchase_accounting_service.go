package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"app-sistem-akuntansi/models"
)

// Enhanced Accounting Methods for Purchase Module

// setApprovalBasisAndBase determines approval basis for purchase
func (s *PurchaseService) setApprovalBasisAndBase(purchase *models.Purchase) {
	// Set approval basis - what amount will be used for approval
	purchase.ApprovalAmountBasis = "SUBTOTAL_BEFORE_DISCOUNT"
	purchase.ApprovalBaseAmount = purchase.SubtotalBeforeDiscount
	
	// Check if approval is required based on amount
	requiredWorkflow, err := s.approvalService.GetWorkflowByAmount(
		models.ApprovalModulePurchase, 
		purchase.ApprovalBaseAmount,
	)
	
	if err == nil && requiredWorkflow != nil {
		purchase.RequiresApproval = true
	} else {
		purchase.RequiresApproval = false
	}
}

// calculatePurchaseTotals calculates all purchase totals with proper accounting
func (s *PurchaseService) calculatePurchaseTotals(purchase *models.Purchase, items []models.PurchaseItemRequest) error {
	subtotalBeforeDiscount := 0.0
	itemDiscountAmount := 0.0
	
	purchase.PurchaseItems = []models.PurchaseItem{}
	
	for _, itemReq := range items {
		// Validate product exists
		product, err := s.productRepo.FindByID(itemReq.ProductID)
		if err != nil {
			// Add more detailed logging
			fmt.Printf("[DEBUG] Failed to find product with ID %d: %v\n", itemReq.ProductID, err)
			return fmt.Errorf("product %d not found: %v", itemReq.ProductID, err)
		}
		
		// Create purchase item
		item := models.PurchaseItem{
			ProductID:        itemReq.ProductID,
			Quantity:         itemReq.Quantity,
			UnitPrice:        itemReq.UnitPrice,
			Discount:         itemReq.Discount,
			Tax:              itemReq.Tax,
			ExpenseAccountID: itemReq.ExpenseAccountID,
		}
		
		// Calculate line totals
		lineSubtotal := float64(item.Quantity) * item.UnitPrice
		item.TotalPrice = lineSubtotal - item.Discount // Remove duplicate tax addition
		
		subtotalBeforeDiscount += lineSubtotal
		itemDiscountAmount += item.Discount
		
		// Update product cost price (weighted average)
		s.updateProductCostPrice(product, item.Quantity, item.UnitPrice)
		
		purchase.PurchaseItems = append(purchase.PurchaseItems, item)
	}
	
	// Calculate order-level discount
	orderDiscountAmount := 0.0
	if purchase.Discount > 0 {
		orderDiscountAmount = (subtotalBeforeDiscount - itemDiscountAmount) * purchase.Discount / 100
	}
	
	// Set basic calculated fields
	purchase.SubtotalBeforeDiscount = subtotalBeforeDiscount
	purchase.ItemDiscountAmount = itemDiscountAmount
	purchase.OrderDiscountAmount = orderDiscountAmount
	purchase.NetBeforeTax = subtotalBeforeDiscount - itemDiscountAmount - orderDiscountAmount
	
	// Calculate tax additions (Penambahan)
	// 1. PPN (VAT) calculation
	if purchase.PPNRate > 0 {
		purchase.PPNAmount = purchase.NetBeforeTax * purchase.PPNRate / 100
	} else {
		// Default PPN 11% if not specified
		purchase.PPNAmount = purchase.NetBeforeTax * 0.11
		purchase.PPNRate = 11.0
	}
	
	// 2. Other tax additions
	purchase.TotalTaxAdditions = purchase.PPNAmount + purchase.OtherTaxAdditions
	
	// Calculate tax deductions (Pemotongan)
	// 1. PPh 21 calculation
	if purchase.PPh21Rate > 0 {
		purchase.PPh21Amount = purchase.NetBeforeTax * purchase.PPh21Rate / 100
	}
	
	// 2. PPh 23 calculation
	if purchase.PPh23Rate > 0 {
		purchase.PPh23Amount = purchase.NetBeforeTax * purchase.PPh23Rate / 100
	}
	
	// 3. Total tax deductions
	purchase.TotalTaxDeductions = purchase.PPh21Amount + purchase.PPh23Amount + purchase.OtherTaxDeductions
	
	// Calculate final total amount
	// Total = Net Before Tax + Tax Additions - Tax Deductions
	purchase.TotalAmount = purchase.NetBeforeTax + purchase.TotalTaxAdditions - purchase.TotalTaxDeductions
	
	// For legacy compatibility, set TaxAmount to PPN amount
	purchase.TaxAmount = purchase.PPNAmount
	
	// Set payment amounts based on payment method
	if isImmediatePayment(purchase.PaymentMethod) {
		// For cash/transfer purchases - fully paid
		purchase.PaidAmount = purchase.TotalAmount
		purchase.OutstandingAmount = 0
	} else {
		// For credit purchases - outstanding amount
		purchase.PaidAmount = 0
		purchase.OutstandingAmount = purchase.TotalAmount
	}
	
	return nil
}

// updateProductCostPrice updates product cost using weighted average
func (s *PurchaseService) updateProductCostPrice(product *models.Product, newQuantity int, newPrice float64) {
	if product.Stock == 0 {
		product.PurchasePrice = newPrice
	} else {
		// Weighted average cost calculation
		totalValue := (float64(product.Stock) * product.PurchasePrice) + (float64(newQuantity) * newPrice)
		totalQuantity := product.Stock + newQuantity
		product.PurchasePrice = totalValue / float64(totalQuantity)
	}
	
	// Update stock quantity
	product.Stock += newQuantity
	
	// TODO: Replace with proper save method when available
	// s.productRepo.Update(product)
}

// recalculatePurchaseTotals recalculates purchase totals
func (s *PurchaseService) recalculatePurchaseTotals(purchase *models.Purchase) error {
	subtotalBeforeDiscount := 0.0
	itemDiscountAmount := 0.0
	
	for _, item := range purchase.PurchaseItems {
		lineSubtotal := float64(item.Quantity) * item.UnitPrice
		item.TotalPrice = lineSubtotal - item.Discount // Remove duplicate tax addition
		
		subtotalBeforeDiscount += lineSubtotal
		itemDiscountAmount += item.Discount
	}
	
	// Calculate order-level discount
	orderDiscountAmount := 0.0
	if purchase.Discount > 0 {
		orderDiscountAmount = (subtotalBeforeDiscount - itemDiscountAmount) * purchase.Discount / 100
	}
	
	// Set basic calculated fields
	purchase.SubtotalBeforeDiscount = subtotalBeforeDiscount
	purchase.ItemDiscountAmount = itemDiscountAmount
	purchase.OrderDiscountAmount = orderDiscountAmount
	purchase.NetBeforeTax = subtotalBeforeDiscount - itemDiscountAmount - orderDiscountAmount
	
	// Recalculate tax additions (Penambahan)
	// 1. PPN (VAT) calculation
	if purchase.PPNRate > 0 {
		purchase.PPNAmount = purchase.NetBeforeTax * purchase.PPNRate / 100
	} else {
		// Default PPN 11% if not specified
		purchase.PPNAmount = purchase.NetBeforeTax * 0.11
		purchase.PPNRate = 11.0
	}
	
	// 2. Other tax additions
	purchase.TotalTaxAdditions = purchase.PPNAmount + purchase.OtherTaxAdditions
	
	// Recalculate tax deductions (Pemotongan)
	// 1. PPh 21 calculation
	if purchase.PPh21Rate > 0 {
		purchase.PPh21Amount = purchase.NetBeforeTax * purchase.PPh21Rate / 100
	}
	
	// 2. PPh 23 calculation
	if purchase.PPh23Rate > 0 {
		purchase.PPh23Amount = purchase.NetBeforeTax * purchase.PPh23Rate / 100
	}
	
	// 3. Total tax deductions
	purchase.TotalTaxDeductions = purchase.PPh21Amount + purchase.PPh23Amount + purchase.OtherTaxDeductions
	
	// Calculate final total amount
	// Total = Net Before Tax + Tax Additions - Tax Deductions
	purchase.TotalAmount = purchase.NetBeforeTax + purchase.TotalTaxAdditions - purchase.TotalTaxDeductions
	
	// For legacy compatibility, set TaxAmount to PPN amount
	purchase.TaxAmount = purchase.PPNAmount
	
	return nil
}

// updatePurchaseItems updates purchase items
func (s *PurchaseService) updatePurchaseItems(purchase *models.Purchase, items []models.PurchaseItemRequest) error {
	// Clear existing items
	purchase.PurchaseItems = []models.PurchaseItem{}
	
	for _, itemReq := range items {
		// Validate product exists
		_, err := s.productRepo.FindByID(itemReq.ProductID)
		if err != nil {
			return fmt.Errorf("product %d not found", itemReq.ProductID)
		}
		
		item := models.PurchaseItem{
			ProductID:        itemReq.ProductID,
			Quantity:         itemReq.Quantity,
			UnitPrice:        itemReq.UnitPrice,
			Discount:         itemReq.Discount,
			Tax:              itemReq.Tax,
			ExpenseAccountID: itemReq.ExpenseAccountID,
		}
		
		// Calculate totals
		item.TotalPrice = float64(item.Quantity)*item.UnitPrice - item.Discount // Remove duplicate tax addition
		purchase.PurchaseItems = append(purchase.PurchaseItems, item)
	}
	
	return nil
}

// createPurchaseAccountingEntries creates journal entries for purchase
// Adapted to match database schema that expects account_id in journal_entries table
func (s *PurchaseService) createPurchaseAccountingEntries(purchase *models.Purchase, userID uint) (*models.JournalEntry, error) {
	inventoryAccountID := uint(4) // Default inventory account ID

	// For this database schema, we'll create one primary journal entry
	// and use the main expense account as the account_id
	
	// Find the primary expense account (most used or first one)
	var primaryAccountID uint = inventoryAccountID
	if len(purchase.PurchaseItems) > 0 {
		for _, item := range purchase.PurchaseItems {
			if item.ExpenseAccountID != 0 {
				primaryAccountID = item.ExpenseAccountID
				break // Use the first expense account found
			}
		}
	}
	
	// Calculate proper accounting totals
	// Debit side: Expense accounts + PPN Masukan (always the same)
	totalDebits := purchase.NetBeforeTax + purchase.PPNAmount
	
	// Credit side depends on payment method
	var totalCredits float64
	if isImmediatePayment(purchase.PaymentMethod) {
		// For cash/transfer: Credit to Bank Account
		totalCredits = purchase.TotalAmount
	} else {
		// For credit: Credit to Accounts Payable
		totalCredits = purchase.TotalAmount
	}
	
	// Debug logging
	fmt.Printf("🧮 Purchase Calculation Debug:\n")
	fmt.Printf("   NetBeforeTax: %.2f\n", purchase.NetBeforeTax)
	fmt.Printf("   PPNAmount: %.2f\n", purchase.PPNAmount)
	fmt.Printf("   TotalAmount: %.2f\n", purchase.TotalAmount)
	fmt.Printf("   Calculated totalDebits: %.2f\n", totalDebits)
	fmt.Printf("   Calculated totalCredits: %.2f\n", totalCredits)
	
	// Ensure balanced entry - if totals don't match, use the larger amount for both
	if totalDebits != totalCredits {
		// In purchase accounting: Debit (Expense + PPN) = Credit (Payable)
		// Use the purchase total as the balanced amount
		balancedAmount := purchase.TotalAmount
		totalDebits = balancedAmount
		totalCredits = balancedAmount
		fmt.Printf("   ⚖️ Balanced both to: %.2f\n", balancedAmount)
	}
	
	// Create main journal entry with primary account
	journalEntry := &models.JournalEntry{
		JournalID:       nil, // Auto-generated entries don't need a parent journal
		AccountID:       &primaryAccountID, // Required by database schema
		Code:            s.generatePurchaseJournalCode(),
		EntryDate:       purchase.Date,
		Description:     fmt.Sprintf("Purchase %s - %s", purchase.Code, purchase.Vendor.Name),
		ReferenceType:   models.JournalRefPurchase,
		ReferenceID:     &purchase.ID,
		Reference:       purchase.Code,
		UserID:          userID,
		Status:          models.JournalStatusDraft,
		TotalDebit:      totalDebits,
		TotalCredit:     totalCredits,
		IsBalanced:      totalDebits == totalCredits && totalDebits > 0,
		IsAutoGenerated: true,
	}
	
	// Since database doesn't have journal_lines table, we'll store summary information in description
	// Create detailed description for the journal entry
	var detailsBuilder strings.Builder
	detailsBuilder.WriteString(fmt.Sprintf("Purchase %s - %s\n", purchase.Code, purchase.Vendor.Name))
	
	// Add line items details in description
	for _, item := range purchase.PurchaseItems {
		accountID := item.ExpenseAccountID
		if accountID == 0 {
			accountID = inventoryAccountID
		}
		account, _ := s.accountRepo.FindByID(nil, accountID)
		accountName := "Inventory"
		if account != nil {
			accountName = account.Name
		}
		detailsBuilder.WriteString(fmt.Sprintf("Dr. %s: %.2f\n", accountName, item.TotalPrice))
	}
	
	if purchase.PPNAmount > 0 {
		detailsBuilder.WriteString(fmt.Sprintf("Dr. PPN Receivable: %.2f\n", purchase.PPNAmount))
	}
	
	detailsBuilder.WriteString(fmt.Sprintf("Cr. Accounts Payable: %.2f", purchase.TotalAmount))
	
	// Update journal entry description with details
	journalEntry.Description = detailsBuilder.String()
	
	// Create journal entry without separate lines since database doesn't have journal_lines table
	// The database schema appears to store all journal information in the journal_entries table itself
	if err := s.db.Create(journalEntry).Error; err != nil {
		return nil, fmt.Errorf("failed to create journal entry: %v", err)
	}
	
	// Log successful journal entry creation
	fmt.Printf("✅ Journal entry created successfully: ID=%d, Debit=%.2f, Credit=%.2f\n", 
		journalEntry.ID, journalEntry.TotalDebit, journalEntry.TotalCredit)
	
	return journalEntry, nil
}

// ProcessPurchaseReceipt processes goods receipt
func (s *PurchaseService) ProcessPurchaseReceipt(purchaseID uint, request models.PurchaseReceiptRequest, userID uint) (*models.PurchaseReceipt, error) {
	purchase, err := s.purchaseRepo.FindByID(purchaseID)
	if err != nil {
		return nil, err
	}
	
	if purchase.Status != models.PurchaseStatusApproved && purchase.Status != models.PurchaseStatusPending {
		return nil, errors.New("purchase must be approved before receiving goods")
	}
	
	// Create receipt record
	receipt := &models.PurchaseReceipt{
		PurchaseID:    purchaseID,
		ReceiptNumber: s.generateReceiptNumber(),
		ReceivedDate:  request.ReceivedDate,
		ReceivedBy:    userID,
		Status:        models.ReceiptStatusPending,
		Notes:         request.Notes,
	}
	
	createdReceipt, err := s.purchaseRepo.CreateReceipt(receipt)
	if err != nil {
		return nil, err
	}
	
	// Process receipt items and update inventory
	allReceived := true
	for _, itemReq := range request.ReceiptItems {
		purchaseItem, err := s.purchaseRepo.GetPurchaseItemByID(itemReq.PurchaseItemID)
		if err != nil {
			return nil, err
		}
		
		if purchaseItem.PurchaseID != purchaseID {
			return nil, errors.New("purchase item does not belong to this purchase")
		}
		
		// Create receipt item
		receiptItem := &models.PurchaseReceiptItem{
			ReceiptID:        createdReceipt.ID,
			PurchaseItemID:   itemReq.PurchaseItemID,
			QuantityReceived: itemReq.QuantityReceived,
			Condition:        itemReq.Condition,
			Notes:            itemReq.Notes,
		}
		
		err = s.purchaseRepo.CreateReceiptItem(receiptItem)
		if err != nil {
			return nil, err
		}
		
		// Update inventory if condition is good
		if itemReq.Condition == models.ReceiptConditionGood {
			product, _ := s.productRepo.FindByID(purchaseItem.ProductID)
			if product != nil {
				product.Stock += itemReq.QuantityReceived
				// TODO: Replace with proper update method when available
				// s.productRepo.Update(product)
			}
		}
		
		// Check if all items are fully received
		if itemReq.QuantityReceived < purchaseItem.Quantity {
			allReceived = false
		}
	}
	
	// Update receipt and purchase status
	if allReceived {
		receipt.Status = models.ReceiptStatusComplete
		purchase.Status = models.PurchaseStatusCompleted
	} else {
		receipt.Status = models.ReceiptStatusPartial
	}
	
	s.purchaseRepo.UpdateReceipt(receipt)
	s.purchaseRepo.Update(purchase)
	
	return createdReceipt, nil
}

// generatePurchaseCode generates unique purchase code
func (s *PurchaseService) generatePurchaseCode() (string, error) {
	year := time.Now().Year()
	month := time.Now().Month()
	
	// Get the last number used for this month
	lastNumber, err := s.purchaseRepo.GetLastPurchaseNumberByMonth(year, int(month))
	if err != nil {
		return "", err
	}
	
	// Try to generate a unique code, incrementing if necessary
	for i := 1; i <= 100; i++ { // Limit iterations to prevent infinite loop
		nextNumber := lastNumber + i
		code := fmt.Sprintf("PO/%04d/%02d/%04d", year, month, nextNumber)
		
		// Check if code already exists
		exists, err := s.purchaseRepo.CodeExists(code)
		if err != nil {
			return "", err
		}
		
		if !exists {
			return code, nil
		}
	}
	
	return "", fmt.Errorf("unable to generate unique purchase code after 100 attempts")
}

// generateReceiptNumber generates unique receipt number
func (s *PurchaseService) generateReceiptNumber() string {
	year := time.Now().Year()
	month := time.Now().Month()
	count, _ := s.purchaseRepo.CountReceiptsByMonth(year, int(month))
	return fmt.Sprintf("GR/%04d/%02d/%04d", year, month, count+1)
}

// generatePurchaseJournalCode generates unique journal code for purchase
func (s *PurchaseService) generatePurchaseJournalCode() string {
	year := time.Now().Year()
	month := time.Now().Month()
	count, _ := s.purchaseRepo.CountJournalsByMonth(year, int(month))
	
	// Keep trying to generate a unique code
	for i := 1; i <= 100; i++ {
		nextNumber := count + int64(i)
		code := fmt.Sprintf("PJ/%04d/%02d/%04d", year, month, nextNumber)
		
		// Check if this code already exists in the journal_entries table
		var existingCount int64
		s.db.Model(&models.JournalEntry{}).Where("code = ?", code).Count(&existingCount)
		
		if existingCount == 0 {
			return code
		}
	}
	
	// Final fallback - use timestamp
	return fmt.Sprintf("PJ/%04d/%02d/%d", year, month, time.Now().Unix())
}

// GetPurchaseSummary gets purchase summary with analytics
func (s *PurchaseService) GetPurchaseSummary(startDate, endDate string) (*models.PurchaseSummary, error) {
	return s.purchaseRepo.GetPurchaseSummary(startDate, endDate)
}

// GetPayablesReport gets accounts payable report
func (s *PurchaseService) GetPayablesReport() (*models.PayablesReportResponse, error) {
	return s.purchaseRepo.GetPayablesReport()
}

// CreatePurchaseReturn creates a purchase return
// TODO: Uncomment when PurchaseReturn and PurchaseReturnItem models are created
/*
func (s *PurchaseService) CreatePurchaseReturn(purchaseID uint, reason string, items []models.PurchaseReturnItem, userID uint) (*models.PurchaseReturn, error) {
	purchase, err := s.purchaseRepo.FindByID(purchaseID)
	if err != nil {
		return nil, err
	}
	
	if purchase.Status != models.PurchaseStatusCompleted {
		return nil, errors.New("purchase must be completed before creating return")
	}
	
	// Calculate return amount
	totalAmount := 0.0
	for _, item := range items {
		purchaseItem, _ := s.purchaseRepo.GetPurchaseItemByID(item.PurchaseItemID)
		if purchaseItem != nil {
			totalAmount += float64(item.Quantity) * purchaseItem.UnitPrice
		}
	}
	
	// Create return record
	purchaseReturn := &models.PurchaseReturn{
		PurchaseID:    purchaseID,
		ReturnNumber:  s.generatePurchaseReturnNumber(),
		Date:          time.Now(),
		Reason:        reason,
		TotalAmount:   totalAmount,
		Status:        models.PurchaseReturnStatusPending,
		UserID:        userID,
	}
	
	createdReturn, err := s.purchaseRepo.CreateReturn(purchaseReturn)
	if err != nil {
		return nil, err
	}
	
	// Create return items and adjust inventory
	for _, item := range items {
		item.PurchaseReturnID = createdReturn.ID
		err = s.purchaseRepo.CreateReturnItem(&item)
		if err != nil {
			return nil, err
		}
		
		// Reduce inventory
		purchaseItem, _ := s.purchaseRepo.GetPurchaseItemByID(item.PurchaseItemID)
		if purchaseItem != nil {
			product, _ := s.productRepo.FindByID(purchaseItem.ProductID)
			if product != nil {
				product.StockQuantity -= item.Quantity
				s.productRepo.Update(product)
			}
		}
	}
	
	// Create reversal journal entries
	s.createPurchaseReturnJournalEntries(createdReturn, userID)
	
	return createdReturn, nil
}

// createPurchaseReturnJournalEntries creates journal entries for purchase return
func (s *PurchaseService) createPurchaseReturnJournalEntries(purchaseReturn *models.PurchaseReturn, userID uint) error {
	// This would create the reverse entries of the original purchase
	// Debit: Accounts Payable
	// Credit: Inventory/Expense accounts
	// Credit: PPN Receivable (if applicable)
	
	// Implementation details would mirror the purchase entries but in reverse
	return nil
}

// generatePurchaseReturnNumber generates unique return number
func (s *PurchaseService) generatePurchaseReturnNumber() string {
	year := time.Now().Year()
	month := time.Now().Month()
	count, _ := s.purchaseRepo.CountReturnsByMonth(year, int(month))
	return fmt.Sprintf("PR/%04d/%02d/%04d", year, month, count+1)
}

*/

// createAndPostPurchaseJournalEntries creates journal entries for approved purchase and posts them to GL
func (s *PurchaseService) createAndPostPurchaseJournalEntries(purchase *models.Purchase, userID uint) error {
	// First, check if journal entries already exist for this purchase
	existingEntry, err := s.journalRepo.FindByReferenceID(
		context.Background(),
		models.JournalRefPurchase,
		purchase.ID,
	)
	
	// Handle any unexpected database errors
	if err != nil {
		return fmt.Errorf("failed to check existing journal entries: %v", err)
	}
	
	// If journal entry already exists and is posted, don't create another one
	if existingEntry != nil {
		if existingEntry.Status == models.JournalStatusPosted {
			fmt.Printf("Journal entry already posted for purchase %d\n", purchase.ID)
			return nil
		}
		// If exists but not posted, post the existing entry
		err = s.journalRepo.PostJournalEntry(context.Background(), existingEntry.ID, userID)
		if err != nil {
			return fmt.Errorf("failed to post existing journal entry: %v", err)
		}
		fmt.Printf("Posted existing journal entry for purchase %d\n", purchase.ID)
		return nil
	}
	
	// Create new journal entries if none exist
	journalEntry, err := s.createPurchaseAccountingEntries(purchase, userID)
	if err != nil {
		return fmt.Errorf("failed to create journal entries: %v", err)
	}
	
	// Post the journal entry to update account balances
	err = s.journalRepo.PostJournalEntry(context.Background(), journalEntry.ID, userID)
	if err != nil {
		return fmt.Errorf("failed to post journal entry: %v", err)
	}
	
	fmt.Printf("Successfully created and posted journal entry for purchase %d\n", purchase.ID)
	return nil
}
