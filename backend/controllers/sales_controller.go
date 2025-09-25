package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/utils"

	"github.com/gin-gonic/gin"
)

type SalesController struct {
	salesService         *services.SalesService
	paymentService       *services.PaymentService
	salesPaymentService  *services.SalesPaymentService
	unifiedPaymentService *services.UnifiedSalesPaymentService // NEW: Single source of truth
}

func NewSalesController(salesService *services.SalesService, paymentService *services.PaymentService, salesPaymentService *services.SalesPaymentService, unifiedPaymentService *services.UnifiedSalesPaymentService) *SalesController {
	return &SalesController{
		salesService:          salesService,
		paymentService:        paymentService,
		salesPaymentService:   salesPaymentService,
		unifiedPaymentService: unifiedPaymentService, // NEW: Single source of truth
	}
}

// Sales Management

// GetSales gets all sales with pagination and filters
func (sc *SalesController) GetSales(c *gin.Context) {
	log.Printf("📋 Getting sales list with filters")
	
	// Parse and validate pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	// Validate pagination bounds
	if page < 1 {
		utils.SendValidationError(c, "Invalid pagination parameters", map[string]string{
			"page": "Page must be greater than 0",
		})
		return
	}
	if limit < 1 || limit > 100 {
		utils.SendValidationError(c, "Invalid pagination parameters", map[string]string{
			"limit": "Limit must be between 1 and 100",
		})
		return
	}
	
	// Get filter parameters
	status := c.Query("status")
	customerID := c.Query("customer_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	search := c.Query("search")

	filter := models.SalesFilter{
		Status:     status,
		CustomerID: customerID,
		StartDate:  startDate,
		EndDate:    endDate,
		Search:     search,
		Page:       page,
		Limit:      limit,
	}

	log.Printf("🔍 Fetching sales with filters: page=%d, limit=%d, status=%s, customer_id=%s", 
		page, limit, status, customerID)
	
	result, err := sc.salesService.GetSales(filter)
	if err != nil {
		log.Printf("❌ Failed to get sales: %v", err)
		utils.SendInternalError(c, "Failed to retrieve sales data", err.Error())
		return
	}

	log.Printf("✅ Retrieved %d sales (total: %d)", len(result.Data), result.Total)
	
	// Send paginated success response
	utils.SendPaginatedSuccess(c, 
		"Sales retrieved successfully", 
		result.Data, 
		result.Page, 
		result.Limit, 
		result.Total)
}

// GetSale gets a single sale by ID
func (sc *SalesController) GetSale(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		log.Printf("❌ Invalid sale ID parameter: %v", err)
		utils.SendValidationError(c, "Invalid sale ID", map[string]string{
			"id": "Sale ID must be a valid positive number",
		})
		return
	}

	log.Printf("🔍 Getting sale details for ID: %d", id)
	
	sale, err := sc.salesService.GetSaleByID(uint(id))
	if err != nil {
		log.Printf("❌ Sale %d not found: %v", id, err)
		utils.SendSaleNotFound(c, uint(id))
		return
	}

	log.Printf("✅ Retrieved sale %d details successfully", id)
	utils.SendSuccess(c, "Sale retrieved successfully", sale)
}

// CreateSale creates a new sale
func (sc *SalesController) CreateSale(c *gin.Context) {
	log.Printf("🎆 Creating new sale")
	
	var request models.SaleCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("❌ Invalid sale creation request: %v", err)
		utils.SendValidationError(c, "Invalid sale data", map[string]string{
			"request": "Please check the request format and required fields",
		})
		return
	}

	// Get user ID from context with error handling
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		log.Printf("❌ User authentication missing for sale creation")
		utils.SendUnauthorized(c, "User authentication required")
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		log.Printf("❌ Invalid user ID type: %T", userIDInterface)
		utils.SendUnauthorized(c, "Invalid user authentication")
		return
	}

	log.Printf("📄 Creating sale for customer %d by user %d", request.CustomerID, userID)
	
	sale, err := sc.salesService.CreateSale(request, userID)
	if err != nil {
		log.Printf("❌ Failed to create sale: %v", err)
		
		// Handle specific error types
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "customer not found"):
			utils.SendNotFound(c, "Customer not found")
		case strings.Contains(errorMsg, "validation"):
			utils.SendValidationError(c, "Sale validation failed", map[string]string{
				"details": errorMsg,
			})
		case strings.Contains(errorMsg, "inventory"):
			utils.SendBusinessRuleError(c, "Inventory validation failed", map[string]interface{}{
				"details": errorMsg,
			})
		default:
			utils.SendInternalError(c, "Failed to create sale", errorMsg)
		}
		return
	}

	log.Printf("✅ Sale created successfully: ID=%d, Code=%s", sale.ID, sale.Code)
	utils.SendCreated(c, "Sale created successfully", sale)
}

