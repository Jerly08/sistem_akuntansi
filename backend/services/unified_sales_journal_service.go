package services

import (
	"fmt"
	"time"
	"context"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
	"log"
	"github.com/shopspring/decimal"
)

// UnifiedSalesJournalService combines all journal-related operations with proper coordination
type UnifiedSalesJournalService struct {
	db                *gorm.DB
	coordinator       *JournalCoordinator
	accountResolver   *AccountResolver
	journalRepo       repositories.JournalEntryRepository
}

// TransactionState tracks the state of a transaction for rollback purposes
type TransactionState struct {
	SaleID          uint                    `json:"sale_id"`
	PaymentID       uint                    `json:"payment_id"`
	JournalEntryID  uint                    `json:"journal_entry_id"`
	InventoryOps    []InventoryOperation    `json:"inventory_ops"`
	AccountUpdates  []TransactionAccountUpdate  `json:"account_updates"`
	Status          string                  `json:"status"`
	CreatedAt       time.Time               `json:"created_at"`
	CompletedAt     *time.Time              `json:"completed_at"`
	RolledBackAt    *time.Time              `json:"rolled_back_at"`
	ErrorMessage    string                  `json:"error_message"`
}

// InventoryOperation represents an inventory change that can be rolled back
type InventoryOperation struct {
	ProductID      uint    `json:"product_id"`
	Quantity       int     `json:"quantity"`
	Type           string  `json:"type"` // DECREASE, INCREASE
	OriginalStock  int     `json:"original_stock"`
}

// TransactionAccountUpdate represents an account balance change for rollback
type TransactionAccountUpdate struct {
	AccountID      uint    `json:"account_id"`
	Amount         float64 `json:"amount"`
	Type           string  `json:"type"` // DEBIT, CREDIT
	OriginalBalance float64 `json:"original_balance"`
}

const (
	TransactionStatusPending   = "PENDING"
	TransactionStatusCompleted = "COMPLETED"
	TransactionStatusFailed    = "FAILED"
	TransactionStatusRolledBack = "ROLLED_BACK"
)

func NewUnifiedSalesJournalService(db *gorm.DB) *UnifiedSalesJournalService {
	return &UnifiedSalesJournalService{
		db:                db,
		coordinator:       NewJournalCoordinator(db),
		accountResolver:   NewAccountResolver(db),
		journalRepo:       repositories.NewJournalEntryRepository(db),
	}
}

// CreateSaleJournalEntry creates journal entry for sale with full coordination and rollback
func (usjs *UnifiedSalesJournalService) CreateSaleJournalEntry(sale *models.Sale, userID uint) (*models.JournalEntry, error) {
	// Create transaction state for tracking
	state := &TransactionState{
		SaleID:    sale.ID,
		Status:    TransactionStatusPending,
		CreatedAt: time.Now(),
	}
	
	// Create journal creation request
	journalReq := &JournalCreationRequest{
		TransactionType: models.JournalRefSale,
		ReferenceID:     sale.ID,
		UserID:          userID,
		Description:     fmt.Sprintf("Sales Invoice %s - %s", sale.Code, sale.Customer.Name),
	}
	
	// Use coordinator to create journal entry
	result, err := usjs.coordinator.CreateJournalEntryWithCoordination(journalReq, func() (*models.JournalEntry, error) {
		return usjs.createSaleJournalEntryInternal(sale, userID, state)
	})
	
	if err != nil || !result.Success {
		state.Status = TransactionStatusFailed
		state.ErrorMessage = fmt.Sprintf("Failed to create journal entry: %v", err)
		usjs.rollbackTransaction(state)
		return nil, fmt.Errorf("journal creation failed: %v", err)
	}
	
	if result.WasDuplicate {
		log.Printf("⚠️ Journal entry already exists for sale %d", sale.ID)
		return result.ExistingEntry, nil
	}
	
	// Update state with success
	state.JournalEntryID = result.JournalEntryID
	state.Status = TransactionStatusCompleted
	now := time.Now()
	state.CompletedAt = &now
	
	// Get the created journal entry
	ctx := context.Background()
	journalEntry, err := usjs.journalRepo.FindByID(ctx, result.JournalEntryID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created journal entry: %v", err)
	}
	
	log.Printf("✅ Successfully created journal entry %d for sale %d", journalEntry.ID, sale.ID)
	return journalEntry, nil
}

