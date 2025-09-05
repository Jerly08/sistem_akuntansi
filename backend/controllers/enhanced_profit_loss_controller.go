package controllers

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"app-sistem-akuntansi/services"
)

type EnhancedProfitLossController struct {
	enhancedPLService *services.EnhancedProfitLossService
}

func NewEnhancedProfitLossController(enhancedPLService *services.EnhancedProfitLossService) *EnhancedProfitLossController {
	return &EnhancedProfitLossController{
		enhancedPLService: enhancedPLService,
	}
}

// GenerateEnhancedProfitLoss generates comprehensive P&L report
// @Summary Generate Enhanced Profit & Loss Statement
// @Description Generate a comprehensive P&L statement with proper accounting categorization and financial metrics
// @Tags Enhanced Reports
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Param format query string false "Output format (json, pdf)" default(json)
// @Success 200 {object} map[string]interface{} "Enhanced P&L data"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reports/enhanced/profit-loss [get]
func (eplc *EnhancedProfitLossController) GenerateEnhancedProfitLoss(c *gin.Context) {
	// Parse query parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	format := c.DefaultQuery("format", "json")

	// Validate required parameters
	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start_date and end_date are required",
			"code":  "MISSING_PARAMETERS",
		})
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid start_date format. Use YYYY-MM-DD",
			"code":  "INVALID_DATE_FORMAT",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid end_date format. Use YYYY-MM-DD",
			"code":  "INVALID_DATE_FORMAT",
		})
		return
	}

	// Validate date range
	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "end_date must be after start_date",
			"code":  "INVALID_DATE_RANGE",
		})
		return
	}

	// Generate enhanced P&L report
	plData, err := eplc.enhancedPLService.GenerateEnhancedProfitLoss(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate P&L statement: " + err.Error(),
			"code":  "GENERATION_ERROR",
		})
		return
	}

	// Handle different output formats
	switch format {
	case "json":
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    plData,
			"message": "Enhanced P&L statement generated successfully",
		})

	case "pdf":
		// For PDF format, you would generate a PDF file here
		// This is a placeholder - implement PDF generation as needed
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "PDF format not yet implemented",
			"code":  "NOT_IMPLEMENTED",
		})

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid format. Supported formats: json, pdf",
			"code":  "INVALID_FORMAT",
		})
	}
}

// GetFinancialMetrics returns key financial metrics from P&L
// @Summary Get Key Financial Metrics
// @Description Get key profitability and performance metrics from P&L data
// @Tags Enhanced Reports
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "Financial metrics"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reports/enhanced/financial-metrics [get]
func (eplc *EnhancedProfitLossController) GetFinancialMetrics(c *gin.Context) {
	// Parse query parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Validate required parameters
	if startDateStr == "" || endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "start_date and end_date are required",
		})
		return
	}

	// Parse dates
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid start_date format. Use YYYY-MM-DD",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid end_date format. Use YYYY-MM-DD",
		})
		return
	}

	// Generate P&L data
	plData, err := eplc.enhancedPLService.GenerateEnhancedProfitLoss(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate financial metrics: " + err.Error(),
		})
		return
	}

	// Extract key metrics
	metrics := map[string]interface{}{
		"period": map[string]string{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		},
		"revenue_metrics": map[string]interface{}{
			"total_revenue":     plData.Revenue.TotalRevenue,
			"sales_revenue":     plData.Revenue.SalesRevenue.Subtotal,
			"service_revenue":   plData.Revenue.ServiceRevenue.Subtotal,
			"other_revenue":     plData.Revenue.OtherRevenue.Subtotal,
		},
		"profitability_metrics": map[string]interface{}{
			"gross_profit":        plData.GrossProfit,
			"gross_profit_margin": plData.GrossProfitMargin,
			"operating_income":    plData.OperatingIncome,
			"operating_margin":    plData.OperatingMargin,
			"ebitda":             plData.EBITDA,
			"ebitda_margin":      plData.EBITDAMargin,
			"net_income":         plData.NetIncome,
			"net_income_margin":  plData.NetIncomeMargin,
		},
		"cost_structure": map[string]interface{}{
			"total_cogs":             plData.CostOfGoodsSold.TotalCOGS,
			"total_operating_expenses": plData.OperatingExpenses.TotalOpex,
			"tax_expense":            plData.TaxExpense,
			"tax_rate":               plData.TaxRate,
		},
		"efficiency_ratios": map[string]interface{}{
			"cogs_as_percent_of_sales": func() float64 {
				if plData.Revenue.TotalRevenue != 0 {
					return (plData.CostOfGoodsSold.TotalCOGS / plData.Revenue.TotalRevenue) * 100
				}
				return 0
			}(),
			"opex_as_percent_of_sales": func() float64 {
				if plData.Revenue.TotalRevenue != 0 {
					return (plData.OperatingExpenses.TotalOpex / plData.Revenue.TotalRevenue) * 100
				}
				return 0
			}(),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    metrics,
		"message": "Financial metrics calculated successfully",
	})
}

