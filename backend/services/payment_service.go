package services

import (
	"bytes"
	"errors"
	"fmt"
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

// CreateReceivablePayment creates payment for sales/receivables
func (s *PaymentService) CreateReceivablePayment(request PaymentCreateRequest, userID uint) (*models.Payment, error) {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Validate customer
	_, err := s.contactRepo.GetByID(request.ContactID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("customer not found")
	}
	
	// Generate payment code
	code := s.generatePaymentCode("RCV")
	
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
	
	// Process allocations to invoices
	remainingAmount := request.Amount
	for _, allocation := range request.Allocations {
		if remainingAmount <= 0 {
			break
		}
		
		sale, err := s.salesRepo.FindByID(allocation.InvoiceID)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("invoice %d not found", allocation.InvoiceID)
		}
		
		if sale.CustomerID != request.ContactID {
			tx.Rollback()
			return nil, errors.New("invoice does not belong to this customer")
		}
		
		allocatedAmount := allocation.Amount
		if allocatedAmount > remainingAmount {
			allocatedAmount = remainingAmount
		}
		if allocatedAmount > sale.OutstandingAmount {
			allocatedAmount = sale.OutstandingAmount
		}
		
		// Create payment allocation
		paymentAllocation := &models.PaymentAllocation{
			PaymentID:       payment.ID,
			InvoiceID:       &allocation.InvoiceID,
			AllocatedAmount: allocatedAmount,
		}
		
		if err := tx.Create(paymentAllocation).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		
		// Update invoice
		sale.PaidAmount += allocatedAmount
		sale.OutstandingAmount -= allocatedAmount
		
		if sale.OutstandingAmount <= 0 {
			sale.Status = models.SaleStatusPaid
		}
		
		if err := tx.Save(sale).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		
		remainingAmount -= allocatedAmount
	}
	
	// Update cash/bank account
	if request.CashBankID > 0 {
		err = s.updateCashBankBalance(tx, request.CashBankID, request.Amount, "IN", payment.ID, userID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}
	
	// Create journal entries
	err = s.createReceivablePaymentJournal(tx, payment, request.CashBankID, userID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	
	payment.Status = models.PaymentStatusCompleted
	if err := tx.Save(payment).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	return payment, tx.Commit().Error
}

// CreatePayablePayment creates payment for purchases/payables
func (s *PaymentService) CreatePayablePayment(request PaymentCreateRequest, userID uint) (*models.Payment, error) {
	// Start transaction
	tx := s.db.Begin()
	
	// Validate vendor
	_, err := s.contactRepo.GetByID(request.ContactID)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("vendor not found")
	}
	
	// Check cash/bank balance
	if request.CashBankID > 0 {
		cashBank, err := s.cashBankRepo.FindByID(request.CashBankID)
		if err != nil {
			tx.Rollback()
			return nil, errors.New("cash/bank account not found")
		}
		
		if cashBank.Balance < request.Amount {
			tx.Rollback()
			return nil, fmt.Errorf("insufficient balance. Available: %.2f", cashBank.Balance)
		}
	}
	
	// Generate payment code
	code := s.generatePaymentCode("PAY")
	
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
		return nil, err
	}
	
	return payment, tx.Commit().Error
}

// updateCashBankBalance updates cash/bank balance and creates transaction record
func (s *PaymentService) updateCashBankBalance(tx *gorm.DB, cashBankID uint, amount float64, direction string, referenceID uint, userID uint) error {
	var cashBank models.CashBank
	if err := tx.First(&cashBank, cashBankID).Error; err != nil {
		return err
	}
	
	// Update balance
	cashBank.Balance += amount
	
	if cashBank.Balance < 0 {
		return errors.New("insufficient balance in cash/bank account")
	}
	
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
		// TODO: Implement GetAccountByCode in AccountRepository
		// cashAccount, err := s.accountRepo.GetAccountByCode("1100")
		// if err != nil {
		//	return err
		// }
		// cashBankAccountID = cashAccount.ID
		cashBankAccountID = 1 // Default cash account ID - should be from config or db
	}
	
	// TODO: Implement GetAccountByCode in AccountRepository
	// arAccount, err := s.accountRepo.GetAccountByCode("1200") // Accounts Receivable
	// if err != nil {
	//	return err
	// }
	// Use default AR account ID for now
	arAccountID := uint(2) // Default AR account ID - should be from config or db
	
	// Create journal
	journal := &models.Journal{
		Code:          s.generateJournalCode("RCV"),
		Date:          payment.Date,
		Description:   fmt.Sprintf("Customer Payment %s", payment.Code),
		ReferenceType: models.JournalRefTypePayment,
		ReferenceID:   &payment.ID,
		UserID:        userID,
		Status:        models.JournalStatusPosted,
		Period:        payment.Date.Format("2006-01"),
	}
	
	// Journal entries
	entries := []models.JournalEntry{
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
	
	journal.JournalEntries = entries
	journal.TotalDebit = payment.Amount
	journal.TotalCredit = payment.Amount
	
	return tx.Create(journal).Error
}

