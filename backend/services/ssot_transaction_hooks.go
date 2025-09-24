package services

import (
	"fmt"
	"log"

	"app-sistem-akuntansi/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SSOTTransactionHooks provides hooks for integrating all transactions with SSOT journal system
type SSOTTransactionHooks struct {
	db                        *gorm.DB
	unifiedJournalService     *UnifiedJournalService
	ssotReportIntegrationSvc  *SSOTReportIntegrationService
}

// NewSSOTTransactionHooks creates a new SSOT transaction hooks service
func NewSSOTTransactionHooks(
	db *gorm.DB,
	unifiedJournalService *UnifiedJournalService,
	ssotReportIntegrationSvc *SSOTReportIntegrationService,
) *SSOTTransactionHooks {
	return &SSOTTransactionHooks{
		db:                       db,
		unifiedJournalService:    unifiedJournalService,
		ssotReportIntegrationSvc: ssotReportIntegrationSvc,
	}
}

// OnSaleCreated creates journal entries when a sale is created/confirmed
func (h *SSOTTransactionHooks) OnSaleCreated(saleID uint64, userID uint64) error {
	var sale models.Sale
	if err := h.db.Preload("SaleItems.Product").Preload("Customer").First(&sale, saleID).Error; err != nil {
		return fmt.Errorf("failed to load sale: %w", err)
	}

	// Only create journal entries for confirmed/invoiced sales
	if sale.Status != "CONFIRMED" && sale.Status != "INVOICED" {
		return nil
	}

	// Create journal entry request
	journalReq := &JournalEntryRequest{
		SourceType:  models.SSOTSourceTypeSale,
		SourceID:    &saleID,
		Reference:   sale.InvoiceNumber,
	EntryDate:   sale.Date,
		Description: fmt.Sprintf("Sale Invoice #%s - %s", sale.InvoiceNumber, sale.Customer.Name),
		CreatedBy:   userID,
		AutoPost:    true,
	}

	var journalLines []JournalLineRequest

	// Accounts Receivable (Debit) - or Cash if cash sale
	if sale.PaymentMethod == "CASH" {
		// Cash Account (Debit)
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    getCashAccountID(), // You need to implement this
			Description:  "Cash received from sale",
		DebitAmount:  decimal.NewFromFloat(sale.TotalAmount),
			CreditAmount: decimal.Zero,
		})
	} else {
		// Accounts Receivable (Debit)
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    getAccountsReceivableAccountID(), // You need to implement this
			Description:  "Accounts receivable from sale",
		DebitAmount:  decimal.NewFromFloat(sale.TotalAmount),
			CreditAmount: decimal.Zero,
		})
	}

	// Sales Revenue (Credit)
	// Calculate net sales (total - tax)
	totalTax := sale.TotalTaxAdditions // Using new tax field
	netSales := decimal.NewFromFloat(sale.TotalAmount - totalTax)
	journalLines = append(journalLines, JournalLineRequest{
		AccountID:    getSalesRevenueAccountID(), // You need to implement this
		Description:  "Sales revenue",
		DebitAmount:  decimal.Zero,
		CreditAmount: netSales,
	})

	// Tax Payable (Credit) - if there's tax
	if totalTax > 0 {
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    getTaxPayableAccountID(), // You need to implement this
			Description:  "Tax payable",
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromFloat(totalTax),
		})
	}

	// Cost of Goods Sold entries (if inventory items)
	for _, item := range sale.SaleItems {
		if item.Product.CostPrice > 0 {
			cogsAmount := decimal.NewFromFloat(item.Product.CostPrice * float64(item.Quantity))
			
			// COGS (Debit)
			journalLines = append(journalLines, JournalLineRequest{
				AccountID:    getCOGSAccountID(), // You need to implement this
				Description:  fmt.Sprintf("COGS for %s", item.Product.Name),
				DebitAmount:  cogsAmount,
				CreditAmount: decimal.Zero,
			})

			// Inventory (Credit)
			journalLines = append(journalLines, JournalLineRequest{
				AccountID:    getInventoryAccountID(), // You need to implement this
				Description:  fmt.Sprintf("Inventory decrease for %s", item.Product.Name),
				DebitAmount:  decimal.Zero,
				CreditAmount: cogsAmount,
			})
		}
	}

	journalReq.Lines = journalLines

	// Create journal entry
	journalResponse, err := h.unifiedJournalService.CreateJournalEntry(journalReq)
	if err != nil {
		log.Printf("Failed to create journal entry for sale %d: %v", saleID, err)
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	log.Printf("Created journal entry %d for sale %d", journalResponse.ID, saleID)

	// Trigger real-time update
	affectedAccounts := []uint64{getCashAccountID(), getAccountsReceivableAccountID(), getSalesRevenueAccountID()}
	h.ssotReportIntegrationSvc.OnTransactionCreated("SALE", saleID, affectedAccounts)

	return nil
}

