package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
	"github.com/xuri/excelize/v2"
)

type PaymentService struct {
	db              *gorm.DB
	paymentRepo     *repositories.PaymentRepository
	salesRepo       *repositories.SalesRepository
	purchaseRepo    *repositories.PurchaseRepository
	cashBankRepo    *repositories.CashBankRepository
	accountRepo     repositories.AccountRepository
	contactRepo     repositories.ContactRepository
}

func NewPaymentService(
	db *gorm.DB,
	paymentRepo *repositories.PaymentRepository,
	salesRepo *repositories.SalesRepository,
	purchaseRepo *repositories.PurchaseRepository,
	cashBankRepo *repositories.CashBankRepository,
	accountRepo repositories.AccountRepository,
	contactRepo repositories.ContactRepository,
) *PaymentService {
	return &PaymentService{
		db:           db,
		paymentRepo:  paymentRepo,
		salesRepo:    salesRepo,
		purchaseRepo: purchaseRepo,
		cashBankRepo: cashBankRepo,
		accountRepo:  accountRepo,
		contactRepo:  contactRepo,
	}
}

// Payment Types
const (
	PaymentTypeReceivable = "RECEIVABLE" // Payment from customer
	PaymentTypePayable    = "PAYABLE"    // Payment to vendor
	PaymentTypeAdvance    = "ADVANCE"    // Advance payment
	PaymentTypeRefund     = "REFUND"     // Refund payment
)

// CreateReceivablePayment creates payment for sales/receivables (Fixed version)
func (s *PaymentService) CreateReceivablePayment(request PaymentCreateRequest, userID uint) (*models.Payment, error) {
	// Use the fixed version with better logging and error handling
	return s.CreateReceivablePaymentFixed(request, userID)
}

// CreateReceivablePaymentFixed - Fixed version with better error handling and timeout
func (s *PaymentService) CreateReceivablePaymentFixed(request PaymentCreateRequest, userID uint) (*models.Payment, error) {
	startTime := time.Now()
	log.Printf("🚀 Starting CreateReceivablePayment: ContactID=%d, Amount=%.2f, Allocations=%d", 
		request.ContactID, request.Amount, len(request.Allocations))
	
	// Start transaction with extended timeout for complex operations
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", tx.Error)
	}
	
	// Robust transaction handling with automatic rollback
	var committed bool
	defer func() {
		if r := recover(); r != nil {
			log.Printf("❌ PANIC in CreateReceivablePayment: %v", r)
			if !committed {
				tx.Rollback()
			}
			panic(r)
		} else if !committed {
			// Auto-rollback if not committed
			tx.Rollback()
		}
	}()
	
	// Step 1: Validate customer
	log.Printf("📝 Step 1: Validating customer...")
	stepStart := time.Now()
	_, err := s.contactRepo.GetByID(request.ContactID)
	if err != nil {
		log.Printf("❌ Customer validation failed: %v (%.2fms)", err, float64(time.Since(stepStart).Nanoseconds())/1000000)
		return nil, fmt.Errorf("customer not found: %v", err)
	}
	log.Printf("✅ Customer validated (%.2fms)", float64(time.Since(stepStart).Nanoseconds())/1000000)
	
	// Step 2: Generate payment code
	log.Printf("📝 Step 2: Generating payment code...")
	stepStart = time.Now()
	code := s.generatePaymentCode("RCV")
	log.Printf("✅ Payment code generated: %s (%.2fms)", code, float64(time.Since(stepStart).Nanoseconds())/1000000)
	
	// Step 3: Create payment record
	log.Printf("📝 Step 3: Creating payment record...")
	stepStart = time.Now()
	payment := &models.Payment{
		Code:      code,
		ContactID: request.ContactID,
		UserID:    userID,
		Date:      request.Date,
		Amount:    request.Amount,
		Method:    request.Method,
		Reference: request.Reference,
		Status:    models.PaymentStatusPending,
		Notes:     request.Notes,
	}
	
	if err := tx.Create(payment).Error; err != nil {
		log.Printf("❌ Failed to create payment: %v (%.2fms)", err, float64(time.Since(stepStart).Nanoseconds())/1000000)
		return nil, fmt.Errorf("failed to create payment: %v", err)
	}
	log.Printf("✅ Payment record created: ID=%d (%.2fms)", payment.ID, float64(time.Since(stepStart).Nanoseconds())/1000000)
	
	// Step 4: Process allocations
	log.Printf("📝 Step 4: Processing %d allocations...", len(request.Allocations))
	stepStart = time.Now()
	remainingAmount := request.Amount
	
	for i, allocation := range request.Allocations {
		if remainingAmount <= 0 {
			log.Printf("⚠️ No remaining amount, skipping allocation %d", i+1)
			break
		}
		
		log.Printf("📝 Processing allocation %d: InvoiceID=%d, Amount=%.2f", i+1, allocation.InvoiceID, allocation.Amount)
		
		// Get sale
		sale, err := s.salesRepo.FindByID(allocation.InvoiceID)
		if err != nil {
			log.Printf("❌ Invoice %d not found: %v", allocation.InvoiceID, err)
			return nil, fmt.Errorf("invoice %d not found: %v", allocation.InvoiceID, err)
		}
		
		// Validate ownership
		if sale.CustomerID != request.ContactID {
			log.Printf("❌ Invoice ownership mismatch: Sale.CustomerID=%d, Request.ContactID=%d", sale.CustomerID, request.ContactID)
			return nil, fmt.Errorf("invoice does not belong to this customer")
		}
		
		// Calculate allocated amount
		allocatedAmount := allocation.Amount
		if allocatedAmount > remainingAmount {
			allocatedAmount = remainingAmount
			log.Printf("⚠️ Adjusting amount to remaining: %.2f -> %.2f", allocation.Amount, allocatedAmount)
		}
		if allocatedAmount > sale.OutstandingAmount {
			allocatedAmount = sale.OutstandingAmount
			log.Printf("⚠️ Adjusting amount to outstanding: %.2f -> %.2f", allocatedAmount, sale.OutstandingAmount)
		}
		
		// Create payment allocation
		paymentAllocation := &models.PaymentAllocation{
			PaymentID:       payment.ID,
			InvoiceID:       &allocation.InvoiceID,
			AllocatedAmount: allocatedAmount,
		}
		
		if err := tx.Create(paymentAllocation).Error; err != nil {
			log.Printf("❌ Failed to create payment allocation: %v", err)
			return nil, fmt.Errorf("failed to create payment allocation: %v", err)
		}
		log.Printf("✅ Payment allocation created: %.2f", allocatedAmount)
		
		// Update sale amounts
		log.Printf("📝 Updating sale amounts: PaidAmount %.2f -> %.2f, Outstanding %.2f -> %.2f", 
			sale.PaidAmount, sale.PaidAmount + allocatedAmount,
			sale.OutstandingAmount, sale.OutstandingAmount - allocatedAmount)
			
		sale.PaidAmount += allocatedAmount
		sale.OutstandingAmount -= allocatedAmount
		
		// Update status
		if sale.OutstandingAmount <= 0 {
			sale.Status = models.SaleStatusPaid
			log.Printf("✅ Sale status updated to PAID")
		} else if sale.PaidAmount > 0 && sale.Status == models.SaleStatusInvoiced {
			sale.Status = models.SaleStatusInvoiced
			log.Printf("✅ Sale status remains INVOICED (partial payment)")
		}
		
		// Save sale changes
		if err := tx.Save(sale).Error; err != nil {
			log.Printf("❌ Failed to save sale: %v", err)
			return nil, fmt.Errorf("failed to update sale: %v", err)
		}
		log.Printf("✅ Sale updated successfully")
		
		// Create SalePayment cross-reference
		salePayment := &models.SalePayment{
			SaleID:        sale.ID,
			PaymentNumber: fmt.Sprintf("PAY-%s", payment.Code),
			Date:          payment.Date,
			Amount:        allocatedAmount,
			Method:        payment.Method,
			Reference:     fmt.Sprintf("Payment ID: %d", payment.ID),
			Notes:         fmt.Sprintf("Created from Payment Management - %s", payment.Notes),
			CashBankID:    &request.CashBankID,
			UserID:        userID,
			PaymentID:     &payment.ID,
		}
		
		if err := tx.Create(salePayment).Error; err != nil {
			log.Printf("❌ CRITICAL: Failed to create sale payment cross-reference for payment %d, sale %d: %v", payment.ID, sale.ID, err)
			// This is critical - if this fails, return error for auto-rollback
			return nil, fmt.Errorf("failed to create sale payment record: %v", err)
		} else {
			log.Printf("✅ Sale payment cross-reference created: payment_id=%d, sale_id=%d, amount=%.2f", payment.ID, sale.ID, allocatedAmount)
		}
		
		remainingAmount -= allocatedAmount
		log.Printf("✅ Allocation %d complete. Remaining: %.2f", i+1, remainingAmount)
	}
	log.Printf("✅ All allocations processed (%.2fms)", float64(time.Since(stepStart).Nanoseconds())/1000000)
	
	// Step 5: Update cash/bank balance
	log.Printf("📝 Step 5: Updating cash/bank balance...")
	stepStart = time.Now()
	if request.CashBankID > 0 {
		err = s.updateCashBankBalanceWithLogging(tx, request.CashBankID, request.Amount, "IN", payment.ID, userID)
		if err != nil {
			log.Printf("❌ Failed to update cash/bank balance: %v (%.2fms)", err, float64(time.Since(stepStart).Nanoseconds())/1000000)
			return nil, fmt.Errorf("failed to update cash/bank balance: %v", err)
		}
	}
	log.Printf("✅ Cash/bank balance updated (%.2fms)", float64(time.Since(stepStart).Nanoseconds())/1000000)
	
	// Step 6: Create journal entries
	log.Printf("📝 Step 6: Creating journal entries...")
	stepStart = time.Now()
	err = s.createReceivablePaymentJournalWithLogging(tx, payment, request.CashBankID, userID)
	if err != nil {
		log.Printf("❌ Failed to create journal entries: %v (%.2fms)", err, float64(time.Since(stepStart).Nanoseconds())/1000000)
		return nil, fmt.Errorf("failed to create journal entries: %v", err)
	}
	log.Printf("✅ Journal entries created (%.2fms)", float64(time.Since(stepStart).Nanoseconds())/1000000)
	
	// Step 7: Update payment status
	log.Printf("📝 Step 7: Updating payment status to COMPLETED...")
	stepStart = time.Now()
	payment.Status = models.PaymentStatusCompleted
	if err := tx.Save(payment).Error; err != nil {
		log.Printf("❌ Failed to save payment status: %v (%.2fms)", err, float64(time.Since(stepStart).Nanoseconds())/1000000)
		return nil, fmt.Errorf("failed to update payment status: %v", err)
	}
	log.Printf("✅ Payment status updated (%.2fms)", float64(time.Since(stepStart).Nanoseconds())/1000000)
	
	// Step 8: Commit transaction
	log.Printf("📋 Step 8: Committing transaction...")
	stepStart = time.Now()
	commitErr := tx.Commit().Error
	if commitErr != nil {
		log.Printf("❌ CRITICAL: Failed to commit transaction: %v (%.2fms)", commitErr, float64(time.Since(stepStart).Nanoseconds())/1000000)
		return nil, fmt.Errorf("transaction commit failed: %v", commitErr)
	}
	committed = true // Mark as committed to prevent auto-rollback
	log.Printf("✅ Transaction committed successfully (%.2fms)", float64(time.Since(stepStart).Nanoseconds())/1000000)
	
	totalTime := time.Since(startTime)
	log.Printf("🎉 CreateReceivablePayment COMPLETED: ID=%d, Code=%s, ContactID=%d, Amount=%.2f, TotalTime=%.2fms", 
		payment.ID, payment.Code, payment.ContactID, payment.Amount, float64(totalTime.Nanoseconds())/1000000)
	
	return payment, nil
}

