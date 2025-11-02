package services

import (
	"context"
	"fmt"
	"time"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/utils"
	"gorm.io/gorm"
)

// AccountingPeriodService handles accounting period management
type AccountingPeriodService struct {
	db     *gorm.DB
	logger *utils.JournalLogger
}

// NewAccountingPeriodService creates a new accounting period service
func NewAccountingPeriodService(db *gorm.DB) *AccountingPeriodService {
	return &AccountingPeriodService{
		db:     db,
		logger: utils.NewJournalLogger(db),
	}
}

// CreatePeriod creates a new accounting period
func (aps *AccountingPeriodService) CreatePeriod(ctx context.Context, req models.AccountingPeriodRequest) (*models.AccountingPeriod, error) {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("user context required: %v", err)
	}

	// Check if period already exists
	var existingPeriod models.AccountingPeriod
	err = aps.db.Where("year = ? AND month = ?", req.Year, req.Month).First(&existingPeriod).Error
	if err == nil {
		return nil, fmt.Errorf("accounting period %04d-%02d already exists", req.Year, req.Month)
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing period: %v", err)
	}

	period := &models.AccountingPeriod{
		Year:  req.Year,
		Month: req.Month,
		Notes: req.Notes,
	}

	if err := aps.db.Create(period).Error; err != nil {
		aps.logger.LogValidationError(ctx, nil, "period_creation", err)
		return nil, fmt.Errorf("failed to create accounting period: %v", err)
	}

	aps.logger.LogProcessingInfo(ctx, "Accounting period created", map[string]interface{}{
		"period_id":   period.ID,
		"period_name": period.PeriodName,
		"year":        req.Year,
		"month":       req.Month,
		"user_id":     userID,
	})

	return period, nil
}

// GetCurrentPeriod returns the current accounting period
func (aps *AccountingPeriodService) GetCurrentPeriod(ctx context.Context) (*models.AccountingPeriod, error) {
	now := time.Now()
	var period models.AccountingPeriod

	err := aps.db.Where("year = ? AND month = ?", now.Year(), int(now.Month())).
		Preload("ClosedByUser").
		Preload("LockedByUser").
		First(&period).Error

	if err == gorm.ErrRecordNotFound {
		// Auto-create current period if it doesn't exist
		return aps.CreatePeriod(ctx, models.AccountingPeriodRequest{
			Year:  now.Year(),
			Month: int(now.Month()),
			Notes: "Auto-created current period",
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get current period: %v", err)
	}

	return &period, nil
}

// ClosePeriod closes an accounting period
func (aps *AccountingPeriodService) ClosePeriod(ctx context.Context, year, month int, req models.AccountingPeriodCloseRequest) error {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("user context required: %v", err)
	}

	return aps.db.Transaction(func(tx *gorm.DB) error {
		var period models.AccountingPeriod
		if err := tx.Where("year = ? AND month = ?", year, month).First(&period).Error; err != nil {
			return fmt.Errorf("period not found: %v", err)
		}

		// Check if period can be closed
		if period.IsClosed {
			return fmt.Errorf("period %s is already closed", period.PeriodName)
		}

		if period.IsLocked {
			return fmt.Errorf("period %s is locked and cannot be modified", period.PeriodName)
		}

		// Validate all journal entries are balanced for this period
		if err := aps.validatePeriodForClosing(ctx, tx, &period); err != nil {
			return fmt.Errorf("period validation failed: %v", err)
		}

		// Close the period
		if err := period.Close(userID); err != nil {
			return err
		}

		if req.Notes != "" {
			period.Notes = req.Notes
		}

		if err := tx.Save(&period).Error; err != nil {
			return fmt.Errorf("failed to close period: %v", err)
		}

		aps.logger.LogProcessingInfo(ctx, "Accounting period closed", map[string]interface{}{
			"period_id":   period.ID,
			"period_name": period.PeriodName,
			"closed_by":   userID,
			"notes":       req.Notes,
		})

		return nil
	})
}

// LockPeriod locks an accounting period
func (aps *AccountingPeriodService) LockPeriod(ctx context.Context, year, month int, notes string) error {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("user context required: %v", err)
	}

	return aps.db.Transaction(func(tx *gorm.DB) error {
		var period models.AccountingPeriod
		if err := tx.Where("year = ? AND month = ?", year, month).First(&period).Error; err != nil {
			return fmt.Errorf("period not found: %v", err)
		}

		if err := period.Lock(userID); err != nil {
			return err
		}

		if notes != "" {
			period.Notes = notes
		}

		if err := tx.Save(&period).Error; err != nil {
			return fmt.Errorf("failed to lock period: %v", err)
		}

		aps.logger.LogProcessingInfo(ctx, "Accounting period locked", map[string]interface{}{
			"period_id":   period.ID,
			"period_name": period.PeriodName,
			"locked_by":   userID,
			"notes":       notes,
		})

		return nil
	})
}

