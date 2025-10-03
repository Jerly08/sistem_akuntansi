package controllers

import (
	"log"
	"strconv"
	"time"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"app-sistem-akuntansi/utils"

	"github.com/gin-gonic/gin"
)

type InvoiceTypeController struct {
	invoiceTypeService   *services.InvoiceTypeService
	invoiceNumberService *services.InvoiceNumberService
}

func NewInvoiceTypeController(invoiceTypeService *services.InvoiceTypeService, invoiceNumberService *services.InvoiceNumberService) *InvoiceTypeController {
	return &InvoiceTypeController{
		invoiceTypeService:   invoiceTypeService,
		invoiceNumberService: invoiceNumberService,
	}
}

// GetInvoiceTypes gets all invoice types
func (c *InvoiceTypeController) GetInvoiceTypes(ctx *gin.Context) {
	log.Printf("📋 Getting invoice types list")
	
	// Check for active_only query parameter
	activeOnly := ctx.Query("active_only") == "true"
	
	invoiceTypes, err := c.invoiceTypeService.GetInvoiceTypes(activeOnly)
	if err != nil {
		log.Printf("❌ Failed to get invoice types: %v", err)
		utils.SendInternalError(ctx, "Failed to retrieve invoice types", err.Error())
		return
	}

	log.Printf("✅ Retrieved %d invoice types", len(invoiceTypes))
	utils.SendSuccess(ctx, "Invoice types retrieved successfully", invoiceTypes)
}

// GetInvoiceType gets a single invoice type by ID
func (c *InvoiceTypeController) GetInvoiceType(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		log.Printf("❌ Invalid invoice type ID parameter: %v", err)
		utils.SendValidationError(ctx, "Invalid invoice type ID", map[string]string{
			"id": "Invoice type ID must be a valid positive number",
		})
		return
	}

	log.Printf("🔍 Getting invoice type details for ID: %d", id)
	
	invoiceType, err := c.invoiceTypeService.GetInvoiceTypeByID(uint(id))
	if err != nil {
		log.Printf("❌ Invoice type %d not found: %v", id, err)
		utils.SendNotFound(ctx, "Invoice type not found")
		return
	}

	log.Printf("✅ Retrieved invoice type %d details successfully", id)
	utils.SendSuccess(ctx, "Invoice type retrieved successfully", invoiceType)
}