// CreatePayablePayment creates payment for purchases/payables
func (s *PaymentService) CreatePayablePayment(request PaymentCreateRequest, userID uint) (*models.Payment, error) {
	start := time.Now()
	log.Printf("Starting CreatePayablePayment: ContactID=%d, Amount=%.2f", request.ContactID, request.Amount)
	
	// Start transaction with timeout
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in CreatePayablePayment: %v", r)
			tx.Rollback()
		}
	}()
	
	// Validate vendor with optimized query
	var contact models.Contact
	if err := tx.Select("id, name, type").First(&contact, request.ContactID).Error; err != nil {
		tx.Rollback()
		log.Printf("Vendor validation failed: %v", err)
		return nil, errors.New("vendor not found")
	}
	log.Printf("Vendor validated: %s (ID: %d)", contact.Name, contact.ID)
	
	// Check cash/bank balance
	log.Printf("Checking balance for CashBankID: %d", request.CashBankID)
	balanceCheckStart := time.Now()
	if request.CashBankID > 0 {
		cashBank, err := s.cashBankRepo.FindByID(request.CashBankID)
		if err != nil {
			tx.Rollback()
			log.Printf("Cash/bank account not found: %v", err)
			return nil, errors.New("cash/bank account not found")
		}
		
		if cashBank.Balance < request.Amount {
			tx.Rollback()
			log.Printf("Insufficient balance: Available=%.2f, Required=%.2f", cashBank.Balance, request.Amount)
			return nil, fmt.Errorf("insufficient balance. Available: %.2f", cashBank.Balance)
		}
		log.Printf("Balance check passed: %.2f available (%.2fms)", cashBank.Balance, float64(time.Since(balanceCheckStart).Nanoseconds())/1000000)
	}
	
	// Generate payment code
	codeGenStart := time.Now()
	code := s.generatePaymentCode("PAY")
	log.Printf("Payment code generated: %s (%.2fms)", code, float64(time.Since(codeGenStart).Nanoseconds())/1000000)
	
	// Create payment record
	payment := &models.Payment{
		Code:      code,
		ContactID: request.ContactID,
		UserID:    userID,
		Date:      request.Date,
		Amount:    request.Amount,
		Method:    request.Method,
		Reference: request.Reference,
		Status:    models.PaymentStatusPending,
		Notes:     request.Notes,
	}
	
	if err := tx.Create(payment).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	// Process allocations to bills
	remainingAmount := request.Amount
	for _, allocation := range request.BillAllocations {
		if remainingAmount <= 0 {
			break
		}
		
		var purchase models.Purchase
		if err := tx.First(&purchase, allocation.BillID).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("bill %d not found", allocation.BillID)
		}
		
		if purchase.VendorID != request.ContactID {
			tx.Rollback()
			return nil, errors.New("bill does not belong to this vendor")
		}
		
		allocatedAmount := allocation.Amount
		if allocatedAmount > remainingAmount {
			allocatedAmount = remainingAmount
		}
		
		// Calculate outstanding (simplified - would need proper tracking)
		outstandingAmount := purchase.TotalAmount // This should be tracked properly
		if allocatedAmount > outstandingAmount {
			allocatedAmount = outstandingAmount
		}
		
		// Create payment allocation
		paymentAllocation := &models.PaymentAllocation{
			PaymentID:       payment.ID,
			BillID:          &allocation.BillID,
			AllocatedAmount: allocatedAmount,
		}
		
		if err := tx.Create(paymentAllocation).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		
		remainingAmount -= allocatedAmount
	}
	
	// Declare err variable for use in this scope
	var err error
	
	// Update cash/bank account
	if request.CashBankID > 0 {
		err = s.updateCashBankBalance(tx, request.CashBankID, -request.Amount, "OUT", payment.ID, userID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	
	// Create journal entries
	err = s.createPayablePaymentJournal(tx, payment, request.CashBankID, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	
	payment.Status = models.PaymentStatusCompleted
	if err := tx.Save(payment).Error; err != nil {
		tx.Rollback()
		log.Printf("Failed to save payment status: %v", err)
		return nil, err
	}
	
	log.Printf("Payment creation completed, committing transaction...")
	if err := tx.Commit().Error; err != nil {
		log.Printf("Failed to commit payment transaction: %v", err)
		return nil, err
	}
	
	totalTime := time.Since(start)
	log.Printf("✅ CreatePayablePayment completed successfully: ID=%d, Code=%s, Amount=%.2f, TotalTime=%.2fms", 
		payment.ID, payment.Code, payment.Amount, float64(totalTime.Nanoseconds())/1000000)
	
	return payment, nil
}

// updateCashBankBalance updates cash/bank balance and creates transaction record
func (s *PaymentService) updateCashBankBalance(tx *gorm.DB, cashBankID uint, amount float64, direction string, referenceID uint, userID uint) error {
	var cashBank models.CashBank
	if err := tx.First(&cashBank, cashBankID).Error; err != nil {
		return fmt.Errorf("cash/bank account not found: %v", err)
	}
	
	log.Printf("Updating Cash/Bank Balance: ID=%d, Name=%s, CurrentBalance=%.2f, Amount=%.2f, Direction=%s", 
		cashBankID, cashBank.Name, cashBank.Balance, amount, direction)
	
	// For receivable payments (IN), amount should be positive, balance increases
	// For payable payments (OUT), amount should be negative, balance decreases
	
	// For outgoing payments, validate sufficient balance BEFORE updating
	if direction == "OUT" && amount < 0 {
		requiredAmount := -amount // Convert negative to positive
		if cashBank.Balance < requiredAmount {
			return fmt.Errorf("insufficient balance. Available: %.2f, Required: %.2f", cashBank.Balance, requiredAmount)
		}
	}
	
	// Update balance
	newBalance := cashBank.Balance + amount
	
	// Final safety check - balance should never go negative
	if newBalance < 0 {
		return fmt.Errorf("transaction would result in negative balance. Current: %.2f, Change: %.2f, Result: %.2f", 
			cashBank.Balance, amount, newBalance)
	}
	
	cashBank.Balance = newBalance
	log.Printf("Balance updated successfully: %.2f -> %.2f", cashBank.Balance-amount, cashBank.Balance)
	
	if err := tx.Save(&cashBank).Error; err != nil {
		return err
	}
	
	// Create transaction record
	transaction := &models.CashBankTransaction{
		CashBankID:      cashBankID,
		ReferenceType:   "PAYMENT",
		ReferenceID:     referenceID,
		Amount:          amount,
		BalanceAfter:    cashBank.Balance,
		TransactionDate: time.Now(),
		Notes:           fmt.Sprintf("Payment %s", direction),
	}
	
	return tx.Create(transaction).Error
}

// createReceivablePaymentJournal creates journal entries for receivable payment
func (s *PaymentService) createReceivablePaymentJournal(tx *gorm.DB, payment *models.Payment, cashBankID uint, userID uint) error {
	// Get accounts
	var cashBankAccountID uint
	if cashBankID > 0 {
		var cashBank models.CashBank
		if err := tx.First(&cashBank, cashBankID).Error; err != nil {
			return err
		}
		cashBankAccountID = cashBank.AccountID
	} else {
		// If no specific bank account, use default Kas account (1101)
		var kasAccount models.Account
		if err := tx.Where("code = ?", "1101").First(&kasAccount).Error; err != nil {
			return fmt.Errorf("default cash account (1101) not found: %v", err)
		}
		cashBankAccountID = kasAccount.ID
	}

	// Get Accounts Receivable account (Piutang Usaha - 1201)
	var arAccount models.Account
	if err := tx.Where("code = ?", "1201").First(&arAccount).Error; err != nil {
		log.Printf("Warning: Piutang Usaha account (1201) not found, using fallback")
		// Fallback: try to find by name pattern
		if err := tx.Where("LOWER(name) LIKE ?", "%piutang%usaha%").First(&arAccount).Error; err != nil {
			return fmt.Errorf("accounts receivable account not found: %v", err)
		}
	}
	arAccountID := arAccount.ID

	log.Printf("Journal Entry Mapping: CashBank AccountID=%d, AR AccountID=%d (Code=%s)", cashBankAccountID, arAccountID, arAccount.Code)

	// Create journal entry
	journalEntry := &models.JournalEntry{
		// Code will be auto-generated by BeforeCreate hook
		EntryDate:     payment.Date,
		Description:   fmt.Sprintf("Customer Payment %s", payment.Code),
		ReferenceType: models.JournalRefPayment,
		ReferenceID:   &payment.ID,
		Reference:     payment.Code,
		UserID:        userID,
		Status:        models.JournalStatusPosted,
		TotalDebit:    payment.Amount,
		TotalCredit:   payment.Amount,
		IsAutoGenerated: true,
	}

	// Journal lines
	journalLines := []models.JournalLine{
		// Debit: Cash/Bank
		{
			AccountID:    cashBankAccountID,
			Description:  fmt.Sprintf("Payment received - %s", payment.Code),
			DebitAmount:  payment.Amount,
			CreditAmount: 0,
		},
		// Credit: Accounts Receivable
		{
			AccountID:    arAccountID,
			Description:  fmt.Sprintf("AR reduction - %s", payment.Code),
			DebitAmount:  0,
			CreditAmount: payment.Amount,
		},
	}

	journalEntry.JournalLines = journalLines

	// Create journal entry
	if err := tx.Create(journalEntry).Error; err != nil {
		return err
	}

	// Update account balances based on journal lines
	for _, line := range journalLines {
		if err := s.accountRepo.UpdateBalance(context.Background(), line.AccountID, line.DebitAmount, line.CreditAmount); err != nil {
			log.Printf("Warning: Failed to update balance for account %d: %v", line.AccountID, err)
			// Don't fail the entire transaction for balance updates
		}
	}

	return nil
}

// createPayablePaymentJournal creates journal entries for payable payment
func (s *PaymentService) createPayablePaymentJournal(tx *gorm.DB, payment *models.Payment, cashBankID uint, userID uint) error {
	// Get accounts with optimized queries
	var cashBankAccountID uint
	if cashBankID > 0 {
		var cashBank models.CashBank
		if err := tx.Select("account_id").First(&cashBank, cashBankID).Error; err != nil {
			return fmt.Errorf("cash/bank account not found: %v", err)
		}
		cashBankAccountID = cashBank.AccountID
	} else {
		// Get default cash account (Kas - 1101)
		var kasAccount models.Account
		if err := tx.Select("id").Where("code = ?", "1101").First(&kasAccount).Error; err != nil {
			return fmt.Errorf("default cash account (1101) not found: %v", err)
		}
		cashBankAccountID = kasAccount.ID
	}

	// Get Accounts Payable account (Hutang Usaha - 2101) with optimized query
	var apAccount models.Account
	if err := tx.Select("id").Where("code = ?", "2101").First(&apAccount).Error; err != nil {
		log.Printf("Warning: Hutang Usaha account (2101) not found, trying fallback")
		// Fallback: try to find by name pattern
		if err := tx.Select("id").Where("LOWER(name) LIKE ?", "%hutang%usaha%").First(&apAccount).Error; err != nil {
			return fmt.Errorf("accounts payable account not found: %v", err)
		}
	}
	apAccountID := apAccount.ID

	// Create journal entry
	journalEntry := &models.JournalEntry{
		// Code will be auto-generated by BeforeCreate hook
		EntryDate:     payment.Date,
		Description:   fmt.Sprintf("Vendor Payment %s", payment.Code),
		ReferenceType: models.JournalRefPayment,
		ReferenceID:   &payment.ID,
		Reference:     payment.Code,
		UserID:        userID,
		Status:        models.JournalStatusPosted,
		TotalDebit:    payment.Amount,
		TotalCredit:   payment.Amount,
		IsAutoGenerated: true,
	}

	// Journal lines
	journalLines := []models.JournalLine{
		// Debit: Accounts Payable
		{
			AccountID:    apAccountID,
			Description:  fmt.Sprintf("AP reduction - %s", payment.Code),
			DebitAmount:  payment.Amount,
			CreditAmount: 0,
		},
		// Credit: Cash/Bank
		{
			AccountID:    cashBankAccountID,
			Description:  fmt.Sprintf("Payment made - %s", payment.Code),
			DebitAmount:  0,
			CreditAmount: payment.Amount,
		},
	}

	journalEntry.JournalLines = journalLines

	// Create journal entry
	if err := tx.Create(journalEntry).Error; err != nil {
		return err
	}

	// Update account balances based on journal lines (async to improve performance)
	go func() {
		log.Printf("Starting async balance updates for %d journal lines", len(journalLines))
		for i, line := range journalLines {
			if err := s.accountRepo.UpdateBalance(context.Background(), line.AccountID, line.DebitAmount, line.CreditAmount); err != nil {
				log.Printf("Warning: Failed to update balance for account %d (line %d): %v", line.AccountID, i, err)
			} else {
				log.Printf("Balance updated for account %d: Debit=%.2f, Credit=%.2f", line.AccountID, line.DebitAmount, line.CreditAmount)
			}
		}
		log.Printf("Async balance updates completed")
	}()

	return nil
}

// GetPayments retrieves payments with filters
func (s *PaymentService) GetPayments(filter repositories.PaymentFilter) (*repositories.PaymentResult, error) {
	return s.paymentRepo.FindWithFilter(filter)
}

// GetPaymentByID retrieves payment by ID
func (s *PaymentService) GetPaymentByID(id uint) (*models.Payment, error) {
	return s.paymentRepo.FindByID(id)
}

// DeletePayment deletes a payment (admin only)
func (s *PaymentService) DeletePayment(id uint, reason string, userID uint) error {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get payment to verify it exists
	var payment models.Payment
	if err := tx.First(&payment, id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("payment not found: %v", err)
	}

	// Check if payment is already failed/cancelled
	if payment.Status == models.PaymentStatusFailed {
		// If already failed, just soft delete
		if err := tx.Delete(&payment).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete payment: %v", err)
		}
		log.Printf("Deleted failed payment %d (no reversal needed)", id)
	} else {
		// If payment is completed, we need to reverse it first
		log.Printf("Canceling payment %d before deletion", id)
		if err := s.cancelPaymentTransaction(tx, &payment, fmt.Sprintf("Deleted by admin: %s", reason), userID); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to cancel payment before deletion: %v", err)
		}
		
		// Now soft delete the payment
		if err := tx.Delete(&payment).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete payment after cancellation: %v", err)
		}
		log.Printf("Payment %d canceled and deleted successfully", id)
	}

	return tx.Commit().Error
}