// ReopenPeriod reopens a closed accounting period
func (aps *AccountingPeriodService) ReopenPeriod(ctx context.Context, year, month int, reason string) error {
	userID, err := utils.GetUserIDFromContext(ctx)
	if err != nil {
		return fmt.Errorf("user context required: %v", err)
	}

	return aps.db.Transaction(func(tx *gorm.DB) error {
		var period models.AccountingPeriod
		if err := tx.Where("year = ? AND month = ?", year, month).First(&period).Error; err != nil {
			return fmt.Errorf("period not found: %v", err)
		}

		if err := period.Reopen(); err != nil {
			return err
		}

		if reason != "" {
			period.Notes = fmt.Sprintf("Reopened: %s", reason)
		}

		if err := tx.Save(&period).Error; err != nil {
			return fmt.Errorf("failed to reopen period: %v", err)
		}

		aps.logger.LogProcessingInfo(ctx, "Accounting period reopened", map[string]interface{}{
			"period_id":   period.ID,
			"period_name": period.PeriodName,
			"reopened_by": userID,
			"reason":      reason,
		})

		return nil
	})
}

// ValidatePeriodPosting validates if a journal entry can be posted to a specific period
func (aps *AccountingPeriodService) ValidatePeriodPosting(ctx context.Context, entryDate time.Time) error {
	year := entryDate.Year()
	month := int(entryDate.Month())

	var period models.AccountingPeriod
	// Support both standard context and gin.Context
	err := aps.db.WithContext(ctx).Where("year = ? AND month = ?", year, month).First(&period).Error

	if err == gorm.ErrRecordNotFound {
		// Period doesn't exist, check if it's too old or too future
		if entryDate.Before(time.Now().AddDate(-2, 0, 0)) {
			return fmt.Errorf("cannot post entries more than 2 years old")
		}
		if entryDate.After(time.Now().AddDate(0, 0, 7)) {
			return fmt.Errorf("cannot post entries more than 7 days in the future")
		}
		// Auto-create period if within allowed range
		_, err := aps.CreatePeriod(ctx, models.AccountingPeriodRequest{
			Year:  year,
			Month: month,
			Notes: "Auto-created for journal entry",
		})
		return err
	}

	if err != nil {
		return fmt.Errorf("failed to check period: %v", err)
	}

	if !period.CanPost() {
		return fmt.Errorf("cannot post to %s period (status: %s)", period.PeriodName, period.GetStatus())
	}

	return nil
}

// validatePeriodForClosing validates if a period can be closed
func (aps *AccountingPeriodService) validatePeriodForClosing(ctx context.Context, tx *gorm.DB, period *models.AccountingPeriod) error {
	// Calculate period end date
	periodEnd := period.EndDate
	if periodEnd.IsZero() {
		// If EndDate not set, calculate it (last day of month)
		periodEnd = time.Date(period.Year, time.Month(period.Month+1), 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)
	}

	// Check for unbalanced journal entries using BETWEEN (avoids DATE_TRUNC ambiguity)
	var unbalancedCount int64
	err := tx.Model(&models.JournalEntry{}).
		Where("entry_date >= ? AND entry_date <= ?", period.StartDate, periodEnd).
		Where("is_balanced = false").
		Count(&unbalancedCount).Error

	if err != nil {
		return fmt.Errorf("failed to check unbalanced entries: %v", err)
	}

	if unbalancedCount > 0 {
		return fmt.Errorf("cannot close period with %d unbalanced journal entries", unbalancedCount)
	}

	// Check for draft journal entries
	var draftCount int64
	err = tx.Model(&models.JournalEntry{}).
		Where("entry_date >= ? AND entry_date <= ?", period.StartDate, periodEnd).
		Where("status = ?", models.JournalStatusDraft).
		Count(&draftCount).Error

	if err != nil {
		return fmt.Errorf("failed to check draft entries: %v", err)
	}

	if draftCount > 0 {
		aps.logger.LogWarning(ctx, fmt.Sprintf("Period %s has %d draft journal entries", period.PeriodName, draftCount), map[string]interface{}{
			"period_id":    period.ID,
			"draft_count":  draftCount,
		})
	}

	return nil
}

