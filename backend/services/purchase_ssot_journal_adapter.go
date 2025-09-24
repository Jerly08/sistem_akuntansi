package services

import (
	"context"
	"fmt"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// PurchaseSSOTJournalAdapter provides integration between Purchase System and SSOT Journal System
type PurchaseSSOTJournalAdapter struct {
	db                    *gorm.DB
	unifiedJournalService *UnifiedJournalService
	accountRepo           repositories.AccountRepository
}

// NewPurchaseSSOTJournalAdapter creates a new adapter instance
func NewPurchaseSSOTJournalAdapter(
	db *gorm.DB,
	unifiedJournalService *UnifiedJournalService,
	accountRepo repositories.AccountRepository,
) *PurchaseSSOTJournalAdapter {
	return &PurchaseSSOTJournalAdapter{
		db:                    db,
		unifiedJournalService: unifiedJournalService,
		accountRepo:           accountRepo,
	}
}

// CreatePurchaseJournalEntry creates SSOT journal entry for purchase transaction
func (adapter *PurchaseSSOTJournalAdapter) CreatePurchaseJournalEntry(
	ctx context.Context,
	purchase *models.Purchase,
	userID uint64,
) (*models.SSOTJournalEntry, error) {
	
	fmt.Printf("🏗️ Creating SSOT journal entry for purchase %s (ID: %d)\n", purchase.Code, purchase.ID)

	// Get account IDs
	accountIDs, err := adapter.getPurchaseAccountIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase account IDs: %v", err)
	}

	// Create journal entry request
	journalReq := &JournalEntryRequest{
		SourceType:      models.SSOTSourceTypePurchase,
		SourceID:        uint64Ptr(uint64(purchase.ID)),
		Reference:       purchase.Code,
		EntryDate:       purchase.Date,
		Description:     fmt.Sprintf("Purchase Order %s - %s", purchase.Code, purchase.Vendor.Name),
		Lines:           adapter.buildPurchaseJournalLines(purchase, accountIDs),
		AutoPost:        true,
		CreatedBy:       userID,
	}

	// Validate and create journal entry
	journalResponse, err := adapter.unifiedJournalService.CreateJournalEntry(journalReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSOT journal entry: %v", err)
	}

	fmt.Printf("✅ SSOT journal entry created: %s (ID: %d) for purchase %s\n", 
		journalResponse.EntryNumber, journalResponse.ID, purchase.Code)

	// Convert JournalResponse to SSOTJournalEntry for return
	journalEntry := &models.SSOTJournalEntry{
		ID:          journalResponse.ID,
		EntryNumber: journalResponse.EntryNumber,
		Status:      journalResponse.Status,
		TotalDebit:  journalResponse.TotalDebit,
		TotalCredit: journalResponse.TotalCredit,
		IsBalanced:  journalResponse.IsBalanced,
		CreatedAt:   journalResponse.CreatedAt,
		UpdatedAt:   journalResponse.UpdatedAt,
	}

	return journalEntry, nil
}

// GetPurchaseJournalEntries retrieves all journal entries for a purchase
func (adapter *PurchaseSSOTJournalAdapter) GetPurchaseJournalEntries(
	ctx context.Context,
	purchaseID uint64,
) ([]models.SSOTJournalEntry, error) {
	
	filters := JournalFilters{
		SourceType: models.SSOTSourceTypePurchase,
		SourceID:   &purchaseID, // Filter by specific purchase ID
		Page:       1,
		Limit:      100,
	}

	response, err := adapter.unifiedJournalService.GetJournalEntries(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase journal entries: %v", err)
	}

	// Convert JournalResponse to SSOTJournalEntry
	var purchaseEntries []models.SSOTJournalEntry
	for _, journalResp := range response.Data {
		entry := models.SSOTJournalEntry{
			ID:          journalResp.ID,
			EntryNumber: journalResp.EntryNumber,
			Status:      journalResp.Status,
			TotalDebit:  journalResp.TotalDebit,
			TotalCredit: journalResp.TotalCredit,
			IsBalanced:  journalResp.IsBalanced,
			CreatedAt:   journalResp.CreatedAt,
			UpdatedAt:   journalResp.UpdatedAt,
		}
		
		purchaseEntries = append(purchaseEntries, entry)
	}

	return purchaseEntries, nil
}