// CancelPayment cancels a payment and reverses entries
func (s *PaymentService) CancelPayment(id uint, reason string, userID uint) error {
	tx := s.db.Begin()
	
	var payment models.Payment
	if err := tx.First(&payment, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	if payment.Status == models.PaymentStatusFailed {
		tx.Rollback()
		return errors.New("payment is already cancelled")
	}
	
	// Use helper method to cancel payment
	if err := s.cancelPaymentTransaction(tx, &payment, reason, userID); err != nil {
		tx.Rollback()
		return err
	}
	
	return tx.Commit().Error
}

// cancelPaymentTransaction handles the cancellation logic (reusable helper)
func (s *PaymentService) cancelPaymentTransaction(tx *gorm.DB, payment *models.Payment, reason string, userID uint) error {
	// Reverse allocations
	var allocations []models.PaymentAllocation
	tx.Where("payment_id = ?", payment.ID).Find(&allocations)
	
	for _, allocation := range allocations {
		if allocation.InvoiceID != nil && *allocation.InvoiceID > 0 {
			// Reverse invoice payment
			var sale models.Sale
			if err := tx.First(&sale, allocation.InvoiceID).Error; err == nil {
				sale.PaidAmount -= allocation.AllocatedAmount
				sale.OutstandingAmount += allocation.AllocatedAmount
				
				if sale.Status == models.SaleStatusPaid {
					sale.Status = models.SaleStatusInvoiced
				}
				
				tx.Save(&sale)
			}
		}
		
		if allocation.BillID != nil && *allocation.BillID > 0 {
			// Reverse bill payment - would need proper tracking
			// This is simplified
		}
	}
	
	// Reverse cash/bank transaction
	var cashBankTx models.CashBankTransaction
	if err := tx.Where("reference_type = ? AND reference_id = ?", "PAYMENT", payment.ID).First(&cashBankTx).Error; err == nil {
		var cashBank models.CashBank
		if err := tx.First(&cashBank, cashBankTx.CashBankID).Error; err == nil {
			// Reverse the balance change
			cashBank.Balance -= cashBankTx.Amount
			tx.Save(&cashBank)
		}
	}
	
	// Create reversal journal entries
	if err := s.createReversalJournal(tx, payment, reason, userID); err != nil {
		return err
	}
	
	// Update payment status
	payment.Status = models.PaymentStatusFailed
	payment.Notes += fmt.Sprintf("\nCancelled on %s. Reason: %s", time.Now().Format("2006-01-02"), reason)
	
	return tx.Save(payment).Error
}

// createReversalJournal creates reversal journal entries
func (s *PaymentService) createReversalJournal(tx *gorm.DB, payment *models.Payment, reason string, userID uint) error {
	// Find original journal entry
	var originalJournalEntry models.JournalEntry
	if err := tx.Where("reference_type = ? AND reference_id = ?", models.JournalRefPayment, payment.ID).First(&originalJournalEntry).Error; err != nil {
		return err
	}
	
	// Get original journal lines
	var originalLines []models.JournalLine
	if err := tx.Where("journal_entry_id = ?", originalJournalEntry.ID).Find(&originalLines).Error; err != nil {
		return err
	}
	
	// Create reversal journal entry
	reversalEntry := &models.JournalEntry{
		// Code will be auto-generated by BeforeCreate hook
		EntryDate:     time.Now(),
		Description:   fmt.Sprintf("Reversal of %s - %s", payment.Code, reason),
		ReferenceType: models.JournalRefPayment,
		ReferenceID:   &payment.ID,
		Reference:     fmt.Sprintf("REV-%s", payment.Code),
		UserID:        userID,
		Status:        models.JournalStatusPosted,
		TotalDebit:    originalJournalEntry.TotalCredit,  // Swap totals
		TotalCredit:   originalJournalEntry.TotalDebit,
		ReversedID:    &originalJournalEntry.ID,
		IsAutoGenerated: true,
	}
	
	// Create the journal entry first
	if err := tx.Create(reversalEntry).Error; err != nil {
		return err
	}
	
	// Create reversed journal lines
	for i, original := range originalLines {
		reversalLine := models.JournalLine{
			JournalEntryID: reversalEntry.ID,
			AccountID:      original.AccountID,
			Description:    fmt.Sprintf("Reversal - %s", original.Description),
			DebitAmount:    original.CreditAmount, // Swap debit and credit
			CreditAmount:   original.DebitAmount,
			LineNumber:     i + 1,
		}
		if err := tx.Create(&reversalLine).Error; err != nil {
			return err
		}
		
		// Update account balance for reversal
		if err := s.accountRepo.UpdateBalance(context.Background(), reversalLine.AccountID, reversalLine.DebitAmount, reversalLine.CreditAmount); err != nil {
			log.Printf("Warning: Failed to update balance for reversal account %d: %v", reversalLine.AccountID, err)
		}
	}
	
	// Update original entry to mark as reversed
	originalJournalEntry.ReversalID = &reversalEntry.ID
	originalJournalEntry.Status = models.JournalStatusReversed
	if err := tx.Save(&originalJournalEntry).Error; err != nil {
		return err
	}
	
	return nil
}

// Helper functions
func (s *PaymentService) generatePaymentCode(prefix string) string {
	year := time.Now().Year()
	month := time.Now().Month()
	return s.generatePaymentCodeAtomic(prefix, year, int(month))
}

// generatePaymentCodeAtomic generates payment code using atomic database operations
func (s *PaymentService) generatePaymentCodeAtomic(prefix string, year, month int) string {
	// Use atomic UPSERT operation to get next sequence number
	sequenceNum, err := s.getNextSequenceNumber(prefix, year, month)
	if err != nil {
		// Fallback to timestamp-based unique code if sequence fails
		now := time.Now()
		uniqueNum := now.UnixNano() % 999999
		log.Printf("Warning: Failed to generate sequence for %s/%d/%d, using timestamp fallback: %v", prefix, year, month, err)
		return fmt.Sprintf("%s/%04d/%02d/TS%06d", prefix, year, month, uniqueNum)
	}
	
	return fmt.Sprintf("%s/%04d/%02d/%04d", prefix, year, month, sequenceNum)
}

// getNextSequenceNumber atomically gets the next sequence number for payment codes
func (s *PaymentService) getNextSequenceNumber(prefix string, year, month int) (int, error) {
	// Try using database sequence table with atomic operations
	var sequenceRecord models.PaymentCodeSequence
	
	// Use a transaction to ensure atomicity
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Try to find existing record with row lock
	err := tx.Where("prefix = ? AND year = ? AND month = ?", prefix, year, month).
		Set("gorm:query_option", "FOR UPDATE").
		First(&sequenceRecord).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new sequence record
			sequenceRecord = models.PaymentCodeSequence{
				Prefix:         prefix,
				Year:           year,
				Month:          month,
				SequenceNumber: 1,
			}
			
			if err := tx.Create(&sequenceRecord).Error; err != nil {
				tx.Rollback()
				return 0, err
			}
			
			if err := tx.Commit().Error; err != nil {
				return 0, err
			}
			
			return 1, nil
		} else {
			tx.Rollback()
			return 0, err
		}
	}
	
	// Increment sequence number
	nextNum := sequenceRecord.SequenceNumber + 1
	sequenceRecord.SequenceNumber = nextNum
	
	if err := tx.Save(&sequenceRecord).Error; err != nil {
		tx.Rollback()
		return 0, err
	}
	
	if err := tx.Commit().Error; err != nil {
		return 0, err
	}
	
	return nextNum, nil
}