// GetPeriodSummary returns summary statistics for a period
func (aps *AccountingPeriodService) GetPeriodSummary(ctx context.Context, year, month int) (*models.AccountingPeriodSummary, error) {
	var summary models.AccountingPeriodSummary

	// Get total periods
	aps.db.Model(&models.AccountingPeriod{}).Count(&summary.TotalPeriods)

	// Get periods by status
	aps.db.Model(&models.AccountingPeriod{}).Where("is_closed = false AND is_locked = false").Count(&summary.OpenPeriods)
	aps.db.Model(&models.AccountingPeriod{}).Where("is_closed = true AND is_locked = false").Count(&summary.ClosedPeriods)
	aps.db.Model(&models.AccountingPeriod{}).Where("is_locked = true").Count(&summary.LockedPeriods)

	// Get current period
	currentPeriod, err := aps.GetCurrentPeriod(ctx)
	if err == nil {
		summary.CurrentPeriod = currentPeriod
	}

	return &summary, nil
}

// ListPeriods returns a list of accounting periods with filtering
func (aps *AccountingPeriodService) ListPeriods(ctx context.Context, filter models.AccountingPeriodFilter) ([]models.AccountingPeriod, int64, error) {
	var periods []models.AccountingPeriod
	var total int64

	query := aps.db.Model(&models.AccountingPeriod{}).
		Preload("ClosedByUser").
		Preload("LockedByUser")

	// Apply filters
	if filter.Year != nil {
		query = query.Where("year = ?", *filter.Year)
	}

	if filter.Month != nil {
		query = query.Where("month = ?", *filter.Month)
	}

	if filter.IsClosed != nil {
		query = query.Where("is_closed = ?", *filter.IsClosed)
	}

	if filter.IsLocked != nil {
		query = query.Where("is_locked = ?", *filter.IsLocked)
	}

	if filter.Status != "" {
		switch filter.Status {
		case models.PeriodStatusOpen:
			query = query.Where("is_closed = false AND is_locked = false")
		case models.PeriodStatusClosed:
			query = query.Where("is_closed = true AND is_locked = false")
		case models.PeriodStatusLocked:
			query = query.Where("is_locked = true")
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count periods: %v", err)
	}

	// Apply pagination
	if filter.Page > 0 && filter.Limit > 0 {
		offset := (filter.Page - 1) * filter.Limit
		query = query.Offset(offset).Limit(filter.Limit)
	}

	// Order by year and month descending
	query = query.Order("year DESC, month DESC")

	if err := query.Find(&periods).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get periods: %v", err)
	}

	return periods, total, nil
}

// GetPeriodJournalStats returns journal entry statistics for a specific period
func (aps *AccountingPeriodService) GetPeriodJournalStats(ctx context.Context, year, month int) (map[string]interface{}, error) {
	periodStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	var stats struct {
		TotalEntries    int64   `json:"total_entries"`
		PostedEntries   int64   `json:"posted_entries"`
		DraftEntries    int64   `json:"draft_entries"`
		BalancedEntries int64   `json:"balanced_entries"`
		TotalDebit      float64 `json:"total_debit"`
		TotalCredit     float64 `json:"total_credit"`
	}

	// Get total entries
	aps.db.Model(&models.JournalEntry{}).
		Where("entry_date BETWEEN ? AND ?", periodStart, periodEnd).
		Count(&stats.TotalEntries)

	// Get posted entries
	aps.db.Model(&models.JournalEntry{}).
		Where("entry_date BETWEEN ? AND ? AND status = ?", periodStart, periodEnd, models.JournalStatusPosted).
		Count(&stats.PostedEntries)

	// Get draft entries
	aps.db.Model(&models.JournalEntry{}).
		Where("entry_date BETWEEN ? AND ? AND status = ?", periodStart, periodEnd, models.JournalStatusDraft).
		Count(&stats.DraftEntries)

	// Get balanced entries
	aps.db.Model(&models.JournalEntry{}).
		Where("entry_date BETWEEN ? AND ? AND is_balanced = true", periodStart, periodEnd).
		Count(&stats.BalancedEntries)

	// Get total amounts
	aps.db.Model(&models.JournalEntry{}).
		Where("entry_date BETWEEN ? AND ? AND status = ?", periodStart, periodEnd, models.JournalStatusPosted).
		Select("COALESCE(SUM(total_debit), 0) as total_debit, COALESCE(SUM(total_credit), 0) as total_credit").
		Scan(&stats)

	return map[string]interface{}{
		"period":           fmt.Sprintf("%04d-%02d", year, month),
		"total_entries":    stats.TotalEntries,
		"posted_entries":   stats.PostedEntries,
		"draft_entries":    stats.DraftEntries,
		"balanced_entries": stats.BalancedEntries,
		"total_debit":      stats.TotalDebit,
		"total_credit":     stats.TotalCredit,
		"balance_check":    stats.TotalDebit == stats.TotalCredit,
	}, nil
}