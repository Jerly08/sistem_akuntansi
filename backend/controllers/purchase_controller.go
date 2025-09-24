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

type PurchaseController struct {
	purchaseService *services.PurchaseService
	paymentService  *services.PaymentService
}

func NewPurchaseController(purchaseService *services.PurchaseService, paymentService *services.PaymentService) *PurchaseController {
	return &PurchaseController{
		purchaseService: purchaseService,
		paymentService:  paymentService,
	}
}

// Purchase CRUD Operations

// GetPurchases returns paginated list of purchases with filters
func (pc *PurchaseController) GetPurchases(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	vendorID := c.Query("vendor_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	search := c.Query("search")
	approvalStatus := c.Query("approval_status")
	
	var requiresApproval *bool
	if reqApp := c.Query("requires_approval"); reqApp != "" {
		val := reqApp == "true"
		requiresApproval = &val
	}

	filter := models.PurchaseFilter{
		Status:           status,
		VendorID:         vendorID,
		StartDate:        startDate,
		EndDate:          endDate,
		Search:           search,
		ApprovalStatus:   approvalStatus,
		RequiresApproval: requiresApproval,
		Page:             page,
		Limit:            limit,
	}

	result, err := pc.purchaseService.GetPurchases(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetPurchase returns a single purchase by ID
func (pc *PurchaseController) GetPurchase(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	purchase, err := pc.purchaseService.GetPurchaseByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase not found"})
		return
	}

	c.JSON(http.StatusOK, purchase)
}

// CreatePurchase creates a new purchase request
func (pc *PurchaseController) CreatePurchase(c *gin.Context) {
	var request models.PurchaseCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Debug log the incoming request
	fmt.Printf("[DEBUG] CreatePurchase - Incoming request: %+v\n", request)
	for i, item := range request.Items {
		fmt.Printf("[DEBUG] Item %d: ProductID=%d, Qty=%d, Price=%.2f\n", i, item.ProductID, item.Quantity, item.UnitPrice)
	}

	userID := c.MustGet("user_id").(uint)

	purchase, err := pc.purchaseService.CreatePurchase(request, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, purchase)
}

// UpdatePurchase updates an existing purchase
func (pc *PurchaseController) UpdatePurchase(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	var request models.PurchaseUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uint)

	purchase, err := pc.purchaseService.UpdatePurchase(uint(id), request, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, purchase)
}

// DeletePurchase deletes a purchase
func (pc *PurchaseController) DeletePurchase(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	// Get user role from context
	userRole := c.MustGet("user_role").(string)
	
	// Check if purchase exists and get its status
	purchase, err := pc.purchaseService.GetPurchaseByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase not found"})
		return
	}
	
	// Check if user has permission to delete this purchase
	// For APPROVED purchases, only ADMIN can delete
	if purchase.Status == models.PurchaseStatusApproved && userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admin can delete approved purchases"})
		return
	}
	
	// For other non-draft purchases, admin and director can delete
	if purchase.Status != models.PurchaseStatusDraft && purchase.Status != models.PurchaseStatusApproved && userRole != "admin" && userRole != "director" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to delete this purchase"})
		return
	}

	err = pc.purchaseService.DeletePurchase(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Purchase deleted successfully"})
}

// Approval Operations

