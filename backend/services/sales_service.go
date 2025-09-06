package services

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"gorm.io/gorm"
)

type SalesService struct {
	db              *gorm.DB
	salesRepo       *repositories.SalesRepository
	productRepo     *repositories.ProductRepository
	contactRepo     repositories.ContactRepository
	accountRepo     repositories.AccountRepository
	journalService  JournalServiceInterface
	pdfService      PDFServiceInterface
}

// Define interface types to avoid dependency issues
type JournalServiceInterface interface {
	CreateSaleJournalEntries(sale *models.Sale, userID uint) error
	CreatePaymentJournalEntries(payment *models.SalePayment, userID uint) error
	CreateSaleReversalJournalEntries(sale *models.Sale, userID uint, reason string) error
}

type PDFServiceInterface interface {
	GenerateInvoicePDF(sale *models.Sale) ([]byte, error)
	GenerateSalesReportPDF(sales []models.Sale, startDate, endDate string) ([]byte, error)
	GeneratePaymentReportPDF(payments []models.Payment, startDate, endDate string) ([]byte, error)
	GeneratePaymentDetailPDF(payment *models.Payment) ([]byte, error)
}

type SalesResult struct {
	Data       []models.Sale `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

func NewSalesService(db *gorm.DB, salesRepo *repositories.SalesRepository, productRepo *repositories.ProductRepository, contactRepo repositories.ContactRepository, accountRepo repositories.AccountRepository, journalService JournalServiceInterface, pdfService PDFServiceInterface) *SalesService {
	return &SalesService{
		db:              db,
		salesRepo:       salesRepo,
		productRepo:     productRepo,
		contactRepo:     contactRepo,
		accountRepo:     accountRepo,
		journalService:  journalService,
		pdfService:      pdfService,
	}
}

// Sales CRUD Operations

func (s *SalesService) GetSales(filter models.SalesFilter) (*SalesResult, error) {
	sales, total, err := s.salesRepo.FindWithFilter(filter)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))

	return &SalesResult{
		Data:       sales,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *SalesService) GetSaleByID(id uint) (*models.Sale, error) {
	return s.salesRepo.FindByID(id)
}

func (s *SalesService) CreateSale(request models.SaleCreateRequest, userID uint) (*models.Sale, error) {
	// Validate customer exists
	_, err := s.contactRepo.GetByID(request.CustomerID)
	if err != nil {
		return nil, errors.New("customer not found")
	}

	// Validate document type specific requirements
	if err := s.validateDocumentTypeRequirements(request.Type, request.ValidUntil, request.Date); err != nil {
		return nil, err
	}

	// Validate sales person if provided
	if request.SalesPersonID != nil {
		// Check if sales person exists in contacts with type EMPLOYEE
		salesPerson, err := s.contactRepo.GetByID(*request.SalesPersonID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("sales person with ID %d not found", *request.SalesPersonID)
			}
			return nil, fmt.Errorf("error validating sales person: %v", err)
		}
		
		// Validate that the contact is an employee
		if salesPerson.Type != "EMPLOYEE" {
			return nil, fmt.Errorf("contact with ID %d is not an employee", *request.SalesPersonID)
		}
		
		// Check if employee is active
		if !salesPerson.IsActive {
			return nil, fmt.Errorf("sales person with ID %d is inactive", *request.SalesPersonID)
		}
	}

	// Generate sale code and numbers
	code, err := s.generateSaleCode(request.Type)
	if err != nil {
		return nil, err
	}

	// Create sale entity
	sale := &models.Sale{
		Code:            code,
		CustomerID:      request.CustomerID,
		UserID:          userID,
		SalesPersonID:   request.SalesPersonID,
		Type:            request.Type,
		Status:          models.SaleStatusDraft,
		Date:            request.Date,
		DueDate:         request.DueDate,
		ValidUntil:      request.ValidUntil,
		Currency:        s.getDefaultCurrency(request.Currency),
		ExchangeRate:    s.getExchangeRate(request.Currency, s.dereferenceFloat64(request.ExchangeRate)),
		DiscountPercent: request.DiscountPercent,
		PPNPercent:      s.getDefaultPPNPercent(s.dereferenceFloat64(request.PPNPercent)),
		PPhPercent:      request.PPhPercent,
		PPhType:         request.PPhType,
		PaymentTerms:    request.PaymentTerms,
		PaymentMethod:   request.PaymentMethod,
		ShippingMethod:  request.ShippingMethod,
		ShippingCost:    request.ShippingCost,
		ShippingTaxable:  request.ShippingTaxable,
		BillingAddress:  request.BillingAddress,
		ShippingAddress: request.ShippingAddress,
		Notes:           request.Notes,
		InternalNotes:   request.InternalNotes,
		Reference:       request.Reference,
	}

	// Generate specific numbers based on type
	switch request.Type {
	case models.SaleTypeQuotation:
		sale.QuotationNumber = s.generateQuotationNumber()
	case models.SaleTypeOrder:
		// Will be generated when converting from quotation or creating directly
	case models.SaleTypeInvoice:
		sale.InvoiceNumber = s.generateInvoiceNumber()
	}

	// Calculate totals and create sale items
	err = s.calculateSaleItemsFromRequest(sale, request.Items)
	if err != nil {
		return nil, err
	}

	// Save sale with transaction
	createdSale, err := s.salesRepo.Create(sale)
	if err != nil {
		return nil, err
	}

	// Update inventory if needed (for confirmed sales)
	if sale.Status == models.SaleStatusConfirmed {
		err = s.updateInventoryForSale(createdSale)
		if err != nil {
			// Rollback sale creation
			s.salesRepo.Delete(createdSale.ID)
			return nil, err
		}

		// Create journal entries
		err = s.createJournalEntriesForSale(createdSale, userID)
		if err != nil {
			return nil, err
		}
	}

	return s.GetSaleByID(createdSale.ID)
}

func (s *SalesService) UpdateSale(id uint, request models.SaleUpdateRequest, userID uint) (*models.Sale, error) {
	sale, err := s.salesRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Check if sale can be updated (only drafts and pending)
	if sale.Status != models.SaleStatusDraft && sale.Status != models.SaleStatusPending {
		return nil, errors.New("sale cannot be updated in current status")
	}

	// Validate sales person if provided in update
	if request.SalesPersonID != nil {
		// Check if sales person exists in contacts with type EMPLOYEE
		salesPerson, err := s.contactRepo.GetByID(*request.SalesPersonID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("sales person with ID %d not found", *request.SalesPersonID)
			}
			return nil, fmt.Errorf("error validating sales person: %v", err)
		}
		
		// Validate that the contact is an employee
		if salesPerson.Type != "EMPLOYEE" {
			return nil, fmt.Errorf("contact with ID %d is not an employee", *request.SalesPersonID)
		}
		
		// Check if employee is active
		if !salesPerson.IsActive {
			return nil, fmt.Errorf("sales person with ID %d is inactive", *request.SalesPersonID)
		}
	}

	// Update fields if provided
	if request.CustomerID != nil {
		sale.CustomerID = *request.CustomerID
	}
	if request.SalesPersonID != nil {
		sale.SalesPersonID = request.SalesPersonID
	}
	if request.Date != nil {
		sale.Date = *request.Date
	}
	if request.DueDate != nil {
		sale.DueDate = *request.DueDate
	}
	if request.ValidUntil != nil {
		sale.ValidUntil = request.ValidUntil
	}
	if request.DiscountPercent != nil {
		sale.DiscountPercent = *request.DiscountPercent
	}
	if request.PPNPercent != nil {
		sale.PPNPercent = *request.PPNPercent
	}
	if request.PPhPercent != nil {
		sale.PPhPercent = *request.PPhPercent
	}
	if request.PPhType != nil {
		sale.PPhType = *request.PPhType
	}
	if request.PaymentTerms != nil {
		sale.PaymentTerms = *request.PaymentTerms
	}
	if request.PaymentMethod != nil {
		sale.PaymentMethod = *request.PaymentMethod
	}
	if request.ShippingMethod != nil {
		sale.ShippingMethod = *request.ShippingMethod
	}
	if request.ShippingCost != nil {
		sale.ShippingCost = *request.ShippingCost
	}
	if request.ShippingTaxable != nil {
		sale.ShippingTaxable = *request.ShippingTaxable
	}
	if request.BillingAddress != nil {
		sale.BillingAddress = *request.BillingAddress
	}
	if request.ShippingAddress != nil {
		sale.ShippingAddress = *request.ShippingAddress
	}
	if request.Notes != nil {
		sale.Notes = *request.Notes
	}
	if request.InternalNotes != nil {
		sale.InternalNotes = *request.InternalNotes
	}
	if request.Reference != nil {
		sale.Reference = *request.Reference
	}

	// Update items if provided
	if len(request.Items) > 0 {
		err = s.updateSaleItemsFromRequest(sale, request.Items)
		if err != nil {
			return nil, err
		}
	}

	// Recalculate totals
	err = s.recalculateSaleTotals(sale)
	if err != nil {
		return nil, err
	}

	// Save updated sale
	updatedSale, err := s.salesRepo.Update(sale)
	if err != nil {
		return nil, err
	}

	return s.GetSaleByID(updatedSale.ID)
}

func (s *SalesService) DeleteSale(id uint) error {
	sale, err := s.salesRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Check if sale can be deleted (only drafts)
	if sale.Status != models.SaleStatusDraft {
		return errors.New("sale cannot be deleted in current status")
	}

	return s.salesRepo.Delete(id)
}

// DeleteSaleWithRole deletes a sale with role-based permission checking
func (s *SalesService) DeleteSaleWithRole(id uint, userRole string) error {
	sale, err := s.salesRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Admin users can delete sales in any status
	if strings.ToLower(userRole) == "admin" {
		log.Printf("Admin user deleting sale %d with status %s", id, sale.Status)
		
		// If sale was confirmed/invoiced, restore inventory and create reversal entries
		if sale.Status == models.SaleStatusConfirmed || sale.Status == models.SaleStatusInvoiced {
			err = s.restoreInventoryForSale(sale)
			if err != nil {
				log.Printf("Warning: Failed to restore inventory for sale %d: %v", id, err)
			}
			
			// Create reversal journal entries
			err = s.createReversalJournalEntries(sale, 1, "Admin deletion") // Using userID 1 as placeholder
			if err != nil {
				log.Printf("Warning: Failed to create reversal journal entries for sale %d: %v", id, err)
			}
		}
		
		return s.salesRepo.Delete(id)
	}

	// Non-admin users can only delete drafts (existing business rule)
	if sale.Status != models.SaleStatusDraft {
		return errors.New("sale cannot be deleted in current status")
	}

	return s.salesRepo.Delete(id)
}

// Status Management

func (s *SalesService) ConfirmSale(id uint, userID uint) error {
	sale, err := s.salesRepo.FindByID(id)
	if err != nil {
		log.Printf("Error finding sale %d: %v", id, err)
		return fmt.Errorf("failed to find sale: %v", err)
	}

	if sale.Status != models.SaleStatusDraft {
		return errors.New("only draft sales can be confirmed")
	}

	// Update status directly to invoiced with invoice number generation
	sale.Status = models.SaleStatusInvoiced
	sale.InvoiceNumber = s.generateInvoiceNumber()

	// Calculate due date if not set
	if sale.DueDate.IsZero() {
		dueDate := s.calculateDueDate(sale.Date, sale.PaymentTerms)
		sale.DueDate = dueDate
	}

	// Update outstanding amount
	sale.OutstandingAmount = sale.TotalAmount - sale.PaidAmount

	// Update inventory
	err = s.updateInventoryForSale(sale)
	if err != nil {
		log.Printf("Error updating inventory for sale %d: %v", id, err)
		return fmt.Errorf("failed to update inventory: %v", err)
	}

	// Create journal entries for the sale
	err = s.createJournalEntriesForSale(sale, userID)
	if err != nil {
		log.Printf("Error creating journal entries for sale %d: %v", id, err)
		return fmt.Errorf("failed to create journal entries: %v", err)
	}

	// Save the updated sale
	_, err = s.salesRepo.Update(sale)
	if err != nil {
		log.Printf("Error updating sale %d: %v", id, err)
		return fmt.Errorf("failed to update sale: %v", err)
	}

	log.Printf("Successfully confirmed sale %d", id)
	return nil
}

func (s *SalesService) CreateInvoiceFromSale(id uint, userID uint) (*models.Sale, error) {
	sale, err := s.salesRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if sale.Status != models.SaleStatusConfirmed {
		return nil, errors.New("sale must be confirmed before creating invoice")
	}

	// Generate invoice number
	sale.InvoiceNumber = s.generateInvoiceNumber()
	sale.Status = models.SaleStatusInvoiced

	// Calculate due date if not set
	if sale.DueDate.IsZero() {
		dueDate := s.calculateDueDate(sale.Date, sale.PaymentTerms)
		sale.DueDate = dueDate
	}

	// Update outstanding amount
	sale.OutstandingAmount = sale.TotalAmount - sale.PaidAmount

	updatedSale, err := s.salesRepo.Update(sale)
	if err != nil {
		return nil, err
	}

	// Create journal entries for the invoice
	err = s.createJournalEntriesForSale(updatedSale, userID)
	if err != nil {
		return nil, err
	}

	return s.GetSaleByID(updatedSale.ID)
}

func (s *SalesService) CancelSale(id uint, reason string, userID uint) error {
	sale, err := s.salesRepo.FindByID(id)
	if err != nil {
		return err
	}

	if sale.Status == models.SaleStatusPaid || sale.Status == models.SaleStatusCancelled {
		return errors.New("sale cannot be cancelled in current status")
	}

	// Restore inventory if sale was confirmed
	if sale.Status == models.SaleStatusConfirmed || sale.Status == models.SaleStatusInvoiced {
		err = s.restoreInventoryForSale(sale)
		if err != nil {
			return err
		}

		// Create reversal journal entries
		err = s.createReversalJournalEntries(sale, userID, reason)
		if err != nil {
			return err
		}
	}

	sale.Status = models.SaleStatusCancelled
	sale.InternalNotes += fmt.Sprintf("\nCancelled on %s. Reason: %s", time.Now().Format("2006-01-02 15:04"), reason)

	_, err = s.salesRepo.Update(sale)
	return err
}

// Payment Management

func (s *SalesService) GetSalePayments(saleID uint) ([]models.SalePayment, error) {
	return s.salesRepo.FindPaymentsBySaleID(saleID)
}

func (s *SalesService) CreateSalePayment(saleID uint, request models.SalePaymentRequest, userID uint) (*models.SalePayment, error) {
	sale, err := s.salesRepo.FindByID(saleID)
	if err != nil {
		log.Printf("Error finding sale %d: %v", saleID, err)
		return nil, err
	}

	log.Printf("Sale %d status: %s, Outstanding: %.2f, Payment amount: %.2f", saleID, sale.Status, sale.OutstandingAmount, request.Amount)

	if sale.Status != models.SaleStatusInvoiced && sale.Status != models.SaleStatusOverdue {
		log.Printf("Sale status validation failed. Current status: %s, Expected: %s or %s", sale.Status, models.SaleStatusInvoiced, models.SaleStatusOverdue)
		return nil, errors.New(fmt.Sprintf("payments can only be recorded for invoiced sales. Current status: %s", sale.Status))
	}

	if request.Amount > sale.OutstandingAmount {
		log.Printf("Payment amount validation failed. Amount: %.2f, Outstanding: %.2f", request.Amount, sale.OutstandingAmount)
		return nil, errors.New(fmt.Sprintf("payment amount %.2f exceeds outstanding amount %.2f", request.Amount, sale.OutstandingAmount))
	}

	// Create payment record
	payment := &models.SalePayment{
		SaleID:        saleID,
		PaymentNumber: s.generatePaymentNumber(),
		Date:          request.PaymentDate,
		Amount:        request.Amount,
		Method:        request.PaymentMethod,
		Reference:     request.Reference,
		Notes:         request.Notes,
		CashBankID:    request.CashBankID,
		AccountID:     request.AccountID,
		UserID:        userID,
	}

	createdPayment, err := s.salesRepo.CreatePayment(payment)
	if err != nil {
		return nil, err
	}

	// Update sale amounts
	sale.PaidAmount += request.Amount
	sale.OutstandingAmount -= request.Amount

	// Update status if fully paid
	if sale.OutstandingAmount <= 0 {
		sale.Status = models.SaleStatusPaid
	}

	_, err = s.salesRepo.Update(sale)
	if err != nil {
		return nil, err
	}

	// Create journal entries for payment
	err = s.createJournalEntriesForPayment(createdPayment, userID)
	if err != nil {
		return nil, err
	}

	return createdPayment, nil
}

// Sales Returns

func (s *SalesService) CreateSaleReturn(saleID uint, request models.SaleReturnRequest, userID uint) (*models.SaleReturn, error) {
	sale, err := s.salesRepo.FindByID(saleID)
	if err != nil {
		return nil, err
	}

	if sale.Status != models.SaleStatusInvoiced && sale.Status != models.SaleStatusPaid {
		return nil, errors.New("returns can only be created for invoiced or paid sales")
	}

	// Validate return items
	totalAmount := 0.0
	for _, item := range request.ReturnItems {
		saleItem, err := s.salesRepo.FindSaleItemByID(item.SaleItemID)
		if err != nil {
			return nil, fmt.Errorf("sale item %d not found", item.SaleItemID)
		}

		if saleItem.SaleID != saleID {
			return nil, errors.New("sale item does not belong to this sale")
		}

		if item.Quantity > saleItem.Quantity {
			return nil, errors.New("return quantity exceeds sold quantity")
		}

		totalAmount += saleItem.UnitPrice * float64(item.Quantity)
	}

	// Create return record
	saleReturn := &models.SaleReturn{
		SaleID:           saleID,
		ReturnNumber:     s.generateReturnNumber(),
		Date:             request.ReturnDate,
		Reason:           request.Reason,
		TotalAmount:      totalAmount,
		Status:           models.ReturnStatusPending,
		Notes:            request.Notes,
		UserID:           userID,
	}

	// Set type based on business logic (could be parameter or default)
	saleReturn.Type = models.ReturnTypeCreditNote // Default to credit note
	if saleReturn.Type == models.ReturnTypeCreditNote {
		saleReturn.CreditNoteNumber = s.generateCreditNoteNumber()
	}

	createdReturn, err := s.salesRepo.CreateReturn(saleReturn)
	if err != nil {
		return nil, err
	}

	// Create return items
	for _, item := range request.ReturnItems {
		returnItem := &models.SaleReturnItem{
			SaleReturnID: createdReturn.ID,
			SaleItemID:   item.SaleItemID,
			Quantity:     item.Quantity,
			Reason:       item.Reason,
		}

		saleItem, _ := s.salesRepo.FindSaleItemByID(item.SaleItemID)
		returnItem.UnitPrice = saleItem.UnitPrice
		returnItem.TotalAmount = saleItem.UnitPrice * float64(item.Quantity)

		err = s.salesRepo.CreateReturnItem(returnItem)
		if err != nil {
			return nil, err
		}
	}

	return createdReturn, nil
}

func (s *SalesService) GetSaleReturns(page, limit int) ([]models.SaleReturn, error) {
	return s.salesRepo.FindReturns(page, limit)
}

// Reporting and Analytics

func (s *SalesService) GetSalesSummary(startDate, endDate string) (*models.SalesSummaryResponse, error) {
	return s.salesRepo.GetSalesSummary(startDate, endDate)
}

func (s *SalesService) GetSalesAnalytics(period, year string) (*models.SalesAnalyticsResponse, error) {
	return s.salesRepo.GetSalesAnalytics(period, year)
}

func (s *SalesService) GetReceivablesReport() (*models.ReceivablesReportResponse, error) {
	return s.salesRepo.GetReceivablesReport()
}

func (s *SalesService) GetCustomerSales(customerID uint, page, limit int) ([]models.Sale, error) {
	return s.salesRepo.FindByCustomerID(customerID, page, limit)
}

func (s *SalesService) GetCustomerInvoices(customerID uint) ([]models.Sale, error) {
	return s.salesRepo.FindInvoicesByCustomerID(customerID)
}

// PDF Export

func (s *SalesService) ExportInvoicePDF(saleID uint) ([]byte, string, error) {
	sale, err := s.salesRepo.FindByID(saleID)
	if err != nil {
		return nil, "", err
	}

	if sale.InvoiceNumber == "" {
		return nil, "", errors.New("sale does not have an invoice number")
	}

	pdfData, err := s.pdfService.GenerateInvoicePDF(sale)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Invoice_%s.pdf", sale.InvoiceNumber)
	return pdfData, filename, nil
}

func (s *SalesService) ExportSalesReportPDF(startDate, endDate string) ([]byte, string, error) {
	sales, err := s.salesRepo.FindByDateRange(startDate, endDate)
	if err != nil {
		return nil, "", err
	}

	pdfData, err := s.pdfService.GenerateSalesReportPDF(sales, startDate, endDate)
	if err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("Sales_Report_%s_to_%s.pdf", startDate, endDate)
	return pdfData, filename, nil
}

// Private helper methods

func (s *SalesService) generateSaleCode(saleType string) (string, error) {
	prefix := "SAL"
	switch saleType {
	case models.SaleTypeQuotation:
		prefix = "QUO"
	case models.SaleTypeOrder:
		prefix = "ORD"
	case models.SaleTypeInvoice:
		prefix = "INV"
	}

	year := time.Now().Year()
	
	// Use database-level unique code generation with retry logic
	for attempt := 0; attempt < 100; attempt++ {
		// Get a base timestamp for uniqueness
		timestamp := time.Now().UnixMicro()
		baseNumber := (timestamp % 9999) + 1 // Get last 4 digits, avoid 0
		
		// Create code with timestamp-based number to avoid collisions
		code := fmt.Sprintf("%s-%d-%04d", prefix, year, baseNumber)
		
		// Check if this code already exists - use a more direct approach
		exists, err := s.salesRepo.ExistsByCode(code)
		if err != nil {
			return "", fmt.Errorf("error checking code existence: %v", err)
		}
		
		// If code doesn't exist, we can use it
		if !exists {
			return code, nil
		}
		
		// If code exists, add a small random delay and try again
		time.Sleep(time.Millisecond * time.Duration(attempt+1))
	}
	
	// Fallback: use UUID-based approach if all attempts failed
	return s.generateUniqueCodeWithUUID(prefix, year)
}

func (s *SalesService) generateInvoiceNumber() string {
	year := time.Now().Year()
	month := time.Now().Month()
	count, _ := s.salesRepo.CountInvoicesByMonth(year, int(month))
	return fmt.Sprintf("INV/%04d/%02d/%04d", year, month, count+1)
}

func (s *SalesService) generateQuotationNumber() string {
	year := time.Now().Year()
	month := time.Now().Month()
	count, _ := s.salesRepo.CountQuotationsByMonth(year, int(month))
	return fmt.Sprintf("QUO/%04d/%02d/%04d", year, month, count+1)
}

func (s *SalesService) generatePaymentNumber() string {
	year := time.Now().Year()
	month := time.Now().Month()
	count, _ := s.salesRepo.CountPaymentsByMonth(year, int(month))
	return fmt.Sprintf("PAY/%04d/%02d/%04d", year, month, count+1)
}

func (s *SalesService) generateReturnNumber() string {
	year := time.Now().Year()
	month := time.Now().Month()
	count, _ := s.salesRepo.CountReturnsByMonth(year, int(month))
	return fmt.Sprintf("RET/%04d/%02d/%04d", year, month, count+1)
}

func (s *SalesService) generateCreditNoteNumber() string {
	year := time.Now().Year()
	month := time.Now().Month()
	count, _ := s.salesRepo.CountCreditNotesByMonth(year, int(month))
	return fmt.Sprintf("CN/%04d/%02d/%04d", year, month, count+1)
}

func (s *SalesService) getDefaultCurrency(currency string) string {
	if currency == "" {
		return "IDR"
	}
	return currency
}

func (s *SalesService) getExchangeRate(currency string, rate float64) float64 {
	if currency == "IDR" || currency == "" {
		return 1.0
	}
	if rate > 0 {
		return rate
	}
	// In a real implementation, you would fetch current exchange rates
	return 1.0
}

func (s *SalesService) getDefaultPPNPercent(ppnPercent float64) float64 {
	if ppnPercent > 0 {
		return ppnPercent
	}
	return 11.0 // Default PPN 11%
}

func (s *SalesService) getDefaultPaymentTerms(terms string) string {
	if terms == "" {
		return "NET30"
	}
	return terms
}

func (s *SalesService) calculateDueDate(saleDate time.Time, paymentTerms string) time.Time {
	// Handle special payment terms first
	switch paymentTerms {
	case "COD":
		// Cash on Delivery - same day payment
		return saleDate
	case "EOM":
		// End of Month - due on last day of current month
		year, month, _ := saleDate.Date()
		// Get last day of the current month
		lastDayOfMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, saleDate.Location())
		return lastDayOfMonth
	case "2_10_NET_30":
		// 2/10, Net 30 - 2% discount if paid within 10 days, otherwise net 30 days
		return saleDate.AddDate(0, 0, 30)
	default:
		// Handle standard NET terms
		days := s.getDaysFromPaymentTerms(paymentTerms)
		dueDate := saleDate.AddDate(0, 0, days)
		
		// Optional: Skip weekends for business days calculation
		// Uncomment if business requires due dates to fall on business days
		// dueDate = s.adjustToBusinessDay(dueDate)
		
		return dueDate
	}
}

// getDaysFromPaymentTerms extracts the number of days from payment terms
func (s *SalesService) getDaysFromPaymentTerms(paymentTerms string) int {
	switch paymentTerms {
	case "NET15":
		return 15
	case "NET30":
		return 30
	case "NET45":
		return 45
	case "NET60":
		return 60
	case "NET90":
		return 90
	default:
		// Default to NET30 if unknown term
		return 30
	}
}

// adjustToBusinessDay adjusts the due date to the next business day if it falls on weekend
// This is optional and can be enabled based on business requirements
func (s *SalesService) adjustToBusinessDay(date time.Time) time.Time {
	weekday := date.Weekday()
	switch weekday {
	case time.Saturday:
		// Move to Monday
		return date.AddDate(0, 0, 2)
	case time.Sunday:
		// Move to Monday  
		return date.AddDate(0, 0, 1)
	default:
		// It's a weekday, return as is
		return date
	}
}

// Helper function to safely dereference float64 pointer
func (s *SalesService) dereferenceFloat64(ptr *float64) float64 {
	if ptr != nil {
		return *ptr
	}
	return 0
}

// validateSaleItemsStock validates that all items have sufficient stock
func (s *SalesService) validateSaleItemsStock(items []models.SaleItemRequest) error {
	for _, item := range items {
		product, err := s.productRepo.FindByID(item.ProductID)
		if err != nil {
			return fmt.Errorf("product %d not found", item.ProductID)
		}
		
		// Check stock availability for stockable products (non-service products need stock tracking)
		if !product.IsService && product.Stock < item.Quantity {
			return fmt.Errorf("insufficient stock for product %s. Available: %d, Required: %d", 
				product.Name, product.Stock, item.Quantity)
		}
	}
	return nil
}

// validateBusinessRules validates business rules for sales
func (s *SalesService) validateBusinessRules(customer *models.Contact, request models.SaleCreateRequest) error {
	// Check if customer is active
	if !customer.IsActive {
		return errors.New("cannot create sale for inactive customer")
	}
	
	// Validate items stock
	if err := s.validateSaleItemsStock(request.Items); err != nil {
		return err
	}
	
	// Calculate total for credit limit check
	totalAmount := s.calculateEstimatedTotal(request.Items, request.DiscountPercent, request.ShippingCost)
	
	// Check customer credit limit if applicable
	if customer.CreditLimit > 0 {
		outstandingAmount, err := s.salesRepo.GetCustomerOutstandingAmount(customer.ID)
		if err != nil {
			return fmt.Errorf("failed to check customer credit limit: %v", err)
		}
		
		if (outstandingAmount + totalAmount) > customer.CreditLimit {
			return fmt.Errorf("credit limit exceeded. Available credit: %.2f, Required: %.2f", 
				customer.CreditLimit - outstandingAmount, totalAmount)
		}
	}
	
	return nil
}

// calculateEstimatedTotal calculates estimated total for validation purposes
func (s *SalesService) calculateEstimatedTotal(items []models.SaleItemRequest, discountPercent, shippingCost float64) float64 {
	subtotal := 0.0
	for _, item := range items {
		lineTotal := float64(item.Quantity) * item.UnitPrice
		// Apply item-level discount
		if item.DiscountPercent > 0 {
			lineTotal -= lineTotal * (item.DiscountPercent / 100)
		} else if item.Discount > 0 {
			lineTotal -= item.Discount
		}
		subtotal += lineTotal
	}
	
	// Apply order-level discount
	if discountPercent > 0 {
		subtotal -= subtotal * (discountPercent / 100)
	}
	
	return subtotal + shippingCost
}

// Helper function to safely dereference time pointer
func (s *SalesService) dereferenceTime(ptr *time.Time) time.Time {
	if ptr != nil {
		return *ptr
	}
	return time.Time{}
}

func (s *SalesService) calculateSaleItemsFromRequest(sale *models.Sale, items []models.SaleItemRequest) error {
	subtotal := 0.0
	totalTax := 0.0

	// Clear existing items
	sale.SaleItems = []models.SaleItem{}

	for _, itemReq := range items {
		// Validate product exists
		product, err := s.productRepo.FindByID(itemReq.ProductID)
		if err != nil {
			return fmt.Errorf("product %d not found", itemReq.ProductID)
		}

		// Set default revenue account if not provided
		revenueAccountID := itemReq.RevenueAccountID
		if revenueAccountID == 0 {
			// Get default sales revenue account
			defaultAccount, err := s.getDefaultSalesRevenueAccount()
			if err != nil {
				return fmt.Errorf("failed to get default sales revenue account: %v", err)
			}
			revenueAccountID = defaultAccount.ID
		}

		// Handle description - use from request or fallback to product name
		description := itemReq.Description
		if description == "" {
			description = product.Name
		}

		// Handle discount percent - use new field or map from legacy field
		discountPercent := itemReq.DiscountPercent
		if discountPercent == 0 && itemReq.Discount > 0 {
			// If legacy discount is provided as flat amount, convert to percentage
			lineSubtotal := float64(itemReq.Quantity) * itemReq.UnitPrice
			if lineSubtotal > 0 {
				discountPercent = (itemReq.Discount / lineSubtotal) * 100
				// Cap discount percent to prevent overflow
				if discountPercent > 100 {
					discountPercent = 100
				}
			}
		}

		// Validate and cap numeric values to prevent overflow
		unitPrice := itemReq.UnitPrice
		if unitPrice > 999999999999.99 { // Max for decimal(15,2)
			unitPrice = 999999999999.99
		}

		if discountPercent > 100 {
			discountPercent = 100
		}

		// Create sale item with enhanced calculation
		item := models.SaleItem{
			ProductID:        itemReq.ProductID,
			Description:      description,
			Quantity:         itemReq.Quantity,
			UnitPrice:        unitPrice,
			DiscountPercent:  discountPercent,
			Taxable:          itemReq.Taxable, // Use field from request
			RevenueAccountID: revenueAccountID,
			// Legacy fields for backward compatibility
			Discount:         itemReq.Discount,
			Tax:              itemReq.Tax,
		}

		// Calculate line totals with proper precedence
		lineSubtotal := float64(item.Quantity) * item.UnitPrice
		discountAmount := item.Discount
		if discountAmount == 0 && item.DiscountPercent > 0 {
			discountAmount = lineSubtotal * (item.DiscountPercent / 100)
		}
		item.DiscountAmount = discountAmount
		item.LineTotal = lineSubtotal - discountAmount

		// Calculate taxes only if item is taxable
		if item.Taxable {
			item.PPNAmount = item.LineTotal * (sale.PPNPercent / 100)
			item.PPhAmount = item.LineTotal * (sale.PPhPercent / 100)
			item.TotalTax = item.PPNAmount - item.PPhAmount // PPN is added, PPh is deducted
		} else {
			item.PPNAmount = 0
			item.PPhAmount = 0
			item.TotalTax = 0
		}
		item.FinalAmount = item.LineTotal + item.TotalTax

		// Set computed fields for frontend compatibility
		item.TotalPrice = item.LineTotal // Frontend expects this field

		// Accumulate totals
		subtotal += item.LineTotal
		totalTax += item.TotalTax

		sale.SaleItems = append(sale.SaleItems, item)
	}

	// Calculate sale totals with enhanced logic
	sale.Subtotal = subtotal
	sale.DiscountAmount = subtotal * (sale.DiscountPercent / 100)
	
	// Calculate taxable amount including shipping if it's taxable
	taxableAmount := subtotal - sale.DiscountAmount
	if sale.ShippingTaxable {
		taxableAmount += sale.ShippingCost
	}
	sale.TaxableAmount = taxableAmount
	
	sale.PPN = sale.TaxableAmount * (sale.PPNPercent / 100)
	sale.PPh = sale.TaxableAmount * (sale.PPhPercent / 100)
	sale.TotalTax = sale.PPN - sale.PPh // PPN is added, PPh is deducted
	sale.TotalAmount = subtotal - sale.DiscountAmount + sale.TotalTax + sale.ShippingCost
	sale.OutstandingAmount = sale.TotalAmount - sale.PaidAmount

	// Set computed/legacy fields for frontend compatibility
	sale.Tax = sale.TotalTax
	sale.SubTotal = sale.Subtotal // Frontend compatibility alias

	return nil
}

func (s *SalesService) updateSaleItemsFromRequest(sale *models.Sale, items []models.SaleItemRequest) error {
	// Clear existing items and recreate from request
	sale.SaleItems = []models.SaleItem{}
	
	for _, itemReq := range items {
		// Validate product exists
		_, err := s.productRepo.FindByID(itemReq.ProductID)
		if err != nil {
			return fmt.Errorf("product %d not found", itemReq.ProductID)
		}

		// Create sale item
		// Set default revenue account if not provided
		revenueAccountID := itemReq.RevenueAccountID
		if revenueAccountID == 0 {
			// Get default sales revenue account
			defaultAccount, err := s.getDefaultSalesRevenueAccount()
			if err != nil {
				return fmt.Errorf("failed to get default sales revenue account: %v", err)
			}
			revenueAccountID = defaultAccount.ID
		}

		item := models.SaleItem{
			ProductID:        itemReq.ProductID,
			Quantity:         itemReq.Quantity,
			UnitPrice:        itemReq.UnitPrice,
			Discount:         itemReq.Discount,
			Tax:              itemReq.Tax,
			RevenueAccountID: revenueAccountID,
		}

		// Calculate totals
		item.TotalPrice = float64(item.Quantity) * item.UnitPrice
		sale.SaleItems = append(sale.SaleItems, item)
	}

	return nil
}

func (s *SalesService) calculateSaleTotals(sale *models.Sale, items []models.SaleItemCreateRequest) error {
	subtotal := 0.0
	_ = 0.0 // totalTax not used in current implementation

	// Clear existing items
	sale.SaleItems = []models.SaleItem{}

	for _, itemReq := range items {
		// Validate product exists
		_, err := s.productRepo.FindByID(itemReq.ProductID)
		if err != nil {
			return fmt.Errorf("product %d not found", itemReq.ProductID)
		}

	// Create sale item
		// Set default revenue account if not provided
		revenueAccountID := itemReq.RevenueAccountID
		if revenueAccountID == 0 {
			// Get default sales revenue account
			defaultAccount, err := s.getDefaultSalesRevenueAccount()
			if err != nil {
				return fmt.Errorf("failed to get default sales revenue account: %v", err)
			}
			revenueAccountID = defaultAccount.ID
		}

		item := models.SaleItem{
			ProductID:        itemReq.ProductID,
			Quantity:         itemReq.Quantity,
			UnitPrice:        itemReq.UnitPrice,
			Discount:         itemReq.Discount,
			Tax:              itemReq.Tax,
			RevenueAccountID: revenueAccountID,
		}

		// Note: Description field not available in current SaleItem model
		// If needed, add Description field to SaleItem model

		// Calculate totals
		item.TotalPrice = float64(item.Quantity) * item.UnitPrice - item.Discount + item.Tax

		subtotal += item.TotalPrice

		sale.SaleItems = append(sale.SaleItems, item)
	}

	// Calculate sale totals
	sale.TotalAmount = subtotal - (subtotal * sale.DiscountPercent / 100) + sale.ShippingCost + sale.Tax
	sale.OutstandingAmount = sale.TotalAmount - sale.PaidAmount

	return nil
}

func (s *SalesService) recalculateSaleTotals(sale *models.Sale) error {
	subtotal := 0.0
	totalTax := 0.0

	for i := range sale.SaleItems {
		item := &sale.SaleItems[i]
		
		// Calculate line totals with proper discount handling
		lineSubtotal := float64(item.Quantity) * item.UnitPrice
		discountAmount := item.Discount
		if discountAmount == 0 && item.DiscountPercent > 0 {
			discountAmount = lineSubtotal * (item.DiscountPercent / 100)
		}
		item.DiscountAmount = discountAmount
		item.LineTotal = lineSubtotal - discountAmount
		
		// Calculate taxes only if item is taxable
		if item.Taxable {
			item.PPNAmount = item.LineTotal * (sale.PPNPercent / 100)
			item.PPhAmount = item.LineTotal * (sale.PPhPercent / 100)
			item.TotalTax = item.PPNAmount - item.PPhAmount
		} else {
			item.PPNAmount = 0
			item.PPhAmount = 0
			item.TotalTax = 0
		}
		item.FinalAmount = item.LineTotal + item.TotalTax
		item.TotalPrice = item.LineTotal // For frontend compatibility
		
		subtotal += item.LineTotal
		totalTax += item.TotalTax
	}

	// Calculate sale totals with enhanced logic
	sale.Subtotal = subtotal
	sale.DiscountAmount = subtotal * (sale.DiscountPercent / 100)
	
	// Calculate taxable amount including shipping if it's taxable
	taxableAmount := subtotal - sale.DiscountAmount
	if sale.ShippingTaxable {
		taxableAmount += sale.ShippingCost
	}
	sale.TaxableAmount = taxableAmount
	
	sale.PPN = sale.TaxableAmount * (sale.PPNPercent / 100)
	sale.PPh = sale.TaxableAmount * (sale.PPhPercent / 100)
	sale.TotalTax = sale.PPN - sale.PPh
	sale.TotalAmount = subtotal - sale.DiscountAmount + sale.TotalTax + sale.ShippingCost
	sale.OutstandingAmount = sale.TotalAmount - sale.PaidAmount
	
	// Set computed/legacy fields for frontend compatibility
	sale.Tax = sale.TotalTax
	sale.SubTotal = sale.Subtotal

	return nil
}

func (s *SalesService) updateSaleItems(sale *models.Sale, items []models.SaleItemUpdateRequest) error {
	// Update existing items based on SaleItemUpdateRequest structure
	for i := range sale.SaleItems {
		item := &sale.SaleItems[i]
		
		// Find matching update request by index or other logic
		// Note: This is a simplified implementation
		// In practice, you'd need a way to identify which item to update
		if i < len(items) {
			itemReq := items[i]
			
			// Update fields if provided in request
			if itemReq.Quantity != nil {
				item.Quantity = *itemReq.Quantity
			}
			if itemReq.UnitPrice != nil {
				item.UnitPrice = *itemReq.UnitPrice
			}
			if itemReq.Discount != nil {
				item.Discount = *itemReq.Discount
			}
			if itemReq.Tax != nil {
				item.Tax = *itemReq.Tax
			}
			if itemReq.RevenueAccountID != nil {
				item.RevenueAccountID = *itemReq.RevenueAccountID
			}
			
			// Recalculate total price
			item.TotalPrice = float64(item.Quantity) * item.UnitPrice - item.Discount + item.Tax
		}
	}

	return nil
}

func (s *SalesService) updateInventoryForSale(sale *models.Sale) error {
	// Implementation for inventory updates
	// This would interact with inventory service/repository
	return nil
}

func (s *SalesService) restoreInventoryForSale(sale *models.Sale) error {
	// Implementation for restoring inventory
	return nil
}

func (s *SalesService) createJournalEntriesForSale(sale *models.Sale, userID uint) error {
	// Create journal entries for the sale
	// Debit: Accounts Receivable (or Cash if COD)
	// Credit: Sales Revenue
	// Credit/Debit: Tax accounts
	
	// Try to use journal service interface first
	if s.journalService != nil {
		return s.journalService.CreateSaleJournalEntries(sale, userID)
	}
	
	// Fallback to direct accounting service implementation
	log.Printf("Using direct accounting service for sale %d", sale.ID)
	return s.createSaleAccountingEntries(sale, userID)
}

func (s *SalesService) createJournalEntriesForPayment(payment *models.SalePayment, userID uint) error {
	// Create journal entries for payment
	// Debit: Cash/Bank Account
	// Credit: Accounts Receivable
	
	// Try to use journal service interface first
	if s.journalService != nil {
		return s.journalService.CreatePaymentJournalEntries(payment, userID)
	}
	
	// Direct implementation when journal service is not available
	log.Printf("Using direct implementation for payment journal entries for payment %d", payment.ID)
	
	// Get Accounts Receivable account
	accountsReceivable, err := s.accountRepo.GetAccountByCode("1201") // Piutang Usaha
	if err != nil {
		// Try fallback account codes
		accountsReceivable, err = s.accountRepo.GetAccountByCode("1200")
		if err != nil {
			log.Printf("Error: No accounts receivable account found: %v", err)
			return fmt.Errorf("accounts receivable account not found: %v", err)
		}
	}
	
	// Get Cash/Bank account
	var cashBankAccount *models.Account
	if payment.CashBankID != nil && *payment.CashBankID > 0 {
		// Try to get account from CashBank record
		var cashBank models.CashBank
		if err := s.db.First(&cashBank, *payment.CashBankID).Error; err == nil {
			cashBankAccount, err = s.accountRepo.FindByID(context.Background(), cashBank.AccountID)
			if err != nil {
				log.Printf("Warning: Could not find GL account for cash bank %d, using default cash account", *payment.CashBankID)
			}
		}
	}
	
	// If no specific bank account found, use default cash account
	if cashBankAccount == nil {
		cashBankAccount, err = s.accountRepo.GetAccountByCode("1104") // Bank Mandiri or default bank
		if err != nil {
			// Try other bank accounts
			cashBankAccount, err = s.accountRepo.GetAccountByCode("1102") // Bank BCA
			if err != nil {
				// Try cash account
				cashBankAccount, err = s.accountRepo.GetAccountByCode("1101") // Kas
				if err != nil {
					return fmt.Errorf("no cash/bank account found for payment")
				}
			}
		}
	}
	
	// Get customer name for description
	customerName := "Unknown Customer"
	// Check if payment.Sale.Customer is loaded by checking the Customer.ID
	if payment.Sale.ID > 0 && payment.Sale.Customer.ID > 0 {
		customerName = payment.Sale.Customer.Name
	} else {
		// Load sale and customer if not preloaded
		var sale models.Sale
		if err := s.db.Preload("Customer").First(&sale, payment.SaleID).Error; err == nil {
			if sale.Customer.ID > 0 {
				customerName = sale.Customer.Name
			}
		}
	}
	
	// Create journal entry
	journalEntry := &models.JournalEntry{
		EntryDate:       payment.Date,
		Description:     fmt.Sprintf("Customer Payment %s - %s", payment.PaymentNumber, customerName),
		Reference:       payment.PaymentNumber,
		ReferenceType:   models.JournalRefPayment,
		ReferenceID:     &payment.ID,
		UserID:          userID,
		Status:          models.JournalStatusDraft,
		TotalDebit:      payment.Amount,
		TotalCredit:     payment.Amount,
		IsBalanced:      true,
		IsAutoGenerated: true,
		AccountID:       &cashBankAccount.ID, // Primary account for the entry
	}
	
	// Create journal entry
	if err := s.db.Create(journalEntry).Error; err != nil {
		return fmt.Errorf("failed to create payment journal entry: %v", err)
	}
	
	// Update account balances manually
	// 1. Debit: Cash/Bank Account (increase cash/bank balance)
	err = s.accountRepo.UpdateBalance(context.Background(), cashBankAccount.ID, payment.Amount, 0)
	if err != nil {
		return fmt.Errorf("failed to update cash/bank account balance: %v", err)
	}
	
	// 2. Credit: Accounts Receivable (decrease AR balance)
	err = s.accountRepo.UpdateBalance(context.Background(), accountsReceivable.ID, 0, payment.Amount)
	if err != nil {
		return fmt.Errorf("failed to update accounts receivable balance: %v", err)
	}
	
	// Update journal entry status to posted since we've updated balances
	now := time.Now()
	journalEntry.Status = models.JournalStatusPosted
	journalEntry.PostingDate = &now
	journalEntry.PostedBy = &userID
	
	if err := s.db.Save(journalEntry).Error; err != nil {
		return fmt.Errorf("failed to update journal entry status: %v", err)
	}
	
	log.Printf("Successfully created and posted payment journal entry %d", journalEntry.ID)
	return nil
}

func (s *SalesService) createReversalJournalEntries(sale *models.Sale, userID uint, reason string) error {
	// Create reversal journal entries
	
	// Try to use journal service interface first
	if s.journalService != nil {
		return s.journalService.CreateSaleReversalJournalEntries(sale, userID, reason)
	}
	
	// Fallback to direct accounting service implementation
	log.Printf("Using direct reversal accounting service for sale %d", sale.ID)
	return s.createSaleReversalJournalEntries(sale, userID, reason)
}

// validateDocumentTypeRequirements validates document type specific requirements
func (s *SalesService) validateDocumentTypeRequirements(saleType string, validUntil *time.Time, date time.Time) error {
	switch saleType {
	case models.SaleTypeQuotation:
		// Quotations must have a valid until date
		if validUntil == nil || validUntil.IsZero() {
			return errors.New("quotations must have a valid until date")
		}
		// Valid until date must be after the quotation date
		if validUntil.Before(date) {
			return errors.New("valid until date must be after the quotation date")
		}
	case models.SaleTypeInvoice:
		// Invoices should have due dates calculated based on payment terms
		// This validation is handled in the creation process
	}
	return nil
}

// getDefaultSalesRevenueAccount gets the default sales revenue account
func (s *SalesService) getDefaultSalesRevenueAccount() (*models.Account, error) {
	// Try to find a sales revenue account
	// First, try to find an account with "Sales" in the name and type REVENUE
	accounts, err := s.accountRepo.FindByType(context.Background(), "REVENUE")
	if err != nil {
		return nil, err
	}

	// Look for an account with "Sales" or "Revenue" in the name
	for _, account := range accounts {
		if account.Name == "Sales Revenue" || account.Name == "Sales" || 
		   account.Code == "4000" || account.Code == "400" {
			return &account, nil
		}
	}

	// If no specific sales account found, return the first revenue account
	if len(accounts) > 0 {
		return &accounts[0], nil
	}

	// If no revenue accounts exist, create a default one
	return s.createDefaultSalesRevenueAccount()
}

// createDefaultSalesRevenueAccount creates a default sales revenue account if none exists
func (s *SalesService) createDefaultSalesRevenueAccount() (*models.Account, error) {
	// Create a default sales revenue account request
	defaultAccountReq := &models.AccountCreateRequest{
		Code:           "4000",
		Name:           "Sales Revenue",
		Type:           models.AccountType(models.AccountTypeRevenue),
		Description:    "Default sales revenue account",
		OpeningBalance: 0,
	}

	// Use account repository to create the account
	createdAccount, err := s.accountRepo.Create(context.Background(), defaultAccountReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create default sales revenue account: %v", err)
	}

	return createdAccount, nil
}

// generateUniqueCodeWithUUID creates a unique code using a UUID-based approach as fallback
func (s *SalesService) generateUniqueCodeWithUUID(prefix string, year int) (string, error) {
	// Generate 4 random bytes for uniqueness
	randomBytes := make([]byte, 4)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}
	
	// Convert to a 4-digit number
	randomNumber := uint32(randomBytes[0])<<24 | uint32(randomBytes[1])<<16 | uint32(randomBytes[2])<<8 | uint32(randomBytes[3])
	randomNumber = (randomNumber % 9999) + 1
	
	// Try multiple variations if needed
	for attempt := 0; attempt < 50; attempt++ {
		code := fmt.Sprintf("%s-%d-%04d", prefix, year, (randomNumber+uint32(attempt))%10000)
		
		exists, err := s.salesRepo.ExistsByCode(code)
		if err != nil {
			return "", fmt.Errorf("error checking UUID-based code existence: %v", err)
		}
		
		if !exists {
			return code, nil
		}
	}
	
	return "", fmt.Errorf("failed to generate unique code even with UUID fallback")
}
