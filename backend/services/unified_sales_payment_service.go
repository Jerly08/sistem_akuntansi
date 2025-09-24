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

// UnifiedSalesPaymentService - SINGLE SOURCE OF TRUTH untuk semua operasi pembayaran sales
// Menggabungkan semua logika pembayaran dalam satu tempat untuk menghindari konflik
type UnifiedSalesPaymentService struct {
	db               *gorm.DB
	salesRepo        *repositories.SalesRepository
	accountRepo      repositories.AccountRepository
	ssotJournalService *SSOTSalesJournalService  // ✅ SSOT integration for proper journal entries
}

func NewUnifiedSalesPaymentService(db *gorm.DB, salesRepo *repositories.SalesRepository, accountRepo repositories.AccountRepository) *UnifiedSalesPaymentService {
	return &UnifiedSalesPaymentService{
		db:                 db,
		salesRepo:          salesRepo,
		accountRepo:        accountRepo,
		ssotJournalService: NewSSOTSalesJournalService(db),  // ✅ Initialize SSOT journal service
	}
}

// CreateSalesPayment - UNIFIED method untuk membuat pembayaran sales dengan journal entries
func (s *UnifiedSalesPaymentService) CreateSalesPayment(saleID uint, request models.SalePaymentRequest, userID uint) (*models.SalePayment, error) {
	log.Printf("🚀 [UNIFIED] Starting payment creation for sale %d, amount: %.2f", saleID, request.Amount)
	
	var payment *models.SalePayment
	
	// Use single database transaction for ALL operations
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Step 1: Lock sale record to prevent race conditions
		log.Printf("🔒 [UNIFIED] Locking sale %d", saleID)
		var sale models.Sale
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Customer").
			First(&sale, saleID).Error; err != nil {
			return fmt.Errorf("sale not found or could not be locked: %v", err)
		}
		
		log.Printf("📊 [UNIFIED] Sale locked: Status=%s, Total=%.2f, Paid=%.2f, Outstanding=%.2f", 
			sale.Status, sale.TotalAmount, sale.PaidAmount, sale.OutstandingAmount)
		
		// Step 2: Validate sale status
		if err := s.validateSaleForPayment(&sale); err != nil {
			return err
		}
		
		// Step 3: Validate payment amount
		if err := s.validatePaymentAmount(request.Amount, sale.OutstandingAmount); err != nil {
			return err
		}
		
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
		
		log.Printf("💳 [UNIFIED] Payment record created: ID=%d, Amount=%.2f", payment.ID, payment.Amount)
		
		// Step 5: Update sale amounts atomically
		newPaidAmount := sale.PaidAmount + request.Amount
		newOutstandingAmount := sale.OutstandingAmount - request.Amount
		newStatus := s.calculateSaleStatus(newOutstandingAmount, sale.TotalAmount, sale.DueDate)
		
		updateData := map[string]interface{}{
			"paid_amount":        newPaidAmount,
			"outstanding_amount": newOutstandingAmount,
			"status":            newStatus,
			"updated_at":        time.Now(),
		}
		
		if err := tx.Model(&sale).Updates(updateData).Error; err != nil {
			return fmt.Errorf("failed to update sale amounts: %v", err)
		}
		
		log.Printf("📈 [UNIFIED] Sale updated: Paid=%.2f->%.2f, Outstanding=%.2f->%.2f, Status=%s->%s", 
			sale.PaidAmount, newPaidAmount, sale.OutstandingAmount, newOutstandingAmount, sale.Status, newStatus)
		
		// Step 6: Create SSOT journal entries (FIXED to use proper system)
		log.Printf("📝 [UNIFIED] Creating SSOT journal entries for payment")
		if _, err := s.ssotJournalService.CreatePaymentJournalEntry(payment, userID); err != nil {
			return fmt.Errorf("failed to create SSOT journal entries: %v", err)
		}
		
		// Step 7: Update cash/bank balance (still needed for CashBank table)
		if request.CashBankID != nil && *request.CashBankID > 0 {
			if err := s.updateCashBankBalance(tx, *request.CashBankID, request.Amount, userID); err != nil {
				return fmt.Errorf("failed to update cash/bank balance: %v", err)
			}
		}
		
		log.Printf("✅ [UNIFIED] Payment transaction completed successfully")
		return nil
	})
	
	if err != nil {
		log.Printf("❌ [UNIFIED] Payment creation failed: %v", err)
		return nil, err
	}
	
	// Return payment with preloaded relations
	var completedPayment models.SalePayment
	if err := s.db.Preload("Sale").Preload("Sale.Customer").
		Preload("CashBank").Preload("User").
		First(&completedPayment, payment.ID).Error; err != nil {
		log.Printf("⚠️ [UNIFIED] Payment created but failed to load relations: %v", err)
		return payment, nil
	}
	
	log.Printf("🎉 [UNIFIED] Payment creation completed: ID=%d, Amount=%.2f", completedPayment.ID, completedPayment.Amount)
	return &completedPayment, nil
}

// validateSaleForPayment validates if sale can receive payments
func (s *UnifiedSalesPaymentService) validateSaleForPayment(sale *models.Sale) error {
	allowedStatuses := []string{
		models.SaleStatusInvoiced,
		models.SaleStatusOverdue,
		models.SaleStatusConfirmed,
	}
	
	statusAllowed := false
	for _, status := range allowedStatuses {
		if sale.Status == status {
			statusAllowed = true
			break
		}
	}
	
	if !statusAllowed {
		return fmt.Errorf("sale status '%s' cannot receive payments. Allowed: %v", sale.Status, allowedStatuses)
	}
	
	if sale.OutstandingAmount <= 0 {
		return fmt.Errorf("sale has no outstanding amount (%.2f)", sale.OutstandingAmount)
	}
	
	return nil
}