// CreatePurchasePaymentJournalEntry creates journal entry for purchase payment
func (adapter *PurchaseSSOTJournalAdapter) CreatePurchasePaymentJournalEntry(
	ctx context.Context,
	purchase *models.Purchase,
	paymentAmount decimal.Decimal,
	bankAccountID uint64,
	userID uint64,
	reference string,
	notes string,
) (*models.SSOTJournalEntry, error) {

	fmt.Printf("💰 Creating SSOT journal entry for purchase payment %s (Amount: %s)\n", 
		purchase.Code, paymentAmount.String())

	// Get account IDs
	accountIDs, err := adapter.getPurchaseAccountIDs()
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase account IDs: %v", err)
	}

	// Get bank account ID from CashBank
	var actualBankAccountID uint64
	if bankAccountID > 0 {
		// Find the account_id from cash_banks table
		var cashBank models.CashBank
		err := adapter.db.Select("account_id").Where("id = ?", bankAccountID).First(&cashBank).Error
		if err != nil {
			return nil, fmt.Errorf("cash/bank account not found: %v", err)
		}
		actualBankAccountID = uint64(cashBank.AccountID)
	} else {
		// Use default Kas account (1101)
		var kasAccount models.Account
		err := adapter.db.Select("id").Where("code = ?", "1101").First(&kasAccount).Error
		if err != nil {
			return nil, fmt.Errorf("default Kas account (1101) not found: %v", err)
		}
		actualBankAccountID = uint64(kasAccount.ID)
	}

	// Build payment journal lines
	lines := []JournalLineRequest{
		{
			AccountID:    accountIDs.AccountsPayableID,
			Description:  fmt.Sprintf("Payment to %s - %s", purchase.Vendor.Name, reference),
			DebitAmount:  paymentAmount,
			CreditAmount: decimal.Zero,
		},
		{
			AccountID:    actualBankAccountID,
			Description:  fmt.Sprintf("Bank payment - %s", reference),
			DebitAmount:  decimal.Zero,
			CreditAmount: paymentAmount,
		},
	}

		// Create journal entry request
	journalReq := &JournalEntryRequest{
		SourceType:      models.SSOTSourceTypePayment,
		SourceID:        uint64Ptr(uint64(purchase.ID)),
		Reference:       reference,
		EntryDate:       time.Now(),
		Description:     fmt.Sprintf("Purchase Payment %s - %s (%s)", purchase.Code, purchase.Vendor.Name, paymentAmount.String()),
		Lines:           lines,
		AutoPost:        true,
		CreatedBy:       userID,
	}

		// Create journal entry
	journalResponse, err := adapter.unifiedJournalService.CreateJournalEntry(journalReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment journal entry: %v", err)
	}

	fmt.Printf("✅ Payment journal entry created: %s (ID: %d) for purchase %s\n", 
		journalResponse.EntryNumber, journalResponse.ID, purchase.Code)

	// Convert JournalResponse to SSOTJournalEntry for return
	journalEntry := &models.SSOTJournalEntry{
		ID:          journalResponse.ID,
		EntryNumber: journalResponse.EntryNumber,
		Status:      journalResponse.Status,
		TotalDebit:  journalResponse.TotalDebit,
		TotalCredit: journalResponse.TotalCredit,
		IsBalanced:  journalResponse.IsBalanced,
		CreatedAt:   journalResponse.CreatedAt,
		UpdatedAt:   journalResponse.UpdatedAt,
	}

	return journalEntry, nil
}

// buildPurchaseJournalLines builds journal lines for purchase transaction
func (adapter *PurchaseSSOTJournalAdapter) buildPurchaseJournalLines(
	purchase *models.Purchase,
	accountIDs *SSOTPurchaseAccountIDs,
) []JournalLineRequest {
	
	var lines []JournalLineRequest

	// DEBIT SIDE - Assets and Expenses
	
	// 1. Inventory/Expense accounts for each item
	for _, item := range purchase.PurchaseItems {
		accountID := accountIDs.InventoryAccountID // Default to inventory
		description := fmt.Sprintf("Purchase - %s", item.Product.Name)
		
		// Use specific expense account if provided
		if item.ExpenseAccountID != 0 {
			accountID = uint64(item.ExpenseAccountID)
		}
		
		itemTotalDecimal := decimal.NewFromFloat(item.TotalPrice)
		
		lines = append(lines, JournalLineRequest{
			AccountID:    accountID,
			Description:  description,
			DebitAmount:  itemTotalDecimal,
			CreditAmount: decimal.Zero,
		})
	}

	// 2. PPN Masukan (Input VAT) if applicable
	if purchase.PPNAmount > 0 {
		ppnAmountDecimal := decimal.NewFromFloat(purchase.PPNAmount)
		lines = append(lines, JournalLineRequest{
			AccountID:    accountIDs.PPNInputAccountID,
			Description:  "PPN Masukan (Input VAT)",
			DebitAmount:  ppnAmountDecimal,
			CreditAmount: decimal.Zero,
		})
	}

	// CREDIT SIDE - Based on Payment Method
	
	// Check payment method to determine credit account
	if purchase.PaymentMethod == models.PurchasePaymentCash ||
		purchase.PaymentMethod == models.PurchasePaymentTransfer ||
		purchase.PaymentMethod == models.PurchasePaymentCheck {
		// For immediate payment (CASH/BANK/TRANSFER): Credit to Bank Account
		// Get bank account ID from purchase
		var creditAccountID uint64
		if purchase.BankAccountID != nil {
			// Find the account_id from cash_banks table
			var cashBank models.CashBank
			err := adapter.db.Select("account_id").Where("id = ?", *purchase.BankAccountID).First(&cashBank).Error
			if err == nil && cashBank.AccountID != 0 {
				creditAccountID = uint64(cashBank.AccountID)
			} else {
				// Fallback to default Kas account (1101)
				var kasAccount models.Account
				err := adapter.db.Select("id").Where("code = ?", "1101").First(&kasAccount).Error
				if err == nil {
					creditAccountID = uint64(kasAccount.ID)
				} else {
					creditAccountID = accountIDs.AccountsPayableID // Final fallback
				}
			}
		} else {
			// No bank account specified, use default Kas account (1101)
			var kasAccount models.Account
			err := adapter.db.Select("id").Where("code = ?", "1101").First(&kasAccount).Error
			if err == nil {
				creditAccountID = uint64(kasAccount.ID)
			} else {
				creditAccountID = accountIDs.AccountsPayableID // Final fallback
			}
		}
		
		totalAmountDecimal := decimal.NewFromFloat(purchase.TotalAmount)
		lines = append(lines, JournalLineRequest{
			AccountID:    creditAccountID,
			Description:  fmt.Sprintf("%s Payment - %s", purchase.PaymentMethod, purchase.Vendor.Name),
			DebitAmount:  decimal.Zero,
			CreditAmount: totalAmountDecimal,
		})
	} else {
		// For credit purchases: Credit to Accounts Payable
		totalAmountDecimal := decimal.NewFromFloat(purchase.TotalAmount)
		lines = append(lines, JournalLineRequest{
			AccountID:    accountIDs.AccountsPayableID,
			Description:  fmt.Sprintf("Accounts Payable - %s", purchase.Vendor.Name),
			DebitAmount:  decimal.Zero,
			CreditAmount: totalAmountDecimal,
		})
	}

	// 4. Tax withholdings (PPh 21, PPh 23) if applicable
	if purchase.PPh21Amount > 0 && accountIDs.PPh21PayableID != 0 {
		pph21AmountDecimal := decimal.NewFromFloat(purchase.PPh21Amount)
		lines = append(lines, JournalLineRequest{
			AccountID:    accountIDs.PPh21PayableID,
			Description:  "PPh 21 Withholding",
			DebitAmount:  decimal.Zero,
			CreditAmount: pph21AmountDecimal,
		})
	}

	if purchase.PPh23Amount > 0 && accountIDs.PPh23PayableID != 0 {
		pph23AmountDecimal := decimal.NewFromFloat(purchase.PPh23Amount)
		lines = append(lines, JournalLineRequest{
			AccountID:    accountIDs.PPh23PayableID,
			Description:  "PPh 23 Withholding",
			DebitAmount:  decimal.Zero,
			CreditAmount: pph23AmountDecimal,
		})
	}

	return lines
}

