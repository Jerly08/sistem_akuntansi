package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"

	"github.com/gin-gonic/gin"
)

type PurchaseController struct {
	purchaseService *services.PurchaseService
}

func NewPurchaseController(purchaseService *services.PurchaseService) *PurchaseController {
	return &PurchaseController{
		purchaseService: purchaseService,
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
	userRole := c.MustGet("role").(string)
	
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
	userRole := c.MustGet("role").(string)

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
	userRole := c.MustGet("role").(string)

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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.MustGet("user_id").(uint)

	receipt, err := pc.purchaseService.CreatePurchaseReceipt(request, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, receipt)
}

// GetPurchaseReceipts returns receipts for a purchase
func (pc *PurchaseController) GetPurchaseReceipts(c *gin.Context) {
	_, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase ID"})
		return
	}

	// This would need to be implemented in the service
	c.JSON(http.StatusOK, gin.H{"message": "Get receipts endpoint - to be implemented"})
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
	userRole := c.MustGet("user_role").(string)

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
		"data":  filteredPurchases,
		"count": len(filteredPurchases),
	})
}

// Dashboard endpoint for purchases
func (pc *PurchaseController) GetPurchaseDashboard(c *gin.Context) {
	_ = c.MustGet("user_id").(uint) // userID not used in current implementation
	userRole := c.MustGet("user_role").(string)

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
