package services

import (
	"errors"
	"fmt"
	"log"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"

	"gorm.io/gorm"
)

// SalesValidationService handles all validation logic for sales operations
type SalesValidationService struct {
	db          *gorm.DB
	contactRepo repositories.ContactRepository
	productRepo *repositories.ProductRepository
	accountRepo repositories.AccountRepository
}

// SalesValidationResult represents the result of a validation operation
type SalesValidationResult struct {
	IsValid    bool              `json:"is_valid"`
	Errors     []ValidationError `json:"errors,omitempty"`
	Warnings   []ValidationError `json:"warnings,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Field    string                 `json:"field"`
	Code     string                 `json:"code"`
	Message  string                 `json:"message"`
	Value    interface{}            `json:"value,omitempty"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// Validation error codes
const (
	ValidationCodeRequired           = "REQUIRED"
	ValidationCodeInvalidFormat      = "INVALID_FORMAT"
	ValidationCodeInvalidRange       = "INVALID_RANGE"
	ValidationCodeNotFound          = "NOT_FOUND"
	ValidationCodeInvalidStatus     = "INVALID_STATUS"
	ValidationCodeBusinessRule     = "BUSINESS_RULE"
	ValidationCodeInsufficientStock = "INSUFFICIENT_STOCK"
	ValidationCodeCreditLimit      = "CREDIT_LIMIT_EXCEEDED"
	ValidationCodeDuplicateValue   = "DUPLICATE_VALUE"
	ValidationCodeInvalidDate      = "INVALID_DATE"
)

func NewSalesValidationService(db *gorm.DB, contactRepo repositories.ContactRepository, productRepo *repositories.ProductRepository, accountRepo repositories.AccountRepository) *SalesValidationService {
	return &SalesValidationService{
		db:          db,
		contactRepo: contactRepo,
		productRepo: productRepo,
		accountRepo: accountRepo,
	}
}

// ValidateSaleCreateRequest validates a sale creation request
func (svs *SalesValidationService) ValidateSaleCreateRequest(request models.SaleCreateRequest) *SalesValidationResult {
	log.Printf("🔍 Validating sale creation request for customer %d", request.CustomerID)
	
	result := &SalesValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
		Metadata: make(map[string]interface{}),
	}
	
	// 1. Validate basic required fields
	svs.validateBasicFields(request, result)
	
	// 2. Validate customer
	customer, err := svs.validateCustomer(request.CustomerID, result)
	if err == nil && customer != nil {
		// 3. Validate credit limit
		svs.validateCreditLimit(customer, request, result)
	}
	
	// 4. Validate sales person if provided
	if request.SalesPersonID != nil {
		svs.validateSalesPerson(*request.SalesPersonID, result)
	}
	
	// 5. Validate date logic
	svs.validateDateLogic(request, result)
	
	// 6. Validate tax configuration
	svs.validateTaxConfiguration(request, result)
	
	// 7. Validate items
	svs.validateSaleItems(request.Items, result)
	
	// 8. Validate document type requirements
	svs.validateDocumentType(request, result)
	
	// Set overall validation result
	result.IsValid = len(result.Errors) == 0
	
	if result.IsValid {
		log.Printf("✅ Sale creation validation passed with %d warnings", len(result.Warnings))
	} else {
		log.Printf("❌ Sale creation validation failed with %d errors", len(result.Errors))
	}
	
	return result
}

