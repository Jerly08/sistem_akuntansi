package controllers

import (
	"net/http"
	"time"
	"app-sistem-akuntansi/utils"

	"github.com/gin-gonic/gin"
)

type SalesValidationController struct {
	dateUtils *utils.DateUtils
}

func NewSalesValidationController() *SalesValidationController {
	return &SalesValidationController{
		dateUtils: utils.NewDateUtils(),
	}
}

// ValidatePaymentTermsRequest represents the request for payment terms validation
type ValidatePaymentTermsRequest struct {
	InvoiceDate  string `json:"invoice_date" binding:"required"`
	PaymentTerms string `json:"payment_terms" binding:"required"`
}

// ValidatePaymentTermsResponse represents the response for payment terms validation
type ValidatePaymentTermsResponse struct {
	Valid                bool   `json:"valid"`
	DueDate             string `json:"due_date"`
	DueDateFormatted    string `json:"due_date_formatted"`
	PaymentDescription  string `json:"payment_description"`
	DaysFromInvoice     int    `json:"days_from_invoice"`
	IsBusinessDay       bool   `json:"is_business_day"`
	Error               string `json:"error,omitempty"`
}

// CalculateDueDateRequest represents the request for due date calculation
type CalculateDueDateRequest struct {
	InvoiceDate  string `json:"invoice_date" binding:"required"`
	PaymentTerms string `json:"payment_terms" binding:"required"`
}

// CalculateDueDateResponse represents the response for due date calculation
type CalculateDueDateResponse struct {
	InvoiceDate         string `json:"invoice_date"`
	InvoiceDateFormatted string `json:"invoice_date_formatted"`
	PaymentTerms        string `json:"payment_terms"`
	DueDate             string `json:"due_date"`
	DueDateFormatted    string `json:"due_date_formatted"`
	DaysFromInvoice     int    `json:"days_from_invoice"`
	PaymentDescription  string `json:"payment_description"`
	IsBusinessDay       bool   `json:"is_business_day"`
	BusinessDayAdjusted string `json:"business_day_adjusted,omitempty"`
}

// PaymentTermsOptionsResponse represents available payment terms
type PaymentTermsOptionsResponse struct {
	Options []map[string]string `json:"options"`
}

// ValidatePaymentTerms validates payment terms and calculates due date
func (c *SalesValidationController) ValidatePaymentTerms(ctx *gin.Context) {
	var request ValidatePaymentTermsRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Parse invoice date
	invoiceDate, err := time.Parse("2006-01-02", request.InvoiceDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ValidatePaymentTermsResponse{
			Valid: false,
			Error: "Invalid invoice date format. Use YYYY-MM-DD",
		})
		return
	}

	// Validate payment terms
	if err := c.dateUtils.ValidatePaymentTerms(request.PaymentTerms); err != nil {
		ctx.JSON(http.StatusBadRequest, ValidatePaymentTermsResponse{
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	// Calculate due date
	dueDate := c.dateUtils.CalculateDueDateFromPaymentTerms(invoiceDate, request.PaymentTerms)
	
	// Calculate days difference
	daysFromInvoice := int(dueDate.Sub(invoiceDate).Hours() / 24)

	response := ValidatePaymentTermsResponse{
		Valid:               true,
		DueDate:             c.dateUtils.FormatDateForAPI(dueDate),
		DueDateFormatted:    c.dateUtils.FormatDateForIndonesia(dueDate),
		PaymentDescription:  c.dateUtils.GetPaymentTermsDescription(request.PaymentTerms, invoiceDate),
		DaysFromInvoice:     daysFromInvoice,
		IsBusinessDay:       c.dateUtils.IsBusinessDay(dueDate),
	}

	ctx.JSON(http.StatusOK, response)
}

// CalculateDueDate calculates due date with detailed information
func (c *SalesValidationController) CalculateDueDate(ctx *gin.Context) {
	var request CalculateDueDateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Parse invoice date
	invoiceDate, err := time.Parse("2006-01-02", request.InvoiceDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid invoice date format. Use YYYY-MM-DD",
		})
		return
	}

	// Validate payment terms
	if err := c.dateUtils.ValidatePaymentTerms(request.PaymentTerms); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Calculate due date
	dueDate := c.dateUtils.CalculateDueDateFromPaymentTerms(invoiceDate, request.PaymentTerms)
	
	// Calculate days difference
	daysFromInvoice := int(dueDate.Sub(invoiceDate).Hours() / 24)

	response := CalculateDueDateResponse{
		InvoiceDate:         c.dateUtils.FormatDateForAPI(invoiceDate),
		InvoiceDateFormatted: c.dateUtils.FormatDateForIndonesia(invoiceDate),
		PaymentTerms:        request.PaymentTerms,
		DueDate:             c.dateUtils.FormatDateForAPI(dueDate),
		DueDateFormatted:    c.dateUtils.FormatDateForIndonesia(dueDate),
		DaysFromInvoice:     daysFromInvoice,
		PaymentDescription:  c.dateUtils.GetPaymentTermsDescription(request.PaymentTerms, invoiceDate),
		IsBusinessDay:       c.dateUtils.IsBusinessDay(dueDate),
	}

	// If due date falls on weekend, provide business day adjusted date
	if !response.IsBusinessDay {
		adjustedDate := c.dateUtils.AdjustToBusinessDay(dueDate)
		response.BusinessDayAdjusted = c.dateUtils.FormatDateForAPI(adjustedDate)
	}

	ctx.JSON(http.StatusOK, response)
}

// GetPaymentTermsOptions returns available payment terms options
func (c *SalesValidationController) GetPaymentTermsOptions(ctx *gin.Context) {
	options := c.dateUtils.GetPaymentTermsOptions()
	
	response := PaymentTermsOptionsResponse{
		Options: options,
	}

	ctx.JSON(http.StatusOK, response)
}

// ValidateInvoiceDate validates if the invoice date is reasonable
func (c *SalesValidationController) ValidateInvoiceDate(ctx *gin.Context) {
	dateStr := ctx.Query("date")
	if dateStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Date parameter is required",
		})
		return
	}

	invoiceDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid date format. Use YYYY-MM-DD",
		})
		return
	}

	now := time.Now()
	
	// Check if date is too far in the past (more than 1 year)
	oneYearAgo := now.AddDate(-1, 0, 0)
	if invoiceDate.Before(oneYearAgo) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invoice date cannot be more than 1 year in the past",
			"valid": false,
		})
		return
	}

	// Check if date is too far in the future (more than 30 days)
	thirtyDaysFromNow := now.AddDate(0, 0, 30)
	if invoiceDate.After(thirtyDaysFromNow) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invoice date cannot be more than 30 days in the future",
			"valid": false,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"valid": true,
		"date": c.dateUtils.FormatDateForAPI(invoiceDate),
		"date_formatted": c.dateUtils.FormatDateForIndonesia(invoiceDate),
		"is_business_day": c.dateUtils.IsBusinessDay(invoiceDate),
	})
}
