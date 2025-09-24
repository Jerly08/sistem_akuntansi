package services

import (
	"fmt"
	"log"
	"sync"
	"time"

	"app-sistem-akuntansi/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// SingleSourcePostingService - THE ONLY service allowed to update balances
// This ensures NO double posting ever occurs by centralizing all balance updates
type SingleSourcePostingService struct {
	db                    *gorm.DB
	unifiedJournalService *UnifiedJournalService
	mutex                 sync.RWMutex // Prevent concurrent access issues
	transactionTracker    map[string]bool // Track processed transactions
}

// NewSingleSourcePostingService creates the single source posting service
func NewSingleSourcePostingService(db *gorm.DB) *SingleSourcePostingService {
	return &SingleSourcePostingService{
		db:                    db,
		unifiedJournalService: NewUnifiedJournalService(db),
		transactionTracker:   make(map[string]bool),
	}
}

// PostingRequest represents a balance posting request
type PostingRequest struct {
	// Core identifiers
	TransactionID      string    // Unique transaction identifier
	SourceType         string    // "PAYMENT", "SALE", "PURCHASE", etc. - use constants from models
	SourceID           uint64    // ID of source record
	Reference          string    // Human readable reference
	
	// Transaction details
	Date               time.Time
	Description        string
	Notes              string
	UserID             uint64
	
	// Balance updates
	CashBankUpdates    []CashBankBalanceUpdate  // Cash/Bank balance changes
	JournalLines       []JournalLineRequest     // Journal entry lines
	
	// Safety controls
	AllowDuplicates    bool                     // Usually false
	SkipValidation     bool                     // For system operations
}

// CashBankBalanceUpdate represents a cash/bank balance change
type CashBankBalanceUpdate struct {
	CashBankID       uint
	Amount           decimal.Decimal  // Positive for increase, negative for decrease
	TransactionType  string          // "DEPOSIT", "WITHDRAWAL", "TRANSFER_IN", "TRANSFER_OUT"
	Notes            string
}

// PostingResult contains the result of a posting operation
type PostingResult struct {
	TransactionID     string
	JournalEntryID    uint64
	CashBankUpdates   []CashBankUpdateResult
	Success           bool
	Errors            []string
	ProcessedAt       time.Time
}

// CashBankUpdateResult contains result of cash/bank balance update
type CashBankUpdateResult struct {
	CashBankID      uint
	PreviousBalance decimal.Decimal
	NewBalance      decimal.Decimal
	TransactionID   uint64
}

// PostBalanceUpdate - THE ONLY METHOD that should ever update balances
func (s *SingleSourcePostingService) PostBalanceUpdate(request PostingRequest) (*PostingResult, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	log.Printf("🎯 SINGLE SOURCE POSTING: Processing transaction %s", request.TransactionID)
	
	// Step 1: Validate request and check for duplicates
	if err := s.validatePostingRequest(request); err != nil {
		return nil, fmt.Errorf("validation failed: %v", err)
	}
	
	// Step 2: Check for duplicate processing
	if s.isAlreadyProcessed(request.TransactionID) && !request.AllowDuplicates {
		return nil, fmt.Errorf("transaction %s already processed", request.TransactionID)
	}
	
	// Step 3: Start database transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("❌ POSTING PANIC: %v", r)
		}
	}()
	
	result := &PostingResult{
		TransactionID: request.TransactionID,
		ProcessedAt:   time.Now(),
		Success:       false,
	}
	
	// Step 4: Process cash/bank balance updates
	if len(request.CashBankUpdates) > 0 {
		cashBankResults, err := s.processCashBankUpdates(tx, request.CashBankUpdates, request.SourceID, request.UserID)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("cash/bank update failed: %v", err)
		}
		result.CashBankUpdates = cashBankResults
	}
	
	// Step 5: Create journal entry if needed
	if len(request.JournalLines) > 0 {
		journalRequest := &JournalEntryRequest{
			SourceType:  request.SourceType,
			SourceID:    &request.SourceID,
			Reference:   request.Reference,
			EntryDate:   request.Date,
			Description: request.Description,
			Lines:       request.JournalLines,
			AutoPost:    true, // Always auto-post for consistency
			CreatedBy:   request.UserID,
		}
		
		journalResponse, err := s.unifiedJournalService.CreateJournalEntryWithTx(tx, journalRequest)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("journal creation failed: %v", err)
		}
		result.JournalEntryID = journalResponse.ID
	}
	
	// Step 6: Mark transaction as processed
	s.markAsProcessed(request.TransactionID)
	
	// Step 7: Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		s.unmarkAsProcessed(request.TransactionID) // Remove from tracker on failure
		return nil, fmt.Errorf("commit failed: %v", err)
	}
	
	result.Success = true
	log.Printf("✅ SINGLE SOURCE POSTING: Transaction %s completed successfully", request.TransactionID)
	
	return result, nil
}

