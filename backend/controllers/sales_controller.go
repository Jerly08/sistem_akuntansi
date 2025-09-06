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

	"github.com/gin-gonic/gin"
)

type SalesController struct {
	salesService  *services.SalesService
	paymentService *services.PaymentService
}

func NewSalesController(salesService *services.SalesService, paymentService *services.PaymentService) *SalesController {
	return &SalesController{
		salesService:  salesService,
		paymentService: paymentService,
	}
}

// Sales Management

// GetSales gets all sales with pagination and filters
func (sc *SalesController) GetSales(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
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

	result, err := sc.salesService.GetSales(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetSale gets a single sale by ID
func (sc *SalesController) GetSale(c *gin.Context) {
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

	c.JSON(http.StatusOK, sale)
}

// CreateSale creates a new sale
func (sc *SalesController) CreateSale(c *gin.Context) {
	var request models.SaleCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID := c.MustGet("user_id").(uint)

	sale, err := sc.salesService.CreateSale(request, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sale)
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

// CreateSalePayment creates a payment for a sale
func (sc *SalesController) CreateSalePayment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sale ID",
			"details": err.Error(),
		})
		return
	}

	var request models.SalePaymentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
			"validation_error": true,
		})
		return
	}

	// Set the sale ID from the URL parameter
	request.SaleID = uint(id)

	// Get user ID from context - handle potential panic
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"details": "user_id not found in context",
		})
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid user authentication",
			"details": "user_id has invalid type",
		})
		return
	}

	payment, err := sc.salesService.CreateSalePayment(uint(id), request, userID)
	if err != nil {
		// Determine appropriate HTTP status based on error
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "status") || strings.Contains(err.Error(), "validation") || strings.Contains(err.Error(), "exceeds") {
			status = http.StatusBadRequest
		}
		
		c.JSON(status, gin.H{
			"error": "Failed to create payment",
			"details": err.Error(),
			"sale_id": id,
		})
		return
	}

	c.JSON(http.StatusCreated, payment)
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
			"error": "Failed to create payment",
			"details": err.Error(),
		})
		return
	}
	log.Printf("Payment created successfully: ID=%d, Code=%s", payment.ID, payment.Code)

	// Return response with both payment info and updated sale status
	updatedSale, err := sc.salesService.GetSaleByID(uint(id))
	if err != nil {
		// If we can't get updated sale info, still return success but with basic info
		log.Printf("Warning: Could not fetch updated sale info after payment creation: %v", err)
		c.JSON(http.StatusCreated, gin.H{
			"payment": payment,
			"message": "Payment created successfully via Payment Management",
			"note": "Payment created but updated sale info unavailable",
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"payment": payment,
		"updated_sale": gin.H{
			"id": updatedSale.ID,
			"status": updatedSale.Status,
			"paid_amount": updatedSale.PaidAmount,
			"outstanding_amount": updatedSale.OutstandingAmount,
		},
		"message": "Payment created successfully via Payment Management",
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

	pdfData, filename, err := sc.salesService.ExportSalesReportPDF(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
