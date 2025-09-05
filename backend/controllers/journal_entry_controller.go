package controllers

import (
	"net/http"
	"strconv"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/repositories"
	"app-sistem-akuntansi/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type JournalEntryController struct {
	journalRepo repositories.JournalEntryRepository
	accountRepo repositories.AccountRepository
}

func NewJournalEntryController(db *gorm.DB) *JournalEntryController {
	return &JournalEntryController{
		journalRepo: repositories.NewJournalEntryRepository(db),
		accountRepo: repositories.NewAccountRepository(db),
	}
}

// CreateJournalEntry creates a new journal entry
// @Summary Create journal entry
// @Description Create a new journal entry with balanced debit and credit lines
// @Tags journal-entries
// @Accept json
// @Produce json
// @Param request body models.JournalEntryCreateRequest true "Journal entry data"
// @Success 201 {object} models.JournalEntry
// @Failure 400 {object} utils.ErrorResponse
// @Router /journal-entries [post]
func (c *JournalEntryController) CreateJournalEntry(ctx *gin.Context) {
	var req models.JournalEntryCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		appError := utils.NewBadRequestError("Invalid request payload")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	// Get current user ID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		appError := utils.NewUnauthorizedError("User not authenticated")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	entry, err := c.journalRepo.Create(ctx.Request.Context(), &req)
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			ctx.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to create journal entry", err)
			ctx.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	// Set user ID
	entry.UserID = userID.(uint)

	ctx.JSON(http.StatusCreated, gin.H{"data": entry})
}

// GetJournalEntry gets a journal entry by ID
// @Summary Get journal entry
// @Description Get journal entry details by ID
// @Tags journal-entries
// @Produce json
// @Param id path int true "Journal Entry ID"
// @Success 200 {object} models.JournalEntry
// @Failure 404 {object} utils.ErrorResponse
// @Router /journal-entries/{id} [get]
func (c *JournalEntryController) GetJournalEntry(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		appError := utils.NewBadRequestError("Invalid journal entry ID")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	entry, err := c.journalRepo.FindByID(ctx.Request.Context(), uint(id))
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			ctx.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to get journal entry", err)
			ctx.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": entry})
}

// GetJournalEntries gets journal entries with filtering
// @Summary List journal entries
// @Description Get paginated list of journal entries with optional filtering
// @Tags journal-entries
// @Produce json
// @Param status query string false "Filter by status (DRAFT, POSTED, REVERSED)"
// @Param reference_type query string false "Filter by reference type"
// @Param account_id query string false "Filter by account ID"
// @Param start_date query string false "Filter by start date (YYYY-MM-DD)"
// @Param end_date query string false "Filter by end date (YYYY-MM-DD)"
// @Param search query string false "Search in description, reference, or code"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20)"
// @Success 200 {object} object{data=[]models.JournalEntry,total=int64}
// @Router /journal-entries [get]
func (c *JournalEntryController) GetJournalEntries(ctx *gin.Context) {
	filter := &models.JournalEntryFilter{
		Status:        ctx.Query("status"),
		ReferenceType: ctx.Query("reference_type"),
		AccountID:     ctx.Query("account_id"),
		StartDate:     ctx.Query("start_date"),
		EndDate:       ctx.Query("end_date"),
		Search:        ctx.Query("search"),
	}

	// Parse pagination
	if page := ctx.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filter.Page = p
		}
	}

	if limit := ctx.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	entries, total, err := c.journalRepo.FindAll(ctx.Request.Context(), filter)
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			ctx.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to get journal entries", err)
			ctx.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  entries,
		"total": total,
	})
}

// UpdateJournalEntry updates a journal entry (only if DRAFT)
// @Summary Update journal entry
// @Description Update journal entry details (only draft entries can be updated)
// @Tags journal-entries
// @Accept json
// @Produce json
// @Param id path int true "Journal Entry ID"
// @Param request body models.JournalEntryUpdateRequest true "Updated journal entry data"
// @Success 200 {object} models.JournalEntry
// @Failure 400 {object} utils.ErrorResponse
// @Router /journal-entries/{id} [put]
func (c *JournalEntryController) UpdateJournalEntry(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		appError := utils.NewBadRequestError("Invalid journal entry ID")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	var req models.JournalEntryUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		appError := utils.NewBadRequestError("Invalid request payload")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	entry, err := c.journalRepo.Update(ctx.Request.Context(), uint(id), &req)
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			ctx.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to update journal entry", err)
			ctx.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": entry})
}

// DeleteJournalEntry deletes a journal entry (only if DRAFT)
// @Summary Delete journal entry
// @Description Delete a journal entry (only draft entries can be deleted)
// @Tags journal-entries
// @Param id path int true "Journal Entry ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} utils.ErrorResponse
// @Router /journal-entries/{id} [delete]
func (c *JournalEntryController) DeleteJournalEntry(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		appError := utils.NewBadRequestError("Invalid journal entry ID")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	err = c.journalRepo.Delete(ctx.Request.Context(), uint(id))
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			ctx.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to delete journal entry", err)
			ctx.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Journal entry deleted successfully"})
}