// checkPaymentCodeExists checks if a payment code already exists
func (s *PaymentService) checkPaymentCodeExists(code string) bool {
	var count int64
	s.db.Model(&models.Payment{}).Where("code = ?", code).Count(&count)
	return count > 0
}

// generateJournalCode is no longer used - journal codes are auto-generated by the JournalEntry BeforeCreate hook

// DTOs
type PaymentCreateRequest struct {
	ContactID       uint                     `json:"contact_id" binding:"required"`
	CashBankID      uint                     `json:"cash_bank_id"`
	Date            time.Time                `json:"date" binding:"required"`
	Amount          float64                  `json:"amount" binding:"required,min=0"`
	Method          string                   `json:"method" binding:"required"`
	Reference       string                   `json:"reference"`
	Notes           string                   `json:"notes"`
	Allocations     []InvoiceAllocation      `json:"allocations"`
	BillAllocations []BillAllocation         `json:"bill_allocations"`
}

type InvoiceAllocation struct {
	InvoiceID uint    `json:"invoice_id"`
	Amount    float64 `json:"amount"`
}

type BillAllocation struct {
	BillID uint    `json:"bill_id"`
	Amount float64 `json:"amount"`
}

// PaymentAllocation is defined in repositories package

// GetUnpaidInvoices gets outstanding invoices for a customer
func (s *PaymentService) GetUnpaidInvoices(customerID uint) ([]OutstandingInvoice, error) {
	// Get sales from sales repository where customer_id = customerID and outstanding_amount > 0
	var sales []models.Sale
	err := s.db.Where("customer_id = ? AND outstanding_amount > ?", customerID, 0).Find(&sales).Error
	if err != nil {
		return nil, err
	}
	
	// Convert to OutstandingInvoice format
	var invoices []OutstandingInvoice
	for _, sale := range sales {
		invoice := OutstandingInvoice{
			ID:               sale.ID,
			Code:             sale.Code,
			Date:             sale.Date.Format("2006-01-02"),
			TotalAmount:      sale.TotalAmount,
			OutstandingAmount: sale.OutstandingAmount,
		}
		
		// Add due date if available (sales usually don't have due date, but we can calculate it)
		// For now, we'll use a 30-day payment term from invoice date
		dueDate := sale.Date.AddDate(0, 0, 30).Format("2006-01-02")
		invoice.DueDate = &dueDate
		
		invoices = append(invoices, invoice)
	}
	
	return invoices, nil
}