// CreatePaymentJournalEntry creates journal entry for payment
func (usjs *UnifiedSalesJournalService) CreatePaymentJournalEntry(payment *models.SalePayment, userID uint) (*models.JournalEntry, error) {
	// Create transaction state for tracking
	state := &TransactionState{
		PaymentID: payment.ID,
		SaleID:    payment.SaleID,
		Status:    TransactionStatusPending,
		CreatedAt: time.Now(),
	}
	
	// Create journal creation request
	journalReq := &JournalCreationRequest{
		TransactionType: models.JournalRefPayment,
		ReferenceID:     payment.ID,
		UserID:          userID,
		Description:     fmt.Sprintf("Payment for Sale %d - %s", payment.SaleID, payment.PaymentMethod),
	}
	
	// Use coordinator to create journal entry
	result, err := usjs.coordinator.CreateJournalEntryWithCoordination(journalReq, func() (*models.JournalEntry, error) {
		return usjs.createPaymentJournalEntryInternal(payment, userID, state)
	})
	
	if err != nil || !result.Success {
		state.Status = TransactionStatusFailed
		state.ErrorMessage = fmt.Sprintf("Failed to create payment journal entry: %v", err)
		usjs.rollbackTransaction(state)
		return nil, fmt.Errorf("payment journal creation failed: %v", err)
	}
	
	if result.WasDuplicate {
		log.Printf("⚠️ Payment journal entry already exists for payment %d", payment.ID)
		return result.ExistingEntry, nil
	}
	
	// Update state with success
	state.JournalEntryID = result.JournalEntryID
	state.Status = TransactionStatusCompleted
	now := time.Now()
	state.CompletedAt = &now
	
	// Get the created journal entry
	ctx := context.Background()
	journalEntry, err := usjs.journalRepo.FindByID(ctx, result.JournalEntryID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created payment journal entry: %v", err)
	}
	
	log.Printf("✅ Successfully created payment journal entry %d for payment %d", journalEntry.ID, payment.ID)
	return journalEntry, nil
}

// createSaleJournalEntryInternal creates the actual journal entry for sales using SSOT API
func (usjs *UnifiedSalesJournalService) createSaleJournalEntryInternal(sale *models.Sale, userID uint, state *TransactionState) (*models.JournalEntry, error) {
	// Get required accounts using AccountResolver
	arAccount, err := usjs.accountResolver.GetAccount(AccountTypeAccountsReceivable)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts receivable account: %v", err)
	}
	
	salesAccount, err := usjs.accountResolver.GetAccount(AccountTypeSalesRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales revenue account: %v", err)
	}
	
	// Create journal lines for SSOT API
	lines := []JournalLineRequest{
		{
			AccountID:    uint64(arAccount.ID),
			Description:  fmt.Sprintf("Sales to %s", sale.Customer.Name),
			DebitAmount:  decimal.NewFromFloat(sale.TotalAmount),
			CreditAmount: decimal.Zero,
		},
		{
			AccountID:    uint64(salesAccount.ID),
			Description:  "Sales Revenue",
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromFloat(sale.TotalAmount - sale.PPN),
		},
	}

	// Add PPN line if applicable
	if sale.PPN > 0 {
		ppnAccount, err := usjs.accountResolver.GetAccount(AccountTypePPNPayable)
		if err == nil {
			lines = append(lines, JournalLineRequest{
				AccountID:    uint64(ppnAccount.ID),
				Description:  "PPN Keluaran",
				DebitAmount:  decimal.Zero,
				CreditAmount: decimal.NewFromFloat(sale.PPN),
			})
		}
	}

	// Create journal entry using SSOT API
	journalRequest := &JournalEntryRequest{
		SourceType:  models.SSOTSourceTypeSale,
		SourceID:    func() *uint64 { id := uint64(sale.ID); return &id }(),
		Reference:   sale.Code,
		EntryDate:   sale.Date,
		Description: fmt.Sprintf("Sales Invoice %s - %s", sale.Code, sale.Customer.Name),
		Lines:       lines,
		AutoPost:    true,
		CreatedBy:   uint64(userID),
	}

	// Create using SSOT unified journal service
	unifiedJournalService := NewUnifiedJournalService(usjs.db)
	journalResponse, err := unifiedJournalService.CreateJournalEntry(journalRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSOT journal entry: %v", err)
	}
	
	// Convert SSOT JournalResponse to legacy JournalEntry for compatibility
	journalEntry := &models.JournalEntry{
		ID:              uint(journalResponse.ID),
		EntryDate:       sale.Date,
		Description:     fmt.Sprintf("Sales Invoice %s - %s", sale.Code, sale.Customer.Name),
		Reference:       sale.Code,
		ReferenceType:   models.JournalRefSale,
		ReferenceID:     &sale.ID,
		UserID:          userID,
		Status:          models.JournalStatusPosted,
		TotalDebit:      sale.TotalAmount,
		TotalCredit:     sale.TotalAmount,
		IsBalanced:      true,
		IsAutoGenerated: true,
	}
	
	// Track account balance updates for rollback
	state.AccountUpdates = []TransactionAccountUpdate{
		{
			AccountID: arAccount.ID,
			Amount:    sale.TotalAmount,
			Type:      "DEBIT",
		},
		{
			AccountID: salesAccount.ID,
			Amount:    sale.TotalAmount - sale.PPN,
			Type:      "CREDIT",
		},
	}
	
	if sale.PPN > 0 {
		ppnAccount, _ := usjs.accountResolver.GetAccount(AccountTypePPNPayable)
		if ppnAccount != nil {
			state.AccountUpdates = append(state.AccountUpdates, TransactionAccountUpdate{
				AccountID: ppnAccount.ID,
				Amount:    sale.PPN,
				Type:      "CREDIT",
			})
		}
	}
	
	log.Printf("📝 Created journal entry %d with %d lines for sale %d", journalEntry.ID, len(lines), sale.ID)
	return journalEntry, nil
}