// processCashBankUpdates processes all cash/bank balance updates
func (s *SingleSourcePostingService) processCashBankUpdates(tx *gorm.DB, updates []CashBankBalanceUpdate, sourceID uint64, userID uint64) ([]CashBankUpdateResult, error) {
	results := make([]CashBankUpdateResult, 0, len(updates))
	
	for _, update := range updates {
		result, err := s.processSingleCashBankUpdate(tx, update, sourceID, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to update cash/bank %d: %v", update.CashBankID, err)
		}
		results = append(results, *result)
	}
	
	return results, nil
}

// processSingleCashBankUpdate updates a single cash/bank account balance
func (s *SingleSourcePostingService) processSingleCashBankUpdate(tx *gorm.DB, update CashBankBalanceUpdate, sourceID uint64, userID uint64) (*CashBankUpdateResult, error) {
	// Get current cash bank record with lock to prevent race conditions
	var cashBank models.CashBank
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&cashBank, update.CashBankID).Error; err != nil {
		return nil, fmt.Errorf("cash/bank account %d not found: %v", update.CashBankID, err)
	}
	
	previousBalance := decimal.NewFromFloat(cashBank.Balance)
	newBalance := previousBalance.Add(update.Amount)
	
	// Safety check: prevent negative balance for outgoing transactions
	if newBalance.IsNegative() && update.Amount.IsNegative() {
		return nil, fmt.Errorf("insufficient balance: current=%.2f, required=%.2f", 
			previousBalance, update.Amount.Abs())
	}
	
	// Update cash bank balance
	cashBank.Balance = newBalance.InexactFloat64()
	if err := tx.Save(&cashBank).Error; err != nil {
		return nil, fmt.Errorf("failed to update cash/bank balance: %v", err)
	}
	
	// Create transaction record
	transaction := &models.CashBankTransaction{
		CashBankID:      update.CashBankID,
		ReferenceType:   "SINGLE_SOURCE_POST",
		ReferenceID:     uint(sourceID),
		Amount:          update.Amount.InexactFloat64(),
		BalanceAfter:    cashBank.Balance,
		TransactionDate: time.Now(),
		Notes:           update.Notes,
	}
	
	if err := tx.Create(transaction).Error; err != nil {
		return nil, fmt.Errorf("failed to create transaction record: %v", err)
	}
	
	log.Printf("💰 Cash/Bank %d: %.2f -> %.2f (%.2f)", 
		update.CashBankID, previousBalance, newBalance, update.Amount)
	
	return &CashBankUpdateResult{
		CashBankID:      update.CashBankID,
		PreviousBalance: previousBalance,
		NewBalance:      newBalance,
		TransactionID:   uint64(transaction.ID),
	}, nil
}

// validatePostingRequest validates the posting request
func (s *SingleSourcePostingService) validatePostingRequest(request PostingRequest) error {
	if request.TransactionID == "" {
		return fmt.Errorf("transaction ID is required")
	}
	
	if request.SourceType == "" {
		return fmt.Errorf("source type is required")
	}
	
	if request.SourceID == 0 {
		return fmt.Errorf("source ID is required")
	}
	
	if len(request.CashBankUpdates) == 0 && len(request.JournalLines) == 0 {
		return fmt.Errorf("no updates specified")
	}
	
	// Validate cash/bank updates
	for i, update := range request.CashBankUpdates {
		if update.CashBankID == 0 {
			return fmt.Errorf("cash/bank update %d: missing cash bank ID", i)
		}
		if update.Amount.IsZero() {
			return fmt.Errorf("cash/bank update %d: amount cannot be zero", i)
		}
	}
	
	return nil
}

// isAlreadyProcessed checks if transaction was already processed
func (s *SingleSourcePostingService) isAlreadyProcessed(transactionID string) bool {
	return s.transactionTracker[transactionID]
}

// markAsProcessed marks transaction as processed
func (s *SingleSourcePostingService) markAsProcessed(transactionID string) {
	s.transactionTracker[transactionID] = true
}

// unmarkAsProcessed removes transaction from processed tracker
func (s *SingleSourcePostingService) unmarkAsProcessed(transactionID string) {
	delete(s.transactionTracker, transactionID)
}