// UpdateSale updates an existing sale
func (sc *SalesController) UpdateSale(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale ID"})
		return
	}

	var request models.SaleUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uint)

	sale, err := sc.salesService.UpdateSale(uint(id), request, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sale)
}

// DeleteSale deletes a sale
func (sc *SalesController) DeleteSale(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale ID"})
		return
	}

	// Get user role from context for permission checking
	userRole, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
		return
	}
	roleStr := userRole.(string)

	if err := sc.salesService.DeleteSaleWithRole(uint(id), roleStr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sale deleted successfully"})
}

// Sales Status Management

// ConfirmSale confirms a sale (changes status to CONFIRMED)
func (sc *SalesController) ConfirmSale(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale ID"})
		return
	}

	userID := c.MustGet("user_id").(uint)

	if err := sc.salesService.ConfirmSale(uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sale confirmed successfully"})
}

// InvoiceSale creates an invoice from a sale
func (sc *SalesController) InvoiceSale(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale ID"})
		return
	}

	userID := c.MustGet("user_id").(uint)

	invoice, err := sc.salesService.CreateInvoiceFromSale(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoice)
}

// CancelSale cancels a sale
func (sc *SalesController) CancelSale(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale ID"})
		return
	}

	var request struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&request)

	userID := c.MustGet("user_id").(uint)

	if err := sc.salesService.CancelSale(uint(id), request.Reason, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sale cancelled successfully"})
}

// Payment Management

// GetSalePayments gets all payments for a sale
func (sc *SalesController) GetSalePayments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale ID"})
		return
	}

	payments, err := sc.salesService.GetSalePayments(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payments)
}

// CreateSalePayment creates a payment for a sale with proper race condition protection
func (sc *SalesController) CreateSalePayment(c *gin.Context) {
	log.Printf("🚀 Starting payment creation process")
	
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		log.Printf("❌ Invalid sale ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Invalid sale ID",
			"details": err.Error(),
			"code":    "INVALID_SALE_ID",
		})
		return
	}

	var request models.SalePaymentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("❌ Invalid request data for sale %d: %v", id, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":           "error",
			"error":            "Invalid request data",
			"details":          err.Error(),
			"code":             "VALIDATION_ERROR",
			"validation_error": true,
		})
		return
	}

	// Set the sale ID from the URL parameter
	request.SaleID = uint(id)
	log.Printf("💰 Processing payment request for sale %d: amount=%.2f, method=%s", 
		id, request.Amount, request.PaymentMethod)

	// Get user ID from context with proper error handling
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		log.Printf("❌ User authentication missing for sale %d payment", id)
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"error":   "User not authenticated",
			"details": "user_id not found in context",
			"code":    "AUTH_MISSING",
		})
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		log.Printf("❌ Invalid user ID type for sale %d payment: %T", id, userIDInterface)
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"error":   "Invalid user authentication",
			"details": "user_id has invalid type",
			"code":    "AUTH_INVALID",
		})
		return
	}

	// Validate payment request before processing
	if err := sc.unifiedPaymentService.ValidatePaymentRequest(request); err != nil {
		log.Printf("❌ Payment validation failed for sale %d: %v", id, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"error":   "Payment validation failed",
			"details": err.Error(),
			"code":    "PAYMENT_VALIDATION_ERROR",
		})
		return
	}

	// Use the UNIFIED payment service (SINGLE SOURCE OF TRUTH)
	payment, err := sc.unifiedPaymentService.CreateSalesPayment(uint(id), request, userID)
	if err != nil {
		log.Printf("❌ [UNIFIED] Payment creation failed for sale %d: %v", id, err)
		
		// Determine appropriate HTTP status based on error type
		status := http.StatusInternalServerError
		code := "PAYMENT_CREATION_ERROR"
		
		errorMsg := err.Error()
		switch {
		case strings.Contains(errorMsg, "not found"):
			status = http.StatusNotFound
			code = "SALE_NOT_FOUND"
		case strings.Contains(errorMsg, "exceeds outstanding"):
			status = http.StatusBadRequest
			code = "AMOUNT_EXCEEDS_OUTSTANDING"
		case strings.Contains(errorMsg, "cannot receive payments"):
			status = http.StatusBadRequest
			code = "INVALID_SALE_STATUS"
		case strings.Contains(errorMsg, "no outstanding amount"):
			status = http.StatusBadRequest
			code = "NO_OUTSTANDING_AMOUNT"
		case strings.Contains(errorMsg, "validation"):
			status = http.StatusBadRequest
			code = "VALIDATION_ERROR"
		}
		
		c.JSON(status, gin.H{
			"status":   "error",
			"error":    "Failed to create payment",
			"details":  errorMsg,
			"code":     code,
			"sale_id":  id,
			"user_id":  userID,
		})
		return
	}

	log.Printf("✅ Payment created successfully for sale %d: payment_id=%d, amount=%.2f", 
		id, payment.ID, payment.Amount)

	// Return success response with comprehensive data
	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Payment created successfully with race condition protection",
		"data":    payment,
		"meta": gin.H{
			"sale_id":    payment.SaleID,
			"payment_id": payment.ID,
			"user_id":    userID,
			"created_at": payment.CreatedAt,
		},
	})
}