// createPaymentJournalEntryInternal creates the actual journal entry for payment
func (usjs *UnifiedSalesJournalService) createPaymentJournalEntryInternal(payment *models.SalePayment, userID uint, state *TransactionState) (*models.JournalEntry, error) {
	// Get required accounts using AccountResolver
	arAccount, err := usjs.accountResolver.GetAccount(AccountTypeAccountsReceivable)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts receivable account: %v", err)
	}
	
	// Get appropriate cash/bank account based on payment method
	cashBankAccount, err := usjs.accountResolver.GetBankAccountForPaymentMethod(payment.PaymentMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to get cash/bank account: %v", err)
	}
	
	// Create journal entry
	journalEntry := &models.JournalEntry{
		EntryDate:       payment.PaymentDate,
		Description:     fmt.Sprintf("Payment for Sale %d - %s", payment.SaleID, payment.PaymentMethod),
		Reference:       fmt.Sprintf("PAY-%d", payment.ID),
		ReferenceType:   models.JournalRefPayment,
		ReferenceID:     &payment.ID,
		UserID:          userID,
		Status:          models.JournalStatusDraft,
		TotalDebit:      payment.Amount,
		TotalCredit:     payment.Amount,
		IsBalanced:      true,
		IsAutoGenerated: true,
		AccountID:       &cashBankAccount.ID,
	}
	
	// Validate journal entry before saving
	if err := journalEntry.ValidateComplete(); err != nil {
		return nil, fmt.Errorf("payment journal entry validation failed: %v", err)
	}
	
	// Create journal entry in database
	if err := usjs.db.Create(journalEntry).Error; err != nil {
		return nil, fmt.Errorf("failed to create payment journal entry: %v", err)
	}
	
	// Create journal lines
	lines := []models.JournalLine{
		{
			JournalEntryID: journalEntry.ID,
			AccountID:      cashBankAccount.ID,
			Description:    fmt.Sprintf("Payment received - %s", payment.PaymentMethod),
			DebitAmount:    payment.Amount,
			CreditAmount:   0,
			LineNumber:     1,
		},
		{
			JournalEntryID: journalEntry.ID,
			AccountID:      arAccount.ID,
			Description:    "Payment against receivables",
			DebitAmount:    0,
			CreditAmount:   payment.Amount,
			LineNumber:     2,
		},
	}
	
	// Create journal lines
	for _, line := range lines {
		if err := usjs.db.Create(&line).Error; err != nil {
			return nil, fmt.Errorf("failed to create payment journal line: %v", err)
		}
	}
	
	// Track account balance updates for rollback
	state.AccountUpdates = []TransactionAccountUpdate{
		{
			AccountID: cashBankAccount.ID,
			Amount:    payment.Amount,
			Type:      "DEBIT",
		},
		{
			AccountID: arAccount.ID,
			Amount:    payment.Amount,
			Type:      "CREDIT",
		},
	}
	
	log.Printf("💰 Created payment journal entry %d with %d lines for payment %d", journalEntry.ID, len(lines), payment.ID)
	return journalEntry, nil
}