// ValidateSaleUpdateRequest validates a sale update request
func (svs *SalesValidationService) ValidateSaleUpdateRequest(sale *models.Sale, request models.SaleUpdateRequest) *SalesValidationResult {
	log.Printf("🔍 Validating sale update request for sale %d", sale.ID)
	
	result := &SalesValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
		Metadata: map[string]interface{}{
			"sale_id":      sale.ID,
			"current_status": sale.Status,
		},
	}
	
	// 1. Check if sale can be updated
	if !svs.canSaleBeUpdated(sale.Status) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "status",
			Code:    ValidationCodeInvalidStatus,
			Message: fmt.Sprintf("Sale with status '%s' cannot be updated", sale.Status),
			Value:   sale.Status,
			Context: map[string]interface{}{
				"allowed_statuses": []string{models.SaleStatusDraft, models.SaleStatusPending},
			},
		})
	}
	
	// 2. Validate customer if being updated
	if request.CustomerID != nil {
		_, err := svs.validateCustomer(*request.CustomerID, result)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "customer_id",
				Code:    ValidationCodeNotFound,
				Message: "Customer not found",
				Value:   *request.CustomerID,
			})
		}
	}
	
	// 3. Validate sales person if being updated
	if request.SalesPersonID != nil {
		svs.validateSalesPerson(*request.SalesPersonID, result)
	}
	
	// 4. Validate date updates
	if request.Date != nil || request.DueDate != nil || request.ValidUntil != nil {
		svs.validateDateUpdates(sale, request, result)
	}
	
	// 5. Validate items if being updated
	if len(request.Items) > 0 {
		svs.validateSaleItems(request.Items, result)
	}
	
	result.IsValid = len(result.Errors) == 0
	
	if result.IsValid {
		log.Printf("✅ Sale update validation passed")
	} else {
		log.Printf("❌ Sale update validation failed with %d errors", len(result.Errors))
	}
	
	return result
}

// ValidatePaymentRequest validates a payment request
func (svs *SalesValidationService) ValidatePaymentRequest(sale *models.Sale, request models.SalePaymentRequest) *SalesValidationResult {
	log.Printf("🔍 Validating payment request for sale %d", sale.ID)
	
	result := &SalesValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
		Metadata: map[string]interface{}{
			"sale_id":            sale.ID,
			"outstanding_amount": sale.OutstandingAmount,
			"current_status":     sale.Status,
		},
	}
	
	// 1. Validate payment amount
	svs.validatePaymentAmount(sale, request.Amount, result)
	
	// 2. Validate sale status for payment
	svs.validateSaleStatusForPayment(sale.Status, result)
	
	// 3. Validate payment date
	svs.validatePaymentDate(request.PaymentDate, result)
	
	// 4. Validate payment method
	svs.validatePaymentMethod(request.PaymentMethod, result)
	
	// 5. Validate cash/bank account if provided
	if request.CashBankID != nil {
		svs.validateCashBankAccount(*request.CashBankID, result)
	}
	
	result.IsValid = len(result.Errors) == 0
	
	if result.IsValid {
		log.Printf("✅ Payment validation passed")
	} else {
		log.Printf("❌ Payment validation failed with %d errors", len(result.Errors))
	}
	
	return result
}

// Private validation methods

func (svs *SalesValidationService) validateBasicFields(request models.SaleCreateRequest, result *SalesValidationResult) {
	// Validate customer ID
	if request.CustomerID == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "customer_id",
			Code:    ValidationCodeRequired,
			Message: "Customer ID is required",
		})
	}
	
	// Validate type
	if request.Type == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "type",
			Code:    ValidationCodeRequired,
			Message: "Sale type is required",
		})
	} else {
		validTypes := []string{models.SaleTypeQuotation, models.SaleTypeOrder, models.SaleTypeInvoice}
		isValidType := false
		for _, validType := range validTypes {
			if request.Type == validType {
				isValidType = true
				break
			}
		}
		if !isValidType {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "type",
				Code:    ValidationCodeInvalidFormat,
				Message: "Invalid sale type",
				Value:   request.Type,
				Context: map[string]interface{}{
					"allowed_types": validTypes,
				},
			})
		}
	}
	
	// Validate date
	if request.Date.IsZero() {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "date",
			Code:    ValidationCodeRequired,
			Message: "Sale date is required",
		})
	}
	
	// Validate items
	if len(request.Items) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "items",
			Code:    ValidationCodeRequired,
			Message: "At least one sale item is required",
		})
	}
}

