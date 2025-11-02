package middleware

import (
	"fmt"
	"net/http"
	"time"

	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PeriodValidationMiddleware enforces accounting period restrictions
type PeriodValidationMiddleware struct {
	db             *gorm.DB
	periodService  *services.AccountingPeriodService
}

// NewPeriodValidationMiddleware creates a new period validation middleware
func NewPeriodValidationMiddleware(db *gorm.DB, periodService *services.AccountingPeriodService) *PeriodValidationMiddleware {
	return &PeriodValidationMiddleware{
		db:            db,
		periodService: periodService,
	}
}

// ValidateEntryDate validates if a transaction can be posted to a specific date's period
// This middleware should be used on routes that create transactions with entry_date
func (pvm *PeriodValidationMiddleware) ValidateEntryDate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip validation for GET requests (read-only)
		if c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		// Try to extract entry_date from request body
		var req struct {
			EntryDate   string    `json:"entry_date"`
			Date        string    `json:"date"`
			InvoiceDate string    `json:"invoice_date"`
			PaymentDate string    `json:"payment_date"`
		}

		// Bind JSON without strict validation
		if err := c.ShouldBindJSON(&req); err != nil {
			// If can't bind, let the controller handle it
			c.Next()
			return
		}

		// Determine which date field to use
		var entryDate time.Time
		var dateStr string

		if req.EntryDate != "" {
			dateStr = req.EntryDate
		} else if req.Date != "" {
			dateStr = req.Date
		} else if req.InvoiceDate != "" {
			dateStr = req.InvoiceDate
		} else if req.PaymentDate != "" {
			dateStr = req.PaymentDate
		} else {
			// No date field found, skip validation
			c.Next()
			return
		}

		// Parse date
		var err error
		// Try multiple date formats
		formats := []string{
			"2006-01-02",
			"2006-01-02T15:04:05Z07:00",
			time.RFC3339,
		}

		for _, format := range formats {
			entryDate, err = time.Parse(format, dateStr)
			if err == nil {
				break
			}
		}

		if err != nil {
			// Can't parse date, let controller handle it
			c.Next()
			return
		}

		// Validate period
		if err := pvm.periodService.ValidatePeriodPosting(c.Request.Context(), entryDate); err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Period validation failed",
				"details": err.Error(),
				"code":    "PERIOD_CLOSED",
				"period":  entryDate.Format("2006-01"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateEntryDateFromPath validates period for routes with :year/:month params
func (pvm *PeriodValidationMiddleware) ValidateEntryDateFromPath() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip validation for GET requests
		if c.Request.Method == http.MethodGet {
			c.Next()
			return
		}

		// Extract year and month from path params
		yearStr := c.Param("year")
		monthStr := c.Param("month")

		if yearStr == "" || monthStr == "" {
			c.Next()
			return
		}

		// Parse year and month
		var year, month int
		if _, err := fmt.Sscanf(yearStr, "%d", &year); err != nil {
			c.Next()
			return
		}
		if _, err := fmt.Sscanf(monthStr, "%d", &month); err != nil {
			c.Next()
			return
		}

		// Create date from year/month (first day of month)
		entryDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

		// Validate period
		if err := pvm.periodService.ValidatePeriodPosting(c.Request.Context(), entryDate); err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Period validation failed",
				"details": err.Error(),
				"code":    "PERIOD_CLOSED",
				"period":  fmt.Sprintf("%04d-%02d", year, month),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// BypassForAdmin allows admin to bypass period restrictions
func (pvm *PeriodValidationMiddleware) BypassForAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is admin
		role, exists := c.Get("role")
		if exists && role == "admin" {
			// Admin can bypass, continue without validation
			c.Next()
			return
		}

		// Non-admin, apply validation
		pvm.ValidateEntryDate()(c)
	}
}