// GetUnpaidBills gets outstanding bills for a vendor
func (s *PaymentService) GetUnpaidBills(vendorID uint) ([]OutstandingBill, error) {
	// Get purchases from purchase repository where vendor_id = vendorID and outstanding_amount > 0
	var purchases []models.Purchase
	err := s.db.Where("vendor_id = ? AND status IN (?, ?)", vendorID, "APPROVED", "RECEIVED").Find(&purchases).Error
	if err != nil {
		return nil, err
	}
	
	// Convert to OutstandingBill format
	var bills []OutstandingBill
	for _, purchase := range purchases {
		bill := OutstandingBill{
			ID:               purchase.ID,
			Code:             purchase.Code,
			Date:             purchase.Date.Format("2006-01-02"),
			TotalAmount:      purchase.TotalAmount,
			OutstandingAmount: purchase.TotalAmount, // For now, assume full amount is outstanding
		}
		
		// Add due date if available
		// For purchases, we can use the DueDate field
		dueDate := purchase.DueDate.Format("2006-01-02")
		bill.DueDate = &dueDate
		
		bills = append(bills, bill)
	}
	
	return bills, nil
}

// GetSaleByID gets sale details for payment integration
func (s *PaymentService) GetSaleByID(saleID uint) (*models.Sale, error) {
	return s.salesRepo.FindByID(saleID)
}