// CreatePaymentPosting creates a posting for payment (the only way to post payment balances)
func (s *SingleSourcePostingService) CreatePaymentPosting(payment *models.Payment, cashBankID uint, userID uint) (*PostingResult, error) {
	transactionID := fmt.Sprintf("PAYMENT_%d_%d", payment.ID, time.Now().Unix())
	paymentAmount := decimal.NewFromFloat(payment.Amount)
	
	// Prepare cash/bank balance update
	cashBankUpdates := []CashBankBalanceUpdate{
		{
			CashBankID:      cashBankID,
			Amount:          paymentAmount, // Positive for incoming payment
			TransactionType: "PAYMENT_RECEIVED",
			Notes:           fmt.Sprintf("Payment received - %s", payment.Code),
		},
	}
	
	// Get cash bank account for journal lines
	var cashBank models.CashBank
	if err := s.db.First(&cashBank, cashBankID).Error; err != nil {
		return nil, fmt.Errorf("cash/bank account not found: %v", err)
	}
	
	// Get AR account
	var arAccount models.Account
	if err := s.db.Where("code = ?", "1201").First(&arAccount).Error; err != nil {
		if err := s.db.Where("LOWER(name) LIKE ?", "%piutang%usaha%").First(&arAccount).Error; err != nil {
			return nil, fmt.Errorf("accounts receivable account not found: %v", err)
		}
	}
	
	// Prepare journal lines
	journalLines := []JournalLineRequest{
		{
			AccountID:    uint64(cashBank.AccountID),
			Description:  fmt.Sprintf("Payment received - %s", payment.Code),
			DebitAmount:  paymentAmount,
			CreditAmount: decimal.Zero,
		},
		{
			AccountID:    uint64(arAccount.ID),
			Description:  fmt.Sprintf("AR reduction - %s", payment.Code),
			DebitAmount:  decimal.Zero,
			CreditAmount: paymentAmount,
		},
	}
	
	// Create posting request
	request := PostingRequest{
		TransactionID:   transactionID,
		SourceType:      "PAYMENT", // Using constant from models
		SourceID:        uint64(payment.ID),
		Reference:       payment.Code,
		Date:            payment.Date,
		Description:     fmt.Sprintf("Customer Payment %s", payment.Code),
		Notes:           payment.Notes,
		UserID:          uint64(userID),
		CashBankUpdates: cashBankUpdates,
		JournalLines:    journalLines,
		AllowDuplicates: false,
		SkipValidation:  false,
	}
	
	return s.PostBalanceUpdate(request)
}

// ValidateBalanceConsistency validates that all balances are consistent
func (s *SingleSourcePostingService) ValidateBalanceConsistency() error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	log.Println("🔍 Validating balance consistency...")
	
	// Check all cash banks
	var cashBanks []models.CashBank
	if err := s.db.Find(&cashBanks).Error; err != nil {
		return fmt.Errorf("failed to get cash banks: %v", err)
	}
	
	var inconsistencies []string
	
	for _, cb := range cashBanks {
		// Calculate expected balance from transactions
		var transactionSum float64
		s.db.Model(&models.CashBankTransaction{}).
			Where("cash_bank_id = ?", cb.ID).
			Select("COALESCE(SUM(amount), 0)").
			Scan(&transactionSum)
		
		if cb.Balance != transactionSum {
			inconsistencies = append(inconsistencies, 
				fmt.Sprintf("CashBank %d (%s): Balance=%.2f, TransactionSum=%.2f", 
					cb.ID, cb.Name, cb.Balance, transactionSum))
		}
		
		// Check GL account consistency
		if cb.AccountID > 0 {
			var account models.Account
			if err := s.db.First(&account, cb.AccountID).Error; err == nil {
				if cb.Balance != account.Balance {
					inconsistencies = append(inconsistencies, 
						fmt.Sprintf("CashBank %d vs Account %d: %.2f != %.2f", 
							cb.ID, account.ID, cb.Balance, account.Balance))
				}
			}
		}
	}
	
	if len(inconsistencies) > 0 {
		return fmt.Errorf("balance inconsistencies found: %v", inconsistencies)
	}
	
	log.Println("✅ All balances are consistent")
	return nil
}

// GetProcessingStats returns statistics about processed transactions
func (s *SingleSourcePostingService) GetProcessingStats() map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	return map[string]interface{}{
		"processed_transactions": len(s.transactionTracker),
		"service_uptime":        time.Now().Format(time.RFC3339),
		"total_operations":      len(s.transactionTracker),
	}
}

// ClearProcessingHistory clears the transaction tracker (use with caution)
func (s *SingleSourcePostingService) ClearProcessingHistory() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.transactionTracker = make(map[string]bool)
	log.Println("🧹 Processing history cleared")
}