func (svs *SalesValidationService) validateCustomer(customerID uint, result *SalesValidationResult) (*models.Contact, error) {
	customer, err := svs.contactRepo.GetByID(customerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "customer_id",
				Code:    ValidationCodeNotFound,
				Message: "Customer not found",
				Value:   customerID,
			})
			return nil, err
		}
		result.Errors = append(result.Errors, ValidationError{
			Field:   "customer_id",
			Code:    ValidationCodeBusinessRule,
			Message: "Error validating customer",
			Value:   customerID,
			Context: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return nil, err
	}
	
	// Check if customer is active
	if !customer.IsActive {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "customer_id",
			Code:    ValidationCodeInvalidStatus,
			Message: "Customer is not active",
			Value:   customerID,
		})
	}
	
	// Check customer type
	if customer.Type != "CUSTOMER" && customer.Type != "BOTH" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "customer_id",
			Code:    ValidationCodeBusinessRule,
			Message: "Contact is not configured as a customer",
			Value:   customerID,
			Context: map[string]interface{}{
				"customer_type": customer.Type,
			},
		})
	}
	
	return customer, nil
}

func (svs *SalesValidationService) validateSalesPerson(salesPersonID uint, result *SalesValidationResult) {
	salesPerson, err := svs.contactRepo.GetByID(salesPersonID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "sales_person_id",
				Code:    ValidationCodeNotFound,
				Message: "Sales person not found",
				Value:   salesPersonID,
			})
			return
		}
		result.Errors = append(result.Errors, ValidationError{
			Field:   "sales_person_id",
			Code:    ValidationCodeBusinessRule,
			Message: "Error validating sales person",
			Value:   salesPersonID,
			Context: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return
	}
	
	// Validate that the contact is an employee
	if salesPerson.Type != "EMPLOYEE" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "sales_person_id",
			Code:    ValidationCodeInvalidFormat,
			Message: "Contact is not an employee",
			Value:   salesPersonID,
			Context: map[string]interface{}{
				"contact_type": salesPerson.Type,
			},
		})
	}
	
	// Check if employee is active
	if !salesPerson.IsActive {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "sales_person_id",
			Code:    ValidationCodeInvalidStatus,
			Message: "Sales person is not active",
			Value:   salesPersonID,
		})
	}
}

func (svs *SalesValidationService) validateDateLogic(request models.SaleCreateRequest, result *SalesValidationResult) {
	now := time.Now()
	
	// Validate sale date is not too far in the future
	if request.Date.After(now.AddDate(0, 0, 30)) {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "date",
			Code:    ValidationCodeInvalidDate,
			Message: "Sale date is more than 30 days in the future",
			Value:   request.Date.Format("2006-01-02"),
		})
	}
	
	// Validate due date is after sale date
	if !request.DueDate.IsZero() && request.DueDate.Before(request.Date) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "due_date",
			Code:    ValidationCodeInvalidDate,
			Message: "Due date must be on or after sale date",
			Value:   request.DueDate.Format("2006-01-02"),
			Context: map[string]interface{}{
				"sale_date": request.Date.Format("2006-01-02"),
			},
		})
	}
	
	// Validate valid until date for quotations
	if request.Type == models.SaleTypeQuotation {
		if request.ValidUntil == nil {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   "valid_until",
				Code:    ValidationCodeRequired,
				Message: "Valid until date is recommended for quotations",
			})
		} else if request.ValidUntil.Before(request.Date) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "valid_until",
				Code:    ValidationCodeInvalidDate,
				Message: "Valid until date must be on or after sale date",
				Value:   request.ValidUntil.Format("2006-01-02"),
			})
		}
	}
}

