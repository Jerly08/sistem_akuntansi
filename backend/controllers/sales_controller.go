package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"

	"github.com/gin-gonic/gin"
)

type SalesController struct {
	salesService *services.SalesService
}

func NewSalesController(salesService *services.SalesService) *SalesController {
	return &SalesController{
		salesService: salesService,
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

	if err := sc.salesService.DeleteSale(uint(id)); err != nil {
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