// CompareProfitLoss compares P&L between two periods
// @Summary Compare P&L Between Periods
// @Description Compare P&L metrics between current period and previous period
// @Tags Enhanced Reports
// @Accept json
// @Produce json
// @Param current_start query string true "Current period start date (YYYY-MM-DD)"
// @Param current_end query string true "Current period end date (YYYY-MM-DD)"
// @Param previous_start query string true "Previous period start date (YYYY-MM-DD)"
// @Param previous_end query string true "Previous period end date (YYYY-MM-DD)"
// @Success 200 {object} map[string]interface{} "P&L comparison data"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/reports/enhanced/profit-loss-comparison [get]
func (eplc *EnhancedProfitLossController) CompareProfitLoss(c *gin.Context) {
	// Parse current period
	currentStartStr := c.Query("current_start")
	currentEndStr := c.Query("current_end")
	previousStartStr := c.Query("previous_start")
	previousEndStr := c.Query("previous_end")

	if currentStartStr == "" || currentEndStr == "" || previousStartStr == "" || previousEndStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "All date parameters are required",
		})
		return
	}

	// Parse dates
	currentStart, err := time.Parse("2006-01-02", currentStartStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid current_start format",
		})
		return
	}

	currentEnd, err := time.Parse("2006-01-02", currentEndStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid current_end format",
		})
		return
	}

	previousStart, err := time.Parse("2006-01-02", previousStartStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid previous_start format",
		})
		return
	}

	previousEnd, err := time.Parse("2006-01-02", previousEndStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid previous_end format",
		})
		return
	}

	// Generate P&L for both periods
	currentPL, err := eplc.enhancedPLService.GenerateEnhancedProfitLoss(currentStart, currentEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate current period P&L: " + err.Error(),
		})
		return
	}

	previousPL, err := eplc.enhancedPLService.GenerateEnhancedProfitLoss(previousStart, previousEnd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate previous period P&L: " + err.Error(),
		})
		return
	}

	// Calculate growth rates and variances
	comparison := map[string]interface{}{
		"current_period":  currentPL,
		"previous_period": previousPL,
		"variance_analysis": map[string]interface{}{
			"revenue_growth": calculateGrowthRate(currentPL.Revenue.TotalRevenue, previousPL.Revenue.TotalRevenue),
			"gross_profit_growth": calculateGrowthRate(currentPL.GrossProfit, previousPL.GrossProfit),
			"operating_income_growth": calculateGrowthRate(currentPL.OperatingIncome, previousPL.OperatingIncome),
			"net_income_growth": calculateGrowthRate(currentPL.NetIncome, previousPL.NetIncome),
			"margin_changes": map[string]float64{
				"gross_margin_change": currentPL.GrossProfitMargin - previousPL.GrossProfitMargin,
				"operating_margin_change": currentPL.OperatingMargin - previousPL.OperatingMargin,
				"net_margin_change": currentPL.NetIncomeMargin - previousPL.NetIncomeMargin,
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    comparison,
		"message": "P&L comparison completed successfully",
	})
}

// Helper function to calculate growth rate
func calculateGrowthRate(current, previous float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100 // 100% growth from 0
	}
	return ((current - previous) / previous) * 100
}