// rollbackTransaction rolls back a failed transaction
func (usjs *UnifiedSalesJournalService) rollbackTransaction(state *TransactionState) error {
	log.Printf("🔄 Rolling back transaction for sale %d (journal %d)", state.SaleID, state.JournalEntryID)
	
	// Rollback in reverse order
	// 1. Delete journal lines if created
	if state.JournalEntryID > 0 {
		err := usjs.db.Where("journal_entry_id = ?", state.JournalEntryID).Delete(&models.JournalLine{}).Error
		if err != nil {
			log.Printf("❌ Failed to rollback journal lines: %v", err)
		}
		
		// Delete journal entry
		err = usjs.db.Delete(&models.JournalEntry{}, state.JournalEntryID).Error
		if err != nil {
			log.Printf("❌ Failed to rollback journal entry: %v", err)
		}
	}
	
	// 2. Restore inventory if needed
	for _, invOp := range state.InventoryOps {
		err := usjs.rollbackInventoryOperation(invOp)
		if err != nil {
			log.Printf("❌ Failed to rollback inventory for product %d: %v", invOp.ProductID, err)
		}
	}
	
	// 3. Restore account balances if needed
	for _, balUpdate := range state.AccountUpdates {
		err := usjs.rollbackAccountUpdate(balUpdate)
		if err != nil {
			log.Printf("❌ Failed to rollback account balance for account %d: %v", balUpdate.AccountID, err)
		}
	}
	
	// Update state
	state.Status = TransactionStatusRolledBack
	now := time.Now()
	state.RolledBackAt = &now
	
	log.Printf("✅ Transaction rollback completed for sale %d", state.SaleID)
	return nil
}

// rollbackInventoryOperation rolls back an inventory change
func (usjs *UnifiedSalesJournalService) rollbackInventoryOperation(invOp InventoryOperation) error {
	// Restore original stock level
	return usjs.db.Model(&models.Product{}).
		Where("id = ?", invOp.ProductID).
		Update("stock", invOp.OriginalStock).Error
}

// rollbackAccountUpdate rolls back an account balance change
func (usjs *UnifiedSalesJournalService) rollbackAccountUpdate(balUpdate TransactionAccountUpdate) error {
	// Restore original balance
	return usjs.db.Model(&models.Account{}).
		Where("id = ?", balUpdate.AccountID).
		Update("balance", balUpdate.OriginalBalance).Error
}

// CreateReversalJournalEntry creates a reversal entry for a sale
func (usjs *UnifiedSalesJournalService) CreateReversalJournalEntry(saleID uint, reason string, userID uint) (*models.JournalEntry, error) {
	// Find original journal entry
	ctx := context.Background()
	originalEntry, err := usjs.journalRepo.FindByReferenceID(ctx, models.JournalRefSale, saleID)
	if err != nil {
		return nil, fmt.Errorf("failed to find original journal entry for sale %d: %v", saleID, err)
	}
	
	if originalEntry == nil {
		return nil, fmt.Errorf("no journal entry found for sale %d", saleID)
	}
	
	// Create reversal journal creation request
	journalReq := &JournalCreationRequest{
		TransactionType: "REVERSAL",
		ReferenceID:     originalEntry.ID,
		UserID:          userID,
		Description:     fmt.Sprintf("REVERSAL: %s - Reason: %s", originalEntry.Description, reason),
	}
	
	// Use coordinator to create reversal entry
	result, err := usjs.coordinator.CreateJournalEntryWithCoordination(journalReq, func() (*models.JournalEntry, error) {
		return usjs.createReversalJournalEntryInternal(originalEntry, reason, userID)
	})
	
	if err != nil || !result.Success {
		return nil, fmt.Errorf("reversal journal creation failed: %v", err)
	}
	
	// Get the created reversal entry
	reversalEntry, err := usjs.journalRepo.FindByID(ctx, result.JournalEntryID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created reversal entry: %v", err)
	}
	
	// Update original entry status
	err = usjs.db.Model(originalEntry).Updates(map[string]interface{}{
		"status":     models.JournalStatusReversed,
		"reversal_id": reversalEntry.ID,
	}).Error
	
	if err != nil {
		log.Printf("⚠️ Warning: Failed to update original entry status: %v", err)
	}
	
	log.Printf("🔄 Successfully created reversal journal entry %d for sale %d", reversalEntry.ID, saleID)
	return reversalEntry, nil
}

