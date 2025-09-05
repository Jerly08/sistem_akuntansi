package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"app-sistem-akuntansi/models"
	"app-sistem-akuntansi/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardController struct {
	DB                     *gorm.DB
	stockMonitoringService *services.StockMonitoringService
	dashboardService       *services.DashboardService
}

func NewDashboardController(db *gorm.DB, stockMonitoringService *services.StockMonitoringService) *DashboardController {
	return &DashboardController{
		DB:                     db,
		stockMonitoringService: stockMonitoringService,
		dashboardService:       services.NewDashboardService(db),
	}
}

// GetDashboardSummary returns comprehensive dashboard data
func (dc *DashboardController) GetDashboardSummary(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")
	
	summary := make(map[string]interface{})
	
	// Get stock alerts for inventory managers and admins
	userRoleLower := strings.ToLower(userRole)
	if userRoleLower == "admin" || userRoleLower == "inventory_manager" {
		stockAlerts, err := dc.stockMonitoringService.GetStockAlerts()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stock alerts"})
			return
		}
		summary["stock_alerts"] = stockAlerts
		
		// Get active alerts for banner
		activeAlerts, err := dc.stockMonitoringService.GetActiveStockAlerts()
		if err == nil {
			summary["active_stock_alerts"] = activeAlerts
			summary["has_stock_alerts"] = len(activeAlerts) > 0
		}
	}
	
	// Get general statistics
	stats, err := dc.getGeneralStatistics(userRole)
	if err == nil {
		summary["statistics"] = stats
	}
	
	// Get recent activities
	activities, err := dc.getRecentActivities(userID, userRole)
	if err == nil {
		summary["recent_activities"] = activities
	}
	
	// Get notification count
	var unreadCount int64
	dc.DB.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&unreadCount)
	summary["unread_notifications"] = unreadCount
	
	// Get MIN_STOCK notifications count specifically
	var minStockCount int64
	dc.DB.Model(&models.Notification{}).
		Where("user_id = ? AND type = ? AND is_read = ?", 
			userID, models.NotificationTypeLowStock, false).
		Count(&minStockCount)
	summary["min_stock_alerts_count"] = minStockCount
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Dashboard summary retrieved successfully",
		"data":    summary,
	})
}

// GetStockAlertsBanner returns stock alerts specifically for banner display
func (dc *DashboardController) GetStockAlertsBanner(c *gin.Context) {
	userRole := c.GetString("user_role")
	
	// Only for authorized roles
	userRoleLower := strings.ToLower(userRole)
	if userRoleLower != "admin" && userRoleLower != "inventory_manager" && userRoleLower != "director" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to view stock alerts"})
		return
	}
	
	// Get active stock alerts
	activeAlerts, err := dc.stockMonitoringService.GetActiveStockAlerts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get stock alerts"})
		return
	}
	
	// Format alerts for banner display
	var bannerAlerts []map[string]interface{}
	for _, alert := range activeAlerts {
		bannerAlert := map[string]interface{}{
			"id":              alert.ID,
			"product_id":      alert.ProductID,
			"product_name":    alert.Product.Name,
			"product_code":    alert.Product.Code,
			"current_stock":   alert.CurrentStock,
			"threshold_stock": alert.ThresholdStock,
			"alert_type":      alert.AlertType,
			"urgency":         dc.getUrgencyLevel(alert),
			"message":         dc.formatAlertMessage(alert),
		}
		
		if alert.Product.Category != nil {
			bannerAlert["category_name"] = alert.Product.Category.Name
		}
		
		bannerAlerts = append(bannerAlerts, bannerAlert)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Stock alerts retrieved successfully",
		"data": gin.H{
			"alerts":      bannerAlerts,
			"total_count": len(bannerAlerts),
			"show_banner": len(bannerAlerts) > 0,
		},
	})
}

// DismissStockAlert allows users to dismiss a stock alert
func (dc *DashboardController) DismissStockAlert(c *gin.Context) {
	alertID := c.Param("id")
	userRole := c.GetString("user_role")
	
	userRoleLower := strings.ToLower(userRole)
	if userRoleLower != "admin" && userRoleLower != "inventory_manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to dismiss alerts"})
		return
	}
	
	var alert models.StockAlert
	if err := dc.DB.First(&alert, alertID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alert not found"})
		return
	}
	
	alert.Status = models.StockAlertStatusDismissed
	if err := dc.DB.Save(&alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to dismiss alert"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Alert dismissed successfully"})
}

