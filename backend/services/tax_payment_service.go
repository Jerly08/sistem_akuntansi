package services

import (
	"fmt"
	"log"
	"time"

	"app-sistem-akuntansi/models"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// TaxPaymentService handles tax-related payment operations (PPN Input/Output payments)
type TaxPaymentService struct {
	db                 *gorm.DB
	taxAccountService  *TaxAccountService
	journalService     *UnifiedJournalService
}

// NewTaxPaymentService creates a new TaxPaymentService instance
func NewTaxPaymentService(db *gorm.DB) *TaxPaymentService {
	return &TaxPaymentService{
		db:                db,
		taxAccountService: NewTaxAccountService(db),
		journalService:    NewUnifiedJournalService(db),
	}
}

// CreatePPNPaymentRequest represents a request to pay PPN
type CreatePPNPaymentRequest struct {
	PPNType     string    `json:"ppn_type" binding:"required"` // INPUT (Masukan) or OUTPUT (Keluaran)
	Amount      float64   `json:"amount" binding:"required"`
	Date        time.Time `json:"date" binding:"required"`
	CashBankID  uint      `json:"cash_bank_id" binding:"required"`
	Reference   string    `json:"reference"`
	Notes       string    `json:"notes"`
}

// CreatePPNPayment creates a payment for PPN (either input or output)
// Logic: Debit PPN Account (mengurangi hutang/piutang PPN), Credit Cash/Bank (kas berkurang)
func (s *TaxPaymentService) CreatePPNPayment(req CreatePPNPaymentRequest, userID uint) (*models.Payment, error) {
	log.Printf("🏦 Starting PPN Payment: Type=%s, Amount=%.2f", req.PPNType, req.Amount)

	// Validate PPN type
	if req.PPNType != "INPUT" && req.PPNType != "OUTPUT" {
		return nil, fmt.Errorf("ppn_type must be INPUT (Masukan) or OUTPUT (Keluaran)")
	}

	// Validate amount
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}

	// Validate date
	if req.Date.IsZero() {
		return nil, fmt.Errorf("payment date is required")
	}

	// Begin transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", tx.Error)
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("❌ PANIC in CreatePPNPayment: %v", r)
		}
	}()

	// Get Cash/Bank account
	var cashBank models.CashBank
	if err := tx.Preload("Account").First(&cashBank, req.CashBankID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("cash/bank account not found: %v", err)
	}
	log.Printf("📋 Cash/Bank Account: %s (ID: %d)", cashBank.Account.Name, cashBank.AccountID)

	// Get PPN account from settings
	settings, err := s.taxAccountService.GetSettings()
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to get tax settings: %v", err)
	}

	var taxAccountID uint
	if req.PPNType == "INPUT" {
		// PPN Masukan (Purchase VAT) - account 1240
		if settings.PurchaseInputVATAccountID == 0 {
			tx.Rollback()
			return nil, fmt.Errorf("PPN Masukan account not configured in tax settings")
		}
		taxAccountID = settings.PurchaseInputVATAccountID
	} else {
		// PPN Keluaran (Sales VAT) - account 2103
		if settings.SalesOutputVATAccountID == 0 {
			tx.Rollback()
			return nil, fmt.Errorf("PPN Keluaran account not configured in tax settings")
		}
		taxAccountID = settings.SalesOutputVATAccountID
	}

	// Validate tax account exists
	var taxAccount models.Account
	if err := tx.First(&taxAccount, taxAccountID).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("tax account not found: %v", err)
	}
	log.Printf("📋 Tax Account: %s - %s (ID: %d)", taxAccount.Code, taxAccount.Name, taxAccount.ID)

	// Generate payment code
	prefix := "PPN"
	if req.PPNType == "INPUT" {
		prefix = "PPNM" // PPN Masukan
	} else {
		prefix = "PPNK" // PPN Keluaran
	}
	code := s.generatePaymentCode(prefix, tx)

	// Use default system contact (Tax Office)
	contactID := uint(1)

	// Create payment record
	payment := &models.Payment{
		Code:        code,
		ContactID:   contactID,
		UserID:      userID,
		Date:        req.Date,
		Amount:      req.Amount,
		Method:      models.PaymentMethodBankTransfer, // Default untuk pembayaran PPN
		Reference:   req.Reference,
		Status:      models.PaymentStatusCompleted, // PPN payment langsung completed
		Notes:       req.Notes,
		PaymentType: models.PaymentTypeTaxPPN,
	}

	if err := tx.Create(payment).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create payment record: %v", err)
	}
	log.Printf("✅ Payment record created: ID=%d, Code=%s", payment.ID, payment.Code)

	// Create journal entry
	// Logika PPN Payment:
	// - PPN Masukan (Asset): Debit mengurangi piutang PPN
	// - PPN Keluaran (Liability): Debit mengurangi hutang PPN  
	// - Cash/Bank: Credit karena kas keluar
	ppnLabel := "PPN Masukan"
	if req.PPNType == "OUTPUT" {
		ppnLabel = "PPN Keluaran"
	}
	
	journalLines := []JournalLineRequest{
		{
			AccountID:    uint64(taxAccountID),
			Description:  fmt.Sprintf("Pembayaran %s - %s", ppnLabel, payment.Code),
			DebitAmount:  decimal.NewFromFloat(req.Amount),
			CreditAmount: decimal.Zero,
		},
		{
			AccountID:    uint64(cashBank.AccountID),
			Description:  fmt.Sprintf("Pembayaran %s - %s", ppnLabel, payment.Code),
			DebitAmount:  decimal.Zero,
			CreditAmount: decimal.NewFromFloat(req.Amount),
		},
	}

	// Create SSOT journal entry
	journalRequest := &JournalEntryRequest{
		SourceType:  models.SSOTSourceTypePayment,
		SourceID:    uint64(payment.ID),
		Reference:   payment.Code,
		EntryDate:   payment.Date,
		Description: fmt.Sprintf("Pembayaran %s %s", ppnLabel, payment.Code),
		Lines:       journalLines,
		AutoPost:    true,
		CreatedBy:   uint64(userID),
	}

	journalService := NewUnifiedJournalService(tx)
	journalResponse, err := journalService.CreateJournalEntryWithTx(tx, journalRequest)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create journal entry: %v", err)
	}

	log.Printf("✅ Journal entry created: ID=%d, EntryNumber=%s", journalResponse.ID, journalResponse.EntryNumber)

	// Update payment with journal reference
	journalEntryID := uint(journalResponse.ID)
	payment.JournalEntryID = &journalEntryID
	if err := tx.Save(payment).Error; err != nil {
		log.Printf("⚠️ Warning: Failed to update payment with journal reference: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("✅ PPN Payment successfully created: %s", payment.Code)

	// Load relations for response
	if err := s.db.Preload("Contact").Preload("User").First(payment, payment.ID).Error; err != nil {
		log.Printf("⚠️ Warning: Failed to load payment relations: %v", err)
	}

	return payment, nil
}

// generatePaymentCode generates unique payment code with prefix
func (s *TaxPaymentService) generatePaymentCode(prefix string, tx *gorm.DB) string {
	now := time.Now()
	yearMonth := now.Format("0601") // YYMM format

	// Get last sequence for this month
	var lastPayment models.Payment
	pattern := fmt.Sprintf("%s-%s-%%", prefix, yearMonth)
	
	err := tx.Where("code LIKE ?", pattern).Order("code DESC").First(&lastPayment).Error
	
	sequence := 1
	if err == nil {
		// Extract sequence from last code
		var lastSeq int
		fmt.Sscanf(lastPayment.Code, prefix+"-%s-%d", &yearMonth, &lastSeq)
		sequence = lastSeq + 1
	}

	return fmt.Sprintf("%s-%s-%04d", prefix, yearMonth, sequence)
}

// GetPPNPaymentsByType retrieves PPN payments filtered by type
func (s *TaxPaymentService) GetPPNPaymentsByType(paymentType string, startDate, endDate time.Time) ([]models.Payment, error) {
	var payments []models.Payment
	
	query := s.db.Preload("Contact").Preload("User").Preload("TaxAccount").Preload("CashBank").
		Where("payment_type = ?", paymentType)
	
	if !startDate.IsZero() {
		query = query.Where("date >= ?", startDate)
	}
	
	if !endDate.IsZero() {
		query = query.Where("date <= ?", endDate)
	}
	
	if err := query.Order("date DESC, created_at DESC").Find(&payments).Error; err != nil {
		return nil, fmt.Errorf("failed to get PPN payments: %v", err)
	}
	
	return payments, nil
}

// GetPPNPaymentSummary returns summary of PPN payments
func (s *TaxPaymentService) GetPPNPaymentSummary(startDate, endDate time.Time) (map[string]interface{}, error) {
	type SummaryResult struct {
		PaymentType string
		TotalAmount float64
		Count       int64
	}
	
	var results []SummaryResult
	
	query := s.db.Model(&models.Payment{}).
		Select("payment_type, SUM(amount) as total_amount, COUNT(*) as count").
		Where("payment_type IN ?", []string{models.PaymentTypeTaxPPNInput, models.PaymentTypeTaxPPNOutput})
	
	if !startDate.IsZero() {
		query = query.Where("date >= ?", startDate)
	}
	
	if !endDate.IsZero() {
		query = query.Where("date <= ?", endDate)
	}
	
	if err := query.Group("payment_type").Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to get PPN payment summary: %v", err)
	}
	
	summary := map[string]interface{}{
		"ppn_masukan": map[string]interface{}{
			"total_amount": 0.0,
			"count":        0,
		},
		"ppn_keluaran": map[string]interface{}{
			"total_amount": 0.0,
			"count":        0,
		},
	}
	
	for _, result := range results {
		if result.PaymentType == models.PaymentTypeTaxPPNInput {
			summary["ppn_masukan"] = map[string]interface{}{
				"total_amount": result.TotalAmount,
				"count":        result.Count,
			}
		} else if result.PaymentType == models.PaymentTypeTaxPPNOutput {
			summary["ppn_keluaran"] = map[string]interface{}{
				"total_amount": result.TotalAmount,
				"count":        result.Count,
			}
		}
	}
	
	return summary, nil
}
