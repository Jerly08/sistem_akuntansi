package services

import (
	"errors"
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SalesPaymentService handles payment operations with proper concurrency control
type SalesPaymentService struct {
	db              *gorm.DB
	salesRepo       *repositories.SalesRepository
	paymentService  *PaymentService
}

func NewSalesPaymentService(db *gorm.DB, salesRepo *repositories.SalesRepository, paymentService *PaymentService) *SalesPaymentService {
	return &SalesPaymentService{
		db:             db,
		salesRepo:      salesRepo,
		paymentService: paymentService,
	}
}

// CreateSalePaymentWithLock creates a payment with proper database locking to prevent race conditions
func (s *SalesPaymentService) CreateSalePaymentWithLock(saleID uint, request models.SalePaymentRequest, userID uint) (*models.SalePayment, error) {
	var payment *models.SalePayment
	
	// Use database transaction with proper locking
	err := s.db.Transaction(func(tx *gorm.DB) error {
		log.Printf("🔒 Acquiring lock for sale %d payment creation", saleID)
		
		// Step 1: Lock the sale record to prevent concurrent modifications
		var sale models.Sale
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Customer").
			First(&sale, saleID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("sale with ID %d not found", saleID)
			}
			return fmt.Errorf("failed to lock sale record: %v", err)
		}
		
		log.Printf("🔒 Sale %d locked successfully. Current outstanding: %.2f", saleID, sale.OutstandingAmount)
		
		// Step 2: Validate sale status
		if err := s.validateSaleForPayment(&sale); err != nil {
			return err
		}
		
		// Step 3: Validate payment amount against current outstanding amount
		if request.Amount <= 0 {
			return errors.New("payment amount must be greater than 0")
		}
		
		if request.Amount > sale.OutstandingAmount {
			return fmt.Errorf("payment amount %.2f exceeds outstanding amount %.2f", 
				request.Amount, sale.OutstandingAmount)
		}
		
		log.Printf("💰 Payment validation passed. Amount: %.2f, Outstanding: %.2f", 
			request.Amount, sale.OutstandingAmount)
		
		// Step 4: Create payment record
		payment = &models.SalePayment{
			SaleID:        saleID,
			Amount:        request.Amount,
			PaymentDate:   request.PaymentDate,
			PaymentMethod: request.PaymentMethod,
			Reference:     request.Reference,
			Notes:         request.Notes,
			CashBankID:    request.CashBankID,
			AccountID:     request.AccountID,
			UserID:        userID,
			Status:        "COMPLETED",
			CreatedAt:     time.Now(),
		}
		
		if err := tx.Create(payment).Error; err != nil {
			return fmt.Errorf("failed to create payment record: %v", err)
		}
		
		log.Printf("💳 Payment record created with ID: %d", payment.ID)
		
		// Step 5: Update sale amounts atomically using SQL expressions
		newPaidAmount := sale.PaidAmount + request.Amount
		newOutstandingAmount := sale.OutstandingAmount - request.Amount
		
		updateData := map[string]interface{}{
			"paid_amount":        newPaidAmount,
			"outstanding_amount": newOutstandingAmount,
			"updated_at":        time.Now(),
		}
		
		// Step 6: Determine new status based on payment
		newStatus := s.calculateSaleStatus(newOutstandingAmount, sale.TotalAmount, sale.DueDate)
		if newStatus != sale.Status {
			updateData["status"] = newStatus
			log.Printf("📊 Status changing from %s to %s", sale.Status, newStatus)
		}
		
		if err := tx.Model(&sale).Updates(updateData).Error; err != nil {
			return fmt.Errorf("failed to update sale amounts: %v", err)
		}
		
		log.Printf("✅ Sale %d updated successfully. New paid: %.2f, New outstanding: %.2f, Status: %s", 
			saleID, newPaidAmount, newOutstandingAmount, newStatus)
		
		// Step 7: Journal entries are now handled by PaymentService.CreateReceivablePayment()
		// to avoid double journal entries. This comment documents the architecture.
		log.Printf("💫 Journal entries handled by PaymentService integration")
		
		return nil
	})
	
	if err != nil {
		log.Printf("❌ Payment creation failed for sale %d: %v", saleID, err)
		return nil, err
	}
	
	// Return the created payment with related data
	var createdPayment models.SalePayment
	if err := s.db.Preload("Sale").
		Preload("CashBank").
		Preload("Account").
		Preload("User").
		First(&createdPayment, payment.ID).Error; err != nil {
		log.Printf("⚠️ Payment created but failed to load with relations: %v", err)
		return payment, nil
	}
	
	log.Printf("🎉 Payment creation completed successfully for sale %d", saleID)
	return &createdPayment, nil
}