// createPayablePaymentJournal creates journal entries for payable payment
func (s *PaymentService) createPayablePaymentJournal(tx *gorm.DB, payment *models.Payment, cashBankID uint, userID uint) error {
	// Get accounts
	var cashBankAccountID uint
	if cashBankID > 0 {
		var cashBank models.CashBank
		if err := tx.First(&cashBank, cashBankID).Error; err != nil {
			return err
		}
		cashBankAccountID = cashBank.AccountID
	} else {
		// TODO: Implement GetAccountByCode in AccountRepository
		// cashAccount, err := s.accountRepo.GetAccountByCode("1100")
		// if err != nil {
		//	return err
		// }
		// cashBankAccountID = cashAccount.ID
		cashBankAccountID = 1 // Default cash account ID - should be from config or db
	}
	
	// TODO: Implement GetAccountByCode in AccountRepository
	// apAccount, err := s.accountRepo.GetAccountByCode("2100") // Accounts Payable
	// if err != nil {
	//	return err
	// }
	// Use default AP account ID for now
	apAccountID := uint(3) // Default AP account ID - should be from config or db
	
	// Create journal
	journal := &models.Journal{
		Code:          s.generateJournalCode("PAY"),
		Date:          payment.Date,
		Description:   fmt.Sprintf("Vendor Payment %s", payment.Code),
		ReferenceType: models.JournalRefTypePayment,
		ReferenceID:   &payment.ID,
		UserID:        userID,
		Status:        models.JournalStatusPosted,
		Period:        payment.Date.Format("2006-01"),
	}
	
	// Journal entries
	entries := []models.JournalEntry{
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
	
	journal.JournalEntries = entries
	journal.TotalDebit = payment.Amount
	journal.TotalCredit = payment.Amount
	
	return tx.Create(journal).Error
}

// GetPayments retrieves payments with filters
func (s *PaymentService) GetPayments(filter repositories.PaymentFilter) (*repositories.PaymentResult, error) {
	return s.paymentRepo.FindWithFilter(filter)
}

// GetPaymentByID retrieves payment by ID
func (s *PaymentService) GetPaymentByID(id uint) (*models.Payment, error) {
	return s.paymentRepo.FindByID(id)
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
	
	// Reverse allocations
	var allocations []models.PaymentAllocation
	tx.Where("payment_id = ?", id).Find(&allocations)
	
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
	if err := tx.Where("reference_type = ? AND reference_id = ?", "PAYMENT", id).First(&cashBankTx).Error; err == nil {
		var cashBank models.CashBank
		if err := tx.First(&cashBank, cashBankTx.CashBankID).Error; err == nil {
			// Reverse the balance change
			cashBank.Balance -= cashBankTx.Amount
			tx.Save(&cashBank)
		}
	}
	
	// Create reversal journal entries
	s.createReversalJournal(tx, &payment, reason, userID)
	
	// Update payment status
	payment.Status = models.PaymentStatusFailed
	payment.Notes += fmt.Sprintf("\nCancelled on %s. Reason: %s", time.Now().Format("2006-01-02"), reason)
	
	if err := tx.Save(&payment).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	return tx.Commit().Error
}

// createReversalJournal creates reversal journal entries
func (s *PaymentService) createReversalJournal(tx *gorm.DB, payment *models.Payment, reason string, userID uint) error {
	// Find original journal
	var originalJournal models.Journal
	if err := tx.Where("reference_type = ? AND reference_id = ?", models.JournalRefTypePayment, payment.ID).First(&originalJournal).Error; err != nil {
		return err
	}
	
	// Create reversal journal
	journal := &models.Journal{
		Code:          s.generateJournalCode("REV"),
		Date:          time.Now(),
		Description:   fmt.Sprintf("Reversal of %s - %s", payment.Code, reason),
		ReferenceType: models.JournalRefTypePayment,
		ReferenceID:   &payment.ID,
		UserID:        userID,
		Status:        models.JournalStatusPosted,
		Period:        time.Now().Format("2006-01"),
		IsAdjusting:   true,
	}
	
	// Reverse entries
	var originalEntries []models.JournalEntry
	tx.Where("journal_id = ?", originalJournal.ID).Find(&originalEntries)
	
	for _, original := range originalEntries {
		entry := models.JournalEntry{
			AccountID:    original.AccountID,
			Description:  fmt.Sprintf("Reversal - %s", original.Description),
			DebitAmount:  original.CreditAmount, // Swap debit and credit
			CreditAmount: original.DebitAmount,
		}
		journal.JournalEntries = append(journal.JournalEntries, entry)
	}
	
	journal.TotalDebit = originalJournal.TotalCredit
	journal.TotalCredit = originalJournal.TotalDebit
	
	return tx.Create(journal).Error
}

// Helper functions
func (s *PaymentService) generatePaymentCode(prefix string) string {
	year := time.Now().Year()
	month := time.Now().Month()
	count, _ := s.paymentRepo.CountByMonth(year, int(month))
	return fmt.Sprintf("%s/%04d/%02d/%04d", prefix, year, month, count+1)
}

func (s *PaymentService) generateJournalCode(prefix string) string {
	year := time.Now().Year()
	month := time.Now().Month()
	// Get count of existing journals for this month
	count, _ := s.paymentRepo.CountJournalsByMonth(year, int(month))
	return fmt.Sprintf("%s-JV/%04d/%02d/%04d", prefix, year, month, count+1)
}

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
	pdfService := NewPDFService()
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
	pdfService := NewPDFService()
	pdfData, err := pdfService.GeneratePaymentDetailPDF(payment)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Payment_%s.pdf", payment.Code)
	return pdfData, filename, nil
}