func (svs *SalesValidationService) validateTaxConfiguration(request models.SaleCreateRequest, result *SalesValidationResult) {
	// Validate tax rates are within reasonable ranges
	if request.PPNPercent != nil && (*request.PPNPercent < 0 || *request.PPNPercent > 100) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "ppn_percent",
			Code:    ValidationCodeInvalidRange,
			Message: "PPN percentage must be between 0 and 100",
			Value:   *request.PPNPercent,
		})
	}
	
	if request.PPNRate < 0 || request.PPNRate > 100 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "ppn_rate",
			Code:    ValidationCodeInvalidRange,
			Message: "PPN rate must be between 0 and 100",
			Value:   request.PPNRate,
		})
	}
	
	if request.PPhPercent < 0 || request.PPhPercent > 100 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "pph_percent",
			Code:    ValidationCodeInvalidRange,
			Message: "PPh percentage must be between 0 and 100",
			Value:   request.PPhPercent,
		})
	}
	
	if request.DiscountPercent < 0 || request.DiscountPercent > 100 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "discount_percent",
			Code:    ValidationCodeInvalidRange,
			Message: "Discount percentage must be between 0 and 100",
			Value:   request.DiscountPercent,
		})
	}
	
	// Validate shipping cost
	if request.ShippingCost < 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "shipping_cost",
			Code:    ValidationCodeInvalidRange,
			Message: "Shipping cost cannot be negative",
			Value:   request.ShippingCost,
		})
	}
}

func (svs *SalesValidationService) validateSaleItems(items []models.SaleItemRequest, result *SalesValidationResult) {
	for i, item := range items {
		fieldPrefix := fmt.Sprintf("items[%d]", i)
		
		// Validate product exists
		if item.ProductID == 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldPrefix + ".product_id",
				Code:    ValidationCodeRequired,
				Message: "Product ID is required",
			})
			continue
		}
		
		product, err := svs.productRepo.FindByID(item.ProductID)
		if err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldPrefix + ".product_id",
				Code:    ValidationCodeNotFound,
				Message: "Product not found",
				Value:   item.ProductID,
			})
			continue
		}
		
		// Validate quantity
		if item.Quantity <= 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldPrefix + ".quantity",
				Code:    ValidationCodeInvalidRange,
				Message: "Quantity must be greater than 0",
				Value:   item.Quantity,
			})
		}
		
		// Validate stock availability
		if product.Stock < item.Quantity {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldPrefix + ".quantity",
				Code:    ValidationCodeInsufficientStock,
				Message: "Insufficient stock available",
				Value:   item.Quantity,
				Context: map[string]interface{}{
					"available_stock": product.Stock,
					"product_name":    product.Name,
				},
			})
		}
		
		// Validate unit price
		if item.UnitPrice < 0 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldPrefix + ".unit_price",
				Code:    ValidationCodeInvalidRange,
				Message: "Unit price cannot be negative",
				Value:   item.UnitPrice,
			})
		}
		
		// Warn if price is significantly different from product price
		if product.SalePrice > 0 && item.UnitPrice > 0 {
			priceDeviation := (item.UnitPrice - product.SalePrice) / product.SalePrice * 100
			if priceDeviation > 50 || priceDeviation < -50 {
				result.Warnings = append(result.Warnings, ValidationError{
					Field:   fieldPrefix + ".unit_price",
					Code:    ValidationCodeBusinessRule,
					Message: "Unit price significantly different from product sales price",
					Value:   item.UnitPrice,
					Context: map[string]interface{}{
						"product_sales_price": product.SalePrice,
						"deviation":           fmt.Sprintf("%.1f%%", priceDeviation),
					},
				})
			}
		}
		
		// Validate discount
		if item.DiscountPercent < 0 || item.DiscountPercent > 100 {
			result.Errors = append(result.Errors, ValidationError{
				Field:   fieldPrefix + ".discount_percent",
				Code:    ValidationCodeInvalidRange,
				Message: "Discount percentage must be between 0 and 100",
				Value:   item.DiscountPercent,
			})
		}
	}
}