// OnPurchaseCreated creates journal entries when a purchase is created/confirmed
func (h *SSOTTransactionHooks) OnPurchaseCreated(purchaseID uint64, userID uint64) error {
	var purchase models.Purchase
	if err := h.db.Preload("PurchaseItems.Product").Preload("Vendor").First(&purchase, purchaseID).Error; err != nil {
		return fmt.Errorf("failed to load purchase: %w", err)
	}

	// Only create journal entries for confirmed purchases
	if purchase.Status != "CONFIRMED" && purchase.Status != "RECEIVED" {
		return nil
	}

	// Create journal entry request
	journalReq := &JournalEntryRequest{
		SourceType:  models.SSOTSourceTypePurchase,
		SourceID:    &purchaseID,
		Reference:   purchase.Code,
	EntryDate:   purchase.Date,
		Description: fmt.Sprintf("Purchase Order #%s - %s", purchase.Code, purchase.Vendor.Name),
		CreatedBy:   userID,
		AutoPost:    true,
	}

	var journalLines []JournalLineRequest

	// Inventory/Expense Account (Debit)
	for _, item := range purchase.PurchaseItems {
		itemAmount := decimal.NewFromFloat(item.UnitPrice * float64(item.Quantity))
		
		if item.ProductID > 0 {
			// Inventory (Debit)
			journalLines = append(journalLines, JournalLineRequest{
				AccountID:    getInventoryAccountID(),
				Description:  fmt.Sprintf("Inventory purchase - %s", item.Product.Name),
				DebitAmount:  itemAmount,
				CreditAmount: decimal.Zero,
			})
		} else {
			// Expense Account (Debit) - for non-inventory items
			journalLines = append(journalLines, JournalLineRequest{
				AccountID:    getPurchaseExpenseAccountID(), // You need to implement this
				Description:  "Purchase expense",
				DebitAmount:  itemAmount,
				CreditAmount: decimal.Zero,
			})
		}
	}

	// Tax Input/Prepaid Tax (Debit) - if there's tax
	if purchase.TaxAmount > 0 {
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    getTaxInputAccountID(), // You need to implement this
			Description:  "Tax input",
			DebitAmount:  decimal.NewFromFloat(purchase.TaxAmount),
			CreditAmount: decimal.Zero,
		})
	}

	// Accounts Payable (Credit) - or Cash if cash purchase
	if purchase.PaymentMethod == "CASH" {
		// Cash Account (Credit)
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    getCashAccountID(),
			Description:  "Cash paid for purchase",
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromFloat(purchase.TotalAmount),
		})
	} else {
		// Accounts Payable (Credit)
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    getAccountsPayableAccountID(), // You need to implement this
			Description:  "Accounts payable for purchase",
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromFloat(purchase.TotalAmount),
		})
	}

	journalReq.Lines = journalLines

	// Create journal entry
	journalResponse, err := h.unifiedJournalService.CreateJournalEntry(journalReq)
	if err != nil {
		log.Printf("Failed to create journal entry for purchase %d: %v", purchaseID, err)
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	log.Printf("Created journal entry %d for purchase %d", journalResponse.ID, purchaseID)

	// Trigger real-time update
	affectedAccounts := []uint64{getInventoryAccountID(), getCashAccountID(), getAccountsPayableAccountID()}
	h.ssotReportIntegrationSvc.OnTransactionCreated("PURCHASE", purchaseID, affectedAccounts)

	return nil
}