// GetAnalytics returns dashboard analytics data with real growth calculations
func (dc *DashboardController) GetAnalytics(c *gin.Context) {
	userRole := c.GetString("user_role")
	
	// Only allow admin, finance, and director roles
	userRoleLower := strings.ToLower(userRole)
	if userRoleLower != "admin" && userRoleLower != "finance" && userRoleLower != "director" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to view analytics"})
		return
	}
	
	// Get comprehensive analytics with real growth calculations
	analytics, err := dc.dashboardService.GetDashboardAnalytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch dashboard analytics",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, analytics)
}

// GetQuickStats returns quick statistics for dashboard widgets
func (dc *DashboardController) GetQuickStats(c *gin.Context) {
	userRole := c.GetString("user_role")
	
	stats := make(map[string]interface{})
	
	// Total products
	var totalProducts int64
	dc.DB.Model(&models.Product{}).Where("is_active = ?", true).Count(&totalProducts)
	stats["total_products"] = totalProducts
	
	// Low stock products
	var lowStockCount int64
	dc.DB.Model(&models.Product{}).
		Where("stock <= min_stock AND min_stock > 0 AND is_active = ?", true).
		Count(&lowStockCount)
	stats["low_stock_count"] = lowStockCount
	
	// Out of stock products
	var outOfStockCount int64
	dc.DB.Model(&models.Product{}).
		Where("stock = 0 AND is_active = ?", true).
		Count(&outOfStockCount)
	stats["out_of_stock_count"] = outOfStockCount
	
	// Total categories
	var totalCategories int64
	dc.DB.Model(&models.ProductCategory{}).Where("is_active = ?", true).Count(&totalCategories)
	stats["total_categories"] = totalCategories
	
	// Role-specific stats
	userRoleLower := strings.ToLower(userRole)
	if userRoleLower == "admin" || userRoleLower == "finance" {
		// Total sales today
		var todaySales float64
		dc.DB.Model(&models.Sale{}).
			Where("DATE(created_at) = CURRENT_DATE").
			Select("COALESCE(SUM(total_amount), 0)").
			Scan(&todaySales)
		stats["today_sales"] = todaySales
		
		// Total purchases today (only approved)
		var todayPurchases float64
		dc.DB.Model(&models.Purchase{}).
			Where("DATE(created_at) = CURRENT_DATE AND status IN (?)", []string{"APPROVED", "COMPLETED"}).
			Select("COALESCE(SUM(total_amount), 0)").
			Scan(&todayPurchases)
		stats["today_purchases"] = todayPurchases
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Quick stats retrieved successfully",
		"data":    stats,
	})
}

// Private helper methods

func (dc *DashboardController) getGeneralStatistics(role string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Product statistics
	var totalProducts, activeProducts, inactiveProducts int64
	dc.DB.Model(&models.Product{}).Count(&totalProducts)
	dc.DB.Model(&models.Product{}).Where("is_active = ?", true).Count(&activeProducts)
	dc.DB.Model(&models.Product{}).Where("is_active = ?", false).Count(&inactiveProducts)
	
	stats["products"] = map[string]int64{
		"total":    totalProducts,
		"active":   activeProducts,
		"inactive": inactiveProducts,
	}
	
	// Contact statistics
	var totalContacts, customers, vendors int64
	dc.DB.Model(&models.Contact{}).Count(&totalContacts)
	dc.DB.Model(&models.Contact{}).Where("type = ?", models.ContactTypeCustomer).Count(&customers)
	dc.DB.Model(&models.Contact{}).Where("type = ?", models.ContactTypeVendor).Count(&vendors)
	
	stats["contacts"] = map[string]int64{
		"total":     totalContacts,
		"customers": customers,
		"vendors":   vendors,
	}
	
	return stats, nil
}

