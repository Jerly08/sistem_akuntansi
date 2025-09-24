package controllers

import (
	"net/http"
	"strconv"

	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

// SSOTPaymentController handles payments using the Single Source of Truth posting service
type SSOTPaymentController struct {
	enhancedPaymentService *services.EnhancedPaymentServiceWithJournal
}

// NewSSOTPaymentController creates a new SSOT payment controller
func NewSSOTPaymentController(enhancedPaymentService *services.EnhancedPaymentServiceWithJournal) *SSOTPaymentController {
	return &SSOTPaymentController{
		enhancedPaymentService: enhancedPaymentService,
	}
}

// CreateReceivablePayment creates a customer payment with SSOT journal integration
func (ctrl *SSOTPaymentController) CreateReceivablePayment(c *gin.Context) {
	var req services.PaymentWithJournalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Set defaults for receivable payment
	req.Method = "RECEIVABLE"
	req.AutoCreateJournal = true

	// Get user ID from JWT context
	userID := getSSOTUserIDFromContext(c)
	req.UserID = userID

	// Process payment using SSOT enhanced service
	response, err := ctrl.enhancedPaymentService.CreatePaymentWithJournal(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create receivable payment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Receivable payment created successfully",
		"data":    response,
	})
}

// CreatePayablePayment creates a vendor payment with SSOT journal integration
func (ctrl *SSOTPaymentController) CreatePayablePayment(c *gin.Context) {
	var req services.PaymentWithJournalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Set defaults for payable payment
	req.Method = "PAYABLE"
	req.AutoCreateJournal = true

	// Get user ID from JWT context
	userID := getSSOTUserIDFromContext(c)
	req.UserID = userID

	// Process payment using SSOT enhanced service
	response, err := ctrl.enhancedPaymentService.CreatePaymentWithJournal(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create payable payment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Payable payment created successfully",
		"data":    response,
	})
}

// GetPaymentWithJournal retrieves a payment with its journal entry details
func (ctrl *SSOTPaymentController) GetPaymentWithJournal(c *gin.Context) {
	paymentIDStr := c.Param("id")
	paymentID, err := strconv.ParseUint(paymentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payment ID",
		})
		return
	}

	response, err := ctrl.enhancedPaymentService.GetPaymentWithJournal(uint(paymentID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Payment not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// ReversePayment reverses a payment and its journal entries
func (ctrl *SSOTPaymentController) ReversePayment(c *gin.Context) {
	paymentIDStr := c.Param("id")
	paymentID, err := strconv.ParseUint(paymentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payment ID",
		})
		return
	}

	var request struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	userID := getSSOTUserIDFromContext(c)
	
	response, err := ctrl.enhancedPaymentService.ReversePayment(uint(paymentID), request.Reason, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to reverse payment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment reversed successfully",
		"data":    response,
	})
}

// PreviewPaymentJournal previews what journal entry would be created for a payment
func (ctrl *SSOTPaymentController) PreviewPaymentJournal(c *gin.Context) {
	var req services.PaymentWithJournalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Get user ID from JWT context
	userID := getSSOTUserIDFromContext(c)
	req.UserID = userID

	// Preview journal entry
	preview, err := ctrl.enhancedPaymentService.PreviewPaymentJournal(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Failed to preview journal entry",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Journal preview generated successfully",
		"data":    preview,
	})
}

// GetAccountBalanceUpdates retrieves account balance updates from a payment's journal entry
func (ctrl *SSOTPaymentController) GetAccountBalanceUpdates(c *gin.Context) {
	paymentIDStr := c.Param("id")
	paymentID, err := strconv.ParseUint(paymentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid payment ID",
		})
		return
	}

	updates, err := ctrl.enhancedPaymentService.GetAccountBalanceUpdates(uint(paymentID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Account balance updates not found",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": updates,
	})
}

// GetPayments retrieves all payments (legacy compatibility method)
func (ctrl *SSOTPaymentController) GetPayments(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "This endpoint is deprecated. Use enhanced payment endpoints with journal integration.",
		"migration_notes": "Please migrate to use /api/v1/payments/ssot/* endpoints which provide full journal integration",
	})
}

// getSSOTUserIDFromContext extracts user ID from JWT context for SSOT payment controller
func getSSOTUserIDFromContext(c *gin.Context) uint {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(uint); ok {
			return id
		}
	}
	return 0
}