// OnPaymentCreated creates journal entries when a payment is made/received
func (h *SSOTTransactionHooks) OnPaymentCreated(paymentID uint64, userID uint64) error {
	var payment models.Payment
	if err := h.db.Preload("Sale").Preload("Purchase").First(&payment, paymentID).Error; err != nil {
		return fmt.Errorf("failed to load payment: %w", err)
	}

	// Create journal entry request
	journalReq := &JournalEntryRequest{
		SourceType:  models.SSOTSourceTypePayment,
		SourceID:    &paymentID,
		Reference:   payment.Reference,
		EntryDate:   payment.Date,
		Description: fmt.Sprintf("Payment #%s", payment.Reference),
		CreatedBy:   userID,
		AutoPost:    true,
	}

	var journalLines []JournalLineRequest

	if payment.Method == "RECEIVED" { // Assuming Method field indicates direction
		// Payment received from customer
		// Cash/Bank Account (Debit)
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    getCashAccountID(), // or bank account based on payment method
			Description:  "Payment received",
			DebitAmount:  decimal.NewFromFloat(payment.Amount),
			CreditAmount: decimal.Zero,
		})

		// Accounts Receivable (Credit)
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    getAccountsReceivableAccountID(),
			Description:  "Accounts receivable payment",
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromFloat(payment.Amount),
		})
	} else if payment.Method == "PAID" {
		// Payment made to supplier
		// Accounts Payable (Debit)
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    getAccountsPayableAccountID(),
			Description:  "Accounts payable payment",
			DebitAmount:  decimal.NewFromFloat(payment.Amount),
			CreditAmount: decimal.Zero,
		})

		// Cash/Bank Account (Credit)
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    getCashAccountID(), // or bank account based on payment method
			Description:  "Payment made",
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromFloat(payment.Amount),
		})
	}

	journalReq.Lines = journalLines

	// Create journal entry
	journalResponse, err := h.unifiedJournalService.CreateJournalEntry(journalReq)
	if err != nil {
		log.Printf("Failed to create journal entry for payment %d: %v", paymentID, err)
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	log.Printf("Created journal entry %d for payment %d", journalResponse.ID, paymentID)

	// Trigger real-time update
	affectedAccounts := []uint64{getCashAccountID(), getAccountsReceivableAccountID(), getAccountsPayableAccountID()}
	h.ssotReportIntegrationSvc.OnTransactionCreated("PAYMENT", paymentID, affectedAccounts)

	return nil
}

