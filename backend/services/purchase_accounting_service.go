package services

import (
	"errors"
	"fmt"
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
			return fmt.Errorf("product %d not found", itemReq.ProductID)
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
		item.TotalPrice = lineSubtotal - item.Discount + item.Tax
		
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
	
	// Set all calculated fields
	purchase.SubtotalBeforeDiscount = subtotalBeforeDiscount
	purchase.ItemDiscountAmount = itemDiscountAmount
	purchase.OrderDiscountAmount = orderDiscountAmount
	purchase.NetBeforeTax = subtotalBeforeDiscount - itemDiscountAmount - orderDiscountAmount
	
	// Calculate tax on net amount
	if purchase.TaxAmount == 0 && purchase.NetBeforeTax > 0 {
		// Default PPN 11% if not specified
		purchase.TaxAmount = purchase.NetBeforeTax * 0.11
	}
	
	purchase.TotalAmount = purchase.NetBeforeTax + purchase.TaxAmount
	
	// Set outstanding amount to total amount for new purchase
	purchase.OutstandingAmount = purchase.TotalAmount
	
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
		item.TotalPrice = lineSubtotal - item.Discount + item.Tax
		
		subtotalBeforeDiscount += lineSubtotal
		itemDiscountAmount += item.Discount
	}
	
	// Calculate order-level discount
	orderDiscountAmount := 0.0
	if purchase.Discount > 0 {
		orderDiscountAmount = (subtotalBeforeDiscount - itemDiscountAmount) * purchase.Discount / 100
	}
	
	// Set all calculated fields
	purchase.SubtotalBeforeDiscount = subtotalBeforeDiscount
	purchase.ItemDiscountAmount = itemDiscountAmount
	purchase.OrderDiscountAmount = orderDiscountAmount
	purchase.NetBeforeTax = subtotalBeforeDiscount - itemDiscountAmount - orderDiscountAmount
	purchase.TotalAmount = purchase.NetBeforeTax + purchase.TaxAmount
	
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
		item.TotalPrice = float64(item.Quantity)*item.UnitPrice - item.Discount + item.Tax
		purchase.PurchaseItems = append(purchase.PurchaseItems, item)
	}
	
	return nil
}

// createPurchaseAccountingEntries creates journal entries for purchase
func (s *PurchaseService) createPurchaseAccountingEntries(purchase *models.Purchase, userID uint) error {
	// TODO: Implement GetAccountByCode in AccountRepository
	// For now, use default account IDs
	// inventoryAccount, err := s.accountRepo.GetAccountByCode("1300") // Inventory
	// if err != nil {
	//	// Try expense account for non-inventory purchases
	//	inventoryAccount, err = s.accountRepo.GetAccountByCode("5200") // Purchase Expense
	//	if err != nil {
	//		return errors.New("inventory/expense account not found")
	//	}
	// }
	inventoryAccountID := uint(4) // Default inventory account ID
	
	// accountsPayable, err := s.accountRepo.GetAccountByCode("2100") // Accounts Payable
	// if err != nil {
	//	return errors.New("accounts payable account not found")
	// }
	accountsPayableID := uint(3) // Default AP account ID
	
	// ppnAccount, err := s.accountRepo.GetAccountByCode("1240") // PPN Receivable (input tax)
	// if err != nil && purchase.TaxAmount > 0 {
	//	return errors.New("PPN account not found")
	// }
	ppnAccountID := uint(5) // Default PPN account ID
	
	// Create journal entry
	journal := &models.Journal{
		Code:          s.generatePurchaseJournalCode(),
		Date:          purchase.Date,
		Description:   fmt.Sprintf("Purchase %s - %s", purchase.Code, purchase.Vendor.Name),
		ReferenceType: models.JournalRefTypePurchase,
		ReferenceID:   &purchase.ID,
		UserID:        userID,
		Status:        models.JournalStatusPending,
		Period:        purchase.Date.Format("2006-01"),
	}
	
	entries := []models.JournalEntry{}
	
	// Group items by expense account
	accountTotals := make(map[uint]float64)
	for _, item := range purchase.PurchaseItems {
		accountID := item.ExpenseAccountID
		if accountID == 0 {
			accountID = inventoryAccountID
		}
		accountTotals[accountID] += item.TotalPrice - item.Tax
	}
	
	// 1. Debit: Inventory/Expense accounts
	for accountID, amount := range accountTotals {
		account, _ := s.accountRepo.FindByID(nil, accountID)
		accountName := "Inventory"
		if account != nil {
			accountName = account.Name
		}
		entries = append(entries, models.JournalEntry{
			AccountID:    accountID,
			Description:  fmt.Sprintf("%s - Purchase %s", accountName, purchase.Code),
			DebitAmount:  amount,
			CreditAmount: 0,
		})
	}
	
	// 2. Debit: PPN Receivable (if applicable)
	if purchase.TaxAmount > 0 {
		entries = append(entries, models.JournalEntry{
			AccountID:    ppnAccountID,
			Description:  fmt.Sprintf("PPN Input - Purchase %s", purchase.Code),
			DebitAmount:  purchase.TaxAmount,
			CreditAmount: 0,
		})
	}
	
	// 3. Credit: Accounts Payable
	entries = append(entries, models.JournalEntry{
		AccountID:    accountsPayableID,
		Description:  fmt.Sprintf("AP - Purchase %s", purchase.Code),
		DebitAmount:  0,
		CreditAmount: purchase.TotalAmount,
	})
	
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
	return s.purchaseRepo.CreateJournal(journal)
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
	count, err := s.purchaseRepo.CountByMonth(year, int(month))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PO/%04d/%02d/%04d", year, month, count+1), nil
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
	return fmt.Sprintf("PJ/%04d/%02d/%04d", year, month, count+1)
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
