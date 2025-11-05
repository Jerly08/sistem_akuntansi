package middleware

import (
	"encoding/json"
	"net/http"
	"time"

	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

// PeriodValidationMiddleware checks if transaction date is in closed period
type PeriodValidationMiddleware struct {
	periodService *services.PeriodClosingService
}

// NewPeriodValidationMiddleware creates a new period validation middleware
func NewPeriodValidationMiddleware(periodService *services.PeriodClosingService) *PeriodValidationMiddleware {
	return &PeriodValidationMiddleware{
		periodService: periodService,
	}
}

// ValidateTransactionPeriod checks if the transaction date is in a closed period
func (pvm *PeriodValidationMiddleware) ValidateTransactionPeriod() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for GET requests and non-transactional endpoints
		if c.Request.Method == "GET" || c.Request.Method == "DELETE" {
			c.Next()
			return
		}

		// Get the raw body to check for date fields
		bodyBytes, err := c.GetRawData()
		if err != nil {
			c.Next()
			return
		}

		// Restore body for next handlers
		c.Request.Body = &bodyReader{data: bodyBytes}

		// Parse body to check for date fields
		var requestBody map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &requestBody); err != nil {
			c.Next()
			return
		}

		// Check various date field names used in the system
		dateFields := []string{
			"entry_date",     // Journal entries
			"sale_date",      // Sales
			"purchase_date",  // Purchases
			"payment_date",   // Payments
			"date",          // Generic date field
			"transaction_date", // Alternative naming
		}

		var transactionDate *time.Time
		for _, field := range dateFields {
			if dateStr, ok := requestBody[field].(string); ok && dateStr != "" {
				if parsedDate, err := parseDate(dateStr); err == nil {
					transactionDate = &parsedDate
					break
				}
			}
		}

		// If no date found, let the request proceed (will be validated by controller)
		if transactionDate == nil {
			c.Next()
			return
		}

		// Check if date is in closed period
		isClosed, err := pvm.periodService.IsDateInClosedPeriod(c.Request.Context(), *transactionDate)
		if err != nil {
			// Log error but don't block the request
			c.Next()
			return
		}

		if isClosed {
			// Find the period details for better error message
			periodInfo := pvm.periodService.GetPeriodInfoForDate(c.Request.Context(), *transactionDate)
			
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Cannot create or modify transaction in closed period",
				"code":    "PERIOD_CLOSED",
				"details": "The selected date falls within a closed accounting period",
				"period":  periodInfo,
				"date":    transactionDate.Format("2006-01-02"),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// parseDate tries to parse date string in various formats
func parseDate(dateStr string) (time.Time, error) {
	// Try common date formats
	formats := []string{
		"2006-01-02",                // ISO date
		"2006-01-02T15:04:05Z",     // ISO datetime with Z
		"2006-01-02T15:04:05",      // ISO datetime without timezone
		"2006-01-02 15:04:05",      // Space separator
		time.RFC3339,               // Standard RFC3339
	}

	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date, nil
		}
	}

	return time.Time{}, &time.ParseError{
		Layout:     "various formats",
		Value:      dateStr,
		LayoutElem: "",
		ValueElem:  "",
	}
}

// bodyReader implements io.ReadCloser for restoring request body
type bodyReader struct {
	data   []byte
	offset int
}

func (r *bodyReader) Read(p []byte) (n int, err error) {
	if r.offset >= len(r.data) {
		return 0, nil
	}
	n = copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

func (r *bodyReader) Close() error {
	return nil
}