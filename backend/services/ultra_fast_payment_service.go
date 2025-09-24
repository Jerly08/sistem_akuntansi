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

// UltraFastPaymentService provides the fastest possible payment recording
// This service skips all non-critical operations for maximum speed
type UltraFastPaymentService struct {
	db *gorm.DB
}

// UltraFastPaymentRequest minimal request for ultra-fast processing
type UltraFastPaymentRequest struct {
	SaleID     uint    `json:"sale_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	CashBankID uint    `json:"cash_bank_id" binding:"required"`
	UserID     uint    `json:"-"` // Set from JWT
}

// UltraFastPaymentResponse minimal response
type UltraFastPaymentResponse struct {
	Success        bool    `json:"success"`
	PaymentID      uint    `json:"payment_id"`
	PaymentCode    string  `json:"payment_code"`
	ProcessingTime string  `json:"processing_time"`
	Message        string  `json:"message"`
}

// NewUltraFastPaymentService creates ultra-fast payment service
func NewUltraFastPaymentService(db *gorm.DB) *UltraFastPaymentService {
	return &UltraFastPaymentService{
		db: db,
	}
}

// RecordPaymentUltraFast processes payment with absolute minimum operations
func (ufps *UltraFastPaymentService) RecordPaymentUltraFast(ctx context.Context, req *UltraFastPaymentRequest) (*UltraFastPaymentResponse, error) {
	startTime := time.Now()
	
	log.Printf("⚡ ULTRA-FAST: Starting payment recording for sale %d, amount %.2f", req.SaleID, req.Amount)

	var response *UltraFastPaymentResponse
	
	// Set ultra-short timeout (5 seconds max)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Single transaction with only essential operations
	err := ufps.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: Minimal sale validation (single query, no joins)
		var saleCheck struct {
			ID                uint    `json:"id"`
			OutstandingAmount float64 `json:"outstanding_amount"`
			CustomerID        uint    `json:"customer_id"`
		}
		
		if err := tx.Raw("SELECT id, outstanding_amount, customer_id FROM sales WHERE id = ? AND status IN ('INVOICED', 'OVERDUE')", req.SaleID).
			Scan(&saleCheck).Error; err != nil {
			return fmt.Errorf("invalid sale: %w", err)
		}

		if saleCheck.ID == 0 {
			return fmt.Errorf("sale not found or not invoiced")
		}

		if req.Amount > saleCheck.OutstandingAmount {
			return fmt.Errorf("amount exceeds outstanding")
		}

		// Step 2: Generate simple payment code
		paymentCode := fmt.Sprintf("FAST-%d-%d", time.Now().Unix()%100000, req.SaleID)

		// Step 3: Create payment record (minimal fields only)
		payment := &models.Payment{
			ContactID: saleCheck.CustomerID,
			UserID:    req.UserID,
			Date:      time.Now(),
			Amount:    req.Amount,
			Method:    "RECEIVABLE",
			Code:      paymentCode,
			Reference: "Ultra fast payment",
			Status:    models.PaymentStatusCompleted,
		}

		if err := tx.Create(payment).Error; err != nil {
			return fmt.Errorf("failed to create payment: %w", err)
		}

		// Step 4: Update sale amounts (raw SQL for speed)
		newOutstandingAmount := saleCheck.OutstandingAmount - req.Amount
		newStatus := "INVOICED"
		
		if newOutstandingAmount <= 0.01 {
			newStatus = "PAID"
			newOutstandingAmount = 0
		}

		if err := tx.Exec(`
			UPDATE sales 
			SET paid_amount = paid_amount + ?, 
			    outstanding_amount = ?, 
			    status = ?
			WHERE id = ?
		`, req.Amount, newOutstandingAmount, newStatus, req.SaleID).Error; err != nil {
			return fmt.Errorf("failed to update sale: %w", err)
		}

		// Step 5: Update cash/bank balance (raw SQL for speed)
		if err := tx.Exec(`
			UPDATE cash_banks 
			SET balance = balance + ?
			WHERE id = ?
		`, req.Amount, req.CashBankID).Error; err != nil {
			return fmt.Errorf("failed to update balance: %w", err)
		}

		// Build response
		response = &UltraFastPaymentResponse{
			Success:        true,
			PaymentID:      payment.ID,
			PaymentCode:    paymentCode,
			ProcessingTime: time.Since(startTime).String(),
			Message:        "Ultra-fast payment recorded",
		}

		return nil
	})

	if err != nil {
		log.Printf("❌ ULTRA-FAST: Payment failed: %v", err)
		return &UltraFastPaymentResponse{
			Success:        false,
			ProcessingTime: time.Since(startTime).String(),
			Message:        fmt.Sprintf("Payment failed: %v", err),
		}, err
	}

	log.Printf("✅ ULTRA-FAST: Payment recorded in %s", response.ProcessingTime)
	return response, nil
}

// CreateJournalEntryAsync creates journal entry in background (non-blocking)
func (ufps *UltraFastPaymentService) CreateJournalEntryAsync(paymentID uint, amount float64) {
	go func() {
		time.Sleep(100 * time.Millisecond) // Small delay to let payment complete
		
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Printf("🔄 ASYNC: Creating journal entry for payment %d", paymentID)
		
		// Get account IDs (hardcoded for speed)
		cashAccountID := uint64(1) // Assume cash account ID is 1
		arAccountID := uint64(2)   // Assume AR account ID is 2
		
		// Try to find actual account IDs
		var cashAcc, arAcc models.Account
		if err := ufps.db.WithContext(ctx).Where("code = ?", "1101").First(&cashAcc).Error; err == nil {
			cashAccountID = uint64(cashAcc.ID)
		}
		if err := ufps.db.WithContext(ctx).Where("code = ?", "1201").First(&arAcc).Error; err == nil {
			arAccountID = uint64(arAcc.ID)
		}

		// Create SSOT journal entry
		journalService := NewUnifiedJournalService(ufps.db)
		
		paymentIDUint64 := uint64(paymentID)
		journalRequest := &JournalEntryRequest{
			SourceType:  models.SSOTSourceTypePayment,
			SourceID:    &paymentIDUint64,
			Reference:   fmt.Sprintf("FAST-%d", paymentID),
			EntryDate:   time.Now(),
			Description: fmt.Sprintf("Ultra Fast Payment %d", paymentID),
			Lines: []JournalLineRequest{
				{
					AccountID:    cashAccountID,
					Description:  "Ultra fast payment received",
					DebitAmount:  decimal.NewFromFloat(amount),
					CreditAmount: decimal.Zero,
				},
				{
					AccountID:    arAccountID,
					Description:  "AR reduction - ultra fast",
					DebitAmount:  decimal.Zero,
					CreditAmount: decimal.NewFromFloat(amount),
				},
			},
			AutoPost:  true,
			CreatedBy: 1, // Default user ID
		}

		if _, err := journalService.CreateJournalEntry(journalRequest); err != nil {
			log.Printf("⚠️ ASYNC: Failed to create journal entry: %v", err)
		} else {
			log.Printf("✅ ASYNC: Journal entry created for payment %d", paymentID)
		}
	}()
}