func (svs *SalesValidationService) validateDocumentType(request models.SaleCreateRequest, result *SalesValidationResult) {
	switch request.Type {
	case models.SaleTypeQuotation:
		// Quotations should have valid until date
		if request.ValidUntil == nil {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   "valid_until",
				Code:    ValidationCodeBusinessRule,
				Message: "Valid until date is recommended for quotations",
			})
		}
	case models.SaleTypeOrder:
		// Orders should have due date
		if request.DueDate.IsZero() {
			result.Warnings = append(result.Warnings, ValidationError{
				Field:   "due_date",
				Code:    ValidationCodeBusinessRule,
				Message: "Due date is recommended for orders",
			})
		}
	case models.SaleTypeInvoice:
		// Invoices must have due date
		if request.DueDate.IsZero() {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "due_date",
				Code:    ValidationCodeRequired,
				Message: "Due date is required for invoices",
			})
		}
	}
}

func (svs *SalesValidationService) validateCreditLimit(customer *models.Contact, request models.SaleCreateRequest, result *SalesValidationResult) {
	if customer.CreditLimit <= 0 {
		return // No credit limit set
	}
	
	// Calculate total order amount
	totalAmount := svs.calculateOrderTotal(request)
	
	// Get current outstanding amount for customer
	var currentOutstanding float64
	svs.db.Model(&models.Sale{}).
		Where("customer_id = ? AND deleted_at IS NULL", customer.ID).
		Select("COALESCE(SUM(outstanding_amount), 0)").
		Scan(&currentOutstanding)
	
	// Check if new order would exceed credit limit
	if currentOutstanding+totalAmount > customer.CreditLimit {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "customer_id",
			Code:    ValidationCodeCreditLimit,
			Message: "Order would exceed customer credit limit",
			Context: map[string]interface{}{
				"credit_limit":        customer.CreditLimit,
				"current_outstanding": currentOutstanding,
				"order_amount":        totalAmount,
				"available_credit":    customer.CreditLimit - currentOutstanding,
			},
		})
	} else if currentOutstanding+totalAmount > customer.CreditLimit*0.8 {
		// Warning if close to credit limit
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "customer_id",
			Code:    ValidationCodeBusinessRule,
			Message: "Order approaching customer credit limit",
			Context: map[string]interface{}{
				"credit_limit":     customer.CreditLimit,
				"utilization":      fmt.Sprintf("%.1f%%", (currentOutstanding+totalAmount)/customer.CreditLimit*100),
			},
		})
	}
}

func (svs *SalesValidationService) calculateOrderTotal(request models.SaleCreateRequest) float64 {
	subtotal := 0.0
	
	for _, item := range request.Items {
		lineTotal := float64(item.Quantity) * item.UnitPrice
		lineTotal -= lineTotal * item.DiscountPercent / 100
		subtotal += lineTotal
	}
	
	// Apply order discount
	subtotal -= subtotal * request.DiscountPercent / 100
	
	// Add shipping
	subtotal += request.ShippingCost
	
	// Add taxes (simplified calculation)
	if request.PPNPercent != nil && *request.PPNPercent > 0 {
		subtotal += subtotal * (*request.PPNPercent / 100)
	}
	
	return subtotal
}

// Additional validation methods for different operations

func (svs *SalesValidationService) validatePaymentAmount(sale *models.Sale, amount float64, result *SalesValidationResult) {
	if amount <= 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "amount",
			Code:    ValidationCodeInvalidRange,
			Message: "Payment amount must be greater than 0",
			Value:   amount,
		})
		return
	}
	
	if amount > sale.OutstandingAmount {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "amount",
			Code:    ValidationCodeBusinessRule,
			Message: "Payment amount exceeds outstanding amount",
			Value:   amount,
			Context: map[string]interface{}{
				"outstanding_amount": sale.OutstandingAmount,
			},
		})
	}
}

func (svs *SalesValidationService) validateSaleStatusForPayment(status string, result *SalesValidationResult) {
	allowedStatuses := []string{models.SaleStatusInvoiced, models.SaleStatusOverdue, models.SaleStatusConfirmed}
	
	for _, allowedStatus := range allowedStatuses {
		if status == allowedStatus {
			return
		}
	}
	
	result.Errors = append(result.Errors, ValidationError{
		Field:   "status",
		Code:    ValidationCodeInvalidStatus,
		Message: "Sale status does not allow payments",
		Value:   status,
		Context: map[string]interface{}{
			"allowed_statuses": allowedStatuses,
		},
	})
}

