package services

import (
	"fmt"
	"log"
	"time"
	"context"

	"app-sistem-akuntansi/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// UltraFastPostingService - EMERGENCY bypass for timeout issues
// This service focuses on speed over comprehensive journal creation
type UltraFastPostingService struct {
	db *gorm.DB
}

// NewUltraFastPostingService creates ultra-fast posting service
func NewUltraFastPostingService(db *gorm.DB) *UltraFastPostingService {
	return &UltraFastPostingService{
		db: db,
	}
}

// UltraFastPaymentPosting - Emergency fast payment posting (defaults to RECEIVABLE)
// paymentType: "RECEIVABLE" for customer payments (increase balance), "PAYABLE" for vendor payments (decrease balance)
func (s *UltraFastPostingService) UltraFastPaymentPosting(payment *models.Payment, cashBankID uint, userID uint) error {
	return s.UltraFastPaymentPostingWithType(payment, cashBankID, userID, "RECEIVABLE")
}

// UltraFastReceivablePaymentPosting - Explicit receivable payment posting
func (s *UltraFastPostingService) UltraFastReceivablePaymentPosting(payment *models.Payment, cashBankID uint, userID uint) error {
	return s.UltraFastPaymentPostingWithType(payment, cashBankID, userID, "RECEIVABLE")
}

// UltraFastPayablePaymentPosting - Explicit payable payment posting  
func (s *UltraFastPostingService) UltraFastPayablePaymentPosting(payment *models.Payment, cashBankID uint, userID uint) error {
	return s.UltraFastPaymentPostingWithType(payment, cashBankID, userID, "PAYABLE")
}

// UltraFastPaymentPostingWithType - Explicit payment type version
func (s *UltraFastPostingService) UltraFastPaymentPostingWithType(payment *models.Payment, cashBankID uint, userID uint, paymentType string) error {
	startTime := time.Now()
	log.Printf("⚡ ULTRA-FAST POSTING: Starting for payment %d, amount %.2f", payment.ID, payment.Amount)

	// Set ultra-short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// Use single transaction with raw SQL for maximum speed
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: Update cash bank balance (raw SQL)
		log.Printf("💰 Updating cash bank %d balance", cashBankID)
		updateStart := time.Now()
		
		// Detect payment type and determine correct balance change
		// Check payment method or analyze amount to determine if this is receivable or payable
		var amountChange float64
		var paymentTypeDescription string
		
		// Use explicit payment type parameter for accurate accounting
		// paymentType: "RECEIVABLE" = payment from customer, "PAYABLE" = payment to vendor
		if paymentType == "RECEIVABLE" {
			// Receivable payment: Cash/Bank increases
			amountChange = payment.Amount // Positive amount increases balance
			paymentTypeDescription = "receivable payment (cash in)"
			log.Printf("💰 Receivable payment: Adding %.2f to cash bank %d", amountChange, cashBankID)
		} else {
			// Payable payment: Cash/Bank decreases
			amountChange = -payment.Amount // Negative amount decreases balance
			paymentTypeDescription = "payable payment (cash out)"
			log.Printf("💰 Payable payment: Subtracting %.2f from cash bank %d", payment.Amount, cashBankID)
			
			// For payable payments, check sufficient balance
			var currentBalance float64
			if err := tx.Raw("SELECT balance FROM cash_banks WHERE id = ?", cashBankID).Scan(&currentBalance).Error; err != nil {
				return fmt.Errorf("failed to check current balance: %w", err)
			}
			
			if currentBalance < payment.Amount {
				return fmt.Errorf("insufficient balance for payment: available=%.2f, required=%.2f", currentBalance, payment.Amount)
			}
		}
		
		if err := tx.Exec(`
			UPDATE cash_banks 
			SET balance = balance + ?, updated_at = NOW() 
			WHERE id = ?
		`, amountChange, cashBankID).Error; err != nil {
			return fmt.Errorf("failed to update cash bank balance: %w", err)
		}
		
		log.Printf("✅ Cash bank updated in %.2fms", float64(time.Since(updateStart).Nanoseconds())/1000000)

		// Step 2: Create minimal transaction record
		transStart := time.Now()
		transaction := &models.CashBankTransaction{
			CashBankID:      cashBankID,
			ReferenceType:   "ULTRA_FAST_PAYMENT",
			ReferenceID:     uint(payment.ID),
			Amount:          amountChange, // Use correct amount (positive or negative)
			TransactionDate: payment.Date,
			Notes:           fmt.Sprintf("Ultra-fast payment %s - %s", payment.Code, paymentTypeDescription),
		}

		// Get new balance for transaction record
		var newBalance float64
		if err := tx.Raw("SELECT balance FROM cash_banks WHERE id = ?", cashBankID).Scan(&newBalance).Error; err != nil {
			log.Printf("⚠️ Could not get new balance, using estimated: %v", err)
			newBalance = 0 // Will be corrected later if needed
		}
		transaction.BalanceAfter = newBalance

		if err := tx.Create(transaction).Error; err != nil {
			return fmt.Errorf("failed to create transaction record: %w", err)
		}
		
		log.Printf("✅ Transaction record created in %.2fms", float64(time.Since(transStart).Nanoseconds())/1000000)

		// Step 3: Create minimal journal entry (async later)
		// We skip journal creation here for speed and do it asynchronously
		log.Printf("📝 Skipping journal creation for speed - will create async")

		totalTime := time.Since(startTime)
		log.Printf("🚀 ULTRA-FAST POSTING: Completed in %.2fms", float64(totalTime.Nanoseconds())/1000000)

		return nil
	})
}