// validateSaleForPayment validates if a sale can receive payments
func (s *SalesPaymentService) validateSaleForPayment(sale *models.Sale) error {
	// Check if sale status allows payments
	allowedStatuses := []string{
		models.SaleStatusInvoiced,
		models.SaleStatusOverdue,
		models.SaleStatusConfirmed, // Allow payments for confirmed sales
	}
	
	statusAllowed := false
	for _, status := range allowedStatuses {
		if sale.Status == status {
			statusAllowed = true
			break
		}
	}
	
	if !statusAllowed {
		return fmt.Errorf("sale with status '%s' cannot receive payments. Allowed statuses: %v", 
			sale.Status, allowedStatuses)
	}
	
	// Check if there's any outstanding amount
	if sale.OutstandingAmount <= 0 {
		return fmt.Errorf("sale has no outstanding amount (%.2f)", sale.OutstandingAmount)
	}
	
	return nil
}

// calculateSaleStatus determines the new sale status based on payment amounts
func (s *SalesPaymentService) calculateSaleStatus(outstandingAmount, totalAmount float64, dueDate time.Time) string {
	const tolerance = 0.01 // 1 cent tolerance for floating point comparison
	
	// Fully paid
	if outstandingAmount <= tolerance {
		return models.SaleStatusPaid
	}
	
	// Partially paid but overdue
	if time.Now().After(dueDate) && outstandingAmount > tolerance {
		return models.SaleStatusOverdue
	}
	
	// Partially paid but not overdue
	if outstandingAmount < totalAmount-tolerance {
		return models.SaleStatusInvoiced // Keep as invoiced until fully paid or overdue
	}
	
	// Default - keep current status if no change needed
	return models.SaleStatusInvoiced
}

// ValidatePaymentRequest validates payment request data
func (s *SalesPaymentService) ValidatePaymentRequest(request models.SalePaymentRequest) error {
	if request.Amount <= 0 {
		return errors.New("payment amount must be greater than 0")
	}
	
	if request.PaymentDate.IsZero() {
		return errors.New("payment date is required")
	}
	
	if request.PaymentMethod == "" {
		return errors.New("payment method is required")
	}
	
	// Validate payment date is not in future
	if request.PaymentDate.After(time.Now().AddDate(0, 0, 1)) {
		return errors.New("payment date cannot be more than 1 day in the future")
	}
	
	// Validate payment method
	validMethods := []string{"CASH", "BANK_TRANSFER", "CREDIT_CARD", "CHECK", "OTHER"}
	methodValid := false
	for _, method := range validMethods {
		if request.PaymentMethod == method {
			methodValid = true
			break
		}
	}
	
	if !methodValid {
		return fmt.Errorf("invalid payment method '%s'. Valid methods: %v", 
			request.PaymentMethod, validMethods)
	}
	
	return nil
}

// GetSalePaymentSummary returns payment summary for a sale
func (s *SalesPaymentService) GetSalePaymentSummary(saleID uint) (*models.SalePaymentSummary, error) {
	var sale models.Sale
	if err := s.db.Preload("SalePayments").First(&sale, saleID).Error; err != nil {
		return nil, err
	}
	
	summary := &models.SalePaymentSummary{
		SaleID:         sale.ID,
		TotalAmount:    sale.TotalAmount,
		PaidAmount:     sale.PaidAmount,
		OutstandingAmount: sale.OutstandingAmount,
		PaymentCount:   len(sale.SalePayments),
		LastPaymentDate: nil,
	}
	
	// Find last payment date
	for _, payment := range sale.SalePayments {
		if summary.LastPaymentDate == nil || payment.PaymentDate.After(*summary.LastPaymentDate) {
			summary.LastPaymentDate = &payment.PaymentDate
		}
	}
	
	return summary, nil
}