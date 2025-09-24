package models

import (
	"fmt"
	"time"
	"gorm.io/gorm"
)

// AccountingPeriod represents accounting period management
type AccountingPeriod struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Year        int            `json:"year" gorm:"not null;index"`
	Month       int            `json:"month" gorm:"not null;index"`
	PeriodName  string         `json:"period_name" gorm:"size:20"` // e.g., "2024-01"
	StartDate   time.Time      `json:"start_date"`
	EndDate     time.Time      `json:"end_date"`
	IsClosed    bool           `json:"is_closed" gorm:"default:false"`
	IsLocked    bool           `json:"is_locked" gorm:"default:false"`
	ClosedBy    *uint          `json:"closed_by" gorm:"index"`
	ClosedAt    *time.Time     `json:"closed_at"`
	LockedBy    *uint          `json:"locked_by" gorm:"index"`
	LockedAt    *time.Time     `json:"locked_at"`
	Notes       string         `json:"notes" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	ClosedByUser *User `json:"closed_by_user,omitempty" gorm:"foreignKey:ClosedBy"`
	LockedByUser *User `json:"locked_by_user,omitempty" gorm:"foreignKey:LockedBy"`
}

// AccountingPeriodStatus constants
const (
	PeriodStatusOpen   = "OPEN"
	PeriodStatusClosed = "CLOSED"
	PeriodStatusLocked = "LOCKED"
)

// BeforeCreate hook to set period name
func (ap *AccountingPeriod) BeforeCreate(tx *gorm.DB) error {
	if ap.PeriodName == "" {
		ap.PeriodName = fmt.Sprintf("%04d-%02d", ap.Year, ap.Month)
	}
	
	// Set start and end dates if not provided
	if ap.StartDate.IsZero() {
		ap.StartDate = time.Date(ap.Year, time.Month(ap.Month), 1, 0, 0, 0, 0, time.UTC)
	}
	
	if ap.EndDate.IsZero() {
		// Last day of the month
		nextMonth := ap.StartDate.AddDate(0, 1, 0)
		ap.EndDate = nextMonth.Add(-time.Second)
	}
	
	return nil
}

// GetStatus returns the current status of the period
func (ap *AccountingPeriod) GetStatus() string {
	if ap.IsLocked {
		return PeriodStatusLocked
	}
	if ap.IsClosed {
		return PeriodStatusClosed
	}
	return PeriodStatusOpen
}

// CanPost checks if journal entries can be posted to this period
func (ap *AccountingPeriod) CanPost() bool {
	return !ap.IsClosed && !ap.IsLocked
}

// Close closes the accounting period
func (ap *AccountingPeriod) Close(userID uint) error {
	if ap.IsClosed {
		return fmt.Errorf("period %s is already closed", ap.PeriodName)
	}
	
	if ap.IsLocked {
		return fmt.Errorf("period %s is locked and cannot be closed", ap.PeriodName)
	}
	
	now := time.Now()
	ap.IsClosed = true
	ap.ClosedBy = &userID
	ap.ClosedAt = &now
	
	return nil
}

// Lock locks the accounting period (prevents any modifications)
func (ap *AccountingPeriod) Lock(userID uint) error {
	if !ap.IsClosed {
		return fmt.Errorf("period %s must be closed before it can be locked", ap.PeriodName)
	}
	
	if ap.IsLocked {
		return fmt.Errorf("period %s is already locked", ap.PeriodName)
	}
	
	now := time.Now()
	ap.IsLocked = true
	ap.LockedBy = &userID
	ap.LockedAt = &now
	
	return nil
}

// Reopen reopens a closed period (only if not locked)
func (ap *AccountingPeriod) Reopen() error {
	if ap.IsLocked {
		return fmt.Errorf("period %s is locked and cannot be reopened", ap.PeriodName)
	}
	
	if !ap.IsClosed {
		return fmt.Errorf("period %s is already open", ap.PeriodName)
	}
	
	ap.IsClosed = false
	ap.ClosedBy = nil
	ap.ClosedAt = nil
	
	return nil
}

// Request DTOs
type AccountingPeriodRequest struct {
	Year  int    `json:"year" binding:"required,min=2020,max=2030"`
	Month int    `json:"month" binding:"required,min=1,max=12"`
	Notes string `json:"notes"`
}

type AccountingPeriodCloseRequest struct {
	Notes string `json:"notes"`
}

type AccountingPeriodFilter struct {
	Year     *int   `json:"year"`
	Month    *int   `json:"month"`
	Status   string `json:"status"` // OPEN, CLOSED, LOCKED
	IsClosed *bool  `json:"is_closed"`
	IsLocked *bool  `json:"is_locked"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
}

// Response DTOs
type AccountingPeriodSummary struct {
	TotalPeriods   int64 `json:"total_periods"`
	OpenPeriods    int64 `json:"open_periods"`
	ClosedPeriods  int64 `json:"closed_periods"`
	LockedPeriods  int64 `json:"locked_periods"`
	CurrentPeriod  *AccountingPeriod `json:"current_period"`
}