// OnCashBankTransactionCreated creates journal entries for cash/bank transactions
func (h *SSOTTransactionHooks) OnCashBankTransactionCreated(transactionID uint64, userID uint64) error {
	var transaction models.CashBankTransaction
	if err := h.db.Preload("Account").First(&transaction, transactionID).Error; err != nil {
		return fmt.Errorf("failed to load cash bank transaction: %w", err)
	}

	// Create journal entry request
	journalReq := &JournalEntryRequest{
		SourceType:  models.SSOTSourceTypeCashBank,
		SourceID:    &transactionID,
		Reference:   transaction.ReferenceType,
		EntryDate:   transaction.TransactionDate,
		Description: transaction.Notes,
		CreatedBy:   userID,
		AutoPost:    true,
	}

	var journalLines []JournalLineRequest

	if transaction.Amount > 0 { // Positive amount = Receipt
		// Cash/Bank Receipt
		// Cash/Bank Account (Debit)
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    uint64(transaction.CashBankID), // Use CashBankID as account reference
			Description:  "Cash/Bank receipt",
			DebitAmount:  decimal.NewFromFloat(transaction.Amount),
			CreditAmount: decimal.Zero,
		})

		// Contra Account (Credit) - based on transaction category
		contraAccountID := getContraAccountForCashReceipt(transaction.ReferenceType)
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    contraAccountID,
			Description:  "Cash receipt contra account",
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromFloat(transaction.Amount),
		})
	} else { // Negative amount or zero = Payment
		// Cash/Bank Payment
		// Expense/Asset Account (Debit) - based on transaction category
		expenseAccountID := getExpenseAccountForCashPayment(transaction.ReferenceType)
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    expenseAccountID,
			Description:  "Cash payment expense",
			DebitAmount:  decimal.NewFromFloat(transaction.Amount),
			CreditAmount: decimal.Zero,
		})

		// Cash/Bank Account (Credit)
		journalLines = append(journalLines, JournalLineRequest{
			AccountID:    uint64(transaction.CashBankID), // Use CashBankID as account reference
			Description:  "Cash/Bank payment",
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromFloat(transaction.Amount),
		})
	}

	journalReq.Lines = journalLines

	// Create journal entry
	journalResponse, err := h.unifiedJournalService.CreateJournalEntry(journalReq)
	if err != nil {
		log.Printf("Failed to create journal entry for cash bank transaction %d: %v", transactionID, err)
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	log.Printf("Created journal entry %d for cash bank transaction %d", journalResponse.ID, transactionID)

	// Trigger real-time update
	affectedAccounts := []uint64{uint64(transaction.CashBankID)}
	h.ssotReportIntegrationSvc.OnTransactionCreated("CASH_BANK", transactionID, affectedAccounts)

	return nil
}

// OnJournalPosted triggers real-time updates when a journal is posted
func (h *SSOTTransactionHooks) OnJournalPosted(journalID uint64) {
	h.ssotReportIntegrationSvc.OnJournalPosted(journalID)
}

// Helper functions to get account IDs - these should be implemented based on your chart of accounts
// You may want to cache these or get them from configuration

func getCashAccountID() uint64 {
	// TODO: Implement - get cash account ID from chart of accounts
	return 1 // placeholder
}

func getAccountsReceivableAccountID() uint64 {
	// TODO: Implement - get accounts receivable account ID
	return 2 // placeholder
}

func getSalesRevenueAccountID() uint64 {
	// TODO: Implement - get sales revenue account ID
	return 3 // placeholder
}

func getTaxPayableAccountID() uint64 {
	// TODO: Implement - get tax payable account ID
	return 4 // placeholder
}

func getCOGSAccountID() uint64 {
	// TODO: Implement - get COGS account ID
	return 5 // placeholder
}

func getInventoryAccountID() uint64 {
	// TODO: Implement - get inventory account ID
	return 6 // placeholder
}

func getAccountsPayableAccountID() uint64 {
	// TODO: Implement - get accounts payable account ID
	return 7 // placeholder
}

func getTaxInputAccountID() uint64 {
	// TODO: Implement - get tax input account ID
	return 8 // placeholder
}

func getPurchaseExpenseAccountID() uint64 {
	// TODO: Implement - get purchase expense account ID
	return 9 // placeholder
}

func getContraAccountForCashReceipt(category string) uint64 {
	// TODO: Implement - get contra account for cash receipts based on category
	switch category {
	case "SALES":
		return getSalesRevenueAccountID()
	case "OTHER_INCOME":
		return 10 // Other income account
	default:
		return 11 // Miscellaneous income account
	}
}

func getExpenseAccountForCashPayment(category string) uint64 {
	// TODO: Implement - get expense account for cash payments based on category
	switch category {
	case "OFFICE_SUPPLIES":
		return 12 // Office supplies expense account
	case "UTILITIES":
		return 13 // Utilities expense account
	default:
		return 14 // General expense account
	}
}

// All account mapping functions are defined above - no duplicates
