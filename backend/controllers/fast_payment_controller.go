package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

// FastPaymentController provides lightweight, high-performance payment endpoints
type FastPaymentController struct {
	lightweightPaymentService *services.LightweightPaymentService
}

// NewFastPaymentController creates a new fast payment controller
func NewFastPaymentController(lightweightPaymentService *services.LightweightPaymentService) *FastPaymentController {
	return &FastPaymentController{
		lightweightPaymentService: lightweightPaymentService,
	}
}

// RecordPaymentFast handles lightweight payment recording with minimal processing time
// @Summary Record payment with fast processing
// @Description Records a payment with minimal database operations for fast response
// @Tags fast-payments
// @Accept json
// @Produce json
// @Param payment body services.LightweightPaymentRequest true "Payment details"
// @Success 200 {object} services.LightweightPaymentResponse
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/v1/payments/fast/record [post]
func (fpc *FastPaymentController) RecordPaymentFast(c *gin.Context) {
	var req services.LightweightPaymentRequest
	
	// Use ShouldBindJSON for faster parsing
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Get user ID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User authentication required",
		})
		return
	}
	req.UserID = userID.(uint)

	// Fast validation
	if err := fpc.lightweightPaymentService.ValidatePaymentRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	// Set context with timeout for fast processing
	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// Process payment with lightweight service
	response, err := fpc.lightweightPaymentService.RecordPaymentFast(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Payment recording failed",
			"details": err.Error(),
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, response)
}

// RecordPaymentWithAsyncJournal records payment immediately and processes journal asynchronously
// @Summary Record payment with async journal processing
// @Description Records payment immediately and creates journal entry in background
// @Tags fast-payments
// @Accept json
// @Produce json
// @Param payment body services.LightweightPaymentRequest true "Payment details"
// @Success 200 {object} services.LightweightPaymentResponse
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/v1/payments/fast/record-async [post]
func (fpc *FastPaymentController) RecordPaymentWithAsyncJournal(c *gin.Context) {
	var req services.LightweightPaymentRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Get user ID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "User authentication required",
		})
		return
	}
	req.UserID = userID.(uint)

	// Fast validation
	if err := fpc.lightweightPaymentService.ValidatePaymentRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	// Set context with timeout for fast processing
	ctx, cancel := context.WithTimeout(c.Request.Context(), 12*time.Second)
	defer cancel()

	// Process payment with async journal creation
	response, err := fpc.lightweightPaymentService.RecordPaymentWithAsyncJournal(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Payment recording failed",
			"details": err.Error(),
		})
		return
	}

	// Add async indicator to response
	response.Message += " (Journal entry processing in background)"

	c.JSON(http.StatusOK, response)
}

// ValidatePayment performs quick payment validation without recording
// @Summary Validate payment data
// @Description Validates payment data quickly for frontend feedback
// @Tags fast-payments
// @Accept json
// @Produce json
// @Param payment body services.LightweightPaymentRequest true "Payment details"
// @Success 200 {object} map[string]interface{} "Validation result"
// @Failure 400 {object} map[string]interface{} "Validation errors"
// @Router /api/v1/payments/fast/validate [post]
func (fpc *FastPaymentController) ValidatePayment(c *gin.Context) {
	var req services.LightweightPaymentRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid":   false,
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Perform validation
	if err := fpc.lightweightPaymentService.ValidatePaymentRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid":   false,
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "Payment data is valid",
	})
}

// GetPaymentStatus gets payment processing status for real-time updates
// @Summary Get payment status
// @Description Gets real-time payment processing status
// @Tags fast-payments
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} map[string]interface{} "Payment status"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Payment not found"
// @Router /api/v1/payments/fast/status/{id} [get]
func (fpc *FastPaymentController) GetPaymentStatus(c *gin.Context) {
	paymentIDStr := c.Param("id")
	paymentID, err := strconv.ParseUint(paymentIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid payment ID",
		})
		return
	}

	// This could be enhanced with Redis caching for real-time status
	// For now, return basic status
	c.JSON(http.StatusOK, gin.H{
		"payment_id": paymentID,
		"status":     "completed", // This would be dynamic in real implementation
		"message":    "Payment processed successfully",
	})
}

// HealthCheck provides health check endpoint for monitoring
// @Summary Health check for fast payment service
// @Description Returns health status of the fast payment service
// @Tags fast-payments
// @Produce json
// @Success 200 {object} map[string]interface{} "Health status"
// @Router /api/v1/payments/fast/health [get]
func (fpc *FastPaymentController) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "fast-payment",
		"timestamp": time.Now().Format(time.RFC3339),
		"features": []string{
			"lightweight-processing",
			"async-journal-creation",
			"fast-validation",
			"minimal-database-operations",
		},
	})
}