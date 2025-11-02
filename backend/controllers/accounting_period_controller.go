package controllers

import (
	"net/http"
	"strconv"

	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
)

// AccountingPeriodController handles accounting period management
type AccountingPeriodController struct {
	service *services.AccountingPeriodService
}

// NewAccountingPeriodController creates a new accounting period controller
func NewAccountingPeriodController(service *services.AccountingPeriodService) *AccountingPeriodController {
	return &AccountingPeriodController{
		service: service,
	}
}

// ListPeriods godoc
// @Summary List accounting periods
// @Description Get list of accounting periods with filtering
// @Tags Accounting Periods
// @Accept json
// @Produce json
// @Param year query int false "Filter by year"
// @Param month query int false "Filter by month"
// @Param status query string false "Filter by status (OPEN, CLOSED, LOCKED)"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/periods [get]
func (apc *AccountingPeriodController) ListPeriods(c *gin.Context) {
	var filter models.AccountingPeriodFilter

	// Parse query parameters
	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			filter.Year = &year
		}
	}

	if monthStr := c.Query("month"); monthStr != "" {
		if month, err := strconv.Atoi(monthStr); err == nil {
			filter.Month = &month
		}
	}

	if status := c.Query("status"); status != "" {
		filter.Status = status
	}

	if pageStr := c.Query("page"); pageStr != "" {
		filter.Page, _ = strconv.Atoi(pageStr)
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		filter.Limit, _ = strconv.Atoi(limitStr)
	}

	// Default pagination
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.Limit == 0 {
		filter.Limit = 12 // Default to 12 months
	}

	periods, total, err := apc.service.ListPeriods(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to list periods",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    periods,
		"total":   total,
		"page":    filter.Page,
		"limit":   filter.Limit,
	})
}

// GetCurrentPeriod godoc
// @Summary Get current accounting period
// @Description Get the current accounting period
// @Tags Accounting Periods
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/periods/current [get]
func (apc *AccountingPeriodController) GetCurrentPeriod(c *gin.Context) {
	period, err := apc.service.GetCurrentPeriod(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get current period",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    period,
	})
}

// ClosePeriod godoc
// @Summary Close an accounting period
// @Description Close an accounting period for the specified year and month
// @Tags Accounting Periods
// @Accept json
// @Produce json
// @Param year path int true "Year"
// @Param month path int true "Month"
// @Param request body models.AccountingPeriodCloseRequest true "Close request"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/periods/{year}/{month}/close [post]
func (apc *AccountingPeriodController) ClosePeriod(c *gin.Context) {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid year parameter",
		})
		return
	}

	month, err := strconv.Atoi(c.Param("month"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid month parameter",
		})
		return
	}

	var req models.AccountingPeriodCloseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty body
		req = models.AccountingPeriodCloseRequest{}
	}

	// Pass Gin context to preserve user_id from JWT middleware
	err = apc.service.ClosePeriod(c, year, month, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to close period",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Period closed successfully",
	})
}

// ReopenPeriod godoc
// @Summary Reopen a closed accounting period
// @Description Reopen a closed accounting period for the specified year and month
// @Tags Accounting Periods
// @Accept json
// @Produce json
// @Param year path int true "Year"
// @Param month path int true "Month"
// @Param request body map[string]string true "Reopen request with reason"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/periods/{year}/{month}/reopen [post]
func (apc *AccountingPeriodController) ReopenPeriod(c *gin.Context) {
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid year parameter",
		})
		return
	}

	month, err := strconv.Atoi(c.Param("month"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid month parameter",
		})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	if req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Reason is required to reopen a period",
		})
		return
	}

	// Pass Gin context to preserve user_id from JWT middleware
	err = apc.service.ReopenPeriod(c, year, month, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Failed to reopen period",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Period reopened successfully",
	})
}

// GetPeriodSummary godoc
// @Summary Get period summary statistics
// @Description Get summary statistics for accounting periods
// @Tags Accounting Periods
// @Accept json
// @Produce json
// @Param year query int false "Filter by year"
// @Param month query int false "Filter by month"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/periods/summary [get]
func (apc *AccountingPeriodController) GetPeriodSummary(c *gin.Context) {
	year, _ := strconv.Atoi(c.Query("year"))
	month, _ := strconv.Atoi(c.Query("month"))

	summary, err := apc.service.GetPeriodSummary(c.Request.Context(), year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to get period summary",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}