func (svs *SalesValidationService) validatePaymentDate(paymentDate time.Time, result *SalesValidationResult) {
	if paymentDate.IsZero() {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "payment_date",
			Code:    ValidationCodeRequired,
			Message: "Payment date is required",
		})
		return
	}
	
	now := time.Now()
	if paymentDate.After(now.AddDate(0, 0, 1)) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "payment_date",
			Code:    ValidationCodeInvalidDate,
			Message: "Payment date cannot be more than 1 day in the future",
			Value:   paymentDate.Format("2006-01-02"),
		})
	}
	
	if paymentDate.Before(now.AddDate(-1, 0, 0)) {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "payment_date",
			Code:    ValidationCodeBusinessRule,
			Message: "Payment date is more than 1 year in the past",
			Value:   paymentDate.Format("2006-01-02"),
		})
	}
}

func (svs *SalesValidationService) validatePaymentMethod(method string, result *SalesValidationResult) {
	if method == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "payment_method",
			Code:    ValidationCodeRequired,
			Message: "Payment method is required",
		})
		return
	}
	
	validMethods := []string{"CASH", "BANK_TRANSFER", "CREDIT_CARD", "CHECK", "OTHER"}
	for _, validMethod := range validMethods {
		if method == validMethod {
			return
		}
	}
	
	result.Errors = append(result.Errors, ValidationError{
		Field:   "payment_method",
		Code:    ValidationCodeInvalidFormat,
		Message: "Invalid payment method",
		Value:   method,
		Context: map[string]interface{}{
			"allowed_methods": validMethods,
		},
	})
}

func (svs *SalesValidationService) validateCashBankAccount(cashBankID uint, result *SalesValidationResult) {
	// This would validate against cash_banks table
	// For now, just check if ID is valid
	var count int64
	svs.db.Model(&struct{ ID uint `gorm:"column:id"`}{}).
		Table("cash_banks").
		Where("id = ? AND deleted_at IS NULL", cashBankID).
		Count(&count)
	
	if count == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "cash_bank_id",
			Code:    ValidationCodeNotFound,
			Message: "Cash/Bank account not found",
			Value:   cashBankID,
		})
	}
}

// Helper methods

func (svs *SalesValidationService) canSaleBeUpdated(status string) bool {
	updatableStatuses := []string{models.SaleStatusDraft, models.SaleStatusPending}
	for _, updatableStatus := range updatableStatuses {
		if status == updatableStatus {
			return true
		}
	}
	return false
}

func (svs *SalesValidationService) validateDateUpdates(sale *models.Sale, request models.SaleUpdateRequest, result *SalesValidationResult) {
	saleDate := sale.Date
	if request.Date != nil {
		saleDate = *request.Date
	}
	
	dueDate := sale.DueDate
	if request.DueDate != nil {
		dueDate = *request.DueDate
	}
	
	validUntil := sale.ValidUntil
	if request.ValidUntil != nil {
		validUntil = request.ValidUntil
	}
	
	// Validate due date is after sale date
	if !dueDate.IsZero() && dueDate.Before(saleDate) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "due_date",
			Code:    ValidationCodeInvalidDate,
			Message: "Due date must be on or after sale date",
			Value:   dueDate.Format("2006-01-02"),
			Context: map[string]interface{}{
				"sale_date": saleDate.Format("2006-01-02"),
			},
		})
	}
	
	// Validate valid until date
	if validUntil != nil && validUntil.Before(saleDate) {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "valid_until",
			Code:    ValidationCodeInvalidDate,
			Message: "Valid until date must be on or after sale date",
			Value:   validUntil.Format("2006-01-02"),
		})
	}
}