// Outstanding item types
type OutstandingInvoice struct {
	ID               uint    `json:"id"`
	Code             string  `json:"code"`
	Date             string  `json:"date"`
	TotalAmount      float64 `json:"total_amount"`
	OutstandingAmount float64 `json:"outstanding_amount"`
	DueDate          *string `json:"due_date,omitempty"`
}

type OutstandingBill struct {
	ID               uint    `json:"id"`
	Code             string  `json:"code"`
	Date             string  `json:"date"`
	TotalAmount      float64 `json:"total_amount"`
	OutstandingAmount float64 `json:"outstanding_amount"`
	DueDate          *string `json:"due_date,omitempty"`
}

// GetPaymentSummary gets payment summary statistics
func (s *PaymentService) GetPaymentSummary(startDate, endDate string) (*repositories.PaymentSummary, error) {
	return s.paymentRepo.GetPaymentSummary(startDate, endDate)
}

// GetPaymentAnalytics gets comprehensive payment analytics for dashboard
func (s *PaymentService) GetPaymentAnalytics(startDate, endDate string) (*PaymentAnalytics, error) {
	// Get basic summary first
	summary, err := s.paymentRepo.GetPaymentSummary(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Get recent payments for analytics
	recentPayments, err := s.paymentRepo.GetPaymentsByDateRange(
		time.Now().AddDate(0, 0, -30), // Last 30 days
		time.Now(),
	)
	if err != nil {
		return nil, err
	}

	// Create analytics response
	analytics := &PaymentAnalytics{
		TotalReceived:   summary.TotalReceived,
		TotalPaid:       summary.TotalPaid,
		NetFlow:         summary.NetFlow,
		ReceivedGrowth:  0, // TODO: Calculate growth percentage
		PaidGrowth:      0, // TODO: Calculate growth percentage
		FlowGrowth:      0, // TODO: Calculate growth percentage
		TotalOutstanding: 0, // TODO: Calculate outstanding amount
		ByMethod:        summary.ByMethod,
		DailyTrend:      s.generateDailyTrend(startDate, endDate),
		RecentPayments:  recentPayments,
		AvgPaymentTime:  2.5, // TODO: Calculate actual processing time
		SuccessRate:     95.0, // TODO: Calculate actual success rate
	}

	return analytics, nil
}

// generateDailyTrend generates daily payment trend data
func (s *PaymentService) generateDailyTrend(startDate, endDate string) []DailyTrend {
	// Parse dates
	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	var trends []DailyTrend
	
	// Generate daily data points
	for d := start; d.Before(end) || d.Equal(end); d = d.AddDate(0, 0, 1) {
		// TODO: Get actual daily payment data from database
		// For now, generate mock data
		trends = append(trends, DailyTrend{
			Date:     d.Format("2006-01-02"),
			Received: 0, // TODO: Get actual received amount
			Paid:     0, // TODO: Get actual paid amount
		})
	}

	return trends
}

// PaymentAnalytics struct for analytics response
type PaymentAnalytics struct {
	TotalReceived    float64            `json:"total_received"`
	TotalPaid        float64            `json:"total_paid"`
	NetFlow          float64            `json:"net_flow"`
	ReceivedGrowth   float64            `json:"received_growth"`
	PaidGrowth       float64            `json:"paid_growth"`
	FlowGrowth       float64            `json:"flow_growth"`
	TotalOutstanding float64            `json:"total_outstanding"`
	ByMethod         map[string]float64 `json:"by_method"`
	DailyTrend       []DailyTrend       `json:"daily_trend"`
	RecentPayments   []models.Payment   `json:"recent_payments"`
	AvgPaymentTime   float64            `json:"avg_payment_time"`
	SuccessRate      float64            `json:"success_rate"`
}

// DailyTrend represents daily payment trend data
type DailyTrend struct {
	Date     string  `json:"date"`
	Received float64 `json:"received"`
	Paid     float64 `json:"paid"`
}

// PaymentFilter and PaymentResult are defined in repositories package

// Export functions

// ExportPaymentReportExcel generates an Excel report for payments
func (s *PaymentService) ExportPaymentReportExcel(startDate, endDate, status, method string) ([]byte, string, error) {
	// Create filter for payments
	filter := repositories.PaymentFilter{
		Status: status,
		Method: method,
		Page:   1,
		Limit:  10000, // Get all payments for report
	}

	// Parse dates if provided
	if startDate != "" {
		if sd, err := time.Parse("2006-01-02", startDate); err == nil {
			filter.StartDate = sd
		}
	}
	if endDate != "" {
		if ed, err := time.Parse("2006-01-02", endDate); err == nil {
			filter.EndDate = ed
		}
	}

	// Get payments data
	result, err := s.paymentRepo.FindWithFilter(filter)
	if err != nil {
		return nil, "", err
	}

	// Generate Excel using existing export service
	excelData, err := s.generatePaymentExcel(result.Data, startDate, endDate, status, method)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Payment_Report_%s_to_%s.xlsx", startDate, endDate)
	if startDate == "" {
		filename = "Payment_Report_All_Time.xlsx"
	}

	return excelData, filename, nil
}

// generatePaymentExcel creates Excel file for payments
func (s *PaymentService) generatePaymentExcel(payments []models.Payment, startDate, endDate, status, method string) ([]byte, error) {
	// Import excelize
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			// Log error
		}
	}()

	sheetName := "Payment Report"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to create Excel sheet: %v", err)
	}

	// Set active sheet
	f.SetActiveSheet(index)

	// Set title
	f.SetCellValue(sheetName, "A1", "PAYMENT REPORT")
	f.SetCellValue(sheetName, "A2", fmt.Sprintf("Generated on: %s", time.Now().Format("2006-01-02 15:04:05")))
	
	// Add filter information
	row := 3
	if startDate != "" && endDate != "" {
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), fmt.Sprintf("Period: %s to %s", startDate, endDate))
		row++
	}
	if status != "" {
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), fmt.Sprintf("Status Filter: %s", status))
		row++
	}
	if method != "" {
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), fmt.Sprintf("Method Filter: %s", method))
		row++
	}
	
	// Headers row
	headerRow := row + 1
	headers := []string{"Date", "Payment Code", "Contact", "Contact Type", "Amount", "Method", "Status", "Reference", "Notes", "Created At"}
	for i, header := range headers {
		cell := string(rune('A'+i)) + strconv.Itoa(headerRow)
		f.SetCellValue(sheetName, cell, header)
	}

	// Style for headers
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 12,
			Color: "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4472C4"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create header style: %v", err)
	}

	// Apply style to headers
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(headerRow), "J"+strconv.Itoa(headerRow), headerStyle)

	// Data style
	dataStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create data style: %v", err)
	}

	// Currency style
	currencyStyle, err := f.NewStyle(&excelize.Style{
		NumFmt: 4, // Currency format
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create currency style: %v", err)
	}

	// Fill data
	totalAmount := 0.0
	completedCount := 0
	pendingCount := 0
	failedCount := 0
	
	for i, payment := range payments {
		dataRow := headerRow + 1 + i
		
		contactName := "N/A"
		contactType := "N/A"
		if payment.Contact.ID != 0 {
			contactName = payment.Contact.Name
			contactType = payment.Contact.Type
		}

		// Set cell values
		f.SetCellValue(sheetName, "A"+strconv.Itoa(dataRow), payment.Date.Format("2006-01-02"))
		f.SetCellValue(sheetName, "B"+strconv.Itoa(dataRow), payment.Code)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(dataRow), contactName)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(dataRow), contactType)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(dataRow), payment.Amount)
		f.SetCellValue(sheetName, "F"+strconv.Itoa(dataRow), payment.Method)
		f.SetCellValue(sheetName, "G"+strconv.Itoa(dataRow), payment.Status)
		f.SetCellValue(sheetName, "H"+strconv.Itoa(dataRow), payment.Reference)
		f.SetCellValue(sheetName, "I"+strconv.Itoa(dataRow), payment.Notes)
		f.SetCellValue(sheetName, "J"+strconv.Itoa(dataRow), payment.CreatedAt.Format("2006-01-02 15:04:05"))

		// Apply styles
		f.SetCellStyle(sheetName, "A"+strconv.Itoa(dataRow), "D"+strconv.Itoa(dataRow), dataStyle)
		f.SetCellStyle(sheetName, "E"+strconv.Itoa(dataRow), "E"+strconv.Itoa(dataRow), currencyStyle)
		f.SetCellStyle(sheetName, "F"+strconv.Itoa(dataRow), "J"+strconv.Itoa(dataRow), dataStyle)
		
		// Accumulate statistics
		totalAmount += payment.Amount
		switch payment.Status {
		case "COMPLETED":
			completedCount++
		case "PENDING":
			pendingCount++
		case "FAILED":
			failedCount++
		}
	}

	// Summary section
	summaryRow := headerRow + len(payments) + 3
	f.SetCellValue(sheetName, "A"+strconv.Itoa(summaryRow), "SUMMARY")
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(summaryRow), "A"+strconv.Itoa(summaryRow), headerStyle)
	
	summaryRow++
	f.SetCellValue(sheetName, "A"+strconv.Itoa(summaryRow), "Total Payments:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(summaryRow), len(payments))
	
	summaryRow++
	f.SetCellValue(sheetName, "A"+strconv.Itoa(summaryRow), "Total Amount:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(summaryRow), totalAmount)
	f.SetCellStyle(sheetName, "B"+strconv.Itoa(summaryRow), "B"+strconv.Itoa(summaryRow), currencyStyle)
	
	summaryRow++
	f.SetCellValue(sheetName, "A"+strconv.Itoa(summaryRow), "Completed:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(summaryRow), completedCount)
	
	summaryRow++
	f.SetCellValue(sheetName, "A"+strconv.Itoa(summaryRow), "Pending:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(summaryRow), pendingCount)
	
	summaryRow++
	f.SetCellValue(sheetName, "A"+strconv.Itoa(summaryRow), "Failed:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(summaryRow), failedCount)
	
	if len(payments) > 0 {
		summaryRow++
		f.SetCellValue(sheetName, "A"+strconv.Itoa(summaryRow), "Average Amount:")
		f.SetCellValue(sheetName, "B"+strconv.Itoa(summaryRow), totalAmount/float64(len(payments)))
		f.SetCellStyle(sheetName, "B"+strconv.Itoa(summaryRow), "B"+strconv.Itoa(summaryRow), currencyStyle)
	}

	// Auto-fit columns
	cols := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
	for _, col := range cols {
		f.SetColWidth(sheetName, col, col, 15)
	}
	
	// Make specific columns wider
	f.SetColWidth(sheetName, "C", "C", 25) // Contact name
	f.SetColWidth(sheetName, "H", "H", 20) // Reference
	f.SetColWidth(sheetName, "I", "I", 30) // Notes
	f.SetColWidth(sheetName, "J", "J", 20) // Created at

	// Delete default Sheet1 if it exists
	if f.GetSheetName(0) == "Sheet1" {
		f.DeleteSheet("Sheet1")
	}

	// Save to buffer
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("failed to write Excel file: %v", err)
	}

	return buf.Bytes(), nil
}

// ExportPaymentReportPDF generates a PDF report for payments
func (s *PaymentService) ExportPaymentReportPDF(startDate, endDate, status, method string) ([]byte, string, error) {
	// Create filter for payments
	filter := repositories.PaymentFilter{
		Status: status,
		Method: method,
		Page:   1,
		Limit:  1000, // Get all payments for report
	}

	// Parse dates if provided
	if startDate != "" {
		if sd, err := time.Parse("2006-01-02", startDate); err == nil {
			filter.StartDate = sd
		}
	}
	if endDate != "" {
		if ed, err := time.Parse("2006-01-02", endDate); err == nil {
			filter.EndDate = ed
		}
	}

	// Get payments data
	result, err := s.paymentRepo.FindWithFilter(filter)
	if err != nil {
		return nil, "", err
	}

	// Generate PDF using existing PDF service
	pdfService := NewPDFService(s.db)
	pdfData, err := pdfService.GeneratePaymentReportPDF(result.Data, startDate, endDate)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Payment_Report_%s_to_%s.pdf", startDate, endDate)
	if startDate == "" {
		filename = "Payment_Report_All_Time.pdf"
	}

	return pdfData, filename, nil
}

// ExportPaymentDetailPDF generates a PDF for a single payment detail
func (s *PaymentService) ExportPaymentDetailPDF(paymentID uint) ([]byte, string, error) {
	// Get payment details
	payment, err := s.paymentRepo.FindByID(paymentID)
	if err != nil {
		return nil, "", err
	}

	// Generate PDF using existing PDF service
	pdfService := NewPDFService(s.db)
	pdfData, err := pdfService.GeneratePaymentDetailPDF(payment)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Payment_%s.pdf", payment.Code)
	return pdfData, filename, nil
}

// updateCashBankBalanceWithLogging - Version with detailed logging
func (s *PaymentService) updateCashBankBalanceWithLogging(tx *gorm.DB, cashBankID uint, amount float64, direction string, referenceID uint, userID uint) error {
	log.Printf("💰 Updating Cash/Bank Balance: ID=%d, Amount=%.2f, Direction=%s", cashBankID, amount, direction)
	
	var cashBank models.CashBank
	if err := tx.First(&cashBank, cashBankID).Error; err != nil {
		return fmt.Errorf("cash/bank account not found: %v", err)
	}
	
	log.Printf("💰 Current balance: %.2f -> %.2f", cashBank.Balance, cashBank.Balance + amount)
	
	// Update balance
	newBalance := cashBank.Balance + amount
	
	// Safety check - only prevent negative balance for outgoing payments (withdrawals)
	// For incoming payments (receivables), allow negative balance to become positive
	if newBalance < 0 && amount < 0 {
		// Only block if this is a withdrawal/payment OUT that would make balance negative
		return fmt.Errorf("insufficient balance for withdrawal. Current: %.2f, Required: %.2f, Shortfall: %.2f", 
			cashBank.Balance, -amount, -newBalance)
	}
	
	cashBank.Balance = newBalance
	
	if err := tx.Save(&cashBank).Error; err != nil {
		return fmt.Errorf("failed to save cash/bank balance: %v", err)
	}
	
	log.Printf("✅ Balance updated successfully: %.2f", cashBank.Balance)
	
	// Create transaction record
	transaction := &models.CashBankTransaction{
		CashBankID:      cashBankID,
		ReferenceType:   "PAYMENT",
		ReferenceID:     referenceID,
		Amount:          amount,
		BalanceAfter:    cashBank.Balance,
		TransactionDate: time.Now(),
		Notes:           fmt.Sprintf("Payment %s", direction),
	}
	
	if err := tx.Create(transaction).Error; err != nil {
		return fmt.Errorf("failed to create cash/bank transaction: %v", err)
	}
	
	log.Printf("✅ Cash/bank transaction recorded")
	return nil
}

// createReceivablePaymentJournalWithLogging - Version with detailed logging
func (s *PaymentService) createReceivablePaymentJournalWithLogging(tx *gorm.DB, payment *models.Payment, cashBankID uint, userID uint) error {
	log.Printf("📋 Creating journal entries for payment %d", payment.ID)
	
	// Get accounts
	var cashBankAccountID uint
	if cashBankID > 0 {
		var cashBank models.CashBank
		if err := tx.First(&cashBank, cashBankID).Error; err != nil {
			return fmt.Errorf("cash/bank account not found: %v", err)
		}
		cashBankAccountID = cashBank.AccountID
		log.Printf("📋 Using Cash/Bank Account ID: %d", cashBankAccountID)
	} else {
		var kasAccount models.Account
		if err := tx.Where("code = ?", "1101").First(&kasAccount).Error; err != nil {
			return fmt.Errorf("default cash account (1101) not found: %v", err)
		}
		cashBankAccountID = kasAccount.ID
		log.Printf("📋 Using default Cash Account ID: %d", cashBankAccountID)
	}
	
	// Get AR account
	var arAccount models.Account
	if err := tx.Where("code = ?", "1201").First(&arAccount).Error; err != nil {
		log.Printf("⚠️ AR account (1201) not found, trying fallback")
		if err := tx.Where("LOWER(name) LIKE ?", "%piutang%usaha%").First(&arAccount).Error; err != nil {
			return fmt.Errorf("accounts receivable account not found: %v", err)
		}
	}
	log.Printf("📋 Using AR Account ID: %d (Code: %s)", arAccount.ID, arAccount.Code)
	
	// Create journal entry
	journalEntry := &models.JournalEntry{
		EntryDate:       payment.Date,
		Description:     fmt.Sprintf("Customer Payment %s", payment.Code),
		ReferenceType:   models.JournalRefPayment,
		ReferenceID:     &payment.ID,
		Reference:       payment.Code,
		UserID:          userID,
		Status:          models.JournalStatusPosted,
		TotalDebit:      payment.Amount,
		TotalCredit:     payment.Amount,
		IsAutoGenerated: true,
	}
	
	// Journal lines
	journalLines := []models.JournalLine{
		{
			AccountID:    cashBankAccountID,
			Description:  fmt.Sprintf("Payment received - %s", payment.Code),
			DebitAmount:  payment.Amount,
			CreditAmount: 0,
		},
		{
			AccountID:    arAccount.ID,
			Description:  fmt.Sprintf("AR reduction - %s", payment.Code),
			DebitAmount:  0,
			CreditAmount: payment.Amount,
		},
	}
	
	journalEntry.JournalLines = journalLines
	
	if err := tx.Create(journalEntry).Error; err != nil {
		return fmt.Errorf("failed to create journal entry: %v", err)
	}
	
	log.Printf("✅ Journal entry created: ID=%d", journalEntry.ID)
	
	// Update account balances using batch operation with extended context
	log.Printf("📋 Updating account balances for %d accounts...", len(journalLines))
	balanceStart := time.Now()
	
	// Create separate context with extended timeout for balance updates
	balanceCtx, cancelBalance := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelBalance()
	
	// Batch update account balances using raw SQL for better performance
	for _, line := range journalLines {
		// Use optimized raw update within the existing transaction
		var balanceChange float64
		if line.DebitAmount > 0 {
			balanceChange = line.DebitAmount
		} else {
			balanceChange = -line.CreditAmount
		}
		
		// Direct SQL update with shorter timeout
		result := tx.WithContext(balanceCtx).Exec(
			`UPDATE accounts SET 
				balance = CASE 
					WHEN type IN ('ASSET', 'EXPENSE') THEN balance + ?
					ELSE balance - ?
				END,
				updated_at = CURRENT_TIMESTAMP
			 WHERE id = ? AND deleted_at IS NULL`,
			balanceChange, balanceChange, line.AccountID)
			
		if result.Error != nil {
			log.Printf("⚠️ Warning: Failed to update balance for account %d: %v", line.AccountID, result.Error)
			// Continue with other updates instead of failing
		} else {
			log.Printf("✅ Updated balance for account %d (%.2f)", line.AccountID, balanceChange)
		}
	}
	
	log.Printf("✅ Account balances updated (%.2fms)", float64(time.Since(balanceStart).Nanoseconds())/1000000)
	
	return nil
}
