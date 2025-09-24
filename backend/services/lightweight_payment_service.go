package services

import (
	"fmt"
	"log"
	"time"
	"context"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// LightweightPaymentService provides fast, minimal payment recording
type LightweightPaymentService struct {
	db           *gorm.DB
	paymentRepo  repositories.PaymentRepository
	salesRepo    repositories.SalesRepository
	contactRepo  repositories.ContactRepository
	cashBankRepo repositories.CashBankRepository
}

// LightweightPaymentRequest minimal request structure for fast processing
type LightweightPaymentRequest struct {
	SaleID       uint    `json:"sale_id" binding:"required"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
	PaymentDate  string  `json:"payment_date" binding:"required"`
	Method       string  `json:"method" binding:"required"`
	CashBankID   uint    `json:"cash_bank_id" binding:"required"`
	Reference    string  `json:"reference"`
	Notes        string  `json:"notes"`
	UserID       uint    `json:"-"` // Set from JWT context
}

// LightweightPaymentResponse minimal response for fast feedback
type LightweightPaymentResponse struct {
	Success       bool    `json:"success"`
	PaymentID     uint    `json:"payment_id"`
	PaymentCode   string  `json:"payment_code"`
	Amount        float64 `json:"amount"`
	NewStatus     string  `json:"new_status"`
	Outstanding   float64 `json:"outstanding_amount"`
	ProcessingTime string `json:"processing_time"`
	Message       string  `json:"message"`
}

// NewLightweightPaymentService creates a new lightweight payment service
func NewLightweightPaymentService(
	db *gorm.DB,
	paymentRepo repositories.PaymentRepository,
	salesRepo repositories.SalesRepository,
	contactRepo repositories.ContactRepository,
	cashBankRepo repositories.CashBankRepository,
) *LightweightPaymentService {
	return &LightweightPaymentService{
		db:           db,
		paymentRepo:  paymentRepo,
		salesRepo:    salesRepo,
		contactRepo:  contactRepo,
		cashBankRepo: cashBankRepo,
	}
}

// RecordPaymentFast processes payment with minimal database operations
func (lps *LightweightPaymentService) RecordPaymentFast(ctx context.Context, req *LightweightPaymentRequest) (*LightweightPaymentResponse, error) {
	startTime := time.Now()
	
	log.Printf("🚀 FAST: Starting lightweight payment recording for sale %d, amount %.2f", req.SaleID, req.Amount)

	// Use a single optimized transaction for all operations
	var response *LightweightPaymentResponse
	var err error

	// Set timeout for fast processing (10 seconds max)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = lps.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: Fetch only required sale data with minimal joins
		var sale models.Sale
		if err := tx.Select("id, code, invoice_number, customer_id, total_amount, paid_amount, outstanding_amount, status").
			Where("id = ?", req.SaleID).First(&sale).Error; err != nil {
			return fmt.Errorf("sale not found: %w", err)
		}

		// Step 2: Basic validation (fast checks only)
		if sale.Status != models.SaleStatusInvoiced && sale.Status != models.SaleStatusOverdue {
			return fmt.Errorf("sale must be invoiced to receive payment")
		}

		if req.Amount > sale.OutstandingAmount {
			return fmt.Errorf("payment amount (%.2f) exceeds outstanding amount (%.2f)", req.Amount, sale.OutstandingAmount)
		}

		// Step 3: Parse payment date quickly
		paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
		if err != nil {
			paymentDate = time.Now() // Fallback to current time
		}

		// Step 4: Generate payment code (fast sequential)
		paymentCode := fmt.Sprintf("RCV-%d-%04d", time.Now().Year(), time.Now().Unix()%10000)

		// Step 5: Create payment record (minimal fields)
		payment := &models.Payment{
			ContactID: sale.CustomerID,
			UserID:    req.UserID,
			Date:      paymentDate,
			Amount:    req.Amount,
			Method:    "RECEIVABLE", // Fixed for sales payments
			Code:      paymentCode,
			Reference: req.Reference,
			Notes:     req.Notes,
			Status:    models.PaymentStatusCompleted, // Set as completed immediately
		}

		if err := tx.Create(payment).Error; err != nil {
			return fmt.Errorf("failed to create payment: %w", err)
		}

		// Step 6: Create payment allocation (minimal)
		allocation := &models.PaymentAllocation{
			PaymentID:       uint64(payment.ID),
			InvoiceID:       &req.SaleID,
			AllocatedAmount: req.Amount,
		}

		if err := tx.Create(allocation).Error; err != nil {
			return fmt.Errorf("failed to create payment allocation: %w", err)
		}

		// Step 7: Update sale amounts (optimized single query)
		newPaidAmount := sale.PaidAmount + req.Amount
		newOutstandingAmount := sale.OutstandingAmount - req.Amount
		newStatus := sale.Status

		// Determine new status based on outstanding amount
		if newOutstandingAmount <= 0.01 { // Consider fully paid if difference < 1 cent
			newStatus = models.SaleStatusPaid
			newOutstandingAmount = 0 // Ensure exact zero
		}

		if err := tx.Model(&sale).Updates(map[string]interface{}{
			"paid_amount":        newPaidAmount,
			"outstanding_amount": newOutstandingAmount,
			"status":            newStatus,
		}).Error; err != nil {
			return fmt.Errorf("failed to update sale: %w", err)
		}

		// Step 8: Update cash/bank balance (optimized single operation)
		if err := tx.Model(&models.CashBank{}).
			Where("id = ?", req.CashBankID).
			UpdateColumn("balance", gorm.Expr("balance + ?", req.Amount)).Error; err != nil {
			return fmt.Errorf("failed to update cash/bank balance: %w", err)
		}

		// Step 9: Create minimal transaction record for audit
		transaction := &models.CashBankTransaction{
			CashBankID:      req.CashBankID,
			ReferenceType:   "PAYMENT",
			ReferenceID:     payment.ID,
			Amount:          req.Amount,
			TransactionDate: paymentDate,
			Notes:           fmt.Sprintf("Payment %s", paymentCode),
		}

		if err := tx.Create(transaction).Error; err != nil {
			log.Printf("Warning: Failed to create transaction record: %v", err)
			// Don't fail the whole transaction for audit record
		}

		// Build response
		response = &LightweightPaymentResponse{
			Success:        true,
			PaymentID:      payment.ID,
			PaymentCode:    paymentCode,
			Amount:         req.Amount,
			NewStatus:      newStatus,
			Outstanding:    newOutstandingAmount,
			ProcessingTime: time.Since(startTime).String(),
			Message:        "Payment recorded successfully",
		}

		return nil
	})

	if err != nil {
		log.Printf("❌ FAST: Payment recording failed: %v", err)
		return &LightweightPaymentResponse{
			Success:        false,
			ProcessingTime: time.Since(startTime).String(),
			Message:        fmt.Sprintf("Payment recording failed: %v", err),
		}, err
	}

	log.Printf("✅ FAST: Payment recorded successfully in %s", response.ProcessingTime)
	return response, nil
}

// RecordPaymentWithAsyncJournal records payment immediately and processes journal entry asynchronously
func (lps *LightweightPaymentService) RecordPaymentWithAsyncJournal(ctx context.Context, req *LightweightPaymentRequest) (*LightweightPaymentResponse, error) {
	// First, record payment with lightweight method
	response, err := lps.RecordPaymentFast(ctx, req)
	if err != nil {
		return response, err
	}

	// Queue journal entry creation asynchronously (non-blocking)
	go func() {
		asyncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Printf("🔄 ASYNC: Creating journal entry for payment %d", response.PaymentID)
		
		// Create journal entry in background
		if err := lps.createJournalEntryAsync(asyncCtx, response.PaymentID); err != nil {
			log.Printf("⚠️ ASYNC: Failed to create journal entry for payment %d: %v", response.PaymentID, err)
			// Could implement retry mechanism or dead letter queue here
		} else {
			log.Printf("✅ ASYNC: Journal entry created for payment %d", response.PaymentID)
		}
	}()

	return response, nil
}

// createJournalEntryAsync creates journal entry in background
func (lps *LightweightPaymentService) createJournalEntryAsync(ctx context.Context, paymentID uint) error {
	// Get payment details
	var payment models.Payment
	if err := lps.db.WithContext(ctx).
		Preload("Contact").
		First(&payment, paymentID).Error; err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}

	// Initialize SSOT journal service
	journalService := NewUnifiedJournalService(lps.db)

	// Get account IDs for journal entry
	var cashAccount models.Account
	if err := lps.db.WithContext(ctx).
		Where("code = ?", "1101").First(&cashAccount).Error; err != nil {
		return fmt.Errorf("cash account not found: %w", err)
	}

	var arAccount models.Account
	if err := lps.db.WithContext(ctx).
		Where("code = ?", "1201").First(&arAccount).Error; err != nil {
		return fmt.Errorf("accounts receivable account not found: %w", err)
	}

	// Create journal entry
	journalLines := []JournalLineRequest{
		{
			AccountID:    uint64(cashAccount.ID),
			Description:  fmt.Sprintf("Payment received - %s", payment.Code),
			DebitAmount:  decimal.NewFromFloat(payment.Amount),
			CreditAmount: decimal.Zero,
		},
		{
			AccountID:    uint64(arAccount.ID),
			Description:  fmt.Sprintf("AR reduction - %s", payment.Code),
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromFloat(payment.Amount),
		},
	}

	paymentIDUint64 := uint64(payment.ID)
	journalRequest := &JournalEntryRequest{
		SourceType:  models.SSOTSourceTypePayment,
		SourceID:    &paymentIDUint64,
		Reference:   payment.Code,
		EntryDate:   payment.Date,
		Description: fmt.Sprintf("Customer Payment %s", payment.Code),
		Lines:       journalLines,
		AutoPost:    true,
		CreatedBy:   uint64(payment.UserID),
	}

	_, err := journalService.CreateJournalEntry(journalRequest)
	if err != nil {
		return fmt.Errorf("failed to create journal entry: %w", err)
	}

	// Update payment with journal entry reference
	// This is optional and can be done later if needed

	return nil
}

// ValidatePaymentRequest performs fast validation
func (lps *LightweightPaymentService) ValidatePaymentRequest(req *LightweightPaymentRequest) error {
	if req.SaleID == 0 {
		return fmt.Errorf("sale ID is required")
	}
	
	if req.Amount <= 0 {
		return fmt.Errorf("amount must be greater than zero")
	}
	
	if req.Method == "" {
		return fmt.Errorf("payment method is required")
	}
	
	if req.CashBankID == 0 {
		return fmt.Errorf("cash/bank account is required")
	}
	
	if req.PaymentDate == "" {
		return fmt.Errorf("payment date is required")
	}

	// Validate date format
	if _, err := time.Parse("2006-01-02", req.PaymentDate); err != nil {
		return fmt.Errorf("invalid payment date format, use YYYY-MM-DD")
	}
	
	return nil
}