// SubmitForApproval submits a purchase for approval
func (pc *PurchaseController) SubmitForApproval(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	userID := c.MustGet("user_id").(uint)

	err = pc.purchaseService.SubmitForApproval(uint(id), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Purchase submitted for approval"})
}

// ApprovePurchase approves a purchase
func (pc *PurchaseController) ApprovePurchase(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	userRole := c.MustGet("user_role").(string)
	
	// SAFETY CHECK: Ensure purchase is not in DRAFT status
	purchase, err := pc.purchaseService.GetPurchaseByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase not found"})
		return
	}
	
	if purchase.Status == models.PurchaseStatusDraft {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot approve DRAFT purchase - please submit for approval first",
			"current_status": purchase.Status,
			"required_action": "Use 'Submit for Approval' button first",
		})
		return
	}

	// Parse request body to check for escalation
	var request struct {
		Comments            string `json:"comments"`
		EscalateToDirector  bool   `json:"escalate_to_director"`
	}
	c.ShouldBindJSON(&request)

	// Process approval with escalation logic
	result, err := pc.purchaseService.ProcessPurchaseApprovalWithEscalation(
		uint(id), 
		true, 
		userID, 
		userRole,
		request.Comments,
		request.EscalateToDirector,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RejectPurchase rejects a purchase
func (pc *PurchaseController) RejectPurchase(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	userID := c.MustGet("user_id").(uint)
	userRole := c.MustGet("user_role").(string)

	// Parse request body to get comments
	var request struct {
		Comments string `json:"comments" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Comments are required for rejection: " + err.Error()})
		return
	}

	// Process rejection with escalation logic (similar to approve but with rejection)
	result, err := pc.purchaseService.ProcessPurchaseApprovalWithEscalation(
		uint(id), 
		false, // false = reject
		userID, 
		userRole,
		request.Comments,
		false, // no escalation for rejection
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Receipt Operations

// CreatePurchaseReceipt creates a new purchase receipt
func (pc *PurchaseController) CreatePurchaseReceipt(c *gin.Context) {
	var request models.PurchaseReceiptRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Get user ID with proper error handling
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		log.Printf("Error: user_id not found in context for receipt creation")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"details": "user_id not found in context",
		})
		return
	}
	
	userID, ok := userIDInterface.(uint)
	if !ok {
		log.Printf("Error: user_id has invalid type for receipt creation: %T", userIDInterface)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid user authentication",
			"details": "user_id has invalid type",
		})
		return
	}
	
	log.Printf("Creating receipt for purchase %d with user ID %d", request.PurchaseID, userID)

	receipt, err := pc.purchaseService.CreatePurchaseReceipt(request, userID)
	if err != nil {
		log.Printf("Error creating receipt: %v", err)
		
		// Check for specific error types
		errorMessage := err.Error()
		if strings.Contains(errorMessage, "fk_purchase_receipts_receiver") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid user reference",
				"details": "The user ID in your session is not valid. Please log out and log in again.",
				"user_id": userID,
			})
			return
		}
		
		if strings.Contains(errorMessage, "foreign key constraint") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Database constraint violation",
				"details": "There is a reference issue in the database. Please contact administrator.",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create receipt",
			"details": err.Error(),
		})
		return
	}
	
	log.Printf("Receipt created successfully: ID=%d, Number=%s", receipt.ID, receipt.ReceiptNumber)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"receipt": receipt,
		"message": "Receipt created successfully",
	})
}

// GetPurchaseReceipts returns receipts for a purchase
func (pc *PurchaseController) GetPurchaseReceipts(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	// Check if only completed receipts are requested
	completedOnly := c.Query("completed_only") == "true"

	var receipts []models.PurchaseReceipt
	if completedOnly {
		// Get only completed receipts
		receipts, err = pc.purchaseService.GetCompletedPurchaseReceipts(uint(id))
	} else {
		// Get all receipts
		receipts, err = pc.purchaseService.GetPurchaseReceipts(uint(id))
	}
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": receipts,
		"count": len(receipts),
		"completed_only": completedOnly,
	})
}

// GetReceiptPDF generates PDF for a specific receipt
func (pc *PurchaseController) GetReceiptPDF(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("receipt_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid receipt ID"})
		return
	}

	// Generate PDF
	pdfBytes, receipt, err := pc.purchaseService.GenerateReceiptPDF(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set headers for PDF download
	filename := fmt.Sprintf("receipt_%s.pdf", receipt.ReceiptNumber)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	// Send PDF
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// GetAllReceiptsPDF generates combined PDF for all receipts of a purchase
func (pc *PurchaseController) GetAllReceiptsPDF(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	// Generate combined PDF
	pdfBytes, purchase, err := pc.purchaseService.GenerateAllReceiptsPDF(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Set headers for PDF download
	filename := fmt.Sprintf("receipts_%s.pdf", purchase.Code)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))

	// Send PDF
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// Document Operations

// UploadDocument uploads a document for a purchase
func (pc *PurchaseController) UploadDocument(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	// Handle file upload
	file, header, err := c.Request.FormFile("document")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	documentType := c.PostForm("document_type")
	if documentType == "" {
		documentType = models.PurchaseDocumentInvoice
	}

	userID := c.MustGet("user_id").(uint)

	// In a real implementation, you would save the file to storage
	// For now, we'll simulate the file path
	filePath := "/uploads/purchases/" + header.Filename

	err = pc.purchaseService.UploadDocument(
		uint(id),
		documentType,
		header.Filename,
		filePath,
		header.Size,
		header.Header.Get("Content-Type"),
		userID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Document uploaded successfully"})
}

// GetPurchaseDocuments returns documents for a purchase
func (pc *PurchaseController) GetPurchaseDocuments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	documents, err := pc.purchaseService.GetPurchaseDocuments(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, documents)
}

// DeleteDocument deletes a purchase document
func (pc *PurchaseController) DeleteDocument(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("document_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	err = pc.purchaseService.DeleteDocument(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}

// Three-way Matching Operations

// GetPurchaseMatching returns matching data for three-way matching
func (pc *PurchaseController) GetPurchaseMatching(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	matching, err := pc.purchaseService.GetPurchaseMatching(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, matching)
}

// ValidateThreeWayMatching validates three-way matching
func (pc *PurchaseController) ValidateThreeWayMatching(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	isValid, err := pc.purchaseService.ValidateThreeWayMatching(uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    err.Error(),
			"is_valid": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_valid": isValid,
		"message":  "Three-way matching validation completed",
	})
}

// Analytics and Reporting Operations

// GetPurchasesSummary returns purchase summary statistics
func (pc *PurchaseController) GetPurchasesSummary(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	summary, err := pc.purchaseService.GetPurchasesSummary(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetVendorPurchaseSummary returns purchase summary for a specific vendor
func (pc *PurchaseController) GetVendorPurchaseSummary(c *gin.Context) {
	vendorID, err := strconv.ParseUint(c.Param("vendor_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid vendor ID"})
		return
	}

	summary, err := pc.purchaseService.GetVendorPurchaseSummary(uint(vendorID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// GetPendingApprovals returns purchases pending approval for current user
func (pc *PurchaseController) GetPendingApprovals(c *gin.Context) {
	_ = c.MustGet("user_id").(uint) // userID not used in current implementation
	userRole := c.MustGet("role").(string)

	// Filter purchases requiring approval that user can approve
	filter := models.PurchaseFilter{
		ApprovalStatus: models.PurchaseApprovalPending,
		Page:          1,
		Limit:         100,
	}

	result, err := pc.purchaseService.GetPurchases(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filter based on user role - in a real app, this would be more sophisticated
	var filteredPurchases []models.Purchase
	for _, purchase := range result.Data {
		// Finance can approve all purchases
		if userRole == "finance" || userRole == "admin" {
			filteredPurchases = append(filteredPurchases, purchase)
		}
		// Director can approve all purchases (removed amount restriction)
		if userRole == "director" {
			filteredPurchases = append(filteredPurchases, purchase)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": filteredPurchases,
		"total": len(filteredPurchases),
		"user_role": userRole,
	})
}

// Purchase Payment Integration Operations

// GetPurchaseForPayment godoc
// @Summary Get purchase for payment
// @Description Get purchase details formatted for payment processing
// @Tags Purchases
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Purchase ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/purchases/{id}/for-payment [get]
func (pc *PurchaseController) GetPurchaseForPayment(c *gin.Context) {
	purchaseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	// Get purchase with vendor details
	purchase, err := pc.purchaseService.GetPurchaseByID(uint(purchaseID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase not found"})
		return
	}

	// Check if purchase can receive payment
	canReceivePayment := purchase.Status == models.PurchaseStatusApproved && 
						 purchase.PaymentMethod == "CREDIT" && 
						 purchase.OutstandingAmount > 0

	// Format response for payment processing
	response := gin.H{
		"purchase_id":        purchase.ID,
		"bill_number":        purchase.Code,
		"vendor": gin.H{
			"id":   purchase.Vendor.ID,
			"name": purchase.Vendor.Name,
			"type": "VENDOR",
		},
		"total_amount":       purchase.TotalAmount,
		"paid_amount":        purchase.PaidAmount,
		"outstanding_amount": purchase.OutstandingAmount,
		"status":             purchase.Status,
		"payment_method":     purchase.PaymentMethod,
		"date":               purchase.Date.Format("2006-01-02"),
		"can_receive_payment": canReceivePayment,
	}

	if !purchase.DueDate.IsZero() {
		response["due_date"] = purchase.DueDate.Format("2006-01-02")
	}

	if canReceivePayment {
		response["payment_url_suggestion"] = fmt.Sprintf("/api/purchases/%d/integrated-payment", purchase.ID)
	}

	c.JSON(http.StatusOK, response)
}

// CreateIntegratedPayment godoc
// @Summary Create integrated payment for purchase
// @Description Create payment record in both Purchase and Payment Management systems
// @Tags Purchases
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Purchase ID"
// @Param payment body models.APIResponse true "Payment details"
// @Success 200 {object} map[string]interface{}
// @Router /api/purchases/{id}/integrated-payment [post]
func (pc *PurchaseController) CreateIntegratedPayment(c *gin.Context) {
	purchaseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	var request struct {
		Amount      float64 `json:"amount" binding:"required,gt=0"`
		Date        string  `json:"date" binding:"required"`
		Method      string  `json:"method" binding:"required"`
		CashBankID  *uint   `json:"cash_bank_id"`
		Reference   string  `json:"reference"`
		Notes       string  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uint)

	// Parse date
	paymentDate, err := time.Parse(time.RFC3339, request.Date)
	if err != nil {
		// Try alternative date format
		paymentDate, err = time.Parse("2006-01-02", request.Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use RFC3339 or YYYY-MM-DD"})
			return
		}
	}

	// Create integrated payment via service
	result, err := pc.purchaseService.CreateIntegratedPayment(
		uint(purchaseID),
		request.Amount,
		paymentDate,
		request.Method,
		request.CashBankID,
		request.Reference,
		request.Notes,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create payment",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetPurchasePayments godoc
// @Summary Get purchase payments
// @Description Get all payments for a specific purchase
// @Tags Purchases
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Purchase ID"
// @Success 200 {array} models.PurchasePayment
// @Router /api/purchases/{id}/payments [get]
func (pc *PurchaseController) GetPurchasePayments(c *gin.Context) {
	purchaseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	payments, err := pc.purchaseService.GetPurchasePayments(uint(purchaseID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, payments)
}

// CreatePurchasePayment creates a payment for a purchase via Payment Management
func (pc *PurchaseController) CreatePurchasePayment(c *gin.Context) {
	purchaseID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	var request struct {
		Amount        float64 `json:"amount" binding:"required,gt=0"`
		Date          string  `json:"date" binding:"required"`
		Method        string  `json:"method" binding:"required"`
		CashBankID    uint    `json:"cash_bank_id" binding:"required"`
		Reference     string  `json:"reference"`
		Notes         string  `json:"notes"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("Payment creation validation error for purchase %d: %v", purchaseID, err)
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
	log.Printf("Received payment request for purchase %d: amount=%.2f, method=%s, cash_bank_id=%d, date=%s", purchaseID, request.Amount, request.Method, request.CashBankID, request.Date)
	
	// Parse date - support both RFC3339 and YYYY-MM-DD formats
	paymentDate, err := time.Parse(time.RFC3339, request.Date)
	if err != nil {
		// Try alternative date format (YYYY-MM-DD)
		paymentDate, err = time.Parse("2006-01-02", request.Date)
		if err != nil {
			log.Printf("Date parsing error for purchase %d: %v, date=%s", purchaseID, err, request.Date)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid date format",
				"details": "Date must be in RFC3339 format (2006-01-02T15:04:05Z) or YYYY-MM-DD format",
				"received_date": request.Date,
			})
			return
		}
		// If we parsed YYYY-MM-DD format, set to start of day
		paymentDate = time.Date(paymentDate.Year(), paymentDate.Month(), paymentDate.Day(), 0, 0, 0, 0, time.UTC)
	}
	log.Printf("Parsed payment date: %v", paymentDate)
	
	// Get purchase details to validate and get vendor ID
	purchase, err := pc.purchaseService.GetPurchaseByID(uint(purchaseID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Purchase not found",
			"details": err.Error(),
		})
		return
	}

	// Validate purchase status - only APPROVED credit purchases can receive payments
	if purchase.Status != models.PurchaseStatusApproved {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Purchase must be approved to receive payments",
			"purchase_status": purchase.Status,
		})
		return
	}
	
	if purchase.PaymentMethod != models.PurchasePaymentCredit {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Only credit purchases can receive payments",
			"payment_method": purchase.PaymentMethod,
		})
		return
	}

	// Validate payment amount
	if request.Amount > purchase.OutstandingAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Payment amount exceeds outstanding amount",
			"outstanding_amount": purchase.OutstandingAmount,
			"requested_amount": request.Amount,
		})
		return
	}

	// Get user ID with error handling
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		log.Printf("Error: user_id not found in context for purchase %d payment", purchaseID)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
			"details": "user_id not found in context",
		})
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		log.Printf("Error: user_id has invalid type for purchase %d payment: %T", purchaseID, userIDInterface)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid user authentication",
			"details": "user_id has invalid type",
		})
		return
	}

	// Create payment request for Payment Management service
	paymentRequest := services.PaymentCreateRequest{
		ContactID:   purchase.VendorID,
		CashBankID:  request.CashBankID,
		Date:        paymentDate, // Use the parsed paymentDate
		Amount:      request.Amount,
		Method:      request.Method,
		Reference:   request.Reference,
		Notes:       fmt.Sprintf("Payment for Purchase %s - %s", purchase.Code, request.Notes),
		Allocations: []services.InvoiceAllocation{
			{
				InvoiceID: uint(purchaseID),
				Amount:    request.Amount,
			},
		},
	}

	// Use Payment Management service
	log.Printf("Calling PaymentService.CreatePayablePayment for purchase %d with amount %.2f", purchaseID, paymentRequest.Amount)
	payment, err := pc.paymentService.CreatePayablePayment(paymentRequest, userID)
	if err != nil {
		log.Printf("Error in CreatePayablePayment for purchase %d: %v", purchaseID, err)
		
		// Check if this is an insufficient balance error
		errorMessage := err.Error()
		if strings.Contains(errorMessage, "insufficient balance") {
			// Extract available balance from error message
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": "Saldo rekening tidak mencukupi",
				"error_type": "INSUFFICIENT_BALANCE",
				"details": errorMessage,
				"requested_amount": request.Amount,
				"status": "error",
				"message": "Saldo di rekening bank yang dipilih tidak mencukupi untuk melakukan pembayaran ini",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error": "Failed to create payment",
				"details": err.Error(),
				"status": "error",
			})
		}
		return
	}
	log.Printf("✅ Payment created successfully: ID=%d, Code=%s", payment.ID, payment.Code)

	// 🔥 NEW: Create SSOT journal entry for purchase payment
	log.Printf("🧾 Creating SSOT journal entry for purchase payment...")
	err = pc.purchaseService.CreatePurchasePaymentJournal(
		uint(purchaseID),
		request.Amount,
		request.CashBankID,
		fmt.Sprintf("PAY-%s", purchase.Code),
		fmt.Sprintf("Payment for Purchase %s", purchase.Code),
		userID,
	)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to create SSOT journal entry for payment: %v", err)
		// Don't fail the payment process, but log the issue
	} else {
		log.Printf("✅ SSOT journal entry created for purchase payment")
	}

	// CRITICAL FIX: Update purchase payment amounts after successful payment
	log.Printf("🔄 Updating purchase payment amounts for purchase %d...", purchaseID)
	
	// Calculate new paid and outstanding amounts
	newPaidAmount := purchase.PaidAmount + request.Amount
	newOutstandingAmount := purchase.TotalAmount - newPaidAmount
	
	// DO NOT change status to PAID automatically when payment is made
	// The purchase should remain APPROVED to allow receipt creation
	// Only change to COMPLETED when both payment AND receipt are complete
	newStatus := purchase.Status
	if newOutstandingAmount < 0 {
		newOutstandingAmount = 0 // Ensure it doesn't go negative
	}
	// Status remains APPROVED - will be changed to COMPLETED only when receipt is created
	
	log.Printf("📊 Payment amounts: Paid %.2f → %.2f, Outstanding %.2f → %.2f, Status %s → %s", 
		purchase.PaidAmount, newPaidAmount, purchase.OutstandingAmount, newOutstandingAmount, purchase.Status, newStatus)
	
	// Update purchase in database
	err = pc.purchaseService.UpdatePurchasePaymentAmounts(uint(purchaseID), newPaidAmount, newOutstandingAmount, newStatus)
	if err != nil {
		log.Printf("❌ Critical error: Failed to update purchase payment amounts: %v", err)
		// Payment was created successfully, but purchase amounts weren't updated
		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"payment": payment,
			"warning": "Payment created but purchase amounts could not be updated",
			"error_details": err.Error(),
			"status": "partial_success",
			"message": "Payment recorded but purchase status may not reflect the payment. Please refresh the page.",
		})
		return
	}
	
	log.Printf("✅ Purchase payment amounts updated successfully")
	
	// Return response with both payment info and updated purchase status
	updatedPurchase, err := pc.purchaseService.GetPurchaseByID(uint(purchaseID))
	if err != nil {
		// If we can't get updated purchase info, still return success but with calculated values
		log.Printf("Warning: Could not fetch updated purchase info after payment creation: %v", err)
		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"payment": payment,
			"updated_purchase": gin.H{
				"id": uint(purchaseID),
				"status": newStatus,
				"paid_amount": newPaidAmount,
				"outstanding_amount": newOutstandingAmount,
			},
			"message": "Payment created successfully via Payment Management",
			"status": "success",
		})
		return
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"payment": payment,
		"updated_purchase": gin.H{
			"id": updatedPurchase.ID,
			"status": updatedPurchase.Status,
			"paid_amount": updatedPurchase.PaidAmount,
			"outstanding_amount": updatedPurchase.OutstandingAmount,
		},
		"message": "Payment created successfully via Payment Management",
		"status": "success",
	})
}

// Dashboard endpoint for purchases
func (pc *PurchaseController) GetPurchaseDashboard(c *gin.Context) {
	_ = c.MustGet("user_id").(uint) // userID not used in current implementation
	userRole := c.MustGet("role").(string)

	// Get summary
	summary, err := pc.purchaseService.GetPurchasesSummary("", "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get pending approvals count
	pendingFilter := models.PurchaseFilter{
		ApprovalStatus: models.PurchaseApprovalPending,
		Page:          1,
		Limit:         1000,
	}

	pendingResult, err := pc.purchaseService.GetPurchases(pendingFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"summary":         summary,
		"pending_count":   len(pendingResult.Data),
		"user_role":       userRole,
		"can_approve":     userRole == "finance" || userRole == "director" || userRole == "admin",
	}

	c.JSON(http.StatusOK, response)
}

// GetPurchaseJournalEntries retrieves SSOT journal entries for a purchase
func (pc *PurchaseController) GetPurchaseJournalEntries(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	// Get journal entries for the purchase
	entries, err := pc.purchaseService.GetPurchaseJournalEntries(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"purchase_id": id,
		"journal_entries": entries,
		"count": len(entries),
	})
}
