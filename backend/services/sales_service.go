package services

import (
	"errors"
	"fmt"
	"math"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
)

type SalesService struct {
	salesRepo       *repositories.SalesRepository
	productRepo     *repositories.ProductRepository
	contactRepo     repositories.ContactRepository
	accountRepo     repositories.AccountRepository
	journalService  JournalServiceInterface
	pdfService      PDFServiceInterface
	approvalService *ApprovalService
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
}

type SalesResult struct {
	Data       []models.Sale `json:"data"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	Limit      int           `json:"limit"`
	TotalPages int           `json:"total_pages"`
}

func NewSalesService(salesRepo *repositories.SalesRepository, productRepo *repositories.ProductRepository, contactRepo repositories.ContactRepository, accountRepo repositories.AccountRepository, journalService JournalServiceInterface, pdfService PDFServiceInterface, approvalService *ApprovalService) *SalesService {
	return &SalesService{
		salesRepo:       salesRepo,
		productRepo:     productRepo,
		contactRepo:     contactRepo,
		accountRepo:     accountRepo,
		journalService:  journalService,
		pdfService:      pdfService,
		approvalService: approvalService,
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

	// Validate sales person if provided
	if request.SalesPersonID != nil {
		// Check if sales person exists (implement user repository check)
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

// Status Management

func (s *SalesService) ConfirmSale(id uint, userID uint) error {
	sale, err := s.salesRepo.FindByID(id)
	if err != nil {
		return err
	}

	if sale.Status != models.SaleStatusDraft && sale.Status != models.SaleStatusPending {
		return errors.New("sale cannot be confirmed in current status")
	}

	// Update status
	sale.Status = models.SaleStatusConfirmed

	// Update inventory
	err = s.updateInventoryForSale(sale)
	if err != nil {
		return err
	}

	// Create journal entries
	err = s.createJournalEntriesForSale(sale, userID)
	if err != nil {
		return err
	}

	_, err = s.salesRepo.Update(sale)
	return err
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
		return nil, err
	}

	if sale.Status != models.SaleStatusInvoiced && sale.Status != models.SaleStatusOverdue {
		return nil, errors.New("payments can only be recorded for invoiced sales")
	}

	if request.Amount > sale.OutstandingAmount {
		return nil, errors.New("payment amount exceeds outstanding amount")
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
	count, err := s.salesRepo.CountByTypeAndYear(saleType, year)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%d-%04d", prefix, year, count+1), nil
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
	days := 30 // default
	switch paymentTerms {
	case "COD":
		days = 0
	case "NET15":
		days = 15
	case "NET30":
		days = 30
	case "NET45":
		days = 45
	case "NET60":
		days = 60
	case "NET90":
		days = 90
	}
	return saleDate.AddDate(0, 0, days)
}

// Helper function to safely dereference float64 pointer
func (s *SalesService) dereferenceFloat64(ptr *float64) float64 {
	if ptr != nil {
		return *ptr
	}
	return 0
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

	// Clear existing items
	sale.SaleItems = []models.SaleItem{}

	for _, itemReq := range items {
		// Validate product exists
		_, err := s.productRepo.FindByID(itemReq.ProductID)
		if err != nil {
			return fmt.Errorf("product %d not found", itemReq.ProductID)
		}

		// Create sale item
		item := models.SaleItem{
			ProductID:        itemReq.ProductID,
			Quantity:         itemReq.Quantity,
			UnitPrice:        itemReq.UnitPrice,
			Discount:         itemReq.Discount,
			Tax:              itemReq.Tax,
			RevenueAccountID: itemReq.RevenueAccountID,
		}

		// Calculate totals
		item.TotalPrice = float64(item.Quantity) * item.UnitPrice
		subtotal += item.TotalPrice

		sale.SaleItems = append(sale.SaleItems, item)
	}

	// Calculate sale totals
	sale.TotalAmount = subtotal - (subtotal * sale.DiscountPercent / 100) + sale.ShippingCost
	sale.OutstandingAmount = sale.TotalAmount - sale.PaidAmount

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
		item := models.SaleItem{
			ProductID:        itemReq.ProductID,
			Quantity:         itemReq.Quantity,
			UnitPrice:        itemReq.UnitPrice,
			Discount:         itemReq.Discount,
			Tax:              itemReq.Tax,
			RevenueAccountID: itemReq.RevenueAccountID,
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
		item := models.SaleItem{
			ProductID:        itemReq.ProductID,
			Quantity:         itemReq.Quantity,
			UnitPrice:        itemReq.UnitPrice,
			Discount:         itemReq.Discount,
			Tax:              itemReq.Tax,
			RevenueAccountID: itemReq.RevenueAccountID,
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
	_ = 0.0 // totalTax not used in current implementation

	for i := range sale.SaleItems {
		item := &sale.SaleItems[i]
		
		// Calculate totals
		item.TotalPrice = float64(item.Quantity) * item.UnitPrice - item.Discount + item.Tax
		subtotal += item.TotalPrice
	}

	// Calculate sale totals
	sale.TotalAmount = subtotal - (subtotal * sale.DiscountPercent / 100) + sale.ShippingCost + sale.Tax
	sale.OutstandingAmount = sale.TotalAmount - sale.PaidAmount

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
	return s.journalService.CreateSaleJournalEntries(sale, userID)
}

func (s *SalesService) createJournalEntriesForPayment(payment *models.SalePayment, userID uint) error {
	// Create journal entries for payment
	// Debit: Cash/Bank Account
	// Credit: Accounts Receivable
	return s.journalService.CreatePaymentJournalEntries(payment, userID)
}

func (s *SalesService) createReversalJournalEntries(sale *models.Sale, userID uint, reason string) error {
	// Create reversal journal entries
	return s.journalService.CreateSaleReversalJournalEntries(sale, userID, reason)
}

// Approval Integration Methods

// SubmitForApproval submits a sale for approval if required
func (s *SalesService) SubmitForApproval(id uint, userID uint) error {
	sale, err := s.salesRepo.FindByID(id)
	if err != nil {
		return err
	}

	if sale.Status != models.SaleStatusDraft {
		return errors.New("only draft sales can be submitted for approval")
	}

	// Check if approval is required based on amount and workflow configuration
	requiresApproval := s.checkIfApprovalRequired(sale.TotalAmount)
	
	if !requiresApproval {
		// No approval required, move directly to confirmed
		sale.Status = models.SaleStatusConfirmed
		_, err = s.salesRepo.Update(sale)
		return err
	}

	// Create approval request
	approvalReq := models.CreateApprovalRequestDTO{
		EntityType:     models.EntityTypeSale,
		EntityID:       sale.ID,
		Amount:         sale.TotalAmount,
		Priority:       models.ApprovalPriorityNormal,
		RequestTitle:   fmt.Sprintf("Sale Approval - %s", sale.Code),
		RequestMessage: fmt.Sprintf("Approval request for sale %s with total amount %.2f", sale.Code, sale.TotalAmount),
	}

	// Determine priority based on amount
	if sale.TotalAmount > 100000000 { // 100M IDR
		approvalReq.Priority = models.ApprovalPriorityUrgent
	} else if sale.TotalAmount > 50000000 { // 50M IDR
		approvalReq.Priority = models.ApprovalPriorityHigh
	}

	_, err = s.approvalService.CreateApprovalRequest(approvalReq, userID)
	if err != nil {
		return err
	}

	// Update sale status
	now := time.Now()
	sale.Status = models.SaleStatusPending
	sale.UpdatedAt = now

	_, err = s.salesRepo.Update(sale)
	return err
}

// checkIfApprovalRequired determines if approval is required based on amount
func (s *SalesService) checkIfApprovalRequired(amount float64) bool {
	// Check if there's an active workflow for this amount
	workflow, err := s.approvalService.GetWorkflowByAmount(models.ApprovalModuleSales, amount)
	return err == nil && workflow != nil
}

// ProcessSaleApproval handles the approval/rejection of a sale
func (s *SalesService) ProcessSaleApproval(saleID uint, approved bool, userID uint) error {
	sale, err := s.salesRepo.FindByID(saleID)
	if err != nil {
		return err
	}

	if sale.Status != models.SaleStatusPending {
		return errors.New("sale is not pending approval")
	}

	now := time.Now()
	if approved {
		// Sale approved - can proceed to confirmation
		sale.Status = models.SaleStatusConfirmed
	} else {
		// Sale rejected - back to draft or cancelled
		sale.Status = models.SaleStatusCancelled
	}

	sale.UpdatedAt = now
	_, err = s.salesRepo.Update(sale)
	return err
}