// PostJournalEntry posts a journal entry to update account balances
// @Summary Post journal entry
// @Description Post a journal entry to update account balances (changes status to POSTED)
// @Tags journal-entries
// @Param id path int true "Journal Entry ID"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} utils.ErrorResponse
// @Router /journal-entries/{id}/post [post]
func (c *JournalEntryController) PostJournalEntry(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		appError := utils.NewBadRequestError("Invalid journal entry ID")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	// Get current user ID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		appError := utils.NewUnauthorizedError("User not authenticated")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	err = c.journalRepo.PostJournalEntry(ctx.Request.Context(), uint(id), userID.(uint))
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			ctx.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to post journal entry", err)
			ctx.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Journal entry posted successfully"})
}

// ReverseJournalEntry creates a reversing entry
// @Summary Reverse journal entry
// @Description Create a reversing journal entry to undo the effects of a posted entry
// @Tags journal-entries
// @Accept json
// @Produce json
// @Param id path int true "Journal Entry ID"
// @Param request body object{reason=string} true "Reversal reason"
// @Success 201 {object} models.JournalEntry
// @Failure 400 {object} utils.ErrorResponse
// @Router /journal-entries/{id}/reverse [post]
func (c *JournalEntryController) ReverseJournalEntry(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		appError := utils.NewBadRequestError("Invalid journal entry ID")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		appError := utils.NewBadRequestError("Reversal reason is required")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	// Get current user ID from context
	userID, exists := ctx.Get("user_id")
	if !exists {
		appError := utils.NewUnauthorizedError("User not authenticated")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	reversingEntry, err := c.journalRepo.ReverseJournalEntry(ctx.Request.Context(), uint(id), userID.(uint), req.Reason)
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			ctx.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to reverse journal entry", err)
			ctx.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": reversingEntry})
}

// GetJournalEntrySummary gets summary statistics
// @Summary Get journal entry summary
// @Description Get summary statistics for journal entries
// @Tags journal-entries
// @Produce json
// @Success 200 {object} models.JournalEntrySummary
// @Router /journal-entries/summary [get]
func (c *JournalEntryController) GetJournalEntrySummary(ctx *gin.Context) {
	summary, err := c.journalRepo.GetSummary(ctx.Request.Context())
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			ctx.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to get journal entry summary", err)
			ctx.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": summary})
}

// GetAccountJournalEntries gets journal entries for a specific account
// @Summary Get account journal entries
// @Description Get journal entries that affect a specific account (General Ledger view)
// @Tags journal-entries
// @Produce json
// @Param account_id path int true "Account ID"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 20)"
// @Success 200 {object} object{data=[]models.JournalEntry,total=int64}
// @Router /accounts/{account_id}/journal-entries [get]
func (c *JournalEntryController) GetAccountJournalEntries(ctx *gin.Context) {
	accountIDStr := ctx.Param("account_id")
	accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
	if err != nil {
		appError := utils.NewBadRequestError("Invalid account ID")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	// Verify account exists
	_, err = c.accountRepo.FindByID(ctx.Request.Context(), uint(accountID))
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			ctx.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to find account", err)
			ctx.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	filter := &models.JournalEntryFilter{
		AccountID: accountIDStr,
		StartDate: ctx.Query("start_date"),
		EndDate:   ctx.Query("end_date"),
	}

	// Parse pagination
	if page := ctx.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			filter.Page = p
		}
	}

	if limit := ctx.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filter.Limit = l
		}
	}

	entries, total, err := c.journalRepo.FindAll(ctx.Request.Context(), filter)
	if err != nil {
		if appErr := utils.GetAppError(err); appErr != nil {
			ctx.JSON(appErr.StatusCode, appErr.ToErrorResponse(""))
		} else {
			internalErr := utils.NewInternalError("Failed to get account journal entries", err)
			ctx.JSON(internalErr.StatusCode, internalErr.ToErrorResponse(""))
		}
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  entries,
		"total": total,
	})
}

// AutoGenerateFromSale creates journal entry from sale transaction
// @Summary Auto-generate from sale
// @Description Automatically generate journal entry from a sale transaction
// @Tags journal-entries
// @Accept json
// @Produce json
// @Param request body object{sale_id=uint} true "Sale ID"
// @Success 201 {object} models.JournalEntry
// @Failure 400 {object} utils.ErrorResponse
// @Router /journal-entries/auto-generate/sale [post]
func (c *JournalEntryController) AutoGenerateFromSale(ctx *gin.Context) {
	var req struct {
		SaleID uint `json:"sale_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		appError := utils.NewBadRequestError("Sale ID is required")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	// This would require integration with sales repository
	// For now, return a placeholder response
	ctx.JSON(http.StatusNotImplemented, gin.H{
		"message": "Auto-generation from sales will be implemented when sales integration is complete",
		"sale_id": req.SaleID,
	})
}

// AutoGenerateFromPurchase creates journal entry from purchase transaction
// @Summary Auto-generate from purchase
// @Description Automatically generate journal entry from a purchase transaction
// @Tags journal-entries
// @Accept json
// @Produce json
// @Param request body object{purchase_id=uint} true "Purchase ID"
// @Success 201 {object} models.JournalEntry
// @Failure 400 {object} utils.ErrorResponse
// @Router /journal-entries/auto-generate/purchase [post]
func (c *JournalEntryController) AutoGenerateFromPurchase(ctx *gin.Context) {
	var req struct {
		PurchaseID uint `json:"purchase_id" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		appError := utils.NewBadRequestError("Purchase ID is required")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	// This would require integration with purchase repository
	// For now, return a placeholder response
	ctx.JSON(http.StatusNotImplemented, gin.H{
		"message": "Auto-generation from purchases will be implemented when purchase integration is complete",
		"purchase_id": req.PurchaseID,
	})
}