func (dc *DashboardController) getRecentActivities(userID uint, role string) ([]map[string]interface{}, error) {
	var activities []map[string]interface{}
	
	// Get recent audit logs
	var auditLogs []models.AuditLog
	query := dc.DB.Order("created_at DESC").Limit(10)
	
	if role != "admin" {
		query = query.Where("user_id = ?", userID)
	}
	
	if err := query.Find(&auditLogs).Error; err != nil {
		return activities, err
	}
	
	for _, log := range auditLogs {
		activity := map[string]interface{}{
			"id":          log.ID,
			"action":      log.Action,
			"table_name":  log.TableName,
			"record_id":   log.RecordID,
			"user_id":     log.UserID,
			"created_at":  log.CreatedAt,
		}
		activities = append(activities, activity)
	}
	
	return activities, nil
}

func (dc *DashboardController) getUrgencyLevel(alert models.StockAlert) string {
	percentageOfMin := float64(alert.CurrentStock) / float64(alert.ThresholdStock) * 100
	
	if percentageOfMin <= 25 {
		return "critical"
	} else if percentageOfMin <= 50 {
		return "high"
	} else if percentageOfMin <= 75 {
		return "medium"
	}
	return "low"
}

func (dc *DashboardController) formatAlertMessage(alert models.StockAlert) string {
	switch alert.AlertType {
	case models.StockAlertTypeLowStock:
		return fmt.Sprintf("%s is running low. Current stock: %d (Min: %d)",
			alert.Product.Name, alert.CurrentStock, alert.ThresholdStock)
	case models.StockAlertTypeOutOfStock:
		return fmt.Sprintf("%s is out of stock!", alert.Product.Name)
	default:
		return fmt.Sprintf("%s requires attention. Current stock: %d",
			alert.Product.Name, alert.CurrentStock)
	}
}

// getMonthlySalesData gets sales data for the last 7 months
func (dc *DashboardController) getMonthlySalesData() []map[string]interface{} {
	type MonthlyData struct {
		Month string  `json:"month"`
		Value float64 `json:"value"`
	}
	
	var results []MonthlyData
	dc.DB.Raw(`
		SELECT 
			TO_CHAR(created_at, 'Mon') as month,
			COALESCE(SUM(total_amount), 0) as value
		FROM sales 
		WHERE created_at >= CURRENT_DATE - INTERVAL '7 months'
			AND deleted_at IS NULL
		GROUP BY EXTRACT(YEAR FROM created_at), EXTRACT(MONTH FROM created_at), TO_CHAR(created_at, 'Mon')
		ORDER BY EXTRACT(YEAR FROM created_at), EXTRACT(MONTH FROM created_at)
	`).Scan(&results)
	
	// Convert to interface{} slice
	var data []map[string]interface{}
	for _, result := range results {
		data = append(data, map[string]interface{}{
			"month": result.Month,
			"value": result.Value,
		})
	}
	
	// Return only real data from database - no dummy data
	
	return data
}

// getMonthlyPurchasesData gets purchase data for the last 7 months
func (dc *DashboardController) getMonthlyPurchasesData() []map[string]interface{} {
	type MonthlyData struct {
		Month string  `json:"month"`
		Value float64 `json:"value"`
	}
	
	var results []MonthlyData
	dc.DB.Raw(`
		SELECT 
			TO_CHAR(created_at, 'Mon') as month,
			COALESCE(SUM(total_amount), 0) as value
		FROM purchases 
		WHERE created_at >= CURRENT_DATE - INTERVAL '7 months'
			AND deleted_at IS NULL
			AND status IN ('APPROVED', 'COMPLETED')
		GROUP BY EXTRACT(YEAR FROM created_at), EXTRACT(MONTH FROM created_at), TO_CHAR(created_at, 'Mon')
		ORDER BY EXTRACT(YEAR FROM created_at), EXTRACT(MONTH FROM created_at)
	`).Scan(&results)
	
	// Convert to interface{} slice
	var data []map[string]interface{}
	for _, result := range results {
		data = append(data, map[string]interface{}{
			"month": result.Month,
			"value": result.Value,
		})
	}
	
	// Return only real data from database - no dummy data
	
	return data
}

