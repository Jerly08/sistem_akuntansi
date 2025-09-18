package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type EnhancedPLJournalController struct {
	db *gorm.DB
}

func NewEnhancedPLJournalController(db *gorm.DB) *EnhancedPLJournalController {
	return &EnhancedPLJournalController{
		db: db,
	}
}

// JournalEntriesForPLLine handles journal entry drill-down for specific P&L line items
// This endpoint is specifically designed for Enhanced Profit & Loss Statement integration
func (c *EnhancedPLJournalController) JournalEntriesForPLLine(ctx *gin.Context) {
	// Parse query parameters
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")
	lineItem := ctx.Query("line_item") // e.g., "total_revenue", "gross_profit", "operating_expenses"
	accountCodes := ctx.Query("account_codes") // comma-separated account codes

	// Validate required parameters
	if startDateStr == "" || endDateStr == "" {
		appError := utils.NewBadRequestError("start_date and end_date are required")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		appError := utils.NewBadRequestError("Invalid start_date format. Use YYYY-MM-DD")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		appError := utils.NewBadRequestError("Invalid end_date format. Use YYYY-MM-DD")
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	// Parse pagination
	page := 1
	limit := 20
	if p := ctx.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if l := ctx.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Get account IDs based on line item type if not provided
	var accountIDs []uint
	if accountCodes != "" {
		// Parse account codes from query parameter
		c.db.Model(&models.Account{}).
			Select("id").
			Where("code IN ?", parseAccountCodes(accountCodes)).
			Pluck("id", &accountIDs)
	} else if lineItem != "" {
		// Get relevant account IDs based on line item type
		accountIDs = c.getAccountIDsForPLLineItem(lineItem)
	}

	// Build journal entry query
	query := c.db.Model(&models.JournalEntry{}).
		Preload("Account").
		Preload("JournalLines").
		Preload("JournalLines.Account").
		Preload("Creator").
		Where("entry_date >= ? AND entry_date <= ?", startDate, endDate).
		Where("status = ?", models.JournalStatusPosted)

	// Filter by accounts if specified
	if len(accountIDs) > 0 {
		// Get journal entries that have lines with the specified accounts
		subQuery := c.db.Model(&models.JournalLine{}).
			Select("journal_entry_id").
			Where("account_id IN ?", accountIDs)
		query = query.Where("id IN (?)", subQuery)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		appError := utils.NewInternalError("Failed to count journal entries", err)
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	// Apply pagination and ordering
	offset := (page - 1) * limit
	query = query.Offset(offset).Limit(limit).
		Order("entry_date DESC, created_at DESC")

	// Execute query
	var journalEntries []models.JournalEntry
	if err := query.Find(&journalEntries).Error; err != nil {
		appError := utils.NewInternalError("Failed to fetch journal entries", err)
		ctx.JSON(appError.StatusCode, appError.ToErrorResponse(""))
		return
	}

	// Calculate summary
	summary := c.calculateSummary(journalEntries)

	// Prepare response
	response := gin.H{
		"data": gin.H{
			"journal_entries": journalEntries,
			"total":          total,
			"page":           page,
			"limit":          limit,
			"summary":        summary,
			"metadata": gin.H{
				"line_item":     lineItem,
				"account_codes": accountCodes,
				"date_range": gin.H{
					"start": startDate.Format("2006-01-02"),
					"end":   endDate.Format("2006-01-02"),
				},
				"generated_at": time.Now(),
			},
		},
		"success": true,
		"message": fmt.Sprintf("Found %d journal entries for P&L line item", len(journalEntries)),
	}

	ctx.JSON(http.StatusOK, response)
}

// getAccountIDsForPLLineItem returns relevant account IDs based on P&L line item type
func (c *EnhancedPLJournalController) getAccountIDsForPLLineItem(lineItem string) []uint {
	var accountIDs []uint
	
	switch lineItem {
	case "total_revenue", "revenue", "sales_revenue":
		// Get revenue accounts (4xxx)
		c.db.Model(&models.Account{}).
			Select("id").
			Where("code LIKE '4%'").
			Where("is_header = false").
			Pluck("id", &accountIDs)
		
	case "gross_profit", "cogs", "cost_of_goods_sold":
		// Get COGS accounts (5101, 5xxx for cost accounts)
		c.db.Model(&models.Account{}).
			Select("id").
			Where("code LIKE '5101%' OR (code LIKE '5%' AND name LIKE '%cost%')").
			Where("is_header = false").
			Pluck("id", &accountIDs)
			
	case "operating_expenses", "operating_performance":
		// Get operating expense accounts (5xxx excluding COGS)
		c.db.Model(&models.Account{}).
			Select("id").
			Where("code LIKE '5%'").
			Where("code NOT LIKE '5101%'").
			Where("is_header = false").
			Pluck("id", &accountIDs)
			
	case "net_income", "income_before_tax":
		// Get all revenue and expense accounts
		c.db.Model(&models.Account{}).
			Select("id").
			Where("code LIKE '4%' OR code LIKE '5%'").
			Where("is_header = false").
			Pluck("id", &accountIDs)
	}
	
	return accountIDs
}

// calculateSummary calculates summary statistics for the journal entries
func (c *EnhancedPLJournalController) calculateSummary(entries []models.JournalEntry) gin.H {
	var totalDebit, totalCredit float64
	entryCount := len(entries)
	
	for _, entry := range entries {
		totalDebit += entry.TotalDebit
		totalCredit += entry.TotalCredit
	}
	
	return gin.H{
		"total_debit":  totalDebit,
		"total_credit": totalCredit,
		"net_amount":   totalDebit - totalCredit,
		"entry_count":  entryCount,
	}
}

// parseAccountCodes parses comma-separated account codes
func parseAccountCodes(accountCodes string) []string {
	var codes []string
	for _, code := range splitAndTrim(accountCodes, ",") {
		if code != "" {
			codes = append(codes, code)
		}
	}
	return codes
}

// splitAndTrim helper function
func splitAndTrim(input, separator string) []string {
	var result []string
	if input == "" {
		return result
	}
	
	parts := []string{}
	current := ""
	for _, char := range input {
		if string(char) == separator {
			parts = append(parts, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	parts = append(parts, current)
	
	for _, part := range parts {
		trimmed := ""
		for _, char := range part {
			if char != ' ' && char != '\t' && char != '\n' && char != '\r' {
				trimmed += string(char)
			}
		}
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}