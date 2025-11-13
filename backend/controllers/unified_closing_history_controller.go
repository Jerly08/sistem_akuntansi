package controllers

import (
	"net/http"
	"fmt"
	
	"app-sistem-akuntansi/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UnifiedClosingHistoryController handles closing history from both sources
type UnifiedClosingHistoryController struct {
	db *gorm.DB
}

// NewUnifiedClosingHistoryController creates a new unified closing history controller
func NewUnifiedClosingHistoryController(db *gorm.DB) *UnifiedClosingHistoryController {
	return &UnifiedClosingHistoryController{
		db: db,
	}
}

// GetUnifiedClosingHistory returns combined history from all closing sources
func (c *UnifiedClosingHistoryController) GetUnifiedClosingHistory(ctx *gin.Context) {
	var history []map[string]interface{}
	
	// 1. First, check accounting_periods table (UnifiedPeriodClosing creates records here)
	var accountingPeriods []models.AccountingPeriod
	err := c.db.Where("is_closed = ? OR is_locked = ?", true, true).
		Order("end_date DESC").
		Limit(20).
		Find(&accountingPeriods).Error
	
	if err == nil && len(accountingPeriods) > 0 {
		fmt.Printf("[UnifiedClosingHistory] Found %d closed periods in accounting_periods\n", len(accountingPeriods))
		for _, period := range accountingPeriods {
			// Generate a code for display
			periodCode := fmt.Sprintf("PC-%s", period.EndDate.Format("2006-01-02"))
			
			history = append(history, map[string]interface{}{
				"id":          period.ID,
				"code":        periodCode,
				"description": period.Description,
				"entry_date":  period.EndDate,
				"start_date":  period.StartDate,
				"created_at":  period.CreatedAt,
				"total_debit": period.NetIncome,
				"net_income":  period.NetIncome,
				"source":      "accounting_periods",
			})
		}
	}
	
	// 2. Also check SSOT journal entries for closing entries
	var sootJournalEntries []models.SSOTJournalEntry
	err = c.db.Where("source_type IN (?) OR description LIKE ?", 
		[]string{"CLOSING", "PERIOD_CLOSING", "FISCAL_CLOSING"},
		"%closing%").
		Order("entry_date DESC").
		Limit(10).
		Find(&sootJournalEntries).Error
		
	if err == nil && len(sootJournalEntries) > 0 {
		fmt.Printf("[UnifiedClosingHistory] Found %d SSOT closing entries\n", len(sootJournalEntries))
		for _, entry := range sootJournalEntries {
			// Generate a code
			entryCode := fmt.Sprintf("CLO-%d", entry.ID)
			
			history = append(history, map[string]interface{}{
				"id":          entry.ID,
				"code":        entryCode,
				"description": entry.Description,
				"entry_date":  entry.EntryDate,
				"created_at":  entry.CreatedAt,
				"total_debit": entry.TotalDebit.InexactFloat64(),
				"source":      "ssot_journal",
			})
		}
	}
	
	// Legacy journal_entries removed - only using SSOT system
	
	fmt.Printf("[UnifiedClosingHistory] Total history entries: %d\n", len(history))
	
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    history,
		"count":   len(history),
	})
}