// calculateCashFlow calculates cash flow from sales and purchases data
func (dc *DashboardController) calculateCashFlow(sales, purchases []map[string]interface{}) []map[string]interface{} {
	var cashFlow []map[string]interface{}
	
	maxLen := len(sales)
	if len(purchases) > maxLen {
		maxLen = len(purchases)
	}
	
	for i := 0; i < maxLen; i++ {
		var month string
		var salesValue, purchasesValue, balance float64
		
		if i < len(sales) {
			month = sales[i]["month"].(string)
			salesValue = sales[i]["value"].(float64)
		}
		
		if i < len(purchases) {
			if month == "" {
				month = purchases[i]["month"].(string)
			}
			purchasesValue = purchases[i]["value"].(float64)
		}
		
		balance = salesValue - purchasesValue
		
		cashFlow = append(cashFlow, map[string]interface{}{
			"month":   month,
			"inflow":  salesValue,
			"outflow": purchasesValue,
			"balance": balance,
		})
	}
	
	return cashFlow
}

// getTopAccounts gets the top 5 accounts by balance
func (dc *DashboardController) getTopAccounts() []map[string]interface{} {
	type AccountData struct {
		Name    string  `json:"name"`
		Balance float64 `json:"balance"`
		Type    string  `json:"type"`
	}

	var results []AccountData
	dc.DB.Raw(`
		SELECT 
			name,
			ABS(balance) as balance,
			type
		FROM accounts 
		WHERE deleted_at IS NULL 
			AND is_active = true
			AND balance != 0
		ORDER BY ABS(balance) DESC
		LIMIT 5
	`).Scan(&results)

	// Convert to interface{} slice
	var data []map[string]interface{}
	for _, result := range results {
		data = append(data, map[string]interface{}{
			"name":    result.Name,
			"balance": result.Balance,
			"type":    result.Type,
		})
	}

	// Return empty array if no accounts found - no dummy data
	return data
}

// getRecentTransactions gets recent transactions based on user role
func (dc *DashboardController) getRecentTransactions(userRole string) []map[string]interface{} {
	type TransactionData struct {
		ID             uint    `json:"id"`
		TransactionID  string  `json:"transaction_id"`
		Description    string  `json:"description"`
		Amount         float64 `json:"amount"`
		Date           string  `json:"date"`
		Type           string  `json:"type"`
		AccountName    string  `json:"account_name"`
		ContactName    *string `json:"contact_name"`
		Status         string  `json:"status"`
	}

	var results []TransactionData
	
	// Get recent sales data
	dc.DB.Raw(`
		SELECT 
			s.id,
			s.code as transaction_id,
			s.notes as description,
			s.total_amount as amount,
			TO_CHAR(s.date, 'YYYY-MM-DD') as date,
			'SALE' as type,
			'Sales Transaction' as account_name,
			c.name as contact_name,
			s.status
		FROM sales s 
		LEFT JOIN contacts c ON s.customer_id = c.id
		WHERE s.deleted_at IS NULL
		ORDER BY s.created_at DESC
		LIMIT 5
	`).Scan(&results)

	// Get recent purchases data and append
	var purchaseResults []TransactionData
	dc.DB.Raw(`
		SELECT 
			p.id,
			p.code as transaction_id,
			p.notes as description,
			p.total_amount as amount,
			TO_CHAR(p.date, 'YYYY-MM-DD') as date,
			'PURCHASE' as type,
			'Purchase Transaction' as account_name,
			c.name as contact_name,
			p.status
		FROM purchases p 
		LEFT JOIN contacts c ON p.vendor_id = c.id
		WHERE p.deleted_at IS NULL
		ORDER BY p.created_at DESC
		LIMIT 5
	`).Scan(&purchaseResults)

	// Combine both sales and purchases
	results = append(results, purchaseResults...)

	// Convert to interface{} slice
	var data []map[string]interface{}
	for _, result := range results {
		data = append(data, map[string]interface{}{
			"id":             result.ID,
			"transaction_id": result.TransactionID,
			"description":    result.Description,
			"amount":         result.Amount,
			"date":           result.Date,
			"type":           result.Type,
			"account_name":   result.AccountName,
			"contact_name":   result.ContactName,
			"status":         result.Status,
		})
	}

	return data
}