// Integrated Payment Management - uses Payment Service for comprehensive payment tracking

// CreateIntegratedPayment creates payment via Payment Management for better tracking
func (sc *SalesController) CreateIntegratedPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale ID"})
		return
	}

	var request struct {
		Amount        float64   `json:"amount" binding:"required,min=0"`
		Date          time.Time `json:"date" binding:"required"`
		Method        string    `json:"method" binding:"required"`
		CashBankID    uint      `json:"cash_bank_id" binding:"required"`
		Reference     string    `json:"reference"`
		Notes         string    `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("Payment creation validation error for sale %d: %v", id, err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
			"expected_fields": map[string]string{
				"amount": "number (required, min=0)",
				"date": "datetime string (required, ISO format)",
				"method": "string (required)",
				"cash_bank_id": "number (required)",
				"reference": "string (optional)",
				"notes": "string (optional)",
			},
		})
		return
	}

	// Log successful request parsing
	log.Printf("Received integrated payment request for sale %d: amount=%.2f, method=%s, cash_bank_id=%d", id, request.Amount, request.Method, request.CashBankID)
	
	// Get sale details to validate and get customer ID
	sale, err := sc.salesService.GetSaleByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Sale not found",
			"details": err.Error(),
		})
		return
	}

	// Validate sale status
	if sale.Status != models.SaleStatusInvoiced && sale.Status != models.SaleStatusOverdue {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Sale must be invoiced to receive payments",
			"sale_status": sale.Status,
		})
		return
	}

	// Validate payment amount
	if request.Amount > sale.OutstandingAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Payment amount exceeds outstanding amount",
			"outstanding_amount": sale.OutstandingAmount,
			"requested_amount": request.Amount,
		})
		return
	}

	// Get user ID with error handling
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		log.Printf("Error: user_id not found in context for sale %d payment", id)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"details": "user_id not found in context",
		})
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		log.Printf("Error: user_id has invalid type for sale %d payment: %T", id, userIDInterface)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid user authentication",
			"details": "user_id has invalid type",
		})
		return
	}

	// Create payment request for Payment Management service
	paymentRequest := services.PaymentCreateRequest{
		ContactID:   sale.CustomerID,
		CashBankID:  request.CashBankID,
		Date:        request.Date,
		Amount:      request.Amount,
		Method:      request.Method,
		Reference:   request.Reference,
		Notes:       fmt.Sprintf("Payment for Invoice %s - %s", sale.InvoiceNumber, request.Notes),
		Allocations: []services.InvoiceAllocation{
			{
				InvoiceID: uint(id),
				Amount:    request.Amount,
			},
		},
	}

	// Use Payment Management service (needs to be injected)
	log.Printf("Calling PaymentService.CreateReceivablePayment for sale %d with amount %.2f", id, paymentRequest.Amount)
	payment, err := sc.paymentService.CreateReceivablePayment(paymentRequest, userID)
	if err != nil {
		log.Printf("Error in CreateReceivablePayment for sale %d: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": "Failed to create payment",
			"details": err.Error(),
			"status": "error",
		})
		return
	}
	log.Printf("✅ Payment created successfully: ID=%d, Code=%s", payment.ID, payment.Code)

	// Return response with both payment info and updated sale status
	updatedSale, err := sc.salesService.GetSaleByID(uint(id))
	if err != nil {
		// If we can't get updated sale info, still return success but with basic info
		log.Printf("Warning: Could not fetch updated sale info after payment creation: %v", err)
		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"payment": payment,
			"message": "Payment created successfully via Payment Management",
			"note": "Payment created but updated sale info unavailable",
			"status": "success",
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"payment": payment,
		"updated_sale": gin.H{
			"id": updatedSale.ID,
			"status": updatedSale.Status,
			"paid_amount": updatedSale.PaidAmount,
			"outstanding_amount": updatedSale.OutstandingAmount,
		},
		"message": "Payment created successfully via Payment Management",
		"status": "success",
	})
}

// GetSaleForPayment gets sale details formatted for payment creation
func (sc *SalesController) GetSaleForPayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale ID"})
		return
	}

	sale, err := sc.salesService.GetSaleByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sale not found"})
		return
	}

	// Format response for payment creation
	response := gin.H{
		"sale_id": sale.ID,
		"invoice_number": sale.InvoiceNumber,
		"customer": gin.H{
			"id": sale.Customer.ID,
			"name": sale.Customer.Name,
			"type": sale.Customer.Type,
		},
		"total_amount": sale.TotalAmount,
		"paid_amount": sale.PaidAmount,
		"outstanding_amount": sale.OutstandingAmount,
		"status": sale.Status,
		"date": sale.Date.Format("2006-01-02"),
		"due_date": sale.DueDate.Format("2006-01-02"),
		"can_receive_payment": sale.Status == models.SaleStatusInvoiced || sale.Status == models.SaleStatusOverdue,
		"payment_url_suggestion": fmt.Sprintf("/api/sales/%d/integrated-payment", sale.ID),
	}

	c.JSON(http.StatusOK, response)
}

// Sales Returns

// CreateSaleReturn creates a return for a sale
func (sc *SalesController) CreateSaleReturn(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale ID"})
		return
	}

	var request models.SaleReturnRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uint)

	saleReturn, err := sc.salesService.CreateSaleReturn(uint(id), request, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, saleReturn)
}

// GetSaleReturns gets all returns
func (sc *SalesController) GetSaleReturns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	returns, err := sc.salesService.GetSaleReturns(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, returns)
}

// Reporting and Analytics

// GetSalesSummary gets sales summary statistics
func (sc *SalesController) GetSalesSummary(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	summary, err := sc.salesService.GetSalesSummary(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetSalesAnalytics gets sales analytics data
func (sc *SalesController) GetSalesAnalytics(c *gin.Context) {
	period := c.DefaultQuery("period", "monthly")
	year := c.DefaultQuery("year", "2024")

	analytics, err := sc.salesService.GetSalesAnalytics(period, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetReceivablesReport gets accounts receivable report
func (sc *SalesController) GetReceivablesReport(c *gin.Context) {
	receivables, err := sc.salesService.GetReceivablesReport()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, receivables)
}

// PDF Export

// ExportSaleInvoicePDF exports sale invoice as PDF
func (sc *SalesController) ExportSaleInvoicePDF(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale ID"})
		return
	}

	pdfData, filename, err := sc.salesService.ExportInvoicePDF(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdfData)
}

// ExportSalesReportPDF exports sales report as PDF
func (sc *SalesController) ExportSalesReportPDF(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// If no dates provided, default to last 30 days to keep report size reasonable
	if startDate == "" && endDate == "" {
		end := time.Now()
		start := end.AddDate(0, 0, -30)
		startDate = start.Format("2006-01-02")
		endDate = end.Format("2006-01-02")
	}

	pdfData, filename, err := sc.salesService.ExportSalesReportPDF(startDate, endDate)
	if err != nil {
		log.Printf("❌ Error generating sales report PDF (start=%s, end=%s): %v", startDate, endDate, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate sales report PDF", "details": err.Error()})
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/pdf", pdfData)
}


// Customer Portal

// GetCustomerSales gets sales for a specific customer (for customer portal)
func (sc *SalesController) GetCustomerSales(c *gin.Context) {
	customerID, err := strconv.ParseUint(c.Param("customer_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	sales, err := sc.salesService.GetCustomerSales(uint(customerID), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sales)
}

// GetCustomerInvoices gets invoices for a specific customer
func (sc *SalesController) GetCustomerInvoices(c *gin.Context) {
	customerID, err := strconv.ParseUint(c.Param("customer_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	invoices, err := sc.salesService.GetCustomerInvoices(uint(customerID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoices)
}