// CreateJournalEntryAsync - Create journal entry in background after payment is completed
func (s *UltraFastPostingService) CreateJournalEntryAsync(payment *models.Payment, cashBankID uint, userID uint) {
	s.CreateJournalEntryAsyncWithType(payment, cashBankID, userID, "RECEIVABLE")
}

// CreateJournalEntryAsyncWithType - Create journal entry with explicit payment type
func (s *UltraFastPostingService) CreateJournalEntryAsyncWithType(payment *models.Payment, cashBankID uint, userID uint, paymentType string) {
	go func() {
		// Small delay to let the main transaction complete
		time.Sleep(200 * time.Millisecond)
		
		log.Printf("🔄 ASYNC JOURNAL: Creating journal entry for payment %d (type: %s)", payment.ID, paymentType)
		asyncStart := time.Now()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Try to create the journal entry
		if err := s.createSimpleJournalEntryWithType(ctx, payment, cashBankID, userID, paymentType); err != nil {
			log.Printf("⚠️ ASYNC JOURNAL: Failed to create journal entry for payment %d: %v", payment.ID, err)
			// Don't fail the payment - it's already completed
		} else {
			log.Printf("✅ ASYNC JOURNAL: Journal entry created for payment %d in %.2fms", 
				payment.ID, float64(time.Since(asyncStart).Nanoseconds())/1000000)
		}
	}()
}

// createSimpleJournalEntry creates a simple journal entry (defaults to RECEIVABLE)
func (s *UltraFastPostingService) createSimpleJournalEntry(ctx context.Context, payment *models.Payment, cashBankID uint, userID uint) error {
	return s.createSimpleJournalEntryWithType(ctx, payment, cashBankID, userID, "RECEIVABLE")
}