// validatePaymentAmount validates payment amount
func (s *UnifiedSalesPaymentService) validatePaymentAmount(amount, outstanding float64) error {
	if amount <= 0 {
		return errors.New("payment amount must be greater than 0")
	}
	
	if amount > outstanding {
		return fmt.Errorf("payment amount %.2f exceeds outstanding %.2f", amount, outstanding)
	}
	
	return nil
}

// calculateSaleStatus determines new sale status after payment
func (s *UnifiedSalesPaymentService) calculateSaleStatus(outstanding, total float64, dueDate time.Time) string {
	const tolerance = 0.01
	
	if outstanding <= tolerance {
		return models.SaleStatusPaid
	}
	
	if time.Now().After(dueDate) && outstanding > tolerance {
		return models.SaleStatusOverdue
	}
	
	if outstanding < total-tolerance {
		return models.SaleStatusInvoiced
	}
	
	return models.SaleStatusInvoiced
}

// createPaymentJournalEntries - DEPRECATED: Now using SSOT journal system
// This method has been replaced by SSOTSalesJournalService.CreatePaymentJournalEntry()
// Left here for reference but should not be used
func (s *UnifiedSalesPaymentService) createPaymentJournalEntries_DEPRECATED(tx *gorm.DB, payment *models.SalePayment, sale *models.Sale, userID uint) error {
	log.Printf("⚠️ [DEPRECATED] createPaymentJournalEntries called - this should use SSOT system instead")
	return fmt.Errorf("deprecated method called - use SSOT journal system instead")
}

// updateCashBankBalance updates cash/bank balance
func (s *UnifiedSalesPaymentService) updateCashBankBalance(tx *gorm.DB, cashBankID uint, amount float64, userID uint) error {
	var cashBank models.CashBank
	if err := tx.First(&cashBank, cashBankID).Error; err != nil {
		return fmt.Errorf("cash/bank account not found: %v", err)
	}
	
	log.Printf("💰 [UNIFIED] Updating %s balance: %.2f + %.2f = %.2f", 
		cashBank.Name, cashBank.Balance, amount, cashBank.Balance+amount)
	
	cashBank.Balance += amount
	
	if err := tx.Save(&cashBank).Error; err != nil {
		return fmt.Errorf("failed to update cash/bank balance: %v", err)
	}
	
	// Create transaction record
	transaction := &models.CashBankTransaction{
		CashBankID:      cashBankID,
		ReferenceType:   "PAYMENT",
		ReferenceID:     0, // Will be set after payment creation
		Amount:          amount,
		BalanceAfter:    cashBank.Balance,
		TransactionDate: time.Now(),
		Notes:           "Sales payment received",
	}
	
	if err := tx.Create(transaction).Error; err != nil {
		return fmt.Errorf("failed to create cash/bank transaction: %v", err)
	}
	
	log.Printf("💰 [UNIFIED] Cash/bank balance updated successfully")
	return nil
}

// updateAccountBalances - DEPRECATED: SSOT journal system handles account balance updates automatically
// This method is no longer needed as the SSOT system automatically updates account balances
// when journal entries are posted with AutoPost=true
func (s *UnifiedSalesPaymentService) updateAccountBalances_DEPRECATED(tx *gorm.DB, lines []models.JournalLine) error {
	log.Printf("⚠️ [DEPRECATED] updateAccountBalances called - SSOT system handles this automatically")
	return fmt.Errorf("deprecated method called - SSOT system handles account balance updates automatically")
}

// GetSalePayments returns all payments for a sale
func (s *UnifiedSalesPaymentService) GetSalePayments(saleID uint) ([]models.SalePayment, error) {
	return s.salesRepo.FindPaymentsBySaleID(saleID)
}

// GetPaymentSummary returns payment summary for a sale
func (s *UnifiedSalesPaymentService) GetPaymentSummary(saleID uint) (*models.SalePaymentSummary, error) {
	var sale models.Sale
	if err := s.db.Preload("SalePayments").First(&sale, saleID).Error; err != nil {
		return nil, err
	}
	
	summary := &models.SalePaymentSummary{
		SaleID:            sale.ID,
		TotalAmount:       sale.TotalAmount,
		PaidAmount:        sale.PaidAmount,
		OutstandingAmount: sale.OutstandingAmount,
		PaymentCount:      len(sale.SalePayments),
	}
	
	// Find last payment date
	for _, payment := range sale.SalePayments {
		if summary.LastPaymentDate == nil || payment.PaymentDate.After(*summary.LastPaymentDate) {
			summary.LastPaymentDate = &payment.PaymentDate
		}
	}
	
	return summary, nil
}

// ValidatePaymentRequest validates payment request
func (s *UnifiedSalesPaymentService) ValidatePaymentRequest(request models.SalePaymentRequest) error {
	if request.Amount <= 0 {
		return errors.New("payment amount must be greater than 0")
	}
	
	if request.PaymentDate.IsZero() {
		return errors.New("payment date is required")
	}
	
	if request.PaymentMethod == "" {
		return errors.New("payment method is required")
	}
	
	if request.PaymentDate.After(time.Now().AddDate(0, 0, 1)) {
		return errors.New("payment date cannot be more than 1 day in the future")
	}
	
	validMethods := []string{"CASH", "BANK_TRANSFER", "CREDIT_CARD", "CHECK", "OTHER"}
	methodValid := false
	for _, method := range validMethods {
		if request.PaymentMethod == method {
			methodValid = true
			break
		}
	}
	
	if !methodValid {
		return fmt.Errorf("invalid payment method '%s'. Valid methods: %v", request.PaymentMethod, validMethods)
	}
	
	return nil
}