// CreateInvoiceType creates a new invoice type
func (c *InvoiceTypeController) CreateInvoiceType(ctx *gin.Context) {
	log.Printf("🆕 Creating new invoice type")
	
	var request models.InvoiceTypeCreateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Printf("❌ Invalid invoice type creation request: %v", err)
		utils.SendValidationError(ctx, "Invalid invoice type data", map[string]string{
			"request": "Please check the request format and required fields",
		})
		return
	}

	// Get user ID from context
	userIDInterface, exists := ctx.Get("user_id")
	if !exists {
		log.Printf("❌ User authentication missing for invoice type creation")
		utils.SendUnauthorized(ctx, "User authentication required")
		return
	}
	userID, ok := userIDInterface.(uint)
	if !ok {
		log.Printf("❌ Invalid user ID type: %T", userIDInterface)
		utils.SendUnauthorized(ctx, "Invalid user authentication")
		return
	}

	log.Printf("📄 Creating invoice type '%s' (%s) by user %d", request.Name, request.Code, userID)
	
	invoiceType, err := c.invoiceTypeService.CreateInvoiceType(request, userID)
	if err != nil {
		log.Printf("❌ Failed to create invoice type: %v", err)
		utils.SendBusinessRuleError(ctx, "Failed to create invoice type", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Invoice type created successfully: %d", invoiceType.ID)
	utils.SendCreated(ctx, "Invoice type created successfully", invoiceType)
}

// UpdateInvoiceType updates an existing invoice type
func (c *InvoiceTypeController) UpdateInvoiceType(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendValidationError(ctx, "Invalid invoice type ID", map[string]string{
			"id": "Invoice type ID must be a valid positive number",
		})
		return
	}

	log.Printf("📝 Updating invoice type %d", id)
	
	var request models.InvoiceTypeUpdateRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		log.Printf("❌ Invalid invoice type update request: %v", err)
		utils.SendValidationError(ctx, "Invalid update data", map[string]string{
			"request": "Please check the request format",
		})
		return
	}

	invoiceType, err := c.invoiceTypeService.UpdateInvoiceType(uint(id), request)
	if err != nil {
		log.Printf("❌ Failed to update invoice type %d: %v", id, err)
		utils.SendBusinessRuleError(ctx, "Failed to update invoice type", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Invoice type %d updated successfully", id)
	utils.SendSuccess(ctx, "Invoice type updated successfully", invoiceType)
}

// DeleteInvoiceType deletes an invoice type
func (c *InvoiceTypeController) DeleteInvoiceType(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendValidationError(ctx, "Invalid invoice type ID", map[string]string{
			"id": "Invoice type ID must be a valid positive number",
		})
		return
	}

	log.Printf("🗑️ Deleting invoice type %d", id)
	
	if err := c.invoiceTypeService.DeleteInvoiceType(uint(id)); err != nil {
		log.Printf("❌ Failed to delete invoice type %d: %v", id, err)
		utils.SendBusinessRuleError(ctx, "Failed to delete invoice type", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Invoice type %d deleted successfully", id)
	utils.SendSuccess(ctx, "Invoice type deleted successfully", nil)
}

// ToggleInvoiceType toggles the active status of an invoice type
func (c *InvoiceTypeController) ToggleInvoiceType(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendValidationError(ctx, "Invalid invoice type ID", map[string]string{
			"id": "Invoice type ID must be a valid positive number",
		})
		return
	}

	log.Printf("🔄 Toggling invoice type %d status", id)
	
	invoiceType, err := c.invoiceTypeService.ToggleInvoiceType(uint(id))
	if err != nil {
		log.Printf("❌ Failed to toggle invoice type %d status: %v", id, err)
		utils.SendBusinessRuleError(ctx, "Failed to toggle invoice type status", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	status := "activated"
	if !invoiceType.IsActive {
		status = "deactivated"
	}

	log.Printf("✅ Invoice type %d %s successfully", id, status)
	utils.SendSuccess(ctx, "Invoice type "+status+" successfully", invoiceType)
}

// GetActiveInvoiceTypes gets only active invoice types (for dropdowns)
func (c *InvoiceTypeController) GetActiveInvoiceTypes(ctx *gin.Context) {
	log.Printf("📋 Getting active invoice types for dropdown")
	
	invoiceTypes, err := c.invoiceTypeService.GetActiveInvoiceTypes()
	if err != nil {
		log.Printf("❌ Failed to get active invoice types: %v", err)
		utils.SendInternalError(ctx, "Failed to retrieve active invoice types", err.Error())
		return
	}

	log.Printf("✅ Retrieved %d active invoice types", len(invoiceTypes))
	utils.SendSuccess(ctx, "Active invoice types retrieved successfully", invoiceTypes)
}

// PreviewInvoiceNumber previews what the next invoice number would be for a given type and date
func (c *InvoiceTypeController) PreviewInvoiceNumber(ctx *gin.Context) {
	var request models.InvoiceNumberRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		utils.SendValidationError(ctx, "Invalid request data", map[string]string{
			"request": "Please provide invoice_type_id and date",
		})
		return
	}

	log.Printf("🔍 Previewing invoice number for type %d, date %s", request.InvoiceTypeID, request.Date.Format("2006-01-02"))
	
	preview, err := c.invoiceNumberService.PreviewInvoiceNumber(request.InvoiceTypeID, request.Date)
	if err != nil {
		log.Printf("❌ Failed to preview invoice number: %v", err)
		utils.SendBusinessRuleError(ctx, "Failed to preview invoice number", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	log.Printf("✅ Invoice number preview generated: %s", preview.InvoiceNumber)
	utils.SendSuccess(ctx, "Invoice number preview generated successfully", preview)
}

// PreviewInvoiceNumberByID previews the next invoice number using path param and optional date query (?date=YYYY-MM-DD)
func (c *InvoiceTypeController) PreviewInvoiceNumberByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil || id == 0 {
		utils.SendValidationError(ctx, "Invalid invoice type ID", map[string]string{
			"id": "Invoice type ID must be a valid positive number",
		})
		return
	}

	dateStr := ctx.Query("date")
	var date time.Time
	if dateStr == "" {
		date = time.Now()
	} else {
		parsed, perr := time.Parse("2006-01-02", dateStr)
		if perr != nil {
			utils.SendValidationError(ctx, "Invalid date format", map[string]string{
				"date": "Use YYYY-MM-DD format",
			})
			return
		}
		date = parsed
	}

	log.Printf("🔍 Previewing invoice number for type %d via GET, date %s", id, date.Format("2006-01-02"))
	preview, svcErr := c.invoiceNumberService.PreviewInvoiceNumber(uint(id), date)
	if svcErr != nil {
		log.Printf("❌ Failed to preview invoice number: %v", svcErr)
		utils.SendBusinessRuleError(ctx, "Failed to preview invoice number", map[string]interface{}{
			"details": svcErr.Error(),
		})
		return
	}

	log.Printf("✅ Invoice number preview generated: %s", preview.InvoiceNumber)
	utils.SendSuccess(ctx, "Invoice number preview generated successfully", preview)
}

// GetCounterHistory gets counter history for a specific invoice type
func (c *InvoiceTypeController) GetCounterHistory(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		utils.SendValidationError(ctx, "Invalid invoice type ID", map[string]string{
			"id": "Invoice type ID must be a valid positive number",
		})
		return
	}

	log.Printf("📊 Getting counter history for invoice type %d", id)
	
	history, err := c.invoiceNumberService.GetCounterHistory(uint(id))
	if err != nil {
		log.Printf("❌ Failed to get counter history: %v", err)
		utils.SendInternalError(ctx, "Failed to retrieve counter history", err.Error())
		return
	}

	log.Printf("✅ Retrieved counter history with %d entries", len(history))
	utils.SendSuccess(ctx, "Counter history retrieved successfully", history)
}