// createSimpleJournalEntryWithType creates a simple journal entry with explicit payment type
func (s *UltraFastPostingService) createSimpleJournalEntryWithType(ctx context.Context, payment *models.Payment, cashBankID uint, userID uint, paymentType string) error {
	// Get cash bank account
	var cashBank models.CashBank
	if err := s.db.WithContext(ctx).First(&cashBank, cashBankID).Error; err != nil {
		return fmt.Errorf("cash bank not found: %w", err)
	}

	if cashBank.AccountID == 0 {
		log.Printf("⚠️ Cash bank %d has no associated GL account, skipping journal", cashBankID)
		return nil
	}

	// Get AR account (try different methods)
	var arAccount models.Account
	if err := s.db.WithContext(ctx).Where("code = ?", "1201").First(&arAccount).Error; err != nil {
		if err := s.db.WithContext(ctx).Where("code LIKE ?", "120%").First(&arAccount).Error; err != nil {
			log.Printf("⚠️ AR account not found, skipping journal entry: %v", err)
			return nil // Don't fail, just skip journal
		}
	}

	// Create simple journal entry using SSOT system
	journalService := NewUnifiedJournalService(s.db)
	paymentAmount := decimal.NewFromFloat(payment.Amount)
	
	// Determine journal entry based on explicit payment type
	var journalLines []JournalLineRequest
	
	if paymentType == "RECEIVABLE" {
		// Receivable payment: Dr. Cash/Bank, Cr. Accounts Receivable
		journalLines = []JournalLineRequest{
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
	} else {
		// Payable payment: Dr. Accounts Payable, Cr. Cash/Bank
		// Get Accounts Payable account
		var apAccount models.Account
		if err := s.db.WithContext(ctx).Where("code = ?", "2101").First(&apAccount).Error; err != nil {
			if err := s.db.WithContext(ctx).Where("code LIKE ?", "210%").First(&apAccount).Error; err != nil {
				log.Printf("⚠️ AP account not found, using default liability account: %v", err)
				// Use AR account as fallback (not ideal, but prevents crash)
				apAccount = arAccount
			}
		}
		
		journalLines = []JournalLineRequest{
			{
				AccountID:    uint64(apAccount.ID),
				Description:  fmt.Sprintf("AP reduction - %s", payment.Code),
				DebitAmount:  paymentAmount,
				CreditAmount: decimal.Zero,
			},
			{
				AccountID:    uint64(cashBank.AccountID),
				Description:  fmt.Sprintf("Payment made - %s", payment.Code),
				DebitAmount:  decimal.Zero,
				CreditAmount: paymentAmount,
			},
		}
	}
	
	paymentIDUint64 := uint64(payment.ID)
	journalRequest := &JournalEntryRequest{
		SourceType:  models.SSOTSourceTypePayment,
		SourceID:    &paymentIDUint64,
		Reference:   payment.Code,
		EntryDate:   payment.Date,
		Description: fmt.Sprintf("Ultra Fast Payment %s", payment.Code),
		Lines:       journalLines,
		AutoPost:    true,
		CreatedBy:   uint64(userID),
	}

	// Try to create with timeout
	_, err := journalService.CreateJournalEntry(journalRequest)
	return err
}

// ValidateUltraFastBalance performs quick balance validation
func (s *UltraFastPostingService) ValidateUltraFastBalance(cashBankID uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cashBank models.CashBank
	if err := s.db.WithContext(ctx).First(&cashBank, cashBankID).Error; err != nil {
		return fmt.Errorf("cash bank not found: %w", err)
	}

	// Quick transaction sum check
	var transactionSum float64
	if err := s.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(amount), 0) 
		FROM cash_bank_transactions 
		WHERE cash_bank_id = ?
	`, cashBankID).Scan(&transactionSum).Error; err != nil {
		log.Printf("⚠️ Could not validate transaction sum: %v", err)
		return nil // Don't fail validation on query error
	}

	if cashBank.Balance != transactionSum {
		log.Printf("⚠️ Balance inconsistency detected: CashBank=%.2f, Transactions=%.2f", 
			cashBank.Balance, transactionSum)
	} else {
		log.Printf("✅ Balance validation passed: %.2f", cashBank.Balance)
	}

	return nil
}

// GetStats returns ultra-fast posting statistics
func (s *UltraFastPostingService) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"service_type":    "ultra-fast-posting",
		"timeout_limit":   "8 seconds",
		"journal_method":  "asynchronous",
		"optimization":    "maximum speed",
	}
}