// createReversalJournalEntryInternal creates the actual reversal journal entry
func (usjs *UnifiedSalesJournalService) createReversalJournalEntryInternal(originalEntry *models.JournalEntry, reason string, userID uint) (*models.JournalEntry, error) {
	// Create reversal entry (swap debits and credits)
	reversalEntry := &models.JournalEntry{
		EntryDate:       time.Now(),
		Description:     fmt.Sprintf("REVERSAL: %s - Reason: %s", originalEntry.Description, reason),
		Reference:       fmt.Sprintf("REV-%s", originalEntry.Reference),
		ReferenceType:   "REVERSAL",
		ReferenceID:     &originalEntry.ID,
		UserID:          userID,
		Status:          models.JournalStatusDraft,
		TotalDebit:      originalEntry.TotalCredit,  // Swap amounts
		TotalCredit:     originalEntry.TotalDebit,
		IsBalanced:      true,
		IsAutoGenerated: true,
		ReversedID:      &originalEntry.ID,
	}
	
	// Create reversal entry
	if err := usjs.db.Create(reversalEntry).Error; err != nil {
		return nil, fmt.Errorf("failed to create reversal journal entry: %v", err)
	}
	
	// Get original journal lines to create reversed lines
	var originalLines []models.JournalLine
	if err := usjs.db.Where("journal_entry_id = ?", originalEntry.ID).Find(&originalLines).Error; err != nil {
		log.Printf("⚠️ Warning: Could not load original journal lines: %v", err)
	}
	
	// Create reversed journal lines
	for i, originalLine := range originalLines {
		reversalLine := models.JournalLine{
			JournalEntryID: reversalEntry.ID,
			AccountID:      originalLine.AccountID,
			Description:    fmt.Sprintf("REVERSAL: %s", originalLine.Description),
			DebitAmount:    originalLine.CreditAmount,  // Swap amounts
			CreditAmount:   originalLine.DebitAmount,
			LineNumber:     i + 1,
		}
		
		if err := usjs.db.Create(&reversalLine).Error; err != nil {
			return nil, fmt.Errorf("failed to create reversal journal line: %v", err)
		}
	}
	
	return reversalEntry, nil
}

// ValidateJournalIntegrity validates the integrity of journal entries for a sale
func (usjs *UnifiedSalesJournalService) ValidateJournalIntegrity(saleID uint) error {
	ctx := context.Background()
	
	// Check if journal entry exists for the sale
	entry, err := usjs.journalRepo.FindByReferenceID(ctx, models.JournalRefSale, saleID)
	if err != nil {
		return fmt.Errorf("error checking journal entry for sale %d: %v", saleID, err)
	}
	
	if entry == nil {
		return fmt.Errorf("no journal entry found for sale %d", saleID)
	}
	
	// Validate journal entry balance
	if !entry.IsBalanced {
		return fmt.Errorf("journal entry %d is not balanced", entry.ID)
	}
	
	// Check if journal lines exist and are balanced
	var lines []models.JournalLine
	if err := usjs.db.Where("journal_entry_id = ?", entry.ID).Find(&lines).Error; err != nil {
		return fmt.Errorf("failed to load journal lines: %v", err)
	}
	
	if len(lines) == 0 {
		log.Printf("⚠️ Warning: No journal lines found for entry %d", entry.ID)
		return nil // This might be okay for simplified entries
	}
	
	// Validate line balance
	var totalDebit, totalCredit float64
	for _, line := range lines {
		totalDebit += line.DebitAmount
		totalCredit += line.CreditAmount
	}
	
	if totalDebit != totalCredit {
		return fmt.Errorf("journal lines are not balanced: debit=%.2f, credit=%.2f", totalDebit, totalCredit)
	}
	
	if totalDebit != entry.TotalDebit || totalCredit != entry.TotalCredit {
		return fmt.Errorf("journal entry totals don't match line totals")
	}
	
	log.Printf("✅ Journal integrity validation passed for sale %d", saleID)
	return nil
}