// getPurchaseAccountIDs retrieves required account IDs for purchase journal entries
func (adapter *PurchaseSSOTJournalAdapter) getPurchaseAccountIDs() (*SSOTPurchaseAccountIDs, error) {
	accountIDs := &SSOTPurchaseAccountIDs{}
	
	// Get inventory account (1301)
	if inventoryAccount, err := adapter.accountRepo.FindByCode(nil, "1301"); err == nil {
		accountIDs.InventoryAccountID = uint64(inventoryAccount.ID)
		accountIDs.PrimaryAccountID = uint64(inventoryAccount.ID) // Default to inventory
	} else {
		return nil, fmt.Errorf("inventory account 1301 not found: %v", err)
	}
	
	// Get PPN Input account (2102)
	if ppnAccount, err := adapter.accountRepo.FindByCode(nil, "2102"); err == nil {
		accountIDs.PPNInputAccountID = uint64(ppnAccount.ID)
	} else {
		return nil, fmt.Errorf("PPN input account 2102 not found: %v", err)
	}
	
	// Get Accounts Payable (2101)
	if apAccount, err := adapter.accountRepo.FindByCode(nil, "2101"); err == nil {
		accountIDs.AccountsPayableID = uint64(apAccount.ID)
	} else {
		return nil, fmt.Errorf("accounts payable account 2101 not found: %v", err)
	}
	
	// Get PPh 21 Payable (2111) - optional
	if pph21Account, err := adapter.accountRepo.FindByCode(nil, "2111"); err == nil {
		accountIDs.PPh21PayableID = uint64(pph21Account.ID)
	}
	
	// Get PPh 23 Payable (2112) - optional
	if pph23Account, err := adapter.accountRepo.FindByCode(nil, "2112"); err == nil {
		accountIDs.PPh23PayableID = uint64(pph23Account.ID)
	}
	
	return accountIDs, nil
}

// Helper functions

func uint64Ptr(v uint64) *uint64 {
	return &v
}

func decimalZeroPtr() *decimal.Decimal {
	zero := decimal.Zero
	return &zero
}

// SSOTPurchaseAccountIDs holds the account IDs needed for SSOT purchase journal entries
type SSOTPurchaseAccountIDs struct {
	PrimaryAccountID      uint64 // Inventory or main expense account
	InventoryAccountID    uint64 // 1301 - Persediaan Barang Dagangan
	PPNInputAccountID     uint64 // 2102 - PPN Masukan
	AccountsPayableID     uint64 // 2101 - Utang Usaha
	PPh21PayableID        uint64 // 2111 - Utang PPh 21
	PPh23PayableID        uint64 // 2112 - Utang PPh 23
}
